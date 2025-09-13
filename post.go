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
	"os"
	"path/filepath"
	"strings"
	"time"

	 // 3rd Party
	_ "github.com/glebarez/go-sqlite"
	//"github.com/mmcdole/gofeed"
	ext "github.com/mmcdole/gofeed/extensions"
)

// Post will add a CommonMark document as a feed item and if the postPath and link
// are provided it will convert the CommonMark document to HTML and save it in the postPath.
func (app *AntennaApp) Post(cfgName string, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("expected a collection name and filepath in the collection")
	}
	cfg := &AppConfig{}
	if err := cfg.LoadConfig(cfgName); err != nil {
		return err
	}
	cName, fName := strings.TrimSpace(args[0]), strings.TrimSpace(args[1])
	collection, err := cfg.GetCollection(cName)
	if err != nil {
		return err
	}
	dsn := collection.DbName
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	src, err := os.ReadFile(fName)
	if err != nil {
		return err
	}
	doc := &CommonMark{}
	if err := doc.Parse(src); err != nil {
		return err
	}
	// Convert our document text to HTML
	innerHTML, err := doc.ToHTML()
	if err != nil {
		return err
	}
	title := doc.GetAttributeString("title", "")
	authors, err := doc.GetPersons("author", false)
	if err != nil {
		return err
	}
	description := doc.GetAttributeString("description", "")
	if description == "" {
		description = doc.GetAttributeString("abstract", "")
	}
	link := doc.GetAttributeString("link", "")
	postPath := doc.GetAttributeString("postPath", "")
	pubDate := doc.GetAttributeString("pubDate", "")
	if pubDate == "" {
		pubDate = doc.GetAttributeString("datePublished", "")
	}
	dateModified := doc.GetAttributeString("dateModified",pubDate)
	draft := doc.GetAttributeBool("draft", false)
	channel := doc.GetAttributeString("channel", collection.Link)
	guid := doc.GetAttributeString("guid", link)
	status := "review"
	if draft || pubDate == ""{
		return fmt.Errorf("%s is set to draft or is missing pubDate %q", fName, pubDate)
	} else {
		status = "published" 
	}

	// When no description is provided the description is set with the body text
	if description == "" {
		description = innerHTML
	}
	if title == "" && description == "" {
		return fmt.Errorf("missing both title and description")
	}
	if postPath != "" {
		if link == "" {
			if cfg.BaseURL != "" {
				if strings.HasSuffix(postPath, ".md") {
					link = cfg.BaseURL + "/" + strings.TrimSuffix(postPath, ".md") + ".html"
				}  else {
					link = cfg.BaseURL + "/" + postPath
				}
			} else {
				return fmt.Errorf("missing baseUrl in antenna YAML, could not form link using postPath %q", postPath)
			}
		}
		// Write out an HTML page to the postPath, if Markdown doc, normalize .html
		htmlName := filepath.Join(cfg.Htdocs, postPath)
		if strings.HasSuffix(htmlName, ".md") {
			htmlName = strings.TrimSuffix(htmlName, ".md") + ".html"
		}
		dName := filepath.Dir(htmlName)
		if _, err := os.Stat(dName); err != nil {
			if err := os.MkdirAll(dName, 0775); err != nil {
				return err
			}
		}
		gen, err := NewGenerator(app.appName, cfg.BaseURL)
		if err != nil {
			return err
		}
		collection, err := cfg.GetCollection(cName)
		if err != nil {
			return err
		}
		if err := gen.LoadConfig(collection.Generator); err != nil {
			return err
		}
		if err := gen.WriteHtmlPage(htmlName, link, postPath, pubDate, innerHTML); err != nil {
			return err
		} 
	}
	// FIXME: Need to handle getting enclosures and publishing them to posts tree
	// NOTE: Insert/update item in collection
	// FIXME: need to populate the enclosures
	enclosures := []*Enclosure{}
	// FIXME: need to populate the Dublin Core extension
	dcExt := &ext.DublinCoreExtension{}
	updated := time.Now().Format(time.RFC3339)
	if dateModified != "" {
		d, err := time.Parse("2006-01-02", dateModified)
		if err != nil {
			return fmt.Errorf("failed to parse dateModified: %q, %s", dateModified, err)
		}
		updated = d.Format(time.RFC3339)
	}

	label := collection.Title
	authorsSrc := []byte{}
	if authors != nil {
		authorsSrc, err = json.Marshal(authors)
		if err != nil {
			return fmt.Errorf("failed to marshal author, %s", err)
		}
	}
	return updateItem(db, link, title, description, fmt.Sprintf("%s", authorsSrc),
		enclosures, guid, pubDate, dcExt, channel, status, updated, label, postPath)
}

func (app *AntennaApp) Unpost(cfgName string, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("expected a collection name and url in the collection")
	}
	cfg := &AppConfig{}
	if err := cfg.LoadConfig(cfgName); err != nil {
		return err
	}
	cName, link := strings.TrimSpace(args[0]), strings.TrimSpace(args[1])
	collection, err := cfg.GetCollection(cName)
	if err != nil {
		return err
	}
	dsn := collection.DbName
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	return removeItem(db, link)
}

// updateItem will perform an "upsert" to insert or update the row
func updateItem(db *sql.DB, link string, title string, description string, authors string,
	enclosures []*Enclosure, guid string, pubDate string, dcExt *ext.DublinCoreExtension,
	channel, status string, updated string, label string, postPath string) error {
	enclosuresSrc, err := json.Marshal(enclosures)
	if err != nil {
		return nil
	}
	dcExtSrc, err := json.Marshal(dcExt)
	if err != nil {
		return nil
	}
	_, err = db.Exec(SQLUpdateItem, link, title, description, authors,
		enclosuresSrc, guid, pubDate, dcExtSrc,
		channel, status, updated, label, postPath)
	if err != nil {
		return err
	}
	return nil
}

func removeItem(db *sql.DB, link string) error {
	_, err := db.Exec(SQLDeleteItemByLink, link)
	if err != nil {
		return err
	}
	return nil
}
