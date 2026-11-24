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
	"strings"
	"testing"
)

// genCollection builds a site with one collection "news" (the rows given, a
// description unless noDescription), runs the real Collection.Generate with
// the given page.yaml text, and returns htdocs and what generate said.
func genCollection(t *testing.T, pageYAML string, before func(htdocs string), rows ...string) (htdocs, eout string) {
	t.Helper()
	cfg := sitemapSite(t, "news.md")
	htdocs = filepath.Join(t.TempDir(), "htdocs")
	if err := os.MkdirAll(htdocs, 0775); err != nil {
		t.Fatal(err)
	}
	cfg.Htdocs = htdocs
	col := cfg.Collections[0]
	col.Title = "News"
	if pageYAML != "" {
		if err := os.WriteFile("page.yaml", []byte(pageYAML), 0664); err != nil {
			t.Fatal(err)
		}
		col.Generator = "page.yaml"
	}
	sitemapExec(t, "news.db", rows...)
	if before != nil {
		before(htdocs)
	}
	var buf bytes.Buffer
	if err := col.Generate(&buf, &buf, "antenna", cfg); err != nil {
		t.Fatalf("Generate: %s\n%s", err, buf.String())
	}
	return htdocs, buf.String()
}

func markdownAlternate(html string) string {
	for _, line := range strings.Split(html, "\n") {
		if strings.Contains(line, "<link") && strings.Contains(line, `type="text/markdown"`) {
			return line
		}
	}
	return ""
}

// A collection with nothing to show gets no items file, no alternate link and
// no warning about a missing description for a file that was never written.
func TestNoItemsMeansNoItemsFileLinkOrWarning(t *testing.T) {
	htdocs, eout := genCollection(t, "", nil)
	if _, err := os.Stat(filepath.Join(htdocs, "news_items.md")); err == nil {
		t.Error("news_items.md was written for a collection with no items")
	}
	// DR-0034 5: nothing to show, so no items page either, and so no
	// alternate link to a file that was never written.
	for _, name := range []string{"news_items.html", "news.html"} {
		if _, err := os.Stat(filepath.Join(htdocs, name)); err == nil {
			t.Errorf("%s was written for a collection with no items", name)
		}
	}
	if strings.Contains(eout, "news_items.md") {
		t.Errorf("warning about a file that is not written: %q", eout)
	}
}

// An items file left by an earlier run, when the collection had items, is
// removed when it has none now, so the site does not serve a stale page.
func TestAStaleItemsFileIsRemovedWhenTheCollectionIsEmpty(t *testing.T) {
	htdocs, _ := genCollection(t, "", func(htdocs string) {
		os.WriteFile(filepath.Join(htdocs, "news_items.md"), []byte("---\nitemCount: 3\n---\n# stale\n"), 0664)
	})
	if _, err := os.Stat(filepath.Join(htdocs, "news_items.md")); err == nil {
		t.Error("the stale news_items.md is still there")
	}
}

// Items that exist but are all left out of the page (items.link.missing: omit)
// leave nothing to show either.
func TestItemsAllOmittedMeansNoItemsFile(t *testing.T) {
	htdocs, _ := genCollection(t, "items:\n  link:\n    missing: omit\n",
		nil, harvested("", "No link", "d", "2026-11-18 10:00:00", "published", "Feed"))
	if _, err := os.Stat(filepath.Join(htdocs, "news_items.md")); err == nil {
		t.Error("news_items.md was written although every item was omitted")
	}
	if _, err := os.Stat(filepath.Join(htdocs, "news_items.html")); err == nil {
		t.Error("news_items.html was written although every item was omitted")
	}
}

// With items, the file, the link and the description warning stay as before.
func TestAnItemsFileStillComesWithItems(t *testing.T) {
	htdocs, eout := genCollection(t, "", nil,
		harvested("https://a.example/one", "One", "d", "2026-11-18 10:00:00", "published", "Feed"))
	if _, err := os.Stat(filepath.Join(htdocs, "news_items.md")); err != nil {
		t.Fatalf("no items file: %s", err)
	}
	html, _ := os.ReadFile(filepath.Join(htdocs, "news_items.html"))
	if l := markdownAlternate(string(html)); !strings.Contains(l, `href="news_items.md"`) {
		t.Errorf("alternate link: %q", l)
	}
	if !strings.Contains(eout, "no description") {
		t.Errorf("the missing-description warning is gone: %q", eout)
	}
}

