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
	"strings"
	"testing"
)

// DR-0021: the site-wide author in antenna.yaml is the default author of what
// antenna writes for the site's own pages and posts. An author in a
// document's front matter wins; the author's source file is never changed;
// harvested items never get it.

func frontMatterOf(t *testing.T, name string) string {
	t.Helper()
	txt := readText(t, name)
	if !strings.HasPrefix(txt, "---\n") {
		return ""
	}
	end := strings.Index(txt[4:], "\n---\n")
	if end < 0 {
		return ""
	}
	return txt[4 : 4+end]
}

func TestPageMarkdownGetsTheSiteAuthor(t *testing.T) {
	markdownPageSite(t)
	setConfig(t, "author: R. S. Doiel")
	if err := runIn(t, "page", "index.txt", "index.md", "index.html"); err != nil {
		t.Fatal(err)
	}
	fm := frontMatterOf(t, "index.md")
	if !strings.Contains(fm, "author: R. S. Doiel") || !strings.Contains(fm, "title: Home") {
		t.Errorf("index.md front matter should hold the site author and keep the title:\n%s", fm)
	}
	if strings.Contains(readText(t, "index.txt"), "author:") {
		t.Errorf("the source must not be changed")
	}
}

func TestPageMarkdownKeepsTheDocumentsOwnAuthor(t *testing.T) {
	markdownPageSite(t)
	setConfig(t, "author: R. S. Doiel")
	os.WriteFile("guest.txt", []byte("---\ntitle: Guest\nauthor: A. Guest\n---\nBody.\n"), 0o664)
	if err := runIn(t, "page", "guest.txt", "guest.md"); err != nil {
		t.Fatal(err)
	}
	fm := frontMatterOf(t, "guest.md")
	if !strings.Contains(fm, "author: A. Guest") || strings.Contains(fm, "R. S. Doiel") {
		t.Errorf("the document's author wins:\n%s", fm)
	}
	// A blank author is no author.
	os.WriteFile("blank.txt", []byte("---\ntitle: Blank\nauthor: \"  \"\n---\nBody.\n"), 0o664)
	if err := runIn(t, "page", "blank.txt", "blank.md"); err != nil {
		t.Fatal(err)
	}
	if fm := frontMatterOf(t, "blank.md"); !strings.Contains(fm, "R. S. Doiel") {
		t.Errorf("a blank author gets the default:\n%s", fm)
	}
}

func TestPageMarkdownWithNoSiteAuthorIsUnchanged(t *testing.T) {
	markdownPageSite(t)
	src := "---\ntitle: Plain\n---\nBody.\n"
	os.WriteFile("plain.txt", []byte(src), 0o664)
	if err := runIn(t, "page", "plain.txt", "plain.md"); err != nil {
		t.Fatal(err)
	}
	if got := readText(t, "plain.md"); got != src {
		t.Errorf("with no site author and a title the copy is the source:\n%q", got)
	}
}

// DR-0019 6: a title from the first heading when the source has none; a source
// with no front matter at all gets a block.
func TestPageMarkdownDerivesTheTitleFromTheHeading(t *testing.T) {
	markdownPageSite(t)
	setConfig(t, "author: R. S. Doiel")
	os.WriteFile("bare.txt", []byte("# The *Real* Title\n\nBody.\n"), 0o664)
	if err := runIn(t, "page", "bare.txt", "bare.md"); err != nil {
		t.Fatal(err)
	}
	got := readText(t, "bare.md")
	fm := frontMatterOf(t, "bare.md")
	if !strings.Contains(fm, "title: The Real Title") || !strings.Contains(fm, "author: R. S. Doiel") {
		t.Errorf("a bare source gets a title and the author:\n%s", got)
	}
	if !strings.HasSuffix(got, "# The *Real* Title\n\nBody.\n") {
		t.Errorf("the body is unchanged:\n%q", got)
	}
	// Run twice: nothing changes.
	if err := runIn(t, "page", "bare.txt", "bare.md"); err != nil {
		t.Fatal(err)
	}
	if again := readText(t, "bare.md"); again != got {
		t.Errorf("a second run changed the copy")
	}
}

