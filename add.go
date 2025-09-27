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

	// 3rd Party
	_ "github.com/glebarez/go-sqlite"
)


func (app *AntennaApp) Add(cfgName string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing collection name")
	}
	// create a cfg object
	cfg := &AppConfig{}
	// Load configuration
	if err := cfg.LoadConfig(cfgName); err != nil {
		return err
	}
	// Get the collection name
	cName := args[0]
	src, err := os.ReadFile(cName)
	if err != nil {
		return err
	}
	// Parse the Markdown collection of links and make sure it
	// makes sense.
	doc := &CommonMark{}
	if err := doc.Parse(src); err != nil {
		return err
	}
	// Now process the Markdown and pull out the list of links and label and description.
	if _, err := doc.GetLinks(); err != nil {
		return err
	}
	// If we get this parse we can update the configuration and include this link.
	if cfg.Collections == nil {
		cfg.Collections = []*Collection{}
	}
	collection := &Collection{
		File: cName,
	}
	if doc.FrontMatter != nil {
		collection.UpdateFrontMatter(doc.FrontMatter, cfg)
	}
	if collection.Title == "" {
		collection.Title = cName
	}
	// Add some default SQL filters
	if len(collection.Filters) == 0 {
		collection.Filters = []string{
			SQLUpdateStatusToReview,
			SQLSetStatusPublishedForRecentlyPublished,
		}
	}

	// Make sure we have a SQLite3 database name
	bName := filepath.Base(cName)
	xName := filepath.Ext(bName)
	if collection.DbName == "" {
		collection.DbName = strings.TrimSuffix(cName, xName) + ".db"
	}
	if collection.Link == "" {
		collection.Link = strings.TrimSuffix(cName, xName) + ".xml"
	}
	// Make sure we have a reasonable Generator YAML filename
	if collection.Generator == "" {
		collection.Generator = cfg.Generator
	}
	// Do I need to "add" the collection or replace the collection?
	appendCollection := true
	for i, col := range cfg.Collections {
		if filepath.Base(col.File) == filepath.Base(collection.File) {
			// update collection
			cfg.Collections[i] = collection
			appendCollection = false
		}
	}
	if appendCollection {
		// We're really adding the new collection
		cfg.Collections = append(cfg.Collections, collection)
	}
	// Create SQLite3 database for collection if not exists
	if err := setupDatabase(cName, collection.DbName); err != nil {
		return err
	}
	// Create generator YAML if not exists
	if _, err := os.Stat(collection.Generator); os.IsNotExist(err) {
		if err := InitPageGenerator(collection.Generator); err != nil {
			return err
		}
	}
	// Save all the updates
	if err := cfg.SaveConfig(cfgName); err != nil {
		return err
	}
	return nil
}

// setupDatabase checks to see if anything needs to be setup (or fixed) for AntennaApp to run.
func setupDatabase(cName string, dbName string) error {
	// Check to see if we have an existing SQLite3 file
	if _, err := os.Stat(dbName); os.IsNotExist(err) {
		stmt := fmt.Sprintf(SQLCreateTables, cName, time.Now().Format("2006-01-02"))
		dsn := dbName
		db, err := sql.Open("sqlite", dsn)
		if err != nil {
			return err
		}
		defer db.Close()
		if db == nil {
			return fmt.Errorf("%s opened and returned nil", dbName)
		}
		_, err = db.Exec(stmt)
		if err != nil {
			return fmt.Errorf("%s\nstmt: %s", err, stmt)
		}
	}
	return nil
}

// Del removes one or more collections from your Antenna instance.
// NOTE: It does not remove generated files, e.g. SQLite3 database,
// YAML, HTML or RSS files.
func (app *AntennaApp) Del(cfgName string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing collection name")
	}

	// create a cfg object
	cfg := &AppConfig{}
	// Load configuration
	if err := cfg.LoadConfig(cfgName); err != nil {
		return err
	}
	for _, cName := range args {
		i := cfg.CollectionIndex(cName)
		if i > -1 {
			cfg.Collections = append(cfg.Collections[:i], cfg.Collections[i+1:]...)	
		}
	}
	// Save all the updates
	if err := cfg.SaveConfig(cfgName); err != nil {
		return err
	}
	return nil
}
