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
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const discoverRSS = `<?xml version="1.0"?><rss version="2.0"><channel><title>Found</title><link>https://example.org/</link>
<item><title>One</title><link>https://example.org/one</link><guid>one</guid></item></channel></rss>`

// route is one path a test site answers.
type route struct {
	contentType string
	body        string
}

// fakeSite serves routes, answers everything else 404 and records the hits.
func fakeSite(routes map[string]route) (*httptest.Server, *hitLog) {
	hits := &hitLog{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.add(r.URL.Path)
		rt, ok := routes[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", rt.contentType)
		w.Write([]byte(strings.ReplaceAll(rt.body, "{{SITE}}", "http://"+r.Host)))
	}))
	return srv, hits
}

func rssRoute() route { return route{"application/rss+xml", discoverRSS} }

func tocWith(body string) route {
	return route{"text/xml", `<?xml version="1.0" encoding="UTF-8"?><opml version="2.0"><head><title>t</title></head><body>` + body + `</body></opml>`}
}

func TestResolveFeed_ListedAddressIsAFeed(t *testing.T) {
	srv, hits := fakeSite(map[string]route{"/rss": rssRoute()})
	defer srv.Close()
	f, _ := testFetcher(0)
	got, err := resolveFeed(f, srv.URL+"/rss", "")
	if err != nil {
		t.Fatal(err)
	}
	if got.address != srv.URL+"/rss" || got.feed.Title != "Found" {
		t.Errorf("address %q title %q", got.address, got.feed.Title)
	}
	if hits.count("/rss") != 1 {
		t.Errorf("the feed was requested %d times, want 1", hits.count("/rss"))
	}
}

func TestResolveFeed_TocWithRSSOutlineGivesItsFeed(t *testing.T) {
	srv, hits := fakeSite(map[string]route{
		"/.well-known/toc.opml": tocWith(`<outline text="Posts" url="{{SITE}}/p/one.md"/><outline text="Feed" type="rss" xmlUrl="{{SITE}}/rss.xml"/>`),
		"/rss.xml":              rssRoute(),
		"/p/one.md":             {"text/markdown", "# one"},
	})
	defer srv.Close()
	f, _ := testFetcher(0)
	got, err := resolveFeed(f, srv.URL+"/.well-known/toc.opml", "")
	if err != nil {
		t.Fatal(err)
	}
	if got.address != srv.URL+"/rss.xml" {
		t.Errorf("address %q, want the rss outline's xmlUrl", got.address)
	}
	if hits.count("/p/one.md") != 0 {
		t.Error("a page the table of contents lists was requested")
	}
}

func TestResolveFeed_TocWithOnlyPagesIsNoFeedAndFetchesNoPage(t *testing.T) {
	srv, hits := fakeSite(map[string]route{
		"/.well-known/toc.opml": tocWith(`<outline text="Posts"><outline text="One" url="{{SITE}}/p/one.md"/><outline text="Two" url="{{SITE}}/p/two.md"/></outline><outline type="include" text="Sub" url="{{SITE}}/sub/.well-known/toc.opml"/>`),
		"/p/one.md":             {"text/markdown", "# one"},
		"/p/two.md":             {"text/markdown", "# two"},
	})
	defer srv.Close()
	f, _ := testFetcher(0)
	_, err := resolveFeed(f, srv.URL+"/.well-known/toc.opml", "")
	if err == nil || !strings.Contains(err.Error(), "no feed") {
		t.Fatalf("err = %v, want a no-feed error", err)
	}
	wantClass(t, "table of contents with no feed", err, classData)
	if !strings.Contains(err.Error(), srv.URL+"/.well-known/toc.opml") {
		t.Errorf("the error does not name the link: %v", err)
	}
	for _, p := range []string{"/p/one.md", "/p/two.md", "/sub/.well-known/toc.opml"} {
		if hits.count(p) != 0 {
			t.Errorf("%s was requested; harvest never reads what a table of contents lists", p)
		}
	}
}

func TestResolveFeed_FeedListOPMLIsADataError(t *testing.T) {
	srv, _ := fakeSite(map[string]route{
		"/list.opml": tocWith(`<outline text="A" type="rss" xmlUrl="https://a.example/feed" url="https://a.example/feed"/><outline text="B" type="rss" xmlUrl="https://b.example/feed" url="https://b.example/feed"/>`),
	})
	defer srv.Close()
	f, _ := testFetcher(0)
	_, err := resolveFeed(f, srv.URL+"/list.opml", "")
	if err == nil || !strings.Contains(err.Error(), "feed list") {
		t.Fatalf("err = %v, want it to say this is a feed list", err)
	}
	wantClass(t, "feed list", err, classData)
}

func TestResolveFeed_HTMLAutodiscovery(t *testing.T) {
	page := `<!doctype html><html><head><title>Blog</title>
<link rel="stylesheet" href="/s.css">
<link rel="alternate" type="text/html" href="/de/">
<link rel="alternate" type="application/atom+xml" title="Atom" href="/atom-feed">
<link rel="alternate" type="application/rss+xml" href="/rss-feed">
</head><body></body></html>`
	srv, hits := fakeSite(map[string]route{
		"/":          {"text/html; charset=utf-8", page},
		"/atom-feed": rssRoute(),
		"/rss-feed":  rssRoute(),
	})
	defer srv.Close()
	f, _ := testFetcher(0)
	got, err := resolveFeed(f, srv.URL+"/", "")
	if err != nil {
		t.Fatal(err)
	}
	if got.address != srv.URL+"/atom-feed" {
		t.Errorf("address %q, want the first feed link, resolved against the page", got.address)
	}
	if hits.count("/feed") != 0 || hits.count("/feed/") != 0 {
		t.Error("probed paths although the page named its feed")
	}
}

