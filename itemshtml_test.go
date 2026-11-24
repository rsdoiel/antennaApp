package antennaApp

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// DR-0034 1: an aggregate collection writes its items page as STEM_items.html,
// beside STEM_items.md, and no STEM.html. The page declares STEM_items.md as
// its Markdown alternate; STEM.xml is written as before.
func TestGenerateWritesItemsPageAsItemsHTML(t *testing.T) {
	_, htdocs := itemsMDSite(t,
		harvested("https://a.example/one", "One", "d1", "2026-11-18 10:00:00", "published", "Feed A"))
	page, err := os.ReadFile(filepath.Join(htdocs, "news_items.html"))
	if err != nil {
		t.Fatalf("news_items.html was not written: %s", err)
	}
	if !strings.Contains(string(page), "https://a.example/one") {
		t.Errorf("news_items.html does not list the item:\n%s", page)
	}
	if l := markdownAlternate(string(page)); !strings.Contains(l, `href="news_items.md"`) {
		t.Errorf("news_items.html does not declare news_items.md as its alternate: %q", l)
	}
	if _, err := os.Stat(filepath.Join(htdocs, "news.xml")); err != nil {
		t.Errorf("news.xml missing: %s", err)
	}
	if _, err := os.Stat(filepath.Join(htdocs, "news_items.md")); err != nil {
		t.Errorf("news_items.md missing: %s", err)
	}
	if _, err := os.Stat(filepath.Join(htdocs, "news.html")); err == nil {
		t.Error("news.html was written; the author's feed list has no HTML page (DR-0034 2)")
	}
}

// DR-0034 5: a collection with nothing to show removes a stale items page.
func TestAStaleItemsPageIsRemovedWhenTheCollectionIsEmpty(t *testing.T) {
	htdocs, _ := genCollection(t, "", func(htdocs string) {
		os.WriteFile(filepath.Join(htdocs, "news_items.html"), []byte("<html>stale</html>\n"), 0664)
	})
	if _, err := os.Stat(filepath.Join(htdocs, "news_items.html")); err == nil {
		t.Error("the stale news_items.html is still there")
	}
}

// DR-0034 3, 4: a page-index collection keeps STEM.html and gets no items page.
func TestAPageIndexCollectionKeepsItsPlainPage(t *testing.T) {
	cfg := sitemapSite(t, "news.md")
	htdocs := filepath.Join(t.TempDir(), "htdocs")
	if err := os.MkdirAll(htdocs, 0775); err != nil {
		t.Fatal(err)
	}
	cfg.Htdocs = htdocs
	col := cfg.Collections[0]
	col.Title = "News"
	col.Mode = "page-index"
	if err := col.Generate(os.Stderr, os.Stderr, "antenna", cfg); err != nil {
		t.Fatalf("Generate: %s", err)
	}
	if _, err := os.Stat(filepath.Join(htdocs, "news.html")); err != nil {
		t.Errorf("a page-index collection lost news.html: %s", err)
	}
	if _, err := os.Stat(filepath.Join(htdocs, "news_items.html")); err == nil {
		t.Error("a page-index collection wrote an items page")
	}
}

// DR-0031, DR-0034: a Markdown-only collection writes no HTML at all, and an
// items page left from before is removed.
func TestAMarkdownOnlyCollectionWritesNoItemsPage(t *testing.T) {
	cfg := sitemapSite(t, "news.md")
	htdocs := filepath.Join(t.TempDir(), "htdocs")
	if err := os.MkdirAll(htdocs, 0775); err != nil {
		t.Fatal(err)
	}
	cfg.Htdocs = htdocs
	off := false
	col := cfg.Collections[0]
	col.Title = "News"
	col.HTML = &off
	sitemapExec(t, "news.db",
		harvested("https://a.example/one", "One", "d", "2026-11-18 10:00:00", "published", "Feed A"))
	os.WriteFile(filepath.Join(htdocs, "news_items.html"), []byte("<html>stale</html>\n"), 0664)
	if err := col.Generate(os.Stderr, os.Stderr, "antenna", cfg); err != nil {
		t.Fatalf("Generate: %s", err)
	}
	for _, name := range []string{"news_items.html", "news.html"} {
		if _, err := os.Stat(filepath.Join(htdocs, name)); err == nil {
			t.Errorf("%s exists for a Markdown-only collection", name)
		}
	}
	if _, err := os.Stat(filepath.Join(htdocs, "news_items.md")); err != nil {
		t.Errorf("the items Markdown is missing: %s", err)
	}
}

