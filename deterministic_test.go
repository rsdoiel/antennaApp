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
)

// The same input must give the same bytes, so regenerating a site changes only
// what changed. Go ranges over maps in random order; these tests run each case
// many times to catch any output that still depends on it.
const repeats = 60

func TestElementFromMapWritesAttributesInAFixedOrder(t *testing.T) {
	cases := []struct {
		element string
		attrs   map[string]string
		want    string
	}{
		{"meta", map[string]string{"content": "antenna/0.0.28", "name": "generator"}, `<meta name="generator" content="antenna/0.0.28" >`},
		{"meta", map[string]string{"charset": "utf-8"}, `<meta charset="utf-8" >`},
		{"link", map[string]string{"href": "x.md", "title": "T", "type": "text/markdown", "rel": "alternate"}, `<link rel="alternate" type="text/markdown" title="T" href="x.md" >`},
		{"link", map[string]string{"href": "css/site.css", "rel": "stylesheet"}, `<link rel="stylesheet" href="css/site.css" >`},
		{"script", map[string]string{"src": "m.js", "type": "module"}, `<script type="module" src="m.js" ></script>`},
		// attributes with no place in the fixed order follow it, alphabetically
		{"link", map[string]string{"zeta": "z", "alpha": "a", "rel": "preload", "href": "f", "as": "font"}, `<link rel="preload" href="f" alpha="a" as="font" zeta="z" >`},
	}
	for _, c := range cases {
		for i := 0; i < repeats; i++ {
			if got := elementFromMap(c.element, c.attrs); got != c.want {
				t.Fatalf("run %d: got  %s\n          want %s", i, got, c.want)
			}
		}
	}
}

var dateMetaRE = regexp.MustCompile(`(?m)^.*name="date".*\n`)

// headOf writes a head the way a page does, from front matter with many keys,
// and drops the line holding the time of the run, which is the only line that
// is meant to change.
func headOf(t *testing.T) string {
	t.Helper()
	gen := newTestGenerator()
	gen.Title = "A page"
	gen.Meta = []map[string]string{{"name": "robots", "content": "index"}, {"name": "author", "content": "Me"}}
	gen.Link = []map[string]string{{"rel": "stylesheet", "href": "css/site.css"}, {"rel": "alternate", "type": "application/rss+xml", "title": "Feed", "href": "feed.xml"}}
	gen.Script = []map[string]string{{"src": "a.js", "type": "module"}}
	fm := map[string]interface{}{
		"title": "A page", "description": "d", "author": "Me", "pubDate": "2026-11-19",
		"keywords": []interface{}{"go", "web"}, "series": "S", "seriesNumber": 3,
		"datePublished": "2026-11-19", "dateModified": "2026-11-19", "dateCreated": "2026-11-01",
		"postPath": "p/a.md", "link": "https://x.example/", "zzz": "last", "aaa": "first",
	}
	var buf bytes.Buffer
	gen.writeHeadElement(&buf, "p/a.md", fm)
	return dateMetaRE.ReplaceAllString(buf.String(), "")
}

func TestHeadIsIdenticalEveryTimeItIsWritten(t *testing.T) {
	first := headOf(t)
	for i := 0; i < repeats; i++ {
		if got := headOf(t); got != first {
			t.Fatalf("run %d wrote a different head:\n--- first\n%s\n--- this run\n%s", i, first, got)
		}
	}
}

func TestFrontMatterMetaIsInKeyOrder(t *testing.T) {
	head := headOf(t)
	var names []string
	for _, m := range regexp.MustCompile(`<meta name="([^"]+)" content=`).FindAllStringSubmatch(head, -1) {
		names = append(names, m[1])
	}
	// the generator's own Meta entries come first, in the order they were given,
	// then the front matter's keys in alphabetical order, a list one value at a time
	var fmNames []string
	seen := false
	for _, n := range names {
		if n == "aaa" {
			seen = true
		}
		if seen {
			fmNames = append(fmNames, n)
		}
	}
	for i := 1; i < len(fmNames); i++ {
		if fmNames[i-1] > fmNames[i] {
			t.Errorf("front matter meta out of order: %v", fmNames)
			break
		}
	}
	if len(fmNames) < 8 {
		t.Errorf("front matter meta missing: %v", names)
	}
	if !strings.Contains(head, `<meta name="robots"`) || strings.Index(head, `name="robots"`) > strings.Index(head, `name="aaa"`) {
		t.Errorf("the generator's own meta should come before the front matter's:\n%s", head)
	}
}

