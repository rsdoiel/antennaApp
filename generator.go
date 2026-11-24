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
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	// 3rd Party Packages
	"gopkg.in/yaml.v3"
)

// Generator supports the generation of HTML pages from a YAML configuration
type Generator struct {
	// AppName holds the name of application running the generator
	AppName string `json:"appName,omitempty" yaml:"appName,omitempty"`

	// BaseURL used to form the feed Link
	BaseURL string `json:"base_url,omitempty" yaml:"base_url,omitempty"`

	// Version holds the version of the generator application
	// used when generating the "generator" metadata
	Version string `json:"version,omitempty" yaml:"version,omitempty"`

	// DbName holds the path to the SQLite3 database
	DBName string `json:"dbName,omitempty" yaml:"dbName,omitempty"`

	// DefaultAuthor is the site-wide author (antenna.yaml author). The site's
	// own posts with no author of their own get it; harvested items never do
	// (DR-0021).
	DefaultAuthor string `json:"-" yaml:"-"`

	// MarkdownAlt is the Markdown copy beside the page being written, declared
	// in its head when the page's own source is not Markdown (DR-0023).
	MarkdownAlt string `json:"-" yaml:"-"`

	// Title if this is set the title will be included
	// when generating the markdown of saved items
	Title string `json:"title,omitempty" yaml:"title,omitempty"`

	// Description, included as metadata in head element
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// CMarkFilters are Lua filters applied to the CommonMark document when
	// rendering HTML.
	CMarkFilters []string `json:"cm_filters,omitempty" yaml:"cm_filters,omitempty"`

	/*
	 * HTML page elements
	 */

	// Meta holds a list of of meta elements rendered into the head element of HTML pages
	Meta []map[string]string `json:"meta,omitempty" yaml:"meta,omitempty"`

	// Link holds the list of links elements rendered into the head element of HTML pages
	Link []map[string]string `json:"link,omitempty" yaml:"link,omitempty"`

	// Script holds a list of script elements rendered into the head element of HTML pages
	Script []map[string]string `json:"script,omitempty" yaml:"script,omitempty"`

	// Style holds an explicit Style blog that gets inserted as the last into the HTML head element
	Style string `json:"style:omitempty" yaml:"style,omitempty"`

	// Header hold the HTML markdup of the Header element. If not included
	// then it will be generated using the Title and timestamp
	Header string `json:"header,omitempty" yaml:"header,omitempty"`

	// Nav holds the HTML markup for navigation
	Nav string `json:"nav,omitempty" yaml:"nav,omitempty"`

	// TopContent holds HTML that comes before the selecton element
	// containing articles
	TopContent string `json:"top_content,omitempty" yaml:"top_content,omitempty"`

	// BottomContent holds HTML that comes before the selecton element
	// containing articles
	BottomContent string `json:"bottom_content,omitempty" yaml:"bottom_content,omitempty"`

	// Footer holds the HTML markup for the footer
	Footer string `json:"footer,omitempty" yaml:"footer,omitempty"`

	// AllowedMetaFields, when non-empty, limits which front matter keys are
	// emitted as HTML metadata for posts and pages. When empty, all keys
	// are emitted (default).
	AllowedMetaFields []string `json:"allowed_meta_fields,omitempty" yaml:"allowed_meta_fields,omitempty"`

	// Lang is the BCP 47 language tag written into <html lang="...">.
	// Defaults to "en-US" when not set via YAML.
	Lang string `json:"lang,omitempty" yaml:"lang,omitempty"`

	// Items, when set, controls how harvested feed items render in
	// aggregate collection pages. Has no effect on local post/page
	// rendering (DEC-031).
	Items ItemsConfig `json:"items,omitempty" yaml:"items,omitempty"`

	out  io.Writer
	eout io.Writer
}

