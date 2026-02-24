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
	"bufio"
	"fmt"
	"os"
	"strings"

	// My packages
	"github.com/rsdoiel/termlib"

	// 3rd Party
	//_ "github.com/glebarez/go-sqlite"
)

var (
	// Help and Menus for related actions
	collectionsMenuHelp = fmt.Sprintf(`
Collections Menu Help

   (optional parameters can be completed using resulting prompts)

%sa%sdd %s[COLLECTION_FILE [NAME] [DESCRIPTION]]%s
: Add the feed collection name by COLLECTION_FILE to your Antenna configuration.
A COLLECTION_FILE is a Markdown document containing one or more links in a list. You 
can include a short name that will be displayed when the HTML was generated. You may
also supply a DESCRIPTION associated with the collection. These can also be set in
the Front Matter of the Markdown document.

%sd%sel %s[COLLECTION_FILE]%s
: Remove a collection from the Antenna configuration.

%sl%sist
: List the collections defined in the Antenna configuration.

%sha%srvest %s[COLLECTION_NAME]%s
: The harvest retrieves feed content. If COLLECTION_NAME is provided then only the 
the single collection will be harvested otherwise all collections defined in your
Antenna YAML configuration are harvested.

%sg%senerate %s[COLLECTION_NAME]%s
: This process the collections rendering HTML pages and RSS 2.0 feeds for each collection.
If the collection name is provided then only that HTML page will be generated.

%sr%sss %s[COLLECTION_NAME [RSS_FILENAME] [COUNT | FROM_DATE TO_DATE]]%s
: Generate an RSS feed from posts. The optional parameters are applied
like the posts action.

%ss%sitemap
: This will generate a set of sitemap files for pages and posts found through the
{app_name}.yaml file. (e.g. sitemap_index.xml, sitemap_1.xml, sitemap_2.xml, ...)

%sap%sply %s[THEME_PATH [YAML_FILE_PATH]]%s
: This will apply the content THEME_PATH and update the YAML generator file described
by YAML_FILE_PATH. If YAML_FILE_PATH is not provided then that YAML generator file
will be replaced by the theme.

%sp%sreview
: Let's your preview the rendered your Antenna instance as a localhost website using
your favorite web browser.

`, 
	Yellow+Bold, Reset, Cyan, Reset,
	Yellow+Bold, Reset, Cyan, Reset,
	Yellow+Bold, Reset,
	Yellow+Bold, Reset, Cyan, Reset,
	Yellow+Bold, Reset, Cyan, Reset,
	Yellow+Bold, Reset, Cyan, Reset,
	Yellow+Bold, Reset,
	Yellow+Bold, Reset, Cyan, Reset,
	Yellow+Bold, Reset)

	collectionsMenu = fmt.Sprintf(`
Collections Menu

	%sa%sdd %s[COLLECTION_FILE [NAME] [DESCRIPTION]]%s

	%sd%sel %s[COLLECTION_FILE]%s

	%sl%sist

	%sha%srvest %s[COLLECTION_NAME]%s

	%sg%senerate %s[COLLECTION_NAME]%s

	%sr%sss %s[COLLECTION_NAME [RSS_FILENAME] [COUNT | FROM_DATE TO_DATE]]%s

	%ss%sitemap

	%sap%sply %s[THEME_PATH [YAML_FILE_PATH]]%s

	%sp%sreview

`,	
	Yellow+Bold, Reset, Cyan, Reset,
	Yellow+Bold, Reset, Cyan, Reset,
	Yellow+Bold, Reset,
	Yellow+Bold, Reset, Cyan, Reset,
	Yellow+Bold, Reset, Cyan, Reset,
	Yellow+Bold, Reset, Cyan, Reset,
	Yellow+Bold, Reset,
	Yellow+Bold, Reset, Cyan, Reset,
	Yellow+Bold, Reset)

 )


/**
 * Collections methods
 */

// helpCurateCollections explains how the options in the collection menu
func helpCurateCollections(scanner *bufio.Scanner) {
	term.Clear()
	defer term.Clear()
	term.ResetStyle()
	term.Printf(collectionsMenuHelp)
	term.Printf("\n(%sq%suit, to exit help): ", Green, Reset)
	term.Refresh()
	scanner.Scan()
}

