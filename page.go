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
	"encoding/json"
	"gopkg.in/yaml.v3"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var (
	DefaultPageCollectionMarkdown = `---
title: An Antenna Website
description: This is the default websites created by the antenna init action.
---

# Welcome to your Antenna

`
	DefaultGeneratorYaml = `##
# Example YAML HTML page generator.
##
meta:
  - http-equiv: "Content-Type"
    content: "text/html; charset=utf-8"
  - name: language
    content: "en-US"
  - name: viewport
    content: "width=device-width, initial-scale=1.0"
#link:
#  - rel: stylesheet
#    type: text/css
#    href: /css/site.css
#script:
#  - type: module
#    src: /modules/date_filter.js
#style: |
#  /* CSS can go here */

##
#  Visible HTML page elements
##
header: |
  <!-- your custom header element's inner HTML goes here -->

nav: |
  <!-- your custom nav element's inner HTML goes here -->

top_content: |
  <!-- your custom HTML content before the section's element goes here -->

bottom_content: |
  <!-- your custom HTML content after the section's element goes here -->

footer: |
  <!-- your custom footer element's inner HTML goes here -->
`
)

// InitPageGenerator initializes a YAML configuration for the default page layout.
// It takes a filename and returns an error
func InitPageGenerator(pageName string) error {
	if _, err := os.Stat(pageName); err == nil {
		//FIXME: read in pageName and Make sure it has a valid structure
		return nil
	} else {
		// NOTE: Create a default page pagefooter Generator
		if err := os.WriteFile(pageName, []byte(DefaultGeneratorYaml), 0664); err != nil {
			return fmt.Errorf("failed to create %q, %w", pageName, asCreate(err))
		}
	}
	return nil
}

// Page will add a CommonMark document as an HTML page based on postPath. It **will not** get added
// to an RSS feed or to the collection's db. It is designed to relieve the requirement of using Pandoc
// for a handful of page.
func (app *AntennaApp) Page(cfgName string, args []string) error {
	if len(args) < 1 {
		return usageErrorf("expected filepath for Markdown content")
	}
	cfg := &AppConfig{}
	if err := cfg.LoadConfig(cfgName); err != nil {
		return err
	}
	fName := strings.TrimSpace(args[0])
	outputs := make([]string, 0, len(args)-1)
	for _, o := range args[1:] {
		outputs = append(outputs, strings.TrimSpace(o))
	}
	return cfg.PageOutputs(fName, outputs, os.Stderr)
}

/** PageOutputs renders one source file to each output named, the extension of
 * an output saying what it is (DR-0019): .html is the HTML page, .md is the
 * Markdown with the include directives expanded.
 *
 * Parameters:
 *   fName (string) — the source file (Markdown, a text file with directives, or ODT)
 *   outputs ([]string) — output paths under htdocs; none means the one HTML page as before
 *   eout (io.Writer) — where warnings about links go
 *
 * Returns:
 *   error — usage (an output that is neither .html nor .md, a Markdown output
 *   for an ODT source), data (an output that is the source itself), or from reading or writing
 *
 * Example:
 *   err := cfg.PageOutputs("index.txt", []string{"index.md", "index.html"}, os.Stderr)
 */
func (cfg *AppConfig) PageOutputs(fName string, outputs []string, eout io.Writer) error {
	var htmls, mds []string
	src := filepath.Clean(fName)
	for _, o := range outputs {
		if o == "" {
			continue
		}
		if filepath.Clean(o) == src || filepath.Clean(filepath.Join(cfg.Htdocs, o)) == src {
			return dataErrorf("output %q is the source file; antenna never overwrites a source", o)
		}
	}
	for _, o := range outputs {
		switch strings.ToLower(filepath.Ext(o)) {
		case ".html":
			htmls = append(htmls, o)
		case ".md":
			mds = append(mds, o)
		default:
			return usageErrorf("output %q must end in .html or .md", o)
		}
	}
	if len(mds) > 0 && isODTFile(fName) {
		return usageErrorf("a Markdown output needs a text source, not %q", fName)
	}
	if len(outputs) == 0 {
		// One HTML page, as before. A source that is not .md, .odt or .ott
		// (index.txt) gets a name derived from its own, never its own.
		name := ""
		switch ext := filepath.Ext(fName); strings.ToLower(ext) {
		case ".md", ".odt", ".ott":
		default:
			name = strings.TrimSuffix(fName, ext) + ".html"
		}
		htmls = append(htmls, name)
	}
	var includes []string
	if len(mds) > 0 {
		inc, err := cfg.writeMarkdownCopies(fName, mds, eout)
		if err != nil {
			return err
		}
		includes = inc
	}
	mdAlt := ""
	if len(mds) > 0 {
		mdAlt = mds[0]
	}
	recorded := append([]string{}, outputs...)
	for i, h := range htmls {
		res, err := cfg.pageTo(fName, h, eout, mdAlt)
		if err != nil {
			return err
		}
		if len(mds) == 0 {
			includes = res.Includes
		}
		if len(outputs) == 0 && i == 0 {
			recorded = []string{res.Output} // the name pageTo chose
		}
	}
	return cfg.recordPage(fName, recorded, includes)
}