// ItemsConfig controls how harvested feed items are rendered into
// aggregate collection pages (the `items:` block in page.yaml).
// See antennaApp/design_item_formatting.md and decisions.md DEC-022–031.
type ItemsConfig struct {
	// Fields is an ordered allowlist of visible body fields to render.
	// Known values: "title", "link", "pubDate", "content", "source".
	// Empty means all fields (default, applied by applyDefaults()).
	Fields []string `json:"fields,omitempty" yaml:"fields,omitempty"`

	Link LinkConfig `json:"link,omitempty" yaml:"link,omitempty"`

	// DateFormat is a Go reference-layout string applied to pubDate/updated.
	// Default: "2006-01-02".
	DateFormat string `json:"date_format,omitempty" yaml:"date_format,omitempty"`

	// ContentMaxLength truncates resolved pre-render source text on a word
	// boundary. Zero means no truncation (default).
	ContentMaxLength int `json:"content_max_length,omitempty" yaml:"content_max_length,omitempty"`

	// ShowSource controls whether the originating feed/channel label is
	// rendered. Default: true. Pointer so an explicit `show_source: false`
	// can be distinguished from "not set" (Go's bool zero value is false).
	ShowSource *bool `json:"show_source,omitempty" yaml:"show_source,omitempty"`

	// HTML is one of "strip" (default), "escape", "unsafe" — DEC-024.
	HTML string `json:"html,omitempty" yaml:"html,omitempty"`
}

// LinkConfig controls the anchor generated for each feed item.
type LinkConfig struct {
	// LabelField names the item field supplying anchor text, or the
	// literal sentinel "static". Default: "static" — a deliberate
	// accessibility default (DEC-026): screen readers read a raw URL
	// character-by-character, which is a poor listening experience. Set
	// to "link" to restore the pre-existing URL-as-anchor-text behavior,
	// or "title" to use the item title instead.
	LabelField string `json:"label_field,omitempty" yaml:"label_field,omitempty"`

	// LabelFallback is used when LabelField names a field whose value is
	// empty/missing, or unconditionally when LabelField == "static" (the
	// default). Default: "Continue reading".
	LabelFallback string `json:"label_fallback,omitempty" yaml:"label_fallback,omitempty"`

	// Required, if true, fails collection generation when an item's link
	// is empty, instead of applying Missing. Default: false.
	Required bool `json:"required,omitempty" yaml:"required,omitempty"`

	// Missing is one of "unlinked" (default), "omit", "source_link" — DEC-027.
	Missing string `json:"missing,omitempty" yaml:"missing,omitempty"`
}

// applyDefaults fills in zero-valued fields with this feature's documented
// defaults (DEC-022–031). Called once per ItemsConfig before it is used to
// render items, so a collection with no `items:` block in page.yaml still
// renders against a fully-defaulted configuration rather than a bare zero
// value.
func (cfg *ItemsConfig) applyDefaults() {
	if len(cfg.Fields) == 0 {
		cfg.Fields = []string{"title", "source", "pubDate", "content"}
	}
	if cfg.Link.LabelField == "" {
		cfg.Link.LabelField = "static" // DEC-026 — accessibility default, not "link"
	}
	if cfg.Link.LabelFallback == "" {
		cfg.Link.LabelFallback = "Continue reading"
	}
	if cfg.Link.Missing == "" {
		cfg.Link.Missing = "unlinked"
	}
	if cfg.DateFormat == "" {
		cfg.DateFormat = "2006-01-02"
	}
	if cfg.ShowSource == nil {
		t := true
		cfg.ShowSource = &t
	}
	if cfg.HTML == "" {
		cfg.HTML = "strip"
	}
}

// validate checks enum-valued fields and returns a descriptive error for
// typos, rather than letting an invalid value silently fall through to a
// default at render time. Intended to be called at config-load time.
func (cfg *ItemsConfig) validate() error {
	switch cfg.HTML {
	case "", "strip", "escape", "unsafe":
	default:
		return configErrorf("items.html: invalid value %q (want strip, escape, or unsafe)", cfg.HTML)
	}
	switch cfg.Link.Missing {
	case "", "unlinked", "omit", "source_link":
	default:
		return configErrorf("items.link.missing: invalid value %q (want unlinked, omit, or source_link)", cfg.Link.Missing)
	}
	return nil
}

