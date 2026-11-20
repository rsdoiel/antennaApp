/*
antennaApp is a package for creating and curating blog, link blogs and social websites
Copyright (C) 2025 R. S. Doiel

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
*/
package antennaApp

import (
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// macroPattern matches {{page.KEY}}, with optional spaces inside the braces
// as Jekyll allows (DR-0009 6). KEY is letters, digits, underscore and hyphen.
var macroPattern = regexp.MustCompile(`\{\{\s*page\.([A-Za-z0-9_-]+)\s*\}\}`)

/** MacroExpander expands BonBon's {{page.NAME}} macros while a page or post is
 * rendered to HTML (antennaApp DR-0009 6). Set it as CommonMark.Macros; nil
 * (the default, and what harvested feed items get) expands nothing.
 *
 * Expansion works on the parsed document's text nodes, never on the source
 * text, so a value is inserted as escaped text and can never become Markdown,
 * a link or HTML. Code spans and code blocks are never touched. It makes one
 * pass over the page's own text, so a value holding a macro is shown as
 * written. A key the page does not have leaves the macro visible and is
 * reported once per page. When the page has a `link` value, the expanded
 * `title` is a link to it (BonBon's rule); an address that is not http, https,
 * mailto or relative is ignored with a warning.
 *
 * Fields:
 *   From (string) — the page's output path, named in warnings
 *   Warn (io.Writer) — receives one "warning: ..." line per problem; nil
 *     discards them
 */
type MacroExpander struct {
	From string
	Warn io.Writer
}

// warnf writes one warning line.
func (m *MacroExpander) warnf(format string, args ...any) {
	if m.Warn != nil {
		fmt.Fprintf(m.Warn, "warning: %s: %s\n", m.From, fmt.Sprintf(format, args...))
	}
}

// macroValue returns the text of a front matter value for a macro. ok is false
// for an absent key, a null, or a value that is not text, a number, a date or a
// list of those.
func macroValue(fm map[string]interface{}, key string) (string, bool) {
	switch v := fm[key].(type) {
	case nil:
		return "", false
	case string:
		return v, true
	case bool, int, int64, uint64, float64:
		return fmt.Sprint(v), true
	case time.Time:
		if v.Hour() == 0 && v.Minute() == 0 && v.Second() == 0 {
			return v.Format("2006-01-02"), true
		}
		return v.Format(time.RFC3339), true
	case []interface{}:
		var parts []string
		for _, item := range v {
			if s, ok := macroValue(map[string]interface{}{"x": item}, "x"); ok {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, ", "), true
	}
	return "", false
}

// macroTransformer is the goldmark AST transformer that applies a MacroExpander.
type macroTransformer struct {
	m  *MacroExpander
	fm map[string]interface{}
}

// Transform expands the macros in the text nodes of doc.
func (t macroTransformer) Transform(doc *ast.Document, reader text.Reader, _ parser.Context) {
	src := reader.Source()
	var texts []*ast.Text
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		switch v := n.(type) {
		case *ast.CodeSpan:
			return ast.WalkSkipChildren, nil
		case *ast.Text:
			if entering {
				texts = append(texts, v)
			}
		}
		return ast.WalkContinue, nil
	})
	seen := map[string]bool{}
	merged := map[*ast.Text]bool{}
	for _, n := range texts {
		if merged[n] {
			continue
		}
		// goldmark splits a run of text at characters such as _ and *: join
		// touching text nodes when the joined text holds a macro, so a key
		// such as pub_date is found whole.
		run := []*ast.Text{n}
		stop := n.Segment.Stop
		for last := n; !last.SoftLineBreak() && !last.HardLineBreak() && !last.IsRaw(); {
			next, ok := last.NextSibling().(*ast.Text)
			if !ok || next.Segment.Start != stop || next.IsRaw() {
				break
			}
			run = append(run, next)
			stop = next.Segment.Stop
			last = next
		}
		if len(run) > 1 && macroPattern.Match(src[n.Segment.Start:stop]) {
			last := run[len(run)-1]
			n.Segment = n.Segment.WithStop(stop)
			n.SetSoftLineBreak(last.SoftLineBreak())
			n.SetHardLineBreak(last.HardLineBreak())
			for _, extra := range run[1:] {
				merged[extra] = true
				extra.Parent().RemoveChild(extra.Parent(), extra)
			}
		}
		t.expandText(n, src, seen)
	}
}

