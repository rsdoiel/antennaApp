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

	"gopkg.in/yaml.v3"
)

// genOptOut runs the real Collection.Generate for a published item in the
// collection file given, with markdownItems set as given (nil leaves it
// unset), after writing stale to the items file name when stale is not empty.
func genOptOut(t *testing.T, file string, markdownItems *bool, stale string) (htdocs, itemsName string) {
	t.Helper()
	cfg := sitemapSite(t, file)
	htdocs = filepath.Join(t.TempDir(), "htdocs")
	if err := os.MkdirAll(htdocs, 0775); err != nil {
		t.Fatal(err)
	}
	cfg.Htdocs = htdocs
	col := cfg.Collections[0]
	col.Title = "Coll"
	col.MarkdownItems = markdownItems
	sitemapExec(t, strings.TrimSuffix(file, ".md")+".db",
		harvested("https://a.example/one", "One", "d", "2026-11-18 10:00:00", "published", "Feed"))
	itemsName = itemsMarkdownName(file)
	if stale != "" {
		os.WriteFile(filepath.Join(htdocs, itemsName), []byte(stale), 0664)
	}
	var buf bytes.Buffer
	if err := col.Generate(&buf, &buf, "antenna", cfg); err != nil {
		t.Fatalf("Generate: %s\n%s", err, buf.String())
	}
	return htdocs, itemsName
}

func boolPtr(b bool) *bool { return &b }

// The pages collection, the site's own pages, gets no items file by default,
// no alternate link, and a stale file from an earlier version is removed.
func TestPagesCollectionHasNoItemsFileByDefault(t *testing.T) {
	htdocs, name := genOptOut(t, "pages.md", nil, "# stale\n")
	if _, err := os.Stat(filepath.Join(htdocs, name)); err == nil {
		t.Errorf("%s was written for the pages collection", name)
	}
	html, err := os.ReadFile(filepath.Join(htdocs, "pages.html"))
	if err != nil {
		t.Fatal(err)
	}
	if l := markdownAlternate(string(html)); l != "" {
		t.Errorf("alternate link for an items file that is not written: %s", l)
	}
}

// markdownItems: true brings the file back for the pages collection.
func TestPagesCollectionCanOptIn(t *testing.T) {
	htdocs, name := genOptOut(t, "pages.md", boolPtr(true), "")
	if _, err := os.Stat(filepath.Join(htdocs, name)); err != nil {
		t.Fatalf("no items file: %s", err)
	}
	html, _ := os.ReadFile(filepath.Join(htdocs, "pages.html"))
	if l := markdownAlternate(string(html)); !strings.Contains(l, `href="pages_items.md"`) {
		t.Errorf("alternate link: %q", l)
	}
}

// markdownItems: false skips any collection and leaves its HTML and RSS.
func TestMarkdownItemsFalseSkipsACollection(t *testing.T) {
	htdocs, name := genOptOut(t, "news.md", boolPtr(false), "# stale\n")
	if _, err := os.Stat(filepath.Join(htdocs, name)); err == nil {
		t.Errorf("%s was written although markdownItems is false", name)
	}
	for _, f := range []string{"news.html", "news.xml"} {
		if _, err := os.Stat(filepath.Join(htdocs, f)); err != nil {
			t.Errorf("%s: %s", f, err)
		}
	}
	html, _ := os.ReadFile(filepath.Join(htdocs, "news.html"))
	if l := markdownAlternate(string(html)); l != "" {
		t.Errorf("alternate link: %s", l)
	}
}

// Other collections keep the file when the setting is absent.
func TestOtherCollectionsKeepTheItemsFile(t *testing.T) {
	htdocs, name := genOptOut(t, "news.md", nil, "")
	if _, err := os.Stat(filepath.Join(htdocs, name)); err != nil {
		t.Fatalf("no items file: %s", err)
	}
}

// The setting is read from antenna.yaml as markdownItems.
func TestMarkdownItemsIsReadFromYAML(t *testing.T) {
	var col Collection
	if err := yaml.Unmarshal([]byte("file: pages.md\nmarkdownItems: false\n"), &col); err != nil {
		t.Fatal(err)
	}
	if col.MarkdownItems == nil || *col.MarkdownItems {
		t.Errorf("markdownItems: %v", col.MarkdownItems)
	}
}
