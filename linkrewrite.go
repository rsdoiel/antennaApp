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
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

/** LinkTargets maps the site-relative path of a Markdown file antenna renders
 * (slash separated, for example "docs/guide.md") to the site-relative path of
 * the HTML page it renders to ("docs/guide.html"). It is what antenna knows
 * about the site from its own data (DR-0007 5, 6): the pages table, the
 * collections in antenna.yaml and the posts.
 *
 * A key that ends in "/" is a directory, not a file: it stands for a subsite
 * on this host (the subsites setting, antennaApp DR-0027 2), and any .md path
 * under it maps to the .html twin of the same name. The value is unused.
 */
type LinkTargets map[string]string

/** RewriteDestination returns the destination a link should have in HTML
 * output: a relative link to a Markdown file antenna renders becomes a link to
 * the HTML page (DR-0007). Absolute and scheme-relative URLs, links that are
 * only a fragment or query, and links to anything but .md are returned
 * unchanged. A relative .md link whose target antenna does not render is
 * returned unchanged with a warning, as is one that points outside the site.
 * The fragment and query are kept.
 *
 * Parameters:
 *   from (string) — the site-relative path of the page the link is on
 *   dest (string) — the destination as the author wrote it
 *   targets (LinkTargets) — what antenna renders
 *
 * Returns:
 *   string — the destination to use
 *   string — a warning line, or "" when there is nothing to say
 *
 * Example:
 *   dest, warn := RewriteDestination("index.md", "about.md#team", targets)
 *   // dest == "about.html#team", warn == ""
 */
func RewriteDestination(from, dest string, targets LinkTargets) (string, string) {
	if dest == "" {
		return dest, ""
	}
	u, err := url.Parse(dest)
	if err != nil || u.Scheme != "" || u.Host != "" || u.Path == "" {
		return dest, ""
	}
	ext := path.Ext(u.Path)
	if strings.ToLower(ext) != ".md" {
		return dest, ""
	}
	fromDir := path.Dir(filepath.ToSlash(from))
	root := strings.HasPrefix(u.Path, "/")
	var site string
	if root {
		site = path.Clean(strings.TrimPrefix(u.Path, "/"))
	} else {
		site = path.Join(fromDir, u.Path)
	}
	// Only the extension is case-folded; file names are exact.
	site = strings.TrimSuffix(site, ext) + ".md"
	out, ok := targets[site]
	if !ok && site != ".." && !strings.HasPrefix(site, "../") && inSubsite(targets, site) {
		out, ok = strings.TrimSuffix(site, ".md")+".html", true
	}
	if !ok || site == ".." || strings.HasPrefix(site, "../") {
		return dest, fmt.Sprintf("link to %s on %s: antenna does not render that file, so the link is left as Markdown", dest, from)
	}
	if root {
		u.Path = "/" + out
	} else {
		rel, err := filepath.Rel(filepath.FromSlash(fromDir), filepath.FromSlash(out))
		if err != nil {
			return dest, fmt.Sprintf("link to %s on %s: %s", dest, from, err)
		}
		u.Path = filepath.ToSlash(rel)
	}
	u.RawPath = ""
	return u.String(), ""
}

// inSubsite reports whether the site-relative path lies under one of the
// directory keys (those ending in "/") of targets.
func inSubsite(targets LinkTargets, site string) bool {
	for k := range targets {
		if strings.HasSuffix(k, "/") && strings.HasPrefix(site, k) {
			return true
		}
	}
	return false
}

/** LinkRewriter rewrites the links of one page while it is rendered to HTML.
 * Set it as CommonMark.Links.
 *
 * Fields:
 *   From (string) — the site-relative path of the page, so relative links
 *     are judged by where they point
 *   Targets (LinkTargets) — what antenna renders
 *   Warn (io.Writer) — receives one "warning: ..." line per link that cannot
 *     be mapped; nil discards them
 */
type LinkRewriter struct {
	From    string
	Targets LinkTargets
	Warn    io.Writer
}

// linkTransformer is the goldmark AST transformer that applies a LinkRewriter.
// It sees only link nodes, so code spans, code blocks and images are never
// touched.
type linkTransformer struct{ r *LinkRewriter }

// Transform rewrites the destination of every link in the document.
func (t linkTransformer) Transform(doc *ast.Document, _ text.Reader, _ parser.Context) {
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if link, ok := n.(*ast.Link); ok && entering {
			dest, warn := RewriteDestination(t.r.From, string(link.Destination), t.r.Targets)
			if warn != "" && t.r.Warn != nil {
				fmt.Fprintf(t.r.Warn, "warning: %s\n", warn)
			}
			link.Destination = []byte(dest)
		}
		return ast.WalkContinue, nil
	})
}

// parserOptions returns the goldmark parser options the document needs: the
// macro transformer when a MacroExpander is set, then the link transformer
// when a LinkRewriter is set (so a title link a macro makes is rewritten too).
func (doc *CommonMark) parserOptions() []parser.Option {
	var opts []parser.Option
	if opt := doc.macroOption(); opt != nil {
		opts = append(opts, opt)
	}
	if doc.Links != nil {
		opts = append(opts, parser.WithASTTransformers(util.Prioritized(linkTransformer{doc.Links}, 100)))
	}
	return opts
}

