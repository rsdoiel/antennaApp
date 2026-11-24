package antennaApp

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

// DR-0034 6 (plan step 4): reading the code showed that the sitemap, prune,
// the table of contents and preview do not need the items page's name. These
// tests pin that, so a later change that makes one of them name the items
// page, or lose it, is a decision and not an accident.

func siblingsSite(t *testing.T) (*AppConfig, string) {
	t.Helper()
	return itemsMDSite(t,
		harvested("https://a.example/one", "One", "d1", "2026-11-18 10:00:00", "published", "Feed A"))
}

// The sitemap lists an aggregate collection's items page, STEM_items.html,
// once it is written; the feed list STEM.md has no HTML page to list. It
// was left out of the sitemap by oversight.
func TestSitemapListsTheItemsPage(t *testing.T) {
	cfg, htdocs := siblingsSite(t)
	if _, err := os.Stat(filepath.Join(htdocs, "news_items.html")); err != nil {
		t.Fatalf("no items page: %s", err)
	}
	urls, err := collectionURLs(cfg, cfg.Collections[0])
	if err != nil {
		t.Fatal(err)
	}
	var found *URL
	for i, u := range urls {
		if strings.HasSuffix(u.Loc, "news.html") {
			t.Errorf("the sitemap lists %s, which is not written", u.Loc)
		}
		if u.Loc == "https://example.com/news_items.html" {
			found = &urls[i]
		}
	}
	if found == nil {
		t.Fatalf("news_items.html is not in the sitemap: %+v", urls)
	}
	if found.LastMod == "" {
		t.Error("the items page has no lastmod")
	}
}

// A collection with no items page written (nothing to show, or Markdown-only)
// lists nothing for it.
func TestSitemapOmitsAnItemsPageThatIsNotThere(t *testing.T) {
	cfg, htdocs := siblingsSite(t)
	os.Remove(filepath.Join(htdocs, "news_items.html"))
	urls, err := collectionURLs(cfg, cfg.Collections[0])
	if err != nil {
		t.Fatal(err)
	}
	for _, u := range urls {
		if strings.Contains(u.Loc, "news") {
			t.Errorf("the sitemap lists %s although the page is not written", u.Loc)
		}
	}
}

// changefreq and priority rules apply to the items page like any other.
func TestSitemapRulesApplyToTheItemsPage(t *testing.T) {
	cfg, _ := siblingsSite(t)
	cfg.FreqRules = map[string]string{"news_items": "hourly"}
	cfg.PriRules = map[string]string{"news_items": "0.9"}
	urls, err := collectionURLs(cfg, cfg.Collections[0])
	if err != nil {
		t.Fatal(err)
	}
	for _, u := range urls {
		if strings.HasSuffix(u.Loc, "news_items.html") && (u.ChangeFreq != "hourly" || u.Priority != "0.9") {
			t.Errorf("rules not applied: %+v", u)
		}
	}
}

// The table of contents lists the items Markdown, never an HTML page.
func TestTOCListsTheItemsMarkdownNotAnHTMLPage(t *testing.T) {
	cfg, htdocs := siblingsSite(t)
	cols := cfg.tocCollections(htdocs)
	if len(cols) != 1 || cols[0].md != "news_items.md" {
		t.Fatalf("collections = %+v, want news_items.md", cols)
	}
}

// Preview serves the files in htdocs as they are: the items page is there,
// the old plain name is not.
func TestPreviewServesTheItemsPageAtItsOwnName(t *testing.T) {
	_, htdocs := siblingsSite(t)
	srv := httptest.NewServer(previewMux(http.Dir(htdocs)))
	defer srv.Close()
	get := func(p string) (int, string) {
		res, err := http.Get(srv.URL + p)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		b, _ := io.ReadAll(res.Body)
		return res.StatusCode, string(b)
	}
	if code, body := get("/news_items.html"); code != 200 || !strings.Contains(body, "https://a.example/one") {
		t.Errorf("/news_items.html: %d\n%.200s", code, body)
	}
	if code, _ := get("/news.html"); code != 404 {
		t.Errorf("/news.html: %d, want 404", code)
	}
}