// curateCollections provides the interaction loop for curating collections.
func curateCollections(scanner *bufio.Scanner, cfgName string, cfg *AppConfig) error {
	term.Clear()
	defer term.Clear()
	for quit := false; quit == false; {
		term.Move(1, 1)
		term.ClrToEOL()
		// - list next actions
		term.Printf(collectionsMenu)
		term.Printf("\n(%sh%selp or %sq%suit): ", Green+Bold, Reset, Green+Bold, Reset)
		term.ClrToEOL()
		term.Refresh()
		
		// Read entry
		if !scanner.Scan() {
			continue
		}
		answer, options, err := parseAnswer(scanner.Text())
		if err != nil {
			displayErrorStatus("%s", err)
		}
		answer = strings.ToLower(answer)
		switch {
		case strings.HasPrefix(answer, "ap"):
			// apply a theme to a collection
			if err := applyThemeFiles(scanner, options, cfgName, cfg); err != nil {
				displayErrorStatus("%s", err)
				continue
			}
		case strings.HasPrefix(answer,"a"):
			// add a collection(s)
			if err := addCollection(scanner, options, cfgName, cfg); err != nil {
				displayErrorStatus("%s", err)
				continue
			}	
			if err := cfg.LoadConfig(cfgName); err != nil {
				displayErrorStatus("%s", err)
			}
		case strings.HasPrefix(answer, "d"):
			// del a collection(s)
			if err := deleteCollection(scanner, options, cfgName, cfg); err != nil {
				displayErrorStatus("%s", err)
				continue
			}
			if err := cfg.LoadConfig(cfgName); err != nil {
				displayErrorStatus("%s", err)
			}
		case strings.HasPrefix(answer, "l"):
			// list collection names
			if err := listCollectionNames(scanner, options, cfgName, cfg); err != nil {
				displayErrorStatus("%s", err)
				continue
			}
		case strings.HasPrefix(answer, "ha"):
			// harvest a collection(s)
			if err := harvestCollection(scanner, options, cfgName, cfg); err != nil {
				displayErrorStatus("%s", err)
				continue
			}
		case strings.HasPrefix(answer, "g"):
			// Generate pages, posts, rss and sitemaps
			if err := generateCollection(scanner, options, cfgName, cfg); err != nil {
				displayErrorStatus("%s", err)
				continue	
			}
		case strings.HasPrefix(answer, "r"):
			// Generate RSS files
			if err := generateRSSFiles(scanner, options, cfgName, cfg); err != nil {
				displayErrorStatus("%s", err)
				continue
			}
		case strings.HasPrefix(answer, "s"):
			// Generate sitemaps
			if err := generateSitemapFiles(scanner, options, cfgName, cfg); err != nil {
				displayErrorStatus("%s", err)
				continue
			}
		case strings.HasPrefix(answer, "h"):
			helpCurateCollections(scanner)
			continue
		case strings.HasPrefix(answer, "q"):
			quit = true
		}
		term.Clear()
	}
	return nil
}

// listCollectionNames will output a list of collection names
func listCollectionNames(scanner *bufio.Scanner, options []string, cfgName string, cfg *AppConfig) error {
	// Clear the screen
	term.Clear()
	defer term.Clear()
	term.ResetStyle()
	args := []string{}
	pageSize := int((term.GetTerminalHeight() - 5) / 2)
	curPos := 0
	args = append(args, fmt.Sprintf("%d", pageSize))
	names, err := cfg.ListCollectionFiles(cfgName)
	if err != nil {
		return err
	}
	for quit := false; !quit; {
		term.Move(1, 1)
		term.ClrToEOL()
		term.Printf("Collections\n\n")
		tot := len(names)
		for i := curPos; i < tot && i < (curPos+pageSize); i++ {
			name := names[i]
			term.ClrToEOL()
			term.Printf("%4d. %s%s%s\n",
				i+1, termlib.Bold+termlib.Italic, name, termlib.Reset)
		}
		// Display prompt
		term.ResetStyle()
		term.Printf("\n(%d/%d, %s+N%s, %s-N%s, COLLECTION_NAME, %sh%selp  or %sq%suit): ",
			curPos + 1, tot,
			Green+Bold, Reset,
			Green+Bold, Reset,
			Green+Bold, Reset,
			Green+Bold, Reset)
		term.ClrToEOL()
		term.Refresh()
		if !scanner.Scan() {
			continue
		}
		answer, _, err := parseAnswer(scanner.Text())
		answer = strings.ToLower(answer)
		switch {
		case answer == "":
			curPos = normalizePos(curPos+pageSize, pageSize, tot)
		case strings.HasPrefix(answer, "+"):
			curPos, err = pageTo(answer, curPos, pageSize, tot)
			if err != nil { 
				displayErrorStatus("%s", err)
				continue
			}
		case strings.HasPrefix(answer, "-"):
			curPos, err = pageTo(answer, curPos, pageSize, tot)
			if err != nil { 
				displayErrorStatus("%s", err)
				continue
			}
		case strings.HasPrefix(answer, "q"):
			quit = true
		default:
			// If the answer is a number, go to item number
			if val, err := extractInt(answer); err == nil {
				curPos = normalizePos(val-1, pageSize, tot)
			} else {
				displayErrorStatus("%q, unknown command", answer)
				continue
			}
		}
		term.Clear()
	}
	return nil
}


/**
 * Collections methods
 */
 
