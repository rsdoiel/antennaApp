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
	"io"
	"os"
	"path/filepath"
	"strings"

	// 3rd Party Packages
	"gopkg.in/yaml.v3"
)

// Generator supports the generation of HTML pages from a YAML configuration
type Generator struct {
	// AppName holds the name of application running the generator
	AppName string `json:"appName,omitempty" yaml:"appName,omitempty"`

	// BaseURL used to form the feed Link
	BaseURL string `json:"baseUrl,omitempty" yaml:"baseUrl,omitempty"`

	// Version holds the version of the genliction
	// used when generating the "generator" metadata
	Version string `json:"version,omitempty" yaml:"version,omitempty"`

	// DbName holds the path to the SQLite3 database
	DBName string `json:"dbName,omitempty" yaml:"dbName,omitempty"`

	// Title if this is set the title will be included
	// when generating the markdown of saved items
	Title string `json:"title,omitempty" yaml:"title,omitempty"`

	// Description, included as metadata in head element
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// Link points to the RSS feed associated with the page.
	Link string `json:"link,omitempty" yaml:"link,omitempty"`

	// CSS is the path to a CSS file
	CSS string `json:"css,omitempty" yaml:"css,omitempty"`

	// Modules is a list for ES6 diles
	Modules []string `json:"modules,omitempty" yaml:"modules,omitempty"`

	// Header hold the HTML markdup of the Header element. If not included
	// then it will be generated using the Title and timestamp
	Header string `json:"header,omitempty" yaml:"header,omitempty"`

	// Nav holds the HTML markup for navigation
	Nav string `json:"nav,omitempty" yaml:"nav,omitempty"`

	// TopContent holds HTML that comes before the selecton element
	// containing articles
	TopContent string `json:"topContent,omitempty" yaml:"topContent,omitempty"`

	// BottomContent holds HTML that comes before the selecton element
	// containing articles
	BottomContent string `json:"bottomContent,omitempty" yaml:"bottomContent,omitempty"`

	// Footer holds the HTML markup for the footer
	Footer string `json:"footer,omitempty" yaml:"footer,omitempty"`

	out  io.Writer
	eout io.Writer
}

// NewGenerator initialized a new Generator struct
func NewGenerator(appName string, BaseURL string) (*Generator, error) {
	gen := new(Generator)
	gen.AppName = appName
	gen.Version = Version
	gen.BaseURL = BaseURL
	gen.out = os.Stdout
	gen.eout = os.Stderr
	return gen, nil
}

func getDsnAndCfgName(args []string) (string, string) {
	dsn := args[0]
	if len(args) == 2 {
		return args[0], args[1]
	}
	// Figure out if we have a YAML config or not
	cfgName := strings.TrimSuffix(dsn, ".db") + ".yaml"
	if _, err := os.Stat(cfgName); err != nil {
		return dsn, ""
	}
	return dsn, cfgName
}

// LoadConfig read in the generator configuration (not AppConfig)
// and map the settings into the Generator object.
func (gen *Generator) LoadConfig(cfgName string) error {
	src, err := os.ReadFile(cfgName)
	if err != nil {
		return err
	}
	obj := Generator{}
	if err := yaml.Unmarshal(src, &obj); err != nil {
		return err
	}
	// Pull in the configuration elements
	if obj.AppName != "" {
		gen.AppName = obj.AppName
	}
	if obj.BaseURL != "" {
		gen.BaseURL = obj.BaseURL
	}
	if obj.Version != "" {
		gen.Version = obj.Version
	}
	if obj.Title != "" {
		gen.Title = obj.Title
	}
	if obj.Description != "" {
		gen.Description = obj.Description
	}
	if obj.CSS != "" {
		gen.CSS = obj.CSS
	}
	if obj.Modules != nil && len(obj.Modules) > 0 {
		gen.Modules = obj.Modules[:]
	}
	if obj.Header != "" {
		gen.Header = obj.Header
	}
	if obj.Nav != "" {
		gen.Nav = obj.Nav
	}
	if obj.TopContent != "" {
		gen.TopContent = obj.TopContent
	}
	if obj.BottomContent != "" {
		gen.BottomContent = obj.BottomContent
	}
	if obj.Footer != "" {
		gen.Footer = obj.Footer
	}
	return nil
}

