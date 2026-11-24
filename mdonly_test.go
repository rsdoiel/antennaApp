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
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// DR-0031: a collection with html: false writes Markdown, feeds and the
// table of contents, and no HTML. Posts go to the pages collection by
// default, so that is the collection these tests switch.

const mdOnlyPost = "blog/2026/11/19/one.md"

// setCollectionLines adds lines to the pages collection entry in antenna.yaml,
// right after its `file:` line.
func setCollectionLines(t *testing.T, lines ...string) {
	t.Helper()
	b, err := os.ReadFile("antenna.yaml")
	if err != nil {
		t.Fatal(err)
	}
	add := ""
	for _, l := range lines {
		add += "      " + l + "\n"
	}
	s := strings.Replace(string(b), "      file: pages.md\n", "      file: pages.md\n"+add, 1)
	if s == string(b) {
		t.Fatalf("no pages.md collection entry in antenna.yaml:\n%s", b)
	}
	if err := os.WriteFile("antenna.yaml", []byte(s), 0o664); err != nil {
		t.Fatal(err)
	}
}

// postOne writes and posts a post named by path.
func postOne(t *testing.T, p, title string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o775); err != nil {
		t.Fatal(err)
	}
	src := "---\ntitle: " + title + "\npostPath: " + p + "\npubDate: \"2026-11-19\"\n---\nBody of " + title + ".\n"
	if err := os.WriteFile(p, []byte(src), 0o664); err != nil {
		t.Fatal(err)
	}
	if err := runIn(t, "post", p); err != nil {
		t.Fatalf("post %s: %s", p, err)
	}
}

// itemLinks returns link and guid of every row of pages.db with a postPath
// (a post's guid is its link unless its front matter sets one).
func itemLinks(t *testing.T) map[string][2]string {
	t.Helper()
	db, err := sql.Open("sqlite", "pages.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows, err := db.Query(`SELECT postPath, link, guid FROM items WHERE postPath != ''`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	out := map[string][2]string{}
	for rows.Next() {
		var p, l, g string
		if err := rows.Scan(&p, &l, &g); err != nil {
			t.Fatal(err)
		}
		out[p] = [2]string{l, g}
	}
	return out
}

func TestHTMLSettingLoadsFromAConfigurationFile(t *testing.T) {
	newWorkspace(t, true)
	cfg := &AppConfig{}
	if err := cfg.LoadConfig("antenna.yaml"); err != nil {
		t.Fatal(err)
	}
	if cfg.Collections[0].HTML != nil || !cfg.Collections[0].WantsHTML() {
		t.Errorf("an unset html must mean HTML is written")
	}
	setCollectionLines(t, "html: false")
	cfg = &AppConfig{}
	if err := cfg.LoadConfig("antenna.yaml"); err != nil {
		t.Fatal(err)
	}
	if col := cfg.Collections[0]; col.HTML == nil || col.WantsHTML() {
		t.Errorf("html: false must turn HTML off")
	}
}

func TestHTMLFalseWithNothingToShowIsAConfigurationError(t *testing.T) {
	for name, lines := range map[string][]string{
		"markdownItems false": {"html: false", "markdownItems: false"},
		"page-index":          {"html: false", "mode: page-index"},
	} {
		newWorkspace(t, true)
		setCollectionLines(t, lines...)
		wantClass(t, name, runIn(t, "generate"), classConfig)
	}
}

func TestMarkdownOnlyGenerateWritesNoHTML(t *testing.T) {
	newWorkspace(t, true)
	setCollectionLines(t, "html: false")
	postOne(t, mdOnlyPost, "One")
	if err := runIn(t, "generate"); err != nil {
		t.Fatalf("generate: %s", err)
	}
	for _, gone := range []string{"pages.html", "blog/2026/11/19/one.html"} {
		if exists(gone) {
			t.Errorf("%s was written for a Markdown-only collection", gone)
		}
	}
	// The feed-list OPML is written only for a collection with feeds, as before;
	// this one has only posts.
	for _, want := range []string{"pages_items.md", "pages.xml", mdOnlyPost} {
		if !exists(want) {
			t.Errorf("%s was not written", want)
		}
	}
	if !strings.Contains(readText(t, "pages_items.md"), "One") {
		t.Errorf("the items file does not show the post:\n%s", readText(t, "pages_items.md"))
	}
}

func TestACollectionWithoutTheSettingStillWritesHTML(t *testing.T) {
	newWorkspace(t, true)
	postOne(t, mdOnlyPost, "One")
	if err := runIn(t, "generate"); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"pages.html", "blog/2026/11/19/one.html"} {
		if !exists(want) {
			t.Errorf("%s missing from an ordinary collection", want)
		}
	}
	if got := itemLinks(t)[mdOnlyPost][0]; got != "http://localhost:8000/blog/2026/11/19/one.html" {
		t.Errorf("link %q, want the .html address as before", got)
	}
}

func TestPostIntoMarkdownOnlyCollectionStoresTheMarkdownAddress(t *testing.T) {
	newWorkspace(t, true)
	setCollectionLines(t, "html: false")
	postOne(t, mdOnlyPost, "One")
	if exists("blog/2026/11/19/one.html") {
		t.Error("post wrote an HTML page for a Markdown-only collection")
	}
	got := itemLinks(t)[mdOnlyPost]
	want := "http://localhost:8000/blog/2026/11/19/one.md"
	if got[0] != want || got[1] != want {
		t.Errorf("link %q guid %q, want both %q", got[0], got[1], want)
	}
}

