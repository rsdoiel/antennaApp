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
	"database/sql"
	"encoding/xml"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// sitemapSite builds a temporary site: a working directory holding one
// database per collection, each made by the same code antenna add uses, so
// the tables are the real schema. Returns the config.
func sitemapSite(t *testing.T, collections ...string) *AppConfig {
	t.Helper()
	dir := t.TempDir()
	t.Chdir(dir)
	cfg := &AppConfig{BaseURL: "https://example.com", ChunkSize: 100}
	for _, file := range collections {
		db := strings.TrimSuffix(file, ".md") + ".db"
		if err := setupDatabase(file, db); err != nil {
			t.Fatalf("setupDatabase(%s): %s", file, err)
		}
		cfg.Collections = append(cfg.Collections, &Collection{File: file, DbName: db})
	}
	return cfg
}

func sitemapExec(t *testing.T, dbName string, stmts ...string) {
	t.Helper()
	db, err := sql.Open("sqlite", dbName)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			t.Fatalf("%s: %s\n%s", dbName, err, s)
		}
	}
}

// addPage, addPost and addHarvested put rows in a collection's database the
// way the actions do: a page is a pages row; a post is an item with a
// postPath; a harvested item is an item with a pubDate and no postPath.
func addPage(t *testing.T, db, out string) {
	t.Helper()
	sitemapExec(t, db, `INSERT INTO pages (inputPath, outputPath, updated)
		VALUES ('`+strings.TrimSuffix(out, ".html")+`.md', '`+out+`', '2026-10-01 12:00:00')`)
}

func addPost(t *testing.T, db, postPath string) {
	t.Helper()
	sitemapExec(t, db, `INSERT INTO items (link, postPath, title, description, pubDate, status, updated)
		VALUES ('https://example.com/`+postPath+`', '`+postPath+`', 'T', 'D', '2026-10-01 12:00:00', 'published', '2026-10-02 12:00:00')`)
}

func addHarvested(t *testing.T, db string, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		sitemapExec(t, db, `INSERT INTO items (link, title, description, pubDate, status, updated)
			VALUES ('https://elsewhere.example/`+string(rune('a'+i))+`', 'T', 'D', '2026-10-01 12:00:00', 'published', '2026-10-01 12:00:00')`)
	}
}

// sitemapLocs returns every <loc> in every sitemap_N.xml in dir, sorted, and
// the files they came from.
func sitemapLocs(t *testing.T, dir string) (locs []string, files []string) {
	t.Helper()
	files, _ = filepath.Glob(filepath.Join(dir, "sitemap_[0-9]*.xml"))
	sort.Strings(files)
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		var set URLSet
		if err := xml.Unmarshal(src, &set); err != nil {
			t.Fatalf("%s is not valid sitemap XML: %s", f, err)
		}
		for _, u := range set.URLs {
			locs = append(locs, u.Loc)
		}
	}
	sort.Strings(locs)
	return locs, files
}

func indexLocs(t *testing.T, dir string) []string {
	t.Helper()
	src, err := os.ReadFile(filepath.Join(dir, "sitemap_index.xml"))
	if err != nil {
		t.Fatalf("no sitemap_index.xml: %s", err)
	}
	var idx SitemapIndex
	if err := xml.Unmarshal(src, &idx); err != nil {
		t.Fatalf("sitemap_index.xml is not valid: %s", err)
	}
	var out []string
	for _, s := range idx.Sitemaps {
		out = append(out, s.Loc)
	}
	sort.Strings(out)
	return out
}

func wantLocs(t *testing.T, got []string, want ...string) {
	t.Helper()
	sort.Strings(want)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("sitemap URLs differ.\n got:\n  %s\nwant:\n  %s", strings.Join(got, "\n  "), strings.Join(want, "\n  "))
	}
}

