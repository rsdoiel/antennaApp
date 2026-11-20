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
	"strings"
	"testing"
)

// runMigrate runs antenna migrate-links through the real dispatcher in the
// current directory, with the arguments given, and returns what it wrote to
// standard output and the error.
func runMigrate(t *testing.T, args ...string) (string, error) {
	t.Helper()
	app := &AntennaApp{appName: "antenna"}
	var out, eout bytes.Buffer
	err := app.Run(strings.NewReader(""), &out, &eout, "antenna.yaml", "migrate-links", args)
	return out.String(), err
}

// migrateDoc is a post with front matter, then the body given; the body starts
// on line 6.
func migrateDoc(body string) string {
	return "---\ntitle: T\npostPath: docs/p.md\npubDate: \"Thu, 19 Nov 2026 12:00:00 -0800\"\n---\n" + body
}

// DR-0017 8: a dry run shows a diff and writes nothing.
func TestMigrateLinksDryRunShowsADiffAndWritesNothing(t *testing.T) {
	bodySite(t)
	src := migrateDoc("See [About](../about.html#team).\n")
	write(t, "docs/p.md", src)
	out, err := runMigrate(t, "docs/p.md")
	if err != nil {
		t.Fatalf("migrate-links: %s\n%s", err, out)
	}
	for _, want := range []string{"--- docs/p.md", "+++ docs/p.md", "@@ -6 +6 @@", "-See [About](../about.html#team).", "+See [About](../about.md#team)."} {
		if !strings.Contains(out, want) {
			t.Errorf("the diff lacks %q:\n%s", want, out)
		}
	}
	if !strings.Contains(out, "1 link in 1 file would change") {
		t.Errorf("no count of what would change:\n%s", out)
	}
	if got := readFile(t, "docs/p.md"); got != src {
		t.Errorf("a dry run changed the file:\n%s", got)
	}
}

// DR-0017 7, 8: -w changes only the bytes of the destinations. The expected
// result is the original with each destination replaced, byte for byte.
func TestMigrateLinksWriteChangesOnlyTheDestinations(t *testing.T) {
	bodySite(t)
	body := "" +
		"A [page](../about.html), a [fragment](../about.html#team), a [query](../about.html?x=1#y),\n" +
		"a [root](/about.html), a [titled](../about.html \"The title\"), an [angle](<../about.html>).\n\n" +
		"[ref link][r] and [other][o].\n\n" +
		"[r]: ../about.html\n" +
		"[o]: https://other.example/about.html\n\n" +
		"[collection](../news.html)\n\n" +
		"Untouched: [external](https://example.org/about.html), [feed](../index.xml), ![image](../about.html), `[code](../about.html)`,\n" +
		"[md already](../about.md), [dir](../series/), [frag](#top), [unknown](../handwritten.html).\n\n" +
		"```\n[fenced](../about.html)\n```\n\n" +
		"    [indented](../about.html)\n"
	src := migrateDoc(body)
	write(t, "docs/p.md", src)
	out, err := runMigrate(t, "-w", "docs/p.md")
	if err != nil {
		t.Fatalf("migrate-links -w: %s\n%s", err, out)
	}
	want := src
	for _, r := range [][2]string{
		{"[page](../about.html)", "[page](../about.md)"},
		{"[fragment](../about.html#team)", "[fragment](../about.md#team)"},
		{"[query](../about.html?x=1#y)", "[query](../about.md?x=1#y)"},
		{"[root](/about.html)", "[root](/about.md)"},
		{"[titled](../about.html \"The title\")", "[titled](../about.md \"The title\")"},
		{"[angle](<../about.html>)", "[angle](<../about.md>)"},
		{"[r]: ../about.html", "[r]: ../about.md"},
		{"[collection](../news.html)", "[collection](../news_items.md)"},
	} {
		if !strings.Contains(want, r[0]) {
			t.Fatalf("test bug: %q not in the source", r[0])
		}
		want = strings.Replace(want, r[0], r[1], 1)
	}
	if got := readFile(t, "docs/p.md"); got != want {
		t.Errorf("-w changed more or less than the destinations.\n--- got\n%s\n--- want\n%s", got, want)
	}
	if !strings.Contains(out, "8 links in 1 file changed") {
		t.Errorf("summary:\n%s", out)
	}
}

// Line endings and the front matter are kept exactly.
func TestMigrateLinksKeepsCRLFAndFrontMatter(t *testing.T) {
	bodySite(t)
	src := strings.ReplaceAll(migrateDoc("See [About](../about.html).\n\nMore text.\n"), "\n", "\r\n")
	write(t, "docs/p.md", src)
	if _, err := runMigrate(t, "-w", "docs/p.md"); err != nil {
		t.Fatal(err)
	}
	want := strings.Replace(src, "../about.html", "../about.md", 1)
	if got := readFile(t, "docs/p.md"); got != want {
		t.Errorf("CRLF or front matter changed:\n%q\nwant\n%q", got, want)
	}
}

