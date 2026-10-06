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
	"archive/zip"
	"bytes"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// The table of workspace DR-0003. A change here is a breaking change for
// every caller, so the numbers are pinned.
func TestExitClasses_TableIsDR0003(t *testing.T) {
	want := map[string]int{
		"ok": 0, "negative": 1, "usage": 2, "data": 65, "no_input": 66,
		"unavailable": 69, "internal": 70, "cant_create": 73, "io": 74,
		"temp_fail": 75, "no_permission": 77, "config": 78,
	}
	if len(allExitClasses) != len(want) {
		t.Errorf("expected %d classes, got %d", len(want), len(allExitClasses))
	}
	for _, c := range allExitClasses {
		code, ok := want[c.Name]
		if !ok {
			t.Errorf("class %q is not in DR-0003", c.Name)
			continue
		}
		if c.Code != code {
			t.Errorf("class %q: code %d, want %d", c.Name, c.Code, code)
		}
		// The shell owns 126, 127 and everything above 128.
		if c.Code > 125 {
			t.Errorf("class %q: code %d is in the shell's range", c.Name, c.Code)
		}
	}
}

func TestClassify_Constructors(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want exitClass
	}{
		{"usage", usageErrorf("bad %s", "x"), classUsage},
		{"negative", negativef("nope"), classNegative},
		{"data", dataErrorf("bad content"), classData},
		{"no_input", noInputf("missing"), classNoInput},
		{"unavailable", unavailablef("down"), classUnavailable},
		{"internal", internalErrorf("bug"), classInternal},
		{"cant_create", cantCreatef("exists"), classCantCreate},
		{"io", ioErrorf("half written"), classIO},
		{"config", configErrorf("bad yaml"), classConfig},
	}
	for _, c := range cases {
		got, ok := classify(c.err)
		if !ok || got != c.want {
			t.Errorf("%s: got %v (classified=%v), want %v", c.name, got, ok, c.want)
		}
	}
}

func TestClassify_NilIsOK(t *testing.T) {
	if got, ok := classify(nil); !ok || got != classOK {
		t.Errorf("nil: got %v, %v", got, ok)
	}
	if ExitCodeFor(nil).Code != 0 {
		t.Errorf("ExitCodeFor(nil) = %d, want 0", ExitCodeFor(nil).Code)
	}
}

func TestClassify_OutermostClassWins(t *testing.T) {
	inner := usageErrorf("inner")
	outer := fmt.Errorf("while reading: %w", dataErrorf("wrapped %w", inner))
	if got, _ := classify(outer); got != classData {
		t.Errorf("got %v, want data (the nearest class around the cause)", got)
	}
}

func TestClassify_Stdlib(t *testing.T) {
	dir := t.TempDir()
	_, notExist := os.Open(filepath.Join(dir, "missing.yaml"))
	exists := os.Mkdir(dir, 0o755)
	_, notDir := os.ReadDir(filepath.Join(dir, "missing"))
	_, zipErr := zip.NewReader(bytes.NewReader([]byte("not a zip archive at all")), 24)
	xmlErr := xml.Unmarshal([]byte("<a><b></a>"), &struct{}{})
	var jsonErr error = json.Unmarshal([]byte("{nope"), &struct{}{})
	// yaml.v3 reports a syntax error as a plain fmt error, so only a type
	// mismatch (*yaml.TypeError) can be recognised by type; the sites that parse
	// YAML classify their own syntax errors.
	var yamlErr error = yaml.Unmarshal([]byte("a: [1]"), &struct{ A int }{})
	cases := []struct {
		name string
		err  error
		want exitClass
	}{
		{"fs.ErrNotExist", notExist, classNoInput},
		{"wrapped ErrNotExist", fmt.Errorf("open config: %w", notExist), classNoInput},
		{"ReadDir of missing", notDir, classNoInput},
		{"fs.ErrExist", exists, classCantCreate},
		{"fs.ErrPermission", fmt.Errorf("x: %w", fs.ErrPermission), classNoPermission},
		{"json syntax", jsonErr, classData},
		{"yaml type error", yamlErr, classData},
		{"unexpected EOF", io.ErrUnexpectedEOF, classData},
		{"zip.ErrFormat", zipErr, classData},
		{"xml syntax", xmlErr, classData},
		{"PathError", &fs.PathError{Op: "write", Path: "/x", Err: errors.New("disk full")}, classIO},
		{"url.Error", &url.Error{Op: "Get", URL: "http://x", Err: errors.New("refused")}, classUnavailable},
		{"net.OpError", &net.OpError{Op: "dial", Err: errors.New("refused")}, classUnavailable},
		{"net.DNSError", &net.DNSError{Err: "no such host", Name: "x"}, classUnavailable},
	}
	for _, c := range cases {
		got, ok := classify(c.err)
		if !ok || got != c.want {
			t.Errorf("%s: got %v (classified=%v), want %v", c.name, got, ok, c.want)
		}
	}
}

func TestClassify_UnclassifiedIsInternal(t *testing.T) {
	err := errors.New("something nobody classified")
	if _, ok := classify(err); ok {
		t.Errorf("a plain error must not report itself classified")
	}
	if got := ExitCodeFor(err).Code; got != 70 {
		t.Errorf("ExitCodeFor(plain error) = %d, want 70 (internal)", got)
	}
}

