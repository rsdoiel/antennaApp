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
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// DR-0007: authors write links to the Markdown file, and the HTML output
// links to the HTML page antenna renders from it.

func testTargets() LinkTargets {
	return LinkTargets{
		"index.md":                  "index.html",
		"about.md":                  "about.html",
		"README.md":                 "README.html",
		"docs/guide.md":             "docs/guide.html",
		"socal_north.md":            "socal_north.html",
		"socal_north_items.md":      "socal_north.html",
		"blog/2026/06/30/a-post.md": "blog/2026/06/30/a-post.html",
	}
}

func TestRewriteDestination(t *testing.T) {
	cases := []struct {
		name, from, dest, want, warn string
	}{
		{"sibling page", "index.md", "about.md", "about.html", ""},
		{"fragment kept", "index.md", "about.md#team", "about.html#team", ""},
		{"query kept", "index.md", "about.md?x=1", "about.html?x=1", ""},
		{"extension in any case", "index.md", "about.MD", "about.html", ""},
		{"page in a directory", "index.md", "docs/guide.md", "docs/guide.html", ""},
		{"parent directory", "docs/guide.md", "../about.md", "../about.html", ""},
		{"same directory from a subdirectory", "docs/other.md", "guide.md", "guide.html", ""},
		{"site-root path", "docs/guide.md", "/about.md", "/about.html", ""},
		{"collection source means the collection page", "index.md", "socal_north.md", "socal_north.html", ""},
		{"collection items file is the same page", "index.md", "socal_north_items.md", "socal_north.html", ""},
		{"post", "index.md", "blog/2026/06/30/a-post.md", "blog/2026/06/30/a-post.html", ""},
		{"dot slash", "index.md", "./about.md", "about.html", ""},
		{"absolute URL", "index.md", "https://example.org/x.md", "https://example.org/x.md", ""},
		{"scheme-relative URL", "index.md", "//example.org/x.md", "//example.org/x.md", ""},
		{"mailto", "index.md", "mailto:a@example.org", "mailto:a@example.org", ""},
		{"not Markdown", "index.md", "feed.xml", "feed.xml", ""},
		{"already html", "index.md", "about.html", "about.html", ""},
		{"fragment only", "index.md", "#top", "#top", ""},
		{"empty", "index.md", "", "", ""},
		{"unrendered target stays and warns", "index.md", "missing.md", "missing.md", "missing.md"},
		{"unrendered target keeps its fragment", "index.md", "missing.md#x", "missing.md#x", "missing.md"},
		{"outside the site", "index.md", "../../etc.md", "../../etc.md", "etc.md"},
	}
	for _, c := range cases {
		got, warn := RewriteDestination(c.from, c.dest, testTargets())
		if got != c.want {
			t.Errorf("%s: %q from %q = %q, want %q", c.name, c.dest, c.from, got, c.want)
		}
		if (c.warn == "") != (warn == "") || !strings.Contains(warn, c.warn) {
			t.Errorf("%s: warning %q, want one containing %q", c.name, warn, c.warn)
		}
	}
}

const rewriteSample = "See [about](about.md#team), [guide](docs/guide.md), [feed](feed.xml),\n" +
	"[outside](https://example.org/x.md) and [lost](missing.md).\n\n" +
	"A reference [link][r] and a code span `[c](about.md)`.\n\n" +
	"```\n[fenced](about.md)\n```\n\n" +
	"![picture](about.md)\n\n[r]: index.md\n"

func renderWith(t *testing.T, unsafe bool) (string, string) {
	t.Helper()
	var warnings bytes.Buffer
	doc := &CommonMark{Text: rewriteSample, Links: &LinkRewriter{From: "x.md", Targets: testTargets(), Warn: &warnings}}
	var html string
	var err error
	if unsafe {
		html, err = doc.ToUnsafeHTML()
	} else {
		html, err = doc.ToHTML()
	}
	if err != nil {
		t.Fatal(err)
	}
	return html, warnings.String()
}

func TestRenderRewritesLinksInBothRenderers(t *testing.T) {
	for _, unsafe := range []bool{false, true} {
		html, warnings := renderWith(t, unsafe)
		for _, want := range []string{
			`href="about.html#team"`, `href="docs/guide.html"`, `href="feed.xml"`,
			`href="https://example.org/x.md"`, `href="missing.md"`, `href="index.html"`,
		} {
			if !strings.Contains(html, want) {
				t.Errorf("unsafe=%v: no %s in\n%s", unsafe, want, html)
			}
		}
		if !strings.Contains(warnings, "missing.md") || strings.Count(warnings, "\n") != 1 {
			t.Errorf("unsafe=%v: warnings %q, want one line about missing.md", unsafe, warnings)
		}
	}
}

