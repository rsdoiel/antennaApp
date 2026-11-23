package antennaApp

import (
	"bytes"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runCapture runs one verb through the dispatcher in the current directory
// and returns its standard output, standard error and error.
func runCapture(t *testing.T, action string, args ...string) (string, string, error) {
	t.Helper()
	app := &AntennaApp{appName: "antenna"}
	var out, eout bytes.Buffer
	err := app.Run(strings.NewReader(""), &out, &eout, "antenna.yaml", action, args)
	return out.String(), eout.String(), err
}

// pruneSite makes a workspace with the pages named, each paged so the pages
// table records its output, name.html.
func pruneSite(t *testing.T, names ...string) {
	t.Helper()
	newWorkspace(t, true)
	for _, n := range names {
		src := "---\ntitle: " + n + "\n---\n\nBody of " + n + "\n"
		if err := os.WriteFile(n+".md", []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := runIn(t, "page", n+".md"); err != nil {
			t.Fatalf("page %s: %s", n, err)
		}
	}
}

func pageRowCount(t *testing.T) int {
	t.Helper()
	db, err := sql.Open("sqlite", "pages.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM pages`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func insertPageRow(t *testing.T, input, outputsJSON string) {
	t.Helper()
	db, err := sql.Open("sqlite", "pages.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`INSERT INTO pages (inputPath, outputPath, updated, outputs, includes)
		VALUES (?, '', '2026-11-20T10:00:00-08:00', ?, '')`, input, outputsJSON); err != nil {
		t.Fatal(err)
	}
}

// DR-0028 1: with no option prune is a dry run. It lists the row and the
// output it would delete, exits 1 (drift found), and changes nothing.
func TestPruneDryRunChangesNothing(t *testing.T) {
	pruneSite(t, "keep", "gone")
	os.Remove("gone.md")
	out, _, err := runCapture(t, "prune")
	wantClass(t, "dry run with a stale row", err, classNegative)
	if !strings.Contains(out, "gone.md") || !strings.Contains(out, "gone.html") {
		t.Errorf("dry run does not name the row and its output:\n%s", out)
	}
	if !exists("gone.html") || pageRowCount(t) != 2 {
		t.Errorf("dry run changed something: gone.html exists=%v rows=%d", exists("gone.html"), pageRowCount(t))
	}
}

// DR-0028 1, 2: -w deletes the row and its recorded outputs, leaves the page
// whose source is there, and a second run has nothing to do (0).
func TestPruneWriteRemovesRowAndOutputs(t *testing.T) {
	pruneSite(t, "keep", "gone")
	os.Remove("gone.md")
	if _, _, err := runCapture(t, "prune", "-w"); err != nil {
		t.Fatalf("prune -w: %v", err)
	}
	if exists("gone.html") {
		t.Error("gone.html was not deleted")
	}
	if !exists("keep.html") || !exists("keep.md") {
		t.Error("the page whose source exists was touched")
	}
	if n := pageRowCount(t); n != 1 {
		t.Errorf("rows = %d, want 1", n)
	}
	if _, _, err := runCapture(t, "prune"); err != nil {
		t.Errorf("second prune: want nothing to do (0), got %v", err)
	}
}

// DR-0028 2: an output a surviving row also records is kept and reported.
func TestPruneKeepsSharedOutput(t *testing.T) {
	pruneSite(t, "keep")
	insertPageRow(t, "old.md", `["keep.html"]`)
	_, eout, err := runCapture(t, "prune", "-w")
	if err != nil {
		t.Fatalf("prune -w: %v", err)
	}
	if !exists("keep.html") {
		t.Error("an output shared with a surviving row was deleted")
	}
	if !strings.Contains(eout, "keep.html") {
		t.Errorf("the kept output is not reported on standard error: %q", eout)
	}
	if n := pageRowCount(t); n != 1 {
		t.Errorf("rows = %d, want the stale row gone (1)", n)
	}
}

// DR-0028 2: an output that leaves htdocs is never deleted.
func TestPruneRefusesOutputOutsideHtdocs(t *testing.T) {
	dir := newWorkspace(t, true)
	outside := filepath.Join(filepath.Dir(dir), "outside-"+filepath.Base(dir)+".txt")
	if err := os.WriteFile(outside, []byte("precious"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(outside) })
	if err := os.WriteFile("live.md", []byte("# Live\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runIn(t, "page", "live.md"); err != nil {
		t.Fatal(err)
	}
	insertPageRow(t, "old.md", `["../`+filepath.Base(outside)+`", "`+outside+`"]`)
	_, eout, _ := runCapture(t, "prune", "-w")
	if !exists(outside) {
		t.Error("a file outside htdocs was deleted")
	}
	if !strings.Contains(eout, filepath.Base(outside)) {
		t.Errorf("the refused output is not reported: %q", eout)
	}
}

// DR-0028 3: when every row's source is missing, prune refuses unless -all.
func TestPruneRefusesWhenEverySourceIsMissing(t *testing.T) {
	pruneSite(t, "a", "b")
	os.Remove("a.md")
	os.Remove("b.md")
	_, _, err := runCapture(t, "prune", "-w")
	wantClass(t, "every source missing", err, classNegative)
	if !exists("a.html") || !exists("b.html") || pageRowCount(t) != 2 {
		t.Error("prune -w deleted something it refused to")
	}
	if _, _, err := runCapture(t, "prune", "-w", "-all"); err != nil {
		t.Fatalf("prune -w -all: %v", err)
	}
	if exists("a.html") || exists("b.html") || pageRowCount(t) != 0 {
		t.Error("-all did not prune everything")
	}
}

// DR-0028 5: exit classes through the dispatcher.
func TestPruneExitClasses(t *testing.T) {
	pruneSite(t, "keep")
	if _, _, err := runCapture(t, "prune"); err != nil {
		t.Errorf("nothing to prune: want 0, got %v", err)
	}
	_, _, err := runCapture(t, "prune", "-nosuch")
	wantClass(t, "unknown option", err, classUsage)
	_, _, err = runCapture(t, "prune", "extra")
	wantClass(t, "extra argument", err, classUsage)
	newWorkspace(t, false)
	_, _, err = runCapture(t, "prune")
	wantClass(t, "no workspace", err, classNoInput)
}
