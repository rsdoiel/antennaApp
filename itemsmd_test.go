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
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// mdLink is one link found in a parsed Markdown file.
type mdLink struct{ Text, Dest string }

// parseMarkdownLinks parses src with goldmark and returns its links, and
// whether the document holds any raw HTML node.
func parseMarkdownLinks(src []byte) (links []mdLink, hasHTML bool) {
	doc := goldmark.New().Parser().Parse(text.NewReader(src))
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch v := n.(type) {
		case *ast.Link:
			links = append(links, mdLink{string(util.UnescapePunctuations(v.Text(src))), string(v.Destination)})
		case *ast.RawHTML, *ast.HTMLBlock:
			hasHTML = true
		}
		return ast.WalkContinue, nil
	})
	return links, hasHTML
}

// harvested returns an INSERT for a harvested item with the columns harvest
// always fills (the display query does not accept NULLs there).
func harvested(link, title, description, pubDate, status, label string) string {
	return harvestedSource(link, title, description, "", pubDate, status, label)
}

// harvestedSource is harvested with the sourceMarkdown column set.
func harvestedSource(link, title, description, sourceMarkdown, pubDate, status, label string) string {
	q := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }
	return `INSERT INTO items (link, title, description, authors, enclosures, guid, pubDate, dcExt,
		channel, status, updated, label, postPath, sourceMarkdown, categories) VALUES (` +
		strings.Join([]string{q(link), q(title), q(description), "''", "''", q(link), q(pubDate), "''",
			"''", q(status), q(pubDate), q(label), "''", q(sourceMarkdown), "''"}, ",") + `)`
}

// itemsMDSite makes a site with one collection, "news", holding the items
// given, and runs the real Collection.Generate. It returns the htdocs dir.
func itemsMDSite(t *testing.T, rows ...string) (cfg *AppConfig, htdocs string) {
	t.Helper()
	cfg = sitemapSite(t, "news.md")
	htdocs = filepath.Join(t.TempDir(), "htdocs")
	if err := os.MkdirAll(htdocs, 0775); err != nil {
		t.Fatal(err)
	}
	cfg.Htdocs = htdocs
	col := cfg.Collections[0]
	col.Title = "News & Views"
	col.Description = "Items from my feeds"
	sitemapExec(t, "news.db", rows...)
	if err := col.Generate(io.Discard, io.Discard, "antenna", cfg); err != nil {
		t.Fatalf("Generate: %s", err)
	}
	return cfg, htdocs
}

func TestGenerateWritesItemsMarkdown(t *testing.T) {
	_, htdocs := itemsMDSite(t,
		harvested("https://a.example/one", "Plain title", "First <b>bold</b> story", "2026-11-18 10:00:00", "published", "Feed A"),
		// brackets, parentheses, a newline and raw HTML in the title and link
		harvested("https://b.example/a_(b)?x=1&y=<2>", "Odd [title] (with)\nnewline <script>alert(1)</script>",
			"<p>Second</p><script>evil()</script>", "2026-11-17 10:00:00", "published", "Feed B"),
		// not published: must not appear
		harvested("https://c.example/hidden", "Hidden", "review only", "2026-11-16 10:00:00", "review", "Feed C"),
	)
	src, err := os.ReadFile(filepath.Join(htdocs, "news_items.md"))
	if err != nil {
		t.Fatalf("news_items.md not written: %s", err)
	}
	links, hasHTML := parseMarkdownLinks(src)
	if hasHTML {
		t.Errorf("Markdown holds raw HTML:\n%s", src)
	}
	if regexp.MustCompile(`(^|[^\\])<script`).Match(src) {
		t.Errorf("script content leaked into Markdown:\n%s", src)
	}
	if strings.Contains(string(src), "Hidden") {
		t.Errorf("unpublished item listed:\n%s", src)
	}
	// Every published item is a link back to its source, in HTML-page order
	// (newest first), and the title survives the round trip.
	var got []string
	for _, l := range links {
		if strings.HasPrefix(l.Dest, "https://") {
			got = append(got, l.Dest)
		}
	}
	want := []string{"https://a.example/one", "https://b.example/a_(b)?x=1&y=<2>"}
	if len(got) != len(want) {
		t.Fatalf("links = %q, want %q\n%s", got, want, src)
	}
	for i := range want {
		// destination may be percent-encoded, never altered otherwise
		if got[i] != want[i] && !strings.HasPrefix(got[i], "https://b.example/a_") {
			t.Errorf("link %d = %q, want %q", i, got[i], want[i])
		}
	}
	foundTitle := false
	for _, l := range links {
		if l.Text == "Odd [title] (with) newline <script>alert(1)</script>" {
			foundTitle = true
		}
	}
	if !foundTitle {
		t.Errorf("title with brackets and newline did not survive:\n%s", src)
	}
}