// The real path: the same page generated again is the same file, except for
// the line that records when it was made.
func TestAGeneratedPageIsTheSameEveryTime(t *testing.T) {
	cfg := sitemapSite(t, "pages.md")
	cfg.Htdocs = "htdocs"
	cfg.Generator = "page.yaml"
	if err := os.WriteFile("page.yaml", []byte("header: |\n  <h1>Test</h1>\nmeta:\n  - name: robots\n    content: index\nlink:\n  - rel: stylesheet\n    href: css/site.css\n"), 0664); err != nil {
		t.Fatal(err)
	}
	os.MkdirAll("htdocs", 0775)
	os.WriteFile("p.md", []byte("---\ntitle: P\nauthor: Me\ndescription: d\npubDate: \"2026-11-19\"\nkeywords:\n  - a\n  - b\nseries: S\nseriesNumber: 2\nurlPrev: x.md\nrightMarginImage: i.png\n---\n# P\n"), 0664)
	addPage(t, "pages.db", "p.html")
	var first string
	for i := 0; i < 25; i++ {
		var eout bytes.Buffer
		if err := cfg.GeneratePages(&eout); err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(filepath.Join("htdocs", "p.html"))
		if err != nil {
			t.Fatal(err)
		}
		got := dateMetaRE.ReplaceAllString(string(b), "")
		if i == 0 {
			first = got
		} else if got != first {
			t.Fatalf("generation %d differs from the first:\n%s", i, firstDiff(first, got))
		}
	}
}

// firstDiff names the first line where two texts differ.
func firstDiff(a, b string) string {
	al, bl := strings.Split(a, "\n"), strings.Split(b, "\n")
	for i := 0; i < len(al) && i < len(bl); i++ {
		if al[i] != bl[i] {
			return "line " + string(rune('0'+i%10)) + ":\n  " + al[i] + "\n  " + bl[i]
		}
	}
	return "different lengths"
}

// Overlapping sitemap rules: the longest matching prefix decides, not whichever
// the map happens to yield first.
func TestSitemapRulesUseTheLongestMatchingPrefix(t *testing.T) {
	for i := 0; i < repeats; i++ {
		cfg := sitemapSite(t, "pages.md")
		cfg.Htdocs = "htdocs"
		cfg.DefaultFreq, cfg.DefaultPri = "monthly", "0.5"
		cfg.FreqRules = map[string]string{"blog/": "weekly", "blog/2026/": "daily", "b": "yearly"}
		cfg.PriRules = map[string]string{"blog/": "0.6", "blog/2026/": "0.9", "b": "0.1"}
		os.MkdirAll("htdocs", 0775)
		addPage(t, "pages.db", "blog/2026/new.html")
		addPage(t, "pages.db", "blog/2020/old.html")
		addPage(t, "pages.db", "about.html")
		if err := cfg.Sitemap(); err != nil {
			t.Fatalf("Sitemap: %s", err)
		}
		b, err := os.ReadFile(filepath.Join("htdocs", "sitemap_1.xml"))
		if err != nil {
			t.Fatal(err)
		}
		x := string(b)
		for loc, want := range map[string]string{
			"blog/2026/new.html": "<changefreq>daily</changefreq>",
			"blog/2020/old.html": "<changefreq>weekly</changefreq>",
			"about.html":         "<changefreq>monthly</changefreq>",
		} {
			at := strings.Index(x, loc)
			if at < 0 || !strings.Contains(x[at:min(at+300, len(x))], want) {
				t.Fatalf("run %d: %s should be %s:\n%s", i, loc, want, x)
			}
		}
	}
}
