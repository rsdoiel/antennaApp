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
	"database/sql"
	"encoding/xml"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/rsdoiel/opml"
)

const (
	// SQLTocListPages lists the pages of the pages collection with the Markdown
	// each was made from, in output path order, which is the sitemap's order.
	SQLTocListPages = `SELECT inputPath, outputPath FROM pages ORDER BY outputPath;`

	// SQLTocListPosts lists the local posts of a collection (DR-0006): items
	// with a postPath and a pubDate. Harvested items have no postPath.
	SQLTocListPosts = `SELECT ifnull(title, ''), postPath, pubDate FROM items
WHERE postPath IS NOT NULL AND postPath != '' AND pubDate IS NOT NULL AND pubDate != '';`
)

// tocFileName is where a site's table of contents goes, relative to htdocs
// (antennaApp DR-0010 5, DR-0013 1).
var tocFileName = filepath.Join(".well-known", "toc.opml")

// tocEntry is one page, collection items page or post in the table of contents.
type tocEntry struct {
	text string // the outline's text
	md   string // the published Markdown page, slash separated, relative to htdocs
	when time.Time
}

/** GenerateTOC writes the site's table of contents, <htdocs>/.well-known/toc.opml,
 * for Markdown readers such as BonBon and cmarkbrowser (antennaApp DR-0013).
 * It is one OPML outline in up to three groups, each left out when empty:
 *
 *   - Pages: the pages collection's pages in output path order, each listed
 *     only when its Markdown is published in htdocs; the title comes from the
 *     page's front matter, else its name.
 *   - Collections: each collection that has an items page (DR-0008), in
 *     antenna.yaml order. The author's feed list is never listed.
 *   - Posts: the local posts of every collection (DR-0006), listed oldest
 *     first (BonBon's arrows read the lines downward), grouped by year then month, each listed only when its Markdown is
 *     published in htdocs.
 *
 * Addresses are absolute, made from base_url, which is required: without one
 * a warning is written and no file is. With nothing to list no file is written
 * and a stale one is removed. The file is written with writeStable, so it is
 * left alone when only its dateModified would change (DR-0011).
 *
 * Parameters:
 *   eout (io.Writer) — receives the warnings, one line each
 *
 * Returns:
 *   error — nil, or the failure of the first collection database that could
 *     not be read (generate finishes everything first); a missing base_url or
 *     nothing to list is a warning, not an error
 *
 * Example:
 *   err := cfg.GenerateTOC(os.Stderr)
 */
