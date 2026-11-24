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
	"database/sql"
	"os"
	"strings"
	"testing"
)

// DR-0030: harvest finds a site's feed from the address the author lists,
// stores the answer in the channel row, and respects robots.txt. These tests
// run the real harvest verb against a test server and read the real SQLite
// file afterwards.

const selfLinkRSS = `<?xml version="1.0"?><rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom"><channel><title>Found</title><link>https://example.org/</link>
<atom:link href="https://example.org/declared-self.xml" rel="self" type="application/rss+xml"/>
<item><title>One</title><link>https://example.org/one</link><guid>one</guid><description>First</description></item></channel></rss>`

// harvestSite makes a workspace whose news.md lists links, with the pause
// turned off in antenna.yaml so no test sleeps.
func harvestSite(t *testing.T, links ...string) {
	t.Helper()
	newWorkspace(t, true)
	cfg, err := os.ReadFile("antenna.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("antenna.yaml", append(cfg, []byte("\nharvest:\n  pause: 0s\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	md := "# News\n\n"
	for _, l := range links {
		md += "- " + l + "\n"
	}
	if err := os.WriteFile("news.md", []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runIn(t, "add", "news.md"); err != nil {
		t.Fatalf("add: %s", err)
	}
}

type channelRow struct{ link, feedLink, discovered string }

func channelRows(t *testing.T) []channelRow {
	t.Helper()
	db, err := sql.Open("sqlite", "news.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows, err := db.Query(`SELECT link, COALESCE(feed_link, ''), COALESCE(discovered_link, '') FROM channels ORDER BY link`)
	if err != nil {
		t.Fatalf("reading channels: %s", err)
	}
	defer rows.Close()
	var out []channelRow
	for rows.Next() {
		var r channelRow
		if err := rows.Scan(&r.link, &r.feedLink, &r.discovered); err != nil {
			t.Fatal(err)
		}
		out = append(out, r)
	}
	return out
}

func itemCount(t *testing.T) int {
	t.Helper()
	db, err := sql.Open("sqlite", "news.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM items`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestHarvest_SiteAddressFindsTheFeedAndRemembersIt(t *testing.T) {
	srv, hits := fakeSite(map[string]route{
		"/":             {"text/html", `<html><head><link rel="alternate" type="application/rss+xml" href="/the-feed.xml"></head><body>hi</body></html>`},
		"/the-feed.xml": {"application/rss+xml", selfLinkRSS},
	})
	defer srv.Close()
	harvestSite(t, "[Site]("+srv.URL+"/)")
	if _, _, err := runCapture(t, "harvest", "news.md"); err != nil {
		t.Fatalf("harvest: %s", err)
	}
	rows := channelRows(t)
	if len(rows) != 1 {
		t.Fatalf("channels: %+v", rows)
	}
	if rows[0].link != srv.URL+"/" {
		t.Errorf("link %q, want the address the author listed", rows[0].link)
	}
	if rows[0].discovered != srv.URL+"/the-feed.xml" {
		t.Errorf("discovered_link %q, want the feed address", rows[0].discovered)
	}
	if rows[0].feedLink != "https://example.org/declared-self.xml" {
		t.Errorf("feed_link %q: gofeed's own FeedLink was overwritten", rows[0].feedLink)
	}
	if itemCount(t) != 1 {
		t.Errorf("items = %d, want 1", itemCount(t))
	}
	// The second run goes straight to the remembered feed: the page is not
	// asked for again and the answer is kept.
	if _, _, err := runCapture(t, "harvest", "news.md"); err != nil {
		t.Fatalf("second harvest: %s", err)
	}
	if hits.count("/") != 1 {
		t.Errorf("the site address was requested %d times over two runs, want 1", hits.count("/"))
	}
	if hits.count("/the-feed.xml") != 2 {
		t.Errorf("the feed was requested %d times, want 2", hits.count("/the-feed.xml"))
	}
	if rows := channelRows(t); rows[0].discovered != srv.URL+"/the-feed.xml" {
		t.Errorf("discovered_link lost on the second run: %+v", rows)
	}
}

func TestHarvest_ListedFeedHasNoDiscoveredLink(t *testing.T) {
	srv, _ := fakeSite(map[string]route{"/rss": {"application/rss+xml", selfLinkRSS}})
	defer srv.Close()
	harvestSite(t, "[Feed]("+srv.URL+"/rss)")
	if _, _, err := runCapture(t, "harvest", "news.md"); err != nil {
		t.Fatal(err)
	}
	if rows := channelRows(t); len(rows) != 1 || rows[0].discovered != "" {
		t.Errorf("channels: %+v, want an empty discovered_link for a listed feed", rows)
	}
}

// A collection database made before this release has no discovered_link
// column. Harvest adds it and keeps what is there.
func TestHarvest_AddsTheColumnToAnOldDatabase(t *testing.T) {
	srv, _ := fakeSite(map[string]route{"/rss": {"application/rss+xml", selfLinkRSS}})
	defer srv.Close()
	harvestSite(t, "[Feed]("+srv.URL+"/rss)")
	db, err := sql.Open("sqlite", "news.db")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`ALTER TABLE channels DROP COLUMN discovered_link`); err != nil {
		t.Fatalf("making an old database: %s", err)
	}
	if _, err := db.Exec(`INSERT INTO channels (link, title) VALUES ('https://old.example/feed', 'Old')`); err != nil {
		t.Fatal(err)
	}
	db.Close()
	if _, _, err := runCapture(t, "harvest", "news.md"); err != nil {
		t.Fatalf("harvest on an old database: %s", err)
	}
	rows := channelRows(t)
	if len(rows) != 2 {
		t.Fatalf("channels: %+v, want the old row kept and the new one added", rows)
	}
}

func TestHarvest_SiteWithNoFeedFailsThatLinkOnly(t *testing.T) {
	srv, hits := fakeSite(map[string]route{
		"/md/":                     {"text/markdown", "# a markdown site"},
		"/md/.well-known/toc.opml": tocWith(`<outline text="One" url="{{SITE}}/md/one.md"/>`),
		"/md/one.md":               {"text/markdown", "# one"},
		"/good":                    {"application/rss+xml", selfLinkRSS},
	})
	defer srv.Close()
	harvestSite(t, "[Markdown site]("+srv.URL+"/md/)", "[Good]("+srv.URL+"/good)")
	_, eout, err := runCapture(t, "harvest", "news.md")
	wantClass(t, "a site with no feed", err, classData)
	if !strings.Contains(eout, srv.URL+"/md/") || !strings.Contains(eout, "no feed") {
		t.Errorf("the warning does not name the link and say no feed:\n%s", eout)
	}
	if itemCount(t) != 1 {
		t.Errorf("items = %d: the feed after the failing link was not harvested", itemCount(t))
	}
	if hits.count("/md/one.md") != 0 {
		t.Error("a page listed in a table of contents was requested")
	}
}

func TestHarvest_RobotsDisallowedFeedIsSkippedNotFailed(t *testing.T) {
	srv, hits := fakeSite(map[string]route{
		"/robots.txt":  {"text/plain", "User-agent: *\nDisallow: /private\n"},
		"/private/rss": {"application/rss+xml", selfLinkRSS},
		"/good":        {"application/rss+xml", selfLinkRSS},
	})
	defer srv.Close()
	harvestSite(t, "[Private]("+srv.URL+"/private/rss)", "[Good]("+srv.URL+"/good)")
	_, eout, err := runCapture(t, "harvest", "news.md")
	if err != nil {
		t.Fatalf("a robots.txt skip made harvest fail: %v", err)
	}
	if hits.count("/private/rss") != 0 {
		t.Error("a disallowed feed was requested")
	}
	if !strings.Contains(eout, "robots.txt") {
		t.Errorf("the skip was not reported:\n%s", eout)
	}
	if itemCount(t) != 1 {
		t.Errorf("items = %d, want the allowed feed harvested", itemCount(t))
	}
}

// DR-0030 9: a bad harvest setting is a configuration error.
func TestHarvest_BadPauseIsAConfigurationError(t *testing.T) {
	for _, bad := range []string{"soon", "-5s"} {
		newWorkspace(t, true)
		cfg, _ := os.ReadFile("antenna.yaml")
		os.WriteFile("antenna.yaml", append(cfg, []byte("\nharvest:\n  pause: "+bad+"\n")...), 0o644)
		_, _, err := runCapture(t, "harvest")
		wantClass(t, "harvest pause "+bad, err, classConfig)
	}
}