// NewGenerator initialized a new Generator struct
func NewGenerator(appName string, BaseURL string) (*Generator, error) {
	gen := new(Generator)
	gen.AppName = appName
	gen.Version = Version
	gen.BaseURL = BaseURL
	gen.Lang = "en-US"
	gen.out = os.Stdout
	gen.eout = os.Stderr
	return gen, nil
}

func getDsnAndCfgName(args []string) (string, string) {
	dsn := args[0]
	if len(args) == 2 {
		return args[0], args[1]
	}
	// Figure out if we have a YAML config or not
	cfgName := strings.TrimSuffix(dsn, ".db") + ".yaml"
	if _, err := os.Stat(cfgName); err != nil {
		return dsn, ""
	}
	return dsn, cfgName
}

// LoadConfig read in the generator configuration (not AppConfig)
// and map the settings into the Generator object.
func (gen *Generator) LoadConfig(cfgName string) error {
	src, err := os.ReadFile(cfgName)
	if err != nil {
		return err
	}
	obj := new(Generator)
	if err := yaml.Unmarshal(src, &obj); err != nil {
		return configErrorf("%s: %w", cfgName, err)
	}
	// Pull in the configuration elements
	if obj.AppName != "" {
		gen.AppName = obj.AppName
	}
	if obj.BaseURL != "" {
		gen.BaseURL = obj.BaseURL
	}
	if obj.Version != "" {
		gen.Version = obj.Version
	}
	if obj.Title != "" {
		gen.Title = obj.Title
	}
	if obj.Description != "" {
		gen.Description = obj.Description
	}
	if obj.Meta != nil && len(obj.Meta) > 0 {
		gen.Meta = obj.Meta[:]
	} else {
		gen.Meta = nil
	}
	if obj.Link != nil && len(obj.Link) > 0 {
		gen.Link = append(gen.Link, obj.Link...)
	} else {
		gen.Link = nil
	}
	if obj.Script != nil && len(obj.Script) > 0 {
		gen.Script = obj.Script[:]
	} else {
		gen.Script = nil
	}
	if obj.Style != "" {
		gen.Style = obj.Style
	} else {
		gen.Style = ""
	}
	if obj.Header != "" {
		gen.Header = obj.Header
	} else {
		gen.Header = ""
	}
	if obj.Nav != "" {
		gen.Nav = obj.Nav
	} else {
		gen.Nav = ""
	}
	if obj.TopContent != "" {
		gen.TopContent = obj.TopContent
	} else {
		gen.TopContent = ""
	}
	if obj.BottomContent != "" {
		gen.BottomContent = obj.BottomContent
	} else {
		gen.BottomContent = ""
	}
	if obj.Footer != "" {
		gen.Footer = obj.Footer
	} else {
		gen.Footer = ""
	}
	if obj.Lang != "" {
		gen.Lang = obj.Lang
	}
	if obj.AllowedMetaFields != nil {
		gen.AllowedMetaFields = obj.AllowedMetaFields[:]
	}
	gen.Items = obj.Items
	return nil
}

/** Generate rebuilds the entire website from the collections defined in antenna.yaml.
 * When called with no args it processes every collection; otherwise only the named
 * collections are processed.  For each collection it regenerates: the aggregation
 * page (HTML + RSS + OPML), all individual post HTML pages, and all pages tracked
 * in the pages table.
 *
 * Parameters:
 *   out    (io.Writer) — progress messages
 *   eout   (io.Writer) — warning and error messages
 *   cfgName (string)   — path to antenna.yaml
 *   args   ([]string)  — optional list of collection filenames to restrict regeneration
 *
 * Returns:
 *   error — first fatal error encountered, or nil on success
 *
 * Example:
 *   err := app.Generate(os.Stdout, os.Stderr, "antenna.yaml", nil)
 */