// DR-0017 10: running -w twice changes nothing the second time.
func TestMigrateLinksIsIdempotent(t *testing.T) {
	bodySite(t)
	write(t, "docs/p.md", migrateDoc("[About](../about.html)\n"))
	if _, err := runMigrate(t, "-w", "docs/p.md"); err != nil {
		t.Fatal(err)
	}
	first := readFile(t, "docs/p.md")
	out, err := runMigrate(t, "-w", "docs/p.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "nothing to migrate") {
		t.Errorf("the second run should say there is nothing to migrate:\n%s", out)
	}
	if readFile(t, "docs/p.md") != first {
		t.Error("the second run changed the file")
	}
}

// DR-0017 7: a collection with no items Markdown is left, and said so.
func TestMigrateLinksLeavesACollectionWithNoItemsFileAndSaysSo(t *testing.T) {
	bodySite(t)
	os.Remove("news_items.md")
	src := migrateDoc("[News](../news.html)\n")
	write(t, "docs/p.md", src)
	out, err := runMigrate(t, "-w", "docs/p.md")
	if err != nil {
		t.Fatal(err)
	}
	if readFile(t, "docs/p.md") != src {
		t.Error("a link with no Markdown to point at was changed")
	}
	if !strings.Contains(out, "news.html") || !strings.Contains(out, "left") {
		t.Errorf("the left link is not reported:\n%s", out)
	}
}

// DR-0017 8: the whole site is the Markdown the site's data names; a PATH
// limits it to a file or a directory.
func TestMigrateLinksScope(t *testing.T) {
	bodySite(t)
	link := "[About](../about.html)\n"
	write(t, "docs/a.md", migrateDoc(link))
	write(t, "docs/sub/b.md", migrateDoc("[About](../../about.html)\n"))
	write(t, "elsewhere/c.md", migrateDoc("[About](../about.html)\n"))
	// the page the site renders is in scope for the whole-site run
	write(t, "series.md", "---\ntitle: S\n---\n[About](about.html)\n")
	if err := runIn(t, "page", "series.md", "series.html"); err != nil {
		t.Fatal(err)
	}
	if _, err := runMigrate(t, "-w", "docs"); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"docs/a.md", "docs/sub/b.md"} {
		if strings.Contains(readFile(t, f), "about.html") {
			t.Errorf("%s was not migrated by PATH docs", f)
		}
	}
	if !strings.Contains(readFile(t, "elsewhere/c.md"), "about.html") {
		t.Error("a file outside PATH was changed")
	}
	if _, err := runMigrate(t, "-w"); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(readFile(t, "series.md"), "about.html") {
		t.Error("the whole-site run did not migrate a page of the site")
	}
	if !strings.Contains(readFile(t, "elsewhere/c.md"), "about.html") {
		t.Error("the whole-site run changed a file the site's data does not name")
	}
}

// DR-0017 9: exit codes for the command line.
func TestMigrateLinksExitCodes(t *testing.T) {
	bodySite(t)
	_, err := runMigrate(t, "nosuch.md")
	wantClass(t, "a missing PATH", err, classNoInput)
	_, err = runMigrate(t, "-nosuchflag")
	wantClass(t, "an unknown flag", err, classUsage)
	_, err = runMigrate(t, "a.md", "b.md")
	wantClass(t, "two paths", err, classUsage)
	t.Chdir(t.TempDir())
	write(t, "p.md", "[a](b.html)\n")
	_, err = runMigrate(t, "p.md")
	wantClass(t, "no antenna.yaml", err, classNoInput)
}

// DR-0017 9: a file it cannot parse is reported (65) after the others are done.
func TestMigrateLinksUnparseableFileIsDataAndTheOthersAreDone(t *testing.T) {
	bodySite(t)
	write(t, "docs/a.md", "---\ntitle: never closed\n\n[About](../about.html)\n")
	write(t, "docs/b.md", migrateDoc("[About](../about.html)\n"))
	out, err := runMigrate(t, "-w", "docs")
	wantClass(t, "an unparseable file", err, classData)
	if strings.Contains(readFile(t, "docs/b.md"), "about.html") {
		t.Errorf("the good file was not migrated:\n%s", out)
	}
	if !strings.Contains(readFile(t, "docs/a.md"), "about.html") {
		t.Error("the unparseable file was changed")
	}
}

// DR-0017 9: a failed write is 74.
func TestMigrateLinksFailedWriteIsIO(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can write anywhere")
	}
	bodySite(t)
	write(t, "docs/p.md", migrateDoc("[About](../about.html)\n"))
	os.Chmod("docs", 0o555)
	t.Cleanup(func() { os.Chmod("docs", 0o775) })
	_, err := runMigrate(t, "-w", "docs/p.md")
	wantClass(t, "a read-only folder", err, classIO)
}

// The pair works together: after -w, validate has no link-html left.
func TestMigrateLinksClearsWhatValidateFound(t *testing.T) {
	bodySite(t)
	write(t, "docs/p.md", migrateDoc("[About](../about.html) and [News](../news.html)\n"))
	out, _ := runValidate(t, "docs/p.md")
	if strings.Count(out, "link-html") != 2 {
		t.Fatalf("validate should find two link-html:\n%s", out)
	}
	if _, err := runMigrate(t, "-w", "docs/p.md"); err != nil {
		t.Fatal(err)
	}
	out, _ = runValidate(t, "docs/p.md")
	if strings.Contains(out, "link-html") {
		t.Errorf("validate still finds link-html after migrate-links -w:\n%s", out)
	}
}
