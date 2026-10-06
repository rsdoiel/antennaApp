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
	"bytes"
	"strings"
	"testing"
)

func TestItemsFromDB_Basic(t *testing.T) {
	db := newTestItemsDB(t)
	defer db.Close()
	db.Exec(`INSERT INTO items (link, title, pubDate, status, label)
		VALUES ('https://example.com/a', 'Article A', '2020-04-11', 'published', 'My Feed')`)
	db.Exec(`INSERT INTO items (link, title, pubDate, status, label)
		VALUES ('https://example.com/b', 'Article B', '2020-03-01', 'review', 'Other Feed')`)

	var buf bytes.Buffer
	if err := itemsFromDB(&buf, db); err != nil {
		t.Fatalf("itemsFromDB: %s", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Article A") {
		t.Errorf("expected Article A in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Article B") {
		t.Errorf("expected Article B in output, got:\n%s", out)
	}
	if !strings.Contains(out, "2020-04-11") {
		t.Errorf("expected pubDate 2020-04-11, got:\n%s", out)
	}
}

func TestItemsFromDB_Label(t *testing.T) {
	db := newTestItemsDB(t)
	defer db.Close()
	db.Exec(`INSERT INTO items (link, title, pubDate, status, label)
		VALUES ('https://example.com/a', 'Title', '2020-04-11', 'published', 'My Feed')`)

	var buf bytes.Buffer
	if err := itemsFromDB(&buf, db); err != nil {
		t.Fatalf("itemsFromDB: %s", err)
	}
	out := buf.String()
	if !strings.Contains(out, "My Feed") {
		t.Errorf("expected label 'My Feed' in output, got:\n%s", out)
	}
}

func TestItemsFromDB_EmptyTitle(t *testing.T) {
	db := newTestItemsDB(t)
	defer db.Close()
	db.Exec(`INSERT INTO items (link, title, pubDate, status)
		VALUES ('https://example.com/no-title', '', '2020-04-11', 'published')`)

	var buf bytes.Buffer
	if err := itemsFromDB(&buf, db); err != nil {
		t.Fatalf("itemsFromDB: %s", err)
	}
	out := buf.String()
	// When title is empty, link should be used as display text
	if !strings.Contains(out, "https://example.com/no-title") {
		t.Errorf("expected link as display text when title is empty, got:\n%s", out)
	}
}

func TestItemsFromDB_EmptyIsNotAnError(t *testing.T) {
	db := newTestItemsDB(t)
	defer db.Close()

	var buf bytes.Buffer
	// A listing that matches nothing is not a failure (workspace DR-0003,
	// item 3): it prints nothing and succeeds, as ls of an empty directory does.
	if err := itemsFromDB(&buf, db); err != nil {
		t.Errorf("expected no error when DB has no items, got %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output for an empty listing, got %q", buf.String())
	}
}

func TestItemsFromDB_LongPubDate(t *testing.T) {
	db := newTestItemsDB(t)
	defer db.Close()
	db.Exec(`INSERT INTO items (link, title, pubDate, status)
		VALUES ('https://example.com/a', 'Title', '2020-04-11T15:30:00Z', 'published')`)

	var buf bytes.Buffer
	if err := itemsFromDB(&buf, db); err != nil {
		t.Fatalf("itemsFromDB: %s", err)
	}
	out := buf.String()
	if !strings.Contains(out, "2020-04-11") {
		t.Errorf("expected date truncated to YYYY-MM-DD, got:\n%s", out)
	}
	if strings.Contains(out, "T15:30:00") {
		t.Errorf("time component should be stripped from pubDate, got:\n%s", out)
	}
}
