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
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/rsdoiel/opml"
)

// tocPost inserts a local post the way antenna post does: an item with a
// postPath, a title and a pubDate.
func tocPost(t *testing.T, db, title, postPath, pubDate string) {
	t.Helper()
	sitemapExec(t, db, `INSERT INTO items (link, postPath, title, description, pubDate, status, updated)
		VALUES ('https://example.org/`+postPath+`', '`+postPath+`', '`+strings.ReplaceAll(title, "'", "''")+`', 'd', '`+pubDate+`', 'published', '`+pubDate+`')`)
}

// publish writes a file under htdocs, so its Markdown counts as published.
func publish(t *testing.T, htdocs, name string) {
	t.Helper()
	p := filepath.Join(htdocs, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(p), 0o775); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("# "+name+"\n"), 0o664); err != nil {
		t.Fatal(err)
	}
}

// tocSite builds a site with the three kinds of content a table of contents
// lists: two pages, a collection with an items page, and posts in a blog.
func tocSite(t *testing.T) (cfg *AppConfig, htdocs string) {
	t.Helper()
	cfg = sitemapSite(t, "pages.md", "news.md", "blog.md")
	htdocs = filepath.Join(t.TempDir(), "htdocs")
	cfg.Htdocs, cfg.BaseURL = htdocs, "https://example.org"
	cfg.Collections[1].Title = "Local News"
	// pages: the Markdown source is read for its title, the published copy is checked
	os.MkdirAll("series", 0o775)
	os.WriteFile("about.md", []byte("---\ntitle: About Me\n---\n# About\n"), 0o664)
	os.WriteFile("series/books.md", []byte("# Books\n"), 0o664)
	addPage(t, "pages.db", "series/books.html")
	addPage(t, "pages.db", "about.html")
	publish(t, htdocs, "about.md")
	publish(t, htdocs, "series/books.md")
	// a collection with an items page, and the feed list that is never listed
	publish(t, htdocs, "news_items.md")
	publish(t, htdocs, "news.md")
	// posts in two years and three months
	for _, p := range []struct{ title, path, date string }{
		{"Second November", "blog/2026/11/02/b.md", "2026-11-02 09:00:00"},
		{"Third November", "blog/2026/11/19/c.md", "2026-11-19 09:00:00"},
		{"First October", "blog/2026/10/30/a.md", "2026-10-30 09:00:00"},
		{"An old one", "blog/2025/12/25/old.md", "2025-12-25 09:00:00"},
	} {
		tocPost(t, "blog.db", p.title, p.path, p.date)
		publish(t, htdocs, p.path)
	}
	return cfg, htdocs
}

// tocStampRE drops the dateModified line, the one part meant to change.
var tocStampRE = regexp.MustCompile(`(?m)^.*<dateModified>.*\n`)

// readTOC parses the written file with a real XML parser.
func readTOC(t *testing.T, htdocs string) (*opml.OPML, string) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(htdocs, ".well-known", "toc.opml"))
	if err != nil {
		t.Fatalf("toc.opml not written: %s", err)
	}
	o, err := opml.Parse(b)
	if err != nil {
		t.Fatalf("toc.opml does not parse: %s\n%s", err, b)
	}
	return o, string(b)
}

func texts(os []*opml.Outline) []string {
	var out []string
	for _, o := range os {
		out = append(out, o.Text)
	}
	return out
}

func find(os []*opml.Outline, text string) *opml.Outline {
	for _, o := range os {
		if o.Text == text {
			return o
		}
	}
	return nil
}