const staleItemsPage = `<!doctype html>
<html lang="en-US">
<head>
  <meta charset="utf-8" >
  <meta name="generator" content="antenna/0.1.2" >
</head>
<body><main id="main-content">old items</main></body>
</html>
`

// genWithStaleHTML generates the "news" collection with an old news.html in
// htdocs and returns htdocs and what generate said.
func genWithStaleHTML(t *testing.T, stale string, pagesRow bool) (htdocs, eout, second string) {
	t.Helper()
	cfg := sitemapSite(t, "news.md", "pages.md")
	htdocs = filepath.Join(t.TempDir(), "htdocs")
	if err := os.MkdirAll(htdocs, 0775); err != nil {
		t.Fatal(err)
	}
	cfg.Htdocs = htdocs
	cfg.Collections[1].Mode = "page-index"
	col := cfg.Collections[0]
	col.Title = "News"
	sitemapExec(t, "news.db",
		harvested("https://a.example/one", "One", "d", "2026-11-18 10:00:00", "published", "Feed A"))
	if pagesRow {
		addPage(t, "pages.db", "news.html")
	}
	if err := os.WriteFile(filepath.Join(htdocs, "news.html"), []byte(stale), 0664); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := col.Generate(&buf, &buf, "antenna", cfg); err != nil {
		t.Fatalf("Generate: %s\n%s", err, buf.String())
	}
	eout = buf.String()
	// a second run
	var again bytes.Buffer
	if err := col.Generate(&again, &again, "antenna", cfg); err != nil {
		t.Fatalf("second Generate: %s\n%s", err, again.String())
	}
	return htdocs, eout, again.String()
}

// DR-0034 4: the items page 0.1.2 wrote at STEM.html is removed, with one note.
func TestTheOldItemsPageAtPlainNameIsRemovedWithANote(t *testing.T) {
	htdocs, eout, second := genWithStaleHTML(t, staleItemsPage, false)
	if _, err := os.Stat(filepath.Join(htdocs, "news.html")); err == nil {
		t.Error("the stale news.html is still there")
	}
	if _, err := os.Stat(filepath.Join(htdocs, "news_items.html")); err != nil {
		t.Errorf("news_items.html missing: %s", err)
	}
	notes := 0
	for _, l := range strings.Split(eout, "\n") {
		if strings.Contains(l, "news.html") {
			notes++
		}
	}
	if notes != 1 {
		t.Errorf("%d notes mention news.html, want one:\n%s", notes, eout)
	}
	if strings.Contains(second, "news.html") {
		t.Errorf("the second run mentions news.html again:\n%s", second)
	}
	if !strings.Contains(eout, "news_items.html") {
		t.Errorf("the note does not say where the items are now:\n%s", eout)
	}
}

// An author's own news.html, with no antenna stamp, is kept and named.
func TestAnAuthorsPlainPageIsKeptWithAWarning(t *testing.T) {
	htdocs, eout, _ := genWithStaleHTML(t, "<html><body>my own page</body></html>\n", false)
	b, err := os.ReadFile(filepath.Join(htdocs, "news.html"))
	if err != nil || !strings.Contains(string(b), "my own page") {
		t.Errorf("the author's news.html was not kept: %v", err)
	}
	if !strings.Contains(eout, "warning") || !strings.Contains(eout, "news.html") {
		t.Errorf("no warning naming news.html:\n%s", eout)
	}
}