func (cfg *AppConfig) GenerateTOC(eout io.Writer) error {
	warnf := func(format string, args ...any) {
		fmt.Fprintf(eout, "warning: toc.opml: "+format+"\n", args...)
	}
	if strings.TrimSpace(cfg.BaseURL) == "" {
		warnf("no base_url in the configuration, so no table of contents is written")
		return nil
	}
	htdocs := cfg.Htdocs
	if htdocs == "" {
		htdocs = "."
	}
	tally := &failureTally{}
	pages := cfg.tocPages(htdocs, warnf, tally)
	collections := cfg.tocCollections(htdocs)
	posts := cfg.tocPosts(htdocs, warnf, tally)

	tocPath := filepath.Join(htdocs, tocFileName)
	if len(pages)+len(collections)+len(posts) == 0 {
		if err := os.Remove(tocPath); err != nil && !os.IsNotExist(err) {
			return err
		}
		warnf("nothing to list (no published pages, collection items pages or posts), so no table of contents is written")
		return tally.err("table of contents")
	}

	base := strings.TrimRight(cfg.BaseURL, "/")
	o := opml.New()
	o.Head.Title = tocTitle(base)
	o.Head.Modified = rssDate(clock())
	o.Body.Outline = nil
	if len(pages) > 0 {
		o.Body.Outline = append(o.Body.Outline, tocGroup("Pages", base, pages))
	}
	if len(collections) > 0 {
		o.Body.Outline = append(o.Body.Outline, tocGroup("Collections", base, collections))
	}
	if len(posts) > 0 {
		o.Body.Outline = append(o.Body.Outline, tocPostsGroup(base, posts))
	}
	src, err := xml.MarshalIndent(o, "", "\t")
	if err != nil {
		return fmt.Errorf("toc.opml: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(tocPath), 0o775); err != nil {
		return asCreate(err)
	}
	if _, err := writeStable(tocPath, []byte(xml.Header+string(src)+"\n"), 0o666, tocStamps...); err != nil {
		return err
	}
	return tally.err("table of contents")
}

// tocTitle is the host and path of the site's base address, as Dave Winer's
// root table of contents names his site by its host.
func tocTitle(base string) string {
	u, err := url.Parse(base)
	if err != nil || u.Host == "" {
		return base
	}
	return u.Host + strings.TrimRight(u.Path, "/")
}

// tocAddress is base plus a slash separated path, each segment escaped.
func tocAddress(base, md string) string {
	return base + (&url.URL{Path: "/" + md}).EscapedPath()
}

func tocGroup(text, base string, entries []tocEntry) *opml.Outline {
	g := &opml.Outline{Text: text}
	for _, e := range entries {
		g.Outline = append(g.Outline, &opml.Outline{Text: e.text, URL: tocAddress(base, e.md)})
	}
	return g
}

// tocPostsGroup groups posts, held newest first and written oldest first, by
// year then month in the shape of Dave Winer's published table of contents: calendarYear and
// calendarMonth outlines with a name, and each post a leaf with text and url.
func tocPostsGroup(base string, posts []tocEntry) *opml.Outline {
	g := &opml.Outline{Text: "Posts"}
	var year, month *opml.Outline
	// Written oldest first, whatever order the caller holds them in: BonBon
	// reads the lines top to bottom for its own arrows (previous is the line
	// above), so oldest first makes them agree with ours (DR-0013 amendment).
	for i := len(posts) - 1; i >= 0; i-- {
		p := posts[i]
		if year == nil || year.Text != fmt.Sprintf("%d", p.when.Year()) {
			year = &opml.Outline{Text: fmt.Sprintf("%d", p.when.Year()), Type: "calendarYear",
				OtherAttr: opml.CustomAttrs{{Name: xml.Name{Local: "name"}, Value: fmt.Sprintf("%d", p.when.Year())}}}
			g.Outline = append(g.Outline, year)
			month = nil
		}
		label := p.when.Format("January 2006")
		if month == nil || month.Text != label {
			month = &opml.Outline{Text: label, Type: "calendarMonth",
				OtherAttr: opml.CustomAttrs{{Name: xml.Name{Local: "name"}, Value: p.when.Format("01")}}}
			year.Outline = append(year.Outline, month)
		}
		month.Outline = append(month.Outline, &opml.Outline{Text: p.text, URL: tocAddress(base, p.md)})
	}
	return g
}

// published reports whether a Markdown page exists under htdocs.
func published(htdocs, md string) bool {
	info, err := os.Stat(filepath.Join(htdocs, filepath.FromSlash(md)))
	return err == nil && !info.IsDir()
}

// markdownPathFor returns the published Markdown path for an output or post
// path: x.html and x.htm become x.md, x.md and x.markdown stay, anything else
// (an ODT post, say) has none.
func markdownPathFor(p string) string {
	p = strings.TrimPrefix(path.Clean("/"+filepath.ToSlash(p)), "/")
	switch strings.ToLower(path.Ext(p)) {
	case ".md", ".markdown":
		return p
	case ".html", ".htm":
		return strings.TrimSuffix(p, path.Ext(p)) + ".md"
	}
	return ""
}

func (cfg *AppConfig) tocPages(htdocs string, warnf func(string, ...any), tally *failureTally) []tocEntry {
	var out []tocEntry
	for _, col := range cfg.Collections {
		if !isPagesCollection(col) {
			continue
		}
		db, err := openIfExists(col.DbName)
		if err != nil {
			tally.add(err)
			continue
		}
		if db == nil {
			continue
		}
		func() {
			defer db.Close()
			var n int
			if err := db.QueryRow(SQLHasPagesTable).Scan(&n); err != nil || n == 0 {
				return
			}
			rows, err := db.Query(SQLTocListPages)
			if err != nil {
				tally.add(fmt.Errorf("listing pages in %s: %w", col.DbName, err))
				return
			}
			defer rows.Close()
			seen := map[string]bool{} // two rows can make one page (index.md and index.txt)
			for rows.Next() {
				var inputPath, outputPath string
				if err := rows.Scan(&inputPath, &outputPath); err != nil {
					tally.add(dataErrorf("reading a page row in %s: %w", col.DbName, err))
					continue
				}
				tally.add(nil)
				md := markdownPathFor(outputPath)
				if seen[outputPath] {
					continue
				}
				seen[outputPath] = true
				if md == "" || !published(htdocs, md) {
					warnf("page %s is not listed: its Markdown %s is not published in %s", outputPath, md, htdocs)
					continue
				}
				title := pageDisplayName(inputPath)
				if doc, err := LoadCommonMark(inputPath); err == nil {
					if t := strings.Join(strings.Fields(doc.GetAttributeString("title", "")), " "); t != "" {
						title = t
					}
				}
				out = append(out, tocEntry{text: title, md: md})
			}
		}()
	}
	return out
}

func (cfg *AppConfig) tocCollections(htdocs string) []tocEntry {
	var out []tocEntry
	for _, col := range cfg.Collections {
		if isPagesCollection(col) || col.Mode == "page-index" {
			continue
		}
		name := itemsMarkdownName(col.File)
		if !published(htdocs, name) {
			continue
		}
		text := strings.TrimSpace(col.Title)
		if text == "" {
			text = strings.TrimSuffix(name, "_items.md")
		}
		out = append(out, tocEntry{text: text, md: name})
	}
	return out
}

func (cfg *AppConfig) tocPosts(htdocs string, warnf func(string, ...any), tally *failureTally) []tocEntry {
	var out []tocEntry
	seen := map[string]bool{}
	for _, col := range cfg.Collections {
		db, err := openIfExists(col.DbName)
		if err != nil {
			tally.add(err)
			continue
		}
		if db == nil {
			continue
		}
		func() {
			defer db.Close()
			rows, err := db.Query(SQLTocListPosts)
			if err != nil {
				tally.add(fmt.Errorf("listing posts in %s: %w", col.DbName, err))
				return
			}
			defer rows.Close()
			for rows.Next() {
				var title, postPath, pubDate string
				if err := rows.Scan(&title, &postPath, &pubDate); err != nil {
					tally.add(dataErrorf("reading a post row in %s: %w", col.DbName, err))
					continue
				}
				tally.add(nil)
				md := markdownPathFor(postPath)
				when, dated := parseStoredDate(pubDate)
				switch {
				case md == "":
					warnf("post %s is not listed: it has no Markdown page", postPath)
					continue
				case !dated:
					warnf("post %s is not listed: its date %q cannot be read", postPath, pubDate)
					continue
				case !published(htdocs, md):
					warnf("post %s is not listed: its Markdown %s is not published in %s", postPath, md, htdocs)
					continue
				case seen[md]:
					continue
				}
				seen[md] = true
				text := strings.Join(strings.Fields(title), " ")
				if text == "" {
					text = md
				}
				out = append(out, tocEntry{text: text, md: md, when: when})
			}
		}()
	}
	// newest first, ties by path, latest first (a blog's paths carry their dates)
	sort.SliceStable(out, func(i, j int) bool {
		if !out[i].when.Equal(out[j].when) {
			return out[i].when.After(out[j].when)
		}
		return out[i].md > out[j].md
	})
	return out
}

// openIfExists opens a collection's database, or returns nil when its file
// does not exist: looking at a collection must never create its database.
func openIfExists(dbName string) (*sql.DB, error) {
	if dbName == "" {
		return nil, nil
	}
	if _, err := os.Stat(dbName); err != nil {
		return nil, nil
	}
	db, err := sql.Open("sqlite", dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database (%s): %w", dbName, err)
	}
	return db, nil
}

// tocNeighbours are the Markdown pages before and after one in the outline, by
// the rule of antennaApp DR-0013 5. "" means there is none.
type tocNeighbours struct{ prev, next string }

// tocNav maps a published Markdown page (site-relative, slash separated) to its
// neighbours in the table of contents.
type tocNav map[string]tocNeighbours

/** buildTOCNav works out the previous and next page of every page and post the
 * table of contents lists (antennaApp DR-0013 5). In the Pages group, previous
 * is the page above and next the page below. In Posts, which are listed newest
 * first, previous is the older post, which is the next line down, and next is
 * the newer one: the convention of Dave Winer's explicit urlPrev and urlNext,
 * and the reverse of the line order. The first and last entry have only one.
 * Collections have none.
 *
 * Parameters:
 *   pages ([]tocEntry) — the Pages group in outline order
 *   posts ([]tocEntry) — the posts, newest first
 *
 * Returns:
 *   tocNav — the neighbours of each listed Markdown page
 *
 * Example:
 *   nav := buildTOCNav(pages, posts)
 */
func buildTOCNav(pages, posts []tocEntry) tocNav {
	nav := tocNav{}
	for i, p := range pages {
		var n tocNeighbours
		if i > 0 {
			n.prev = pages[i-1].md
		}
		if i+1 < len(pages) {
			n.next = pages[i+1].md
		}
		nav[p.md] = n
	}
	for i, p := range posts {
		var n tocNeighbours
		if i+1 < len(posts) {
			n.prev = posts[i+1].md // older
		}
		if i > 0 {
			n.next = posts[i-1].md // newer
		}
		nav[p.md] = n
	}
	return nav
}

// computeTOCNav reads the same pages and posts GenerateTOC lists and returns
// their neighbours. It is quiet: GenerateTOC does the warning. A database that
// cannot be read gives no arrows at all, not wrong ones.
func (cfg *AppConfig) computeTOCNav() tocNav {
	htdocs := cfg.Htdocs
	if htdocs == "" {
		htdocs = "."
	}
	quiet := func(string, ...any) {}
	tally := &failureTally{}
	pages := cfg.tocPages(htdocs, quiet, tally)
	posts := cfg.tocPosts(htdocs, quiet, tally)
	if tally.err("table of contents") != nil {
		return nil
	}
	return buildTOCNav(pages, posts)
}

// attachNav gives doc the site's neighbours, for the page or post at from (its
// output path or post path), so its HTML can show previous and next when its
// front matter names none. It does nothing outside generate, which is what
// computes them.
func (cfg *AppConfig) attachNav(doc *CommonMark, from string) {
	if cfg.tocNav == nil {
		return
	}
	doc.Nav, doc.NavFrom = cfg.tocNav, markdownPathFor(from)
}

// relativeMarkdownPath writes the path of the Markdown page to, relative to
// the folder of the Markdown page from; both are site-relative and slash
// separated. A link written this way is what the .md-to-.html rewriting of
// DR-0007 expects.
func relativeMarkdownPath(from, to string) string {
	rel, err := filepath.Rel(filepath.FromSlash(path.Dir(from)), filepath.FromSlash(to))
	if err != nil {
		return to
	}
	return filepath.ToSlash(rel)
}
