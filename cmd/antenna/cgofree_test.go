package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The release binaries are cross-compiled, and a cross-compile builds with
// cgo off. Code that names a cgo-only symbol (the mattn sqlite3 driver's Error
// type) then fails to build at all, and code that opens the "sqlite3" driver
// builds but fails at run time with "requires cgo". Native tests run with cgo
// on and see neither, so this builds and runs the release configuration.

// cgoFreeBinary builds antenna with CGO_ENABLED=0 for the host, once per test
// binary, and returns its path.
func cgoFreeBinary(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "antenna-nocgo")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("building with CGO_ENABLED=0 failed (a cross-compile will too):\n%s", out)
	}
	return bin
}

func TestCgoFree_BuildsLikeACrossCompile(t *testing.T) {
	cgoFreeBinary(t)
}

// sitemap is the one command that read databases through the cgo-only
// driver. In the release configuration it must still work.
func TestCgoFree_SitemapWorks(t *testing.T) {
	bin := cgoFreeBinary(t)
	ws := t.TempDir()
	run := func(args ...string) (int, string) {
		cmd := exec.Command(bin, args...)
		cmd.Dir = ws
		out, err := cmd.CombinedOutput()
		if err == nil {
			return 0, string(out)
		}
		if ee, ok := err.(*exec.ExitError); ok {
			return ee.ExitCode(), string(out)
		}
		t.Fatalf("%v: %s", args, err)
		return -1, ""
	}
	if code, out := run("init"); code != 0 {
		t.Fatalf("init exited %d: %s", code, out)
	}
	if err := os.WriteFile(filepath.Join(ws, "about.md"), []byte("---\ntitle: About\n---\n\nAbout.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, out := run("page", "about.md"); code != 0 {
		t.Fatalf("page exited %d: %s", code, out)
	}
	code, out := run("sitemap")
	if code != 0 {
		t.Fatalf("sitemap exited %d in a cgo-free build: %s", code, out)
	}
	src, err := os.ReadFile(filepath.Join(ws, "sitemap_1.xml"))
	if err != nil {
		t.Fatalf("no sitemap written: %s\n%s", err, out)
	}
	if !strings.Contains(string(src), "about.html") {
		t.Errorf("the page is missing from the sitemap:\n%s", src)
	}
}
