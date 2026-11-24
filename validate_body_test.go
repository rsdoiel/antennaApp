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

// bodySite is a workspace made the way a person makes one: antenna init, the
// page about.md rendered with antenna page, and a news.md collection with its
// generated items page published beside it. It returns nothing; the files are
// in the working directory (htdocs is the working directory, the default).
func bodySite(t *testing.T) {
	t.Helper()
	newWorkspace(t, true)
	write(t, "about.md", "---\ntitle: About\n---\n# About\n")
	if err := runIn(t, "page", "about.md", "about.html"); err != nil {
		t.Fatalf("page: %s", err)
	}
	write(t, "news.md", "---\ntitle: News\ndescription: d\n---\n# News\n\n- [Feed](https://example.org/feed.xml)\n")
	if err := runIn(t, "add", "news.md"); err != nil {
		t.Fatalf("add: %s", err)
	}
	write(t, "news_items.md", "---\naggregateOf: news.md\n---\n# generated\n")
	write(t, "unlisted.md", "---\ntitle: Unlisted\n---\n# Unlisted\n")
}

// post is a Markdown file whose front matter is complete, so only its body
// can draw a finding; body starts on line 6.
func bodyDoc(body string) string {
	return "---\ntitle: T\npostPath: docs/p.md\npubDate: \"Thu, 19 Nov 2026 12:00:00 -0800\"\n---\n" + body
}

func TestValidateLinkHTMLNamesTheMarkdownToLinkTo(t *testing.T) {
	bodySite(t)
	write(t, "docs/p.md", bodyDoc("[About](../about.html#team)\n\n[News](../news_items.html)\n\n[Elsewhere](https://other.example/x.html)\n\n[feed](../index.xml)\n\n```\n[code](../about.html)\n```\n\n`[span](../about.html)`\n"))
	out, err := runValidate(t, "docs/p.md")
	wantClass(t, "link-html findings", err, classNegative)
	line, ok := finding(out, "docs/p.md", "warning", "link-html")
	if !ok || !strings.HasPrefix(line, "docs/p.md:6: ") || !strings.Contains(line, "../about.md#team") {
		t.Errorf("about.html: %q\n%s", line, out)
	}
	if !strings.Contains(out, "../news_items.md") {
		t.Errorf("a collection's HTML page should name its items Markdown:\n%s", out)
	}
	if n := strings.Count(out, ": link-html: "); n != 2 {
		t.Errorf("want exactly 2 link-html findings (not the external, the .xml, the code), got %d:\n%s", n, out)
	}
}

func TestValidateLinkHTMLToACollectionWithNoItemsFileIsLeftAlone(t *testing.T) {
	bodySite(t)
	os.Remove("news_items.md")
	write(t, "docs/p.md", bodyDoc("[News](../news_items.html)\n"))
	out, _ := runValidate(t, "docs/p.md")
	if strings.Contains(out, "link-html") {
		t.Errorf("no items Markdown exists, so there is nothing to link to:\n%s", out)
	}
}

func TestValidateLinkUnrenderedAndLinkMissing(t *testing.T) {
	bodySite(t)
	write(t, "docs/p.md", bodyDoc("[exists](../unlisted.md)\n\n[gone](../nowhere.md)\n\n[txt](../notes.txt)\n\n[dir](../series/)\n\n[frag](#top)\n\n[mail](mailto:a@example.org)\n"))
	out, _ := runValidate(t, "docs/p.md")
	if _, ok := finding(out, "docs/p.md", "warning", "link-unrendered"); !ok || !strings.Contains(out, "unlisted.md") {
		t.Errorf("no link-unrendered for a file antenna does not render:\n%s", out)
	}
	if n := strings.Count(out, ": link-missing: "); n != 2 {
		t.Errorf("want link-missing for nowhere.md and notes.txt only, got %d:\n%s", n, out)
	}
	if strings.Contains(out, "series/") || strings.Contains(out, "#top") || strings.Contains(out, "mailto") {
		t.Errorf("a directory, a fragment or a mailto link was reported:\n%s", out)
	}
}

