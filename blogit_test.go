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
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const blogitDoc = "---\ntitle: A post\n---\n\nbody\n"

func writeBlogitDoc(t *testing.T, name string) {
	t.Helper()
	if err := os.WriteFile(name, []byte(blogitDoc), 0o644); err != nil {
		t.Fatal(err)
	}
}

// blogit takes [COLLECTION_NAME] FILEPATH [POST_DATE]. Whether the second of
// two arguments is a date or a file is decided by whether it is a date, not by
// whether it contains a hyphen: hyphenated file names are everyday.
func TestBlogIt_ArgumentForms(t *testing.T) {
	today := time.Now().Format("2006/01/02")
	cases := []struct {
		label string
		args  []string
		dir   string // where the post must land under blog/
	}{
		{"file only", []string{"plain.md"}, today},
		{"hyphenated file only", []string{"my-post.md"}, today},
		{"collection and hyphenated file", []string{"pages.md", "my-post.md"}, today},
		{"file and date", []string{"my-post.md", "2026-01-02"}, "2026/01/02"},
		{"collection, file and date", []string{"pages.md", "my-post.md", "2026-01-02"}, "2026/01/02"},
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			newWorkspace(t, true)
			file := c.args[0]
			if len(c.args) >= 2 && strings.HasSuffix(c.args[1], ".md") {
				file = c.args[1]
			}
			writeBlogitDoc(t, file)
			if err := runIn(t, "blogit", c.args...); err != nil {
				t.Fatalf("blogit %v: %s", c.args, err)
			}
			want := filepath.Join("blog", filepath.FromSlash(c.dir), file)
			if _, err := os.Stat(want); err != nil {
				t.Errorf("blogit %v: expected %s: %s", c.args, want, err)
			}
		})
	}
}

// A date that is shaped like one but is not a real date is the user's typo,
// and the message names the argument that failed to parse, not another one.
func TestBlogIt_BadDateIsUsageAndNamesTheDate(t *testing.T) {
	cases := [][]string{
		{"my-post.md", "2026-13-45"},
		{"pages.md", "my-post.md", "2026-13-45"},
	}
	for _, args := range cases {
		newWorkspace(t, true)
		writeBlogitDoc(t, "my-post.md")
		err := runIn(t, "blogit", args...)
		wantClass(t, "blogit "+strings.Join(args, " "), err, classUsage)
		if err != nil && !strings.Contains(err.Error(), `failed to parse "2026-13-45" as post date`) {
			t.Errorf("blogit %v: the message should say it failed to parse the date %q, got: %v", args, "2026-13-45", err)
		}
	}
}

// DR-0016: blogit writes what a published post needs. The expected dates are
// worked out here in the local zone, independently of the code.

// stderrOf runs f and returns what it wrote to standard error.
func stderrOf(t *testing.T, f func()) string {
	t.Helper()
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w
	f()
	w.Close()
	os.Stderr = old
	b, _ := io.ReadAll(r)
	return string(b)
}

func readFile(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// noon is 12:00 local on the given day in the form BonBon expects.
func noon(y int, m time.Month, d int) string {
	return time.Date(y, m, d, 12, 0, 0, 0, time.Local).Format("Mon, 02 Jan 2006 15:04:05 -0700")
}

func TestBlogIt_AddsPubDateAndTitleFromTheHeading(t *testing.T) {
	newWorkspace(t, true)
	os.WriteFile("hello.md", []byte("# Hello *there*\n\nA first post.\n"), 0o644)
	if err := runIn(t, "blogit", "hello.md", "2026-11-19"); err != nil {
		t.Fatalf("blogit: %s", err)
	}
	got := readFile(t, "blog/2026/11/19/hello.md")
	if !strings.Contains(got, "pubDate: "+noon(2026, time.November, 19)) && !strings.Contains(got, `pubDate: "`+noon(2026, time.November, 19)+`"`) {
		t.Errorf("no pubDate %q in:\n%s", noon(2026, time.November, 19), got)
	}
	if !strings.Contains(got, "title: Hello there") {
		t.Errorf("title not taken from the heading as plain text:\n%s", got)
	}
	if !strings.Contains(got, "# Hello *there*\n\nA first post.") {
		t.Errorf("the body changed:\n%s", got)
	}
	for _, key := range []string{"author:", "description:", "link:", "urlPrev:", "urlNext:", "rightMarginImage:"} {
		if strings.Contains(got, key) {
			t.Errorf("blogit wrote %s:\n%s", key, got)
		}
	}
}

// The post is published, so a Markdown reader finds it in the table of contents.
func TestBlogIt_PostIsPublishedAndInTheTOC(t *testing.T) {
	newWorkspace(t, true)
	os.WriteFile("hello.md", []byte("# Hello\n\nA first post.\n"), 0o644)
	if err := runIn(t, "blogit", "hello.md", "2026-11-19"); err != nil {
		t.Fatalf("blogit: %s", err)
	}
	db, err := sql.Open("sqlite", "pages.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var status, pub string
	if err := db.QueryRow(`SELECT status, pubDate FROM items WHERE postPath = 'blog/2026/11/19/hello.md'`).Scan(&status, &pub); err != nil {
		t.Fatal(err)
	}
	if status != "published" || pub == "" {
		t.Errorf("status %q pubDate %q, want a published post with a date", status, pub)
	}
	if err := runIn(t, "generate"); err != nil {
		t.Fatalf("generate: %s", err)
	}
	if toc := readFile(t, ".well-known/toc.opml"); !strings.Contains(toc, "blog/2026/11/19/hello.md") {
		t.Errorf("the post is not in toc.opml:\n%s", toc)
	}
}

// With no POST_DATE the post is dated now, in the same local day as its directory.
func TestBlogIt_NoDateMeansNow(t *testing.T) {
	newWorkspace(t, true)
	os.WriteFile("now.md", []byte("# Now\n\nbody\n"), 0o644)
	before := time.Now().Add(-time.Second)
	if err := runIn(t, "blogit", "now.md"); err != nil {
		t.Fatalf("blogit: %s", err)
	}
	after := time.Now().Add(time.Second)
	got := readFile(t, filepath.Join("blog", time.Now().Format("2006/01/02"), "now.md"))
	var raw string
	for _, line := range strings.Split(got, "\n") {
		if strings.HasPrefix(line, "pubDate:") {
			raw = strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, "pubDate:")), `"'`)
		}
	}
	when, err := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", raw)
	if err != nil {
		t.Fatalf("pubDate %q is not the RSS form with a numeric offset: %s\n%s", raw, err, got)
	}
	if when.Before(before.Truncate(time.Second)) || when.After(after) {
		t.Errorf("pubDate %s is not now (%s to %s)", when, before, after)
	}
}