// addCollections provides the prompts to add a new collection
func addCollection(scanner *bufio.Scanner, options []string, cfgName string, cfg *AppConfig) error {
	var (
		cName string
		title string
		description string
	)
	if len(options) < 1 {
		term.Printf("Enter a collection name: ")
		scanner.Scan()
		cName = scanner.Text()
	} else {
		cName = options[0]
	}
	if cName == "" {
		return fmt.Errorf("Missing collection name")
	}
	
	if _, err := os.Stat(cName); os.IsNotExist(err) {
		if len(options) < 2 {
			term.Printf("Enter title: ")
			scanner.Scan()
			title = scanner.Text()
		}
		if len(options) < 3 {
			term.Printf("Enter description: ")
			scanner.Scan()
			description = scanner.Text()			
		}
		text := fmt.Sprintf(`---
title: %q
description: %q
---

# %s

%s

(enter any feeds you want to aggregate here as a Markdown list of links)

`, title, description, title, description)
		if err := os.WriteFile(cName, []byte(text), 0664); err != nil {
			return err
		}
	}
	return cfg.AddCollection(cfgName, cName)	
}

// deleteCollection provides the prompts to delete a collection	scanner := bufio.NewScanner(os.Stdin)
func deleteCollection(scanner *bufio.Scanner, options []string, cfgName string, cfg *AppConfig) error {
	if cfg.Collections == nil {
		return fmt.Errorf("no collections to delete")
	}	
	var cName string
	if len(options) < 1 {
		term.Printf("Enter a collection name: ")
		scanner.Scan()
		cName = scanner.Text()
	} else {
		cName = options[0]
	}
	// Find the collection name to delete.
	for _, col := range cfg.Collections {
		if col.File == cName {
				term.Printf("Remove %s? yes/NO ", cName)
				scanner.Scan()
				answer, _, _ := parseAnswer(scanner.Text())
				if (answer != "yes" && answer != "y") {
					return fmt.Errorf("delete %s cancelled", cName)
				}
				return cfg.DelCollection(cfgName, cName)	
		}
	}
	if cName == "" {
		return fmt.Errorf("Missing collection name")
	}
	return fmt.Errorf("%q not found", cName)
}

// harvestCollection will retrieve and aggregate collection items
func harvestCollection(scanner *bufio.Scanner, options []string, cfgName string, cfg *AppConfig) error {
	term.Clear()
	defer term.Clear()
	term.ResetStyle()
	args := []string{}
	for _, cName := range options {
		if val, err := extractInt(cName); err == nil {
			// Figure out the cName for collection
			val = val - 1 // Adjust for zero based array
			if val >= 0 && val < len(cfg.Collections) {
				cName = cfg.Collections[val].File
			}
		}
		args = append(args, cName)
	}

	if len(args) == 0 {
		for _, col := range cfg.Collections {
			args = append(args, col.File)
		}
	}
	term.Clear()
	for _, cName := range args {
		col, err := cfg.GetCollection(cName)
		if err != nil {
			return err
		}
		if col == nil {
			displayErrorStatus("warning could not retrieve %q, skipping\n", cName)
			continue
		}
		term.Printf("Harvesting %s\n", cName)
		// Harvest the collection
		if err := col.Harvest(os.Stdout, os.Stderr, cfg.UserAgent); err != nil {
			displayErrorStatus("warning %s: %s\n", col.File, err)
		}
	}
	term.Printf("\n\tPress enter to return to collections menu.\n")
	term.Refresh()
	scanner.Scan()
	scanner.Text()
	return nil
}

// generateCollection will generate pages and posts
func generateCollection(scanner *bufio.Scanner, options []string, cfgName string, cfg *AppConfig) error {
	term.Clear()
	defer term.Clear()
	term.ResetStyle()
	args := []string{}
	for _, cName := range options {
		if val, err := extractInt(cName); err == nil {
			// Figure out the cName for collection
			val = val - 1 // Adjust for zero based array
			if val >= 0 && val < len(cfg.Collections) {
				cName = cfg.Collections[val].File
			}
		}
		args = append(args, cName)
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
			displayErrorStatus("warning could not retrieve %q, skipping\n", cName)
			continue
		}
		term.Printf("Generating %s\n", cName)
		// Generate the aggregated page
		if err := col.Generate(os.Stdout, os.Stderr, "curate", cfg); err != nil {
			displayErrorStatus("warning %s: %s\n", col.File, err)
		}
	}
	term.Printf("\n\tPress enter to return to collections menu.\n")
	term.Refresh()
	scanner.Scan()
	scanner.Text()
	return nil
}

// generateRSSFiles will generate RSS files for collections all collections
func generateRSSFiles(scanner *bufio.Scanner, options []string, cfgName string, cfg *AppConfig) error {
	return fmt.Errorf("generateRSSFiles() not implemented yet.")
}

// generateSitemapFiles will generate Sitemap XML files for all collections
func generateSitemapFiles(scanner *bufio.Scanner, options []string, cfgName string, cfg *AppConfig) error {
	return fmt.Errorf("generateSitemapFiles() not implemented yet.")
}

// appleThemeFiles apply a theme directory to a given collection
func applyThemeFiles(scanner *bufio.Scanner, options []string, cfgName string, cfg *AppConfig) error {
	return fmt.Errorf("applyThemeFiles() not implemented yet.")
}
