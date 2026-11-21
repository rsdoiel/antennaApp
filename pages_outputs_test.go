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
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// DR-0024: the pages table records every output and include of a page.

// runOut runs a verb and returns what it wrote to standard output and to the
// error stream.
func runOut(t *testing.T, action string, args ...string) (string, string, error) {
	t.Helper()
	app := &AntennaApp{appName: "antenna"}
	var out, eout bytes.Buffer
	err := app.Run(strings.NewReader(""), &out, &eout, "antenna.yaml", action, args)
	return out.String(), eout.String(), err
}

func openPagesForTest(t *testing.T) *sql.DB {
	t.Helper()
	cfg := &AppConfig{}
	if err := cfg.LoadConfig("antenna.yaml"); err != nil {
		t.Fatal(err)
	}
	col, err := cfg.GetCollection("pages.md")
	if err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", col.DbName)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func rowFor(t *testing.T, input string) pageRow {
	t.Helper()
	rows, err := queryPageRows(openPagesForTest(t), "inputPath")
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range rows {
		if r.Input == input {
			return r
		}
	}
	t.Fatalf("no pages row for %s in %v", input, rows)
	return pageRow{}
}

func jsonOf(v []string) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func TestPageRecordsItsOutputsAndIncludes(t *testing.T) {
	markdownPageSite(t)
	if err := runIn(t, "page", "index.txt", "index.md", "index.html"); err != nil {
		t.Fatal(err)
	}
	r := rowFor(t, "index.txt")
	if jsonOf(r.Outputs) != `["index.md","index.html"]` {
		t.Errorf("outputs in command-line order: %s", jsonOf(r.Outputs))
	}
	if r.Output != "index.html" {
		t.Errorf("outputPath is the first .html: %q", r.Output)
	}
	if jsonOf(r.Includes) != `["author.md","blog/recent.txt","hello.go"]` {
		t.Errorf("includes in the order met: %s", jsonOf(r.Includes))
	}
}

func TestMarkdownOnlyRunIsRecorded(t *testing.T) {
	markdownPageSite(t)
	if err := runIn(t, "page", "index.txt", "index.md"); err != nil {
		t.Fatal(err)
	}
	r := rowFor(t, "index.txt")
	if jsonOf(r.Outputs) != `["index.md"]` || r.Output != "" {
		t.Errorf("a Markdown-only run is a row: outputs %s outputPath %q", jsonOf(r.Outputs), r.Output)
	}
}

func TestTheLastRunWins(t *testing.T) {
	markdownPageSite(t)
	runIn(t, "page", "index.txt", "index.md", "index.html")
	runIn(t, "page", "index.txt", "index.html")
	if r := rowFor(t, "index.txt"); jsonOf(r.Outputs) != `["index.html"]` {
		t.Errorf("the last run replaces the outputs: %s", jsonOf(r.Outputs))
	}
}

func TestSeveralOutputsOfEachKindAreAllRecorded(t *testing.T) {
	markdownPageSite(t)
	os.MkdirAll("site", 0o775)
	if err := runIn(t, "page", "index.txt", "index.md", "site/home.md", "index.html", "site/home.html"); err != nil {
		t.Fatal(err)
	}
	if r := rowFor(t, "index.txt"); jsonOf(r.Outputs) != `["index.md","site/home.md","index.html","site/home.html"]` || r.Output != "index.html" {
		t.Errorf("outputs %s outputPath %q", jsonOf(r.Outputs), r.Output)
	}
	for _, f := range []string{"index.md", "site/home.md", "index.html", "site/home.html"} {
		if _, err := os.Stat(f); err != nil {
			t.Errorf("%s was not written", f)
		}
	}
}

// An old database has no outputs or includes columns.
func oldPagesDB(t *testing.T) {
	t.Helper()
	newWorkspace(t, true)
	db := openPagesForTest(t)
	for _, stmt := range []string{
		`DROP TABLE pages`,
		`CREATE TABLE pages (inputPath PRIMARY KEY, outputPath TEXT DEFAULT '', updated DATETIME)`,
		`INSERT INTO pages VALUES ('about.md', 'about.html', '2026-11-01T10:00:00-08:00')`,
	} {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatal(err)
		}
	}
	os.WriteFile("about.md", []byte("---\ntitle: About\n---\nAbout.\n"), 0o664)
	os.WriteFile("about.html", []byte("<html></html>\n"), 0o664)
}

