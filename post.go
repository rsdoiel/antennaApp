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
	"strconv"
	"strings"

	// 3rd Party
	_ "github.com/glebarez/go-sqlite"
)


func saveMarkdown(fName string, doc *CommonMark) error {
	backupName := strings.TrimSuffix(fName, ".md") + ".bak"
	if _, err := os.Stat(backupName); err == nil {
		if err := os.Remove(backupName); err != nil {
			return fmt.Errorf("failed to back %q as %q, %s", fName, backupName, err)
		}
	}
	if err := os.Rename(fName, backupName); err != nil {
		return err
	}
	txt := doc.String()
	if err := os.WriteFile(fName, []byte(txt), 0666); err != nil {
		return err
	}
	return nil
}

// Post will add a CommonMark document as a feed item and if the postPath and link
// are provided it will convert the CommonMark document to HTML and save it in the postPath.
func (app *AntennaApp) Post(cfgName string, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("expected a Markdown filename or collection name and Markdown filename")
	}
	cfg := &AppConfig{}
	if err := cfg.LoadConfig(cfgName); err != nil {
		return err
	}
	cName, fName := "pages.md", ""
	if len(args) == 1 {
		fName = strings.TrimSpace(args[0])
	} else {
		cName, fName = strings.TrimSpace(args[0]), strings.TrimSpace(args[1])
	}
	return cfg.Post(cName, fName)
}

// This lists published posts
func (app *AntennaApp) Posts(cfgName string, args []string) error {
	cfg := &AppConfig{}
	if err := cfg.LoadConfig(cfgName); err != nil {
		return err
	}
	cName := "pages.md"
	if len(args) > 0 {
		cName = strings.TrimSpace(args[0])
	}
	return cfg.Posts(cName, args)
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
	return cfg.Unpost(cName, link)
}

// RssPosts, gernate RSS to stdout for posts
func (app *AntennaApp) RssPosts(cfgName string, args []string) error {
	cfg := &AppConfig{}
	if err := cfg.LoadConfig(cfgName); err != nil {
		return err
	}
	cName := "pages.md"
	rssFeed := ""
	if len(args) > 0 {
		cName = strings.TrimSpace(args[0])
	} else {
		return fmt.Errorf("missing collection to use for RSS feed")
	}
	if len(args) > 1 {
		rssFeed = strings.TrimSpace(args[1])
	} else {
		return fmt.Errorf("missing RSS filename to generate")
	}
	feedLink := fmt.Sprintf("%s/%s", cfg.BaseURL, rssFeed)
	out, err := os.Create(rssFeed)
	if err != nil {
		return err
	}
	defer out.Close()

	collection, err := cfg.GetCollection(cName)
	if err != nil {
		return fmt.Errorf("%s, %s", cName, err)
	}
	dsn := collection.DbName
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	// NOTES: posts action supports three different SQL statements
	var (
		sqlStmt string
	)

	gen, err := NewGenerator(app.appName, cfg.BaseURL)
	if err != nil {
		return err
	}

	switch {
	case len(args) == 4:
		fromDate, toDate := args[2], args[3]
		sqlStmt = fmt.Sprintf(SQLRssDateRangePosts, fromDate, toDate)
	case len(args) == 3:
		count, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("%q, %s", args[2], err)
		}
		sqlStmt = fmt.Sprintf(SQLRssRecentPosts, count)
		if err != nil {
			return fmt.Errorf("%s\n%s, %s", SQLRssRecentPosts, dsn, err)
		}
	default:
		sqlStmt = SQLRssPosts
	}
	return gen.WriteCustomRSS(out, db, sqlStmt, feedLink, app.appName, collection)
}