/** writeMarkdownCopies writes the Markdown copy of a page to each path in mds
 * (under htdocs): the source with its include directives expanded and the
 * derived front matter. An output that would not change is left alone.
 *
 * Parameters:
 *   fName (string) — the source file
 *   mds ([]string) — the .md output paths, under htdocs
 *   eout (io.Writer) — warnings
 *
 * Returns:
 *   []string — every file the source includes, nested ones too
 *   error — from reading the source, expanding it (a data error for a cycle or
 *     nesting deeper than five levels) or writing a copy
 *
 * Example:
 *   includes, err := cfg.writeMarkdownCopies("index.txt", []string{"index.md"}, os.Stderr)
 */
func (cfg *AppConfig) writeMarkdownCopies(fName string, mds []string, eout io.Writer) ([]string, error) {
	raw, err := os.ReadFile(fName)
	if err != nil {
		return nil, fmt.Errorf("failed to read %q: %w", fName, err)
	}
	expanded, includes, err := expandIncludes(string(raw), eout)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fName, err)
	}
	text := cfg.deriveFrontMatter(expanded)
	for _, o := range mds {
		mdName := filepath.Join(cfg.Htdocs, o)
		if err := os.MkdirAll(filepath.Dir(mdName), 0775); err != nil {
			return nil, asCreate(err)
		}
		if _, err := writeStable(mdName, []byte(text), 0664); err != nil {
			return nil, fmt.Errorf("failed to write %q, %w", mdName, asCreate(err))
		}
	}
	return includes, nil
}

// expandIncludes expands the include directives in text, the same expansion
// the HTML page and the Markdown copy get (DR-0024 12). It returns the text,
// every file the directives reached, and a data error for a cycle or for
// nesting deeper than five levels. What it could not read is reported on eout.
func expandIncludes(text string, eout io.Writer) (string, []string, error) {
	return expandIncludesReport(text, eout)
}

// Unpage will remove a CommonMark document filePath
func (app *AntennaApp) Unpage(cfgName string, args []string) error {
	if len(args) < 1 {
		return usageErrorf("expected filepath for Markdown content")
	}
	cfg := &AppConfig{}
	if err := cfg.LoadConfig(cfgName); err != nil {
		return err
	}
	fName := strings.TrimSpace(args[0])
	return cfg.Unpage(fName)
}

// Pages will list the pages in the pages collection
func (app *AntennaApp) Pages(out io.Writer, cfgName string, args []string) error {
	cfg := &AppConfig{}
	if err := cfg.LoadConfig(cfgName); err != nil {
		return err
	}
	onlyStale := false
	for _, a := range args {
		switch a {
		case "-stale", "--stale":
			onlyStale = true
		default:
			return usageErrorf("unexpected argument %q; try 'antenna help pages'", a)
		}
	}
	return cfg.Pages(out, onlyStale)
}


// Unpage delete a page from the pages table (does not remove files from disk)
// The deletion will happen for pages with either an inputPath or outputPath matching
// the fName parameter.
func (cfg *AppConfig) Unpage(fName string) error {
	// NOTE: remove the page from pages table.
	collection, err := cfg.GetCollection("pages.md")
	if err != nil {
		return err
	}
	dsn := collection.DbName
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("failed to open %s, %w", dsn, err)
	}
	defer db.Close()

	res, err := db.Exec(SQLDeletePageByPath, fName, fName)
	if err != nil {
		return fmt.Errorf("%s, %w", dsn, err)
	}
	// No matching record is a negative answer (exit 1), not a success.
	if n, err := res.RowsAffected(); err == nil && n == 0 {
		return negativef("no page record matches %q", fName)
	}
	return nil
}

func (cfg *AppConfig) Page(fName string, oName string) error {
	var outputs []string
	if oName != "" {
		outputs = []string{oName}
	}
	return cfg.PageOutputs(fName, outputs, os.Stderr)
}

