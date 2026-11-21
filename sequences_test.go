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
	"path/filepath"
	"strings"
	"testing"
)

// DR-0026: antenna.yaml names the paths that hold a sequence.

const (
	postA = "blog/2026/10/30/a.html" // oldest
	postB = "blog/2026/11/02/b.html"
	postC = "blog/2026/11/19/c.html" // newest
)

// seqSite is derivedNavSite (pages a, b, c and posts a, b, c, published under
// htdocs) with sequences added to antenna.yaml.
func seqSite(t *testing.T, sequences string) (htdocs string) {
	t.Helper()
	htdocs = derivedNavSite(t)
	if sequences != "" {
		setConfig(t, sequences)
	}
	return htdocs
}

// seqList writes a list file series/s.md, a page published with its
// Markdown, whose body is body.
func seqList(t *testing.T, htdocs, body string) {
	t.Helper()
	src := "---\ntitle: S\n---\n" + body
	os.MkdirAll("series", 0o775)
	if err := os.WriteFile("series/s.md", []byte(src), 0o664); err != nil {
		t.Fatal(err)
	}
	if err := runIn(t, "page", "series/s.md", "series/s.html"); err != nil {
		t.Fatal(err)
	}
	os.MkdirAll(filepath.Join(htdocs, "series"), 0o775)
	os.WriteFile(filepath.Join(htdocs, "series/s.md"), []byte(src), 0o664)
}

func arrows(t *testing.T, htdocs, html string) [2]string {
	t.Helper()
	prev, next := neighbours(t, filepath.Join(htdocs, html))
	return [2]string{prev, next}
}

const seqSequence = "sequences:\n  - path: series/s.md\n"

// The author's order is the list's order, not the posts' dates; links that are
// not published Markdown, the list itself and repeats are skipped.
func TestSequenceListFileGivesItsMembersNeighboursInDocumentOrder(t *testing.T) {
	htdocs := seqSite(t, seqSequence)
	seqList(t, htdocs, "- [C](/blog/2026/11/19/c.md)\n- [A](/blog/2026/10/30/a.md)\n- [B](../blog/2026/11/02/b.md)\n"+
		"- [elsewhere](https://example.org/x.md)\n- [gone](/blog/2030/01/01/nope.md)\n- [self](s.md)\n- [again](/blog/2026/11/19/c.md)\n")
	if err := runIn(t, "generate"); err != nil {
		t.Fatal(err)
	}
	for html, want := range map[string][2]string{
		postC: {"", "../../10/30/a.html"},
		postA: {"../../11/19/c.html", "../../11/02/b.html"},
		postB: {"../../10/30/a.html", ""},
	} {
		if got := arrows(t, htdocs, html); got != want {
			t.Errorf("%s: %v, want %v", html, got, want)
		}
	}
	if got := arrows(t, htdocs, "series/s.html"); got != [2]string{"", ""} {
		t.Errorf("the list file is not a member of its own sequence: %v", got)
	}
}

func TestAPostOutsideASequenceKeepsTheChronologicalNeighbours(t *testing.T) {
	htdocs := seqSite(t, seqSequence)
	seqList(t, htdocs, "- [C](/blog/2026/11/19/c.md)\n- [A](/blog/2026/10/30/a.md)\n")
	if err := runIn(t, "generate"); err != nil {
		t.Fatal(err)
	}
	if got := arrows(t, htdocs, postB); got != [2]string{"../../10/30/a.html", "../19/c.html"} {
		t.Errorf("b is in no sequence: %v", got)
	}
	if got := arrows(t, htdocs, postA); got != [2]string{"../../11/19/c.html", ""} {
		t.Errorf("a is in the sequence, second: %v", got)
	}
}

func TestAPageInNoSequenceGetsNoneAndAMemberKeepsItsOwnKeys(t *testing.T) {
	htdocs := seqSite(t, seqSequence)
	seqList(t, htdocs, "- [C](/blog/2026/11/19/c.md)\n- [A](/blog/2026/10/30/a.md)\n")
	os.WriteFile("post1.md", []byte("---\ntitle: Post A\npostPath: blog/2026/10/30/a.md\npubDate: \"2026-10-30\"\nurlPrev: https://elsewhere.example/p\n---\nBody.\n"), 0o664)
	runIn(t, "post", "post1.md")
	runIn(t, "generate")
	if got := arrows(t, htdocs, "a.html"); got != [2]string{"", ""} {
		t.Errorf("a page in no sequence: %v", got)
	}
	if got := arrows(t, htdocs, postA); got != [2]string{"https://elsewhere.example/p", ""} {
		t.Errorf("an explicit key wins: %v", got)
	}
}

// A directory: its pages, index.md left out, in name or date order.
func dirSite(t *testing.T, order string) (htdocs string) {
	t.Helper()
	seq := "sequences:\n  - path: docs/\n"
	if order != "" {
		seq += "    order: " + order + "\n"
	}
	htdocs = seqSite(t, seq)
	os.MkdirAll("docs", 0o775)
	os.MkdirAll(filepath.Join(htdocs, "docs"), 0o775)
	for name, date := range map[string]string{"a": "2026-03-01", "b": "2026-01-01", "c": "2026-02-01", "index": "2026-04-01"} {
		src := "---\ntitle: " + name + "\npubDate: \"" + date + "\"\n---\n# " + name + "\n"
		os.WriteFile("docs/"+name+".md", []byte(src), 0o664)
		os.WriteFile(filepath.Join(htdocs, "docs", name+".md"), []byte(src), 0o664)
		if err := runIn(t, "page", "docs/"+name+".md", "docs/"+name+".html"); err != nil {
			t.Fatal(err)
		}
	}
	if err := runIn(t, "generate"); err != nil {
		t.Fatal(err)
	}
	return htdocs
}

