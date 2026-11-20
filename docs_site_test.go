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
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/rsdoiel/opml"
)

// The documentation is a Markdown site (README.md and the pages beside it) that
// is published as HTML and as Markdown for BonBon and other Markdown readers.
// These tests keep it that way: the pages link to each other as Markdown, the
// table of contents lists every page, each HTML page declares its Markdown, and
// the built site is a Markdown site. The pattern is cmarkbrowser's.

// makeVar reads a "NAME = value" line of website.mak as fields.
func makeVar(t *testing.T, name string) []string {
	t.Helper()
	b, err := os.ReadFile("website.mak")
	if err != nil {
		t.Fatal(err)
	}
	m := regexp.MustCompile(`(?m)^` + name + ` = (.*)$`).FindSubmatch(b)
	if m == nil {
		t.Fatalf("website.mak has no %s", name)
	}
	return strings.Fields(string(m[1]))
}

// manualPage reports whether a file is a generated manual page (antenna.1.md,
// antenna-TOPIC.7.md): the website lists those itself, apart from PAGE_ORDER.
func manualPage(name string) bool {
	return regexp.MustCompile(`^antenna(-[a-z-]+)?\.[0-9]\.md$`).MatchString(name)
}

// publishedPages are the Markdown files in the folder that the site publishes.
func publishedPages(t *testing.T) []string {
	t.Helper()
	skip := map[string]bool{}
	for _, f := range makeVar(t, "NO_PUBLISH") {
		skip[f] = true
	}
	all, _ := filepath.Glob("*.md")
	var out []string
	for _, f := range all {
		if !skip[f] {
			out = append(out, f)
		}
	}
	sort.Strings(out)
	return out
}

func siteNeeds(t *testing.T, tools ...string) {
	t.Helper()
	for _, tool := range tools {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s is not installed", tool)
		}
	}
}

// siteRun runs a command in dir and returns its standard output and exit code.
func siteRun(t *testing.T, dir, name string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	err := cmd.Run()
	if err == nil {
		return out.String(), 0
	}
	if ee, ok := err.(*exec.ExitError); ok {
		return out.String() + errb.String(), ee.ExitCode()
	}
	t.Fatalf("%s: %s", name, err)
	return "", -1
}

// outlineURLs returns the url of every leaf outline, in order, and the text of each.
func outlineEntries(t *testing.T, src string) (urls, texts []string) {
	t.Helper()
	o, err := opml.Parse([]byte(src))
	if err != nil {
		t.Fatalf("not an outline: %s\n%s", err, src)
	}
	var walk func([]*opml.Outline)
	walk = func(os []*opml.Outline) {
		for _, n := range os {
			if n.URL != "" {
				urls = append(urls, n.URL)
				texts = append(texts, n.Text)
			}
			walk(n.Outline)
		}
	}
	walk(o.Body.Outline)
	return urls, texts
}

// Every published page is listed in the table of contents: PAGE_ORDER names the
// guides, and the manual pages are the ones the program generates.
func TestEveryPublishedPageIsInTheSite(t *testing.T) {
	order := makeVar(t, "PAGE_ORDER")
	if order[0] != "README.md" {
		t.Errorf("the home page must come first: %v", order[:1])
	}
	listed := append([]string(nil), order...)
	for _, p := range publishedPages(t) {
		if manualPage(p) {
			listed = append(listed, p)
		}
	}
	sort.Strings(listed)
	if got, want := strings.Join(listed, " "), strings.Join(publishedPages(t), " "); got != want {
		t.Errorf("website.mak lists\n  %s\nbut the published pages are\n  %s", got, want)
	}
	for _, p := range order {
		if manualPage(p) {
			t.Errorf("%s is a manual page; website.mak lists those itself, not in PAGE_ORDER", p)
		}
	}
}

// The Markdown is what a Markdown reader follows, so a link to another page is
// written as .md (the HTML build changes it).
func TestSitePagesLinkToEachOtherAsMarkdown(t *testing.T) {
	skip := map[string]bool{}
	for _, f := range makeVar(t, "NO_PUBLISH") {
		skip[f] = true
	}
	link := regexp.MustCompile(`\]\(([^)\s]+)\)`)
	for _, page := range publishedPages(t) {
		b, err := os.ReadFile(page)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range link.FindAllStringSubmatch(withoutCode(string(b)), -1) {
			dest := m[1]
			if strings.Contains(dest, "://") || strings.HasPrefix(dest, "#") || strings.HasPrefix(dest, "mailto:") {
				continue
			}
			file := strings.SplitN(strings.SplitN(dest, "#", 2)[0], "?", 2)[0]
			if file == "" {
				continue
			}
			if strings.HasSuffix(file, ".html") {
				t.Errorf("%s links to %q: a link to another page is written as .md (the HTML build changes it)", page, dest)
				continue
			}
			if !strings.HasSuffix(file, ".md") {
				continue // a file such as LICENSE or a folder, not a page
			}
			if _, err := os.Stat(file); err != nil || skip[file] {
				t.Errorf("%s links to %q, which is not a published page", page, dest)
			}
		}
	}
}

