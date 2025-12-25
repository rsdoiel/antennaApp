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
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"

	// My packages
	"github.com/rsdoiel/termlib"

	// 3rd Party
	_ "github.com/glebarez/go-sqlite"
)

var term = termlib.New(os.Stdout)

// getString retrieves a string by the map's key, returns
// empty string if not found.
func getString(m map[string]string, key string) string {
	if s, ok := m[key]; ok {
		return s
	}
	return ""
}

// parseAnswer takes a string, parses it into args using
// flag package.
func parseAnswer(s string) (string, []string, error) {
	options := strings.Fields(s)	
	if len(options) > 1 {
		return options[0], options[1:], nil
	}
	return s, []string{}, nil
}

func extractInt(s string) (int, error) {
	var numStr string
	for _, r := range s {
		if unicode.IsDigit(r) {
			numStr += string(r)
		}
	}
	return strconv.Atoi(numStr)
}

func normalizePos(curPos int, pageSize int, tot int) int {
	if curPos >= tot {
		curPos = tot - pageSize
	}
	if curPos < 0 {
		curPos = 0
	}	
	return curPos
}

// helpCurateItems explains how the options in the collection menu
func helpCurateItems(scanner *bufio.Scanner) {
	term.Clear()
	defer term.Clear()
	term.ResetStyle()
	term.Printf(`

%sCurate items. Command syntax.
%s
  NUMBER ENTER
  ACTION [PARAMETERS] ENTER_KEY
%s
Actions:

NUMBER
: Move to item NUMBER

+NUMBER or -NUMBER
: Page by NUMBER of items through list

[f]ilter
: apply SQL filters to items in the collection

[p]ublish NUMBER [NUMBER ...]
: set item NUMBER to published status

[r]eview NUMBER [NUMBER ...]
: set item NUMBER to review status

[c]lear NUMBER [NUMBER ...]
: clear item NUMBER to status value

[h]elp
: This help page

[q]uit
: Exit the items menu

(NOTE: Pressing enter without an action will page through results)

Press enter to exit help.
`, termlib.Cyan, termlib.Italic, termlib.Reset)
	term.Refresh()
	scanner.Scan()
}

// applyFilter, runs the SQL filters defined for the collection
func applyFilter(collection *Collection) error {
	dsn := collection.DbName
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	if db == nil {
		return fmt.Errorf("%s opened and returned nil", dsn)
	}
	return collection.ApplyFilters(db)
}

// setItemStatus sets the status column to "published" for each
// item number provided
func setItemStatus(status string, options []string, items []map[string]string, collection *Collection) error {
	dsn := collection.DbName
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	if db == nil {
		return fmt.Errorf("%s opened and returned nil", dsn)
	}
	for _, option := range options {
		val, err := extractInt(option)
		if err != nil {
			displayErrorStatus("failed to parse item number %q, %s", option, err)
			continue
		}
		itemNo := val - 1
		if itemNo >= 0 && itemNo < len(items) {
			link := getString(items[itemNo], "link")
			if _, err := db.Exec(SQLSetItemStatus, status, link); err != nil {
				displayErrorStatus("set to %s failed for item %d, %s", status, itemNo, err)
			}
		} else {
			displayErrorStatus("%d is out of item range", itemNo)
			continue
		}
	}
	return nil
}


// pageTo calculates the new position based on a string indicating distance
// the current position, pagesize and total items. If an error occurs
// the current position is returned along with the error. If no error then
// the new position is returned along with a nil error value.
func pageTo(s string, curPos int, pageSize int, tot int) (int, error)  {
	val, err := extractInt(s)
	if  err != nil {
		return curPos, err
	}
	switch {
	case strings.HasPrefix(s, "-"):
		curPos = normalizePos(curPos-val, pageSize, tot)
	case strings.HasPrefix(s, "+"):
		curPos = normalizePos(val+curPos, pageSize, tot)
	default:
		return curPos, fmt.Errorf("unable to parse %q", s)
	}
	return curPos, nil
}