// pageResult is what rendering one page's HTML found out: the output name used
// (relative to htdocs) and every file the source includes.
type pageResult struct {
	Output   string
	Includes []string
}

// pageTo renders one page's HTML; warnings about links go to eout. mdAlt is the
// site-relative path of the page's Markdown copy, or "" when none is recorded,
// in which case the copy beside the HTML is declared when it exists (DR-0023).
// The pages row is written by the caller, not here.
func (cfg *AppConfig) pageTo(fName string, oName string, eout io.Writer, mdAlt string) (pageResult, error) {
	doc, err := LoadCommonMark(fName)
	if err != nil {
		return pageResult{}, err
	}
	// NOTE: This is trusted content so I can support commonMarkDoc
	// processor extensions safely.
	expanded, includes, err := expandIncludes(doc.Text, eout)
	if err != nil {
		return pageResult{}, fmt.Errorf("%s: %w", fName, err)
	}
	doc.Text = expanded

	// Convert our document text to HTML
	// NOTE: Pages are allowed to have "unsafe" embedded HTML because they are
	// not reading from a feed, they are being read from your file system.
	from := oName
	if from == "" {
		from = doc.GetAttributeString("postPath", fName)
	}
	doc.Links = cfg.rewriterFor(from, eout)
	doc.Macros = &MacroExpander{From: from, Warn: eout}
	cfg.attachNav(doc, from)
	innerHTML, err := doc.ToUnsafeHTML()
	if err != nil {
		return pageResult{}, err
	}
	postPath := doc.GetAttributeString("postPath", fName)
	htmlName := filepath.Join(cfg.Htdocs, postPath)
	if oName != "" {
		htmlName = filepath.Join(cfg.Htdocs, oName)
	} else if postPath != "" {
		oName = normalizeToHTMLExt(postPath)
	} else {
		oName = normalizeToHTMLExt(fName)
	}

	// Normalize the HTML filename, replacing any source document extension with .html
	htmlName = normalizeToHTMLExt(htmlName)
	dName := filepath.Dir(htmlName)
	if _, err := os.Stat(dName); err != nil {
		if err := os.MkdirAll(dName, 0775); err != nil {
			return pageResult{}, err
		}
	}
	gen, err := NewGenerator(path.Base(os.Args[0]), cfg.BaseURL)
	if err != nil {
		return pageResult{}, err
	}
	if err := gen.LoadConfig(cfg.Generator); err != nil {
		return pageResult{}, err
	}
	// The Markdown copy the row records is declared, as a path relative to the
	// HTML; with none recorded, a page made from a text source declares the
	// copy beside its HTML when it exists, and a page whose source is .md
	// declares the source (DR-0023, DR-0024).
	if mdAlt != "" {
		if rel, err := filepath.Rel(filepath.Dir(htmlName), filepath.Join(cfg.Htdocs, mdAlt)); err == nil {
			gen.MarkdownAlt = filepath.ToSlash(rel)
		}
	} else if !strings.HasSuffix(postPath, ".md") {
		copyMD := strings.TrimSuffix(htmlName, filepath.Ext(htmlName)) + ".md"
		if info, err := os.Stat(copyMD); err == nil && !info.IsDir() {
			gen.MarkdownAlt = filepath.Base(copyMD)
		}
	}
	if err := gen.WriteHtmlPage(htmlName, "", postPath, "", innerHTML, doc.htmlFrontMatter()); err != nil {
		return pageResult{}, err
	}
	return pageResult{Output: oName, Includes: includes}, nil
}

// GetPages returns a list of page maps for a collection: inputPath,
// outputPath (deprecated), updated, and outputs and includes as the compact
// JSON arrays (empty when none). A table made before DR-0024 is read as it is.
func (cfg *AppConfig) GetPages() ([]map[string]string, error) {
	rows, err := cfg.GetPageRows()
	if err != nil {
		return nil, err
	}
	pages := []map[string]string{}
	for _, r := range rows {
		outJSON, _ := json.Marshal(r.Outputs)
		incJSON := ""
		if len(r.Includes) > 0 {
			b, _ := json.Marshal(r.Includes)
			incJSON = string(b)
		}
		if len(r.Outputs) == 0 {
			outJSON = nil
		}
		pages = append(pages, map[string]string{
			"inputPath":  r.Input,
			"outputPath": r.Output,
			"updated":    r.Updated,
			"outputs":    string(outJSON),
			"includes":   incJSON,
		})
	}
	// No pages is an empty list, not an error (DR-0003).
	return pages, nil
}

