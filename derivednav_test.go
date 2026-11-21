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
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// derivedNavSite is a workspace made the way a person makes one: antenna init, three
// pages (a, b, c, in output path order) made with antenna page, and three posts
// in different folders made with antenna post. The Markdown of each is
// published in htdocs. Nothing is generated yet.
func derivedNavSite(t *testing.T) (htdocs string) {
	t.Helper()
	newWorkspace(t, true)
	cfgYAML, _ := os.ReadFile("antenna.yaml")
	edited := strings.Replace(string(cfgYAML), "base_url: http://localhost:8000", "base_url: https://example.org\nhtdocs: htdocs", 1)
	if edited == string(cfgYAML) {
		t.Fatalf("antenna.yaml is not what the test expects:\n%s", cfgYAML)
	}
	os.WriteFile("antenna.yaml", []byte(edited), 0o664)
	htdocs = "htdocs"
	os.MkdirAll(htdocs, 0o775)
	for _, p := range []string{"a", "b", "c"} {
		os.WriteFile(p+".md", []byte("---\ntitle: Page "+p+"\n---\n# "+p+"\n"), 0o664)
		if err := runIn(t, "page", p+".md", p+".html"); err != nil {
			t.Fatalf("page %s: %s", p, err)
		}
		publish(t, htdocs, p+".md")
	}
	for i, p := range []struct{ path, date string }{
		{"blog/2026/10/30/a.md", "2026-10-30"},
		{"blog/2026/11/02/b.md", "2026-11-02"},
		{"blog/2026/11/19/c.md", "2026-11-19"},
	} {
		src := "---\ntitle: Post " + string(rune('A'+i)) + "\npostPath: " + p.path + "\npubDate: \"" + p.date + "\"\n---\nBody.\n"
		name := "post" + string(rune('1'+i)) + ".md"
		os.WriteFile(name, []byte(src), 0o664)
		os.MkdirAll(filepath.Join(htdocs, filepath.Dir(p.path)), 0o775)
		if err := runIn(t, "post", name); err != nil {
			t.Fatalf("post %s: %s", name, err)
		}
		publish(t, htdocs, p.path)
	}
	return htdocs
}

var (
	prevRE = regexp.MustCompile(`<a rel="prev" href="([^"]*)">`)
	nextRE = regexp.MustCompile(`<a rel="next" href="([^"]*)">`)
)

// neighbours returns the previous and next addresses a generated page shows
// ("" for none), and checks the <link> elements in the head agree with the nav.
func neighbours(t *testing.T, file string) (prev, next string) {
	t.Helper()
	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	page := string(b)
	if m := prevRE.FindStringSubmatch(page); m != nil {
		prev = m[1]
	}
	if m := nextRE.FindStringSubmatch(page); m != nil {
		next = m[1]
	}
	head := page[:strings.Index(page, "<body>")]
	for rel, want := range map[string]string{"prev": prev, "next": next} {
		has := regexp.MustCompile(`<link[^>]*rel="` + rel + `"[^>]*>`).MatchString(head)
		if has != (want != "") {
			t.Errorf("%s: <link rel=%s> in the head is %v but the visible navigation is %q", file, rel, has, want)
		}
	}
	return prev, next
}

// DR-0025: pages are not a sequence, so none is derived for them; the table of
// contents lists them in file-name order, which means nothing.
func TestPagesGetNoDerivedPreviousAndNext(t *testing.T) {
	htdocs := derivedNavSite(t)
	if err := runIn(t, "generate"); err != nil {
		t.Fatalf("generate: %s", err)
	}
	for _, file := range []string{"a.html", "b.html", "c.html"} {
		prev, next := neighbours(t, filepath.Join(htdocs, file))
		if prev != "" || next != "" {
			t.Errorf("%s: a page gets no derived arrows, got previous %q next %q", file, prev, next)
		}
	}
}