func TestTheTOCListsPagesCollectionsAndPostsInThreeGroups(t *testing.T) {
	cfg, htdocs := tocSite(t)
	var eout bytes.Buffer
	if err := cfg.GenerateTOC(&eout); err != nil {
		t.Fatalf("GenerateTOC: %s\n%s", err, eout.String())
	}
	if eout.Len() != 0 {
		t.Errorf("warnings for a complete site: %s", eout.String())
	}
	o, raw := readTOC(t, htdocs)
	if !strings.HasPrefix(raw, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n") {
		t.Errorf("no XML declaration:\n%.120s", raw)
	}
	if !strings.Contains(raw, "\n\t<head>") || !strings.Contains(raw, "\n\t\t<title>") {
		t.Errorf("not indented:\n%s", raw)
	}
	if o.Head.Title != "example.org" || o.Head.Modified == "" {
		t.Errorf("head %+v", o.Head)
	}
	if got := texts(o.Body.Outline); strings.Join(got, ",") != "Pages,Collections,Posts" {
		t.Fatalf("top-level groups %v, want Pages, Collections, Posts", got)
	}
	for _, g := range o.Body.Outline {
		if g.URL != "" || g.Type != "" {
			t.Errorf("group %q should have text only: %+v", g.Text, g)
		}
	}

	pages := find(o.Body.Outline, "Pages").Outline
	if len(pages) != 2 || pages[0].Text != "About Me" || pages[0].URL != "https://example.org/about.md" ||
		pages[1].Text != "Books" || pages[1].URL != "https://example.org/series/books.md" {
		t.Errorf("pages (output path order, titles from front matter): %+v", pages)
	}

	cols := find(o.Body.Outline, "Collections").Outline
	if len(cols) != 1 || cols[0].Text != "Local News" || cols[0].URL != "https://example.org/news_items.md" {
		t.Errorf("collections (the items page, never the feed list): %+v", cols)
	}

	years := find(o.Body.Outline, "Posts").Outline
	// oldest first, so BonBon's own arrows (previous is the line above) agree
	// with antenna's (DR-0013 amendment, read from BonBon's source 2026-11-19)
	if got := texts(years); strings.Join(got, ",") != "2025,2026" {
		t.Fatalf("years %v, want oldest first", got)
	}
	if years[1].Type != "calendarYear" || attrValue(years[1], "name") != "2026" {
		t.Errorf("year outline %+v", years[1])
	}
	months := years[1].Outline
	if got := texts(months); strings.Join(got, ",") != "October 2026,November 2026" {
		t.Fatalf("months %v", got)
	}
	if months[1].Type != "calendarMonth" || attrValue(months[1], "name") != "11" {
		t.Errorf("month outline %+v", months[1])
	}
	nov := months[1].Outline
	if len(nov) != 2 || nov[0].Text != "Second November" || nov[1].Text != "Third November" {
		t.Errorf("posts in a month, oldest first: %v", texts(nov))
	}
	if nov[1].URL != "https://example.org/blog/2026/11/19/c.md" || nov[1].Type != "" || len(nov[1].Outline) != 0 {
		t.Errorf("a post is a leaf with text and url only: %+v", nov[1])
	}
	if old := years[0].Outline[0].Outline; len(old) != 1 || old[0].Text != "An old one" {
		t.Errorf("2025: %+v", years[0])
	}
}

func attrValue(o *opml.Outline, name string) string {
	for _, a := range o.OtherAttr {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}

func TestNoBaseURLWarnsOnceAndWritesNothing(t *testing.T) {
	cfg, htdocs := tocSite(t)
	cfg.BaseURL = ""
	var eout bytes.Buffer
	if err := cfg.GenerateTOC(&eout); err != nil {
		t.Fatalf("a missing base_url is a warning, not an error: %s", err)
	}
	if _, err := os.Stat(filepath.Join(htdocs, ".well-known", "toc.opml")); err == nil {
		t.Error("a toc.opml was written with no base_url")
	}
	if strings.Count(eout.String(), "\n") != 1 || !strings.Contains(eout.String(), "base_url") || !strings.Contains(eout.String(), "warning") {
		t.Errorf("warnings: %q", eout.String())
	}
}

func TestNothingToListWritesNoFileAndRemovesAStaleOne(t *testing.T) {
	cfg := sitemapSite(t, "pages.md")
	htdocs := filepath.Join(t.TempDir(), "htdocs")
	cfg.Htdocs, cfg.BaseURL = htdocs, "https://example.org"
	publish(t, htdocs, ".well-known/toc.opml") // left by an earlier run
	var eout bytes.Buffer
	if err := cfg.GenerateTOC(&eout); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(htdocs, ".well-known", "toc.opml")); err == nil {
		t.Error("the stale toc.opml is still there")
	}
	if strings.Count(eout.String(), "\n") != 1 || !strings.Contains(eout.String(), "nothing") {
		t.Errorf("want one warning that there is nothing to list: %q", eout.String())
	}
}