// The site's own pages come from the pages collection; local posts count in
// any collection; harvested feed items are not pages of this site.
func TestSitemap_ListsPagesAndLocalPostsNotHarvestedItems(t *testing.T) {
	cfg := sitemapSite(t, "pages.md", "news.md")
	addPage(t, "pages.db", "about.html")
	addPage(t, "pages.db", "contact.html")
	addPost(t, "pages.db", "blog/2026/10/01/first.md")
	addHarvested(t, "news.db", 5)
	addPost(t, "news.db", "blog/2026/10/02/second.md")
	if err := generateSitemaps(cfg, &bytes.Buffer{}, false); err != nil {
		t.Fatalf("generateSitemaps: %s", err)
	}
	locs, _ := sitemapLocs(t, ".")
	wantLocs(t, locs,
		"https://example.com/about.html",
		"https://example.com/contact.html",
		"https://example.com/blog/2026/10/01/first.html",
		"https://example.com/blog/2026/10/02/second.html")
}

// Harvested items have no postPath. They used to come out as the site root,
// once each: 61,663 times for the antenna site.
func TestSitemap_NeverListsTheRootForHarvestedItems(t *testing.T) {
	cfg := sitemapSite(t, "pages.md")
	addPage(t, "pages.db", "about.html")
	addHarvested(t, "pages.db", 4)
	if err := generateSitemaps(cfg, &bytes.Buffer{}, false); err != nil {
		t.Fatal(err)
	}
	locs, _ := sitemapLocs(t, ".")
	wantLocs(t, locs, "https://example.com/about.html")
}

// Pages belong to the pages collection. A pages row in any other collection's
// database is not a page of this site.
func TestSitemap_PagesOnlyFromThePagesCollection(t *testing.T) {
	cfg := sitemapSite(t, "pages.md", "news.md")
	addPage(t, "pages.db", "about.html")
	addPage(t, "news.db", "stray.html")
	if err := generateSitemaps(cfg, &bytes.Buffer{}, false); err != nil {
		t.Fatal(err)
	}
	locs, _ := sitemapLocs(t, ".")
	wantLocs(t, locs, "https://example.com/about.html")
}

// An older database has no pages table. That is "no pages", not a failure,
// and the posts in it still count (the pages query used to fail first and
// return before the posts were read).
func TestSitemap_CollectionWithoutPagesTableStillListsItsPosts(t *testing.T) {
	cfg := sitemapSite(t, "pages.md", "old.md")
	addPage(t, "pages.db", "about.html")
	addPost(t, "old.db", "blog/2025/01/01/older.md")
	sitemapExec(t, "old.db", `DROP TABLE pages`)
	var warn bytes.Buffer
	if err := generateSitemaps(cfg, &warn, false); err != nil {
		t.Fatalf("a missing pages table must not fail the command: %s", err)
	}
	locs, _ := sitemapLocs(t, ".")
	wantLocs(t, locs,
		"https://example.com/about.html",
		"https://example.com/blog/2025/01/01/older.html")
	if strings.Contains(warn.String(), "no such table") {
		t.Errorf("a missing pages table is not worth a warning, got: %s", warn.String())
	}
}

// Nothing to point at: warn, write nothing, succeed. An aggregation-only site
// has no sitemap and that is not a failure (and an empty sitemap is not valid
// per the sitemaps.org schema).
func TestSitemap_NothingToMapWarnsAndWritesNothing(t *testing.T) {
	cfg := sitemapSite(t, "pages.md", "news.md")
	addHarvested(t, "news.db", 3)
	var warn bytes.Buffer
	if err := generateSitemaps(cfg, &warn, false); err != nil {
		t.Fatalf("nothing to map must not be an error, got: %s", err)
	}
	if !strings.Contains(warn.String(), "no pages or posts") {
		t.Errorf("expected a warning that there are no pages or posts, got: %q", warn.String())
	}
	matches, _ := filepath.Glob("sitemap*.xml")
	if len(matches) != 0 {
		t.Errorf("expected no sitemap files, found %v", matches)
	}
}