// In Posts, previous is the older post and next the newer one; the addresses
// are relative to the post, across folders.
func TestPostsGetDerivedPreviousAsOlderAndNextAsNewer(t *testing.T) {
	htdocs := derivedNavSite(t)
	if err := runIn(t, "generate"); err != nil {
		t.Fatalf("generate: %s", err)
	}
	for file, want := range map[string][2]string{
		"blog/2026/10/30/a.html": {"", "../../11/02/b.html"},             // oldest: only a newer one
		"blog/2026/11/02/b.html": {"../../10/30/a.html", "../19/c.html"}, // older and newer, in other folders
		"blog/2026/11/19/c.html": {"../02/b.html", ""},                   // newest: only an older one
	} {
		prev, next := neighbours(t, filepath.Join(htdocs, file))
		if prev != want[0] || next != want[1] {
			t.Errorf("%s: previous %q next %q, want %q and %q", file, prev, next, want[0], want[1])
		}
	}
}

// The author's own keys always win, on a page (DR-0025 2) and on a post, and a
// missing one is still filled in for a post.
func TestExplicitKeysWinOnAPageAndAPost(t *testing.T) {
	htdocs := derivedNavSite(t)
	os.WriteFile("b.md", []byte("---\ntitle: Page b\nurlPrev: https://elsewhere.example/start\nurlNext: c.html\n---\n# b\n"), 0o664)
	publish(t, htdocs, "b.md")
	if err := runIn(t, "page", "b.md", "b.html"); err != nil {
		t.Fatal(err)
	}
	os.WriteFile("post2.md", []byte("---\ntitle: Post B\npostPath: blog/2026/11/02/b.md\npubDate: \"2026-11-02\"\nurlPrev: https://elsewhere.example/before\n---\nBody.\n"), 0o664)
	if err := runIn(t, "post", "post2.md"); err != nil {
		t.Fatal(err)
	}
	if err := runIn(t, "generate"); err != nil {
		t.Fatal(err)
	}
	prev, next := neighbours(t, filepath.Join(htdocs, "b.html"))
	if prev != "https://elsewhere.example/start" || next != "c.html" {
		t.Errorf("page: previous %q next %q: want the explicit previous and next", prev, next)
	}
	prev, next = neighbours(t, filepath.Join(htdocs, "blog/2026/11/02/b.html"))
	if prev != "https://elsewhere.example/before" || next != "../19/c.html" {
		t.Errorf("post: previous %q next %q: want the explicit previous and the derived next", prev, next)
	}
}

// DR-0009 5: nothing is written into the author's Markdown, published or not.
func TestTheAuthorsMarkdownIsNeverChanged(t *testing.T) {
	htdocs := derivedNavSite(t)
	before := map[string]string{}
	for _, f := range []string{"a.md", "b.md", "post1.md", "post2.md"} {
		b, _ := os.ReadFile(f)
		before[f] = string(b)
	}
	pub, _ := os.ReadFile(filepath.Join(htdocs, "b.md"))
	if err := runIn(t, "generate"); err != nil {
		t.Fatal(err)
	}
	for f, want := range before {
		if got, _ := os.ReadFile(f); string(got) != want {
			t.Errorf("%s changed:\n%s", f, got)
		}
	}
	if got, _ := os.ReadFile(filepath.Join(htdocs, "b.md")); string(got) != string(pub) || strings.Contains(string(got), "urlPrev") {
		t.Errorf("the published Markdown changed:\n%s", got)
	}
}