func (app AntennaApp) Generate(out io.Writer, eout io.Writer, cfgName string, args []string) error {
	cfg := &AppConfig{}
	if err := cfg.LoadConfig(cfgName); err != nil {
		return err
	}
	if len(args) == 0 {
		for _, col := range cfg.Collections {
			args = append(args, col.File)
		}
	}
	for _, cName := range args {
		col, err := cfg.GetCollection(cName)
		if err != nil {
			return err
		}
        if col == nil {
            fmt.Fprintf(eout, "warning could not retrieve %q, skipping\n", cName)
            continue
        }
		// Generate the aggregated page
		if err := col.Generate(out, eout, app.appName, cfg); err != nil {
			fmt.Fprintf(eout, "warning %s: %s\n", col.File, err)
		}
	}
	return nil
}

func (collection *Collection) ApplyFilters(db *sql.DB) error {
	if len(collection.Filters) == 0 {
		return nil
	}
	for _, stmt := range collection.Filters {
        if strings.TrimSpace(stmt) != "" {
            _, err := db.Exec(stmt)
            if err != nil {
                return fmt.Errorf("%s\nstmt: %s", err, stmt)
            }
        }
	}
	return nil
}

func (collection *Collection) Generate(out io.Writer, eout io.Writer, appName string, cfg *AppConfig) error {
	gen, err := NewGenerator(appName, cfg.BaseURL)
	if err != nil {
		return err
	}
	if collection.Generator == "" {
		collection.Generator = cfg.Generator
	}
	if _, err := os.Stat(collection.Generator); err == nil {
		src, err := os.ReadFile(collection.Generator)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(src, &gen); err != nil {
			return err
		}
	} else {
		if err := yaml.Unmarshal([]byte(DefaultGeneratorYaml), &gen); err != nil {
			return err
		}
	}
	return gen.Generate(eout, appName, cfg, collection)
}

func (gen *Generator) Generate(eout io.Writer, appName string, cfg *AppConfig, collection *Collection) error {
	// Open DB so we have a place to write data.
	dsn := collection.DbName
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

    // Run the collection filter to determine which items to publish
	if err := collection.ApplyFilters(db); err != nil {
		return err
	}

	// figure out the name and path to write the HTML file to.
	bName := filepath.Base(collection.File)
	xName := filepath.Ext(bName)
	htmlName := filepath.Join(cfg.Htdocs, strings.TrimSuffix(bName, xName)+".html")
	rssName := filepath.Join(cfg.Htdocs, strings.TrimSuffix(bName, xName)+".xml")
	opmlName := filepath.Join(cfg.Htdocs, strings.TrimSuffix(bName, xName)+".opml")

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

	// Write out HTML page
	if err := gen.WriteHTML(out, db, appName, collection); err != nil {
		return err
	}
	out.Close()
 
    // clear existing page
	if _, err := os.Stat(rssName); err == nil {
		if err := os.Remove(rssName); err != nil {
			return nil
		}
	}

    // Create the RSS file
	out, err = os.Create(rssName)
	if err != nil {
		return err
	}

    // Write out RSS page
	if err := gen.WriteRSS(out, db, appName, collection); err != nil {
		return err
	}
	out.Close()


    // clear existing page
	if _, err := os.Stat(opmlName); err == nil {
		if err := os.Remove(opmlName); err != nil {
			return nil
		}
	}

    // Create the OPML file
	out, err = os.Create(opmlName)
	if err != nil {
		return err
	}
	
	// Write out OPML page
	if err := gen.WriteOPML(out, db, appName, collection); err != nil {
		return err
	}
	defer out.Close()
	return nil
}