func TestSequenceDirectoryOrders(t *testing.T) {
	for order, want := range map[string]map[string][2]string{
		"": {"docs/a.html": {"", "b.html"}, "docs/b.html": {"a.html", "c.html"}, "docs/c.html": {"b.html", ""}},
		"date": { // b (Jan), c (Feb), a (Mar)
			"docs/b.html": {"", "c.html"}, "docs/c.html": {"b.html", "a.html"}, "docs/a.html": {"c.html", ""}},
		"date-desc": { // a, c, b
			"docs/a.html": {"", "c.html"}, "docs/c.html": {"a.html", "b.html"}, "docs/b.html": {"c.html", ""}},
	} {
		htdocs := dirSite(t, order)
		for html, w := range want {
			if got := arrows(t, htdocs, html); got != w {
				t.Errorf("order %q %s: %v, want %v", order, html, got, w)
			}
		}
		if got := arrows(t, htdocs, "docs/index.html"); got != [2]string{"", ""} {
			t.Errorf("order %q: index.md is not a member: %v", order, got)
		}
	}
}

// The generated Markdown copy carries the keys; the source is not changed.
func TestTheMarkdownCopyOfAMemberCarriesTheKeysAndTheSourceIsUntouched(t *testing.T) {
	htdocs := seqSite(t, "sequences:\n  - path: docs/\n")
	os.MkdirAll("docs", 0o775)
	os.MkdirAll(filepath.Join(htdocs, "docs"), 0o775)
	for _, n := range []string{"a", "b"} {
		src := "---\ntitle: " + n + "\n---\n# " + n + "\n"
		os.WriteFile("docs/"+n+".md", []byte(src), 0o664)
		os.WriteFile(filepath.Join(htdocs, "docs", n+".md"), []byte(src), 0o664)
		runIn(t, "page", "docs/"+n+".md", "docs/"+n+".html")
	}
	source := "---\ntitle: c\n---\n# c\n"
	os.WriteFile("docs/c.txt", []byte(source), 0o664)
	if err := runIn(t, "page", "docs/c.txt", "docs/c.md", "docs/c.html"); err != nil {
		t.Fatal(err)
	}
	if err := runIn(t, "generate"); err != nil {
		t.Fatal(err)
	}
	fm := frontMatterOf(t, filepath.Join(htdocs, "docs/c.md"))
	if !strings.Contains(fm, "urlPrev: b.md") || strings.Contains(fm, "urlNext") {
		t.Errorf("the copy of the last member has a previous only:\n%s", fm)
	}
	if got := readText(t, "docs/c.txt"); got != source {
		t.Errorf("the source was changed:\n%s", got)
	}
	first := readText(t, filepath.Join(htdocs, "docs/c.md"))
	runIn(t, "generate")
	if readText(t, filepath.Join(htdocs, "docs/c.md")) != first {
		t.Errorf("a second generate changed the copy")
	}
	// The author's own published Markdown is not rewritten.
	if b := readText(t, filepath.Join(htdocs, "docs/b.md")); strings.Contains(b, "urlPrev") {
		t.Errorf("a served-in-place document was rewritten:\n%s", b)
	}
}

func TestADocumentInTwoSequencesTakesTheFirstAndValidateSaysSo(t *testing.T) {
	htdocs := seqSite(t, "sequences:\n  - path: series/s.md\n  - path: series/t.md\n")
	seqList(t, htdocs, "- [C](/blog/2026/11/19/c.md)\n- [A](/blog/2026/10/30/a.md)\n")
	os.WriteFile(filepath.Join(htdocs, "series/t.md"), []byte("- [B](/blog/2026/11/02/b.md)\n- [A](/blog/2026/10/30/a.md)\n"), 0o664)
	runIn(t, "generate")
	if got := arrows(t, htdocs, postA); got != [2]string{"../../11/19/c.html", ""} {
		t.Errorf("a is in both and takes the first sequence: %v", got)
	}
	out, _ := runValidate(t)
	if line, ok := finding(out, "antenna.yaml", "info", "sequence-overlap"); !ok || !strings.Contains(line, "blog/2026/10/30/a.md") || !strings.Contains(line, "series/t.md") {
		t.Errorf("sequence-overlap:\n%s", out)
	}
}

func TestSequenceConfigurationErrors(t *testing.T) {
	for name, yaml := range map[string]string{
		"no path":         "sequences:\n  - order: name\n",
		"not a dir or md": "sequences:\n  - path: docs\n",
		"unknown order":   "sequences:\n  - path: docs/\n    order: random\n",
		"order on a list": "sequences:\n  - path: series/s.md\n    order: name\n",
		"duplicate path":  "sequences:\n  - path: docs/\n  - path: docs/\n",
		"not a list":      "sequences: docs/\n",
	} {
		newWorkspace(t, true)
		setConfig(t, yaml)
		_, err := runValidate(t)
		wantClass(t, name, err, classConfig)
	}
}

func TestValidateSequencePathMissingAndEmpty(t *testing.T) {
	htdocs := seqSite(t, "sequences:\n  - path: nowhere/\n  - path: gone.md\n  - path: empty/\n")
	os.MkdirAll(filepath.Join(htdocs, "empty"), 0o775)
	out, _ := runValidate(t)
	for _, path := range []string{"nowhere/", "gone.md"} {
		if _, ok := finding(out, "antenna.yaml", "warning", "sequence-path-missing"); !ok || !strings.Contains(out, path) {
			t.Errorf("sequence-path-missing for %s:\n%s", path, out)
		}
	}
	if line, ok := finding(out, "antenna.yaml", "info", "sequence-empty"); !ok || !strings.Contains(line, "empty/") {
		t.Errorf("sequence-empty:\n%s", out)
	}
}