// Items arriving later bring the file and the link back.
func TestItemsAppearingLaterRestoreTheFile(t *testing.T) {
	cfg := sitemapSite(t, "news.md")
	htdocs := filepath.Join(t.TempDir(), "htdocs")
	os.MkdirAll(htdocs, 0775)
	cfg.Htdocs = htdocs
	col := cfg.Collections[0]
	var buf bytes.Buffer
	if err := col.Generate(&buf, &buf, "antenna", cfg); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(htdocs, "news_items.md")); err == nil {
		t.Fatal("file written while empty")
	}
	sitemapExec(t, "news.db", harvested("https://a.example/one", "One", "d", "2026-11-18 10:00:00", "published", "Feed"))
	if err := col.Generate(&buf, &buf, "antenna", cfg); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(htdocs, "news_items.md")); err != nil {
		t.Errorf("file not written once there are items: %s", err)
	}
}

// One item row that cannot be read must not stop the collection's other
// output. The row is reported (generate exits with its class), but the HTML
// page, the items Markdown and the RSS feed are still written from the rows
// that were readable. On main the HTML page was written despite the bad row;
// the Markdown step added by DR-0008 must not take that away.
func TestAnUnreadableItemRowDoesNotStopTheOtherOutput(t *testing.T) {
	cfg := sitemapSite(t, "news.md")
	htdocs := filepath.Join(t.TempDir(), "htdocs")
	if err := os.MkdirAll(htdocs, 0775); err != nil {
		t.Fatal(err)
	}
	cfg.Htdocs = htdocs
	col := cfg.Collections[0]
	col.Title = "News"
	sitemapExec(t, "news.db",
		harvested("https://example.org/good", "Good item", "d", "2026-11-18 10:00:00", "published", "Feed"),
		// authors is NULL, which the row reader cannot scan into a string
		`INSERT INTO items (link, title, description, status, pubDate) VALUES ('https://example.org/bad', 'Bad item', 'd', 'published', '2026-11-17 10:00:00')`)
	var buf bytes.Buffer
	err := col.Generate(&buf, &buf, "antenna", cfg)
	wantClass(t, "generate with one unreadable row", err, classData)
	for _, name := range []string{"news_items.html", "news_items.md"} {
		b, rerr := os.ReadFile(filepath.Join(htdocs, name))
		if rerr != nil {
			t.Errorf("%s was not written: %s", name, rerr)
			continue
		}
		if !strings.Contains(string(b), "Good item") {
			t.Errorf("%s does not hold the readable item", name)
		}
	}
}

// A collection switched to page-index mode shows no items, so an items file an
// earlier aggregate run left behind is removed, not served as stale.
func TestAStaleItemsFileIsRemovedWhenTheCollectionBecomesAPageIndex(t *testing.T) {
	cfg := sitemapSite(t, "news.md")
	htdocs := filepath.Join(t.TempDir(), "htdocs")
	if err := os.MkdirAll(htdocs, 0775); err != nil {
		t.Fatal(err)
	}
	cfg.Htdocs = htdocs
	col := cfg.Collections[0]
	col.Title = "News"
	col.Mode = "page-index"
	stale := filepath.Join(htdocs, "news_items.md")
	if err := os.WriteFile(stale, []byte("---\nitemCount: 3\n---\n# stale\n"), 0664); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := col.Generate(&buf, &buf, "antenna", cfg); err != nil {
		t.Fatalf("Generate: %s\n%s", err, buf.String())
	}
	if _, err := os.Stat(stale); err == nil {
		t.Error("the stale news_items.md is still there")
	}
}

// The RSS feed, like the HTML page, is written from the rows that can be read.
// A row that cannot be scanned used to end the feed at that row, so the items
// after it were missing from news.xml; generate must still report the row.
func TestAnUnreadableItemRowDoesNotCutTheRSSFeedShort(t *testing.T) {
	cfg := sitemapSite(t, "news.md")
	htdocs := filepath.Join(t.TempDir(), "htdocs")
	if err := os.MkdirAll(htdocs, 0775); err != nil {
		t.Fatal(err)
	}
	cfg.Htdocs = htdocs
	col := cfg.Collections[0]
	col.Title = "News"
	sitemapExec(t, "news.db",
		harvested("https://example.org/new", "Newest item", "d", "2026-11-19 10:00:00", "published", "Feed"),
		// authors is NULL, which the row reader cannot scan into a string
		`INSERT INTO items (link, title, description, status, pubDate) VALUES ('https://example.org/bad', 'Bad item', 'd', 'published', '2026-11-18 10:00:00')`,
		harvested("https://example.org/old", "Oldest item", "d", "2026-11-17 10:00:00", "published", "Feed"))
	var buf bytes.Buffer
	err := col.Generate(&buf, &buf, "antenna", cfg)
	wantClass(t, "generate with one unreadable row", err, classData)
	b, rerr := os.ReadFile(filepath.Join(htdocs, "news.xml"))
	if rerr != nil {
		t.Fatalf("news.xml was not written: %s", rerr)
	}
	for _, want := range []string{"Newest item", "Oldest item"} {
		if !strings.Contains(string(b), want) {
			t.Errorf("news.xml lacks %q, the feed stopped at the unreadable row:\n%s", want, b)
		}
	}
}