func TestPageMarkdownNeverInventsATitle(t *testing.T) {
	markdownPageSite(t)
	os.WriteFile("nohead.txt", []byte("Just text.\n"), 0o664)
	if err := runIn(t, "page", "nohead.txt", "nohead.md"); err != nil {
		t.Fatal(err)
	}
	if got := readText(t, "nohead.md"); got != "Just text.\n" {
		t.Errorf("no heading, no site author: nothing to add:\n%q", got)
	}
}

// Posts: the author lands in what generate writes for the site's own posts.
// The collection page marks each item with author:NAME for search.
func authorSite(t *testing.T) {
	t.Helper()
	newWorkspace(t, true)
	setConfig(t, "author: R. S. Doiel")
	os.MkdirAll("blog/2026/11/19", 0o775)
	os.WriteFile("blog/2026/11/19/mine.md", []byte("---\ntitle: Mine\npostPath: blog/2026/11/19/mine.md\npubDate: \"2026-11-19\"\n---\nBody.\n"), 0o664)
	os.WriteFile("blog/2026/11/19/guest.md", []byte("---\ntitle: Guest\nauthor: A. Guest\npostPath: blog/2026/11/19/guest.md\npubDate: \"2026-11-19\"\n---\nBody.\n"), 0o664)
	for _, p := range []string{"blog/2026/11/19/mine.md", "blog/2026/11/19/guest.md"} {
		if err := runIn(t, "post", p); err != nil {
			t.Fatalf("post %s: %s", p, err)
		}
	}
	if err := runIn(t, "generate"); err != nil {
		t.Fatalf("generate: %s", err)
	}
}

func TestPostsGetTheSiteAuthorInTheCollectionPage(t *testing.T) {
	authorSite(t)
	html := readText(t, "pages.html")
	if !strings.Contains(html, "author:R. S. Doiel") {
		t.Errorf("a post with no author should carry the site's:\n%s", html)
	}
	if !strings.Contains(html, "author:A. Guest") {
		t.Errorf("a guest post keeps its author")
	}
	if strings.Count(html, "author:R. S. Doiel") != 1 {
		t.Errorf("only the post without an author gets the default")
	}
	if got := readText(t, "blog/2026/11/19/mine.md"); strings.Contains(got, "author:") {
		t.Errorf("the post file must not be changed:\n%s", got)
	}
}

func TestHarvestedItemsNeverGetTheSiteAuthor(t *testing.T) {
	newWorkspace(t, true)
	setConfig(t, "author: R. S. Doiel")
	cfg := &AppConfig{}
	if err := cfg.LoadConfig("antenna.yaml"); err != nil {
		t.Fatal(err)
	}
	// An item with no postPath is someone else's: build it through the db the
	// generator reads.
	collection, err := cfg.GetCollection("pages.md")
	if err != nil {
		t.Fatal(err)
	}
	if err := insertHarvestedItem(t, collection.DbName); err != nil {
		t.Fatal(err)
	}
	if err := runIn(t, "generate"); err != nil {
		t.Fatalf("generate: %s", err)
	}
	if strings.Contains(readText(t, "pages.html"), "author:R. S. Doiel") {
		t.Errorf("someone else's item was given the site's author")
	}
}

// insertHarvestedItem adds a published item that has no postPath, the way a
// harvested feed item is stored, and no author of its own.
func insertHarvestedItem(t *testing.T, dbName string) error {
	t.Helper()
	db, err := sql.Open("sqlite", dbName)
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(`INSERT INTO items (link, title, pubDate, status, label, postPath)
		VALUES ('https://other.example/a', 'Elsewhere', '2026-11-19', 'published', 'Other', '')`)
	return err
}

func TestPageMarkdownFillsAnEmptyFrontMatterBlock(t *testing.T) {
	markdownPageSite(t)
	setConfig(t, "author: R. S. Doiel")
	os.WriteFile("empty.txt", []byte("---\n---\n# Empty\n\nText.\n"), 0o664)
	if err := runIn(t, "page", "empty.txt", "empty.md"); err != nil {
		t.Fatal(err)
	}
	got := readText(t, "empty.md")
	if !strings.HasPrefix(got, "---\ntitle: Empty\nauthor: R. S. Doiel\n---\n# Empty") || !strings.HasSuffix(got, "# Empty\n\nText.\n") {
		t.Errorf("empty block:\n%q", got)
	}
}
