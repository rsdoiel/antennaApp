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
	"encoding/json"
	"fmt"
	"html"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	// 3rd Party Packages
	"github.com/mmcdole/gofeed"
	ext "github.com/mmcdole/gofeed/extensions"
)

// Write HTML for an item. cfg controls body-content rendering per the
// items: block in page.yaml (DEC-022–031); it does not affect the
// data-pagefind-filter metadata built below (DEC-025). The returned bool
// is true when the item was omitted entirely (items.link.missing: omit —
// DEC-027) and no output was written for it.
func (gen *Generator) WriteItem(out io.Writer, link string, title string, description string, authors []*gofeed.Person,
	sourceMarkdown string, enclosures []*Enclosure, guid string, pubDate string, dcExtSrc string,
	channel string, status string, updated string, label string, categories string, cfg ItemsConfig) (bool, error) {
	cfg.applyDefaults()

	linkRes, err := resolveItemLink(link, title, channel, cfg.Link)
	if err != nil {
		return false, err
	}
	if linkRes.Omit {
		return true, nil
	}

	showField := func(name string) bool {
		if len(cfg.Fields) == 0 {
			return true
		}
		for _, f := range cfg.Fields {
			if f == name {
				return true
			}
		}
		return false
	}

	// Build PageFind filter attribute values
	var filters []string

	// RSS/Atom categories
	if categories != "" {
		var cats []string
		if err := json.Unmarshal([]byte(categories), &cats); err == nil {
			for _, c := range cats {
				if c != "" {
					filters = append(filters, "category:"+c)
				}
			}
		}
	}

	// Dublin Core extension fields
	if dcExtSrc != "" {
		var dc ext.DublinCoreExtension
		if err := json.Unmarshal([]byte(dcExtSrc), &dc); err == nil {
			dcFields := []struct {
				key    string
				values []string
			}{
				{"dc_title", dc.Title},
				{"dc_creator", dc.Creator},
				{"dc_author", dc.Author},
				{"dc_subject", dc.Subject},
				{"dc_description", dc.Description},
				{"dc_publisher", dc.Publisher},
				{"dc_contributor", dc.Contributor},
				{"dc_date", dc.Date},
				{"dc_type", dc.Type},
				{"dc_format", dc.Format},
				{"dc_identifier", dc.Identifier},
				{"dc_source", dc.Source},
				{"dc_language", dc.Language},
				{"dc_relation", dc.Relation},
				{"dc_coverage", dc.Coverage},
				{"dc_rights", dc.Rights},
			}
			for _, f := range dcFields {
				for _, v := range f.values {
					if v != "" {
						filters = append(filters, f.key+":"+v)
					}
				}
			}
		}
	}

	// Native authors
	for _, a := range authors {
		if a != nil && a.Name != "" {
			filters = append(filters, "author:"+a.Name)
		}
	}

	// Publication date
	if pubDate != "" {
		filters = append(filters, "datePublished:"+pubDate)
	}

	// Feed provenance
	if label != "" {
		filters = append(filters, "label:"+label)
	}
	if channel != "" {
		filters = append(filters, "channel:"+channel)
	}

	// Build heading — h2 keeps article titles subordinate to the page h1.
	// title/label are feed-supplied text, so they must be HTML-escaped
	// before being embedded — otherwise a title like "Q&A: ..." breaks
	// the page's markup. htmlEscapeFeedText also normalizes titles some
	// feeds have already (over-)escaped, rather than double-escaping them.
	var headingHTML string
	if showField("title") {
		if title == "" {
			headingHTML = fmt.Sprintf("<h2>@%s</h2>", htmlEscapeFeedText(label))
		} else {
			headingHTML = fmt.Sprintf("<h2>%s</h2>", htmlEscapeFeedText(title))
		}
	}

	// Build date using <time> elements so the date is machine-readable and
	// not mixed into the heading's accessible name. The datetime attribute
	// always uses an ISO-style layout regardless of cfg.DateFormat, since
	// it is machine-readable metadata, not display text (DEC-028).
	var dateHTML string
	if showField("pubDate") {
		machinePub := formatItemDate(pubDate, "2006-01-02")
		displayPub := formatItemDate(pubDate, cfg.DateFormat)
		if machinePub != "" {
			dateHTML = fmt.Sprintf(`<time datetime=%q>%s</time>`, machinePub, displayPub)
			if updated != "" {
				machineUpd := formatItemDate(updated, "2006-01-02")
				displayUpd := formatItemDate(updated, cfg.DateFormat)
				if machineUpd != machinePub {
					dateHTML += fmt.Sprintf(`, updated: <time datetime=%q>%s</time>`, machineUpd, displayUpd)
				}
			}
		}
	}

	var content string
	var contentIsBlockHTML bool
	if showField("content") {
		content, contentIsBlockHTML, err = resolveItemContent(description, sourceMarkdown, cfg)
		if err != nil {
			return false, err
		}
	}

	// Source attribution line — additive to the existing footer, per the
	// same "supplementary information about the article" rationale as
	// DEC-017's <footer> choice.
	var sourceHTML string
	if showField("source") && cfg.ShowSource != nil && *cfg.ShowSource && label != "" {
		sourceHTML = fmt.Sprintf(`<p class="source">via %s</p>`, htmlEscapeFeedText(label))
	}

	// linkRes.Label is feed-supplied text (title/link) and linkRes.Href is
	// a feed-supplied URL; both must be HTML-escaped before embedding.
	var footerInner string
	if linkRes.AsPlainText {
		footerInner = htmlEscapeFeedText(linkRes.Label)
	} else {
		footerInner = fmt.Sprintf(`<a href="%s">%s</a>`, html.EscapeString(linkRes.Href), htmlEscapeFeedText(linkRes.Label))
	}
	footerHTML := "<footer>\n        " + footerInner
	if sourceHTML != "" {
		footerHTML += "\n        " + sourceHTML
	}
	footerHTML += "\n      </footer>"

	// Build the body from only the sections cfg.Fields selects, so an
	// excluded field's wrapper element is omitted entirely rather than
	// left behind as an empty <p></p> (found via Phase 7 smoke testing).
	var bodyParts []string
	if headingHTML != "" {
		bodyParts = append(bodyParts, headingHTML)
	}
	if showField("pubDate") {
		bodyParts = append(bodyParts, fmt.Sprintf("<p>%s</p>", dateHTML))
	}
	if showField("content") {
		// content is only plain text (safe to wrap in a single <p>) when it
		// came from the "strip"/"escape" raw-description paths. Rendered
		// sourceMarkdown and "unsafe" passthrough are already block-level
		// HTML with their own <p>/<ul>/<blockquote> elements — wrapping
		// those in another <p> produces invalid nested markup that browsers
		// silently mangle by auto-closing the outer <p>.
		if contentIsBlockHTML {
			bodyParts = append(bodyParts, content)
		} else {
			bodyParts = append(bodyParts, fmt.Sprintf("<p>%s</p>", content))
		}
	}
	bodyParts = append(bodyParts, footerHTML)
	bodyHTML := strings.Join(bodyParts, "\n      ")

	// pubDate/link/filters are feed-supplied text embedded in HTML
	// attributes; they must be HTML-escaped, not just Go-%q-quoted, or a
	// query-string link like "?a=1&b=2" produces an invalid raw "&" in
	// the attribute value.
	if len(filters) > 0 {
		fmt.Fprintf(out, `
    <article data-published="%s" data-link="%s" data-pagefind-filter="%s">
      %s
    </article>
`, html.EscapeString(pubDate), html.EscapeString(linkRes.Href), html.EscapeString(strings.Join(filters, ", ")), bodyHTML)
	} else {
		fmt.Fprintf(out, `
    <article data-published="%s" data-link="%s">
      %s
    </article>
`, html.EscapeString(pubDate), html.EscapeString(linkRes.Href), bodyHTML)
	}
	return false, nil
}