func (app AntennaApp) Generate(out io.Writer, eout io.Writer, cfgName string, args []string) error {
	cfg := &AppConfig{}
	if err := cfg.LoadConfig(cfgName); err != nil {
		return err
	}
	// The table of contents lists the whole site, so it is written only when the
	// whole site is generated, not for generate COLLECTION_NAME.
	wholeSite := len(args) == 0
	if wholeSite {
		for _, col := range cfg.Collections {
			args = append(args, col.File)
		}
	}
	// What antenna renders is read once for the whole run (DR-0007).
	defer cfg.holdLinkTargets(eout)()
	// The previous and next page of every page and post, from the same data the
	// table of contents is made from, so a page or post that names neither gets
	// them (DR-0013 5). Worked out before anything is rendered.
	cfg.tocNav = cfg.computeTOCNav()
	defer func() { cfg.tocNav = nil }()
	// The old items pages removed over the whole run are reported once.
	retired := []string{}
	cfg.retiredItemsPages = &retired
	defer func() { cfg.retiredItemsPages = nil }()
	// Everything asked for is generated even when part of it fails; the
	// command then reports the failures and exits with the class of the first
	// (DR-0003 item 4), not 0.
	tally := &failureTally{}
	for _, cName := range args {
		col, err := cfg.GetCollection(cName)
		if err != nil {
			return err
		}
		if col == nil {
			fmt.Fprintf(eout, "warning could not retrieve %q, skipping\n", cName)
			continue
		}
		// Regenerate the aggregation page (HTML + RSS + OPML)
		colErr := col.Generate(out, eout, app.appName, cfg)
		if colErr != nil {
			fmt.Fprintf(eout, "warning %s: %s\n", col.File, colErr)
		}
		// Regenerate individual post HTML pages stored in this collection
		if err := col.GeneratePosts(eout, app.appName, cfg); err != nil {
			fmt.Fprintf(eout, "warning generating posts for %s: %s\n", col.File, err)
			if colErr == nil {
				colErr = err
			}
		}
		tally.add(colErr)
	}
	if len(retired) > 0 {
		fmt.Fprintln(eout, retiredItemsNote(retired))
	}
	// Regenerate all pages tracked in the pages table
	pagesErr := cfg.GeneratePages(eout)
	if pagesErr != nil {
		fmt.Fprintf(eout, "warning generating pages: %s\n", pagesErr)
	}
	// The table of contents for Markdown readers (DR-0013), after everything it
	// lists has been written.
	var tocErr error
	if wholeSite {
		if tocErr = cfg.GenerateTOC(eout); tocErr != nil {
			fmt.Fprintf(eout, "warning writing the table of contents: %s\n", tocErr)
		}
		// bonbon.json, from the bonbon block of antenna.yaml (DR-0015 5).
		if err := cfg.GenerateBonbonJSON(eout); err != nil {
			fmt.Fprintf(eout, "warning writing bonbon.json: %s\n", err)
			if tocErr == nil {
				tocErr = err
			}
		}
	}
	if err := tally.err("collections"); err != nil {
		return err
	}
	if pagesErr != nil {
		return pagesErr
	}
	return tocErr
}

/** GeneratePosts re-renders the HTML file for every post (item with postPath set)
 * in the collection, using the sourceMarkdown stored in the database.
 *
 * Parameters:
 *   eout    (io.Writer) — warning and error messages
 *   appName (string)    — name of the running application
 *   cfg     (*AppConfig) — loaded antenna.yaml configuration
 *
 * Returns:
 *   error — database query error; or nil on success; or, when some posts
 *           failed, the class of the first failure with the counts. Every
 *           per-post error is also written to eout and does not stop
 *           processing
 *
 * Example:
 *   err := col.GeneratePosts(os.Stderr, "antenna", cfg)
 */