// withoutCode removes fenced and indented code blocks and inline code spans, so
// an example of a link is not mistaken for one.
func withoutCode(md string) string {
	var keep []string
	fence := ""
	for _, line := range strings.Split(md, "\n") {
		trimmed := strings.TrimSpace(line)
		switch {
		case fence != "":
			if strings.HasPrefix(trimmed, fence) {
				fence = ""
			}
		case strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~"):
			fence = trimmed[:3]
		case strings.HasPrefix(line, "    ") || strings.HasPrefix(line, "\t"):
			// an indented code block (or a list continuation; neither is a page link here)
		default:
			keep = append(keep, regexp.MustCompile("`[^`]*`").ReplaceAllString(line, ""))
		}
	}
	return strings.Join(keep, "\n")
}

func TestSiteTocScriptWritesAnOutlineTheReaderUnderstands(t *testing.T) {
	siteNeeds(t, "bash")
	order := makeVar(t, "PAGE_ORDER")
	manuals := []string{"antenna.1.md", "antenna-generate.7.md", "antenna-markdown-web.7.md"}
	args := append(append([]string{"toc.bash", "https://example.org/docs/", "Docs & <More>"}, order...), manuals...)
	out, code := siteRun(t, ".", "bash", args...)
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, out)
	}
	urls, texts := outlineEntries(t, out)
	all := append(append([]string(nil), order...), manuals...)
	if len(urls) != len(all) || urls[0] != "https://example.org/docs/index.md" {
		t.Fatalf("entries %v", urls)
	}
	for i, p := range all[1:] {
		if urls[i+1] != "https://example.org/docs/"+p {
			t.Errorf("entry %d is %s, want %s", i+1, urls[i+1], p)
		}
	}
	if !strings.Contains(out, "Docs &amp; &lt;More&gt;") {
		t.Errorf("the title was not escaped:\n%s", out)
	}
	// a manual page is named for its command, not by its first heading, "NAME"
	for i, p := range manuals {
		want := map[string]string{"antenna.1.md": "antenna", "antenna-generate.7.md": "antenna generate", "antenna-markdown-web.7.md": "antenna markdown-web"}[p]
		if got := texts[len(order)+i]; got != want {
			t.Errorf("%s is titled %q, want %q", p, got, want)
		}
	}
	if again, _ := siteRun(t, ".", "bash", args...); again != out {
		t.Error("the same input gave a different table of contents")
	}
}

func TestSiteTocScriptRefusesWhatItCannotDo(t *testing.T) {
	siteNeeds(t, "bash")
	if _, code := siteRun(t, ".", "bash", "toc.bash", "https://x.org/docs", "T", "README.md"); code != 2 {
		t.Errorf("a base URL with no trailing slash: exit %d, want 2", code)
	}
	if _, code := siteRun(t, ".", "bash", "toc.bash", "https://x.org/", "T", "no-such-page.md"); code != 66 {
		t.Errorf("a missing page: exit %d, want 66", code)
	}
	if _, code := siteRun(t, ".", "bash", "toc.bash"); code != 2 {
		t.Errorf("no arguments: exit %d, want 2", code)
	}
}

// copySite copies the files the website build reads into a new folder.
func copySite(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, f := range []string{"website.mak", "toc.bash", "page.tmpl", "links-to-html.lua", "add-col-scope.lua"} {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		os.WriteFile(filepath.Join(dir, f), b, 0o755)
	}
	for _, p := range append(publishedPages(t), "TODO.md") {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		os.WriteFile(filepath.Join(dir, p), b, 0o644)
	}
	return dir
}