// A post whose Markdown is not published is not in the outline, so it is
// skipped: the arrows jump over it, as the outline does.
func TestAnUnpublishedPostIsSkipped(t *testing.T) {
	htdocs := derivedNavSite(t)
	os.Remove(filepath.Join(htdocs, "blog/2026/11/02/b.md"))
	if err := runIn(t, "generate"); err != nil {
		t.Fatal(err)
	}
	if prev, next := neighbours(t, filepath.Join(htdocs, "blog/2026/10/30/a.html")); prev != "" || next != "../../11/19/c.html" {
		t.Errorf("a: previous %q next %q, want it to skip b", prev, next)
	}
	if prev, next := neighbours(t, filepath.Join(htdocs, "blog/2026/11/02/b.html")); prev != "" || next != "" {
		t.Errorf("b is not in the outline, so it has no derived arrows: %q %q", prev, next)
	}
}

// Arrows come from generate, which knows the whole site; antenna post alone does not derive them.
func TestOnlyGenerateDerivesTheArrows(t *testing.T) {
	htdocs := derivedNavSite(t)
	if prev, next := neighbours(t, filepath.Join(htdocs, "blog/2026/11/02/b.html")); prev != "" || next != "" {
		t.Errorf("antenna post derived arrows: %q %q", prev, next)
	}
	// generate of one collection still gives every post its arrows
	if err := runIn(t, "generate", "pages.md"); err != nil {
		t.Fatal(err)
	}
	if prev, next := neighbours(t, filepath.Join(htdocs, "blog/2026/11/02/b.html")); prev != "../../10/30/a.html" || next != "../19/c.html" {
		t.Errorf("generate pages.md: %q %q", prev, next)
	}
}

// DR-0025 3: a page and a post together do not borrow each other's neighbours.
func TestAPageAndPostsDoNotShareNeighbours(t *testing.T) {
	htdocs := derivedNavSite(t)
	if err := runIn(t, "generate"); err != nil {
		t.Fatal(err)
	}
	if _, next := neighbours(t, filepath.Join(htdocs, "blog/2026/11/19/c.html")); next != "" {
		t.Errorf("the newest post must not point at a page: %q", next)
	}
	if prev, _ := neighbours(t, filepath.Join(htdocs, "blog/2026/10/30/a.html")); prev != "" {
		t.Errorf("the oldest post must not point at a page: %q", prev)
	}
}

// The outline's own rule, without a site: posts by date (DR-0025); pages have none.
func TestBuildNavFromTheOutlinesOrder(t *testing.T) {
	posts := []tocEntry{{md: "new.md"}, {md: "mid.md"}, {md: "old.md"}} // newest first, as in the outline
	nav := buildTOCNav(posts)
	want := map[string]tocNeighbours{
		"new.md": {prev: "mid.md", next: ""},
		"mid.md": {prev: "old.md", next: "new.md"},
		"old.md": {prev: "", next: "mid.md"},
	}
	for md, w := range want {
		if nav[md] != w {
			t.Errorf("%s: %+v, want %+v", md, nav[md], w)
		}
	}
	if len(nav) != len(want) {
		t.Errorf("nav has %d entries, want %d", len(nav), len(want))
	}
}

func TestRelativeMarkdownPath(t *testing.T) {
	for _, c := range []struct{ from, to, want string }{
		{"about.md", "series/books.md", "series/books.md"},
		{"series/books.md", "about.md", "../about.md"},
		{"blog/2026/11/02/b.md", "blog/2026/11/19/c.md", "../19/c.md"},
		{"blog/2026/11/02/b.md", "blog/2026/10/30/a.md", "../../10/30/a.md"},
		{"a/b.md", "a/c.md", "c.md"},
	} {
		if got := relativeMarkdownPath(c.from, c.to); got != c.want {
			t.Errorf("relativeMarkdownPath(%q, %q) = %q, want %q", c.from, c.to, got, c.want)
		}
	}
}

func TestTheManualsSayWhereDerivedArrowsComeFrom(t *testing.T) {
	for name, text := range map[string]string{"generate": GenerateHelpText, "metadata": MetadataHelpText} {
		if !strings.Contains(text, "table of contents") || !strings.Contains(text, "older") {
			t.Errorf("antenna help %s does not explain the derived previous and next", name)
		}
	}
}