// CurateCollection displays items in a collection so you can select items for publication
func curateItems(scanner *bufio.Scanner, collection *Collection) error {
	// Clear the screen
	term.Clear()
	defer term.Clear()
	term.ResetStyle()
	args := []string{}
	pageSize := int((term.GetTerminalHeight() - 5) / 2)
	curPos := 0
	args = append(args, fmt.Sprintf("%d", pageSize))
	items, err := listItems(collection, args)
	if err != nil {
		displayErrorStatus("%s", err)
	}
	tot := len(items)
	term.Clear()
	for quit := false; !quit; {
		term.Move(1, 1)
		term.ClrToEOL()

		term.Printf("Items in %s\n\n", collection.File)
		for i := curPos; i < tot && i < (curPos+pageSize); i++ {
			//link := getString(items[i], "link")
			//channel := getString(items[i], "channel")
			title := getString(items[i], "title")
			postPath := getString(items[i], "postPath")
			status := getString(items[i], "status")
			label := getString(items[i], "label")
			pubDate := getString(items[i], "pubDate")
			updated := getString(items[i], "updated")
			term.ClrToEOL()
			term.Printf("%4d %s%s%s\n\t%q %s %s %s %s\n",
				i+1, termlib.Bold+termlib.Italic, title, termlib.Reset, status, label, postPath, pubDate, updated)
		}
		// Display prompt
		term.ResetStyle()
		term.Printf("\n(%d/%d, [h]elp or [q]uit): ", curPos + 1,tot)
		term.ClrToEOL()
		term.Refresh()
		if !scanner.Scan() {
			continue
		}
		answer, options, err := parseAnswer(scanner.Text())
		answer = strings.TrimSpace(strings.ToLower(answer))
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
		case strings.HasPrefix(answer, "f"):
			if err := applyFilter(collection); err != nil {
				displayErrorStatus("%s", err)
				continue
			}
			items, err = listItems(collection, []string{})
			if err != nil {
				displayErrorStatus("%s", err)
			}
			tot = len(items)
		case strings.HasPrefix(answer, "p"):
			if err := setItemStatus("published", options, items, collection); err != nil {
				displayErrorStatus("%s", err)
				continue
			}
			items, err = listItems(collection, args)
			if err != nil {
				displayErrorStatus("%s", err)
			}
			tot = len(items)
		case strings.HasPrefix(answer, "r"):
			if err := setItemStatus("review", options, items, collection); err != nil {
				displayErrorStatus("%s", err)
				continue
			}
			items, err = listItems(collection, args)
			if err != nil {
				displayErrorStatus("%s", err)
			}
			tot = len(items)
		case strings.HasPrefix(answer, "c"):
			if err := setItemStatus("", options, items, collection); err != nil {
				displayErrorStatus("%s", err)
				continue
			}
			items, err = listItems(collection, args)
			if err != nil {
				displayErrorStatus("%s", err)
			}
			tot = len(items)
		case strings.HasPrefix(answer, "h"):
			helpCurateItems(scanner)
			continue
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

// addCollections provides the prompts to add a new collection
func addCollection(scanner *bufio.Scanner, options []string, cfgName string, cfg *AppConfig) error {
	var (
		cName string
		title string
		description string
	)
	displayStatus("DEBUG options list -> %s", strings.Join(options,", "))
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
		// See if this is a collection name or number
		if val, err := extractInt(options[0]); err == nil {
			// Figure out the cName for collection
			val = val - 1 // Adjust for zero based array
			if val >= 0 && val < len(cfg.Collections) {
				cName = cfg.Collections[val].File
				term.Printf("Remove %s? yes/NO ", cName)
				scanner.Scan()
				answer, _, _ := parseAnswer(scanner.Text())
				if (answer != "yes" && answer != "y") {
					return fmt.Errorf("delete %s cancelled", cName)
				}
			} else {
				return fmt.Errorf("%d is not a collection number", val + 1);
			}
		}
	}
	if cName == "" {
		return fmt.Errorf("Missing collection name")
	}
	return cfg.DelCollection(cfgName, cName)	

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


// helpCurateCollections explains how the options in the collection menu
func helpCurateCollections(scanner *bufio.Scanner) {
	term.Clear()
	defer term.Clear()
	term.ResetStyle()
	term.Printf(`

%sCurate collection(s). Command syntax.
%s
  NUMBER ENTER_KEY
  ACTION [NAME] ENTER_KEY
%s
Actions:

NUMBER
: curate collection NUMBER

+NUMBER or -NUMBER
: Page by NUMBER of items through list

[a]dd
: Add a new collection. You'll be prompted for a Markdown filename, a title
and description. If Markdown file already exists it'll be used along with any front matter providing title and description.

[d]elete NAME|NUMBER
: Remove collection with NAME or collection NUMBER from configuration. It does not remove the files from disk.

[H]arvest [NAME|NUMBER]
: Harvest all collections or one specified by NAME
or NUMBER

[g]enerate [NAME|NUMBER]
: Generate all collections or one specified by NAME
or NUMBER

[h]elp
: Display this help

[q]uit
: To quit

(NOTE: Pressing enter without an action will page through results)

Press enter to exit help.
`, termlib.Cyan, termlib.Italic, termlib.Reset)
	term.Refresh()
	scanner.Scan()
}

// display the status line
func displayStatus(format string, options ...interface{}) {
	// Get the current position
	row, col := term.GetCurPos()
	// Calc where the status line should go
	statusRow, statusCol := term.GetTerminalHeight(), 1
	term.Move(statusRow, statusCol)
	term.ClrToEOL()
	term.Printf(format, options...)
	term.Refresh()
	// Return to original position
	term.Move(row, col)
}

// displayErrorStatus, show a status message in Red
func displayErrorStatus(format string, options ...interface{}) {
	fgColor := term.GetFgColor()
	newFormat := fmt.Sprintf("%s%s%s", termlib.Red, format, fgColor)
	displayStatus(newFormat, options...)
}


// listItems returns a list of items for a collection
func listItems(collection *Collection, args []string) ([]map[string]string, error) {
	dsn := collection.DbName
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var (
		rows *sql.Rows
	)
	switch {
	case len(args) == 3:
		fromDate, toDate := args[1], args[2]
		rows, err = db.Query(SQLListDateRangeItems, fromDate, toDate)
		if err != nil {
			return nil, fmt.Errorf("%s\n%s, %s", SQLListDateRangeItems, dsn, err)
		}
	case len(args) == 2:
		count, err := strconv.Atoi(args[1])
		if err != nil {
			return nil, err
		}
		rows, err = db.Query(SQLListRecentItems, count)
		if err != nil {
			return nil, fmt.Errorf("%s\n%s, %s", SQLListRecentItems, dsn, err)
		}
	default:
		rows, err = db.Query(SQLListItems)
		if err != nil {
			return nil, fmt.Errorf("%s\n%s, %s", SQLListItems, dsn, err)
		}
	}
	if rows != nil {
		defer rows.Close()
	}

	i := 0
	items := []map[string]string{}
	for rows.Next() {
		var (
			link           string
			title          string
			description    string
			sourceMarkdown string
			pubDate        string
			postPath       string
			status         string
			channel        string
			label          string
			updated        string
		)
		if err := rows.Scan(&link, &title, &description, &sourceMarkdown, &pubDate, &postPath, &status, &channel, &label, &updated); err != nil {
			displayErrorStatus("failed to read row (%d), %s\n", i, err)
			continue
		}
		if strings.Contains(pubDate, "T") {
			parts := strings.SplitN(pubDate, "T", 2)
			pubDate = parts[0]
		}
		if i == 0 {
			i++
		}
		item := map[string]string{
			"link":           link,
			"title":          title,
			"description":    description,
			"sourceMarkdown": sourceMarkdown,
			"pubDate":        pubDate,
			"postPath":       postPath,
			"status":         status,
			"channel":        channel,
			"label":          label,
			"updated":        updated,
		}
		fmt.Fprintf(os.Stderr, "DEBUG item link %q, postPath %q\n", link, postPath) // DEBUG
		items = append(items, item)
	}
	if i == 0 {
		return nil, fmt.Errorf("no published posts")
	}
	return items, nil
}

// curateCollections provides the interaction loop for curating collections.
func curateCollections(scanner *bufio.Scanner, cfgName string, cfg *AppConfig) error {
	term.Clear()
	pageSize := int((term.GetTerminalHeight() - 5) / 2)
	curPos := 0
	for quit := false; quit == false; {
		term.Move(1, 1)
		term.ClrToEOL()
		// - A List collections
		term.Printf("Curate Collections\n\n")
		term.SetBold()
		tot := len(cfg.Collections)
		for i := curPos; i < tot && i < (curPos+pageSize); i++ {
			col := cfg.Collections[i]
			term.ClrToEOL()
			term.Printf("\t%2d: %s, %s%s%s\n", i+1, col.File, termlib.Bold + termlib.Italic, col.Title, termlib.Reset)
		}
		term.ResetStyle()
		term.Printf("\n(%d/%d, [h]elp or [q]uit): ", curPos + 1,tot)
		term.ClrToEOL()
		term.Refresh()
		// Read entry
		if !scanner.Scan() {
			continue
		}
		answer, options, err := parseAnswer(scanner.Text())
		switch {
		case strings.HasPrefix(answer, "+"):
			// page forward by N
			curPos, err = pageTo(answer, curPos, pageSize, tot)
			if err != nil { 
				displayErrorStatus("%s", err)
				continue
			}
		case strings.HasPrefix(answer, "-"):
			// page backward by N
			curPos, err = pageTo(answer, curPos, pageSize, tot)
			if err != nil { 
				displayErrorStatus("%s", err)
				continue
			}
		case answer == "":
			// next page
			curPos = normalizePos(curPos+pageSize, pageSize, tot)
		case strings.HasPrefix(answer,"a"):
			// Add a collection
			if err := addCollection(scanner, options, cfgName, cfg); err != nil {
				displayErrorStatus("%s", err)
				continue
			}	
			if err := cfg.LoadConfig(cfgName); err != nil {
				displayErrorStatus("%s", err)
			}
		case strings.HasPrefix(answer, "d"):
			if err := deleteCollection(scanner, options, cfgName, cfg); err != nil {
				displayErrorStatus("%s", err)
				continue
			}
		case strings.HasPrefix(answer, "H"):
			// Harvest collection(s)
			if err := harvestCollection(scanner, options, cfgName, cfg); err != nil {
				displayErrorStatus("%s", err)
				continue
			}
		case strings.HasPrefix(answer, "g"):
			// Generate pages and posts
			if err := generateCollection(scanner, options, cfgName, cfg); err != nil {
				displayErrorStatus("%s", err)
				continue	
			}
		case strings.HasPrefix(answer, "h"):
			helpCurateCollections(scanner)
			continue
		case strings.HasPrefix(answer, "q"):
			quit = true
		default:
			// If the answer is a number, go to item number
			if val, err := extractInt(answer); err == nil {
				if (val < 1) || (val > tot) {
					displayErrorStatus("enter a number the range of 1 to %d", tot)
					continue
				}
				// calc collection number to curate
				if err := curateItems(scanner, cfg.Collections[val - 1]); err != nil {
					displayErrorStatus("%q", err)
					continue
				}
			} else {
				displayErrorStatus("%q, unknown command", answer)
				continue
			}

		}
		term.Clear()
	}
	return nil
}

// listPages returns a list of items for a collection
func listPages(collection *Collection, args []string) ([]map[string]string, error) {
	dsn := collection.DbName
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var (
		rows *sql.Rows
	)
	rows, err = db.Query(SQLListPages)
	if err != nil {
		return nil, fmt.Errorf("%s\n%s, %s", SQLListItems, dsn, err)
	}
	if rows != nil {
		defer rows.Close()
	}

	i := 0
	pages := []map[string]string{}
	for rows.Next() {
		var (
			inputPath string
			outputPath string
			updated string
		)
		if err := rows.Scan(&inputPath, &outputPath, &updated)
			displayErrorStatus("failed to read row (%d), %s\n", i, err)
			continue
		}
		if i == 0 {
			i++
		}
		item := map[string]string{
			"inputPath": inputPath,
			"outputPath": outputPath,
			"updated": updated,
		}
		fmt.Fprintf(os.Stderr, "DEBUG page inputPath %q, outputPath %q, updated: %s\n", inputPath, outputPath, updated) // DEBUG
		items = append(items, item)
	}
	if i == 0 {
		return nil, fmt.Errorf("no pages found")
	}
	return items, nil
}

// helpCuratePages explains how the options in the collection pages menu
func helpCuratePages(scanner *bufio.Scanner) {
	term.Clear()
	defer term.Clear()
	term.ResetStyle()
	term.Printf(`

%sCurate pages. Command syntax.
%s
  NUMBER ENTER
  ACTION [PARAMETERS] ENTER_KEY
%s
Actions:

NUMBER
: Move to item NUMBER

+NUMBER or -NUMBER
: Page by NUMBER of items through list

[a]dd
: Add a page to the collection

[d]el NUMBER|NAME
: Delete a page from the collection. Doesn't delete the file on disk.

[g]enerate NUMBER [NUMBER ...]
: Render HTML for pages indicated by their number or name

[h]elp
: This help page

[q]uit
: Exit the items menu

(NOTE: Pressing enter without an action will page through results)

Press enter to exit help.
`, termlib.Cyan, termlib.Italic, termlib.Reset)
	term.Refresh()
	scanner.Scan()
}


// Curate provides a simple terminal interface to curating collections and 
// feed items for publication in your Antenna site.
func (app *AntennaApp) Curate(cfgName string, args []string) error {
	scanner := bufio.NewScanner(os.Stdin)
	if _, err := os.Stat(cfgName); os.IsNotExist(err) {
		term.Clear()
		term.Printf(`
	%s does not exist. Create it? %syes%s/no `, cfgName, termlib.Bold + termlib.Italic, termlib.Reset)
		scanner.Scan()
		answer, _, _ := parseAnswer(scanner.Text())
		if answer == "y" || answer == "yes" {
			if err := app.Init(cfgName, []string{}); err != nil {
				return err
			}
		}
	}
	cfg := &AppConfig{}
	if err := cfg.LoadConfig(cfgName); err != nil {
		return err
	}
	if cfg.Collections == nil || len(cfg.Collections) == 0 {
		// NOTE: shouldn't see this unless you have a partially
		// initialized project
		return fmt.Errorf("no collections found in %s", cfgName)
	}
	if err := curateCollections(scanner, cfgName, cfg); err != nil {
		return err
	}
	return nil
}