// The built site is a Markdown site: index.md for the site's address, a table
// of contents that lists every page, .nojekyll so GitHub Pages serves
// .well-known, and each HTML page declares its Markdown.
func TestTheBuiltSiteIsAMarkdownSite(t *testing.T) {
	siteNeeds(t, "bash", "make", "pandoc")
	dir := copySite(t)
	if out, code := siteRun(t, dir, "make", "-f", "website.mak"); code != 0 {
		t.Fatalf("make -f website.mak: exit %d\n%s", code, out)
	}
	readme, _ := os.ReadFile("README.md")
	if got, err := os.ReadFile(filepath.Join(dir, "index.md")); err != nil || string(got) != string(readme) {
		t.Errorf("index.md is not a copy of README.md: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".nojekyll")); err != nil {
		t.Error("no .nojekyll: GitHub Pages hides .well-known without it")
	}
	tocSrc, err := os.ReadFile(filepath.Join(dir, ".well-known", "toc.opml"))
	if err != nil {
		t.Fatalf("no table of contents: %s", err)
	}
	urls, _ := outlineEntries(t, string(tocSrc))
	if len(urls) != len(publishedPages(t)) || urls[0] != "https://rsdoiel.github.io/antennaApp/index.md" {
		t.Errorf("the table of contents lists %d pages, want %d, first %q", len(urls), len(publishedPages(t)), urls[0])
	}
	for _, u := range urls {
		name := strings.TrimPrefix(u, "https://rsdoiel.github.io/antennaApp/")
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("the table of contents lists %s, which the site does not have", u)
		}
	}
	for _, tc := range []struct{ html, md string }{{"index.html", "index.md"}, {"user_manual.html", "user_manual.md"}, {"antenna-generate.7.html", "antenna-generate.7.md"}} {
		b, err := os.ReadFile(filepath.Join(dir, tc.html))
		if err != nil {
			t.Fatalf("%s was not built: %s", tc.html, err)
		}
		if want := `<link rel="alternate" type="text/markdown" href="` + tc.md + `">`; !strings.Contains(string(b), want) {
			t.Errorf("%s does not declare its Markdown (%s)", tc.html, want)
		}
	}
	// the stylesheet is relative, so the page works under /antennaApp/
	if b, _ := os.ReadFile(filepath.Join(dir, "index.html")); strings.Contains(string(b), `href="/css/site.css"`) {
		t.Error("the stylesheet link is root-relative; under a project path it points at another site")
	}
	// a rebuild changes nothing, so git shows only real changes
	before, _ := os.ReadFile(filepath.Join(dir, ".well-known", "toc.opml"))
	if _, code := siteRun(t, dir, "make", "-f", "website.mak"); code != 0 {
		t.Fatal("second build failed")
	}
	if after, _ := os.ReadFile(filepath.Join(dir, ".well-known", "toc.opml")); string(after) != string(before) {
		t.Error("rebuilding changed the table of contents")
	}
}

// A page that is not in PAGE_ORDER stops the build, so it is never left out of
// the table of contents by accident.
func TestTheBuildStopsForAPageThatIsNotListed(t *testing.T) {
	siteNeeds(t, "bash", "make", "pandoc")
	dir := copySite(t)
	os.WriteFile(filepath.Join(dir, "stray.md"), []byte("# Stray\n"), 0o644)
	out, code := siteRun(t, dir, "make", "-f", "website.mak")
	if code == 0 || !strings.Contains(out, "stray.md") {
		t.Errorf("exit %d, want a failure that names stray.md:\n%s", code, out)
	}
}

// The HTML links change only for the site's own relative Markdown pages.
func TestSiteHTMLLinksChangeOnlyForThePagesOwnMarkdown(t *testing.T) {
	siteNeeds(t, "pandoc")
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.md"), []byte("# A\n\n[b](b.md#top) [home](README.md) [out](https://example.org/x.md) [abs](/y.md) [md](notes.md.txt)\n"), 0o644)
	lua, _ := os.ReadFile("links-to-html.lua")
	os.WriteFile(filepath.Join(dir, "links-to-html.lua"), lua, 0o644)
	out, code := siteRun(t, dir, "pandoc", "-f", "markdown", "-t", "html5", "--lua-filter=links-to-html.lua", "a.md")
	if code != 0 {
		t.Fatalf("pandoc exit %d", code)
	}
	for _, want := range []string{`href="b.html#top"`, `href="index.html"`, `href="https://example.org/x.md"`, `href="/y.md"`, `href="notes.md.txt"`} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %s in\n%s", want, out)
		}
	}
}

// A page's entry is named by its first "# " heading, else the title in its
// front matter, else its file name with the underscores as spaces.
func TestSiteTocTitles(t *testing.T) {
	siteNeeds(t, "bash")
	script, _ := filepath.Abs("toc.bash")
	dir := t.TempDir()
	for name, content := range map[string]string{
		"heading.md":   "---\ntitle: Front matter\n---\n\n# The heading\n",
		"front.md":     "---\ntitle: \"Only front matter\"\n---\n\nText.\n",
		"Some_name.md": "Text with no heading and no front matter.\n",
	} {
		os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)
	}
	out, code := siteRun(t, dir, "bash", script, "https://example.org/", "T", "heading.md", "front.md", "Some_name.md")
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, out)
	}
	_, texts := outlineEntries(t, out)
	want := []string{"The heading", "Only front matter", "Some name"}
	if strings.Join(texts, "|") != strings.Join(want, "|") {
		t.Errorf("titles %v, want %v", texts, want)
	}
}