// Prune works from the pages table. A generated items page is not a pages
// row, so pruning a missing page leaves it alone.
func TestPruneLeavesTheGeneratedItemsPageAlone(t *testing.T) {
	pruneSite(t, "keep", "gone")
	if err := os.WriteFile("news_items.html", []byte("<html>items</html>\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	os.Remove("gone.md")
	if _, _, err := runCapture(t, "prune", "-w"); err != nil {
		t.Fatalf("prune -w: %v", err)
	}
	if !exists("news_items.html") {
		t.Error("prune removed the generated items page")
	}
	if exists("gone.html") {
		t.Error("prune did not remove the page whose source is gone")
	}
}

// Every page, post and items page expressed as HTML is in the sitemap: all the
// HTML outputs of a pages row (a page may write more than one), and none that
// is not HTML.
func TestSitemapListsEveryHTMLOutputOfAPage(t *testing.T) {
	cfg := sitemapSite(t, "pages.md")
	insertPageRow(t, "a.txt", `["a.md","a.html","b.html"]`)
	insertPageRow(t, "c.txt", `["c.md"]`)
	urls, err := collectionURLs(cfg, cfg.Collections[0])
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, u := range urls {
		got[strings.TrimPrefix(u.Loc, "https://example.com/")] = true
	}
	for _, want := range []string{"a.html", "b.html"} {
		if !got[want] {
			t.Errorf("%s is not in the sitemap: %v", want, got)
		}
	}
	for _, not := range []string{"a.md", "c.md", ""} {
		if got[not] {
			t.Errorf("%q is in the sitemap but is not an HTML page: %v", not, got)
		}
	}
}

// A page-index collection's page, STEM.html, is an HTML page of the site.
func TestSitemapListsAPageIndexCollectionPage(t *testing.T) {
	cfg := sitemapSite(t, "toc.md")
	htdocs := filepath.Join(t.TempDir(), "htdocs")
	if err := os.MkdirAll(htdocs, 0775); err != nil {
		t.Fatal(err)
	}
	cfg.Htdocs = htdocs
	col := cfg.Collections[0]
	col.Mode = "page-index"
	col.Title = "Contents"
	if err := col.Generate(io.Discard, io.Discard, "antenna", cfg); err != nil {
		t.Fatal(err)
	}
	urls, err := collectionURLs(cfg, col)
	if err != nil {
		t.Fatal(err)
	}
	for _, u := range urls {
		if u.Loc == "https://example.com/toc.html" {
			return
		}
	}
	t.Errorf("toc.html is not in the sitemap: %+v", urls)
}

// The usual form: `antenna page a.txt a.md a.html` renders the text source to
// Markdown and to HTML. The sitemap lists the HTML once and not the Markdown or
// the source, and a page with two HTML outputs lists both.
func TestSitemapListsThePagesTheVerbWrote(t *testing.T) {
	newWorkspace(t, true)
	if err := os.WriteFile("a.txt", []byte("A title\n\nSome text.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("b.md", []byte("---\ntitle: B\n---\n\nBody.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runIn(t, "page", "a.txt", "a.md", "a.html"); err != nil {
		t.Fatalf("page a.txt: %s", err)
	}
	if err := runIn(t, "page", "b.md", "b.html", "b2.html"); err != nil {
		t.Fatalf("page b.md: %s", err)
	}
	cfg := &AppConfig{}
	if err := cfg.LoadConfig("antenna.yaml"); err != nil {
		t.Fatal(err)
	}
	pages, err := cfg.GetCollection("pages.md")
	if err != nil {
		t.Fatal(err)
	}
	urls, err := collectionURLs(cfg, pages)
	if err != nil {
		t.Fatal(err)
	}
	count := map[string]int{}
	for _, u := range urls {
		count[path.Base(u.Loc)]++
	}
	for _, want := range []string{"a.html", "b.html", "b2.html"} {
		if count[want] != 1 {
			t.Errorf("%s listed %d times, want once: %v", want, count[want], count)
		}
	}
	for _, not := range []string{"a.txt", "a.md", "b.md", "."} {
		if count[not] != 0 {
			t.Errorf("%s is in the sitemap: %v", not, count)
		}
	}
}
