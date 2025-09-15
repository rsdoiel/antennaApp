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
	"strings"
	"time"

	// 3rd Party Packages
	"github.com/mmcdole/gofeed"
)

// Enclosure holds the data for RSS enclusure support
type Enclosure struct {
	Url    string `json:"url,omitempty" yaml:"url,omitempty"`
	Length string    `json:"length,omitempty" yaml:"length,omitempty"`
	Type   string `json:"type,omitempty" yaml:"type,omitempty"`
}

func toXMLString(input string) string {
	const (
		XML_AMP = "&#38;"
		XML_APOS = "&#39;"
		XML_GT = "&#62;"
		XML_LT = "&#60;"
		XML_QUOT = "&#34;"
	)
	input = strings.ReplaceAll(input, "&", XML_AMP) // Encode ampersand first to avoid double encoding
	input = strings.ReplaceAll(input, "<", XML_LT)  // Less than sign
	input = strings.ReplaceAll(input, ">", XML_GT)  // Greater than sign
	input = strings.ReplaceAll(input, "\"", XML_QUOT) // Double quote
	input = strings.ReplaceAll(input, "'", XML_APOS)  // Apostrophe
	return input
}

func (gen *Generator) WriteItemRSS(out io.Writer, link string, title string, description string, authors []*gofeed.Person,
	enclosures []*Enclosure, guid string, pubDate string, dcExt string,
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
		pressTime += ", updated: " + updated
	}
	// Wrap the Item
	fmt.Fprintf(out, `    <item>
`)
	defer fmt.Fprintf(out, `    </item>`)
	// Setup the Title
	if title != "" {
		fmt.Fprintf(out, "      <title>%s</title>\n", strings.TrimSpace(toXMLString(title)))
	}
	if link != "" {
		fmt.Fprintf(out, "      <link>%s</link>\n", strings.TrimSpace(toXMLString(link)))
	}
	if description != "" {
		fmt.Fprintf(out, `      <description>
        <![CDATA[%s]]>
      </description>
`, indentText(strings.TrimSpace(description), 8))
	}
	if authors != nil {
        for _, author := range authors  {
			if author.Email != "" && author.Name != "" {
				fmt.Fprintf(out, "      <author>%s (%s)</author>\n", author.Email, author.Name)
			}
        }
	}
	if enclosures != nil && len(enclosures) > 0 {
		for _, enclosure := range enclosures {
			fmt.Fprintf(out, `      <enclosure url=%q length=%q type=%q />
`, strings.TrimSpace(enclosure.Url), enclosure.Length, strings.TrimSpace(enclosure.Type))
		}
	}
	if guid != "" {
		fmt.Fprintf(out, "      <guid>cid://%s</guid>\n", strings.TrimSpace(toXMLString(guid)))
	}
	if pubDate != "" {
		d, err := time.Parse("2006-01-02", pubDate)
		if err == nil {
			fmt.Fprintf(out, "      <pubDate>%s</pubDate>\n", d.Format(time.RFC822Z))
		}
	}
	return nil
}

