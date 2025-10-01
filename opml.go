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
	"strings"
	"time"

	// Additional packages/modules
	"github.com/rsdoiel/opml"
	_ "github.com/glebarez/go-sqlite"
)
// WriteOPML writes out the feeds being followed in collection.
func (gen *Generator) WriteOPML(out io.Writer, db *sql.DB, appName string, collection *Collection) error {
		// Build our OPML structure from the collection and channels' table
		o := opml.New()
		o.Head.Title = collection.Title
		o.Head.Created = time.Now().Format(time.RFC822Z)
		o.Body.Outline = []*opml.Outline{}

		stmt := SQLDisplayChannels
		rows, err := db.Query(stmt)
        if err != nil {
                return err
        }
        defer rows.Close()
		// Setup and write out the body
        for rows.Next() {
			var (
				link string
				title string
				description string
				feedLink string
				linksSrc string
				updated string
				published string
				authorsSrc string
				language string
				copyright string
				generator string
				categoriesSrc string
				feedType string
				feedVersion string
			)
			 if err := rows.Scan(&link, &title, &description, &feedLink,
				&linksSrc, &updated, &published, &authorsSrc,
				&language, &copyright, &generator, &categoriesSrc, &feedType, &feedVersion); err != nil {
				return err
			}
			fmt.Printf("DEBUG link: %q, feedLink: %q, links (%T) -> %+v\n", link, feedLink, linksSrc, linksSrc)
			if link != "" && title != "" {
				entry := &opml.Outline{}
				entry.Text = title
				entry.XMLURL = link
				if linksSrc != "" {
					links := []string{}
					if err := json.Unmarshal([]byte(linksSrc), &links); err == nil  && len(links) > 0 {
						 entry.HTMLURL = strings.Join(links, ",")
					}
				}
				if feedLink != "" {
					entry.URL = feedLink
				}
				if feedType != "" {
					entry.Type = feedType 
				}
				if description != "" {
					entry.Description = description
				}
				if categoriesSrc != "" {
					categories := []string{}
					if err := json.Unmarshal([]byte(categoriesSrc), &categories); err == nil  && len(categories) > 0 {
						 entry.Category = strings.Join(categories, ",")
					}
				}
				o.Body.Outline = append(o.Body.Outline, entry)
			}
		}
		// Render the OPML content with the io.Writer
		src, err := opml.Marshal(o)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "%s\n", out)
		return nil
}