func (collection *Collection) GeneratePosts(eout io.Writer, appName string, cfg *AppConfig) error {
	if collection.DbName == "" {
		return nil
	}
	if !collection.WantsHTML() {
		// A Markdown-only collection writes no post HTML (DR-0031 4, 5); say
		// how much from before the switch is left, and leave it.
		return collection.reportLeftoverPostHTML(eout, cfg)
	}
	db, err := sql.Open("sqlite", collection.DbName)
	if err != nil {
		return err
	}
	defer db.Close()

	gen, err := NewGenerator(appName, cfg.BaseURL)
	if err != nil {
		return err
	}
	if collection.Generator == "" {
		collection.Generator = cfg.Generator
	}
	if _, err := os.Stat(collection.Generator); err == nil {
		src, err := os.ReadFile(collection.Generator)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(src, &gen); err != nil {
			return configErrorf("%s: %w", collection.Generator, err)
		}
	} else {
		if err := yaml.Unmarshal([]byte(DefaultGeneratorYaml), &gen); err != nil {
			return err
		}
	}

	rows, err := db.Query(SQLGeneratePosts)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Every post is attempted; failures are counted and returned at the end so
	// generate does not exit 0 with posts missing (DR-0003 item 4).
	tally := &failureTally{}
	for rows.Next() {
		var (
			link           string
			postPath       string
			pubDate        string
			sourceMarkdown string
		)
		if err := rows.Scan(&link, &postPath, &pubDate, &sourceMarkdown); err != nil {
			fmt.Fprintf(eout, "warning reading post row: %s\n", err)
			tally.add(dataErrorf("reading post row: %w", err))
			continue
		}
		if sourceMarkdown == "" {
			continue
		}
		doc := &CommonMark{}
		if err := doc.Parse([]byte(sourceMarkdown)); err != nil {
			// malformed front matter — treat entire source as body text
			doc.Text = sourceMarkdown
		}
		expanded, _, ierr := expandIncludes(doc.Text, eout)
		if ierr != nil {
			fmt.Fprintf(eout, "warning rendering markdown for %q: %s\n", postPath, ierr)
			tally.add(fmt.Errorf("%s: %w", postPath, ierr))
			continue
		}
		doc.Text = expanded
		doc.Links = cfg.rewriterFor(postPath, eout)
		doc.Macros = &MacroExpander{From: postPath, Warn: eout}
		cfg.attachNav(doc, postPath)
		innerHTML, err := doc.ToUnsafeHTML()
		if err != nil {
			fmt.Fprintf(eout, "warning rendering markdown for %q: %s\n", postPath, err)
			tally.add(dataErrorf("rendering markdown for %q: %w", postPath, err))
			continue
		}
		htmlName := normalizeToHTMLExt(filepath.Join(cfg.Htdocs, postPath))
		dName := filepath.Dir(htmlName)
		if _, err := os.Stat(dName); err != nil {
			if err := os.MkdirAll(dName, 0775); err != nil {
				fmt.Fprintf(eout, "warning creating directory %q: %s\n", dName, err)
				tally.add(fmt.Errorf("creating directory %q: %w", dName, asCreate(err)))
				continue
			}
		}
		err = gen.WriteHtmlPage(htmlName, link, postPath, pubDate, innerHTML, doc.htmlFrontMatter())
		if err != nil {
			fmt.Fprintf(eout, "warning writing HTML for %q: %s\n", postPath, err)
			err = fmt.Errorf("writing HTML for %q: %w", postPath, err)
		}
		tally.add(err)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return tally.err("posts")
}

/** GeneratePages re-renders the HTML file for every page tracked in the pages table.
 * Pages are read from disk (using their inputPath) so the source Markdown must still
 * exist.  Render errors for individual pages are written to eout and do not stop
 * processing.
 *
 * Parameters:
 *   eout (io.Writer) — warning and error messages
 *
 * Returns:
 *   error — nil, or the class of the first failure with the counts; every
 *           per-page error is also written to eout
 *
 * Example:
 *   err := cfg.GeneratePages(os.Stderr)
 */
func (cfg *AppConfig) GeneratePages(eout io.Writer) error {
	pages, err := cfg.GetPageRows()
	if err != nil {
		// No pages collection is not an error during generate; a database
		// that cannot be read is.
		if class, _ := classify(err); class == classNegative {
			return nil
		}
		return err
	}
	defer cfg.holdLinkTargets(eout)()
	tally := &failureTally{}
	for _, page := range pages {
		if page.Input == "" {
			continue
		}
		err := cfg.generatePage(page, eout)
		if err != nil {
			fmt.Fprintf(eout, "warning generating page %q: %s\n", page.Input, err)
			err = fmt.Errorf("%s: %w", page.Input, err)
		}
		tally.add(err)
	}
	return tally.err("pages")
}

// generatePage re-writes every output the row records, the Markdown copies
// from the source with its includes expanded and the HTML pages rendered, and
// refreshes the row's includes (DR-0024 6).
func (cfg *AppConfig) generatePage(page pageRow, eout io.Writer) error {
	var htmls, mds []string
	for _, o := range page.Outputs {
		switch strings.ToLower(filepath.Ext(o)) {
		case ".md":
			mds = append(mds, o)
		default:
			htmls = append(htmls, o)
		}
	}
	if len(htmls) == 0 && len(mds) == 0 {
		htmls = []string{page.Output} // an old row with no recorded output: the default HTML
	}
	var includes []string
	if len(mds) > 0 {
		inc, err := cfg.writeMarkdownCopies(page.Input, mds, eout)
		if err != nil {
			return err
		}
		includes = inc
	}
	mdAlt := ""
	if len(mds) > 0 {
		mdAlt = mds[0]
	}
	recorded := append([]string{}, page.Outputs...)
	for i, h := range htmls {
		res, err := cfg.pageTo(page.Input, h, eout, mdAlt)
		if err != nil {
			return err
		}
		if len(mds) == 0 {
			includes = res.Includes
		}
		if len(recorded) == 0 && i == 0 {
			recorded = []string{res.Output}
		}
	}
	return cfg.recordPage(page.Input, recorded, includes)
}

func (collection *Collection) ApplyFilters(db *sql.DB) error {
	if len(collection.Filters) == 0 {
		return nil
	}
	for _, stmt := range collection.Filters {
		if strings.TrimSpace(stmt) != "" {
			_, err := db.Exec(stmt)
			if err != nil {
				return fmt.Errorf("%w\nstmt: %s", err, stmt)
			}
		}
	}
	return nil
}

func (collection *Collection) Generate(out io.Writer, eout io.Writer, appName string, cfg *AppConfig) error {
	gen, err := NewGenerator(appName, cfg.BaseURL)
	if err != nil {
		return err
	}
	gen.DefaultAuthor = strings.TrimSpace(cfg.Author)
	// Warnings go to the stream the caller gave, not always to os.Stderr.
	if eout != nil {
		gen.eout = eout
	}
	if collection.Generator == "" {
		collection.Generator = cfg.Generator
	}
	if _, err := os.Stat(collection.Generator); err == nil {
		src, err := os.ReadFile(collection.Generator)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(src, &gen); err != nil {
			return configErrorf("%s: %w", collection.Generator, err)
		}
	} else {
		if err := yaml.Unmarshal([]byte(DefaultGeneratorYaml), &gen); err != nil {
			return err
		}
	}
	// Use collection title as the HTML page title when the generator yaml has none
	if gen.Title == "" && collection.Title != "" {
		gen.Title = collection.Title
	}
	if collection.Link != "" {
		m := map[string]string{
			"rel":  "alternate",
			"type": "application/rss+xml",
			"href": collection.Link,
		}
		gen.Link = append(gen.Link, m)
	}
	return gen.Generate(eout, appName, cfg, collection)
}

func (gen *Generator) Generate(eout io.Writer, appName string, cfg *AppConfig, collection *Collection) error {
	// Open DB so we have a place to write data.
	dsn := collection.DbName
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	// Run the collection filter to determine which items to publish
	if err := collection.ApplyFilters(db); err != nil {
		return err
	}

	// The Markdown rendering of the items (DR-0008) is made now, before the HTML
	// head is written, because the head declares it only when there is one: a
	// collection with nothing to show has no items file and no alternate link.
	var mdBuf bytes.Buffer
	mdCount := 0
	// rowsErr is the first error that left output partial; it is returned
	// after every file that can be written has been.
	var rowsErr error
	// A Markdown-only collection always has its items file: it is the only
	// rendering of the items left (DR-0031 2).
	wantItemsMD := collection.Mode != "page-index" && (collection.WantsItemsMarkdown() || !collection.WantsHTML())
	if wantItemsMD {
		mdCount, err = gen.WriteItemsMarkdown(&mdBuf, db, collection, clock())
		if err != nil {
			if !isPartial(err) {
				return err
			}
			// An unreadable row is reported when everything else is written.
			rowsErr = err
		}
		if mdCount > 0 {
			gen.Link = append(gen.Link, map[string]string{
				"rel":  "alternate",
				"type": "text/markdown",
				"href": itemsMarkdownName(collection.File),
			})
		}
	}

	// figure out the name and path to write the HTML file to.
	bName := filepath.Base(collection.File)
	xName := filepath.Ext(bName)
	stem := strings.TrimSuffix(bName, xName)
	htmlName := filepath.Join(cfg.Htdocs, stem+".html")
	// An aggregate collection's items page is STEM_items.html, pairing with
	// STEM_items.md; the author's STEM.md has no HTML page (DR-0034 1, 2).
	// A page-index collection and the pages collection keep STEM.html.
	aggregate := collection.Mode != "page-index" && !isPagesCollection(collection)
	if aggregate {
		htmlName = filepath.Join(cfg.Htdocs, stem+"_items.html")
	}
	rssName := filepath.Join(cfg.Htdocs, strings.TrimSuffix(bName, xName)+".xml")
	opmlName := filepath.Join(cfg.Htdocs, strings.TrimSuffix(bName, xName)+".opml")

	// Every output file is rendered in memory first and written with writeStable,
	// which leaves a file alone when only its build-time line would change.

	// Write out HTML page — mode: page-index renders a simple link list from
	// the pages table; the default "aggregate" mode renders feed item cards.
	// With nothing to show an aggregate collection writes no items page either,
	// and a stale one is removed (DR-0034 5); this is known only when the
	// Markdown rendering was made.
	emptyItems := aggregate && wantItemsMD && mdCount == 0
	if collection.WantsHTML() && !emptyItems {
		var out bytes.Buffer
		if collection.Mode == "page-index" {
			// Wrap the link list in a full HTML shell using the same header/nav/footer
			// as the aggregate mode, but replace <main> content with WritePageIndex.
			fmt.Fprintf(&out, "<!doctype html>\n<html lang=%q>\n", gen.Lang)
			gen.writeHeadElement(&out, "", nil)
			fmt.Fprintln(&out, "<body>")
			fmt.Fprintln(&out, `  <a href="#main-content" class="skip-link">Skip to main content</a>`)
			if gen.Header != "" {
				fmt.Fprintf(&out, "  <header>\n    %s\n  </header>\n", indentText(strings.TrimSpace(gen.Header), 4))
			}
			if gen.Nav != "" {
				fmt.Fprintf(&out, "  <nav aria-label=\"Site navigation\">\n    %s\n  </nav>\n", indentText(strings.TrimSpace(gen.Nav), 4))
			}
			if gen.TopContent != "" {
				fmt.Fprintf(&out, "\n    %s\n", indentText(strings.TrimSpace(gen.TopContent), 2))
			}
			if err := gen.WritePageIndex(&out, db); err != nil {
				return err
			}
			if gen.BottomContent != "" {
				fmt.Fprintf(&out, "\n    %s\n", indentText(strings.TrimSpace(gen.BottomContent), 2))
			}
			if gen.Footer != "" {
				fmt.Fprintf(&out, "  <footer>\n    %s\n  </footer>\n", indentText(strings.TrimSpace(gen.Footer), 4))
			}
			fmt.Fprintln(&out, "</body>")
			fmt.Fprintln(&out, "</html>")
		} else {
			if err := gen.WriteHTML(&out, db, appName, collection); err != nil {
				if !isPartial(err) {
					return err
				}
				if rowsErr == nil {
					rowsErr = err
				}
			}
		}
		if _, err := writeStable(htmlName, out.Bytes(), 0o666, htmlStamps...); err != nil {
			return err
		}
	} else if err := os.Remove(htmlName); err != nil && !errors.Is(err, os.ErrNotExist) {
		// A Markdown-only or empty collection: a page written before is
		// removed; it would be served stale (DR-0031 5, DR-0034 5).
		return err
	}

	// The items page 0.1.2 wrote at STEM.html is gone now that STEM_items.html
	// is written (or removed): after the write, so a failed write does not
	// leave the site with neither (DR-0034 4).
	if aggregate {
		if err := cfg.retireOldItemsPage(eout, cfg.Htdocs, stem); err != nil {
			return err
		}
	}

	// The items file sits beside the HTML page. It is generated output, replaced
	// when it changes, and removed when the collection has nothing to show now.
	if wantItemsMD {
		mdName := filepath.Join(cfg.Htdocs, itemsMarkdownName(collection.File))
		if mdCount > 0 {
			if _, err := writeStable(mdName, mdBuf.Bytes(), 0o664, itemsMDStamps...); err != nil {
				return err
			}
		} else if err := os.Remove(mdName); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	} else if err := os.Remove(filepath.Join(cfg.Htdocs, itemsMarkdownName(collection.File))); err != nil && !errors.Is(err, os.ErrNotExist) {
		// A page index, or a collection with markdownItems off, writes no items
		// file; one from an earlier run would be served stale.
		return err
	}

	// The RSS feed
	var rssBuf bytes.Buffer
	if err := gen.WriteRSS(&rssBuf, db, appName, collection); err != nil {
		if !isPartial(err) {
			return err
		}
		if rowsErr == nil {
			rowsErr = err
		}
	}
	if _, err := writeStable(rssName, rssBuf.Bytes(), 0o666, rssStamps...); err != nil {
		return err
	}

	// OPML, only when there are feeds
	var opmlBuf bytes.Buffer
	if err := gen.WriteOPML(&opmlBuf, db, appName, collection); err != nil {
		return err
	}
	if opmlBuf.Len() > 0 {
		if _, err := writeStable(opmlName, opmlBuf.Bytes(), 0o666, opmlStamps...); err != nil {
			return err
		}
	}
	return rowsErr
}

/** reportLeftoverPostHTML warns once when a Markdown-only collection still has
 * post HTML files from before it was switched (DR-0031 5). It deletes nothing:
 * other pages and readers may link to them.
 *
 * Parameters:
 *   eout (io.Writer)   — where the warning goes
 *   cfg  (*AppConfig)  — the loaded configuration, for htdocs
 *
 * Returns:
 *   error — from reading the collection's database
 *
 * Example:
 *   err := col.reportLeftoverPostHTML(os.Stderr, cfg)
 */
func (collection *Collection) reportLeftoverPostHTML(eout io.Writer, cfg *AppConfig) error {
	db, err := sql.Open("sqlite", collection.DbName)
	if err != nil {
		return err
	}
	defer db.Close()
	rows, err := db.Query(`SELECT postPath FROM items WHERE postPath IS NOT NULL AND postPath != ''`)
	if err != nil {
		return err
	}
	defer rows.Close()
	var left []string
	for rows.Next() {
		var postPath string
		if err := rows.Scan(&postPath); err != nil {
			return err
		}
		name := normalizeToHTMLExt(filepath.Join(cfg.Htdocs, postPath))
		if name == filepath.Join(cfg.Htdocs, postPath) {
			continue // not a source that has an HTML twin
		}
		if _, err := os.Stat(name); err == nil {
			left = append(left, name)
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(left) > 0 {
		fmt.Fprintf(eout, "warning: %s is Markdown-only; %d post HTML file(s) from before remain, no longer updated (for example %s); delete them when nothing links to them\n",
			collection.File, len(left), left[0])
	}
	return nil
}