// htmlEscapeFeedText HTML-escapes s for safe embedding as text content,
// first unescaping any HTML entities s already contains. Some feeds
// double-HTML-escape their title text before publishing it (confirmed live
// on koreaherald.com item 10812786: its <title> contains the literal text
// "&quot;", not a real quote character). Escaping such a title directly
// would turn that literal "&quot;" into visible "&amp;quot;" text instead
// of a quote mark. Unescaping first collapses any such pre-existing
// entities to their real characters, so the following escape produces
// exactly one, correct level of encoding regardless of whether the source
// already (over-)encoded the text.
func htmlEscapeFeedText(s string) string {
	return html.EscapeString(html.UnescapeString(s))
}

// stripTagsPattern matches an HTML tag for stripTags. A deliberate
// simplification, not a full HTML parser — acceptable because it only ever
// runs on the raw-description-fallback path (DEC-024), never on the common
// sourceMarkdown path.
var stripTagsPattern = regexp.MustCompile(`<[^>]*>`)

// stripTags removes HTML tags from s, returning the remaining text
// unchanged. Not a full HTML parser: malformed or unclosed tags are
// removed on a best-effort basis (DEC-024).
func stripTags(s string) string {
	return stripTagsPattern.ReplaceAllString(s, "")
}

// truncateWords truncates s to at most maxLen characters, backing off to
// the last preceding whitespace boundary so a word is never cut in half
// (DEC-029). If s is already at or under maxLen, it is returned unchanged.
// If no whitespace exists at or before maxLen, s is hard-truncated to
// maxLen (documented limitation, not treated as an error).
func truncateWords(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	cut := s[:maxLen]
	if idx := strings.LastIndexAny(cut, " \t\n"); idx >= 0 {
		return cut[:idx]
	}
	return cut
}

