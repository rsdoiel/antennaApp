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
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"
)

func TestParseStoredDateAcceptsRSSForms(t *testing.T) {
	want := time.Date(2026, 10, 23, 15, 19, 45, 0, time.UTC)
	for _, raw := range []string{
		"Fri, 23 Oct 2026 15:19:45 GMT",
		"Fri, 23 Oct 2026 15:19:45 +0000",
		"Fri, 23 Oct 2026 17:19:45 +0200",
		"Fri, 23 Oct 2026 15:19:45 UT",
		"23 Oct 2026 15:19:45 GMT",
	} {
		got, ok := parseStoredDate(raw)
		if !ok || !got.Equal(want) {
			t.Errorf("parseStoredDate(%q) = %v, %v; want %v", raw, got, ok, want)
		}
	}
	// day without a leading zero, as many feeds write it
	got, ok := parseStoredDate("Mon, 5 Oct 2026 09:00:00 GMT")
	if !ok || got.Day() != 5 || got.Month() != time.October {
		t.Errorf("unpadded day: %v %v", got, ok)
	}
	// existing forms still work
	for _, raw := range []string{"2026-10-23", "2026-10-23 15:19:45", "2026-10-23T15:19:45Z"} {
		if _, ok := parseStoredDate(raw); !ok {
			t.Errorf("%q no longer parses", raw)
		}
	}
	if _, ok := parseStoredDate("next Tuesday"); ok {
		t.Error("garbage parsed")
	}
	if formatItemDate("Fri, 23 Oct 2026 15:19:45 GMT", "2006-01-02") != "2026-10-23" {
		t.Error("formatItemDate does not read the RSS form")
	}
}

func TestGetAttributeDate(t *testing.T) {
	cases := []struct {
		name, fm  string
		stored    string
		found, ok bool
	}{
		{"quoted date kept as written", `pubDate: "2026-06-27"`, "2026-06-27", true, true},
		{"unquoted date is a date, not unset", `pubDate: 2026-06-27`, "2026-06-27", true, true},
		{"unquoted midnight timestamp is its date", `pubDate: 2022-09-28T00:00:00.000Z`, "2022-09-28", true, true},
		{"unquoted timestamp with an offset keeps it (DR-0018)", `pubDate: 2026-06-27T10:00:00+02:00`, "2026-06-27T10:00:00+02:00", true, true},
		{"RSS form becomes the stored form", `pubDate: Fri, 23 Oct 2026 15:19:45 GMT`, "2026-10-23 15:19:45", true, true},
		{"RSS form with an offset keeps it (DR-0018)", `pubDate: Fri, 23 Oct 2026 17:19:45 +0200`, "2026-10-23T17:19:45+02:00", true, true},
		{"stored datetime kept as written", `pubDate: "2026-10-23 15:19:45"`, "2026-10-23 15:19:45", true, true},
		{"a date nobody can read is found and not ok", `pubDate: next Tuesday`, "next Tuesday", true, false},
		{"absent", `title: x`, "", false, false},
		{"empty", `pubDate: ""`, "", false, false},
	}
	for _, c := range cases {
		doc := &CommonMark{}
		if err := doc.Parse([]byte("---\n" + c.fm + "\n---\nbody\n")); err != nil {
			t.Fatalf("%s: %s", c.name, err)
		}
		stored, found, ok := doc.GetAttributeDate("pubDate")
		if stored != c.stored || found != c.found || ok != c.ok {
			t.Errorf("%s: got (%q, %v, %v), want (%q, %v, %v)", c.name, stored, found, ok, c.stored, c.found, c.ok)
		}
	}
}

// The real path: antenna post stores the date in the form every reader of the
// items table expects, so posts sort and publish, whichever form the author wrote.
func TestPostStoresPubDateInStoredForm(t *testing.T) {
	cfg := sitemapSite(t, "news.md")
	cfg.Htdocs = "htdocs"
	cfg.Generator = "page.yaml"
	cfg.Collections[0].Generator = "page.yaml"
	if err := os.MkdirAll("htdocs/p", 0775); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("page.yaml", []byte("header: |\n  <h1>Test</h1>\n"), 0664); err != nil {
		t.Fatal(err)
	}
	cases := []struct{ file, pubDate, wantDate, wantStatus string }{
		{"a", "Fri, 23 Oct 2026 15:19:45 GMT", "2026-10-23 15:19:45", "published"},
		{"b", "2026-06-27", "2026-06-27", "published"}, // unquoted: was read as unset, so a draft
		{"c", `"2026-06-28"`, "2026-06-28", "published"},
		{"d", "2022-09-28T00:00:00.000Z", "2022-09-28", "published"},
	}
	for _, c := range cases {
		src := "---\ntitle: Post " + c.file + "\npostPath: p/" + c.file + ".md\npubDate: " + c.pubDate + "\n---\nBody.\n"
		if err := os.WriteFile(c.file+".md", []byte(src), 0664); err != nil {
			t.Fatal(err)
		}
		if err := cfg.Post("news.md", c.file+".md"); err != nil {
			t.Fatalf("Post %s: %s", c.file, err)
		}
		db, err := sql.Open("sqlite", "news.db")
		if err != nil {
			t.Fatal(err)
		}
		var gotDate, gotStatus string
		err = db.QueryRow(`SELECT pubDate, status FROM items WHERE postPath = ?`, "p/"+c.file+".md").Scan(&gotDate, &gotStatus)
		db.Close()
		if err != nil {
			t.Fatalf("%s: %s", c.file, err)
		}
		// the pure-Go driver hands a DATETIME column back as RFC 3339, so compare instants
		gotT, gotOK := parseStoredDate(gotDate)
		wantT, _ := parseStoredDate(c.wantDate)
		if !gotOK || !gotT.Equal(wantT) || gotStatus != c.wantStatus {
			t.Errorf("%s (%s): stored (%q, %q), want (%q, %q)", c.file, c.pubDate, gotDate, gotStatus, c.wantDate, c.wantStatus)
		}
	}
}

// A date antenna cannot read does not fail the command: it is kept as written,
// as before, and reported on standard error.
func TestPostWithUnreadableDateWarnsAndStillWorks(t *testing.T) {
	cfg := sitemapSite(t, "news.md")
	cfg.Htdocs = "htdocs"
	cfg.Generator = "page.yaml"
	cfg.Collections[0].Generator = "page.yaml"
	if err := os.MkdirAll("htdocs/p", 0775); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("page.yaml", []byte("header: |\n  <h1>Test</h1>\n"), 0664); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("x.md", []byte("---\ntitle: X\npostPath: p/x.md\npubDate: next Tuesday\n---\nBody.\n"), 0664); err != nil {
		t.Fatal(err)
	}
	stderr := captureStderr(t, func() {
		if err := cfg.Post("news.md", "x.md"); err != nil {
			t.Fatalf("Post: %s", err)
		}
	})
	if !strings.Contains(stderr, "warning:") || !strings.Contains(stderr, "next Tuesday") {
		t.Errorf("stderr = %q", stderr)
	}
}

// captureStderr returns what fn writes to os.Stderr.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w
	defer func() { os.Stderr = old }()
	done := make(chan string)
	go func() {
		var sb strings.Builder
		buf := make([]byte, 4096)
		for {
			n, err := r.Read(buf)
			sb.Write(buf[:n])
			if err != nil {
				break
			}
		}
		done <- sb.String()
	}()
	fn()
	w.Close()
	os.Stderr = old
	return <-done
}
