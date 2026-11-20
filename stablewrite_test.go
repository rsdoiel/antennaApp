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
	"time"
)

// tickingClock makes every build-time stamp differ from the one before, a
// minute apart (the RSS and OPML stamps have only minute resolution).
func tickingClock(t *testing.T) {
	t.Helper()
	old := clock
	at := time.Date(2026, 11, 19, 20, 0, 0, 0, time.UTC)
	clock = func() time.Time {
		at = at.Add(time.Minute)
		return at
	}
	t.Cleanup(func() { clock = old })
}

var past = time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC)

func mtime(t *testing.T, path string) time.Time {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return info.ModTime()
}

func TestWriteStableKeepsAFileThatDiffersOnlyByItsBuildStamp(t *testing.T) {
	cases := []struct {
		name   string
		stamps []*regexp.Regexp
		a, b   string
	}{
		{"html", htmlStamps, "<head>\n  <meta name=\"date\" content=\"2026-11-19T20:00:00-08:00\" >\n  <title>x</title>\n</head>\n", "<head>\n  <meta name=\"date\" content=\"2026-11-19T21:30:00-08:00\" >\n  <title>x</title>\n</head>\n"},
		{"html, older attribute order", htmlStamps, "<meta content=\"2026-11-11T15:06:03-08:00\" name=\"date\" >\n<p>x</p>\n", "<meta name=\"date\" content=\"2026-11-19T20:00:00-08:00\" >\n<p>x</p>\n"},
		{"items", itemsMDStamps, "---\ntitle: T\ndateModified: Fri, 20 Nov 2026 03:54:08 GMT\nitemCount: 3\n---\n", "---\ntitle: T\ndateModified: Fri, 20 Nov 2026 03:59:59 GMT\nitemCount: 3\n---\n"},
		{"rss", rssStamps, "<channel>\n    <lastBuildDate>19 Nov 26 20:00 -0800</lastBuildDate>\n    <generator>antenna</generator>\n", "<channel>\n    <lastBuildDate>19 Nov 26 21:11 -0800</lastBuildDate>\n    <generator>antenna</generator>\n"},
		{"opml", opmlStamps, "<head>\n        <dateCreated>19 Nov 26 20:00 -0800</dateCreated>\n</head>\n", "<head>\n        <dateCreated>19 Nov 26 21:11 -0800</dateCreated>\n</head>\n"},
	}
	for _, c := range cases {
		path := filepath.Join(t.TempDir(), "out")
		if _, err := writeStable(path, []byte(c.a), 0o644, c.stamps...); err != nil {
			t.Fatalf("%s: %s", c.name, err)
		}
		os.Chtimes(path, past, past)
		written, err := writeStable(path, []byte(c.b), 0o644, c.stamps...)
		if err != nil || written {
			t.Errorf("%s: written %v err %v, want the file kept", c.name, written, err)
		}
		got, _ := os.ReadFile(path)
		if string(got) != c.a {
			t.Errorf("%s: the kept file changed:\n%s", c.name, got)
		}
		if !mtime(t, path).Equal(past) {
			t.Errorf("%s: modification time changed to %s", c.name, mtime(t, path))
		}
	}
}

func TestWriteStableRewritesAnyRealChangeAndTakesTheNewStamp(t *testing.T) {
	path := filepath.Join(t.TempDir(), "p.html")
	old := "<meta name=\"date\" content=\"2026-11-19T20:00:00-08:00\" >\n<p>old text</p>\n"
	neu := "<meta name=\"date\" content=\"2026-11-19T21:00:00-08:00\" >\n<p>new text</p>\n"
	writeStable(path, []byte(old), 0o644, htmlStamps...)
	os.Chtimes(path, past, past)
	written, err := writeStable(path, []byte(neu), 0o644, htmlStamps...)
	if err != nil || !written {
		t.Fatalf("written %v err %v", written, err)
	}
	if got, _ := os.ReadFile(path); string(got) != neu {
		t.Errorf("file:\n%s", got)
	}
	if !mtime(t, path).After(past) {
		t.Error("a rewritten file kept its old modification time")
	}
}

// The author's own `date:` front matter is a <meta name="date"> too, written
// without the space before > that the build stamp has. A change to it is a
// change to the page, not a build stamp.
func TestAnAuthorsDateFrontMatterIsNotMistakenForTheBuildStamp(t *testing.T) {
	path := filepath.Join(t.TempDir(), "p.html")
	a := "<meta name=\"date\" content=\"2026-11-19T20:00:00-08:00\" >\n  <meta name=\"date\" content=\"2024-01-01\">\n"
	b := "<meta name=\"date\" content=\"2026-11-19T21:00:00-08:00\" >\n  <meta name=\"date\" content=\"2025-06-30\">\n"
	writeStable(path, []byte(a), 0o644, htmlStamps...)
	written, _ := writeStable(path, []byte(b), 0o644, htmlStamps...)
	if !written {
		t.Error("a change to the page's own date was treated as a build stamp")
	}
}