func TestPagesReadsAnOldDatabaseWithoutChangingIt(t *testing.T) {
	oldPagesDB(t)
	out, _, err := runOut(t, "pages")
	if err != nil {
		t.Fatal(err)
	}
	fields := strings.Split(strings.TrimRight(out, "\n"), "\t")
	if len(fields) != 6 || fields[0] != "about.md" || fields[1] != "about.html" || fields[3] != `["about.html"]` || fields[4] != "" {
		t.Errorf("an old row reads as outputs [outputPath]: %q", fields)
	}
	cols := map[string]bool{}
	rows, _ := openPagesForTest(t).Query(`SELECT name FROM pragma_table_info('pages')`)
	for rows.Next() {
		var n string
		rows.Scan(&n)
		cols[n] = true
	}
	rows.Close()
	if cols["outputs"] || cols["includes"] {
		t.Errorf("a read-only verb must not change the schema: %v", cols)
	}
}

func TestPageAndGenerateUpgradeAnOldDatabase(t *testing.T) {
	oldPagesDB(t)
	if err := runIn(t, "generate"); err != nil {
		t.Fatalf("generate: %s", err)
	}
	cols := map[string]bool{}
	rows, _ := openPagesForTest(t).Query(`SELECT name FROM pragma_table_info('pages')`)
	for rows.Next() {
		var n string
		rows.Scan(&n)
		cols[n] = true
	}
	rows.Close()
	if !cols["outputs"] || !cols["includes"] {
		t.Fatalf("generate should add the columns: %v", cols)
	}
	r := rowFor(t, "about.md")
	if jsonOf(r.Outputs) != `["about.html"]` {
		t.Errorf("outputs backfilled from outputPath, no row lost: %s", jsonOf(r.Outputs))
	}
	// And a second run changes nothing about the schema or the row.
	if err := runIn(t, "generate"); err != nil {
		t.Fatal(err)
	}
	if r2 := rowFor(t, "about.md"); jsonOf(r2.Outputs) != jsonOf(r.Outputs) {
		t.Errorf("a second run changed the row")
	}
}

func TestABadOutputsValueIsADataErrorNamingTheRow(t *testing.T) {
	oldPagesDB(t)
	runIn(t, "generate") // upgrades
	db := openPagesForTest(t)
	if _, err := db.Exec(`UPDATE pages SET outputs = 'not json' WHERE inputPath = 'about.md'`); err != nil {
		t.Fatal(err)
	}
	_, _, err := runOut(t, "pages")
	wantClass(t, "bad outputs", err, classData)
	if err == nil || !strings.Contains(err.Error(), "about.md") {
		t.Errorf("the error names the row: %v", err)
	}
}

func TestTOCListsTheRecordedMarkdownWhateverItIsCalled(t *testing.T) {
	newWorkspace(t, true)
	os.MkdirAll("docs", 0o775)
	os.WriteFile("src.txt", []byte("---\ntitle: Notes\n---\n# Notes\n"), 0o664)
	if err := runIn(t, "page", "src.txt", "docs/notes.md", "docs/home.html"); err != nil {
		t.Fatal(err)
	}
	var eout bytes.Buffer
	if err := runInBuf(t, &eout, "generate"); err != nil {
		t.Fatalf("generate: %s\n%s", err, eout.String())
	}
	toc := readText(t, ".well-known/toc.opml")
	if !strings.Contains(toc, "docs/notes.md") || strings.Contains(toc, "docs/home.md") {
		t.Errorf("the toc lists the recorded Markdown, not a guess from the HTML name:\n%s", toc)
	}
	if strings.Contains(eout.String(), "not listed") {
		t.Errorf("no warning expected:\n%s", eout.String())
	}
}

func TestLinksToARecordedMarkdownCopyBecomeHTML(t *testing.T) {
	newWorkspace(t, true)
	os.MkdirAll("docs", 0o775)
	os.WriteFile("src.txt", []byte("# Notes\n"), 0o664)
	runIn(t, "page", "src.txt", "docs/notes.md", "docs/home.html")
	os.WriteFile("other.md", []byte("[Notes](docs/notes.md)\n"), 0o664)
	if err := runIn(t, "page", "other.md", "other.html"); err != nil {
		t.Fatal(err)
	}
	if html := readText(t, "other.html"); !strings.Contains(html, `href="docs/home.html"`) {
		t.Errorf("a link to the copy lands on the page:\n%s", html)
	}
}

