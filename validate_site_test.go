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

// goodSite is a site made with the real verbs: a page, a post made by blogit,
// and a generated site, so validate has nothing to say about it.
func goodSite(t *testing.T) {
	t.Helper()
	newWorkspace(t, true)
	write(t, "about.md", "---\ntitle: About\n---\n# About\n")
	if err := runIn(t, "page", "about.md", "about.html"); err != nil {
		t.Fatalf("page: %s", err)
	}
	write(t, "hello.md", "# Hello\n\nA first post, see [About](../../../../about.md).\n")
	if err := runIn(t, "blogit", "hello.md", "2026-11-19"); err != nil {
		t.Fatalf("blogit: %s", err)
	}
	if err := runIn(t, "generate"); err != nil {
		t.Fatalf("generate: %s", err)
	}
}

func editConfig(t *testing.T, old, new string) {
	t.Helper()
	src := readFile(t, "antenna.yaml")
	if !strings.Contains(src, old) {
		t.Fatalf("antenna.yaml has no %q:\n%s", old, src)
	}
	write(t, "antenna.yaml", strings.Replace(src, old, new, 1))
}

// The end to end case: a site made and generated with the real verbs is clean.
func TestValidateAGeneratedGoodSiteIsClean(t *testing.T) {
	goodSite(t)
	out, err := runValidate(t)
	if err != nil {
		t.Fatalf("validate of a good site: %s\n%s", err, out)
	}
	if !strings.Contains(out, "0 errors, 0 warnings") {
		t.Errorf("summary:\n%s", out)
	}
}

// A PATH run checks files only: site checks need the whole site.
func TestValidateAPathRunSkipsSiteChecks(t *testing.T) {
	goodSite(t)
	editConfig(t, "base_url: http://localhost:8000\n", "")
	out, _ := runValidate(t, "about.md")
	if strings.Contains(out, "no-base-url") {
		t.Errorf("a file run reported a site finding:\n%s", out)
	}
}

func TestValidateNoBaseURL(t *testing.T) {
	goodSite(t)
	editConfig(t, "base_url: http://localhost:8000\n", "")
	out, err := runValidate(t)
	wantClass(t, "no base_url", err, classNegative)
	if _, ok := finding(out, "antenna.yaml", "warning", "no-base-url"); !ok {
		t.Errorf("no no-base-url:\n%s", out)
	}
}

func TestValidateHtdocsMissing(t *testing.T) {
	goodSite(t)
	editConfig(t, "base_url:", "htdocs: nosuchfolder\nbase_url:")
	out, err := runValidate(t)
	wantClass(t, "htdocs missing", err, classNegative)
	if _, ok := finding(out, "antenna.yaml", "error", "htdocs-missing"); !ok {
		t.Errorf("no htdocs-missing:\n%s", out)
	}
}

// A page row whose input file is gone: the stale index.txt row of 2026-11-20.
func TestValidateRowSourceMissing(t *testing.T) {
	goodSite(t)
	os.Remove("about.md")
	out, err := runValidate(t)
	wantClass(t, "a stale page row", err, classNegative)
	if _, ok := finding(out, "about.md", "error", "row-source-missing"); !ok {
		t.Errorf("no row-source-missing:\n%s", out)
	}
}

// With htdocs apart from the sources, the Markdown has to be published there.
func TestValidateNotPublished(t *testing.T) {
	goodSite(t)
	os.MkdirAll("out", 0o775)
	editConfig(t, "base_url:", "htdocs: out\nbase_url:")
	out, err := runValidate(t)
	wantClass(t, "Markdown not in htdocs", err, classNegative)
	if _, ok := finding(out, "about.md", "warning", "not-published"); !ok {
		t.Errorf("no not-published for the page:\n%s", out)
	}
	if _, ok := finding(out, "blog/2026/11/19/hello.md", "warning", "not-published"); !ok {
		t.Errorf("no not-published for the post:\n%s", out)
	}
}

func TestValidateTOCStaleAndMissing(t *testing.T) {
	goodSite(t)
	// a new page after generate: the table of contents is out of date
	write(t, "more.md", "---\ntitle: More\n---\n# More\n")
	if err := runIn(t, "page", "more.md", "more.html"); err != nil {
		t.Fatal(err)
	}
	out, err := runValidate(t)
	wantClass(t, "stale toc", err, classNegative)
	if _, ok := finding(out, ".well-known/toc.opml", "warning", "toc-stale"); !ok {
		t.Errorf("no toc-stale after adding a page:\n%s", out)
	}
	// generate brings it up to date
	if err := runIn(t, "generate"); err != nil {
		t.Fatal(err)
	}
	if out, err := runValidate(t); err != nil {
		t.Errorf("after generate: %s\n%s", err, out)
	}
	// and a missing one is reported
	os.Remove(".well-known/toc.opml")
	out, _ = runValidate(t)
	if _, ok := finding(out, ".well-known/toc.opml", "warning", "toc-stale"); !ok {
		t.Errorf("no toc-stale for a missing toc:\n%s", out)
	}
}

func TestValidateBonbonJSON(t *testing.T) {
	goodSite(t)
	editConfig(t, "base_url:", "bonbon:\n    urlSearch: https://example.org/search\nbase_url:")
	out, _ := runValidate(t)
	if _, ok := finding(out, ".well-known/bonbon.json", "warning", "bonbon-json"); !ok {
		t.Errorf("a bonbon block with no file should be reported:\n%s", out)
	}
	if err := runIn(t, "generate"); err != nil {
		t.Fatal(err)
	}
	if out, err := runValidate(t); err != nil {
		t.Errorf("after generate: %s\n%s", err, out)
	}
	// a hand-written file with no block: invalid JSON, then a bad value
	goodSite(t)
	write(t, ".well-known/bonbon.json", "{not json")
	out, _ = runValidate(t)
	if _, ok := finding(out, ".well-known/bonbon.json", "warning", "bonbon-json"); !ok {
		t.Errorf("invalid JSON not reported:\n%s", out)
	}
	write(t, ".well-known/bonbon.json", `{"urlSearch": ""}`)
	out, _ = runValidate(t)
	if line, ok := finding(out, ".well-known/bonbon.json", "warning", "bonbon-json"); !ok || !strings.Contains(line, "urlSearch") {
		t.Errorf("an empty urlSearch not reported:\n%s", out)
	}
	write(t, ".well-known/bonbon.json", `{"urlMirrorOf": "http://example.org/", "other": 1}`)
	if out, err := runValidate(t); err != nil {
		t.Errorf("a good hand-written file: %s\n%s", err, out)
	}
}

func TestValidateNoJekyll(t *testing.T) {
	goodSite(t)
	editConfig(t, "base_url: http://localhost:8000", "base_url: https://someone.github.io")
	out, _ := runValidate(t)
	if _, ok := finding(out, ".nojekyll", "warning", "nojekyll"); !ok {
		t.Errorf("no nojekyll warning for a github.io site:\n%s", out)
	}
	write(t, ".nojekyll", "")
	out, _ = runValidate(t)
	if strings.Contains(out, "nojekyll") {
		t.Errorf("reported with the file present:\n%s", out)
	}
}