// GetPageRows returns the rows of the pages table, by inputPath. A missing
// pages collection or database is the error GetCollection or sql gives.
func (cfg *AppConfig) GetPageRows() ([]pageRow, error) {
	db, dsn, err := cfg.openPagesDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	rows, err := queryPageRows(db, "inputPath, updated DESC")
	if err != nil {
		if class, _ := classify(err); class == classData {
			return nil, err
		}
		return nil, fmt.Errorf("%s\n%s, %w", SQLListItems, dsn, err)
	}
	return rows, nil
}

/** Pages writes the pages table to out, one page a line, newest first (DR-0024
 * 7, 13): input path, output path (deprecated), updated, outputs, includes and
 * stale, separated by tabs. outputs and includes are the JSON arrays as
 * stored, or empty. stale is empty for a current page, "stale" when an output
 * is older than its source or an include, "missing" when an output is absent.
 * With -stale only the pages whose stale is not empty are written, and the
 * result is a negative-class error (exit 1) when there are any.
 *
 * Parameters:
 *   out (io.Writer) — the listing
 *   stale (bool) — list only stale or missing pages
 *
 * Returns:
 *   error — nil, a negative error for -stale with something to report, or from
 *   the database (a data error for a bad JSON column)
 *
 * Example:
 *   err := cfg.Pages(os.Stdout, false)
 */
func (cfg *AppConfig) Pages(out io.Writer, onlyStale bool) error {
	db, dsn, err := cfg.openPagesDB()
	if err != nil {
		return err
	}
	defer db.Close()
	rows, err := queryPageRows(db, "updated DESC")
	if err != nil {
		if class, _ := classify(err); class == classData {
			return err
		}
		return fmt.Errorf("%s, %w", dsn, err)
	}
	stale := 0
	for _, r := range rows {
		state := pageStale(r, cfg.Htdocs)
		if state != "" {
			stale++
		} else if onlyStale {
			continue
		}
		outJSON, incJSON := "", ""
		if len(r.Outputs) > 0 {
			b, _ := json.Marshal(r.Outputs)
			outJSON = string(b)
		}
		if len(r.Includes) > 0 {
			b, _ := json.Marshal(r.Includes)
			incJSON = string(b)
		}
		fmt.Fprintf(out, "%s\t%s\t%s\t%s\t%s\t%s\n", r.Input, r.Output, r.Updated, outJSON, incJSON, state)
	}
	if onlyStale && stale > 0 {
		return negativef("%d pages are stale or have a missing output", stale)
	}
	return nil
}

/** deriveFrontMatter adds what antenna can derive to the front matter of a
 * generated Markdown copy of a page (DR-0019 6, DR-0021): a title from the
 * first level-1 heading when there is none, and the site-wide author when the
 * document has none. A key the document has is never changed, and a document
 * that needs nothing is returned as it is. Only the copy is changed; the
 * source file is never written.
 *
 * Parameters:
 *   text (string) — the page's Markdown, directives already expanded
 *
 * Returns:
 *   string — text with the derived keys added to its front matter
 *
 * Example:
 *   out := cfg.deriveFrontMatter("# Hello\n")
 */
func (cfg *AppConfig) deriveFrontMatter(text string) string {
	fm, body, err := SplitFrontMatter([]byte(text))
	if err != nil {
		return text
	}
	doc := &CommonMark{FrontMatter: fm, Text: body}
	var add []string
	if strings.TrimSpace(doc.GetAttributeString("title", "")) == "" {
		if h := firstHeadingText(body); h != "" {
			add = append(add, yamlLine("title", h))
		}
	}
	if site := strings.TrimSpace(cfg.Author); site != "" && !hasAuthor(fm["author"]) {
		add = append(add, yamlLine("author", site))
	}
	if len(add) == 0 {
		return text
	}
	block := strings.Join(add, "")
	if !strings.HasPrefix(text, "---\n") {
		return "---\n" + block + "---\n" + text
	}
	if strings.HasPrefix(text, "---\n---\n") { // an empty block
		return "---\n" + block + text[4:]
	}
	end := strings.Index(text[4:], "\n---")
	if end < 0 {
		return text
	}
	insertAt := 4 + end + 1
	return text[:insertAt] + block + text[insertAt:]
}

// yamlLine writes one front matter line, quoted as YAML needs.
func yamlLine(key, value string) string {
	b, err := yaml.Marshal(map[string]string{key: value})
	if err != nil {
		return key + ": " + value + "\n"
	}
	return string(b)
}
