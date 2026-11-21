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
	"os"
	"strings"
	"testing"
)

// DR-0019: antenna page SOURCE [OUTPUT...] writes each output the extension
// names, so index.txt with include directives gives both index.html and the
// expanded index.md.

// markdownPageSite is a site the way rsdoiel.github.io has one: index.txt
// holds the directives, blog/recent.txt is a fragment, about.md is a page
// that other pages link to.
func markdownPageSite(t *testing.T) {
	t.Helper()
	newWorkspace(t, true)
	os.MkdirAll("blog", 0o775)
	files := map[string]string{
		"index.txt": "---\ntitle: Home\n---\n\n@include-text-block author.md\n\n## Recent\n\n@include-text-block blog/recent.txt\n\n" +
			"[About](about.md)\n\n@include-code-block hello.go go\n",
		"author.md":       "I am the author.\n",
		"blog/recent.txt": "- [Post one](/blog/one.md)\n",
		"hello.go":        "package main\n",
		"about.md":        "---\ntitle: About\n---\nAbout.\n",
	}
	for name, body := range files {
		if err := os.WriteFile(name, []byte(body), 0o664); err != nil {
			t.Fatal(err)
		}
	}
	// about.md is a page antenna renders, so a link to it is rewritten in HTML.
	if err := runIn(t, "page", "about.md", "about.html"); err != nil {
		t.Fatalf("page about: %s", err)
	}
}

func readText(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("%s: %s", name, err)
	}
	return string(b)
}

// The reported bug: the served index.md held the literal directives.
func TestPageWritesExpandedMarkdown(t *testing.T) {
	markdownPageSite(t)
	if err := runIn(t, "page", "index.txt", "index.md", "index.html"); err != nil {
		t.Fatalf("page: %s", err)
	}
	md := readText(t, "index.md")
	if strings.Contains(md, "@include-") {
		t.Errorf("index.md still holds a directive:\n%s", md)
	}
	for _, want := range []string{"I am the author.", "- [Post one](/blog/one.md)", "package main", "title: Home", "[About](about.md)"} {
		if !strings.Contains(md, want) {
			t.Errorf("index.md has no %q:\n%s", want, md)
		}
	}
	html := readText(t, "index.html")
	if !strings.Contains(html, "I am the author.") {
		t.Errorf("index.html has no included text:\n%s", html)
	}
	if !strings.Contains(html, `href="about.html"`) {
		t.Errorf("index.html should link to about.html:\n%s", html)
	}
	if strings.Contains(html, "@include-") {
		t.Errorf("index.html still holds a directive")
	}
}

func TestPageMarkdownLeavesTheSourceAlone(t *testing.T) {
	markdownPageSite(t)
	before := readText(t, "index.txt")
	if err := runIn(t, "page", "index.txt", "index.md", "index.html"); err != nil {
		t.Fatalf("page: %s", err)
	}
	if after := readText(t, "index.txt"); after != before {
		t.Errorf("the source was changed:\n%s", after)
	}
}

func TestPageMarkdownIsStable(t *testing.T) {
	markdownPageSite(t)
	if err := runIn(t, "page", "index.txt", "index.md", "index.html"); err != nil {
		t.Fatal(err)
	}
	first := readText(t, "index.md")
	if err := runIn(t, "page", "index.txt", "index.md", "index.html"); err != nil {
		t.Fatal(err)
	}
	if second := readText(t, "index.md"); second != first {
		t.Errorf("a second run changed index.md")
	}
}

// Order of the outputs does not matter, and one output is enough.
func TestPageMarkdownOnly(t *testing.T) {
	markdownPageSite(t)
	if err := runIn(t, "page", "index.txt", "index.md"); err != nil {
		t.Fatalf("page: %s", err)
	}
	if _, err := os.Stat("index.md"); err != nil {
		t.Errorf("index.md not written: %s", err)
	}
	if _, err := os.Stat("index.html"); err == nil {
		t.Errorf("index.html was written but not asked for")
	}
	// A Markdown only run is not a page: no pages.db record (DR-0019 4).
	cfg := &AppConfig{}
	if err := cfg.LoadConfig("antenna.yaml"); err != nil {
		t.Fatal(err)
	}
	pages, err := cfg.GetPages()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range pages {
		if p["inputPath"] == "index.txt" {
			t.Errorf("index.txt was recorded as a page: %v", p)
		}
	}
}

func TestPageHTMLOnlyStillWorks(t *testing.T) {
	markdownPageSite(t)
	if err := runIn(t, "page", "index.txt", "index.html"); err != nil {
		t.Fatalf("page: %s", err)
	}
	if _, err := os.Stat("index.md"); err == nil {
		t.Errorf("index.md was written but not asked for")
	}
}

func TestPageOutputEqualToSourceIsRefused(t *testing.T) {
	markdownPageSite(t)
	before := readText(t, "index.txt")
	wantClass(t, "output is the source", runIn(t, "page", "index.txt", "index.txt"), classData)
	wantClass(t, "markdown output is the source", runIn(t, "page", "about.md", "about.md", "about.html"), classData)
	if after := readText(t, "index.txt"); after != before {
		t.Errorf("the source was changed")
	}
	if got := readText(t, "about.md"); !strings.Contains(got, "About.") {
		t.Errorf("about.md was overwritten: %q", got)
	}
}

func TestPageUnknownOutputExtension(t *testing.T) {
	markdownPageSite(t)
	wantClass(t, "page .pdf", runIn(t, "page", "index.txt", "index.pdf"), classUsage)
	wantClass(t, "page one good one bad", runIn(t, "page", "index.txt", "index.md", "index.docx"), classUsage)
	if _, err := os.Stat("index.md"); err == nil {
		t.Errorf("nothing should be written when an output is not valid")
	}
}

// A page made from a .txt source also has a Markdown copy; HTML that links to
// that copy must still land on the page (DR-0007), and the derived previous
// and next links must too. The failure seen on rsdoiel.github.io: cv.html got
// <a rel="prev" href="blog/index.md"> once blog/index.txt was the source.
func TestLinkToTheMarkdownCopyOfATextSourcePageBecomesHTML(t *testing.T) {
	markdownPageSite(t)
	if err := runIn(t, "page", "index.txt", "index.md", "index.html"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("other.md", []byte("---\ntitle: Other\n---\n[Home](index.md)\n"), 0o664); err != nil {
		t.Fatal(err)
	}
	if err := runIn(t, "page", "other.md", "other.html"); err != nil {
		t.Fatal(err)
	}
	html := readText(t, "other.html")
	if !strings.Contains(html, `href="index.html"`) || strings.Contains(html, `href="index.md"`) {
		t.Errorf("a link to the Markdown copy should become index.html:\n%s", html)
	}
}
