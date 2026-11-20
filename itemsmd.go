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
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

	// 3rd Party Packages
	"github.com/mmcdole/gofeed"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// itemsMarkdownName is the file name of a collection's Markdown rendering of
// its items, relative to the HTML page's directory (antennaApp DR-0008): the
// collection's source stem plus "_items.md", so the author's own feed list
// (the stem plus ".md") is never overwritten.
func itemsMarkdownName(collectionFile string) string {
	base := collectionFile
	if i := strings.LastIndexAny(base, `/\`); i >= 0 {
		base = base[i+1:]
	}
	if i := strings.LastIndex(base, "."); i > 0 {
		base = base[:i]
	}
	return base + "_items.md"
}

// rssDate formats t in the RSS 2.0 date form BonBon expects in front matter
// (DR-0009), always in GMT.
func rssDate(t time.Time) string {
	return t.UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
}

var mdEntityLike = regexp.MustCompile(`^&(#[0-9]{1,7}|#[xX][0-9a-fA-F]{1,6}|[A-Za-z][A-Za-z0-9]{1,31});`)

// mdEscapeInline returns s as Markdown inline text: whitespace (newlines
// included) is collapsed to single spaces, HTML entities the feed already
// escaped are decoded once, and every character CommonMark could read as
// markup or raw HTML is backslash-escaped, so feed text can neither break out
// of a link label or heading nor introduce HTML.
//
// Parameters:
//
//	s (string) — untrusted feed text
//
// Returns:
//
//	string — text that renders as s and holds no markup
//
// Example:
//
//	mdEscapeInline("Q&A [draft]\n<b>x</b>") // `Q&A \[draft\] \<b\>x\</b\>`
func mdEscapeInline(s string) string {
	s = strings.Join(strings.Fields(html.UnescapeString(s)), " ")
	var b strings.Builder
	for i, r := range s {
		switch r {
		case '\\', '`', '*', '_', '[', ']', '<', '>', '~', '|':
			b.WriteByte('\\')
		case '&':
			if mdEntityLike.MatchString(s[i:]) {
				b.WriteByte('\\')
			}
		}
		b.WriteRune(r)
	}
	return b.String()
}

// mdEscapeBlockStart escapes a paragraph's first characters when they would
// start a heading, quote, list, rule or ordered list instead of text.
func mdEscapeBlockStart(s string) string {
	if s == "" {
		return s
	}
	switch s[0] {
	case '#', '>', '-', '+', '=':
		return `\` + s
	}
	i := 0
	for i < len(s) && unicode.IsDigit(rune(s[i])) {
		i++
	}
	if i > 0 && i < len(s) && (s[i] == '.' || s[i] == ')') {
		return s[:i] + `\` + s[i:]
	}
	return s
}

// mdLinkDestination returns href as a Markdown link destination. Characters
// that end or confuse a destination (space, angle brackets, control
// characters, backslash) are percent-encoded, and the angle-bracket form is
// used when the address holds parentheses.
func mdLinkDestination(href string) string {
	var b strings.Builder
	for _, r := range strings.TrimSpace(href) {
		switch {
		case r == ' ' || r == '<' || r == '>' || r == '\\' || unicode.IsControl(r):
			fmt.Fprintf(&b, "%%%02X", r)
		default:
			b.WriteRune(r)
		}
	}
	dest := b.String()
	if strings.ContainsAny(dest, "()") {
		return "<" + dest + ">"
	}
	return dest
}

// yamlString quotes s as a YAML double-quoted scalar (JSON quoting is valid
// YAML) without escaping &, < or >.
func yamlString(s string) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(s)
	return strings.TrimSpace(buf.String())
}

// stripRawHTMLFromMarkdown returns src with its raw HTML removed and
// everything else, links, images, emphasis, code spans and code blocks,
// left byte for byte. An inline tag is removed and the text around it kept
// (`<b>bold</b>` becomes `bold`); an HTML block is removed whole; a heading
// above level 3 is demoted to level 3, so an item cannot reshape the page
// outline. Raw HTML
// inside code is text and stays. It is how harvested Markdown meets DR-0008
// 2: a Markdown file never holds raw HTML from a feed.
//
// Parameters:
//
//	src (string) — Markdown from a harvested item
//
// Returns:
//
//	string — the same Markdown without raw HTML nodes
//
// Example:
//
//	stripRawHTMLFromMarkdown("a <b>b</b> [c](d)") // "a b [c](d)"
func stripRawHTMLFromMarkdown(src string) string {
	b := []byte(src)
	doc := goldmark.New().Parser().Parse(text.NewReader(b))
	type span struct {
		start, stop int
		repl        string
	}
	var cuts []span
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch v := n.(type) {
		case *ast.RawHTML:
			for i := 0; i < v.Segments.Len(); i++ {
				seg := v.Segments.At(i)
				cuts = append(cuts, span{seg.Start, seg.Stop, ""})
			}
		case *ast.HTMLBlock:
			if lines := v.Lines(); lines.Len() > 0 {
				start, stop := lines.At(0).Start, lines.At(lines.Len()-1).Stop
				if v.HasClosure() && v.ClosureLine.Stop > stop {
					stop = v.ClosureLine.Stop
				}
				cuts = append(cuts, span{start, stop, ""})
			}
		case *ast.Heading:
			// An item's own headings go under the item's level-2 heading.
			if v.Level < 3 && v.Lines().Len() > 0 {
				i := v.Lines().At(0).Start
				for i > 0 && b[i-1] == ' ' {
					i--
				}
				stop := i
				for i > 0 && b[i-1] == '#' {
					i--
				}
				if stop-i == v.Level {
					cuts = append(cuts, span{i, stop, "###"})
				}
			}
		}
		return ast.WalkContinue, nil
	})
	// Cut from the end so earlier offsets stay valid.
	sort.Slice(cuts, func(i, j int) bool { return cuts[i].start > cuts[j].start })
	for _, c := range cuts {
		if c.start >= 0 && c.stop <= len(b) && c.start <= c.stop {
			b = append(append(b[:c.start:c.start], c.repl...), b[c.stop:]...)
		}
	}
	return string(b)
}

// writeItemMarkdown writes one item as a level-2 heading that links to the
// item, a source and date line, and the item's text, following the same
// items: configuration (fields, link, date format, content length, source) as
// WriteItem. HTML in the item is always stripped: a Markdown file never holds
// raw HTML from a harvested feed (DR-0008 2). The returned bool is true when
// the item was omitted (items.link.missing: omit) and nothing was written.
func (gen *Generator) writeItemMarkdown(out io.Writer, link, title, description string,
	authors []*gofeed.Person, sourceMarkdown, pubDate, channel, updated, label string,
	cfg ItemsConfig) (bool, error) {
	cfg.applyDefaults()
	linkRes, err := resolveItemLink(link, title, channel, cfg.Link)
	if err != nil {
		return false, err
	}
	if linkRes.Omit {
		return true, nil
	}
	showField := func(name string) bool {
		if len(cfg.Fields) == 0 {
			return true
		}
		for _, f := range cfg.Fields {
			if f == name {
				return true
			}
		}
		return false
	}
	var parts []string
	linkShown := false
	if showField("title") {
		text := title
		if text == "" {
			text = "@" + label
		}
		heading := mdEscapeInline(text)
		if heading == "" {
			heading = mdEscapeInline(linkRes.Label)
		}
		if !linkRes.AsPlainText {
			heading = "[" + heading + "](" + mdLinkDestination(linkRes.Href) + ")"
			linkShown = true
		}
		parts = append(parts, "## "+heading)
	}
	var meta []string
	if showField("source") && cfg.ShowSource != nil && *cfg.ShowSource && label != "" {
		meta = append(meta, "via "+mdEscapeInline(label))
	}
	if showField("pubDate") {
		if d := formatItemDate(pubDate, cfg.DateFormat); d != "" {
			s := mdEscapeInline(d)
			if updated != "" {
				machinePub := formatItemDate(pubDate, "2006-01-02")
				if machineUpd := formatItemDate(updated, "2006-01-02"); machineUpd != machinePub {
					s += ", updated " + mdEscapeInline(formatItemDate(updated, cfg.DateFormat))
				}
			}
			meta = append(meta, s)
		}
	}
	if len(meta) > 0 {
		parts = append(parts, strings.Join(meta, " · "))
	}
	if showField("content") {
		source := sourceMarkdown
		if source == "" {
			source = description
		}
		if cfg.ContentMaxLength > 0 {
			source = truncateWords(source, cfg.ContentMaxLength)
		}
		if sourceMarkdown != "" {
			// Harvested Markdown stays Markdown, without raw HTML.
			if md := strings.TrimSpace(stripRawHTMLFromMarkdown(source)); md != "" {
				parts = append(parts, md)
			}
		} else {
			for _, para := range regexp.MustCompile(`\n\s*\n`).Split(stripTags(source), -1) {
				if p := mdEscapeInline(para); p != "" {
					parts = append(parts, mdEscapeBlockStart(p))
				}
			}
		}
	}
	// A link that the heading did not carry still gets a line of its own, as
	// the HTML footer gives it.
	if !linkShown && !linkRes.AsPlainText {
		parts = append(parts, "["+mdEscapeInline(linkRes.Label)+"]("+mdLinkDestination(linkRes.Href)+")")
	}
	_, err = fmt.Fprintf(out, "%s\n\n", strings.Join(parts, "\n\n"))
	return false, err
}

// WriteItemsMarkdown writes the Markdown rendering of a collection's
// displayed items (antennaApp DR-0008): front matter that says what the file
// is, a literal heading, then every published item in the order the HTML page
// lists them. It reads the same rows (SQLDisplayItems) and applies the same
// items: configuration as WriteHTML, so the two outputs list the same items.
//
// Parameters:
//
//	out (io.Writer) — receives the Markdown
//	db (*sql.DB) — the collection's database, filters already applied
//	collection (*Collection) — title, description and source file name
//	generated (time.Time) — when the file is generated (dateModified)
//
// Returns:
//
//	int — how many items were written. When it is 0 nothing is written to out
//	      and no warning is given: a collection with nothing to show has no
//	      items file (and its HTML page declares none)
//	error — nil, or the first failure; unreadable rows are reported on the
//	        generator's error stream and returned as a data error after the
//	        rest are written
//
// Example:
//
//	n, err := gen.WriteItemsMarkdown(f, db, col, time.Now())
func (gen *Generator) WriteItemsMarkdown(out io.Writer, db *sql.DB, collection *Collection, generated time.Time) (int, error) {
	gen.Items.applyDefaults()
	if err := gen.Items.validate(); err != nil {
		return 0, err
	}
	rows, err := db.Query(SQLDisplayItems)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var (
		body   bytes.Buffer
		count  int
		newest time.Time
		tally  = &failureTally{}
	)
	for rows.Next() {
		var (
			link, title, description, authorsSrc, enclosuresSrc, guid string
			pubDate, dcExt, channel, status, updated, label, postPath string
			sourceMarkdown, categories                                string
			authors                                                   []*gofeed.Person
		)
		if err := rows.Scan(&link, &title, &description, &authorsSrc,
			&enclosuresSrc, &guid, &pubDate, &dcExt,
			&channel, &status, &updated, &label, &postPath, &sourceMarkdown,
			&categories); err != nil {
			if gen.eout != nil {
				fmt.Fprintf(gen.eout, "error (items markdown): %s\n", err)
			}
			tally.add(dataErrorf("reading item row: %w", err))
			continue
		}
		tally.add(nil)
		if authorsSrc != "" {
			if err := json.Unmarshal([]byte(authorsSrc), &authors); err != nil {
				authors = nil
			}
		}
		omitted, err := gen.writeItemMarkdown(&body, link, title, description, authors,
			sourceMarkdown, pubDate, channel, updated, label, gen.Items)
		if err != nil {
			return 0, err
		}
		if omitted {
			continue
		}
		count++
		if t, ok := parseStoredDate(pubDate); ok && t.After(newest) {
			newest = t
		}
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	if count == 0 {
		// Nothing to show: no file, no front matter, no description warning.
		return 0, partial(tally.err("rows"))
	}
	title := collection.Title
	if title == "" {
		title = gen.Title
	}
	fmt.Fprintln(out, "---")
	if title != "" {
		fmt.Fprintf(out, "title: %s\n", yamlString(title))
	}
	if collection.Description != "" {
		fmt.Fprintf(out, "description: %s\n", yamlString(collection.Description))
	} else if gen.eout != nil {
		fmt.Fprintf(gen.eout, "warning: collection %q has no description for %s\n",
			collection.File, itemsMarkdownName(collection.File))
	}
	if !newest.IsZero() {
		fmt.Fprintf(out, "pubDate: %s\n", rssDate(newest))
	}
	fmt.Fprintf(out, "aggregateOf: %s\n", yamlString(collection.File))
	fmt.Fprintf(out, "dateModified: %s\n", rssDate(generated))
	fmt.Fprintf(out, "itemCount: %d\n", count)
	fmt.Fprintf(out, "generator: %s/%s\n", gen.AppName, gen.Version)
	fmt.Fprintln(out, "---")
	fmt.Fprintln(out)
	if title != "" {
		fmt.Fprintf(out, "# %s\n\n", mdEscapeInline(title))
	}
	if _, err := out.Write(body.Bytes()); err != nil {
		return 0, err
	}
	return count, partial(tally.err("rows"))
}
