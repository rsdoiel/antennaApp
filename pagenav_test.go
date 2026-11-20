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

// navSite makes a site with pages "one", "two" and "three" (each with the
// front matter given for it) and generates them through GeneratePages.
func navSite(t *testing.T, fm map[string]string) (htdocs string, eout string) {
	t.Helper()
	cfg := sitemapSite(t, "pages.md")
	cfg.Htdocs = "htdocs"
	cfg.Generator = "page.yaml"
	if err := os.WriteFile("page.yaml", []byte("header: |\n  <h1>Test</h1>\n"), 0664); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll("htdocs/story", 0775); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"one", "two", "three"} {
		src := "---\ntitle: " + name + "\n" + fm[name] + "---\n# " + name + "\n"
		if err := os.WriteFile(name+".md", []byte(src), 0664); err != nil {
			t.Fatal(err)
		}
		addPage(t, "pages.db", name+".html")
	}
	var buf bytes.Buffer
	if err := cfg.GeneratePages(&buf); err != nil {
		t.Fatalf("GeneratePages: %s\n%s", err, buf.String())
	}
	return "htdocs", buf.String()
}

func readNavPage(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("htdocs", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestPagesGetRelLinksAndVisibleNavigation(t *testing.T) {
	_, eout := navSite(t, map[string]string{"two": "urlPrev: one.md\nurlNext: three.md\n"})
	got := readNavPage(t, "two.html")
	head := got[:strings.Index(got, "<body>")]
	for _, want := range []string{`rel="prev"`, `href="one.html"`, `rel="next"`, `href="three.html"`} {
		if !strings.Contains(head, want) {
			t.Errorf("head lacks %s:\n%s", want, head)
		}
	}
	if !regexp.MustCompile(`<link[^>]*rel="prev"[^>]*>`).MatchString(head) {
		t.Errorf("no <link rel=prev> in head:\n%s", head)
	}
	nav := regexp.MustCompile(`(?s)<nav class="page-nav" aria-label="[^"]+">.*?</nav>`).FindString(got)
	if nav == "" {
		t.Fatalf("no labelled page-nav in:\n%s", got)
	}
	for _, want := range []string{`<a rel="prev" href="one.html">`, `<a rel="next" href="three.html">`, "Previous", "Next"} {
		if !strings.Contains(nav, want) {
			t.Errorf("nav lacks %s:\n%s", want, nav)
		}
	}
	// inside <main>, after the page's own content
	if strings.Index(got, `class="page-nav"`) < strings.Index(got, "<main") || strings.Index(got, `class="page-nav"`) > strings.Index(got, "</main>") {
		t.Errorf("page-nav is not inside <main>:\n%s", got)
	}
	if strings.Contains(eout, "warning") {
		t.Errorf("warnings: %s", eout)
	}
	// the nav is not a metadata echo
	if strings.Contains(head, `name="urlPrev"`) || strings.Contains(head, `name="urlNext"`) {
		t.Errorf("urlPrev/urlNext also written as <meta>:\n%s", head)
	}
}

func TestOnlyTheNamedDirectionIsShown(t *testing.T) {
	navSite(t, map[string]string{"one": "urlNext: two.md\n"})
	got := readNavPage(t, "one.html")
	if !strings.Contains(got, `rel="next"`) || strings.Contains(got, `rel="prev"`) {
		t.Errorf("one.html:\n%s", got)
	}
	// a page that names neither has no navigation at all
	if three := readNavPage(t, "three.html"); strings.Contains(three, "page-nav") || strings.Contains(three, `rel="next"`) || strings.Contains(three, `rel="prev"`) {
		t.Errorf("three.html has navigation:\n%s", three)
	}
}

func TestPageNavigationAddressesFollowTheLinkRules(t *testing.T) {
	_, eout := navSite(t, map[string]string{
		"one": "urlNext: https://x.example/a?b=1&c=2\nurlPrev: gone.md\n",
		"two": "urlNext: \"javascript:alert(1)\"\n",
	})
	one := readNavPage(t, "one.html")
	if !strings.Contains(one, `<a rel="next" href="https://x.example/a?b=1&amp;c=2">`) {
		t.Errorf("absolute address not kept and escaped:\n%s", one)
	}
	if !strings.Contains(one, `href="gone.md"`) || !strings.Contains(eout, "gone.md") {
		t.Errorf("unresolved .md link: want it kept and a warning, warnings %q", eout)
	}
	two := readNavPage(t, "two.html")
	if strings.Contains(two, "javascript:") || strings.Contains(two, "page-nav") {
		t.Errorf("script address used:\n%s", two)
	}
	if !strings.Contains(eout, "urlNext") {
		t.Errorf("no warning for the refused address: %q", eout)
	}
}

// Posts are rendered by Post and by GeneratePosts: both show the navigation.
func TestPostsGetNavigation(t *testing.T) {
	cfg := sitemapSite(t, "news.md")
	cfg.Htdocs = "htdocs"
	cfg.Generator = "page.yaml"
	cfg.Collections[0].Generator = "page.yaml"
	if err := os.WriteFile("page.yaml", []byte("header: |\n  <h1>Test</h1>\n"), 0664); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll("htdocs/p", 0775); err != nil {
		t.Fatal(err)
	}
	src := "---\ntitle: Two\npostPath: p/two.md\npubDate: \"2026-11-19\"\nurlPrev: one.md\nurlNext: three.md\n---\nBody.\n"
	if err := os.WriteFile("two.md", []byte(src), 0664); err != nil {
		t.Fatal(err)
	}
	if err := cfg.Post("news.md", "two.md"); err != nil {
		t.Fatalf("Post: %s", err)
	}
	for _, step := range []string{"Post", "GeneratePosts"} {
		if step == "GeneratePosts" {
			os.Remove("htdocs/p/two.html")
			var eout bytes.Buffer
			if err := cfg.Collections[0].GeneratePosts(&eout, "antenna", cfg); err != nil {
				t.Fatalf("GeneratePosts: %s\n%s", err, eout.String())
			}
		}
		b, err := os.ReadFile("htdocs/p/two.html")
		if err != nil {
			t.Fatalf("%s: %s", step, err)
		}
		// one.md and three.md are not rendered by antenna here: addresses stay as written
		page := string(b)
		for _, want := range []string{`<a rel="prev" href="one.md">`, `<a rel="next" href="three.md">`} {
			if !strings.Contains(page, want) {
				t.Errorf("%s: %q not in:\n%s", step, want, page)
			}
		}
		if !regexp.MustCompile(`<link[^>]*rel="prev"[^>]*href="one.md"|<link[^>]*href="one.md"[^>]*rel="prev"`).MatchString(page) {
			t.Errorf("%s: no <link rel=prev> in head:\n%s", step, page)
		}
	}
}
