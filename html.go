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
)

// Write HTML for an item
func (gen *Generator) WriteItem(out io.Writer, link string, title string, description string, authors []*gofeed.Person,
	sourceMarkdown string, enclosures []*Enclosure, guid string, pubDate string, dcExt string,
	channel string, status string, updated string, label string) error {
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

	// Setup the Title
	if title == "" {
		title = fmt.Sprintf("<h1>@%s</h1>\n\n(date: %s)", label, pressTime)
	} else {
		title = fmt.Sprintf("<h1>%s</h1>\n\n(date: %s)", title, pressTime)
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

	fmt.Fprintf(out, `
    <article data-published=%q data-link=%q>
      %s
      <p>
      %s
      <address>
        <a href=%q>%s</a>
      </address>
    </article>
`, pubDate, link, title, content, link, link)
	return nil
}

// elementFromMap, generate an HTML element from a map[string]string
func elementFromMap(element string, m map[string]string) string {
	parts := []string{}
	parts = append(parts, fmt.Sprintf("<%s", element))
	for k, v := range m {
		parts = append(parts, fmt.Sprintf("%s=%q", k, v))
	}
	parts = append(parts, "/>")
	return strings.Join(parts, " ")
}

// writeHeadElement, writes the head element of the HTML page.
func (gen *Generator) writeHeadElement(out io.Writer, postPath string) {
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
	if gen.Meta != nil && len(gen.Meta) > 0 {
		for _, m := range gen.Meta {
			fmt.Fprintf(out, "  %s\n", elementFromMap("meta", m))
		}
	}
	// Write title (NOTE: title must come after the charset since it may have encoded characters)
	if gen.Title != "" {
		fmt.Fprintf(out, "  <title>%s</title>\n", gen.Title)
	}
	// Write out RSS alt link for Markdown if postPath is not empty string
	if postPath != "" && strings.HasSuffix(postPath, ".md") {
		// NOTE: Posts are written next to the HTML page so the link to the Markdown can be relative
		postLink := filepath.Base(postPath)
		m = map[string]string{
			"title": gen.Title,
			"rel":   "altenate",
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
	fmt.Fprintln(out, `<!doctype html>
<html lang="en-US">`)
	defer fmt.Fprintln(out, "</html>")
	// Setup the metadata in the head element
	gen.writeHeadElement(out, "")
	// Setup body element
	fmt.Fprintln(out, "<body>")
	defer fmt.Fprintln(out, "</body>")
	// Setup header element
	if gen.Header != "" {
		fmt.Fprintf(out, "  <header>\n    %s\n  </header>\n", indentText(strings.TrimSpace(gen.Header), 4))
	}
	// Setup nav element
	if gen.Nav != "" {
		fmt.Fprintf(out, `  <nav>
    %s
  </nav>
`, indentText(strings.TrimSpace(gen.Nav), 4))
	}
	if gen.TopContent != "" {
		fmt.Fprintf(out, `
    %s
`, indentText(strings.TrimSpace(gen.TopContent), 2))
	}
	// Setup section
	fmt.Fprintln(out, "  <section>")
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
		)
		if err := rows.Scan(&link, &title, &description, &authorsSrc,
			&enclosuresSrc, &guid, &pubDate, &dcExt,
			&channel, &status, &updated, &label, &postPath, &sourceMarkdown); err != nil {
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
			channel, status, updated, label); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	fmt.Fprintln(out, "  </section>")
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

// WriteHtmlPage renders a post as an HTML Page using HTML connent and wrapping it based on the
// generator configuration.
func (gen *Generator) WriteHtmlPage(htmlName string, link string, postPath, pubDate string, innerHTML string) error {
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
	fmt.Fprintln(out, `<!doctype html>
<html lang="en-US">`)
	defer fmt.Fprintln(out, "</html>")
	// Setup the metadata in the head element
	gen.writeHeadElement(out, postPath)
	// Setup body element
	fmt.Fprintln(out, "<body>")
	defer fmt.Fprintln(out, "</body>")
	// Setup header element
	if gen.Header != "" {
		fmt.Fprintf(out, "  <header>\n    %s\n  </header>\n", indentText(strings.TrimSpace(gen.Header), 4))
	}
	// Setup nav element
	if gen.Nav != "" {
		fmt.Fprintf(out, `  <nav>
    %s
  </nav>
`, indentText(strings.TrimSpace(gen.Nav), 4))
	}

	if gen.TopContent != "" {
		fmt.Fprintf(out, `
    %s
`, indentText(strings.TrimSpace(gen.TopContent), 4))
	}

	//fmt.Printf("DEBUG date-publised=%q data-link=%q\n", pubDate, link)
	// Now render our innerHTML
	if pubDate != "" && link != "" {
		fmt.Fprintf(out, `
  <section>
    <article data-published=%q data-link=%q>
      %s
    </article>
  </section>
`, pubDate, link, indentText(innerHTML, 6))

	} else {
		fmt.Fprintf(out, `
  <section>
    %s
  </section>
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