// The next four use real driver errors. Constructed ones would not show a
// mismatch between the drivers this package links (glebarez "sqlite" in
// production, mattn "sqlite3" in sitemap.go and the tests).
func openTestDB(t *testing.T, driver, path string) *sql.DB {
	t.Helper()
	db, err := sql.Open(driver, path)
	if err != nil {
		t.Fatalf("sql.Open(%s): %s", driver, err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestClassify_SQLiteConstraint_BothDrivers(t *testing.T) {
	for _, driver := range []string{"sqlite", "sqlite3"} {
		db := openTestDB(t, driver, filepath.Join(t.TempDir(), "c.db"))
		if _, err := db.Exec(`CREATE TABLE t (k PRIMARY KEY)`); err != nil {
			t.Fatalf("%s: %s", driver, err)
		}
		if _, err := db.Exec(`INSERT INTO t VALUES (1)`); err != nil {
			t.Fatalf("%s: %s", driver, err)
		}
		_, err := db.Exec(`INSERT INTO t VALUES (1)`)
		if err == nil {
			t.Fatalf("%s: duplicate insert did not fail", driver)
		}
		if got, ok := classify(err); !ok || got != classData {
			t.Errorf("%s constraint violation: got %v (classified=%v), want data; err=%T %v", driver, got, ok, err, err)
		}
	}
}

func TestClassify_SQLiteLocked_BothDrivers(t *testing.T) {
	for _, driver := range []string{"sqlite", "sqlite3"} {
		path := filepath.Join(t.TempDir(), "l.db")
		holder := openTestDB(t, driver, path)
		holder.SetMaxOpenConns(1)
		if _, err := holder.Exec(`CREATE TABLE t (k)`); err != nil {
			t.Fatalf("%s: %s", driver, err)
		}
		tx, err := holder.Begin()
		if err != nil {
			t.Fatalf("%s: %s", driver, err)
		}
		if _, err := tx.Exec(`INSERT INTO t VALUES (1)`); err != nil {
			t.Fatalf("%s: %s", driver, err)
		}
		// No busy timeout, so the blocked writer fails at once instead of
		// waiting out the driver's default.
		dsn := path + "?_pragma=busy_timeout(0)"
		if driver == "sqlite3" {
			dsn = path + "?_busy_timeout=0"
		}
		other := openTestDB(t, driver, dsn)
		_, err = other.Exec(`INSERT INTO t VALUES (2)`)
		tx.Rollback()
		if err == nil {
			t.Fatalf("%s: second writer was not blocked", driver)
		}
		if got, ok := classify(err); !ok || got != classTempFail {
			t.Errorf("%s locked database: got %v (classified=%v), want temp_fail; err=%T %v", driver, got, ok, err, err)
		}
	}
}

func TestClassify_SQLiteNotADatabase_BothDrivers(t *testing.T) {
	for _, driver := range []string{"sqlite", "sqlite3"} {
		path := filepath.Join(t.TempDir(), "junk.db")
		if err := os.WriteFile(path, []byte("this is not a sqlite database, just text padded out to be long enough to be read as a header"), 0o644); err != nil {
			t.Fatal(err)
		}
		db := openTestDB(t, driver, path)
		_, err := db.Exec(`CREATE TABLE t (k)`)
		if err == nil {
			t.Fatalf("%s: a text file opened as a database", driver)
		}
		if got, ok := classify(err); !ok || got != classData {
			t.Errorf("%s not a database: got %v (classified=%v), want data; err=%T %v", driver, got, ok, err, err)
		}
	}
}

func TestAsCreate(t *testing.T) {
	dir := t.TempDir()
	_, open := os.OpenFile(filepath.Join(dir, "no", "such", "dir", "f"), os.O_WRONLY|os.O_CREATE, 0o644)
	cases := []struct {
		name string
		err  error
		want exitClass
	}{
		{"open failed", open, classCantCreate},
		{"mkdir failed", &fs.PathError{Op: "mkdir", Path: "/x", Err: errors.New("not a directory")}, classCantCreate},
		{"target exists", fs.ErrExist, classCantCreate},
		{"write failed part way", &fs.PathError{Op: "write", Path: "/x", Err: errors.New("disk full")}, classIO},
		{"permission refused", fmt.Errorf("x: %w", fs.ErrPermission), classNoPermission},
		{"already classed", dataErrorf("bad"), classData},
	}
	for _, c := range cases {
		got, ok := classify(asCreate(c.err))
		if !ok || got != c.want {
			t.Errorf("%s: got %v (classified=%v), want %v", c.name, got, ok, c.want)
		}
	}
	if asCreate(nil) != nil {
		t.Errorf("asCreate(nil) must be nil")
	}
}

// A port that is already in use is a busy resource, not an unreachable
// service: a retry may succeed (temp_fail, 75). The real error is used.
func TestClassify_AddressInUseIsTempFail(t *testing.T) {
	first, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	_, err = net.Listen("tcp", first.Addr().String())
	if err == nil {
		t.Fatal("second listener on the same address did not fail")
	}
	if got, ok := classify(fmt.Errorf("cannot start preview server: %w", err)); !ok || got != classTempFail {
		t.Errorf("address in use: got %v (classified=%v), want temp_fail; err=%T %v", got, ok, err, err)
	}
}
