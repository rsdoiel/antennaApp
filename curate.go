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

func normalizePos(curPos int, tot int, pageSize int) {
	if curPos < 0 {
		curPos = 0
	} else if curPos >= tot {
		curPos = tot - pageSize
	}
	return curPos
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
	for quit := false; ! quit; {
		term.Move(1,1)
		term.ClrToEOL()

		term.Printf("Items in %s\n\n", collection.File)
		for i := curPos; i < tot && i < (curPos + pageSize); i++ {
			//link := mappedString(items[i], "link")
			title := mappedString(items[i], "title")
			postPath := mappedString(items[i], "postPath")
			status := mappedString(items[i], "status")
			//channel := mappedString(items[i], "channel")
			label := mappedString(items[i], "label")
			pubDate := mappedString(items[i], "pubDate")
			updated := mappedString(items[i], "updated")
			term.ClrToEOL()
			term.Printf("%3d %s\n\t%q %s %s %s %s\n",
				i + 1, title, status, label, postPath, pubDate, updated) 
		}
		// Display prompt
		term.ResetStyle()
		term.Printf("\n%d items (%d) ([n]exit, [p]rev, [h]elp or [q]uit): ", tot, pageSize)
		term.ClrToEOL()
		term.Refresh()
		if ! scanner.Scan() {
			continue
		}
		answer := scanner.Text()
		answer = strings.TrimSpace(strings.ToLower(answer))
		switch {
		case strings.HasPrefix(answer, "q"):
			quit = true
		case strings.HasPrefix(answer, "^"):
			curPos = 0
		case strings.HasPrefix(answer, "$"):
			curPos = tot - pageSize
		case strings.HasPrefix(answer, "+"):
			curPos = normalizePos(val + curPos)
		case strings.HasPrefix(answer, "+"):
			val, err := extractInt(answer)
			if err == nil {
			curPos = normalizePos(val + curPos, tot, pageSize)
			} else {
				displayErrorStatus("%s", err)
			}
		case strings.HasPrefix(answer, "-"):
			val, err := extractInt(answer)
			if err == nil {
				curPos = normalizePos(val - curPos, tot, pageSize)
			} else {
				displayErrorStatus("%s", err)
			}
		case strings.HasPrefix(answer, "f"):
			//FIXME: need to implement
			displayErrorStatus("apply filters not implemented")
			continue
		case strings.HasPrefix(answer, "p"):
			val, err := extractInt(answer)
			if err == nil {
			   displayStatus("Set %d to published", val)
			   //FIXME: update status to published for ink 
			   continue
			}
		case strings.HasPrefix(answer, "r"):
			val, err := extractInt(answer)
			if err == nil {
			   displayStatus("Set %d to review", val)
			   //FIXME: update status to review for link 
			   continue
			}			
		case answer == "":
			curPos = curPos + pageSize
			if curPos >= len(items) {
				curPos = 0
			}
		default:
			val, err := extractInt(answer)
			if err == nil {
				curPos = normalizePos( val - 1, tot, pageSize)
			}
			displayErrorStatus("%q, unknown command", answer)
		}
		term.Clear()
	}
	return nil
}

// createCollections provides the prompts to add a new collection 
func createCollection(params []string) error {
	return fmt.Errorf("createCollection not implemented")
}

// editCollections provides the prompts to edit an existing collection 
func editCollection(params []string) error {
	return fmt.Errorf("editCollection not implemented")
}

// removeCollection provides the prompts to delete a collection	scanner := bufio.NewScanner(os.Stdin)
func removeCollection(params []string) error {
	return fmt.Errorf("removeCollection not implemented")
}

// helpCollectionMenu explains how the options in the collection menu
func helpCollectionMenu(scanner *bufio.Scanner) {
	term.Clear()
	defer term.Clear()
	term.ResetStyle()
	term.Printf(`

%sCollection menu options. Commands have the following form.
%s
  MENU_NUMBER ENTER_KEY
  ACTION [PARAMETERS] ENTER_KEY
%s
Choices:

- To curate a collection's items, type in the menu number and
  press enter
- To create a new collection type in "n" or "new", a space then
  the collection name and press enter
- To edit a collection's metadata type in "edit", a space then the
  menu number for the collection and press enter
- To remove a collection type "remove", a space and the menu number
  for the collection you want to remove.
- To view help type "h" or "help" and press enter
- To quit type "q" or "quit" and press the enter

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

	colCount := len(cfg.Collections)
	scanner := bufio.NewScanner(os.Stdin)
	term.Clear()
	for quit := false; quit == false; {
		term.Move(1,1)
		term.ClrToEOL()
		// - A List collections
		term.Printf("Curate Collections\n\n")
		term.SetBold()
		for i, col := range cfg.Collections {
			term.ClrToEOL()
			term.Printf("\t%2d: %s, %s\n", i + 1, col.File, col.Title)
		}
		term.ResetStyle()
		term.Printf("\n(q to quit, h for help): ")
		term.ClrToEOL()
		term.Refresh()
		// Read entry
		if ! scanner.Scan() {
			continue
		}
		answer := scanner.Text()
		params := []string{}
		if strings.Contains(answer, " ") {
			params = strings.Split(answer, " ")
			answer, params = params[0], params[1:]
		}
		answer = strings.TrimSpace(strings.ToLower(answer))
		switch  {
		case answer == "":
			// just refresh display
			continue
		case (answer == "q" || answer == "quit"):
			quit = true
		case (answer == "n" || answer == "new"):
			if err := createCollection(params); err != nil {
				displayStatus("%s", err)
				continue
			}
		case (answer == "e" || answer == "edit"):
			if err := editCollection(params); err != nil {
				displayStatus("%s", err)
				continue
			}
		case (answer == "r" || answer == "remove" ):
			if err := removeCollection(params); err != nil {
				displayErrorStatus("%s",err)
				continue
			}
		case (answer == "h" || answer == "help"):
			helpCollectionMenu(scanner)
			continue
		default:
			val, err := strconv.Atoi(answer)
			if err != nil {
				displayErrorStatus("%q", err)
				continue
			}
			if (val < 1) || (val > colCount) {
				displayErrorStatus("enter a number between 1 and %d or enter q to quit", colCount)
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
func listItems (collection *Collection, args []string) ([]map[string]string, error) {
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
			link     string
			title    string
			description string
			sourceMarkdown string
			pubDate  string
			postPath string
			status   string
			channel string
			label string
			updated  string
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
		item := map[string]string {
			"link": link,
			"title": title,
			"description": description,
			"sourceMarkdown": sourceMarkdown,
			"pubDate": pubDate,
			"postPath": postPath,
			"status": status,
			"channel": channel,
			"label": label,
			"updated": updated,
			
		}
		items = append(items, item)
	}
	if i == 0 {
		return nil, fmt.Errorf("no published posts")
	}
	return items, nil
}

