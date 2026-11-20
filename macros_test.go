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
*/
package antennaApp

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// macroDoc parses src as a page and turns macro expansion on, with the
// warnings going to warn.
func macroDoc(t *testing.T, src string, warn *bytes.Buffer) *CommonMark {
	t.Helper()
	doc := &CommonMark{}
	if err := doc.Parse([]byte(src)); err != nil {
		t.Fatal(err)
	}
	doc.Macros = &MacroExpander{From: "p.html", Warn: warn}
	return doc
}

func TestMacrosExpandPageKeys(t *testing.T) {
	var warn bytes.Buffer
	doc := macroDoc(t, "---\ntitle: Hello & Goodbye\nauthor: Dave\npub_date: 2026-11-19\nseriesNumber: 3\n---\n# {{page.title}}\n\nBy {{ page.author }} on {{page.pub_date}}, part {{page.seriesNumber}}.\n", &warn)
	got, err := doc.ToUnsafeHTML()
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"<h1>Hello &amp; Goodbye</h1>", "By Dave on 2026-11-19, part 3."} {
		if !strings.Contains(got, want) {
			t.Errorf("%q not in:\n%s", want, got)
		}
	}
	if warn.Len() != 0 {
		t.Errorf("warning: %s", warn.String())
	}
}

func TestMacrosAreNotExpandedInCode(t *testing.T) {
	var warn bytes.Buffer
	doc := macroDoc(t, "---\ntitle: Hello\n---\nUse `{{page.title}}` here.\n\n```\n{{page.title}}\n```\n\n    {{page.title}}\n\n{{page.title}}\n", &warn)
	got, _ := doc.ToUnsafeHTML()
	if strings.Count(got, "{{page.title}}") != 3 || strings.Count(got, "Hello") != 1 {
		t.Errorf("code expanded, or text not:\n%s", got)
	}
}

func TestMacrosOnePassNoLoop(t *testing.T) {
	var warn bytes.Buffer
	doc := macroDoc(t, "---\ntitle: \"{{page.author}}\"\nauthor: Dave\n---\n{{page.title}}\n", &warn)
	got, _ := doc.ToUnsafeHTML()
	if strings.Contains(got, "Dave") || !strings.Contains(got, "{{page.author}}") {
		t.Errorf("a value was expanded again:\n%s", got)
	}
}

func TestMacrosMissingKeyStaysVisibleAndWarns(t *testing.T) {
	var warn bytes.Buffer
	doc := macroDoc(t, "---\ntitle: T\n---\n{{page.nope}} {{page.nope}} {{page.title}}\n", &warn)
	got, _ := doc.ToUnsafeHTML()
	if !strings.Contains(got, "{{page.nope}} {{page.nope}} T") {
		t.Errorf("got:\n%s", got)
	}
	if strings.Count(warn.String(), "\n") != 1 || !strings.Contains(warn.String(), "nope") || !strings.Contains(warn.String(), "warning:") || !strings.Contains(warn.String(), "p.html") {
		t.Errorf("warnings (want one line naming the page and key): %q", warn.String())
	}
}

func TestMacroValuesCannotInjectMarkupOrHTML(t *testing.T) {
	var warn bytes.Buffer
	doc := macroDoc(t, "---\ntitle: \"<b>x</b> *y* [z](http://evil.example/)\"\n---\n{{page.title}}\n", &warn)
	got, _ := doc.ToUnsafeHTML()
	for _, bad := range []string{"<b>", "<em>", "<a "} {
		if strings.Contains(got, bad) {
			t.Errorf("%q injected:\n%s", bad, got)
		}
	}
	if !strings.Contains(got, "&lt;b&gt;x&lt;/b&gt; *y* [z](http://evil.example/)") {
		t.Errorf("value not shown as text:\n%s", got)
	}
}

func TestMacroTitleBecomesLinkWhenLinkIsSet(t *testing.T) {
	var warn bytes.Buffer
	doc := macroDoc(t, "---\ntitle: Hello\nlink: https://x.example/a?b=1&c=2\n---\n# {{page.title}}\n", &warn)
	got, _ := doc.ToUnsafeHTML()
	if !strings.Contains(got, `<h1><a href="https://x.example/a?b=1&amp;c=2">Hello</a></h1>`) {
		t.Errorf("got:\n%s", got)
	}
	// only the title is linked
	doc = macroDoc(t, "---\ntitle: Hello\nauthor: Dave\nlink: https://x.example/\n---\n{{page.author}}\n", &warn)
	if got, _ := doc.ToUnsafeHTML(); strings.Contains(got, "<a ") {
		t.Errorf("author linked:\n%s", got)
	}
}

func TestMacroTitleLinkRefusesScriptSchemes(t *testing.T) {
	var warn bytes.Buffer
	doc := macroDoc(t, "---\ntitle: Hello\nlink: \"javascript:alert(1)\"\n---\n# {{page.title}}\n", &warn)
	got, _ := doc.ToUnsafeHTML()
	if strings.Contains(got, "javascript:") || strings.Contains(got, "<a ") {
		t.Errorf("script link made:\n%s", got)
	}
	if !strings.Contains(got, "<h1>Hello</h1>") || !strings.Contains(warn.String(), "link") {
		t.Errorf("got:\n%s\nwarn: %q", got, warn.String())
	}
}

func TestMacrosOffByDefault(t *testing.T) {
	// harvested feed items are rendered without a MacroExpander
	doc := &CommonMark{}
	if err := doc.Parse([]byte("---\ntitle: T\n---\n{{page.title}}\n")); err != nil {
		t.Fatal(err)
	}
	if got, _ := doc.ToHTML(); !strings.Contains(got, "{{page.title}}") {
		t.Errorf("expanded without being asked:\n%s", got)
	}
}

// The real path: GeneratePages writes HTML with the macros expanded, and says
// so on stderr for a key the page does not have.
func TestGeneratePagesExpandsMacros(t *testing.T) {
	cfg := sitemapSite(t, "pages.md")
	cfg.Htdocs = "htdocs"
	cfg.Generator = "page.yaml"
	if err := os.WriteFile("page.yaml", []byte("header: |\n  <h1>Test</h1>\n"), 0664); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cfg.Htdocs, 0775); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("story.md", []byte("---\ntitle: A Story\nauthor: Dave\n---\n# {{page.title}}\n\nBy {{page.author}} on {{page.pubDate}}.\n"), 0664); err != nil {
		t.Fatal(err)
	}
	addPage(t, "pages.db", "story.html")
	var eout bytes.Buffer
	if err := cfg.GeneratePages(&eout); err != nil {
		t.Fatalf("GeneratePages: %s\n%s", err, eout.String())
	}
	got, _ := os.ReadFile(filepath.Join("htdocs", "story.html"))
	if !strings.Contains(string(got), "<h1>A Story</h1>") || !strings.Contains(string(got), "By Dave on {{page.pubDate}}.") {
		t.Errorf("story.html:\n%s", got)
	}
	if !strings.Contains(eout.String(), "pubDate") {
		t.Errorf("no warning for pubDate: %q", eout.String())
	}
	// the author's Markdown is untouched (DR-0009 6)
	src, _ := os.ReadFile("story.md")
	if !strings.Contains(string(src), "{{page.title}}") {
		t.Errorf("source Markdown was changed:\n%s", src)
	}
}
