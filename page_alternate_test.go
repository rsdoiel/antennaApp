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
	"regexp"
	"strings"
	"testing"
)

// DR-0023: each page declares its own Markdown alternate.

var linkTag = regexp.MustCompile(`<link[^>]*>`)

// markdownAlternates returns the hrefs of the <link rel="alternate"
// type="text/markdown"> elements in html, and the number of alternates that
// use the wrong media type.
func markdownAlternates(html string) (hrefs []string, wrongType int) {
	for _, tag := range linkTag.FindAllString(html, -1) {
		if !strings.Contains(tag, `rel="alternate"`) && !strings.Contains(tag, `rel="altenate"`) {
			continue
		}
		href := regexp.MustCompile(`href="([^"]*)"`).FindStringSubmatch(tag)
		switch {
		case strings.Contains(tag, `type="text/markdown"`) && href != nil:
			hrefs = append(hrefs, href[1])
		case strings.Contains(tag, `type="application/markdown"`):
			wrongType++
		}
	}
	return hrefs, wrongType
}

func TestPageFromATextSourceDeclaresItsMarkdownCopy(t *testing.T) {
	markdownPageSite(t)
	if err := runIn(t, "page", "index.txt", "index.md", "index.html"); err != nil {
		t.Fatal(err)
	}
	hrefs, wrong := markdownAlternates(readText(t, "index.html"))
	if len(hrefs) != 1 || hrefs[0] != "index.md" || wrong != 0 {
		t.Errorf("index.html should declare index.md once, text/markdown: %v wrong=%d\n%s", hrefs, wrong, readText(t, "index.html"))
	}
}

// A page whose source is Markdown already declared itself; keep that.
func TestPageFromMarkdownStillDeclaresItself(t *testing.T) {
	markdownPageSite(t)
	hrefs, _ := markdownAlternates(readText(t, "about.html"))
	if len(hrefs) != 1 || hrefs[0] != "about.md" {
		t.Errorf("about.html should declare about.md: %v", hrefs)
	}
}

// A .txt source with no Markdown copy has nothing to declare: the .txt is not
// Markdown a reader can use.
func TestPageWithNoMarkdownCopyDeclaresNothing(t *testing.T) {
	markdownPageSite(t)
	if err := runIn(t, "page", "index.txt", "index.html"); err != nil {
		t.Fatal(err)
	}
	if hrefs, _ := markdownAlternates(readText(t, "index.html")); len(hrefs) != 0 {
		t.Errorf("no Markdown copy, no alternate: %v", hrefs)
	}
}

// DR-0024 4: the row records the copy, so generate knows it, and a copy in
// another folder is declared relative to the HTML (DR-0023 1 declared only a
// copy beside the page because generate could not know better).
func TestPageDeclaresAMarkdownCopyInAnotherFolderRelatively(t *testing.T) {
	markdownPageSite(t)
	os.MkdirAll("site/docs", 0o775)
	if err := runIn(t, "page", "index.txt", "site/docs/home.md", "site/docs/home.html"); err != nil {
		t.Fatal(err)
	}
	hrefs, _ := markdownAlternates(readText(t, "site/docs/home.html"))
	if len(hrefs) != 1 || hrefs[0] != "home.md" {
		t.Errorf("same folder, base name: %v", hrefs)
	}
	if err := runIn(t, "page", "index.txt", "site/other.md", "site/docs/other.html"); err != nil {
		t.Fatal(err)
	}
	hrefs, _ = markdownAlternates(readText(t, "site/docs/other.html"))
	if len(hrefs) != 1 || hrefs[0] != "../other.md" {
		t.Errorf("another folder, relative: %v", hrefs)
	}
}

// generate re-renders pages from the pages table: the link must survive.
func TestGenerateKeepsThePageAlternate(t *testing.T) {
	markdownPageSite(t)
	if err := runIn(t, "page", "index.txt", "index.md", "index.html"); err != nil {
		t.Fatal(err)
	}
	if err := runIn(t, "generate"); err != nil {
		t.Fatal(err)
	}
	if hrefs, _ := markdownAlternates(readText(t, "index.html")); len(hrefs) != 1 || hrefs[0] != "index.md" {
		t.Errorf("generate dropped the alternate: %v", hrefs)
	}
}

func TestPageAlternateIsStable(t *testing.T) {
	markdownPageSite(t)
	if err := runIn(t, "page", "index.txt", "index.md", "index.html"); err != nil {
		t.Fatal(err)
	}
	first := readText(t, "index.html")
	if err := runIn(t, "page", "index.txt", "index.md", "index.html"); err != nil {
		t.Fatal(err)
	}
	if again := readText(t, "index.html"); again != first {
		t.Errorf("a second run changed the page")
	}
}

// validate: a Markdown alternate written in the generator YAML is the same for
// every page.
func TestValidateStaticMarkdownAlternateInTheGeneratorYAML(t *testing.T) {
	for _, typ := range []string{"application/markdown", "text/markdown"} {
		bodySite(t)
		b, err := os.ReadFile("page.yaml")
		if err != nil {
			t.Fatal(err)
		}
		os.WriteFile("page.yaml", append(b, []byte("\nlink:\n  - rel: alternate\n    title: Markdown source for page\n    type: "+typ+"\n    href: index.md\n")...), 0o664)
		out, _ := runValidate(t)
		line, ok := finding(out, "page.yaml", "warning", "alternate-static")
		if !ok || !strings.Contains(line, "index.md") {
			t.Errorf("%s: want alternate-static naming the entry: %q\n%s", typ, line, out)
		}
	}
}

func TestValidateAGeneratorYAMLWithNoMarkdownAlternateIsQuiet(t *testing.T) {
	bodySite(t)
	b, _ := os.ReadFile("page.yaml")
	os.WriteFile("page.yaml", append(b, []byte("\nlink:\n  - rel: alternate\n    type: application/rss+xml\n    href: index.xml\n  - rel: search\n    type: application/opensearchdescription+xml\n    href: osd.xml\n")...), 0o664)
	out, _ := runValidate(t)
	if strings.Contains(out, "alternate-static") {
		t.Errorf("an RSS alternate is fine:\n%s", out)
	}
}

// The table of contents lists a page by its Markdown. A page whose source is
// bio.md and whose HTML is about.html has its Markdown at bio.md; the toc
// looked for about.md and warned "page about.html is not listed: its Markdown
// about.md is not published" (rsdoiel.github.io, 2026-11-20).
func TestTOCListsAPageByItsMarkdownSourceNotItsHTMLName(t *testing.T) {
	newWorkspace(t, true)
	os.WriteFile("bio.md", []byte("---\ntitle: Bio\n---\n# Bio\n"), 0o664)
	if err := runIn(t, "page", "bio.md", "about.html"); err != nil {
		t.Fatal(err)
	}
	var eout bytes.Buffer
	if err := runInBuf(t, &eout, "generate"); err != nil {
		t.Fatalf("generate: %s\n%s", err, eout.String())
	}
	if strings.Contains(eout.String(), "not listed") {
		t.Errorf("the page has a published Markdown, bio.md:\n%s", eout.String())
	}
	toc := readText(t, ".well-known/toc.opml")
	if !strings.Contains(toc, "bio.md") || strings.Contains(toc, "about.md") {
		t.Errorf("toc.opml should list bio.md:\n%s", toc)
	}
}
