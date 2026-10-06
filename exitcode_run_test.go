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
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runIn runs one action through the same dispatcher the command uses, from
// the current directory, and returns the error it produced. Handler tests go
// through Run, not the handler, so a wrong argument shape shows up here the
// way it would on the command line.
func runIn(t *testing.T, action string, args ...string) error {
	t.Helper()
	// Several handlers print straight to stdout; keep the test output clean.
	old := os.Stdout
	if null, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0); err == nil {
		os.Stdout = null
		defer func() { os.Stdout = old; null.Close() }()
	}
	app := &AntennaApp{appName: "antenna"}
	var out, eout bytes.Buffer
	return app.Run(strings.NewReader(""), &out, &eout, "antenna.yaml", action, args)
}

// newWorkspace makes a temporary directory, makes it the working directory,
// and, when init is true, runs antenna init in it so antenna.yaml, page.yaml,
// pages.md and pages.db exist.
func newWorkspace(t *testing.T, init bool) string {
	t.Helper()
	dir := t.TempDir()
	t.Chdir(dir)
	if init {
		if err := runIn(t, "init"); err != nil {
			t.Fatalf("init: %s", err)
		}
	}
	return dir
}

// wantClass fails the test unless err classifies as want, and says which
// error text produced the wrong class.
func wantClass(t *testing.T, label string, err error, want exitClass) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: no error, want class %s (%d)", label, want.Name, want.Code)
		return
	}
	if got := ExitCodeFor(err); got != want {
		t.Errorf("%s: class %s (%d), want %s (%d); error: %v",
			label, got.Name, got.Code, want.Name, want.Code, err)
	}
}

// A command line that is wrong gets usage (2) before anything is read or
// written, so these run with no workspace at all.
func TestRun_UsageErrors(t *testing.T) {
	newWorkspace(t, false)
	cases := []struct {
		label  string
		action string
		args   []string
	}{
		{"unknown action", "nosuchaction", nil},
		{"unknown help topic", "help", []string{"nosuchtopic"}},
		{"completion without shell", "completion", nil},
		{"completion with two shells", "completion", []string{"bash", "powershell"}},
		{"completion with unsupported shell", "completion", []string{"tcsh"}},
		{"add without collection", "add", nil},
		{"del without collection", "del", nil},
		{"post without file", "post", nil},
		{"page without file", "page", nil},
		{"rss without collection", "rss", nil},
		{"apply without theme", "apply", nil},
		{"stylefrom without input", "stylefrom", nil},
		{"quote without URL", "quote", nil},
	}
	for _, c := range cases {
		wantClass(t, c.label, runIn(t, c.action, c.args...), classUsage)
	}
}

// A missing antenna.yaml is a missing workspace, not a bug and not a usage
// mistake.
func TestRun_MissingWorkspaceIsNoInput(t *testing.T) {
	newWorkspace(t, false)
	for _, action := range []string{"list", "items", "generate", "harvest", "sitemap"} {
		wantClass(t, action+" with no antenna.yaml", runIn(t, action), classNoInput)
	}
}

func TestRun_MissingInputFileIsNoInput(t *testing.T) {
	dir := newWorkspace(t, true)
	missing := filepath.Join(dir, "missing.md")
	missingHTML := filepath.Join(dir, "missing.html")
	wantClass(t, "page of missing file", runIn(t, "page", missing), classNoInput)
	wantClass(t, "stylefrom of missing file", runIn(t, "stylefrom", missingHTML), classNoInput)
	wantClass(t, "add of missing file", runIn(t, "add", missing), classNoInput)
}

