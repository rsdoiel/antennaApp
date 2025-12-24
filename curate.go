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

func mappedString(m map[string]string, key string) string {
	if s, ok := m[key]; ok {
		return s
	}
	return ""
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

func normalizePos(curPos int, tot int, pageSize int) int {
	if curPos >= tot {
		curPos = tot - pageSize
	}
	if curPos < 0 {
		curPos = 0
	}	
	return curPos
}

// helpItemMenu explains how the options in the collection menu
func helpItemMenu(scanner *bufio.Scanner) {
	term.Clear()
	defer term.Clear()
	term.ResetStyle()
	term.Printf(`

%sItem menu actions. Commands have the following form.
%s
  ACTION [ITEM_NO] ENTER_KEY
%s
Choices:

+NUMBER or -NUMBER
: Page by NUMBER of items through list

f
: apply fitler to items in collection

[g]oto NUMBER
: Move so item NUMBER is at top of page

[p]ublish NUMBER
: set item NUMBER to published status

[r]eview NUMBER
: set item NUMBER to review status

[h]elp
: This help page

[q]uit
: Exit the items menu

Pressing enter without action will move the item to the next set of results.

Press enter to exit help.
`, termlib.Cyan, termlib.Italic, termlib.Reset)
	term.Refresh()
	scanner.Scan()
}


func applyFilter(collection *Collection) error {
	return fmt.Errorf("applyFilter not implemented")
}

func setPublishStatus(itemNo int, items []map[string]string, collection *Collection) error {
	return fmt.Errorf("setPublishStatus(%d, items, collection) not implemented", itemNo)
}

func setReviewStatus(itemNo int, items []map[string]string, collection *Collection) error {
	return fmt.Errorf("setReviewStatus(%d, items, collection) not implemented", itemNo)
}

// CurateCollection displays items in a collection so you can select items for publication
func curateItems(scanner *bufio.Scanner, collection *Collection) error {
	// Clear the screen
	term.Clear()
	defer term.Clear()
	term.ResetStyle()
	args := []string{}
	pageSize := int((term.GetTerminalHeight() - 6) / 2)
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
			//link := mappedString(items[i], "link")
			title := mappedString(items[i], "title")
			postPath := mappedString(items[i], "postPath")
			status := mappedString(items[i], "status")
			//channel := mappedString(items[i], "channel")
			label := mappedString(items[i], "label")
			pubDate := mappedString(items[i], "pubDate")
			updated := mappedString(items[i], "updated")
			term.ClrToEOL()
			term.Printf("%4d %s%s%s\n\t%q %s %s %s %s\n",
				i+1, termlib.Bold+termlib.Italic, title, termlib.Reset, status, label, postPath, pubDate, updated)
		}
		// Display prompt
		term.ResetStyle()
		term.Printf("\n(%d items, [h]elp or [q]uit): ", tot)
		term.ClrToEOL()
		term.Refresh()
		if !scanner.Scan() {
			continue
		}
		answer := scanner.Text()
		answer = strings.TrimSpace(strings.ToLower(answer))
		val, valErr := extractInt(answer)
		switch {
		case strings.HasPrefix(answer, "q"):
			quit = true
		case strings.HasPrefix(answer, "+"):
			if valErr == nil {
				curPos = normalizePos(val+curPos, tot, pageSize)
			} else {
				displayErrorStatus("%s", valErr)
				continue
			}
		case strings.HasPrefix(answer, "-"):
			if valErr == nil {
				curPos = normalizePos(curPos-val, tot, pageSize)
			} else {
				displayErrorStatus("%s", valErr)
				continue
			}
		case strings.HasPrefix(answer, "-"):
			if valErr == nil {
				curPos = normalizePos(val-curPos, tot, pageSize)
			} else {
				displayErrorStatus("%s", valErr)
				continue
			}
		case strings.HasPrefix(answer, "g"):
			if valErr == nil {
				curPos = normalizePos(val-1, tot, pageSize)
			} else {
				displayErrorStatus("%s", valErr)
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
			displayStatus("filters applied")
		case strings.HasPrefix(answer, "p"):
			if valErr == nil {
				if err := setPublishStatus(val - 1, items, collection); err != nil {
					displayErrorStatus("%s", err)
					continue
				}
				displayStatus("published %d", val)
			} else {
				displayErrorStatus("%s", valErr)
				continue
			}
		case strings.HasPrefix(answer, "r"):
			if valErr == nil {
				if err := setReviewStatus(val - 1, items, collection); err != nil {
					displayErrorStatus("%s", err)
					continue
				}
				displayStatus("review %d", val)
			} else {
				displayErrorStatus("%s", valErr)
				continue
			}
		case strings.HasPrefix(answer, "h"):
			helpItemMenu(scanner)
			continue
		case answer == "":
			curPos = normalizePos(curPos+pageSize, tot, pageSize)
		default:
			displayErrorStatus("%q, unknown command", answer)
			continue
		}
		term.Clear()
	}
	return nil
}

// createCollections provides the prompts to add a new collection
func createCollection(params []string) error {
	return fmt.Errorf("createCollection not implemented - %q", strings.Join(params,", "))
}

// editCollections provides the prompts to edit an existing collection
func editCollection(params []string) error {
	return fmt.Errorf("editCollection not implemented - %q", strings.Join(params,", "))
}

// removeCollection provides the prompts to delete a collection	scanner := bufio.NewScanner(os.Stdin)
func removeCollection(params []string) error {
	return fmt.Errorf("removeCollection not implemented - %q", strings.Join(params,", "))
}