// WriteRSS writes aggregated items into an HTML page from the contents of the database
func (gen *Generator) WriteRSS(out io.Writer, db *sql.DB, appName string, collection *Collection) error {
	// Create the outer elements of a page.
	rssLink := strings.TrimSuffix(collection.File, ".md") + ".xml"
	feedLink := fmt.Sprintf("%s/%s", gen.BaseURL, rssLink)
	if collection.Link != "" {
		if strings.Contains(collection.Link, "://") {
			feedLink = collection.Link
		} else {
			feedLink = fmt.Sprintf("%s/%s", gen.BaseURL, collection.Link)
		}
 	}
	fmt.Fprintf(out, `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
  <atom:link href=%q rel="self" type="application/rss+xml" />
  <channel>
`, feedLink)
	defer fmt.Fprintln(out, `
  </channel>
</rss>`)
	// Channel Metadata
	if collection.Title != "" {
		fmt.Fprintf(out, `    <title>%s</title>
`, collection.Title)
	}
	if collection.Description != "" {
		fmt.Fprintf(out, `    <description>
      %s
    </description>
`, indentText(strings.TrimSpace(collection.Description), 6))
	}
	if feedLink != "" {
		fmt.Fprintf(out, `    <link>%s</link>
`, feedLink)
	}
	if collection.Copyright != "" {
		fmt.Fprintf(out, `    <copyright>%s</copyright>
`, strings.TrimSpace(collection.Copyright))
	}
	if collection.ManagingEditor != "" {
		fmt.Fprintf(out, `    <managingEditor>%s</managingEditor>
`, strings.TrimSpace(collection.ManagingEditor))
	}
	if collection.WebMaster != "" {
		fmt.Fprintf(out, `    <webMaster>%s</webMaster>
`, strings.TrimSpace(collection.WebMaster))
	}
	if collection.PubDate != "" {
		fmt.Fprintf(out, `    <pubDate>%s</pubDate>
`, strings.TrimSpace(collection.PubDate))
	}
	// The following are hardcode because they are dependent on the generator and
	// when it executed.
	timestamp := time.Now().Format(time.RFC822Z)
	fmt.Fprintf(out, `    <lastBuildDate>%s</lastBuildDate>
`, timestamp)
	fmt.Fprintf(out, `    <generator>%s/%s</generator>
    <docs>https://cyber.harvard.edu/rss/rss.html</docs>
`, appName, Version)


	// Setup  items
	stmt := SQLDisplayItems
	rows, err := db.Query(stmt)
	if err != nil {
		return err
	}
	defer rows.Close()
	// Setup and write out the body
	for rows.Next() {
		var (
			link          string
			title         string
			description   string
            authorsSrc    string
			authors       []*gofeed.Person
			enclosuresSrc string
			enclosures    []*Enclosure
			guid          string
			pubDate       string
			dcExt         string
			channel       string
			status        string
			updated       string
            label         string
			postPath      string
		)
		if err := rows.Scan(&link, &title, &description, &authorsSrc,
              &enclosuresSrc, &guid, &pubDate, &dcExt,
              &channel, &status, &updated, &label, &postPath); err != nil {
            return err
		}
        if authorsSrc != "" {
			// Do we have a JSON object?
            authors = []*gofeed.Person{}
            if  err := json.Unmarshal([]byte(authorsSrc), &authors); err != nil {
                fmt.Fprintf(gen.eout, "error (%s): %s\n", authorsSrc, err)
                authors = nil
            }
        }
		enclosures = []*Enclosure{}
        if enclosuresSrc != "" {
            if err := json.Unmarshal([]byte(enclosuresSrc), &enclosures); err != nil {
                fmt.Fprintf(gen.eout, "error (%s): %s\n", enclosuresSrc, err)
                enclosures = nil
            }
        }
		if postPath != "" {
			if fi, err := os.Stat(postPath); err == nil {
				enclosure := &Enclosure{
					Url: gen.BaseURL + "/" + postPath,
					Length: fmt.Sprintf("%d", fi.Size()),
					Type: "text/markdown",
				}
				addEnclosure := true
				if len(enclosures) > 0 {
					// Make sure we're not add a duplicate
					for _, item := range enclosures {
						if item.Url == enclosure.Url {
							addEnclosure = false;
							// Let's update the existing enclosure
							item.Url = enclosure.Url
							item.Length = enclosure.Length
							item.Type = enclosure.Type
							break
						}	
					} 
				}
				if addEnclosure {
					enclosures = append(enclosures, enclosure)
				}
			}
		}
		if err := gen.WriteItemRSS(out, link, title, description, authors,
			enclosures, guid, pubDate, dcExt,
			channel, status, updated, label); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	// Close the body via defer
	return nil
}