// expandText replaces the macros in one text node with their values.
func (t macroTransformer) expandText(n *ast.Text, src []byte, seen map[string]bool) {
	seg := n.Segment
	matches := macroPattern.FindAllSubmatchIndex(src[seg.Start:seg.Stop], -1)
	if len(matches) == 0 {
		return
	}
	parent := n.Parent()
	insert := func(node ast.Node) { parent.InsertBefore(parent, n, node) }
	pos := seg.Start
	for _, m := range matches {
		key := string(src[seg.Start+m[2] : seg.Start+m[3]])
		val, ok := macroValue(t.fm, key)
		if !ok {
			if !seen[key] {
				seen[key] = true
				t.m.warnf("{{page.%s}}: the page has no front matter value for %s, so the macro is left as written", key, key)
			}
			continue // stays in the surrounding text
		}
		if before := seg.Start + m[0]; before > pos {
			insert(ast.NewTextSegment(text.NewSegment(pos, before)))
		}
		pos = seg.Start + m[1]
		str := ast.NewString([]byte(val))
		str.SetRaw(true) // written HTML-escaped, never parsed or entity-resolved
		// A title inside a link the author wrote stays plain text: nested links
		// are not valid HTML.
		if dest := t.titleLink(key); dest != "" && !insideLink(n) {
			link := ast.NewLink()
			link.Destination = []byte(dest)
			link.AppendChild(link, str)
			insert(link)
		} else {
			insert(str)
		}
	}
	if pos == seg.Start {
		return // nothing was expanded
	}
	tail := ast.NewTextSegment(text.NewSegment(pos, seg.Stop))
	tail.SetSoftLineBreak(n.SoftLineBreak())
	tail.SetHardLineBreak(n.HardLineBreak())
	insert(tail)
	parent.RemoveChild(parent, n)
}

// insideLink reports whether n sits inside a link or autolink.
func insideLink(n ast.Node) bool {
	for p := n.Parent(); p != nil; p = p.Parent() {
		switch p.(type) {
		case *ast.Link, *ast.AutoLink:
			return true
		}
	}
	return false
}

// titleLink returns the address the expanded title links to, or "" when key is
// not the title or the page has no usable link.
func (t macroTransformer) titleLink(key string) string {
	if key != "title" {
		return ""
	}
	raw, ok := t.fm["link"].(string)
	if !ok || strings.TrimSpace(raw) == "" {
		return ""
	}
	raw = strings.TrimSpace(raw)
	if !isSafePageAddress(raw) {
		t.m.warnf("ignoring link %q: only http, https, mailto and relative addresses make the title a link", raw)
		return ""
	}
	return raw
}

// isSafePageAddress reports whether raw can be used as the address of a link
// that front matter supplies: an http, https or mailto address, or a relative
// one. Script and data schemes are refused.
func isSafePageAddress(raw string) bool {
	u, err := url.Parse(strings.TrimSpace(raw))
	return err == nil && (u.Scheme == "" || u.Scheme == "http" || u.Scheme == "https" || u.Scheme == "mailto")
}

// macroOption returns the parser option for doc's MacroExpander, or nil.
func (doc *CommonMark) macroOption() parser.Option {
	if doc.Macros == nil {
		return nil
	}
	return parser.WithASTTransformers(util.Prioritized(macroTransformer{doc.Macros, doc.FrontMatter}, 50))
}
