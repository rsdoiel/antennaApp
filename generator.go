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

	out  io.Writer
	eout io.Writer
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
		return err
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
	if len(args) == 0 {
		for _, col := range cfg.Collections {
			args = append(args, col.File)
		}
	}
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
		if err := col.Generate(out, eout, app.appName, cfg); err != nil {
			fmt.Fprintf(eout, "warning %s: %s\n", col.File, err)
		}
		// Regenerate individual post HTML pages stored in this collection
		if err := col.GeneratePosts(eout, app.appName, cfg); err != nil {
			fmt.Fprintf(eout, "warning generating posts for %s: %s\n", col.File, err)
		}
	}
	// Regenerate all pages tracked in the pages table
	if err := cfg.GeneratePages(eout); err != nil {
		fmt.Fprintf(eout, "warning generating pages: %s\n", err)
	}
	return nil
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
 *   error — database query error, or nil on success; per-post render errors are
 *           written to eout and do not stop processing
 *
 * Example:
 *   err := col.GeneratePosts(os.Stderr, "antenna", cfg)
 */
func (collection *Collection) GeneratePosts(eout io.Writer, appName string, cfg *AppConfig) error {
	if collection.DbName == "" {
		return nil
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
			return err
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

	for rows.Next() {
		var (
			link           string
			postPath       string
			pubDate        string
			sourceMarkdown string
		)
		if err := rows.Scan(&link, &postPath, &pubDate, &sourceMarkdown); err != nil {
			fmt.Fprintf(eout, "warning reading post row: %s\n", err)
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
		if strings.Contains(doc.Text, "@include-text-block") {
			doc.Text = IncludeTextBlock(doc.Text)
		}
		if strings.Contains(doc.Text, "@include-code-block") {
			doc.Text = IncludeCodeBlock(doc.Text)
		}
		innerHTML, err := doc.ToUnsafeHTML()
		if err != nil {
			fmt.Fprintf(eout, "warning rendering markdown for %q: %s\n", postPath, err)
			continue
		}
		htmlName := normalizeToHTMLExt(filepath.Join(cfg.Htdocs, postPath))
		dName := filepath.Dir(htmlName)
		if _, err := os.Stat(dName); err != nil {
			if err := os.MkdirAll(dName, 0775); err != nil {
				fmt.Fprintf(eout, "warning creating directory %q: %s\n", dName, err)
				continue
			}
		}
		if err := gen.WriteHtmlPage(htmlName, link, postPath, pubDate, innerHTML, doc.FrontMatter); err != nil {
			fmt.Fprintf(eout, "warning writing HTML for %q: %s\n", postPath, err)
		}
	}
	return rows.Err()
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
 *   error — always nil; per-page errors are written to eout
 *
 * Example:
 *   err := cfg.GeneratePages(os.Stderr)
 */
func (cfg *AppConfig) GeneratePages(eout io.Writer) error {
	pages, err := cfg.GetPages()
	if err != nil {
		// No pages collection or no pages found — not an error during generate
		return nil
	}
	for _, page := range pages {
		inputPath := page["inputPath"]
		outputPath := page["outputPath"]
		if inputPath == "" {
			continue
		}
		if err := cfg.Page(inputPath, outputPath); err != nil {
			fmt.Fprintf(eout, "warning generating page %q: %s\n", inputPath, err)
		}
	}
	return nil
}

func (collection *Collection) ApplyFilters(db *sql.DB) error {
	if len(collection.Filters) == 0 {
		return nil
	}
	for _, stmt := range collection.Filters {
		if strings.TrimSpace(stmt) != "" {
			_, err := db.Exec(stmt)
			if err != nil {
				return fmt.Errorf("%s\nstmt: %s", err, stmt)
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
	if collection.Generator == "" {
		collection.Generator = cfg.Generator
	}
	if _, err := os.Stat(collection.Generator); err == nil {
		src, err := os.ReadFile(collection.Generator)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(src, &gen); err != nil {
			return err
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

	// figure out the name and path to write the HTML file to.
	bName := filepath.Base(collection.File)
	xName := filepath.Ext(bName)
	htmlName := filepath.Join(cfg.Htdocs, strings.TrimSuffix(bName, xName)+".html")
	rssName := filepath.Join(cfg.Htdocs, strings.TrimSuffix(bName, xName)+".xml")
	opmlName := filepath.Join(cfg.Htdocs, strings.TrimSuffix(bName, xName)+".opml")

	// clear existing page
	if _, err := os.Stat(htmlName); err == nil {
		if err := os.Remove(htmlName); err != nil {
			return nil
		}
	}

	// Create the HTML file
	out, err := os.Create(htmlName)
	if err != nil {
		return err
	}

	// Write out HTML page
	if err := gen.WriteHTML(out, db, appName, collection); err != nil {
		return err
	}
	out.Close()

	// clear existing page
	if _, err := os.Stat(rssName); err == nil {
		if err := os.Remove(rssName); err != nil {
			return nil
		}
	}

	// Create the RSS file
	out, err = os.Create(rssName)
	if err != nil {
		return err
	}

	// Write out RSS page
	if err := gen.WriteRSS(out, db, appName, collection); err != nil {
		return err
	}
	out.Close()

	// Write OPML to a buffer first — only create the file if there are feeds
	var opmlBuf bytes.Buffer
	if err := gen.WriteOPML(&opmlBuf, db, appName, collection); err != nil {
		return err
	}
	if opmlBuf.Len() > 0 {
		if _, err := os.Stat(opmlName); err == nil {
			if err := os.Remove(opmlName); err != nil {
				return err
			}
		}
		out, err = os.Create(opmlName)
		if err != nil {
			return err
		}
		if _, err := opmlBuf.WriteTo(out); err != nil {
			out.Close()
			return err
		}
		out.Close()
	}
	return nil
}
