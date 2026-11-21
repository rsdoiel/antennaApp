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
	"os"
	"regexp"
	"strings"
	"testing"
)

// DR-0020: validate suggests front matter and colors its report.

const suggestDate = "Thu, 19 Nov 2026 12:00:00 -0800"

// fullDoc has every key validate suggests, so it draws nothing.
func fullDoc(extra string) string {
	return "---\ntitle: T\ndescription: A short description.\nauthor: R. S. Doiel\npostPath: docs/p.md\npubDate: \"" + suggestDate + "\"\n" + extra + "---\nBody.\n"
}

func setConfig(t *testing.T, line string) {
	t.Helper()
	b, err := os.ReadFile("antenna.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("antenna.yaml", append(b, []byte("\n"+line+"\n")...), 0o664); err != nil {
		t.Fatal(err)
	}
}

func TestValidateAFullyDescribedPostDrawsNoSuggestion(t *testing.T) {
	bodySite(t)
	write(t, "docs/p.md", fullDoc(""))
	out, err := runValidate(t, "docs/p.md")
	if err != nil {
		t.Fatalf("%s\n%s", err, out)
	}
	for _, code := range []string{"no-description", "no-author", "title-from-heading", "description-long"} {
		if strings.Contains(out, code) {
			t.Errorf("%s fired on a complete file:\n%s", code, out)
		}
	}
}

func TestValidateNoDescriptionAndNoAuthorAreInfo(t *testing.T) {
	bodySite(t)
	write(t, "docs/p.md", "---\ntitle: T\npostPath: docs/p.md\npubDate: \""+suggestDate+"\"\n---\nBody.\n")
	out, err := runValidate(t, "docs/p.md")
	if err != nil {
		t.Errorf("info findings must not change the exit status: %s", err)
	}
	for _, code := range []string{"no-description", "no-author"} {
		if _, ok := finding(out, "docs/p.md", "info", code); !ok {
			t.Errorf("want an info %s:\n%s", code, out)
		}
	}
	// An empty description is no description.
	write(t, "docs/q.md", "---\ntitle: T\ndescription: \"  \"\nauthor: A\n---\nBody.\n")
	out, _ = runValidate(t, "docs/q.md")
	if _, ok := finding(out, "docs/q.md", "info", "no-description"); !ok {
		t.Errorf("a blank description is none:\n%s", out)
	}
}

func TestValidateTheSiteAuthorSilencesNoAuthor(t *testing.T) {
	bodySite(t)
	setConfig(t, "author: R. S. Doiel")
	write(t, "docs/p.md", "---\ntitle: T\ndescription: d\n---\nBody.\n")
	out, _ := runValidate(t, "docs/p.md")
	if strings.Contains(out, "no-author") {
		t.Errorf("antenna.yaml names an author:\n%s", out)
	}
}

func TestValidateTitleFromHeading(t *testing.T) {
	bodySite(t)
	write(t, "docs/p.md", "---\ndescription: d\nauthor: A\n---\n# The *Real* Title\n\nBody.\n")
	out, _ := runValidate(t, "docs/p.md")
	line, ok := finding(out, "docs/p.md", "info", "title-from-heading")
	if !ok || !strings.Contains(line, "The Real Title") {
		t.Errorf("want title-from-heading naming the heading: %q\n%s", line, out)
	}
	if strings.Contains(out, "no-title") {
		t.Errorf("no-title is for no heading either:\n%s", out)
	}
	write(t, "docs/r.md", "---\ndescription: d\nauthor: A\n---\nBody only.\n")
	out, _ = runValidate(t, "docs/r.md")
	if _, ok := finding(out, "docs/r.md", "warning", "no-title"); !ok || strings.Contains(out, "title-from-heading") {
		t.Errorf("no heading keeps no-title:\n%s", out)
	}
}

func TestValidateDescriptionLong(t *testing.T) {
	bodySite(t)
	long := strings.Repeat("x", 301)
	write(t, "docs/p.md", fullDoc("")+"")
	write(t, "docs/long.md", "---\ntitle: T\ndescription: "+long+"\nauthor: A\n---\nBody.\n")
	write(t, "docs/edge.md", "---\ntitle: T\ndescription: "+strings.Repeat("x", 300)+"\nauthor: A\n---\nBody.\n")
	out, _ := runValidate(t, "docs/long.md", )
	if _, ok := finding(out, "docs/long.md", "info", "description-long"); !ok {
		t.Errorf("301 characters is long:\n%s", out)
	}
	out, _ = runValidate(t, "docs/edge.md")
	if strings.Contains(out, "description-long") {
		t.Errorf("300 characters is the limit, not over it:\n%s", out)
	}
	setConfig(t, "descriptionMax: 500")
	out, _ = runValidate(t, "docs/long.md")
	if strings.Contains(out, "description-long") {
		t.Errorf("descriptionMax 500 allows 301:\n%s", out)
	}
}