// resolveItemContent resolves a feed item's rendered body content following
// the content-source precedence declared in DEC-023: sourceMarkdown is
// preferred when non-empty (rendered via CommonMark, safe or unsafe per
// cfg.HTML — DEC-024), falling back to the raw feed description (tag-
// stripped, escaped, or passed through unchanged per cfg.HTML) only when
// sourceMarkdown is empty. cfg.ContentMaxLength, if set, truncates the
// resolved pre-render source text on a word boundary before conversion
// (DEC-029), never the rendered HTML.
//
// The isBlockHTML return reports whether content is already rendered,
// block-level HTML (sourceMarkdown is always rendered via CommonMark; raw
// description is block HTML only in cfg.HTML == "unsafe" passthrough).
// Callers must not wrap block HTML in another <p>: CommonMark/unsafe
// passthrough content already contains its own <p>/<ul>/<blockquote>
// elements, and re-wrapping it produces invalid nested markup that
// browsers silently mangle by auto-closing the outer <p>.
func resolveItemContent(description, sourceMarkdown string, cfg ItemsConfig) (content string, isBlockHTML bool, err error) {
	source := sourceMarkdown
	usedMarkdown := true
	if source == "" {
		source = description
		usedMarkdown = false
	}
	if cfg.ContentMaxLength > 0 {
		source = truncateWords(source, cfg.ContentMaxLength)
	}
	if usedMarkdown {
		doc := &CommonMark{Text: source}
		if cfg.HTML == "unsafe" {
			rendered, err := doc.ToUnsafeHTML()
			return rendered, true, err
		}
		rendered, err := doc.ToHTML()
		return rendered, true, err
	}
	switch cfg.HTML {
	case "escape":
		return html.EscapeString(source), false, nil
	case "unsafe":
		return source, true, nil
	default: // "strip", or unset
		return stripTags(source), false, nil
	}
}