// A page is listed only when its Markdown is published in htdocs (DR-0007 3).
func TestAPageWhoseMarkdownIsNotPublishedIsLeftOutWithAWarning(t *testing.T) {
	cfg, htdocs := tocSite(t)
	os.Remove(filepath.Join(htdocs, "series", "books.md"))
	var eout bytes.Buffer
	if err := cfg.GenerateTOC(&eout); err != nil {
		t.Fatal(err)
	}
	o, _ := readTOC(t, htdocs)
	pages := find(o.Body.Outline, "Pages").Outline
	if len(pages) != 1 || pages[0].Text != "About Me" {
		t.Errorf("pages %+v", pages)
	}
	if !strings.Contains(eout.String(), "series/books") || strings.Count(eout.String(), "\n") != 1 {
		t.Errorf("want one warning naming the page: %q", eout.String())
	}
	// with every page unpublished the group is left out, not left empty
	os.Remove(filepath.Join(htdocs, "about.md"))
	cfg.GenerateTOC(&bytes.Buffer{})
	o, _ = readTOC(t, htdocs)
	if find(o.Body.Outline, "Pages") != nil {
		t.Errorf("an empty Pages group was written: %v", texts(o.Body.Outline))
	}
}

func TestAGroupWithNothingInItIsLeftOut(t *testing.T) {
	cfg, htdocs := tocSite(t)
	os.Remove(filepath.Join(htdocs, "news_items.md")) // the collection has no items page now
	cfg.GenerateTOC(&bytes.Buffer{})
	o, _ := readTOC(t, htdocs)
	if got := texts(o.Body.Outline); strings.Join(got, ",") != "Pages,Posts" {
		t.Errorf("groups %v", got)
	}
}

func TestPostsThatCannotBeListedAreSkippedWithAWarning(t *testing.T) {
	cfg, htdocs := tocSite(t)
	tocPost(t, "blog.db", "A post that is not Markdown", "blog/2026/11/20/not-markdown.odt", "2026-11-20 09:00:00") // no Markdown page
	tocPost(t, "blog.db", "Bad date", "blog/2026/11/21/bad.md", "next Tuesday")            // date antenna cannot read
	tocPost(t, "blog.db", "Unpublished", "blog/2026/11/22/nope.md", "2026-11-22 09:00:00") // Markdown not in htdocs
	publish(t, htdocs, "blog/2026/11/21/bad.md")
	var eout bytes.Buffer
	if err := cfg.GenerateTOC(&eout); err != nil {
		t.Fatal(err)
	}
	o, raw := readTOC(t, htdocs)
	for _, bad := range []string{"A post that is not Markdown", "Bad date", "Unpublished"} {
		if strings.Contains(raw, bad) {
			t.Errorf("%q should not be listed", bad)
		}
	}
	if n := strings.Count(eout.String(), "\n"); n != 3 {
		t.Errorf("want a warning for each of the three: %q", eout.String())
	}
	if len(find(o.Body.Outline, "Posts").Outline) != 2 {
		t.Errorf("the other posts are still listed")
	}
}