func TestRenderLeavesCodeAndImagesAlone(t *testing.T) {
	html, _ := renderWith(t, true)
	if !strings.Contains(html, "[c](about.md)") || !strings.Contains(html, "[fenced](about.md)") {
		t.Errorf("code was changed:\n%s", html)
	}
	if !strings.Contains(html, `src="about.md"`) {
		t.Errorf("image source was changed:\n%s", html)
	}
}

func TestRenderWithoutARewriterChangesNothing(t *testing.T) {
	doc := &CommonMark{Text: "[a](about.md)\n"}
	html, err := doc.ToHTML()
	if err != nil || !strings.Contains(html, `href="about.md"`) {
		t.Errorf("%v\n%s", err, html)
	}
}

// What antenna renders is known from its own data: the pages table, the
// collections in antenna.yaml and the posts (DR-0007 5, 6).
func TestLinkTargetsComeFromThePagesCollectionsAndPosts(t *testing.T) {
	cfg := sitemapSite(t, "pages.md", "socal_north.md", "columns.md")
	addPage(t, "pages.db", "index.html")
	addPage(t, "pages.db", "docs/guide.html")
	addPost(t, "columns.db", "blog/2026/06/30/a-post.md")
	targets, err := cfg.LinkTargets()
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"index.md":                  "index.html",
		"docs/guide.md":             "docs/guide.html",
		"socal_north.md":            "socal_north.html",
		"socal_north_items.md":      "socal_north.html",
		"columns.md":                "columns.html",
		"columns_items.md":          "columns.html",
		"blog/2026/06/30/a-post.md": "blog/2026/06/30/a-post.html",
	}
	for k, v := range want {
		if targets[k] != v {
			t.Errorf("targets[%q] = %q, want %q", k, targets[k], v)
		}
	}
	if _, ok := targets["pages.md"]; ok {
		t.Error("the pages collection's own file is not a page of the site")
	}
	if _, ok := targets["pages_items.md"]; ok {
		t.Error("the pages collection has no items page")
	}
}

func TestLinkTargetsToleratesACollectionWithoutAPagesTable(t *testing.T) {
	cfg := sitemapSite(t, "pages.md", "old.md")
	sitemapExec(t, "old.db", `DROP TABLE pages`)
	if _, err := cfg.LinkTargets(); err != nil {
		t.Errorf("a database with no pages table is not an error: %s", err)
	}
}

// The real path: GeneratePages renders a page from disk, and the HTML file
// links to the HTML page, and says so on stderr for a link it cannot map.
func TestGeneratePagesRewritesMarkdownLinks(t *testing.T) {
	cfg := sitemapSite(t, "pages.md")
	cfg.Htdocs = "htdocs"
	cfg.Generator = "page.yaml"
	if err := os.WriteFile("page.yaml", []byte("header: |\n  <h1>Test</h1>\n"), 0664); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cfg.Htdocs, 0775); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("index.md", []byte("---\ntitle: Home\n---\n# Home\n\n[About](about.md) and [Gone](gone.md).\n"), 0664); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("about.md", []byte("---\ntitle: About\n---\n# About\n\n[Home](index.md)\n"), 0664); err != nil {
		t.Fatal(err)
	}
	addPage(t, "pages.db", "index.html")
	addPage(t, "pages.db", "about.html")
	var eout bytes.Buffer
	if err := cfg.GeneratePages(&eout); err != nil {
		t.Fatalf("GeneratePages: %s\n%s", err, eout.String())
	}
	index, err := os.ReadFile(filepath.Join("htdocs", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(index), `href="about.html"`) || !strings.Contains(string(index), `href="gone.md"`) {
		t.Errorf("index.html links:\n%s", index)
	}
	about, _ := os.ReadFile(filepath.Join("htdocs", "about.html"))
	if !strings.Contains(string(about), `href="index.html"`) {
		t.Errorf("about.html links:\n%s", about)
	}
	if !strings.Contains(eout.String(), "gone.md") {
		t.Errorf("no warning for gone.md: %q", eout.String())
	}
}

