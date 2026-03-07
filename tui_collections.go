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
	"slices"
	"strconv"
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

%sl%sist
: List the collections defined in the configuration.

%sa%sdd %s[COLLECTION_FILE [NAME] [DESCRIPTION]]%s
: Add the feed collection name by COLLECTION_FILE to your configuration.
A COLLECTION_FILE is a Markdown document containing one or more links in a list. You 
can include a short name that will be displayed when the HTML was generated. You may
also supply a DESCRIPTION associated with the collection. These can also be set in
the Front Matter of the Markdown document.

%sd%sel %s[COLLECTION_FILE]%s
: Remove a collection from the configuration.

%sg%senerate %s[COLLECTION_NAME]%s
: This process the collections rendering HTML pages and RSS 2.0 feeds for each collection.
If the collection name is provided then only that HTML page will be generated.

%sr%sss %s[COLLECTION_NAME [RSS_FILENAME] [COUNT | FROM_DATE TO_DATE]]%s
: Generate an RSS feed from posts. The optional parameters are applied
like the posts action.

%ss%sitemap
: This will generate a set of sitemap files for pages and posts found through the
{app_name}.yaml file. (e.g. sitemap_index.xml, sitemap_1.xml, sitemap_2.xml, ...)

%sharvest%s %s[COLLECTION_NAME]%s
: The harvest retrieves feed content. If COLLECTION_NAME is provided then only the 
the single collection will be harvested otherwise all collections defined in your
Antenna YAML configuration are harvested.

%sthemes%s
: This will list the themes an allow you to apply them

%sp%sreview
: Let's your preview the rendered your instance as a localhost website using
your favorite web browser.

`, 
	Yellow+Bold, Reset,
	Yellow+Bold, Reset, Cyan, Reset,
	Yellow+Bold, Reset, Cyan, Reset,
	Yellow+Bold, Reset, Cyan, Reset,
	Yellow+Bold, Reset, Cyan, Reset,
	Yellow+Bold, Reset,
	Yellow+Bold, Reset, Cyan, Reset,
	Yellow+Bold, Reset,
	Yellow+Bold, Reset)

	collectionsMenu = fmt.Sprintf(`
Collections Menu

    %sl%sist
        (list collections)

    %sa%sdd %s[COLLECTION_FILE [NAME] [DESCRIPTION]]%s
        (add a collection)

    %sd%sel %s[COLLECTION_FILE]%s
        (delete a collection)

    %sg%senerate %s[COLLECTION_NAME]%s
        (generate html)

    %sr%sss %s[COLLECTION_NAME [RSS_FILENAME] [COUNT | FROM_DATE TO_DATE]]%s
        (generate RSS file)

    %ss%sitemap
        (generate site maps)

    %sharvest%s %s[COLLECTION_NAME]%s
        (retrieve feeds)

    %sthemes%s
        (list themes)

    %sp%sreview
        (preview website)

`,	
	Yellow+Bold, Reset,
	Yellow+Bold, Reset, Cyan, Reset,
	Yellow+Bold, Reset, Cyan, Reset,
	Yellow+Bold, Reset, Cyan, Reset,
	Yellow+Bold, Reset, Cyan, Reset,
	Yellow+Bold, Reset,
	Yellow+Bold, Reset, Cyan, Reset,
	Yellow+Bold, Reset, 
	Yellow+Bold, Reset)


 	listCollectionNamesHelpMenu = `
List Collections menu

