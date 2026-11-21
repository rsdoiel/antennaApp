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
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/
package antennaApp

import (
	"bytes"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
)

// bonbonImageExts are the image types BonBon shows (its skill note); anything
// else it drops.
var bonbonImageExts = map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true}

// rawHTMLTag finds the name of the tag an HTML fragment opens.
var rawHTMLTag = regexp.MustCompile(`^<\s*([A-Za-z][A-Za-z0-9-]*)`)

// nodeOffset returns the byte offset in the body where n starts: its own text
// when it has some, else the start of the nearest block that has lines.
func nodeOffset(n ast.Node, src []byte) int {
	switch v := n.(type) {
	case *ast.RawHTML:
		if v.Segments.Len() > 0 {
			return v.Segments.At(0).Start
		}
	case *ast.HTMLBlock:
		if v.Lines().Len() > 0 {
			return v.Lines().At(0).Start
		}
	}
	var first func(ast.Node) int
	first = func(c ast.Node) int {
		if t, ok := c.(*ast.Text); ok {
			return t.Segment.Start
		}
		for k := c.FirstChild(); k != nil; k = k.NextSibling() {
			if off := first(k); off >= 0 {
				return off
			}
		}
		return -1
	}
	if off := first(n); off >= 0 {
		return off
	}
	for p := n.Parent(); p != nil; p = p.Parent() {
		if p.Type() == ast.TypeBlock && p.Lines() != nil && p.Lines().Len() > 0 {
			return p.Lines().At(0).Start
		}
	}
	return 0
}

// lineOf converts a body offset to a file line, given how many lines come
// before the body.
func lineOf(src []byte, offset, before int) int {
	if offset > len(src) {
		offset = len(src)
	}
	return before + 1 + bytes.Count(src[:offset], []byte("\n"))
}

// siteDir returns the directory a relative link on file is resolved against.
func siteDir(file string) string { return path.Dir(relativeToWorkDir(file)) }

// existsInSite reports whether a site-relative path exists in the working
// directory or, when it is set, htdocs.
func (v *validator) existsInSite(site string) bool {
	roots := []string{"."}
	if v.cfg != nil && v.cfg.Htdocs != "" && v.cfg.Htdocs != "." {
		roots = append(roots, v.cfg.Htdocs)
	}
	for _, r := range roots {
		if _, err := os.Stat(filepath.Join(r, filepath.FromSlash(site))); err == nil {
			return true
		}
	}
	return false
}

// htmlToMarkdown maps the HTML page antenna renders to the Markdown sources
// that make it (a collection has two: its feed list and its items page).
func (v *validator) htmlToMarkdown() map[string][]string {
	m := map[string][]string{}
	for md, html := range v.targets {
		m[html] = append(m[html], md)
	}
	return m
}

// markdownFor returns the Markdown a link to the HTML page site should be a
// link to, or "" when there is none: a page or post names its source; a
// collection names its items page when that file exists (DR-0017 7).
func (v *validator) markdownFor(htmlSite string, byHTML map[string][]string) string {
	sources := byHTML[htmlSite]
	switch len(sources) {
	case 0:
		return ""
	case 1:
		return sources[0]
	}
	for _, s := range sources {
		if strings.HasSuffix(strings.TrimSuffix(s, path.Ext(s)), "_items") && v.existsInSite(s) {
			return s
		}
	}
	return ""
}

// checkBody runs the checks that read the Markdown text: links, raw HTML,
// images and macros (DR-0017 4).
func (v *validator) checkBody(file string, raw []byte, doc *CommonMark) {
	src := []byte(doc.Text)
	before := bodyStartLine(raw)
	md := goldmark.New(goldmark.WithExtensions(extension.GFM, extension.Footnote, extension.DefinitionList))
	root := md.Parser().Parse(text.NewReader(src))
	var byHTML map[string][]string
	if v.targets != nil {
		byHTML = v.htmlToMarkdown()
	}
	from := relativeToWorkDir(file)
	_ = ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch node := n.(type) {
		case *ast.Link:
			v.checkLink(file, from, string(node.Destination), lineOf(src, nodeOffset(n, src), before), byHTML)
		case *ast.AutoLink:
			v.checkLink(file, from, string(node.URL(src)), lineOf(src, nodeOffset(n, src), before), byHTML)
		case *ast.Paragraph:
			v.checkIncludeDirectives(file, node, src, before)
		case *ast.Image:
			v.checkImage(file, string(node.Destination), lineOf(src, nodeOffset(n, src), before))
		case *ast.RawHTML:
			var b bytes.Buffer
			for i := 0; i < node.Segments.Len(); i++ {
				seg := node.Segments.At(i)
				b.Write(seg.Value(src))
			}
			v.checkRawHTML(file, b.String(), lineOf(src, nodeOffset(n, src), before))
		case *ast.HTMLBlock:
			var b bytes.Buffer
			for i := 0; i < node.Lines().Len(); i++ {
				seg := node.Lines().At(i)
				b.Write(seg.Value(src))
			}
			v.checkRawHTML(file, b.String(), lineOf(src, nodeOffset(n, src), before))
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})
	v.checkMacros(file, doc, before)
}