func TestResolveFeed_ProbesCommonPathsInOrderAndStops(t *testing.T) {
	srv, hits := fakeSite(map[string]route{
		"/":        {"text/html", "<html><head></head><body>no link here</body></html>"},
		"/feed/":   rssRoute(),
		"/rss.xml": rssRoute(),
	})
	defer srv.Close()
	f, _ := testFetcher(0)
	got, err := resolveFeed(f, srv.URL+"/", "")
	if err != nil {
		t.Fatal(err)
	}
	if got.address != srv.URL+"/feed/" {
		t.Errorf("address %q, want /feed/ (WordPress, Substack)", got.address)
	}
	if hits.count("/rss.xml") != 0 {
		t.Error("kept probing after a hit")
	}
	if hits.count("/feed") != 1 {
		t.Errorf("/feed probed %d times, want once, before /feed/", hits.count("/feed"))
	}
}

func TestResolveFeed_AtMostEightProbes(t *testing.T) {
	srv, hits := fakeSite(map[string]route{"/": {"text/html", "<html><body>x</body></html>"}})
	defer srv.Close()
	f, _ := testFetcher(0)
	_, err := resolveFeed(f, srv.URL+"/", "")
	if err == nil || !strings.Contains(err.Error(), "no feed") {
		t.Fatalf("err = %v, want a no-feed error", err)
	}
	n := 0
	for _, p := range hits.paths {
		if p != "/" && p != "/robots.txt" {
			n++
		}
	}
	if n > 8 {
		t.Errorf("%d probes, want at most 8: %v", n, hits.paths)
	}
	if n < 8 {
		t.Errorf("%d probes, want all 8 tried: %v", n, hits.paths)
	}
}

// A Markdown site's root is not HTML and has no autodiscovery tag; its table
// of contents may name the feed.
func TestResolveFeed_MarkdownRootTriesTheOriginsToc(t *testing.T) {
	srv, _ := fakeSite(map[string]route{
		"/":                     {"text/markdown; charset=utf-8", "# Hello\n\nA Markdown page."},
		"/.well-known/toc.opml": tocWith(`<outline text="Feed" type="rss" xmlUrl="{{SITE}}/scripting.xml"/>`),
		"/scripting.xml":        rssRoute(),
	})
	defer srv.Close()
	f, _ := testFetcher(0)
	got, err := resolveFeed(f, srv.URL+"/", "")
	if err != nil {
		t.Fatal(err)
	}
	if got.address != srv.URL+"/scripting.xml" {
		t.Errorf("address %q", got.address)
	}
}

func TestResolveFeed_MarkdownSiteWithNoFeedIsAnErrorThatNamesTheLink(t *testing.T) {
	srv, hits := fakeSite(map[string]route{
		"/":                     {"text/markdown", "# Hello"},
		"/.well-known/toc.opml": tocWith(`<outline text="One" url="{{SITE}}/one.md"/>`),
		"/one.md":               {"text/markdown", "# one"},
	})
	defer srv.Close()
	f, _ := testFetcher(0)
	_, err := resolveFeed(f, srv.URL+"/", "")
	if err == nil || !strings.Contains(err.Error(), "no feed") || !strings.Contains(err.Error(), srv.URL+"/") {
		t.Fatalf("err = %v", err)
	}
	wantClass(t, "no feed", err, classData)
	if hits.count("/one.md") != 0 {
		t.Error("a page from the table of contents was requested")
	}
}

func TestResolveFeed_RememberedAddressIsTriedFirst(t *testing.T) {
	srv, hits := fakeSite(map[string]route{
		"/":         {"text/html", "<html><body>x</body></html>"},
		"/feed.xml": rssRoute(),
	})
	defer srv.Close()
	f, _ := testFetcher(0)
	got, err := resolveFeed(f, srv.URL+"/", srv.URL+"/feed.xml")
	if err != nil {
		t.Fatal(err)
	}
	if got.address != srv.URL+"/feed.xml" {
		t.Errorf("address %q", got.address)
	}
	if hits.count("/") != 0 || hits.count("/feed") != 0 {
		t.Errorf("went looking although the remembered feed answered: %v", hits.paths)
	}
}

func TestResolveFeed_RememberedAddressThatFailsIsDiscoveredAgain(t *testing.T) {
	srv, _ := fakeSite(map[string]route{"/": {"text/html", "<html><body>x</body></html>"}, "/rss": rssRoute()})
	defer srv.Close()
	f, _ := testFetcher(0)
	got, err := resolveFeed(f, srv.URL+"/", srv.URL+"/gone.xml")
	if err != nil {
		t.Fatal(err)
	}
	if got.address != srv.URL+"/rss" {
		t.Errorf("address %q, want a fresh discovery", got.address)
	}
}

// A listed address that answers 404 stays the unavailable failure it was;
// discovery does not turn a dead link into a hunt.
func TestResolveFeed_ListedAddressThat404sIsUnavailable(t *testing.T) {
	srv, hits := fakeSite(map[string]route{})
	defer srv.Close()
	f, _ := testFetcher(0)
	_, err := resolveFeed(f, srv.URL+"/dead", "")
	if err == nil {
		t.Fatal("a 404 was accepted")
	}
	wantClass(t, "404", err, classUnavailable)
	for _, p := range hits.paths {
		if strings.HasPrefix(p, "/feed") || p == "/rss" {
			t.Errorf("probed %s after a 404 on the listed address", p)
		}
	}
	_ = fmt.Sprint
}