+N, -N
: Move N forward or backward in list (setting the Nth item to the
top of the window

N, COLLECTION_NAME
: Open a collection by number or name

h
: Display this help page

q
: Quit this menu

`
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
		case strings.HasPrefix(answer, "themes"):
			// apply a theme to a collection
			if err := listThemes(scanner, options, cfgName, cfg); err != nil {
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
			if err := generateRssFiles(scanner, options, cfgName, cfg); err != nil {
				displayErrorStatus("%s", err)
				continue
			}
		case strings.HasPrefix(answer, "s"):
			// Generate sitemaps
			if err := generateSitemapFiles(scanner, options, cfgName, cfg); err != nil {
				displayErrorStatus("%s", err)
				continue
			}
		case strings.HasPrefix(answer, "p"):
			// Run web server until "q" pressed
			if err := runPreview(scanner, options, cfgName, cfg); err != nil {
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

// helpListCollectionNames explains the options in the collection names menu
func helpListCollectionNames(scanner *bufio.Scanner) {
	term.Clear()
	defer term.Clear()
	term.ResetStyle()
	term.Printf(listCollectionNamesHelpMenu)
	term.Printf("\n(%sq%suit, to exit help): ", Green, Reset)
	term.Refresh()
	scanner.Scan()
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
		term.Printf("Enter a collection name to curate\n\n")
		tot := len(names)
		for i := curPos; i < tot && i < (curPos+pageSize); i++ {
			name := names[i]
			term.ClrToEOL()
			term.Printf("%4d. %s%s%s\n",
				i+1, termlib.Bold+termlib.Italic, name, termlib.Reset)
		}
		// Display prompt
		term.ResetStyle()
		term.Printf("\n(%d/%d, %s+N%s, %s-N%s, COLLECTION_NAME_OR_NUMBER, %sh%selp  or %sq%suit): ",
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
		case answer == "+":
			// Page forward
			curPos = normalizePos(curPos+pageSize, pageSize, tot)
		case answer == "-":
			// Page back
			curPos = normalizePos(curPos-pageSize, pageSize, tot)
		case strings.HasPrefix(answer, "+"):
			// Page forward to +N
			curPos, err = pageTo(answer, curPos, pageSize, tot)
			if err != nil { 
				displayErrorStatus("%s", err)
				continue
			}
		case strings.HasPrefix(answer, "-"):
			// Page back to +N
			curPos, err = pageTo(answer, curPos, pageSize, tot)
			if err != nil { 
				displayErrorStatus("%s", err)
				continue
			}
		case strings.HasPrefix(answer, "h"):
			helpListCollectionNames(scanner)
			displayErrorStatus("%s not implemented yet", answer)
			continue
		case strings.HasPrefix(answer, "q"):
			quit = true
		default:
			// Check it collection name or number was typed and
			// reconcile
			if val, err := strconv.Atoi(answer); err == nil {
				val -= 1
				if val >= 0 && val < len(names) {
					answer = names[val]
				} else {
					displayErrorStatus("%q does not match a collection name", answer)
					continue
				}
			}
			if slices.Contains(names, answer) {
			// FIXME: Open curate collection name provided
				if err := curateCollection(scanner, answer, cfgName, cfg); err != nil {
					displayErrorStatus("%s, failed to open collection", err)
					continue
				}
			} else {
					displayErrorStatus("%q does not match a collection name", answer)
					continue
			}
			displayErrorStatus("%q is not a known action", answer)
			continue
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

// generateRssFiles will generate RSS files for collections all collections
func generateRssFiles(scanner *bufio.Scanner, options []string, cfgName string, cfg *AppConfig) error {
	var (
		cName = "pages.md"
		rssFeed string
		fromDate string
		toDate string
		count int
		err error
	)
	if len(options) < 1 {
		term.Printf("Enter a collection name: ")
		scanner.Scan()
		cName = scanner.Text()
	} else {
		cName = options[0]
	}
	if len(options) < 2 {
		term.Printf("Enter a RSS filename: ")
		scanner.Scan()
		rssFeed = scanner.Text()
	}  else {
		rssFeed = strings.TrimSpace(options[1])
	}
	// count cases of passing from/to dates or count
	switch {
		case len(options) == 4:
			fromDate, toDate = options[2], options[3]
		case len(options) == 3:
			count, err = strconv.Atoi(options[2])
			if err != nil {
				return fmt.Errorf("%q, %s", options[2], err)
			}
	}
	if fromDate == ""  {
		term.Printf("Enter from date (YYYY-MM-DD or \"\"): ")
		scanner.Scan()
		fromDate = scanner.Text()
	}
	if toDate == "" {
		term.Printf("Enter to date (YYYY-MM-DD or \"\"): ")
		scanner.Scan()
		toDate = scanner.Text()
	}
	if (fromDate == "" && toDate == "") {
		term.Printf("Enter item count for feed (example \"10\")): ")
		scanner.Scan()
		txt := scanner.Text()
		count, err = strconv.Atoi(txt)
		if err != nil {
			return fmt.Errorf("%q, %s", options[2], err)
		}
	}
	return cfg.RssPosts(cName, rssFeed, count, fromDate, toDate)
}

// generateSitemapFiles will generate Sitemap XML files for all collections
func generateSitemapFiles(scanner *bufio.Scanner, options []string, cfgName string, cfg *AppConfig) error {
	return cfg.Sitemap() 
}

// appleThemes list themes available in the project
func listThemes(scanner *bufio.Scanner, options []string, cfgName string, cfg *AppConfig) error {
	return fmt.Errorf("listThemes() not implemented yet.")
}

// runPreview will run the web server on localhost so you can preview your site in the web browser
func runPreview(scanner *bufio.Scanner, options []string, cfgName string, cfg *AppConfig) error {
	return fmt.Errorf("runPreview() not implemented yet.")
}
