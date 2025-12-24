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
	options := strings.Split(s, " ")
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
: apply SQL fitlers to items in collection

[p]ublish NUMBER [NUMBER ...]
: set item NUMBER to published status

[r]eview NUMBER [NUMBER ...]
: set item NUMBER to review status

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
	return fmt.Errorf("applyFilter not implemented")
}

// setPublishStatus sets the status column to "published" for each
// item number provided
func setPublishStatus(options []string, items []map[string]string, collection *Collection) error {
	return fmt.Errorf("setPublishStatus not implemented")
}

// setReviewStatus sets the status column to "published" for each
// item number provided
func setReviewStatus(options []string, items []map[string]string, collection *Collection) error {
	return fmt.Errorf("setReviewStatus not implemented")
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
			title := getString(items[i], "title")
			postPath := getString(items[i], "postPath")
			status := getString(items[i], "status")
			//channel := getString(items[i], "channel")
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
			if err := setPublishStatus(options, items, collection); err != nil {
				displayErrorStatus("%s", err)
				continue
			}
		case strings.HasPrefix(answer, "r"):
			if err := setReviewStatus(options, items, collection); err != nil {
				displayErrorStatus("%s", err)
				continue
			}
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

// createCollections provides the prompts to add a new collection
func createCollection(options []string) error {
	return fmt.Errorf("createCollection not implemented - %q", strings.Join(options,", "))
}

// editCollections provides the prompts to edit an existing collection
func editCollection(options []string) error {
	return fmt.Errorf("editCollection not implemented - %q", strings.Join(options,", "))
}

// removeCollection provides the prompts to delete a collection	scanner := bufio.NewScanner(os.Stdin)
func removeCollection(options []string) error {
	return fmt.Errorf("removeCollection not implemented - %q", strings.Join(options,", "))
}

// harvestCollection will retrieve and aggregate collection items
func harvestCollection(options []string) error {
	return fmt.Errorf("harvestCollection not implemented - %q", strings.Join(options,", "))
}

// generateCollection will generate pages and posts
func generateCollection(options []string) error {
	return fmt.Errorf("generateCollection not implemented - %q", strings.Join(options,", "))
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

[a]dd NAME
: Add a new collection with NAME

[r]emove NAME
: Remove collection with NAME

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
		items = append(items, item)
	}
	if i == 0 {
		return nil, fmt.Errorf("no published posts")
	}
	return items, nil
}

// curateCollections provides the interaction loop for curating collections.
func curateCollections(cfg *AppConfig) error {
	scanner := bufio.NewScanner(os.Stdin)
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
		case answer == "":
			curPos = normalizePos(curPos+pageSize, pageSize, tot)
		case strings.HasPrefix(answer,"a"):
			// Add a collection
			if err := createCollection(options); err != nil {
				displayErrorStatus("%s", err)
				continue
			}
		case strings.HasPrefix(answer, "e"):
			if err := editCollection(options); err != nil {
				displayErrorStatus("%s", err)
				continue
			}
		case strings.HasPrefix(answer, "r"):
			if err := removeCollection(options); err != nil {
				displayErrorStatus("%s", err)
				continue
			}
		case strings.HasPrefix(answer, "H"):
			// Harvest collection(s)
			if err := harvestCollection(options); err != nil {
				displayErrorStatus("%s", err)
				continue
			}
		case strings.HasPrefix(answer, "g"):
			// Generate pages and posts
			if err := generateCollection(options); err != nil {
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


// Curate provides a simple terminal interface to curating collections and 
// feed items for publication in your Antenna site.
func (app *AntennaApp) Curate(cfgName string, args []string) error {
	cfg := &AppConfig{}
	if err := cfg.LoadConfig(cfgName); err != nil {
		return err
	}
	if cfg.Collections == nil || len(cfg.Collections) == 0 {
		return fmt.Errorf("no collections found in %s", cfgName)
	}
	if err := curateCollections(cfg); err != nil {
		return err
	}
	return nil
}

