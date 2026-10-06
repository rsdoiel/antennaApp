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
	"path/filepath"
	"strings"
	"testing"
)

// touch makes a file with some content, creating directories as needed.
func touch(t *testing.T, name string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(name, []byte("old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func exists(name string) bool {
	_, err := os.Lstat(name)
	return err == nil
}

// A site that shrank has sitemap files from the old run beyond the new count;
// the index does not name them. With the clean option they are removed.
func TestSitemapClean_RemovesNumberedFilesThisRunDidNotWrite(t *testing.T) {
	cfg := sitemapSite(t, "pages.md")
	addPage(t, "pages.db", "about.html")
	for _, n := range []string{"sitemap_2.xml", "sitemap_3.xml", "sitemap_17.xml"} {
		touch(t, n)
	}
	var warn bytes.Buffer
	if err := generateSitemaps(cfg, &warn, true); err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{"sitemap_2.xml", "sitemap_3.xml", "sitemap_17.xml"} {
		if exists(n) {
			t.Errorf("%s is stale and should have been removed", n)
		}
		if !strings.Contains(warn.String(), "removed stale "+n) {
			t.Errorf("expected a report that %s was removed, got:\n%s", n, warn.String())
		}
	}
	if !exists("sitemap_1.xml") || !exists("sitemap_index.xml") {
		t.Errorf("the files this run wrote must stay")
	}
	locs, files := sitemapLocs(t, ".")
	wantLocs(t, locs, "https://example.com/about.html")
	if len(files) != 1 {
		t.Errorf("expected just sitemap_1.xml, got %v", files)
	}
}

// Nothing is deleted unless asked.
func TestSitemapClean_OffByDefaultLeavesStaleFiles(t *testing.T) {
	cfg := sitemapSite(t, "pages.md")
	addPage(t, "pages.db", "about.html")
	touch(t, "sitemap_2.xml")
	if err := generateSitemaps(cfg, &bytes.Buffer{}, false); err != nil {
		t.Fatal(err)
	}
	if !exists("sitemap_2.xml") {
		t.Errorf("sitemap_2.xml was removed without the clean option")
	}
}

// Only regular files named exactly sitemap_<number>.xml are stale. Everything
// else in the document root, including look-alikes, is left alone.
func TestSitemapClean_TouchesNothingButNumberedSitemapFiles(t *testing.T) {
	cfg := sitemapSite(t, "pages.md")
	addPage(t, "pages.db", "about.html")
	keep := []string{
		"sitemap.xml", "sitemap_notes.txt", "sitemap_2.xml.bak", "sitemap_2.xml.gz",
		"sitemap_index.xml.bak", "my_sitemap_2.xml", "sitemap_a.xml", "sitemap_.xml",
		"other_2.xml", "index.html", "sub/sitemap_2.xml",
	}
	for _, n := range keep {
		touch(t, n)
	}
	// A directory that matches the pattern is not a sitemap file.
	if err := os.Mkdir("sitemap_9.xml", 0o755); err != nil {
		t.Fatal(err)
	}
	touch(t, "sitemap_9.xml/inside.txt")
	if err := generateSitemaps(cfg, &bytes.Buffer{}, true); err != nil {
		t.Fatal(err)
	}
	for _, n := range append(keep, "sitemap_9.xml/inside.txt") {
		if !exists(n) {
			t.Errorf("%s is not a stale sitemap file and must not be removed", n)
		}
	}
}

// Cleanup works in the document root, like the files it cleans up after.
func TestSitemapClean_WorksInHtdocs(t *testing.T) {
	cfg := sitemapSite(t, "pages.md")
	cfg.Htdocs = "public"
	addPage(t, "pages.db", "about.html")
	touch(t, "public/sitemap_4.xml")
	touch(t, "sitemap_4.xml") // in the working directory: not the document root
	if err := generateSitemaps(cfg, &bytes.Buffer{}, true); err != nil {
		t.Fatal(err)
	}
	if exists("public/sitemap_4.xml") {
		t.Errorf("public/sitemap_4.xml is stale and should be gone")
	}
	if !exists("sitemap_4.xml") {
		t.Errorf("a file outside htdocs must not be touched")
	}
}

// With nothing to list there is no current sitemap at all, so every old file
// is stale, the index included: nothing points at it any more.
func TestSitemapClean_NothingToListRemovesTheOldIndexToo(t *testing.T) {
	cfg := sitemapSite(t, "pages.md", "news.md")
	addHarvested(t, "news.db", 3)
	for _, n := range []string{"sitemap_1.xml", "sitemap_2.xml", "sitemap_index.xml"} {
		touch(t, n)
	}
	var warn bytes.Buffer
	if err := generateSitemaps(cfg, &warn, true); err != nil {
		t.Fatalf("nothing to list is not an error: %s", err)
	}
	for _, n := range []string{"sitemap_1.xml", "sitemap_2.xml", "sitemap_index.xml"} {
		if exists(n) {
			t.Errorf("%s describes a site that has nothing to list and should be gone", n)
		}
	}
	if !strings.Contains(warn.String(), "no pages or posts") {
		t.Errorf("expected the nothing-to-list warning, got:\n%s", warn.String())
	}
}

// And without the option, the old index and files stay.
func TestSitemapClean_NothingToListWithoutCleanKeepsOldFiles(t *testing.T) {
	cfg := sitemapSite(t, "pages.md")
	touch(t, "sitemap_1.xml")
	touch(t, "sitemap_index.xml")
	if err := generateSitemaps(cfg, &bytes.Buffer{}, false); err != nil {
		t.Fatal(err)
	}
	if !exists("sitemap_1.xml") || !exists("sitemap_index.xml") {
		t.Errorf("old files were removed without the clean option")
	}
}

// If a collection could not be read the map is incomplete, so nothing is
// deleted: files that look stale may be the only copy of what that collection
// contributed.
func TestSitemapClean_IsSkippedWhenACollectionFailed(t *testing.T) {
	cfg := sitemapSite(t, "pages.md")
	addPage(t, "pages.db", "about.html")
	cfg.Collections = append(cfg.Collections, &Collection{File: "gone.md", DbName: "gone.db"})
	touch(t, "sitemap_5.xml")
	var warn bytes.Buffer
	err := generateSitemaps(cfg, &warn, true)
	wantClass(t, "sitemap -clean with a missing database", err, classNoInput)
	if !exists("sitemap_5.xml") {
		t.Errorf("cleanup ran although a collection failed")
	}
	if !strings.Contains(warn.String(), "cleanup skipped") {
		t.Errorf("expected a note that cleanup was skipped, got:\n%s", warn.String())
	}
}

// The option on the command line: -clean (or --clean), no other argument.
func TestRun_SitemapCleanOption(t *testing.T) {
	newWorkspace(t, true)
	touch(t, "sitemap_9.xml")
	if err := runIn(t, "sitemap", "-clean"); err != nil {
		t.Fatalf("sitemap -clean: %s", err)
	}
	if exists("sitemap_9.xml") {
		t.Errorf("sitemap -clean left a stale file")
	}
	touch(t, "sitemap_9.xml")
	if err := runIn(t, "sitemap", "--clean"); err != nil {
		t.Fatalf("sitemap --clean: %s", err)
	}
	if exists("sitemap_9.xml") {
		t.Errorf("sitemap --clean left a stale file")
	}
	touch(t, "sitemap_9.xml")
	if err := runIn(t, "sitemap"); err != nil {
		t.Fatalf("sitemap: %s", err)
	}
	if !exists("sitemap_9.xml") {
		t.Errorf("sitemap without -clean removed a file")
	}
}

func TestRun_SitemapBadArgumentsAreUsage(t *testing.T) {
	newWorkspace(t, false) // no workspace: the command line is judged first
	for _, args := range [][]string{{"-bogus"}, {"extra"}, {"-clean", "extra"}, {"-clean", "-clean", "-clean"}} {
		wantClass(t, "sitemap "+strings.Join(args, " "), runIn(t, "sitemap", args...), classUsage)
	}
}