func TestValidateDescriptionMaxFromConfig(t *testing.T) {
	bodySite(t)
	write(t, "docs/s.md", "---\ntitle: T\ndescription: "+strings.Repeat("y", 60)+"\nauthor: A\n---\nBody.\n")
	setConfig(t, "descriptionMax: 50")
	out, _ := runValidate(t, "docs/s.md")
	if _, ok := finding(out, "docs/s.md", "info", "description-long"); !ok {
		t.Errorf("60 over a limit of 50:\n%s", out)
	}
}

func TestValidateDescriptionMaxZeroTurnsItOff(t *testing.T) {
	bodySite(t)
	write(t, "docs/l.md", "---\ntitle: T\ndescription: "+strings.Repeat("z", 5000)+"\nauthor: A\n---\nBody.\n")
	setConfig(t, "descriptionMax: 0")
	out, _ := runValidate(t, "docs/l.md")
	if strings.Contains(out, "description-long") {
		t.Errorf("0 turns the check off:\n%s", out)
	}
}

func TestValidateBadDescriptionMaxIsAConfigError(t *testing.T) {
	for _, bad := range []string{"descriptionMax: -1", "descriptionMax: lots"} {
		bodySite(t)
		setConfig(t, bad)
		write(t, "docs/p.md", fullDoc(""))
		_, err := runValidate(t, "docs/p.md")
		wantClass(t, bad, err, classConfig)
	}
}

func TestValidateGeneratedItemsPagesGetNoSuggestions(t *testing.T) {
	bodySite(t)
	out, _ := runValidate(t, "news_items.md")
	if strings.Contains(out, "no-description") || strings.Contains(out, "no-author") {
		t.Errorf("a generated items page is not the author's to fix:\n%s", out)
	}
}

// ---- color (DR-0020 5) ----

var ansi = regexp.MustCompile("\x1b\\[[0-9;]*m")

func TestValidateColorAlwaysAddsEscapesAndNothingElse(t *testing.T) {
	bodySite(t)
	write(t, "docs/p.md", "---\ntitle: T\n---\n[x](../about.html)\n")
	plain, _ := runValidate(t, "docs/p.md")
	colored, _ := runValidate(t, "-color", "always", "docs/p.md")
	if !strings.Contains(colored, "\x1b[") {
		t.Fatalf("-color always wrote no escapes:\n%s", colored)
	}
	if got := ansi.ReplaceAllString(colored, ""); got != plain {
		t.Errorf("the text must be the same without the escapes:\n%q\n%q", got, plain)
	}
	if strings.Contains(plain, "\x1b[") {
		t.Errorf("auto on a buffer (not a terminal) must not color:\n%s", plain)
	}
	eq, _ := runValidate(t, "-color=always", "docs/p.md")
	if eq != colored {
		t.Errorf("-color=always should equal -color always")
	}
	never, _ := runValidate(t, "-color", "never", "docs/p.md")
	if never != plain {
		t.Errorf("-color never should be plain")
	}
}

func TestValidateColorAutoOnATerminalAndNoColor(t *testing.T) {
	bodySite(t)
	write(t, "docs/p.md", "---\ntitle: T\n---\n[x](../about.html)\n")
	old := outIsTerminal
	outIsTerminal = func(any) bool { return true }
	defer func() { outIsTerminal = old }()
	auto, _ := runValidate(t, "docs/p.md")
	if !strings.Contains(auto, "\x1b[") {
		t.Errorf("auto on a terminal should color:\n%s", auto)
	}
	t.Setenv("NO_COLOR", "1")
	off, _ := runValidate(t, "docs/p.md")
	if strings.Contains(off, "\x1b[") {
		t.Errorf("NO_COLOR turns auto off:\n%s", off)
	}
	forced, _ := runValidate(t, "-color", "always", "docs/p.md")
	if !strings.Contains(forced, "\x1b[") {
		t.Errorf("-color always overrides NO_COLOR")
	}
	t.Setenv("NO_COLOR", "")
	empty, _ := runValidate(t, "docs/p.md")
	if !strings.Contains(empty, "\x1b[") {
		t.Errorf("an empty NO_COLOR is unset")
	}
}

func TestValidateColorNeverTouchesJSON(t *testing.T) {
	bodySite(t)
	write(t, "docs/p.md", "---\ntitle: T\n---\n[x](../about.html)\n")
	out, _ := runValidate(t, "-json", "-color", "always", "docs/p.md")
	if strings.Contains(out, "\x1b") {
		t.Errorf("-json is never colored:\n%s", out)
	}
}

func TestValidateBadColorValueIsUsage(t *testing.T) {
	bodySite(t)
	_, err := runValidate(t, "-color", "rainbow")
	wantClass(t, "bad -color", err, classUsage)
	_, err = runValidate(t, "-color")
	wantClass(t, "-color with no value", err, classUsage)
}