// includeDirective matches a line that is an include directive and nothing
// before it: the form antenna expands.
var includeDirective = regexp.MustCompile(`^\s*@include-(text|code)-block\s+\S`)

// checkIncludeDirectives reports each line of a paragraph that is an include
// directive (DR-0019). Code spans and blocks are not paragraphs, so a page
// that shows a directive in code is left alone.
func (v *validator) checkIncludeDirectives(file string, p *ast.Paragraph, src []byte, before int) {
	for i := 0; i < p.Lines().Len(); i++ {
		seg := p.Lines().At(i)
		if includeDirective.Match(seg.Value(src)) {
			v.add("include-directive", SeverityWarning, file, lineOf(src, seg.Start, before),
				"an include directive in a Markdown file is shown as text by a Markdown reader such as BonBon; keep the source as a .txt file and write this one with antenna page SOURCE.txt OUTPUT.md OUTPUT.html")
		}
	}
}

// checkRawHTML reports raw HTML other than a comment or a closing tag.
func (v *validator) checkRawHTML(file, html string, line int) {
	html = strings.TrimSpace(html)
	if strings.HasPrefix(html, "<!--") || strings.HasPrefix(html, "</") {
		return
	}
	m := rawHTMLTag.FindStringSubmatch(html)
	if m == nil {
		return
	}
	tag := strings.ToLower(m[1])
	extra := ""
	if tag == "script" || tag == "svg" || tag == "style" || tag == "iframe" {
		extra = "; a Markdown reader runs no scripts or styles and shows no SVG"
	}
	v.add("raw-html", SeverityWarning, file, line,
		"raw HTML <%s> is stripped by Markdown readers such as BonBon; write it in Markdown%s", tag, extra)
}

// checkImage reports an image BonBon would not show.
func (v *validator) checkImage(file, dest string, line int) {
	u, err := url.Parse(dest)
	if err != nil || u.Scheme == "data" {
		return
	}
	ext := strings.ToLower(path.Ext(u.Path))
	if ext == "" || bonbonImageExts[ext] {
		return
	}
	v.add("image-type", SeverityWarning, file, line,
		"image %q is not PNG, JPEG, GIF or WebP, the types BonBon shows", dest)
}

// What migrationFor decided about a link to an HTML page.
const (
	migNone    = iota // not a link antenna renders from Markdown: leave it
	migRewrite        // should be the .md link, newDest says which
	migLeft           // antenna renders it but there is no Markdown to point at yet
)

// migrationFor works out what a relative link to the HTML page at site (a
// site-relative path, slash separated) on the page from should be (DR-0017 7):
// the .md source antenna renders it from, written relative to from, with the
// fragment and query kept. A page or post names its source; a collection names
// its items page when that file exists, else migLeft. Both validate (link-html)
// and migrate-links use it, so they cannot disagree.
func (v *validator) migrationFor(from string, u *url.URL, site string, byHTML map[string][]string) (string, int) {
	htmlSite := path.Clean(strings.TrimSuffix(site, path.Ext(site)) + ".html")
	target := v.markdownFor(htmlSite, byHTML)
	if target == "" {
		if len(byHTML[htmlSite]) > 1 {
			return "", migLeft
		}
		return "", migNone
	}
	rel := relativeMarkdownPath(from, target)
	if strings.HasPrefix(u.Path, "/") {
		rel = "/" + target
	}
	return (&url.URL{Path: rel, RawQuery: u.RawQuery, Fragment: u.Fragment}).String(), migRewrite
}

