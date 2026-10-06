package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// binPath is the antenna binary built once for the whole package: these tests
// check what a script sees, the process exit status, which only a real process
// has (workspace DR-0003 item 6).
var binPath string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "antenna-exit-test")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	binPath = filepath.Join(dir, "antenna")
	if out, err := exec.Command("go", "build", "-o", binPath, ".").CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %s\n%s", err, out)
		os.RemoveAll(dir)
		os.Exit(1)
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// antenna runs the binary in dir with the given environment additions and
// returns its exit status and stderr.
func antenna(t *testing.T, dir string, env []string, args ...string) (int, string) {
	t.Helper()
	cmd := exec.Command(binPath, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &bytes.Buffer{}
	err := cmd.Run()
	if err == nil {
		return 0, stderr.String()
	}
	if ee, ok := err.(*exec.ExitError); ok {
		return ee.ExitCode(), stderr.String()
	}
	t.Fatalf("running antenna %v: %s", args, err)
	return -1, ""
}

func workspace(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if code, errs := antenna(t, dir, nil, "init"); code != 0 {
		t.Fatalf("init exited %d: %s", code, errs)
	}
	return dir
}

func expect(t *testing.T, label string, got, want int, stderr string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: exit %d, want %d; stderr: %s", label, got, want, strings.TrimSpace(stderr))
	}
}

func TestExit_OK(t *testing.T) {
	dir := t.TempDir()
	for _, args := range [][]string{{"-version"}, {"-license"}, {"-help"}, {"help", "topics"}, {"completion", "bash"}} {
		code, errs := antenna(t, dir, nil, args...)
		expect(t, strings.Join(args, " "), code, 0, errs)
	}
	ws := workspace(t)
	code, errs := antenna(t, ws, nil, "posts", "pages.md")
	expect(t, "posts of an empty collection", code, 0, errs)
}

func TestExit_Usage(t *testing.T) {
	dir := t.TempDir()
	cases := [][]string{
		{"-nosuchflag"},
		{"-config"},
		{"nosuchaction"},
		{"help", "nosuchtopic"},
		{"-help", "nosuchtopic"},
		{"completion"},
		{"completion", "tcsh"},
		{"add"},
		{"list", "surplus"},
		{"preview", "surplus"},
	}
	for _, args := range cases {
		code, errs := antenna(t, dir, nil, args...)
		expect(t, strings.Join(args, " "), code, 2, errs)
	}
}

func TestExit_Negative(t *testing.T) {
	ws := workspace(t)
	code, errs := antenna(t, ws, nil, "items", "nosuch.md")
	expect(t, "items of an unknown collection", code, 1, errs)
}

func TestExit_Data(t *testing.T) {
	ws := workspace(t)
	os.WriteFile(filepath.Join(ws, "broken.md"), []byte("---\ntitle: never closed\n\nbody\n"), 0o644)
	code, errs := antenna(t, ws, nil, "post", "broken.md")
	expect(t, "post of a document with unclosed front matter", code, 65, errs)
}

func TestExit_NoInput(t *testing.T) {
	code, errs := antenna(t, t.TempDir(), nil, "list")
	expect(t, "list with no antenna.yaml", code, 66, errs)
}

func TestExit_Unavailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "broken", http.StatusInternalServerError)
	}))
	defer srv.Close()
	ws := workspace(t)
	os.WriteFile(filepath.Join(ws, "feeds.md"), []byte("- [Bad]("+srv.URL+"/bad)\n"), 0o644)
	if code, errs := antenna(t, ws, nil, "add", "feeds.md"); code != 0 {
		t.Fatalf("add exited %d: %s", code, errs)
	}
	code, errs := antenna(t, ws, nil, "harvest", "feeds.md")
	expect(t, "harvest of a feed that returns 500", code, 69, errs)
	if !strings.Contains(errs, "1 of 1") {
		t.Errorf("expected the failure count on stderr, got: %s", errs)
	}
}

func TestExit_CantCreate(t *testing.T) {
	// A completion file antenna did not write is not overwritten.
	xdg := t.TempDir()
	dest := filepath.Join(xdg, "bash-completion", "completions")
	os.MkdirAll(dest, 0o755)
	os.WriteFile(filepath.Join(dest, "antenna"), []byte("somebody else's file\n"), 0o644)
	code, errs := antenna(t, t.TempDir(), []string{"XDG_DATA_HOME=" + xdg}, "completion", "bash", "-install")
	expect(t, "completion -install over a foreign file", code, 73, errs)
}

func TestExit_Config(t *testing.T) {
	ws := workspace(t)
	os.WriteFile(filepath.Join(ws, "antenna.yaml"), []byte("port: [unclosed\n"), 0o644)
	code, errs := antenna(t, ws, nil, "list")
	expect(t, "list with a broken antenna.yaml", code, 78, errs)
}

// A command that fails with an error nobody classified exits 70, never 1.
func TestExit_UnclassifiedIsInternal(t *testing.T) {
	if got := exitCodeOf(fmt.Errorf("something nobody classified")); got != 70 {
		t.Errorf("unclassified error: %d, want 70", got)
	}
}
