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
	"strings"
	"testing"
	"time"
)

// newTestPagesDB returns an in-memory SQLite database with a pages table
// seeded with the provided rows (inputPath, outputPath).
func newTestPagesDB(t *testing.T, rows [][2]string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory db: %s", err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS pages (
		inputPath PRIMARY KEY,
		outputPath TEXT DEFAULT '',
		updated    DATETIME
	)`)
	if err != nil {
		t.Fatalf("create pages table: %s", err)
	}
	ts := time.Now().Format(time.RFC3339)
	for _, row := range rows {
		if _, err := db.Exec(`INSERT INTO pages (inputPath, outputPath, updated) VALUES (?, ?, ?)`,
			row[0], row[1], ts); err != nil {
			t.Fatalf("insert row %v: %s", row, err)
		}
	}
	return db
}

func TestWritePageIndex_ProducesUL(t *testing.T) {
	db := newTestPagesDB(t, [][2]string{
		{"about.md", "about.html"},
		{"contact.md", "contact.html"},
	})
	defer db.Close()

	gen := &Generator{}
	var buf bytes.Buffer
	if err := gen.WritePageIndex(&buf, db); err != nil {
		t.Fatalf("WritePageIndex: %s", err)
	}
	out := buf.String()
	if !strings.Contains(out, "<ul>") {
		t.Errorf("expected <ul> in output:\n%s", out)
	}
	if !strings.Contains(out, "</ul>") {
		t.Errorf("expected </ul> in output:\n%s", out)
	}
}

func TestWritePageIndex_LinksToOutputPaths(t *testing.T) {
	db := newTestPagesDB(t, [][2]string{
		{"about.md", "about.html"},
		{"contact.md", "contact.html"},
	})
	defer db.Close()

	gen := &Generator{}
	var buf bytes.Buffer
	if err := gen.WritePageIndex(&buf, db); err != nil {
		t.Fatalf("WritePageIndex: %s", err)
	}
	out := buf.String()
	if !strings.Contains(out, "about.html") {
		t.Errorf("expected link to about.html in output:\n%s", out)
	}
	if !strings.Contains(out, "contact.html") {
		t.Errorf("expected link to contact.html in output:\n%s", out)
	}
}

func TestWritePageIndex_UsesInputPathAsDisplayName(t *testing.T) {
	db := newTestPagesDB(t, [][2]string{
		{"about.md", "about.html"},
	})
	defer db.Close()

	gen := &Generator{}
	var buf bytes.Buffer
	if err := gen.WritePageIndex(&buf, db); err != nil {
		t.Fatalf("WritePageIndex: %s", err)
	}
	out := buf.String()
	// Display name should be derived from "about.md" — at minimum "about" should appear
	if !strings.Contains(strings.ToLower(out), "about") {
		t.Errorf("expected derived display name 'about' in output:\n%s", out)
	}
}

func TestWritePageIndex_EmptyDBProducesEmptyList(t *testing.T) {
	db := newTestPagesDB(t, nil)
	defer db.Close()

	gen := &Generator{}
	var buf bytes.Buffer
	if err := gen.WritePageIndex(&buf, db); err != nil {
		t.Fatalf("WritePageIndex on empty DB: %s", err)
	}
	out := buf.String()
	// Should still produce valid (empty) list markup without error
	if !strings.Contains(out, "<ul>") {
		t.Errorf("expected <ul> even for empty DB:\n%s", out)
	}
}

func TestWritePageIndex_WrappedInMain(t *testing.T) {
	db := newTestPagesDB(t, [][2]string{
		{"about.md", "about.html"},
	})
	defer db.Close()

	gen := &Generator{}
	var buf bytes.Buffer
	if err := gen.WritePageIndex(&buf, db); err != nil {
		t.Fatalf("WritePageIndex: %s", err)
	}
	out := buf.String()
	if !strings.Contains(out, `id="main-content"`) {
		t.Errorf("expected main-content landmark in output:\n%s", out)
	}
}

func TestCollectionMode_FieldExists(t *testing.T) {
	col := &Collection{Mode: "page-index"}
	if col.Mode != "page-index" {
		t.Errorf("expected Mode field to be 'page-index', got %q", col.Mode)
	}
}