// Switching an existing collection: the collection page this run knows is
// removed, post HTML is left and counted, earlier posts keep their identity.
func TestSwitchingACollectionToMarkdownOnly(t *testing.T) {
	newWorkspace(t, true)
	postOne(t, mdOnlyPost, "One")
	if err := runIn(t, "generate"); err != nil {
		t.Fatal(err)
	}
	if !exists("pages.html") || !exists("blog/2026/11/19/one.html") {
		t.Fatal("setup: the HTML site was not written")
	}
	oldLink := itemLinks(t)[mdOnlyPost]
	setCollectionLines(t, "html: false")
	postOne(t, "blog/2026/11/20/two.md", "Two")
	_, eout, err := runCapture(t, "generate")
	if err != nil {
		t.Fatalf("generate: %s", err)
	}
	if exists("pages.html") {
		t.Error("pages.html was left behind")
	}
	if !exists("blog/2026/11/19/one.html") {
		t.Error("an earlier post's HTML was deleted")
	}
	if exists("blog/2026/11/20/two.html") {
		t.Error("a post made after the switch has HTML")
	}
	if !strings.Contains(eout, "1 post") || !strings.Contains(eout, "one.html") {
		t.Errorf("no warning counting the HTML left (1 post, naming one.html):\n%s", eout)
	}
	links := itemLinks(t)
	if links[mdOnlyPost] != oldLink {
		t.Errorf("an earlier post's identity changed: %v -> %v", oldLink, links[mdOnlyPost])
	}
	if got := links["blog/2026/11/20/two.md"][0]; !strings.HasSuffix(got, "/two.md") {
		t.Errorf("a post after the switch has link %q, want its .md address", got)
	}
}

func TestMarkdownOnlyPostsAreListedInTheSitemapByTheirMarkdownAddress(t *testing.T) {
	newWorkspace(t, true)
	setCollectionLines(t, "html: false")
	postOne(t, mdOnlyPost, "One")
	if err := runIn(t, "sitemap"); err != nil {
		t.Fatal(err)
	}
	matches, _ := filepath.Glob("sitemap*.xml")
	all := ""
	for _, m := range matches {
		all += readText(t, m)
	}
	if !strings.Contains(all, "/blog/2026/11/19/one.md</loc>") || strings.Contains(all, "one.html") {
		t.Errorf("the sitemap does not list the .md address alone:\n%s", all)
	}
}

// A link to a post of a Markdown-only collection is left as Markdown: there is
// no HTML page for it to become.
func TestLinkToAMarkdownOnlyPostStaysMarkdown(t *testing.T) {
	newWorkspace(t, true)
	setCollectionLines(t, "html: false")
	postOne(t, mdOnlyPost, "One")
	os.WriteFile("about.md", []byte("---\ntitle: About\n---\n\nSee [the post]("+mdOnlyPost+").\n"), 0o664)
	if err := runIn(t, "page", "about.md"); err != nil {
		t.Fatal(err)
	}
	html := readText(t, "about.html")
	if !strings.Contains(html, `href="`+mdOnlyPost+`"`) || strings.Contains(html, "one.html") {
		t.Errorf("the link was rewritten to a page that does not exist:\n%s", html)
	}
}

func TestValidateWarnsThatAMarkdownOnlyCollectionHasATheme(t *testing.T) {
	newWorkspace(t, true)
	setCollectionLines(t, "html: false")
	out, eout, _ := runCapture(t, "validate")
	all := out + eout
	if !strings.Contains(all, "markdown-only-generator") {
		t.Errorf("no warning that the generator is unused:\n%s", all)
	}
	newWorkspace(t, true)
	out, eout, _ = runCapture(t, "validate")
	if strings.Contains(out+eout, "markdown-only-generator") {
		t.Errorf("an ordinary collection drew the warning:\n%s", out+eout)
	}
}

// A post antenna produces gets a guid when it has none: its link, which is
// unique (the item key) and a URL. Set when the post is made, whether or not the
// collection writes HTML.
func TestPostWithNoGuidIsStoredWithItsLinkAsGuid(t *testing.T) {
	for name, md := range map[string]bool{"HTML collection": false, "Markdown-only collection": true} {
		newWorkspace(t, true)
		if md {
			setCollectionLines(t, "html: false")
		}
		postOne(t, mdOnlyPost, "One")
		got := itemLinks(t)[mdOnlyPost]
		if got[0] == "" || got[1] != got[0] {
			t.Errorf("%s: link %q guid %q, want the guid to be the link", name, got[0], got[1])
		}
	}
}

func TestPostKeepsAGuidFromItsFrontMatter(t *testing.T) {
	newWorkspace(t, true)
	os.MkdirAll("blog/2026/11/19", 0o775)
	os.WriteFile(mdOnlyPost, []byte("---\ntitle: One\npostPath: "+mdOnlyPost+"\npubDate: \"2026-11-19\"\nguid: my-own-guid\n---\nBody.\n"), 0o664)
	if err := runIn(t, "post", mdOnlyPost); err != nil {
		t.Fatal(err)
	}
	if got := itemLinks(t)[mdOnlyPost]; got[1] != "my-own-guid" {
		t.Errorf("guid %q, want the author's own", got[1])
	}
}
