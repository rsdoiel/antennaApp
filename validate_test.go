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
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// runValidate runs antenna validate through the real dispatcher in the current
// directory and returns what it wrote to standard output and the error.
func runValidate(t *testing.T, args ...string) (string, error) {
	t.Helper()
	app := &AntennaApp{appName: "antenna"}
	var out, eout bytes.Buffer
	err := app.Run(strings.NewReader(""), &out, &eout, "antenna.yaml", "validate", args)
	return out.String(), err
}

// finding reports whether a plain output line for file with the code and
// severity is in out, and returns the line.
func finding(out, file, severity, code string) (string, bool) {
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, file) && strings.Contains(line, ": "+severity+": "+code+": ") {
			return line, true
		}
	}
	return "", false
}

func write(t *testing.T, name, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(name), 0o775); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(name, []byte(content), 0o664); err != nil {
		t.Fatal(err)
	}
}

const goodPost = "---\ntitle: Good\npostPath: blog/2026/11/19/good.md\npubDate: \"Thu, 19 Nov 2026 12:00:00 -0800\"\n---\n\n# Good\n\nBody.\n"

// DR-0017 3: a clean file exits 0 and says so.
func TestValidateACleanFileIsClean(t *testing.T) {
	newWorkspace(t, true)
	write(t, "blog/2026/11/19/good.md", goodPost)
	out, err := runValidate(t, "blog/2026/11/19/good.md")
	if err != nil {
		t.Fatalf("validate: %s\n%s", err, out)
	}
	if !strings.Contains(out, "0 errors, 0 warnings, 0 info") {
		t.Errorf("no summary:\n%s", out)
	}
}

// DR-0017 4: one case per front matter check, through the real verb, with the
// file and the line of the key the finding is about.
func TestValidateFrontMatterChecks(t *testing.T) {
	cases := []struct {
		label, content, severity, code string
		line                           int // 0: no line asserted
		exit                           *exitClass
	}{
		{"unclosed front matter", "---\ntitle: never closed\n\nbody\n", "error", "front-matter", 0, &classNegative},
		{"post with no date", "---\ntitle: Draft\npostPath: p.md\n---\nbody\n", "warning", "post-no-pubdate", 0, &classNegative},
		{"unreadable date", "---\ntitle: T\npubDate: \"next tuesday\"\n---\nbody\n", "error", "date-unreadable", 3, &classNegative},
		{"date not in the RSS form is info", "---\ntitle: T\npubDate: \"2026-11-19\"\n---\nbody\n", "info", "date-form", 3, nil},
		{"no title and no heading", "---\ndescription: d\n---\nbody\n", "warning", "no-title", 0, &classNegative},
		{"post path elsewhere", "---\ntitle: T\npostPath: blog/2026/01/01/other.md\npubDate: \"Thu, 19 Nov 2026 12:00:00 -0800\"\n---\nbody\n", "warning", "postpath-mismatch", 3, &classNegative},
		{"urlNext that is a script", "---\ntitle: T\nurlNext: \"javascript:alert(1)\"\n---\nbody\n", "error", "nav-address", 3, &classNegative},
		{"rightMarginImage that is a script", "---\ntitle: T\nrightMarginImage: \"data:text/html,x\"\n---\nbody\n", "error", "nav-address", 3, &classNegative},
	}
	for _, c := range cases {
		newWorkspace(t, true)
		write(t, "p.md", c.content)
		out, err := runValidate(t, "p.md")
		line, ok := finding(out, "p.md", c.severity, c.code)
		if !ok {
			t.Errorf("%s: no %s %s finding in:\n%s", c.label, c.severity, c.code, out)
			continue
		}
		if c.line > 0 && !strings.HasPrefix(line, "p.md:"+itoa(c.line)+": ") {
			t.Errorf("%s: want line %d, got %q", c.label, c.line, line)
		}
		if c.exit == nil {
			if err != nil {
				t.Errorf("%s: an info finding must exit 0, got %v", c.label, err)
			}
		} else {
			wantClass(t, c.label, err, *c.exit)
		}
	}
}

func itoa(n int) string { return strconv.Itoa(n) }

// A heading stands in for a title, and a good file has no finding at all.
func TestValidateAHeadingIsATitle(t *testing.T) {
	newWorkspace(t, true)
	write(t, "p.md", "# Heading only\n\nbody\n")
	out, err := runValidate(t, "p.md")
	if _, bad := finding(out, "p.md", "warning", "no-title"); bad || err != nil {
		t.Errorf("a heading should count as a title: %v\n%s", err, out)
	}
}

// DR-0017 1: a directory is walked; hidden folders and generated files are skipped.
func TestValidateWalksADirectory(t *testing.T) {
	newWorkspace(t, true)
	write(t, "notes/a.md", "---\ndescription: d\n---\nbody\n")     // no title
	write(t, "notes/sub/b.md", "---\ndescription: d\n---\nbody\n") // no title
	write(t, "notes/.hidden/c.md", "---\ndescription: d\n---\nbody\n")
	write(t, "notes/news_items.md", "---\naggregateOf: news.md\n---\n# generated\n")
	out, _ := runValidate(t, "notes")
	for _, want := range []string{"notes/a.md", "notes/sub/b.md"} {
		if _, ok := finding(out, want, "warning", "no-title"); !ok {
			t.Errorf("%s was not checked:\n%s", want, out)
		}
	}
	for _, skipped := range []string{"notes/.hidden/c.md", "notes/news_items.md"} {
		if strings.Contains(out, skipped) {
			t.Errorf("%s should be skipped:\n%s", skipped, out)
		}
	}
}

// DR-0017 3: exit codes for the command line.
func TestValidateExitCodes(t *testing.T) {
	newWorkspace(t, true)
	_, err := runValidate(t, "nosuch.md")
	wantClass(t, "a missing PATH", err, classNoInput)
	_, err = runValidate(t, "-nosuchflag")
	wantClass(t, "an unknown flag", err, classUsage)
	_, err = runValidate(t, "a.md", "b.md")
	wantClass(t, "two paths", err, classUsage)
}

// DR-0017 2: -json writes one object per finding and a summary, and is valid
// JSON line by line.
func TestValidateJSON(t *testing.T) {
	newWorkspace(t, true)
	write(t, "p.md", "---\ndescription: d\n---\nbody\n")
	out, err := runValidate(t, "-json", "p.md")
	wantClass(t, "findings in json mode", err, classNegative)
	var sawFinding, sawSummary bool
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		var obj map[string]any
		if e := json.Unmarshal([]byte(line), &obj); e != nil {
			t.Fatalf("not JSON: %q (%s)", line, e)
		}
		if obj["code"] == "no-title" && obj["severity"] == "warning" && obj["file"] == "p.md" {
			sawFinding = true
		}
		if s, ok := obj["summary"].(map[string]any); ok && s["warnings"] == float64(1) {
			sawSummary = true
		}
	}
	if !sawFinding || !sawSummary {
		t.Errorf("finding=%v summary=%v:\n%s", sawFinding, sawSummary, out)
	}
	// a tool failure is also JSON on standard output (workspace DR-0011)
	out, err = runValidate(t, "-json", "nosuch.md")
	wantClass(t, "missing path in json mode", err, classNoInput)
	var obj map[string]any
	if e := json.Unmarshal([]byte(strings.TrimSpace(out)), &obj); e != nil || obj["class"] != "no_input" || obj["code"] != float64(66) {
		t.Errorf("error object: %q (%v)", out, e)
	}
}
