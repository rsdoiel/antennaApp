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
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	DefaultGeneratorYaml = `###
## Example YAML used to custom a collection's Generator YAML.
##
## Use this CSS file for the CSS link in the head element. To
## pull an other CSS, use a CSS @import statement. (uncomment line
## and update location)
#css: css/site.css
## You list of ES6 modules goes here, (uncomment lines, update locations)
#modules:
#  - modules/date_filter.js
##
## The next set of elements let you include HTML in the page
## area decribed.
##
header: |
  <header><!-- your custom header inner HTML goes here --></header>

nav: |
  <nav><!-- your custom nav innert HTML goes here --></nav>

topContent: |
  <!-- your custom HTML content before the section element goes here -->

bottomContent: |
  <!-- your custom HTML content after the section element goes here -->

footer: |
  <footer><!-- your custom footer innert HTML goes here --></footer>
`
)

// InitPageGenerator initializes a YAML configuration for the default page layout.
// It takes a filename and returns an error
func InitPageGenerator(pageName string) error {
	if _, err := os.Stat(pageName); err == nil {
		//FIXME: read in pageName and Make sure it has a valid structure
		return nil
	} else {
		// NOTE: Create a default page page Generator
		if err := os.WriteFile(pageName, []byte(DefaultGeneratorYaml), 0775); err != nil{
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
		return fmt.Errorf("expected a collection name and filepath for Markdown content")
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
	// Convert our document text to HTML
	innerHTML, err := doc.ToHTML()
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
	return nil
}