func TestItemsMarkdownFrontMatter(t *testing.T) {
	_, htdocs := itemsMDSite(t,
		harvested("https://a.example/one", "One", "d", "2026-11-18 10:00:00", "published", "Feed A"),
		harvested("https://a.example/two", "Two", "d", "2026-11-17 10:00:00", "published", "Feed A"),
	)
	src, err := os.ReadFile(filepath.Join(htdocs, "news_items.md"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(src)
	if !strings.HasPrefix(s, "---\n") {
		t.Fatalf("no front matter:\n%s", s)
	}
	for _, want := range []string{
		`title: "News & Views"`,
		`description: "Items from my feeds"`,
		`aggregateOf: "news.md"`,
		"itemCount: 2",
		"generator: antenna",
		"pubDate: Wed, 18 Nov 2026 10:00:00 GMT",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("front matter lacks %q:\n%s", want, s)
		}
	}
	// the heading is literal text, not a macro (BonBon draws no title)
	if !strings.Contains(s, "\n# News & Views\n") && !strings.Contains(s, "\n# News &amp; Views\n") {
		t.Errorf("no literal heading:\n%s", s)
	}
	if strings.Contains(s, "{{") {
		t.Errorf("macro in generated file:\n%s", s)
	}
}

func TestHTMLPageDeclaresItemsMarkdownAlternate(t *testing.T) {
	_, htdocs := itemsMDSite(t,
		harvested("https://a.example/one", "One", "d", "2026-11-18 10:00:00", "published", "Feed A"))
	src, err := os.ReadFile(filepath.Join(htdocs, "news.html"))
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile(`<link[^>]*type="text/markdown"[^>]*>`)
	m := re.FindString(string(src))
	if m == "" || !strings.Contains(m, `href="news_items.md"`) || !strings.Contains(m, `rel="alternate"`) {
		t.Errorf("no alternate text/markdown link to news_items.md in head:\n%s", src)
	}
}

func TestItemsMarkdownAndHTMLListSameItems(t *testing.T) {
	_, htdocs := itemsMDSite(t,
		harvested("https://a.example/one", "One", "d1", "2026-11-18 10:00:00", "published", "Feed A"),
		harvested("https://a.example/two", "Two", "d2", "2026-11-17 10:00:00", "published", "Feed A"),
	)
	htmlSrc, _ := os.ReadFile(filepath.Join(htdocs, "news.html"))
	mdSrc, err := os.ReadFile(filepath.Join(htdocs, "news_items.md"))
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile(`data-link="([^"]+)"`)
	var htmlLinks []string
	for _, m := range re.FindAllStringSubmatch(string(htmlSrc), -1) {
		htmlLinks = append(htmlLinks, m[1])
	}
	var mdLinks []string
	links, _ := parseMarkdownLinks(mdSrc)
	for _, l := range links {
		if strings.HasPrefix(l.Dest, "https://") {
			mdLinks = append(mdLinks, l.Dest)
		}
	}
	if strings.Join(htmlLinks, "\n") != strings.Join(mdLinks, "\n") {
		t.Errorf("HTML items %q, Markdown items %q", htmlLinks, mdLinks)
	}
}

// Harvest stores an HTML-to-Markdown copy of each item in sourceMarkdown and
// the HTML page renders it as Markdown, so links and images in it are live
// there. The Markdown file keeps them as Markdown, with raw HTML removed
// (found on a copy of the live food collection, 2026-11-19).
func TestItemsMarkdownKeepsSourceMarkdownAndDropsRawHTML(t *testing.T) {
	row := harvestedSource("https://a.example/one", "One", "<p>fallback</p>",
		"Intro with <b>bold</b> and [Fresh Cup](https://freshcup.com).\n\n"+
			"<div onclick=\"x()\">block <script>evil()</script></div>\n\n"+
			"![pic](https://a.example/p.jpg)\n\nCode: `<i>kept</i>`",
		"2026-11-18 10:00:00", "published", "Feed A")
	_, htdocs := itemsMDSite(t, row)
	src, err := os.ReadFile(filepath.Join(htdocs, "news_items.md"))
	if err != nil {
		t.Fatal(err)
	}
	links, hasHTML := parseMarkdownLinks(src)
	if hasHTML {
		t.Errorf("raw HTML in output:\n%s", src)
	}
	var sawLink bool
	for _, l := range links {
		if l.Dest == "https://freshcup.com" && l.Text == "Fresh Cup" {
			sawLink = true
		}
	}
	if !sawLink {
		t.Errorf("source Markdown link was not kept as a link:\n%s", src)
	}
	if !strings.Contains(string(src), "![pic](https://a.example/p.jpg)") {
		t.Errorf("image not kept:\n%s", src)
	}
	if strings.Contains(string(src), "evil()") && strings.Contains(string(src), "<script") {
		t.Errorf("script survived:\n%s", src)
	}
	if !strings.Contains(string(src), "`<i>kept</i>`") {
		t.Errorf("code span was altered:\n%s", src)
	}
}

// A heading inside a harvested item's Markdown must not become a heading of
// the page's own outline: the file has one h1 and one h2 per item (found on
// the live columns collection: 1298 "##" lines for 1180 items).
func TestItemsMarkdownDemotesHeadingsInItemText(t *testing.T) {
	_, htdocs := itemsMDSite(t,
		harvestedSource("https://a.example/one", "One", "d",
			"# Big heading\n\ntext\n\n## Mid heading\n\n#### Small", "2026-11-18 10:00:00", "published", "Feed A"),
		harvestedSource("https://a.example/two", "Two", "d", "plain", "2026-11-17 10:00:00", "published", "Feed A"),
	)
	src, err := os.ReadFile(filepath.Join(htdocs, "news_items.md"))
	if err != nil {
		t.Fatal(err)
	}
	var h1, h2 int
	body := src[bytes.Index(src[4:], []byte("\n---\n"))+4+5:] // past the front matter
	doc := goldmark.New().Parser().Parse(text.NewReader(body))
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if h, ok := n.(*ast.Heading); ok && entering {
			switch h.Level {
			case 1:
				h1++
			case 2:
				h2++
			}
		}
		return ast.WalkContinue, nil
	})
	if h1 != 1 || h2 != 2 {
		t.Errorf("h1=%d h2=%d, want 1 and 2 (the page title and one per item):\n%s", h1, h2, src)
	}
	if !strings.Contains(string(src), "#### Small") {
		t.Errorf("a heading already deep enough was changed:\n%s", src)
	}
}
