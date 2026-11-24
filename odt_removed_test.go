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
	"strings"
	"testing"
)

// DR-0032: antenna no longer reads ODT or OTT. A file with that extension is
// refused with a message that says so, never read as Markdown.

func TestLoadCommonMarkRefusesODT(t *testing.T) {
	newWorkspace(t, false)
	for _, name := range []string{"doc.odt", "doc.ott", "DOC.ODT"} {
		os.WriteFile(name, []byte("PK\x03\x04 not really a zip"), 0o644)
		_, err := LoadCommonMark(name)
		wantClass(t, name, err, classData)
		if err != nil && !strings.Contains(err.Error(), "ODT") {
			t.Errorf("%s: the message does not say ODT is not supported: %v", name, err)
		}
	}
	os.WriteFile("doc.md", []byte("---\ntitle: T\n---\nBody\n"), 0o644)
	if _, err := LoadCommonMark("doc.md"); err != nil {
		t.Errorf("a Markdown file was refused: %v", err)
	}
}

func TestAddAndPostRefuseAnODTFile(t *testing.T) {
	newWorkspace(t, true)
	os.WriteFile("feeds.odt", []byte("PK\x03\x04"), 0o644)
	wantClass(t, "add of an ODT collection", runIn(t, "add", "feeds.odt"), classData)
	wantClass(t, "post of an ODT file", runIn(t, "post", "feeds.odt"), classData)
}

func TestPageRefusesAnODTSourceAsUsage(t *testing.T) {
	newWorkspace(t, true)
	os.WriteFile("doc.odt", []byte("PK\x03\x04"), 0o644)
	err := runIn(t, "page", "doc.odt")
	wantClass(t, "page of an ODT source", err, classUsage)
	if err != nil && !strings.Contains(err.Error(), "ODT") {
		t.Errorf("the message does not say ODT is not supported: %v", err)
	}
	if exists("doc.html") {
		t.Error("a page was written for an ODT source")
	}
}

func TestNormalizeToHTMLExtKnowsOnlyMarkdown(t *testing.T) {
	cases := map[string]string{
		"blog/a/post.md":  "blog/a/post.html",
		"blog/a/post.MD":  "blog/a/post.html",
		"blog/a/post.odt": "blog/a/post.odt",
		"blog/a/post.ott": "blog/a/post.ott",
		"blog/a/post.txt": "blog/a/post.txt",
	}
	for in, want := range cases {
		if got := normalizeToHTMLExt(in); got != want {
			t.Errorf("normalizeToHTMLExt(%q) = %q, want %q", in, got, want)
		}
	}
}