// A news.html that `antenna page news.md` wrote carries the stamp too; it is a
// pages row, so it is the author's page and is kept silently.
func TestAPageAntennaRenderedFromNewsMdIsKept(t *testing.T) {
	htdocs, eout, _ := genWithStaleHTML(t, staleItemsPage, true)
	if _, err := os.Stat(filepath.Join(htdocs, "news.html")); err != nil {
		t.Errorf("a page recorded in the pages table was removed: %s", err)
	}
	if strings.Contains(eout, "news.html") {
		t.Errorf("a recorded page was reported:\n%s", eout)
	}
}

// Exit status (workspace DR-0003): removing a stale page is not a failure, and
// a stale page that cannot be read or removed is an I/O failure, not a crash
// (70) and not a silent 0.
func TestGenerateExitStatusWhenTheOldItemsPageCannotBeRead(t *testing.T) {
	cfg := sitemapSite(t, "news.md")
	htdocs := filepath.Join(t.TempDir(), "htdocs")
	if err := os.MkdirAll(filepath.Join(htdocs, "news.html", "sub"), 0o775); err != nil {
		t.Fatal(err) // news.html is a directory: it cannot be read as a page
	}
	cfg.Htdocs = htdocs
	cfg.Collections[0].Title = "News"
	sitemapExec(t, "news.db",
		harvested("https://a.example/one", "One", "d", "2026-11-18 10:00:00", "published", "Feed A"))
	var buf bytes.Buffer
	err := cfg.Collections[0].Generate(&buf, &buf, "antenna", cfg)
	wantClass(t, "generate with an unreadable old news.html", err, classIO)
	if _, serr := os.Stat(filepath.Join(htdocs, "news_items.html")); serr != nil {
		t.Errorf("the items page is written before the old page is dealt with: %s", serr)
	}
}

// DR-0034 4: one note for the whole run, naming every collection whose old
// items page was removed, not one line per collection.
func TestGenerateWritesOneNoteForAllRemovedItemsPages(t *testing.T) {
	newWorkspace(t, true)
	for _, n := range []string{"news", "food", "arts"} {
		if err := os.WriteFile(n+".md", []byte("# "+n+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := runIn(t, "add", n+".md"); err != nil {
			t.Fatalf("add %s: %s", n, err)
		}
	}
	cfg := &AppConfig{}
	if err := cfg.LoadConfig("antenna.yaml"); err != nil {
		t.Fatal(err)
	}
	htdocs := cfg.htdocsOrDot()
	if err := os.MkdirAll(htdocs, 0o775); err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{"news", "food"} { // arts has no old page
		if err := os.WriteFile(filepath.Join(htdocs, n+".html"), []byte(staleItemsPage), 0o664); err != nil {
			t.Fatal(err)
		}
	}
	_, eout, err := runCapture(t, "generate")
	if err != nil {
		t.Fatalf("generate: %v\n%s", err, eout)
	}
	var notes []string
	for _, l := range strings.Split(eout, "\n") {
		if strings.HasPrefix(l, "note:") {
			notes = append(notes, l)
		}
	}
	if len(notes) != 1 {
		t.Fatalf("%d notes, want one:\n%s", len(notes), eout)
	}
	for _, want := range []string{"news.html", "food.html", "news_items.html", "food_items.html"} {
		if !strings.Contains(notes[0], want) {
			t.Errorf("the note does not name %s: %s", want, notes[0])
		}
	}
	if strings.Contains(notes[0], "arts") {
		t.Errorf("the note names a collection that had no old page: %s", notes[0])
	}
	_, again, _ := runCapture(t, "generate")
	if strings.Contains(again, "note:") {
		t.Errorf("the second run notes again:\n%s", again)
	}
}
