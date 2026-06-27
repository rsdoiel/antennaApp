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
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	// 3rd Party Packages
	"github.com/mmcdole/gofeed"
	ext "github.com/mmcdole/gofeed/extensions"
)

// Write HTML for an item
func (gen *Generator) WriteItem(out io.Writer, link string, title string, description string, authors []*gofeed.Person,
	sourceMarkdown string, enclosures []*Enclosure, guid string, pubDate string, dcExtSrc string,
	channel string, status string, updated string, label string, categories string) error {
	// Setup expressing update time.
	pressTime := pubDate
	if len(pressTime) > 10 {
		pressTime = pressTime[0:10]
	}
	if updated != "" {
		if len(updated) > 10 {
			updated = updated[0:10]
		}
		if pressTime != updated {
			pressTime += ", updated: " + updated
		}
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

	// Build heading — h2 keeps article titles subordinate to the page h1
	var headingHTML string
	if title == "" {
		headingHTML = fmt.Sprintf("<h2>@%s</h2>", label)
	} else {
		headingHTML = fmt.Sprintf("<h2>%s</h2>", title)
	}

	// Build date using <time> elements so the date is machine-readable and
	// not mixed into the heading's accessible name.
	pubShort := pubDate
	if len(pubShort) > 10 {
		pubShort = pubShort[0:10]
	}
	updShort := updated
	if len(updShort) > 10 {
		updShort = updShort[0:10]
	}
	var dateHTML string
	if pubShort != "" {
		dateHTML = fmt.Sprintf(`<time datetime=%q>%s</time>`, pubShort, pubShort)
		if updShort != "" && updShort != pubShort {
			dateHTML += fmt.Sprintf(`, updated: <time datetime=%q>%s</time>`, updShort, updShort)
		}
	}

	content := description
	if sourceMarkdown != "" {
		doc := &CommonMark{
			Text: sourceMarkdown,
		}
		if src, err := doc.ToHTML(); err == nil {
			content = src
		}
	}

	if len(filters) > 0 {
		fmt.Fprintf(out, `
    <article data-published=%q data-link=%q data-pagefind-filter=%q>
      %s
      <p>%s</p>
      <p>%s</p>
      <footer>
        <a href=%q>%s</a>
      </footer>
    </article>
`, pubDate, link, strings.Join(filters, ", "), headingHTML, dateHTML, content, link, link)
	} else {
		fmt.Fprintf(out, `
    <article data-published=%q data-link=%q>
      %s
      <p>%s</p>
      <p>%s</p>
      <footer>
        <a href=%q>%s</a>
      </footer>
    </article>
`, pubDate, link, headingHTML, dateHTML, content, link, link)
	}
	return nil
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
		if err := gen.WriteItem(out, link, title, description, authors,
			sourceMarkdown, enclosures, guid, pubDate, dcExt,
			channel, status, updated, label, categories); err != nil {
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