func TestWriteStableCreatesAMissingFileAndReplacesAnUnreadableOne(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "new.html")
	if written, err := writeStable(path, []byte("x"), 0o644, htmlStamps...); err != nil || !written {
		t.Errorf("new file: %v %v", written, err)
	}
	if _, err := writeStable(filepath.Join(dir, "no", "such", "dir", "f"), []byte("x"), 0o644); err == nil {
		t.Error("no error for a directory that does not exist")
	}
}

func generateTwice(t *testing.T, cfg *AppConfig, col *Collection) {
	t.Helper()
	var buf bytes.Buffer
	if err := col.Generate(&buf, &buf, "antenna", cfg); err != nil {
		t.Fatalf("Generate: %s\n%s", err, buf.String())
	}
}

// The real path: generating a collection again, with the clock moved on and
// nothing else changed, leaves every output file as it was; when something
// does change, the files that depend on it are rewritten.
func TestGenerateLeavesUnchangedFilesAloneAndRewritesChangedOnes(t *testing.T) {
	tickingClock(t)
	cfg := sitemapSite(t, "news.md")
	htdocs := filepath.Join(t.TempDir(), "htdocs")
	os.MkdirAll(htdocs, 0775)
	cfg.Htdocs = htdocs
	col := cfg.Collections[0]
	col.Title, col.Description = "News", "d"
	col.Link = "news.xml"
	sitemapExec(t, "news.db", harvested("https://a.example/one", "One", "d", "2026-11-18 10:00:00", "published", "Feed A"),
		`INSERT INTO channels (link, title, description, feed_link, links, updated, published, authors, language, copyright, generator, categories, feed_type, feed_version)
		 VALUES ('https://a.example/', 'A', 'a', 'https://a.example/feed', '[]', '2026-11-18 10:00:00', '2026-11-18 10:00:00', '[]', 'en', '', '', '[]', 'rss', '2.0')`)
	generateTwice(t, cfg, col)
	files := []string{"news.html", "news_items.md", "news.xml", "news.opml"}
	for _, f := range files {
		p := filepath.Join(htdocs, f)
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("%s not written: %s", f, err)
		}
		os.Chtimes(p, past, past)
	}
	first := map[string]string{}
	for _, f := range files {
		b, _ := os.ReadFile(filepath.Join(htdocs, f))
		first[f] = string(b)
	}
	generateTwice(t, cfg, col) // the clock has moved on; nothing else has
	for _, f := range files {
		p := filepath.Join(htdocs, f)
		if !mtime(t, p).Equal(past) {
			t.Errorf("%s was rewritten although only the build time changed", f)
		}
		if b, _ := os.ReadFile(p); string(b) != first[f] {
			t.Errorf("%s content changed", f)
		}
	}
	// a new item changes the page, the items file and the feed, but not the channels list
	sitemapExec(t, "news.db", harvested("https://a.example/two", "Two", "d", "2026-11-19 10:00:00", "published", "Feed A"))
	generateTwice(t, cfg, col)
	for _, f := range []string{"news.html", "news_items.md", "news.xml"} {
		if !mtime(t, filepath.Join(htdocs, f)).After(past) {
			t.Errorf("%s was not rewritten after a new item", f)
		}
	}
	if b, _ := os.ReadFile(filepath.Join(htdocs, "news_items.md")); !strings.Contains(string(b), "Two") {
		t.Error("the new item is not in news_items.md")
	}
	if !mtime(t, filepath.Join(htdocs, "news.opml")).Equal(past) {
		t.Error("news.opml was rewritten although the channels did not change")
	}
}

func TestGeneratePagesLeavesUnchangedPagesAlone(t *testing.T) {
	tickingClock(t)
	cfg := sitemapSite(t, "pages.md")
	cfg.Htdocs = "htdocs"
	cfg.Generator = "page.yaml"
	os.WriteFile("page.yaml", []byte("header: |\n  <h1>Test</h1>\n"), 0664)
	os.MkdirAll("htdocs", 0775)
	os.WriteFile("p.md", []byte("---\ntitle: P\n---\n# P\n"), 0664)
	addPage(t, "pages.db", "p.html")
	var eout bytes.Buffer
	if err := cfg.GeneratePages(&eout); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join("htdocs", "p.html")
	os.Chtimes(p, past, past)
	if err := cfg.GeneratePages(&eout); err != nil {
		t.Fatal(err)
	}
	if !mtime(t, p).Equal(past) {
		t.Error("p.html was rewritten although only the build time changed")
	}
	os.WriteFile("p.md", []byte("---\ntitle: P\n---\n# P changed\n"), 0664)
	if err := cfg.GeneratePages(&eout); err != nil {
		t.Fatal(err)
	}
	if !mtime(t, p).After(past) {
		t.Error("p.html was not rewritten after the page changed")
	}
}

func TestTheGenerateManualSaysFilesAreRewrittenOnlyWhenTheirContentChanges(t *testing.T) {
	for _, want := range []string{"_items.md", "rewritten only when its content changes", "lastBuildDate"} {
		if !strings.Contains(GenerateHelpText, want) {
			t.Errorf("antenna help generate lacks %q", want)
		}
	}
}