// LinkResolution is the result of resolving a feed item's anchor per
// items.link configuration (DEC-026, DEC-027).
type LinkResolution struct {
	// Href is the anchor's target URL. Empty when AsPlainText is true.
	Href string
	// Label is the anchor's (or plain text's) visible label.
	Label string
	// AsPlainText, when true, means render Label as plain text with no
	// <a> element (Missing == "unlinked", or a source_link fallback with
	// no channel URL available either).
	AsPlainText bool
	// Omit, when true, means exclude the item from output entirely
	// (Missing == "omit"). Callers must check this before writing any
	// output for the item.
	Omit bool
}

// resolveItemLinkLabel computes the anchor label per cfg.LabelField
// (DEC-026): the literal sentinel "static" always uses cfg.LabelFallback;
// otherwise the named field's value is used, falling back to
// cfg.LabelFallback when that value is empty.
func resolveItemLinkLabel(link, title string, cfg LinkConfig) string {
	if cfg.LabelField == "static" {
		return cfg.LabelFallback
	}
	var value string
	switch cfg.LabelField {
	case "link":
		value = link
	case "title":
		value = title
	}
	if value == "" {
		return cfg.LabelFallback
	}
	return value
}

// resolveItemLink resolves a feed item's anchor per items.link
// configuration (DEC-026, DEC-027). channelURL is the parent feed/
// channel's URL, used only for the "source_link" missing-link fallback.
func resolveItemLink(link, title, channelURL string, cfg LinkConfig) (LinkResolution, error) {
	label := resolveItemLinkLabel(link, title, cfg)
	if link != "" {
		return LinkResolution{Href: link, Label: label}, nil
	}
	if cfg.Required {
		return LinkResolution{}, fmt.Errorf("item link is required but empty (items.link.required: true)")
	}
	switch cfg.Missing {
	case "omit":
		return LinkResolution{Omit: true}, nil
	case "source_link":
		if channelURL != "" {
			return LinkResolution{Href: channelURL, Label: label}, nil
		}
		// DEC-027 — channel URL also empty; fall back to unlinked rather
		// than emitting an empty href.
		return LinkResolution{Label: label, AsPlainText: true}, nil
	default: // "unlinked", or unset
		return LinkResolution{Label: label, AsPlainText: true}, nil
	}
}

// storedDateLayouts are the Go reference layouts a stored pubDate/updated
// value may come back as, tried in order. RFC3339 is what the production
// "sqlite" driver (github.com/glebarez/go-sqlite, a pure-Go SQLite driver)
// actually returns when scanning a DATETIME-affinity column into a string —
// confirmed by an end-to-end smoke test (Phase 7) against a real generated
// site, which showed date_format silently having no effect. The
// space-separated layout is what saveItem (harvest.go) writes as a bound
// parameter string and is what the "sqlite3" driver (mattn/go-sqlite3, used
// by this package's own tests) returns verbatim without reinterpreting it.
var storedDateLayouts = []string{
	time.RFC3339,
	"2006-01-02 15:04:05",
}

// formatItemDate formats a stored pubDate/updated value per items.date_format
// (DEC-028). raw is parsed against each of storedDateLayouts in turn and
// reformatted with layout on the first successful parse. When raw matches
// neither — the feed's raw, unparsed date string, in a format gofeed
// couldn't parse — the result falls back to the first 10 characters of raw
// (or raw unchanged if shorter), exactly matching current WriteItem
// truncation behavior, rather than returning the full untruncated string or
// erroring.
func formatItemDate(raw string, layout string) string {
	for _, l := range storedDateLayouts {
		if t, err := time.Parse(l, raw); err == nil {
			return t.Format(layout)
		}
	}
	if len(raw) > 10 {
		return raw[:10]
	}
	return raw
}

// elementFromMap, generate an HTML element from a map[string]string
func elementFromMap(element string, m map[string]string) string {
	parts := []string{}
	parts = append(parts, fmt.Sprintf("<%s", element))
	for k, v := range m {
		parts = append(parts, fmt.Sprintf("%s=%q", k, v))
	}
	if element == "script" {
		parts = append(parts, "></script>")
	} else {
		parts = append(parts, ">")
	}
	return strings.Join(parts, " ")
}