// A directory key (ending in "/") in LinkTargets stands for a whole subsite
// (antennaApp DR-0027 2): a .md link under it becomes its .html twin, with no
// lookup and no warning.
func TestRewriteDestinationIntoASubsite(t *testing.T) {
	targets := testTargets()
	targets["antenna/"] = ""
	cases := []struct {
		name, from, dest, want, warn string
	}{
		{"site-root path", "index.md", "/antenna/index.md", "/antenna/index.html", ""},
		{"fragment kept", "index.md", "/antenna/about.md#x", "/antenna/about.html#x", ""},
		{"nested page", "index.md", "/antenna/a/b.md", "/antenna/a/b.html", ""},
		{"relative path", "index.md", "antenna/index.md", "antenna/index.html", ""},
		{"relative from a subdirectory", "docs/guide.md", "../antenna/index.md", "../antenna/index.html", ""},
		{"the directory itself is untouched", "index.md", "/antenna/", "/antenna/", ""},
		{"a sibling with the same prefix is not the subsite", "index.md", "/antenna2/x.md", "/antenna2/x.md", "antenna2/x.md"},
		{"outside every subsite still warns", "index.md", "missing.md", "missing.md", "missing.md"},
		{"a page of the site still wins", "index.md", "about.md", "about.html", ""},
	}
	for _, c := range cases {
		got, warn := RewriteDestination(c.from, c.dest, targets)
		if got != c.want {
			t.Errorf("%s: %q from %q = %q, want %q", c.name, c.dest, c.from, got, c.want)
		}
		if (c.warn == "") != (warn == "") || !strings.Contains(warn, c.warn) {
			t.Errorf("%s: warning %q, want one containing %q", c.name, warn, c.warn)
		}
	}
}

// Subsites on the site's own host, with a path, become directory targets. A
// subsite on another host, or at the root of the host, does not.
func TestLinkTargetsHoldSameHostSubsiteDirectories(t *testing.T) {
	cfg := sitemapSite(t, "pages.md")
	cfg.BaseURL = "https://example.org"
	cfg.Subsites = []Subsite{
		{URL: "https://example.org/antenna"},
		{URL: "https://example.org/deep/er/"},
		{URL: "https://other.example/blog"},
		{URL: "https://example.org"},
	}
	targets, err := cfg.LinkTargets()
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"antenna/", "deep/er/"} {
		if _, ok := targets[k]; !ok {
			t.Errorf("no directory target %q in %v", k, targets)
		}
	}
	for k := range targets {
		if strings.HasSuffix(k, "/") && k != "antenna/" && k != "deep/er/" {
			t.Errorf("unexpected directory target %q", k)
		}
	}
}

// The real path: GeneratePages writes the subsite link as .html without a
// warning, and a same-host .md link outside every subsite still warns.
func TestGeneratePagesRewritesSubsiteLinks(t *testing.T) {
	cfg := sitemapSite(t, "pages.md")
	cfg.Htdocs = "htdocs"
	cfg.Generator = "page.yaml"
	cfg.BaseURL = "https://example.org"
	cfg.Subsites = []Subsite{{URL: "https://example.org/antenna"}}
	if err := os.WriteFile("page.yaml", []byte("header: |\n  <h1>Test</h1>\n"), 0664); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cfg.Htdocs, 0775); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("index.md", []byte("---\ntitle: Home\n---\n# Home\n\n[Antenna](/antenna/index.md) and [Elsewhere](/other/x.md).\n"), 0664); err != nil {
		t.Fatal(err)
	}
	addPage(t, "pages.db", "index.html")
	var eout bytes.Buffer
	if err := cfg.GeneratePages(&eout); err != nil {
		t.Fatalf("GeneratePages: %s\n%s", err, eout.String())
	}
	index, _ := os.ReadFile(filepath.Join("htdocs", "index.html"))
	if !strings.Contains(string(index), `href="/antenna/index.html"`) || !strings.Contains(string(index), `href="/other/x.md"`) {
		t.Errorf("index.html links:\n%s", index)
	}
	if strings.Contains(eout.String(), "/antenna/index.md") {
		t.Errorf("warning for a subsite link: %q", eout.String())
	}
	if !strings.Contains(eout.String(), "/other/x.md") {
		t.Errorf("no warning for the link outside the subsite: %q", eout.String())
	}
}