func TestTitlesAndPathsAreEscapedForXMLAndAddresses(t *testing.T) {
	cfg, htdocs := tocSite(t)
	tocPost(t, "blog.db", `Q&A: "quotes" <tags> & more`, "blog/2026/11/23/a post.md", "2026-11-23 09:00:00")
	publish(t, htdocs, "blog/2026/11/23/a post.md")
	if err := cfg.GenerateTOC(&bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	o, raw := readTOC(t, htdocs)
	nov := find(find(find(o.Body.Outline, "Posts").Outline, "2026").Outline, "November 2026").Outline
	last := nov[len(nov)-1] // the newest, and the list is oldest first
	if last.Text != `Q&A: "quotes" <tags> & more` {
		t.Errorf("title did not survive: %q\n%s", last.Text, raw)
	}
	if last.URL != "https://example.org/blog/2026/11/23/a%20post.md" {
		t.Errorf("address %q", last.URL)
	}
}

func TestAnUntitledPostIsListedByItsPath(t *testing.T) {
	cfg, htdocs := tocSite(t)
	tocPost(t, "blog.db", "", "blog/2026/11/24/untitled.md", "2026-11-24 09:00:00")
	publish(t, htdocs, "blog/2026/11/24/untitled.md")
	cfg.GenerateTOC(&bytes.Buffer{})
	o, _ := readTOC(t, htdocs)
	nov := find(find(find(o.Body.Outline, "Posts").Outline, "2026").Outline, "November 2026").Outline
	if last := nov[len(nov)-1]; last.Text != "blog/2026/11/24/untitled.md" {
		t.Errorf("untitled post listed as %q", last.Text)
	}
}

// A site under a project path keeps its toc in its own folder, with its own base.
func TestABaseURLWithAPathIsKeptInTheAddressesAndTheTitle(t *testing.T) {
	cfg, htdocs := tocSite(t)
	cfg.BaseURL = "https://rsdoiel.github.io/antenna/"
	cfg.GenerateTOC(&bytes.Buffer{})
	o, _ := readTOC(t, htdocs)
	if o.Head.Title != "rsdoiel.github.io/antenna" {
		t.Errorf("title %q", o.Head.Title)
	}
	if pages := find(o.Body.Outline, "Pages").Outline; pages[0].URL != "https://rsdoiel.github.io/antenna/about.md" {
		t.Errorf("address %q", pages[0].URL)
	}
}

func TestTheSameMarkdownIsListedOnce(t *testing.T) {
	cfg, htdocs := tocSite(t)
	// the same post path in a second collection's database
	tocPost(t, "news.db", "Second November (again)", "blog/2026/11/02/b.md", "2026-11-02 09:00:00")
	cfg.GenerateTOC(&bytes.Buffer{})
	_, raw := readTOC(t, htdocs)
	if n := strings.Count(raw, "blog/2026/11/02/b.md"); n != 1 {
		t.Errorf("listed %d times", n)
	}
}

// DR-0011: an unchanged site leaves the file alone; a change rewrites it with a new date.
func TestTheTOCIsLeftAloneWhenOnlyItsDateWouldChange(t *testing.T) {
	tickingClock(t)
	cfg, htdocs := tocSite(t)
	if err := cfg.GenerateTOC(&bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(htdocs, ".well-known", "toc.opml")
	first, _ := os.ReadFile(p)
	os.Chtimes(p, past, past)
	cfg.GenerateTOC(&bytes.Buffer{}) // the clock has moved on, nothing else has
	if !mtime(t, p).Equal(past) {
		t.Error("toc.opml was rewritten although only its dateModified would change")
	}
	if got, _ := os.ReadFile(p); string(got) != string(first) {
		t.Error("content changed")
	}
	tocPost(t, "blog.db", "A new post", "blog/2026/11/25/new.md", "2026-11-25 09:00:00")
	publish(t, htdocs, "blog/2026/11/25/new.md")
	cfg.GenerateTOC(&bytes.Buffer{})
	got, _ := os.ReadFile(p)
	if !mtime(t, p).After(past) || !strings.Contains(string(got), "A new post") || string(got) == string(first) {
		t.Error("toc.opml was not rewritten after a new post")
	}
}

func TestTheTOCIsTheSameEveryTime(t *testing.T) {
	cfg, htdocs := tocSite(t)
	var first string
	for i := 0; i < 30; i++ {
		cfg.GenerateTOC(&bytes.Buffer{})
		_, raw := readTOC(t, htdocs)
		raw = tocStampRE.ReplaceAllString(raw, "")
		if i == 0 {
			first = raw
		} else if raw != first {
			t.Fatalf("run %d differs:\n%s", i, firstDiff(first, raw))
		}
		os.Remove(filepath.Join(htdocs, ".well-known", "toc.opml"))
	}
}

// The real path: antenna generate writes it, and generate NAME, which makes
// only part of the site, does not.
func TestGenerateWritesTheTOCForTheWholeSiteOnly(t *testing.T) {
	newWorkspace(t, true)
	cfgYAML, _ := os.ReadFile("antenna.yaml")
	// init writes the preview address as base_url; point the site elsewhere
	edited := strings.Replace(string(cfgYAML), "base_url: http://localhost:8000", "base_url: https://example.org\nhtdocs: htdocs", 1)
	if edited == string(cfgYAML) {
		t.Fatalf("antenna.yaml is not what the test expects:\n%s", cfgYAML)
	}
	os.WriteFile("antenna.yaml", []byte(edited), 0o664)
	os.MkdirAll("htdocs", 0o775)
	os.WriteFile("hello.md", []byte("---\ntitle: Hello There\n---\n# Hello\n"), 0o664)
	if err := runIn(t, "page", "hello.md", "hello.html"); err != nil {
		t.Fatalf("page: %s", err)
	}
	publish(t, "htdocs", "hello.md")
	if err := runIn(t, "generate", "pages.md"); err != nil {
		t.Fatalf("generate pages.md: %s", err)
	}
	toc := filepath.Join("htdocs", ".well-known", "toc.opml")
	if _, err := os.Stat(toc); err == nil {
		t.Fatal("generate with a collection name wrote a table of contents for part of the site")
	}
	if err := runIn(t, "generate"); err != nil {
		t.Fatalf("generate: %s", err)
	}
	b, err := os.ReadFile(toc)
	if err != nil {
		t.Fatalf("generate did not write the table of contents: %s", err)
	}
	if !strings.Contains(string(b), "Hello There") || !strings.Contains(string(b), "https://example.org/hello.md") {
		t.Errorf("toc.opml:\n%s", b)
	}
}

func TestTheGenerateManualDescribesTheTOC(t *testing.T) {
	for _, want := range []string{"toc.opml", ".well-known", "base_url"} {
		if !strings.Contains(GenerateHelpText, want) {
			t.Errorf("antenna help generate lacks %q", want)
		}
	}
}

// Two rows of the pages table can make the same page (the live blog has
// index.md and index.txt, both producing index.html): it is listed, or warned
// about, once.
func TestTwoPagesRowsForOnePageAreListedOrWarnedAboutOnce(t *testing.T) {
	cfg, htdocs := tocSite(t)
	sitemapExec(t, "pages.db",
		`INSERT INTO pages (inputPath, outputPath, updated) VALUES ('about.txt', 'about.html', '2026-10-01 12:00:00')`)
	var eout bytes.Buffer
	cfg.GenerateTOC(&eout)
	_, raw := readTOC(t, htdocs)
	if n := strings.Count(raw, "/about.md"); n != 1 {
		t.Errorf("about listed %d times", n)
	}
	// unpublished: one warning, not two
	os.Remove(filepath.Join(htdocs, "about.md"))
	eout.Reset()
	cfg.GenerateTOC(&eout)
	if n := strings.Count(eout.String(), "about.html"); n != 1 {
		t.Errorf("%d warnings for one page: %q", n, eout.String())
	}
}