// Chunk files are numbered across the whole run, not per collection, so one
// collection's files do not overwrite another's; the index names each once.
func TestSitemap_ChunkFilesAreUniqueAndIndexedOnce(t *testing.T) {
	cfg := sitemapSite(t, "pages.md", "news.md", "notes.md")
	cfg.ChunkSize = 2
	addPage(t, "pages.db", "a.html")
	addPage(t, "pages.db", "b.html")
	addPage(t, "pages.db", "c.html")
	addPost(t, "news.db", "blog/n1.md")
	addPost(t, "news.db", "blog/n2.md")
	addPost(t, "notes.db", "blog/o1.md")
	if err := generateSitemaps(cfg, &bytes.Buffer{}, false); err != nil {
		t.Fatal(err)
	}
	locs, files := sitemapLocs(t, ".")
	if len(locs) != 6 {
		t.Errorf("expected 6 URLs across the run, got %d: %v", len(locs), locs)
	}
	if len(files) != 3 {
		t.Errorf("expected 3 chunk files of at most 2 URLs, got %v", files)
	}
	idx := indexLocs(t, ".")
	want := []string{
		"https://example.com/sitemap_1.xml",
		"https://example.com/sitemap_2.xml",
		"https://example.com/sitemap_3.xml",
	}
	wantLocs(t, idx, want...)
}

// The index points at BaseURL/sitemap_N.xml, so the files belong in the
// document root, not the directory the command happened to run from.
func TestSitemap_FilesAreWrittenToHtdocs(t *testing.T) {
	cfg := sitemapSite(t, "pages.md")
	cfg.Htdocs = "public"
	if err := os.Mkdir("public", 0o755); err != nil {
		t.Fatal(err)
	}
	addPage(t, "pages.db", "about.html")
	if err := generateSitemaps(cfg, &bytes.Buffer{}, false); err != nil {
		t.Fatal(err)
	}
	locs, _ := sitemapLocs(t, "public")
	wantLocs(t, locs, "https://example.com/about.html")
	if _, err := os.Stat("sitemap_index.xml"); err == nil {
		t.Errorf("sitemap_index.xml was written to the working directory, not htdocs")
	}
	indexLocs(t, "public")
}

// A collection whose database file is missing is a failure to report, and
// opening it must not create an empty database as a side effect. The other
// collections are still mapped.
func TestSitemap_MissingDatabaseIsReportedNotCreated(t *testing.T) {
	cfg := sitemapSite(t, "pages.md")
	addPage(t, "pages.db", "about.html")
	cfg.Collections = append(cfg.Collections, &Collection{File: "gone.md", DbName: "gone.db"})
	err := generateSitemaps(cfg, &bytes.Buffer{}, false)
	wantClass(t, "sitemap with a missing database", err, classNoInput)
	if _, statErr := os.Stat("gone.db"); statErr == nil {
		t.Errorf("gone.db was created as a side effect of opening it")
	}
	locs, _ := sitemapLocs(t, ".")
	wantLocs(t, locs, "https://example.com/about.html")
}

// A post with no updated date is still a post: it is listed without lastmod.
func TestSitemap_PostWithoutUpdatedDateIsListed(t *testing.T) {
	cfg := sitemapSite(t, "pages.md")
	sitemapExec(t, "pages.db", `INSERT INTO items (link, postPath, title, description, pubDate, status)
		VALUES ('https://example.com/blog/x.md', 'blog/x.md', 'T', 'D', '2026-10-01 12:00:00', 'published')`)
	if err := generateSitemaps(cfg, &bytes.Buffer{}, false); err != nil {
		t.Fatalf("generateSitemaps: %s", err)
	}
	locs, _ := sitemapLocs(t, ".")
	wantLocs(t, locs, "https://example.com/blog/x.html")
}