func TestTheHeadDeclaresACopyInAnotherFolderRelatively(t *testing.T) {
	newWorkspace(t, true)
	os.MkdirAll("docs", 0o775)
	os.WriteFile("src.txt", []byte("# Notes\n"), 0o664)
	if err := runIn(t, "page", "src.txt", "md/notes.md", "docs/home.html"); err != nil {
		t.Fatal(err)
	}
	if hrefs, _ := markdownAlternates(readText(t, "docs/home.html")); len(hrefs) != 1 || hrefs[0] != "../md/notes.md" {
		t.Errorf("relative to the HTML: %v", hrefs)
	}
	// generate, which reads the row, writes the same link.
	before := readText(t, "docs/home.html")
	if err := runIn(t, "generate"); err != nil {
		t.Fatal(err)
	}
	if hrefs, _ := markdownAlternates(readText(t, "docs/home.html")); len(hrefs) != 1 || hrefs[0] != "../md/notes.md" {
		t.Errorf("generate dropped or changed the link: %v\n%s", hrefs, before)
	}
}

func TestGenerateRewritesTheMarkdownAfterAFragmentChanges(t *testing.T) {
	markdownPageSite(t)
	runIn(t, "page", "index.txt", "index.md", "index.html")
	if !strings.Contains(readText(t, "index.md"), "- [Post one](/blog/one.md)") {
		t.Fatal("setup")
	}
	os.WriteFile("blog/recent.txt", []byte("- [Post two](/blog/two.md)\n"), 0o664)
	if err := runIn(t, "generate"); err != nil {
		t.Fatalf("generate: %s", err)
	}
	if md := readText(t, "index.md"); !strings.Contains(md, "Post two") || strings.Contains(md, "Post one") {
		t.Errorf("generate should rewrite the Markdown copy:\n%s", md)
	}
	if err := runIn(t, "generate"); err != nil {
		t.Fatal(err)
	}
	first := readText(t, "index.md")
	runIn(t, "generate")
	if readText(t, "index.md") != first {
		t.Errorf("a second generate changed the copy")
	}
}

func TestGenerateRefreshesIncludes(t *testing.T) {
	markdownPageSite(t)
	runIn(t, "page", "index.txt", "index.md", "index.html")
	os.WriteFile("index.txt", []byte("---\ntitle: Home\n---\n\n@include-text-block author.md\n"), 0o664)
	runIn(t, "generate")
	if r := rowFor(t, "index.txt"); jsonOf(r.Includes) != `["author.md"]` {
		t.Errorf("includes follow the source: %s", jsonOf(r.Includes))
	}
}

func TestNestedIncludesAreRecordedAndTooDeepWritesNothing(t *testing.T) {
	newWorkspace(t, true)
	writeAll(t, chainOf(3))
	os.WriteFile("top.txt", []byte("---\ntitle: Top\n---\n\n@include-text-block l1.txt\n"), 0o664)
	if err := runIn(t, "page", "top.txt", "top.md", "top.html"); err != nil {
		t.Fatal(err)
	}
	if r := rowFor(t, "top.txt"); jsonOf(r.Includes) != `["l1.txt","l2.txt","l3.txt"]` {
		t.Errorf("nested includes are recorded: %s", jsonOf(r.Includes))
	}
	writeAll(t, chainOf(6))
	os.WriteFile("deep.txt", []byte("---\ntitle: Deep\n---\n\n@include-text-block l1.txt\n"), 0o664)
	err := runIn(t, "page", "deep.txt", "deep.md", "deep.html")
	wantClass(t, "six levels", err, classData)
	for _, f := range []string{"deep.md", "deep.html"} {
		if _, e := os.Stat(f); e == nil {
			t.Errorf("%s must not be written for a page that cannot be expanded", f)
		}
	}
}

// ---- antenna pages ----