// checkLink runs the link checks for one link destination (DR-0017 4).
func (v *validator) checkLink(file, from, dest string, line int, byHTML map[string][]string) {
	u, err := url.Parse(dest)
	if err != nil {
		return
	}
	if u.Scheme == "http" || u.Scheme == "https" {
		if v.cfg != nil && strings.TrimSpace(v.cfg.BaseURL) != "" &&
			strings.HasPrefix(dest, strings.TrimRight(v.cfg.BaseURL, "/")+"/") && v.isOwnAddress(u, byHTML) {
			v.add("link-absolute-self", SeverityInfo, file, line,
				"%s is an address on this site; a relative link works for every reader", dest)
		}
		return
	}
	if u.Scheme != "" || u.Host != "" || u.Path == "" {
		return
	}
	var site string
	if strings.HasPrefix(u.Path, "/") {
		site = path.Clean(strings.TrimPrefix(u.Path, "/"))
	} else {
		site = path.Join(siteDir(file), u.Path)
	}
	if strings.HasSuffix(u.Path, "/") || site == ".." || strings.HasPrefix(site, "../") {
		return
	}
	switch ext := strings.ToLower(path.Ext(site)); ext {
	case ".html", ".htm":
		if byHTML == nil {
			return
		}
		if suggest, status := v.migrationFor(from, u, site, byHTML); status == migRewrite {
			v.add("link-html", SeverityWarning, file, line,
				"link to %s should be %s: antenna renders that page from Markdown, and a Markdown reader needs the .md link (antenna migrate-links)", dest, suggest)
		}
	case ".md", ".markdown":
		if _, warn := RewriteDestination(from, dest, v.targets); v.targets != nil && warn == "" {
			return
		}
		if !v.existsInSite(site) {
			v.add("link-missing", SeverityWarning, file, line, "%s does not exist", dest)
		} else if v.targets != nil {
			v.add("link-unrendered", SeverityWarning, file, line,
				"%s exists but antenna does not render it, so the HTML link is left as Markdown", dest)
		}
	case "", ".xml", ".opml", ".rss", ".atom":
		// no extension, or a feed antenna generates: nothing to check
	default:
		if !v.existsInSite(site) {
			v.add("link-missing", SeverityWarning, file, line, "%s does not exist", dest)
		}
	}
}

// isOwnAddress reports whether an absolute address under base_url is a page or
// file of this site. base_url names the whole host, which can hold other sites,
// so an address that is none of this site's pages, posts or files is not "on
// this site".
func (v *validator) isOwnAddress(u *url.URL, byHTML map[string][]string) bool {
	base, err := url.Parse(strings.TrimRight(v.cfg.BaseURL, "/"))
	if err != nil {
		return false
	}
	site := path.Clean(strings.TrimPrefix(strings.TrimPrefix(u.Path, base.Path), "/"))
	if site == "." || site == "" {
		return false
	}
	if _, ok := byHTML[site]; ok {
		return true
	}
	if _, ok := v.targets[site]; ok {
		return true
	}
	return v.existsInSite(site)
}

// checkMacros reports a {{page.NAME}} that names a key the page does not have,
// using the same expander the renderer uses.
func (v *validator) checkMacros(file string, doc *CommonMark, before int) {
	var warn bytes.Buffer
	probe := &CommonMark{FrontMatter: doc.FrontMatter, Text: doc.Text, Macros: &MacroExpander{From: file, Warn: &warn}}
	if _, err := probe.ToUnsafeHTML(); err != nil {
		return
	}
	for _, key := range macroKeysMissing(warn.String()) {
		line := 1
		if i := strings.Index(doc.Text, "page."+key); i >= 0 {
			line = lineOf([]byte(doc.Text), i, before)
		}
		v.add("macro-missing", SeverityWarning, file, line,
			"{{page.%s}} names a key this page does not have, so the macro is left visible", key)
	}
}

var macroWarnKey = regexp.MustCompile(`\{\{page\.([A-Za-z0-9_-]+)\}\}: the page has no front matter value`)

// macroKeysMissing returns the keys the expander warned about, in order.
func macroKeysMissing(warnings string) []string {
	var keys []string
	for _, m := range macroWarnKey.FindAllStringSubmatch(warnings, -1) {
		keys = append(keys, m[1])
	}
	return keys
}
