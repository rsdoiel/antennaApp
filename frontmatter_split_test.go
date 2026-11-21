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
	"strings"
	"testing"
)

// DR-0022: front matter is only the block at the top of a file.

func TestSplitFrontMatterOnlyAtTheTop(t *testing.T) {
	cases := []struct {
		name, src, body string
		title           string // "" when no front matter
		class           ExitClass
		wantErr         bool
	}{
		{"rule in the body", "---\ntitle: T\n---\nBody.\n\n---\n\nafter rule\n", "Body.\n\n---\n\nafter rule\n", "T", ExitClass{}, false},
		{"two rules and no block", "Intro.\n\n---\nnot: yaml\n---\n\nEnd.\n", "Intro.\n\n---\nnot: yaml\n---\n\nEnd.\n", "", ExitClass{}, false},
		{"first line is not a rule", "# Heading\n---\ntitle: X\n---\n", "# Heading\n---\ntitle: X\n---\n", "", ExitClass{}, false},
		{"final newline kept", "---\ntitle: T\n---\nBody.\n", "Body.\n", "T", ExitClass{}, false},
		{"no final newline kept as is", "---\ntitle: T\n---\nBody.", "Body.", "T", ExitClass{}, false},
		{"crlf kept in the body", "---\r\ntitle: T\r\n---\r\nBody.\r\nMore.\r\n", "Body.\r\nMore.\r\n", "T", ExitClass{}, false},
		{"dots close the block", "---\ntitle: T\n...\nBody.\n", "Body.\n", "T", ExitClass{}, false},
		{"dots in the body are text", "---\ntitle: T\n---\nBody.\n...\nMore.\n", "Body.\n...\nMore.\n", "T", ExitClass{}, false},
		{"dots with no block are text", "Body.\n...\nMore.\n", "Body.\n...\nMore.\n", "", ExitClass{}, false},
		{"empty block", "---\n---\n# Empty\n\nText.\n", "# Empty\n\nText.\n", "", ExitClass{}, false},
		{"unclosed block", "---\ntitle: T\nBody.\n", "", "", classData, true},
	}
	for _, c := range cases {
		fm, body, err := SplitFrontMatter([]byte(c.src))
		if c.wantErr {
			wantClass(t, c.name, err, c.class)
			continue
		}
		if err != nil {
			t.Errorf("%s: %s", c.name, err)
			continue
		}
		if body != c.body {
			t.Errorf("%s: body %q, want %q", c.name, body, c.body)
		}
		got, _ := fm["title"].(string)
		if got != c.title {
			t.Errorf("%s: title %q, want %q", c.name, got, c.title)
		}
	}
}

// The reproduction: a thematic break after the front matter made validate say
// "the front matter does not parse".
func TestValidateARuleInTheBodyIsNotFrontMatter(t *testing.T) {
	bodySite(t)
	write(t, "docs/p.md", fullDoc("")+"\n---\n\nafter rule\n")
	out, err := runValidate(t, "docs/p.md")
	if err != nil {
		t.Fatalf("a rule in the body is valid Markdown: %s\n%s", err, out)
	}
	if strings.Contains(out, "front-matter") {
		t.Errorf("front matter finding on a rule:\n%s", out)
	}
}

// Findings after a rule point at the right line.
func TestValidateLineNumbersAfterARule(t *testing.T) {
	bodySite(t)
	write(t, "docs/hr.md", bodyDoc("Para.\n\n---\n\n[About](../about.html)\n"))
	out, _ := runValidate(t, "docs/hr.md")
	line, ok := finding(out, "docs/hr.md", "warning", "link-html")
	// bodyDoc puts the body on line 6: Para 6, blank 7, rule 8, blank 9, link 10.
	if !ok || !strings.HasPrefix(line, "docs/hr.md:10: ") {
		t.Errorf("link-html should be on line 10: %q\n%s", line, out)
	}
}

func TestPostAndPageAcceptARuleInTheBody(t *testing.T) {
	markdownPageSite(t)
	src := "---\ntitle: Rule\ndescription: d\nauthor: A\npostPath: blog/2026/11/19/rule.md\npubDate: \"2026-11-19\"\n---\nAbove.\n\n---\n\nBelow.\n"
	write(t, "blog/2026/11/19/rule.md", src)
	if err := runIn(t, "post", "blog/2026/11/19/rule.md"); err != nil {
		t.Fatalf("post: %s", err)
	}
	write(t, "rule.txt", "---\ntitle: Rule page\n---\nAbove.\n\n---\n\nBelow.\n")
	if err := runIn(t, "page", "rule.txt", "rule.md", "rule.html"); err != nil {
		t.Fatalf("page: %s", err)
	}
	if got := readText(t, "rule.md"); !strings.HasSuffix(got, "Above.\n\n---\n\nBelow.\n") {
		t.Errorf("the copy keeps the rule and the text after it:\n%q", got)
	}
	if html := readText(t, "rule.html"); !strings.Contains(html, "<hr") || !strings.Contains(html, "Below.") {
		t.Errorf("the page has the rule and the text after it:\n%s", html)
	}
}