func TestPagesPrintsOutputsIncludesAndStale(t *testing.T) {
	markdownPageSite(t)
	runIn(t, "page", "index.txt", "index.md", "index.html")
	out, _, err := runOut(t, "pages")
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		f := strings.Split(line, "\t")
		if len(f) != 6 {
			t.Fatalf("six tab-separated fields: %q", line)
		}
		if f[0] == "index.txt" {
			found = true
			if f[1] != "index.html" || f[3] != `["index.md","index.html"]` || f[4] != `["author.md","blog/recent.txt","hello.go"]` || f[5] != "" {
				t.Errorf("fields: %q", f)
			}
		}
	}
	if !found {
		t.Errorf("index.txt not listed:\n%s", out)
	}
}

func setMTime(t *testing.T, name string, when time.Time) {
	t.Helper()
	if err := os.Chtimes(name, when, when); err != nil {
		t.Fatal(err)
	}
}

func TestStaleWhenAnOutputIsOlderThanASourceOrANestedInclude(t *testing.T) {
	newWorkspace(t, true)
	writeAll(t, chainOf(2))
	os.WriteFile("top.txt", []byte("---\ntitle: Top\n---\n\n@include-text-block l1.txt\n"), 0o664)
	if err := runIn(t, "page", "top.txt", "top.md", "top.html"); err != nil {
		t.Fatal(err)
	}
	stale := func() string {
		out, _, err := runOut(t, "pages")
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
			if f := strings.Split(line, "\t"); f[0] == "top.txt" {
				return f[5]
			}
		}
		return "?"
	}
	if got := stale(); got != "" {
		t.Fatalf("just built: %q", got)
	}
	_, _, err := runOut(t, "pages", "-stale")
	if err != nil {
		t.Errorf("nothing stale is exit 0: %v", err)
	}
	future := time.Now().Add(time.Hour)
	setMTime(t, "l2.txt", future) // a nested include changes
	if got := stale(); got != "stale" {
		t.Errorf("an include newer than an output: %q", got)
	}
	out, _, err := runOut(t, "pages", "-stale")
	wantClass(t, "pages -stale with a stale page", err, classNegative)
	if !strings.Contains(out, "top.txt") {
		t.Errorf("-stale lists the page:\n%s", out)
	}
	// Rebuilding clears it, even though the output's bytes do not change.
	setMTime(t, "l2.txt", time.Now().Add(-time.Minute))
	runIn(t, "page", "top.txt", "top.md", "top.html")
	if got := stale(); got != "" {
		t.Errorf("after a rebuild: %q", got)
	}
	os.Remove("top.md")
	if got := stale(); got != "missing" {
		t.Errorf("an absent output: %q", got)
	}
}

// ---- validate ----

func TestValidateRowOutputMissingDuplicateAndIncludeMissing(t *testing.T) {
	markdownPageSite(t)
	runIn(t, "page", "index.txt", "index.md", "index.html")
	os.Remove("index.md")
	// A second row that writes the same HTML, as after a rename.
	os.WriteFile("old.md", []byte("# Old\n"), 0o664)
	runIn(t, "page", "old.md", "index.html")
	// An include that is gone.
	os.WriteFile("gone.txt", []byte("---\ntitle: Gone\n---\n\n@include-text-block nothing-here.txt\n"), 0o664)
	runIn(t, "page", "gone.txt", "gone.html")
	out, _ := runValidate(t)
	if _, ok := finding(out, "index.txt", "warning", "row-output-missing"); !ok || !strings.Contains(out, "index.md") {
		t.Errorf("row-output-missing:\n%s", out)
	}
	if _, ok := finding(out, "index.html", "warning", "row-duplicate-output"); !ok {
		t.Errorf("row-duplicate-output:\n%s", out)
	}
	if line, ok := finding(out, "gone.txt", "warning", "include-missing"); !ok || !strings.Contains(line, "nothing-here.txt") {
		t.Errorf("include-missing:\n%s", out)
	}
}

func TestValidateOutputStaleIsInfo(t *testing.T) {
	markdownPageSite(t)
	runIn(t, "page", "index.txt", "index.md", "index.html")
	future := time.Now().Add(time.Hour)
	setMTime(t, filepath.Join("blog", "recent.txt"), future)
	out, _ := runValidate(t)
	if _, ok := finding(out, "index.txt", "info", "output-stale"); !ok {
		t.Errorf("output-stale:\n%s", out)
	}
}