func TestRun_BrokenConfigIsConfig(t *testing.T) {
	newWorkspace(t, true)
	if err := os.WriteFile("antenna.yaml", []byte("port: [unclosed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	wantClass(t, "list with a broken antenna.yaml", runIn(t, "list"), classConfig)
}

// DR-0003 item 3: a listing that matches nothing is still 0.
func TestRun_EmptyListingsAreOK(t *testing.T) {
	newWorkspace(t, true)
	for _, c := range [][]string{{"posts", "pages.md"}, {"items", "pages.md"}, {"pages"}} {
		if err := runIn(t, c[0], c[1:]...); err != nil {
			t.Errorf("%s of an empty collection: want no error, got %v", c[0], err)
		}
	}
}

// A collection named on the command line that is not in antenna.yaml is "no
// such item": the command ran correctly and the answer is no (1).
func TestRun_UnknownCollectionIsNegative(t *testing.T) {
	newWorkspace(t, true)
	for _, c := range [][]string{{"posts", "nosuch.md"}, {"items", "nosuch.md"}, {"del", "nosuch.md"}} {
		wantClass(t, strings.Join(c, " "), runIn(t, c[0], c[1:]...), classNegative)
	}
}

func TestRun_MalformedDocumentIsData(t *testing.T) {
	newWorkspace(t, true)
	if err := os.WriteFile("broken.md", []byte("---\ntitle: never closed\n\nbody\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	wantClass(t, "post of unclosed front matter", runIn(t, "post", "broken.md"), classData)
	if err := os.WriteFile("fake.odt", []byte("this is not a zip archive"), 0o644); err != nil {
		t.Fatal(err)
	}
	wantClass(t, "post of a text file named .odt", runIn(t, "post", "fake.odt"), classData)
}

func TestRun_BadGeneratorSettingIsConfig(t *testing.T) {
	newWorkspace(t, true)
	src, err := os.ReadFile("page.yaml")
	if err != nil {
		t.Fatal(err)
	}
	src = append(src, []byte("\nitems:\n  html: bogus\n")...)
	if err := os.WriteFile("page.yaml", src, 0o644); err != nil {
		t.Fatal(err)
	}
	wantClass(t, "generate with items.html: bogus", runIn(t, "generate"), classConfig)
}

func TestRun_OutputDirectoryThatCannotBeMadeIsCantCreate(t *testing.T) {
	newWorkspace(t, true)
	// A file where the css directory must go.
	if err := os.WriteFile("css", []byte("in the way"), 0o644); err != nil {
		t.Fatal(err)
	}
	wantClass(t, "css where css is a file", runIn(t, "css"), classCantCreate)
}

// feedServer serves one good RSS feed at /good and a server error at /bad.
func feedServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/good" {
			w.Header().Set("Content-Type", "application/rss+xml")
			w.Write([]byte(`<?xml version="1.0"?><rss version="2.0"><channel><title>Good</title>` +
				`<link>http://example.com/</link><description>d</description>` +
				`<item><title>One</title><link>http://example.com/one</link>` +
				`<description>first</description><pubDate>Mon, 05 Oct 2026 12:00:00 +0000</pubDate></item>` +
				`</channel></rss>`))
			return
		}
		http.Error(w, "broken", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// DR-0003 item 4: a bulk command with failures processes everything first,
// then exits with the class of the first failure.
func TestRun_HarvestWithOneDeadFeedProcessesTheRestThenFails(t *testing.T) {
	newWorkspace(t, true)
	srv := feedServer(t)
	feeds := "# Feeds\n\n- [Bad](" + srv.URL + "/bad)\n- [Good](" + srv.URL + "/good)\n"
	if err := os.WriteFile("feeds.md", []byte(feeds), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runIn(t, "add", "feeds.md"); err != nil {
		t.Fatalf("add: %s", err)
	}
	err := runIn(t, "harvest", "feeds.md")
	wantClass(t, "harvest with a feed that returns 500", err, classUnavailable)
	if err != nil && !strings.Contains(err.Error(), "1 of 2") {
		t.Errorf("expected the failure count in the message, got: %v", err)
	}
	// The good feed after the bad one was still harvested.
	out := &bytes.Buffer{}
	app := &AntennaApp{appName: "antenna"}
	if err := app.Run(strings.NewReader(""), out, out, "antenna.yaml", "items", []string{"feeds.md"}); err != nil {
		t.Fatalf("items: %s", err)
	}
	if !strings.Contains(out.String(), "One") {
		t.Errorf("the feed after the failing one was not harvested; items:\n%s", out.String())
	}
}

// A surplus argument is a usage error (2) and is refused before anything
// runs, however many the verb takes: a verb that took none used to drop them
// and exit 0.
func TestRun_SurplusArgumentsAreUsage(t *testing.T) {
	newWorkspace(t, true)
	cases := []struct {
		verb string
		args []string
	}{
		{"init", []string{"extra"}},
		{"list", []string{"extra"}},
		{"pages", []string{"extra"}},
		{"sitemap", []string{"extra"}},
		{"preview", []string{"extra"}},
		{"css", []string{"a.css", "extra"}},
		{"items", []string{"pages.md", "extra"}},
		{"unpage", []string{"a.md", "extra"}},
		{"quote", []string{"https://example.com/#:~:text=a", "extra"}},
		{"apply", []string{"theme", "page.yaml", "extra"}},
		{"page", []string{"a.md", "a.html", "extra"}},
		{"stylefrom", []string{"a.html", "out", "extra"}},
		{"add", []string{"a.md", "name", "description", "extra"}},
		{"posts", []string{"pages.md", "a", "b", "extra"}},
		{"rss", []string{"pages.md", "a.xml", "1", "2", "extra"}},
		{"unpost", []string{"pages.md", "a", "extra"}},
		{"post", []string{"pages.md", "a.md", "2026-01-01", "extra"}},
	}
	for _, c := range cases {
		wantClass(t, c.verb+" with a surplus argument", runIn(t, c.verb, c.args...), classUsage)
	}
}
