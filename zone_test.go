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
	"strings"
	"testing"
	"time"
)

// DR-0018: a date with a zone is stored with its zone; everything else is as
// it was. Each case is a front matter value as the author writes it.
func TestStoredDateKeepsTheAuthorsZone(t *testing.T) {
	cases := []struct {
		label string
		value string // the YAML after "pubDate: "
		want  string
	}{
		{"RSS with an offset", `"Mon, 30 Nov 2026 21:30:00 -0800"`, "2026-11-30T21:30:00-08:00"},
		{"RSS with a zone word", `"Mon, 30 Nov 2026 21:30:00 PST"`, "2026-11-30T21:30:00-08:00"},
		{"RSS east of Greenwich", `"Tue, 01 Dec 2026 06:30:00 +0900"`, "2026-12-01T06:30:00+09:00"},
		{"RSS in GMT is as before", `"Mon, 30 Nov 2026 21:30:00 GMT"`, "2026-11-30 21:30:00"},
		{"RSS with +0000 is as before", `"Mon, 30 Nov 2026 21:30:00 +0000"`, "2026-11-30 21:30:00"},
		{"unquoted timestamp with an offset", `2026-11-30T21:30:00-08:00`, "2026-11-30T21:30:00-08:00"},
		{"unquoted timestamp without a zone is as before", `2026-11-30 21:30:00`, "2026-11-30 21:30:00"},
		{"unquoted date is as before", `2026-11-30`, "2026-11-30"},
		{"quoted date is kept", `"2026-11-30"`, "2026-11-30"},
		{"quoted datetime is kept", `"2026-11-30 21:30:00"`, "2026-11-30 21:30:00"},
		{"quoted RFC 3339 is kept exactly", `"2026-11-30T21:30:00-08:00"`, "2026-11-30T21:30:00-08:00"},
	}
	for _, c := range cases {
		doc := &CommonMark{}
		if err := doc.Parse([]byte("---\ntitle: T\npubDate: " + c.value + "\n---\nbody\n")); err != nil {
			t.Fatalf("%s: %s", c.label, err)
		}
		got, found, ok := doc.GetAttributeDate("pubDate")
		if !found || !ok || got != c.want {
			t.Errorf("%s: GetAttributeDate = %q found=%v ok=%v, want %q", c.label, got, found, ok, c.want)
		}
	}
}

// DR-0018 2: a post made at 21:30 Pacific on the last evening of the month
// stays in that month and day in the table of contents and in the RSS feed,
// through the real post, generate and rss actions.
func TestAnEveningPostStaysOnItsOwnDay(t *testing.T) {
	newWorkspace(t, true)
	os.MkdirAll("blog/2026/11/30", 0o775)
	src := "---\ntitle: Late post\npubDate: \"Mon, 30 Nov 2026 21:30:00 -0800\"\n---\n\nBody.\n"
	if err := os.WriteFile("blog/2026/11/30/late.md", []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runIn(t, "post", "blog/2026/11/30/late.md"); err != nil {
		t.Fatalf("post: %s", err)
	}
	if err := runIn(t, "generate"); err != nil {
		t.Fatalf("generate: %s", err)
	}
	toc := readFile(t, ".well-known/toc.opml")
	if !strings.Contains(toc, `<outline text="November 2026"`) || strings.Contains(toc, `December 2026`) {
		t.Errorf("the post is not in November in toc.opml:\n%s", toc)
	}
	if err := runIn(t, "rss", "pages.md", "posts.xml"); err != nil {
		t.Fatalf("rss: %s", err)
	}
	if feed := readFile(t, "posts.xml"); !strings.Contains(feed, "<pubDate>Mon, 30 Nov 2026 21:30:00 -0800</pubDate>") {
		t.Errorf("the feed lost the author's zone:\n%s", feed)
	}
}

// DR-0018 3: blogit files a post under the day in the zone of the process, and
// writes that zone's offset. The test sets time.Local, which Go takes from TZ.
func TestBlogItUsesTheProcessZone(t *testing.T) {
	tokyo := time.FixedZone("JST", 9*60*60)
	old := time.Local
	time.Local = tokyo
	t.Cleanup(func() { time.Local = old })

	newWorkspace(t, true)
	os.WriteFile("trip.md", []byte("# Trip\n\nbody\n"), 0o644)
	if err := runIn(t, "blogit", "trip.md"); err != nil {
		t.Fatalf("blogit: %s", err)
	}
	day := time.Now().In(tokyo).Format("2006/01/02")
	got := readFile(t, "blog/"+day+"/trip.md")
	if !strings.Contains(got, "+0900") {
		t.Errorf("the pubDate does not carry the process zone:\n%s", got)
	}
	// and a given day is noon in that zone
	os.WriteFile("dated.md", []byte("# Dated\n\nbody\n"), 0o644)
	if err := runIn(t, "blogit", "dated.md", "2026-11-30"); err != nil {
		t.Fatalf("blogit with a date: %s", err)
	}
	if got := readFile(t, "blog/2026/11/30/dated.md"); !strings.Contains(got, "Mon, 30 Nov 2026 12:00:00 +0900") {
		t.Errorf("noon in the process zone expected:\n%s", got)
	}
}