// writeHeadElement, writes the head element of the HTML page.
func (gen *Generator) writeHeadElement(out io.Writer, postPath string, frontMatter map[string]interface{}) {
	fmt.Fprintln(out, "<head>")
	defer fmt.Fprintln(out, "</head>")
	var m map[string]string
	// Write out charset
	m = map[string]string{
		"charset": "utf-8",
	}
	fmt.Fprintf(out, "  %s\n", elementFromMap("meta", m))
	// Format the date
	formattedDate := time.Now().Format(time.RFC3339)
	m = map[string]string{
		"name":    "generator",
		"content": fmt.Sprintf("%s/%s", gen.AppName, gen.Version),
	}
	fmt.Fprintf(out, "  %s\n", elementFromMap("meta", m))
	m = map[string]string{
		"name":    "date",
		"content": formattedDate,
	}
	fmt.Fprintf(out, "  %s\n", elementFromMap("meta", m))
	// Always emit viewport meta for responsive and accessible rendering
	m = map[string]string{
		"name":    "viewport",
		"content": "width=device-width, initial-scale=1.0",
	}
	fmt.Fprintf(out, "  %s\n", elementFromMap("meta", m))
	if gen.Meta != nil && len(gen.Meta) > 0 {
		for _, m := range gen.Meta {
			fmt.Fprintf(out, "  %s\n", elementFromMap("meta", m))
		}
	}
	// Write title — front matter title takes precedence over gen.Title
	pageTitle := gen.Title
	if frontMatter != nil {
		if t, ok := frontMatter["title"].(string); ok && t != "" {
			pageTitle = t
		}
	}
	if pageTitle != "" {
		fmt.Fprintf(out, "  <title>%s</title>\n", pageTitle)
	}
	// Write out RSS alt link for Markdown if postPath is not empty string
	if postPath != "" && strings.HasSuffix(postPath, ".md") {
		// NOTE: Posts are written next to the HTML page so the link to the Markdown can be relative
		postLink := filepath.Base(postPath)
		m = map[string]string{
			"title": pageTitle,
			"rel":   "alternate",
			"type":  "text/markdown",
			"href":  postLink,
		}
		fmt.Fprintf(out, "  %s\n", elementFromMap("link", m))
	}
	if gen.Link != nil && len(gen.Link) > 0 {
		for _, m := range gen.Link {
			fmt.Fprintf(out, "  %s\n", elementFromMap("link", m))
		}
	}
	if gen.Script != nil && len(gen.Script) > 0 {
		for _, m := range gen.Script {
			fmt.Fprintf(out, "  %s\n", elementFromMap("script", m))
		}
	}
	if gen.Style != "" {
		fmt.Fprintf(out, "  <style>\n%s\n</style>\n", indentText(strings.TrimSpace(gen.Style), 4))
	}
	// Emit front matter fields as standard HTML meta and PageFind filter attributes
	if frontMatter != nil {
		allowed := map[string]bool{}
		for _, k := range gen.AllowedMetaFields {
			allowed[k] = true
		}
		doc := &CommonMark{FrontMatter: frontMatter}
		for key, val := range frontMatter {
			if key == "title" {
				continue // already handled in <title>
			}
			if len(allowed) > 0 && !allowed[key] {
				continue
			}
			values := doc.GetAttributeStringSlice(key)
			if len(values) == 0 {
				if s, ok := val.(string); ok && s != "" {
					values = []string{s}
				}
			}
			for _, v := range values {
				fmt.Fprintf(out, "  <meta name=%q content=%q>\n", key, v)
				fmt.Fprintf(out, "  <meta data-pagefind-filter=%q content=%q>\n", key+"[content]", v)
			}
		}
	}
}

// indentText splits  the string into lines, then prefixes the number of
// spaces specified to each line returning updated text
func indentText(src string, spaces int) string {
	lines := strings.Split(src, "\n")
	return strings.Join(lines, "\n"+strings.Repeat(" ", spaces))
}