/** LinkTargets reads what antenna renders from the site's own data: the rows
 * of the pages table in the pages collection (DR-0006 says pages come only
 * from there), each collection's source file and its items page (a link to
 * either means the collection's HTML page, DR-0007 6), and every item with a
 * postPath. A collection whose database file does not exist, or has no table
 * for something, contributes nothing and is not an error.
 *
 * Returns:
 *   LinkTargets — what antenna renders
 *   error — a database that exists and cannot be read
 *
 * Example:
 *   targets, err := cfg.LinkTargets()
 */
func (cfg *AppConfig) LinkTargets() (LinkTargets, error) {
	targets := LinkTargets{}
	for _, col := range cfg.Collections {
		if col.File == "" || col.File == "pages.md" {
			continue
		}
		stem := strings.TrimSuffix(path.Clean(filepath.ToSlash(col.File)), ".md")
		targets[stem+".md"] = stem + ".html"
		targets[stem+"_items.md"] = stem + ".html"
	}
	// A subsite on this host with a path is a directory whose .md pages have
	// .html twins (DR-0027 2). One on another host, or at the root of this
	// one, is not: the root would take in the site's own unrendered pages.
	if base, err := url.Parse(cfg.BaseURL); err == nil && base.Host != "" {
		for _, sub := range cfg.Subsites {
			u, err := url.Parse(sub.root())
			if err != nil || !strings.EqualFold(u.Host, base.Host) {
				continue
			}
			if dir := strings.Trim(path.Clean("/"+u.Path), "/"); dir != "" {
				targets[dir+"/"] = ""
			}
		}
	}
	pages, err := cfg.GetPageRows()
	if err != nil {
		// No pages collection, or a pages database that does not exist yet,
		// means no pages.
		if class, _ := classify(err); class != classNegative && !strings.Contains(err.Error(), "no such table") {
			return nil, err
		}
	}
	for _, page := range pages {
		html := firstWithExt(page.Outputs, ".html")
		if html == "" {
			html = page.Output
		}
		if page.Input == "" || html == "" {
			continue
		}
		html = path.Clean(filepath.ToSlash(html))
		targets[path.Clean(filepath.ToSlash(page.Input))] = html
		// Every Markdown output of the page is the page (DR-0024). With none
		// recorded, a page made from a text source has the copy named like its
		// HTML (DR-0019).
		recorded := false
		for _, o := range page.Outputs {
			if strings.EqualFold(filepath.Ext(o), ".md") {
				targets[path.Clean(filepath.ToSlash(o))] = html
				recorded = true
			}
		}
		if !recorded && strings.HasSuffix(html, ".html") {
			copyMD := strings.TrimSuffix(html, ".html") + ".md"
			if _, taken := targets[copyMD]; !taken {
				targets[copyMD] = html
			}
		}
	}
	for _, col := range cfg.Collections {
		if col.DbName == "" {
			continue
		}
		if _, err := os.Stat(col.DbName); err != nil {
			continue
		}
		if err := addPostTargets(targets, col.DbName); err != nil {
			return nil, err
		}
	}
	return targets, nil
}

// addPostTargets adds every post (an item with a postPath) of one database.
func addPostTargets(targets LinkTargets, dbName string) error {
	db, err := sql.Open("sqlite", dbName)
	if err != nil {
		return err
	}
	defer db.Close()
	rows, err := db.Query(`SELECT postPath FROM items WHERE postPath IS NOT NULL AND postPath != ''`)
	if err != nil {
		if strings.Contains(err.Error(), "no such table") || strings.Contains(err.Error(), "no such column") {
			return nil
		}
		return fmt.Errorf("%s: %w", dbName, err)
	}
	defer rows.Close()
	for rows.Next() {
		var postPath string
		if err := rows.Scan(&postPath); err != nil {
			return fmt.Errorf("%s: %w", dbName, err)
		}
		p := path.Clean(filepath.ToSlash(postPath))
		targets[p] = filepath.ToSlash(normalizeToHTMLExt(p))
	}
	return rows.Err()
}

/** rewriterFor returns the LinkRewriter for the page at from (site-relative),
 * using the targets cached for the current generate run, or read now when
 * there is no run. When what antenna renders cannot be read it says so on warn
 * and returns nil, so links are left as written and the page is still
 * generated.
 */
func (cfg *AppConfig) rewriterFor(from string, warn io.Writer) *LinkRewriter {
	targets := cfg.linkTargets
	if targets == nil {
		var err error
		if targets, err = cfg.LinkTargets(); err != nil {
			fmt.Fprintf(warn, "warning: %s: could not read what antenna renders, so .md links are left as written: %s\n", from, err)
			return nil
		}
	}
	return &LinkRewriter{From: from, Targets: targets, Warn: warn}
}

// holdLinkTargets reads what antenna renders once for a generate run and
// returns the function that ends the run's cache. A read failure is reported
// on warn and leaves the cache empty, so each page reports it again and keeps
// its links as written.
func (cfg *AppConfig) holdLinkTargets(warn io.Writer) func() {
	if cfg.linkTargets != nil {
		return func() {}
	}
	targets, err := cfg.LinkTargets()
	if err != nil {
		fmt.Fprintf(warn, "warning: could not read what antenna renders, so .md links are left as written: %s\n", err)
		return func() {}
	}
	cfg.linkTargets = targets
	return func() { cfg.linkTargets = nil }
}
