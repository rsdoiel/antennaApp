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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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
			fmt.Errorf("failed to create %q, %s", pageName, err)
		}
	}
	return nil
}

// Page will add a CommonMark document as an HTML page based on postPath. It **will not** get added
// to an RSS feed or to the collection's db. It is designed to relieve the requirement of using Pandoc
// for a handful of page.
func (app *AntennaApp) Page(cfgName string, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("expected filepath for Markdown content")
	}
	cfg := &AppConfig{}
	if err := cfg.LoadConfig(cfgName); err != nil {
		return err
	}
	fName, oName := strings.TrimSpace(args[0]), ""
	if len(args) == 2 {
		oName = strings.TrimSpace(args[1])
	}
	src, err := os.ReadFile(fName)
	if err != nil {
		return err
	}
	doc := &CommonMark{}
	if err := doc.Parse(src); err != nil {
		return err
	}
	// NOTE: This is trusted content so I can support commonMarkDoc
	// processor extensions safely.
	if strings.Contains(doc.Text, "@include-text-block ") {
		doc.Text = IncludeTextBlock(doc.Text)
	}
	if strings.Contains(doc.Text, "@include-code-block ") {
		doc.Text = IncludeCodeBlock(doc.Text)
	}

	// Convert our document text to HTML
	// NOTE: Pages are allowed to have "unsafe" embedded HTML because they are
	// not reading from a feed, they are being read from your file system.
	innerHTML, err := doc.ToUnsafeHTML()
	if err != nil {
		return err
	}
	postPath := doc.GetAttributeString("postPath", fName)
	htmlName := filepath.Join(cfg.Htdocs, postPath)
	if oName != "" {
		htmlName = filepath.Join(cfg.Htdocs, oName)
	} else if strings.HasSuffix(htmlName, ".md") {
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
	if err := gen.LoadConfig(cfg.Generator); err != nil {
		return err
	}
	if err := gen.WriteHtmlPage(htmlName, "", postPath, "", innerHTML); err != nil {
		return err
	}
	// NOTE: I need to add the page to pages.db
	// NOTE: remove the page from pages table.
	dbName := "pages.db"
	db, err := sql.Open("sqlite", dbName)
	if err != nil {
		return fmt.Errorf("failed to open %s, %s", dbName, err)
	}
	defer db.Close()
	timestamp := time.Now().Format(time.RFC3339)
	if _, err := db.Exec(SQLUpdatePage, fName, oName, timestamp); err != nil {
		return fmt.Errorf("%s, %s", dbName, err)
	}
	return nil
}

// Unpage will remove a CommonMark document filePath
func (app *AntennaApp) Unpage(cfgName string, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("expected filepath for Markdown content")
	}
	cfg := &AppConfig{}
	if err := cfg.LoadConfig(cfgName); err != nil {
		return err
	}
	fName, oName := strings.TrimSpace(args[0]), strings.TrimSpace(args[0])
	if len(args) == 2 {
		oName = strings.TrimSpace(args[1])
	}
	// NOTE: remove the page from pages table.
	dbName := "pages.db"
	db, err := sql.Open("sqlite", dbName)
	if err != nil {
		return fmt.Errorf("failed to open %s, %s", dbName, err)
	}
	defer db.Close()

	if _, err := db.Exec(SQLDeletePageByPath, fName, oName); err != nil {
		return fmt.Errorf("%s, %s", dbName, err)
	}
	return nil
}

// Pages will list the pages in the pages collection
func (app *AntennaApp) Pages(cfgName string, args []string) error {
	cfg := &AppConfig{}
	if err := cfg.LoadConfig(cfgName); err != nil {
		return err
	}
	// NOTE: remove the page from pages table.
	dbName := "pages.db"
	db, err := sql.Open("sqlite", dbName)
	if err != nil {
		return fmt.Errorf("failed to open %s, %s", dbName, err)
	}
	defer db.Close()

	rows, err := db.Query(SQLDisplayPage)
	if err != nil {
		return fmt.Errorf("%s, %s", dbName, err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			iName   string
			oName   string
			updated string
		)
		if err := rows.Scan(&iName, &oName, &updated); err != nil {
			fmt.Fprintf(os.Stderr, "failed to read row, %s\n", err)
			continue
		}
		fmt.Printf("%s\t%s\t%s\n", iName, oName, updated)
	}
	return nil
}