// WriteHTML writes aggregated items into an HTML page from the contents of the database
func (gen *Generator) WriteHTML(out io.Writer, db *sql.DB, cfgName string, collection *Collection) error {
	// Apply items: config defaults and validate enum values once, up
	// front, so a typo aborts generation instead of silently misrendering
	// every item (DEC-022–031).
	gen.Items.applyDefaults()
	if err := gen.Items.validate(); err != nil {
		return err
	}
	// Create the outer elements of a page.
	fmt.Fprintf(out, "<!doctype html>\n<html lang=%q>\n", gen.Lang)
	defer fmt.Fprintln(out, "</html>")
	// Setup the metadata in the head element
	gen.writeHeadElement(out, "", nil)
	// Setup body element
	fmt.Fprintln(out, "<body>")
	defer fmt.Fprintln(out, "</body>")
	// Skip navigation link — WCAG 2.4.1: keyboard users bypass repeated nav blocks
	fmt.Fprintln(out, `  <a href="#main-content" class="skip-link">Skip to main content</a>`)
	// Setup header element
	if gen.Header != "" {
		fmt.Fprintf(out, "  <header>\n    %s\n  </header>\n", indentText(strings.TrimSpace(gen.Header), 4))
	} else {
		fmt.Fprintln(gen.eout, "warning: aggregate page has no <h1>; set a 'header' value in the generator YAML")
	}
	// Setup nav element
	if gen.Nav != "" {
		fmt.Fprintf(out, `  <nav aria-label="Site navigation">
    %s
  </nav>
`, indentText(strings.TrimSpace(gen.Nav), 4))
	}
	if gen.TopContent != "" {
		fmt.Fprintf(out, `
    %s
`, indentText(strings.TrimSpace(gen.TopContent), 2))
	}
	// main landmark wraps the primary feed content
	fmt.Fprintln(out, `  <main id="main-content">`)
	stmt := SQLDisplayItems
	rows, err := db.Query(stmt)
	if err != nil {
		return err
	}
	defer rows.Close()
	// Setup and write out the body
	for rows.Next() {
		var (
			link           string
			title          string
			description    string
			authorsSrc     string
			authors        []*gofeed.Person
			enclosuresSrc  string
			enclosures     []*Enclosure
			guid           string
			pubDate        string
			dcExt          string
			channel        string
			status         string
			updated        string
			label          string
			postPath       string
			sourceMarkdown string
			categories     string
		)
		if err := rows.Scan(&link, &title, &description, &authorsSrc,
			&enclosuresSrc, &guid, &pubDate, &dcExt,
			&channel, &status, &updated, &label, &postPath, &sourceMarkdown,
			&categories); err != nil {
			fmt.Fprintf(gen.eout, "error (%s): %s\n", stmt, err)
			continue
		}
		if authorsSrc != "" {
			authors = []*gofeed.Person{}
			if err := json.Unmarshal([]byte(authorsSrc), &authors); err != nil {
				fmt.Fprintf(gen.eout, "error (authors: %s): %s\n", authorsSrc, err)
				authors = nil
			}
		}
		if enclosuresSrc != "" {
			enclosures = []*Enclosure{}
			if err := json.Unmarshal([]byte(enclosuresSrc), &enclosures); err != nil {
				fmt.Fprintf(gen.eout, "error (enclosures: %s): %s\n", err, enclosuresSrc)
				enclosures = nil
			}
		}
		if _, err := gen.WriteItem(out, link, title, description, authors,
			sourceMarkdown, enclosures, guid, pubDate, dcExt,
			channel, status, updated, label, categories, gen.Items); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	fmt.Fprintln(out, "  </main>")
	if gen.BottomContent != "" {
		fmt.Fprintf(out, `
    %s
`, indentText(strings.TrimSpace(gen.BottomContent), 2))
	}
	if gen.Footer != "" {
		fmt.Fprintf(out, "  <footer>\n    %s\n  </footer>\n", indentText(strings.TrimSpace(gen.Footer), 4))
	}
	// close the body
	return nil
}

// pageDisplayName derives a human-readable label from a Markdown input path.
// "about.md" → "About", "my-page.md" → "My Page".
func pageDisplayName(inputPath string) string {
	base := filepath.Base(inputPath)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.ReplaceAll(name, "_", " ")
	if len(name) > 0 {
		name = strings.ToUpper(name[:1]) + name[1:]
	}
	return name
}

// WritePageIndex renders a simple `<ul>` link list from the pages table of db.
// It is used when a collection has mode: page-index. Each row from the pages
// table becomes one `<li><a href="outputPath">displayName</a></li>` entry.
func (gen *Generator) WritePageIndex(out io.Writer, db *sql.DB) error {
	rows, err := db.Query(SQLPageIndexItems)
	if err != nil {
		return err
	}
	defer rows.Close()

	fmt.Fprintln(out, `  <main id="main-content">`)
	fmt.Fprintln(out, "    <ul>")
	for rows.Next() {
		var inputPath, outputPath string
		if err := rows.Scan(&inputPath, &outputPath); err != nil {
			fmt.Fprintf(gen.eout, "error (page-index row): %s\n", err)
			continue
		}
		// Normalise outputPath: ensure it starts with / for web-root linking
		href := "/" + strings.TrimLeft(filepath.ToSlash(outputPath), "/")
		label := pageDisplayName(inputPath)
		fmt.Fprintf(out, "      <li><a href=%q>%s</a></li>\n", href, label)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	fmt.Fprintln(out, "    </ul>")
	fmt.Fprintln(out, "  </main>")
	return nil
}

// WriteHtmlPage renders a post as an HTML Page using HTML connent and wrapping it based on the
// generator configuration.
func (gen *Generator) WriteHtmlPage(htmlName string, link string, postPath, pubDate string, innerHTML string, frontMatter map[string]interface{}) error {
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
	defer out.Close()

	// Create the outer elements of a page.
	fmt.Fprintf(out, "<!doctype html>\n<html lang=%q>\n", gen.Lang)
	defer fmt.Fprintln(out, "</html>")
	// Setup the metadata in the head element
	gen.writeHeadElement(out, postPath, frontMatter)
	// Setup body element
	fmt.Fprintln(out, "<body>")
	defer fmt.Fprintln(out, "</body>")
	// Skip navigation link — WCAG 2.4.1: keyboard users bypass repeated nav blocks
	fmt.Fprintln(out, `  <a href="#main-content" class="skip-link">Skip to main content</a>`)
	// Setup header element
	if gen.Header != "" {
		fmt.Fprintf(out, "  <header>\n    %s\n  </header>\n", indentText(strings.TrimSpace(gen.Header), 4))
	}
	// Setup nav element
	if gen.Nav != "" {
		fmt.Fprintf(out, `  <nav aria-label="Site navigation">
    %s
  </nav>
`, indentText(strings.TrimSpace(gen.Nav), 4))
	}

	if gen.TopContent != "" {
		fmt.Fprintf(out, `
    %s
`, indentText(strings.TrimSpace(gen.TopContent), 4))
	}

	// Now render our innerHTML
	if pubDate != "" && link != "" {
		fmt.Fprintf(out, `
  <main id="main-content">
    <article data-published=%q data-link=%q>
      %s
    </article>
  </main>
`, pubDate, link, indentText(innerHTML, 6))

	} else {
		fmt.Fprintf(out, `
  <main id="main-content">
    %s
  </main>
`, indentText(innerHTML, 4))
	}

	if gen.BottomContent != "" {
		fmt.Fprintf(out, `
    %s
`, indentText(strings.TrimSpace(gen.BottomContent), 4))
	}

	// Wrap up the page
	if gen.Footer != "" {
		fmt.Fprintf(out, "  <footer>\n    %s\n  </footer>\n", indentText(strings.TrimSpace(gen.Footer), 4))
	}
	return nil
}