// What the author already wrote is never changed.
func TestBlogIt_NeverChangesAnExistingPubDateOrTitle(t *testing.T) {
	for label, front := range map[string]string{
		"pubDate":       "pubDate: \"2020-01-02\"",
		"datePublished": "datePublished: \"2020-01-02\"",
	} {
		newWorkspace(t, true)
		os.WriteFile("keep.md", []byte("---\ntitle: Kept title\n"+front+"\n---\n\n# A different heading\n\nbody\n"), 0o644)
		if err := runIn(t, "blogit", "keep.md", "2020-01-02"); err != nil {
			t.Fatalf("%s: blogit: %s", label, err)
		}
		got := readFile(t, "blog/2020/01/02/keep.md")
		if !strings.Contains(got, "title: Kept title") || strings.Contains(got, "A different heading\n\n") && strings.Count(got, "title:") != 1 {
			t.Errorf("%s: the title changed:\n%s", label, got)
		}
		if strings.Count(got, "pubDate:") > 1 || (label == "datePublished" && strings.Contains(got, "pubDate:")) {
			t.Errorf("%s: a second date was added:\n%s", label, got)
		}
		if !strings.Contains(got, "2020-01-02") {
			t.Errorf("%s: the date was lost:\n%s", label, got)
		}
	}
}

// A POST_DATE on another day than the file's own date warns once; the file's
// date stays and the directory follows POST_DATE, as asked.
func TestBlogIt_DateDisagreementWarnsAndKeepsTheFilesDate(t *testing.T) {
	newWorkspace(t, true)
	os.WriteFile("late.md", []byte("---\ntitle: Late\npubDate: \"2020-01-02\"\n---\n\nbody\n"), 0o644)
	var err error
	warn := stderrOf(t, func() { err = runIn(t, "blogit", "late.md", "2026-11-19") })
	if err != nil {
		t.Fatalf("blogit: %s", err)
	}
	if strings.Count(warn, "2020-01-02") < 1 || !strings.Contains(warn, "2026-11-19") {
		t.Errorf("no warning naming both dates: %q", warn)
	}
	got := readFile(t, "blog/2026/11/19/late.md")
	if !strings.Contains(got, "2020-01-02") {
		t.Errorf("the file's own date was changed:\n%s", got)
	}
}

// No title, no heading: nothing invented, and post reports it as before (65).
func TestBlogIt_NoTitleAndNoHeadingIsReported(t *testing.T) {
	newWorkspace(t, true)
	os.WriteFile("bare.md", []byte(""), 0o644)
	wantClass(t, "blogit of an empty file", runIn(t, "blogit", "bare.md", "2026-11-19"), classData)
	if got := readFile(t, "blog/2026/11/19/bare.md"); strings.Contains(got, "title:") {
		t.Errorf("a title was invented:\n%s", got)
	}
}

// Running blogit on its own result changes nothing.
func TestBlogIt_IsIdempotent(t *testing.T) {
	newWorkspace(t, true)
	os.WriteFile("hello.md", []byte("# Hello\n\nbody\n"), 0o644)
	if err := runIn(t, "blogit", "hello.md", "2026-11-19"); err != nil {
		t.Fatal(err)
	}
	first := readFile(t, "blog/2026/11/19/hello.md")
	if err := runIn(t, "blogit", "blog/2026/11/19/hello.md", "2026-11-19"); err != nil {
		t.Fatalf("second blogit: %s", err)
	}
	if second := readFile(t, "blog/2026/11/19/hello.md"); second != first {
		t.Errorf("a second run changed the file:\n--- first\n%s\n--- second\n%s", first, second)
	}
}