// harvestCollection will retrieve and aggregate collection items
func harvestCollection(params []string) error {
	return fmt.Errorf("harvestCollection not implemented - %q", strings.Join(params,", "))
}

// generateCollection will generate pages and posts
func generateCollection(params []string) error {
	return fmt.Errorf("generateCollection not implemented - %q", strings.Join(params,", "))
}


// helpCollectionMenu explains how the options in the collection menu
func helpCollectionMenu(scanner *bufio.Scanner) {
	term.Clear()
	defer term.Clear()
	term.ResetStyle()
	term.Printf(`

%sCollection menu options. Commands have the following form.
%s
  NUMBER ENTER_KEY
  ACTION [NAME] ENTER_KEY
%s
Choices:

NUMBER
: Curate items in collection NUMBER

[a]dd NAME
: Add a new collection with NAME

[r]emove NAME
: Remove collection with NAME

[ha]rvest [NAME]
: Harvest all or NAME collection(s)

[ge]nerate [NAME]
: Generate pages for all or NAME collection

[h]elp
: Display this help

[q]uit
: To quit

Pressing enter at the menu will display the next page of results.

Press enter to exit help.
`, termlib.Cyan, termlib.Italic, termlib.Reset)
	term.Refresh()
	scanner.Scan()
}

// display the status line
func displayStatus(format string, params ...interface{}) {
	// Get the current position
	row, col := term.GetCurPos()
	// Calc where the status line should go
	statusRow, statusCol := term.GetTerminalHeight(), 1
	term.Move(statusRow, statusCol)
	term.ClrToEOL()
	term.Printf(format, params...)
	term.Refresh()
	// Return to original position
	term.Move(row, col)
}

// displayErrorStatus, show a status message in Red
func displayErrorStatus(format string, params ...interface{}) {
	fgColor := term.GetFgColor()
	newFormat := fmt.Sprintf("%s%s%s", termlib.Red, format, fgColor)
	displayStatus(newFormat, params...)
}

// Curate provides a simple terminal interface to curating feed items for publication in your Antenna site.
func (app *AntennaApp) Curate(cfgName string, args []string) error {
	cfg := &AppConfig{}
	if err := cfg.LoadConfig(cfgName); err != nil {
		return err
	}
	if cfg.Collections == nil || len(cfg.Collections) == 0 {
		return fmt.Errorf("no collections found in %s", cfgName)
	}

	scanner := bufio.NewScanner(os.Stdin)
	term.Clear()
	pageSize := int((term.GetTerminalHeight() - 6) / 2)
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
		term.Printf("\n(q to quit, h for help): ")
		term.ClrToEOL()
		term.Refresh()
		// Read entry
		if !scanner.Scan() {
			continue
		}
		answer := scanner.Text()
		params := []string{}
		if strings.Contains(answer, " ") {
			params = strings.Split(answer, " ")
			answer, params = params[0], params[1:]
		}
		answer = strings.TrimSpace(strings.ToLower(answer))
		val, valErr := extractInt(answer)
		switch {
		case strings.HasPrefix(answer, "ha"):
			// Harvest collection(s)
			if err := harvestCollection(params); err != nil {
				displayErrorStatus("%s", err)
				continue
			}
		case strings.HasPrefix(answer, "ge"):
			// Generate pages and posts
			if err := generateCollection(params); err != nil {
				displayErrorStatus("%s", err)
				continue	
			}
		case strings.HasPrefix(answer, "q"):
			quit = true
		case strings.HasPrefix(answer, "+"):
			if valErr == nil {
				curPos = normalizePos(val+curPos, tot, pageSize)
			} else {
				displayErrorStatus("%s", valErr)
				continue
			}
		case strings.HasPrefix(answer, "-"):
			if valErr == nil {
				curPos = normalizePos(curPos-val, tot, pageSize)
			} else {
				displayErrorStatus("%s", valErr)
				continue
			}
		case strings.HasPrefix(answer, "-"):
			if valErr == nil {
				curPos = normalizePos(val-curPos, tot, pageSize)
			} else {
				displayErrorStatus("%s", valErr)
				continue
			}
		case strings.HasPrefix(answer, "g"):
			if valErr == nil {
				curPos = normalizePos(val-1, tot, pageSize)
			} else {
				displayErrorStatus("%s", valErr)
				continue
			}
		case answer == "":
			curPos = normalizePos(curPos+pageSize, tot, pageSize)
		case strings.HasPrefix(answer,"a"):
			// Add a collection
			if err := createCollection(params); err != nil {
				displayErrorStatus("%s", err)
				continue
			}
		case strings.HasPrefix(answer, "e"):
			if err := editCollection(params); err != nil {
				displayErrorStatus("%s", err)
				continue
			}
		case strings.HasPrefix(answer, "r"):
			if err := removeCollection(params); err != nil {
				displayErrorStatus("%s", err)
				continue
			}
		case strings.HasPrefix(answer, "h"):
			helpCollectionMenu(scanner)
			continue
		default:
			if (val < 1) || (val > tot) {
				displayErrorStatus("enter a number between 1 and %d or enter q to quit", tot)
				continue
			}
			// calc collection number to curate
			i := val - 1
			if err := curateItems(scanner, cfg.Collections[i]); err != nil {
				displayErrorStatus("%q", err)
				continue
			}
		}
		term.Clear()
	}
	return nil
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
