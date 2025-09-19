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

// Page will add a CommonMark document as an HTML page based on postPath. It **will not** get added
// to an RSS feed or to the collection's db. It is designed to relieve the requirement of using Pandoc
// for a handful of page.
func (app *AntennaApp) Page(cfgName string, args []string) error {
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
	postPath := doc.GetAttributeString("postPath", "")

	if postPath != "" {
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
		if err := gen.LoadConfig(collection.Generator); err != nil {
			return err
		}
		if err := gen.WriteHtmlPage(htmlName, "", postPath, "", innerHTML); err != nil {
			return err
		} 
	}
	return nil
}