func TestValidateLinkAbsoluteSelfIsInfo(t *testing.T) {
	bodySite(t)
	// base_url covers the whole host, which also holds other sites: only an
	// address that is a page or file of this site is "on this site".
	write(t, "docs/p.md", bodyDoc("[About](http://localhost:8000/about.html)\n\n[Other site](http://localhost:8000/othersite/)\n\n[Elsewhere](http://localhost:8000/mkpage/docs.html)\n"))
	out, err := runValidate(t, "docs/p.md")
	if _, ok := finding(out, "docs/p.md", "info", "link-absolute-self"); !ok {
		t.Errorf("no link-absolute-self:\n%s", out)
	}
	if n := strings.Count(out, "link-absolute-self"); n != 1 {
		t.Errorf("want one finding, for about.html only (not another site on the host), got %d:\n%s", n, out)
	}
	if err != nil {
		t.Errorf("info alone must exit 0, got %v", err)
	}
}

func TestValidateRawHTML(t *testing.T) {
	bodySite(t)
	write(t, "docs/p.md", bodyDoc("<div class=\"x\">block</div>\n\nSome <span>inline</span> text.\n\n<script>alert(1)</script>\n\n<svg><circle/></svg>\n\n<!-- a comment is fine -->\n\n```html\n<div>in code</div>\n```\n"))
	out, _ := runValidate(t, "docs/p.md")
	if n := strings.Count(out, ": raw-html: "); n < 4 {
		t.Errorf("want findings for the div, span, script and svg, got %d:\n%s", n, out)
	}
	for _, tag := range []string{"<div>", "<span>", "<script>", "<svg>"} {
		if !strings.Contains(out, tag) {
			t.Errorf("the finding should name %s:\n%s", tag, out)
		}
	}
	if strings.Contains(out, "comment") || strings.Contains(out, "in code") {
		t.Errorf("a comment or code was reported:\n%s", out)
	}
}

func TestValidateImageType(t *testing.T) {
	bodySite(t)
	write(t, "docs/p.md", bodyDoc("![a](pic.svg)\n\n![b](pic.png)\n\n![c](pic.WEBP)\n\n![d](https://example.org/y.jpeg?w=3)\n\n![e](pic.bmp)\n\n![f](https://example.org/photo)\n"))
	out, _ := runValidate(t, "docs/p.md")
	if n := strings.Count(out, ": image-type: "); n != 2 || !strings.Contains(out, "pic.svg") || !strings.Contains(out, "pic.bmp") {
		t.Errorf("want image-type for pic.svg and pic.bmp only, got %d:\n%s", n, out)
	}
}

func TestValidateMacroMissing(t *testing.T) {
	bodySite(t)
	write(t, "docs/p.md", bodyDoc("# {{page.title}}\n\nBy {{page.nosuchkey}}.\n\nCode `{{page.alsono}}`.\n"))
	out, _ := runValidate(t, "docs/p.md")
	if line, ok := finding(out, "docs/p.md", "warning", "macro-missing"); !ok || !strings.Contains(line, "nosuchkey") || !strings.HasPrefix(line, "docs/p.md:8: ") {
		t.Errorf("no macro-missing for nosuchkey at line 8: %q\n%s", line, out)
	}
	if strings.Contains(out, "alsono") || strings.Count(out, "macro-missing") != 1 {
		t.Errorf("a macro in code or a present key was reported:\n%s", out)
	}
}

// Without antenna.yaml the checks that need no configuration still run; the
// ones that need to know what antenna renders are skipped, never wrong.
func TestValidateAFileWithNoConfigurationStillChecksItsBody(t *testing.T) {
	t.Chdir(t.TempDir())
	write(t, "p.md", "---\ntitle: T\n---\n[gone](nowhere.md)\n\n<div>x</div>\n\n[page](other.html)\n")
	out, err := runValidate(t, "p.md")
	wantClass(t, "findings with no antenna.yaml", err, classNegative)
	for _, code := range []string{"link-missing", "raw-html"} {
		if !strings.Contains(out, ": "+code+": ") {
			t.Errorf("no %s:\n%s", code, out)
		}
	}
	if strings.Contains(out, "link-html") || strings.Contains(out, "link-unrendered") {
		t.Errorf("a check that needs antenna.yaml ran without it:\n%s", out)
	}
}
