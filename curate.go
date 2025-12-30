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

/**
 * TUI
 */
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

// constrainText
func constrainText(s string, length int) string {
    if len(s) < length {
    	return s
    }
    runes := []rune(s)
    if len(runes) > length {
    	// There needs to be an visual indicator the constrained
    	// text is incomplete.
    	return string(runes[:length]) + "[...]"
    }
    return s
}


// Set the new cursor position within bounds of page size and
// total number of things to list.
func normalizePos(curPos int, pageSize int, tot int) int {
	if curPos >= tot {
		curPos = tot - pageSize
	}
	if curPos < 0 {
		curPos = 0
	}	
	return curPos
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



/**
 * Post methods
 */

// viewPost displays the full object retrieved from listPost
func viewPost(scanner *bufio.Scanner, post map[string]string) error {
	term.Clear()
	defer term.Clear()
	parts := []string{}
	title := getString(post, "title")
	link := getString(post, "link")
	description := getString(post, "description")
	postPath := getString(post, "postPath")
	sourceMarkdown := getString(post, "sourceMarkdown")
	pubDate := getString(post, "pubDate")
	status := getString(post, "status")
	//channel := getString(post, "channel")
	label := getString(post, "label")
	if title != "" {
		parts = append(parts, fmt.Sprintf("%s%s%s\n\n",
			termlib.Bold + termlib.Italic,
			title,
			termlib.Reset))
	}
	if sourceMarkdown != "" {
		parts = append(parts, fmt.Sprintf("%s\n\n", sourceMarkdown)) 
	}
	if postPath != "" {
		parts = append(parts, fmt.Sprintf("Post Path: %s\n\n", postPath))
	}
	if description != "" {
		parts = append(parts, fmt.Sprintf("Description: %s\n\n", description))
	}
	if link != "" {
		parts = append(parts, fmt.Sprintf("\n\t%s\n\n", link))
	}
	if pubDate != "" || status != "" || label != "" {
		parts = append(parts, fmt.Sprintf("status: %s %s %s\n", status, label, pubDate))
	}

	pageSize := int((term.GetTerminalHeight() - 5) / 2)
	//pageWidth := term.GetTerminalHeight() - 12
	parts = strings.Split(strings.Join(parts, ""), "\n")
	tot := len(parts)
	curPos := 0
	for quit := false; ! quit; {
		if len(parts) < pageSize {
			term.Println(strings.Join(parts, "\n"))
			displayStatus("press enter to return to menu")
			scanner.Scan()
			quit = true;
		} else {
			for i := curPos; i < tot && i < (curPos+pageSize); i++ {
				term.ClrToEOL()
				term.Printf("%s\n", parts[i])
			}
			displayStatus("press enter for next page, q and enter to quit")
			scanner.Scan()
			answer, _, _ := parseAnswer(scanner.Text())
			switch {
			case answer == "":
				curPos = normalizePos(curPos+pageSize, pageSize, tot)
			case strings.HasPrefix(answer, "q"):
				quit = true
			}
		}
	}
	return nil
}

// listPosts returns a list of posts in a collection
func listPosts(collection *Collection, args []string) ([]map[string]string, error) {
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
		rows, err = db.Query(SQLCurateDateRangePosts, fromDate, toDate)
		if err != nil {
			return nil, fmt.Errorf("%s\n%s, %s", SQLCurateDateRangePosts, dsn, err)
		}
	case len(args) == 2:
		count, err := strconv.Atoi(args[1])
		if err != nil {
			return nil, err
		}
		rows, err = db.Query(SQLCurateRecentPosts, count)
		if err != nil {
			return nil, fmt.Errorf("%s\n%s, %s", SQLCurateRecentPosts, dsn, err)
		}
	default:
		rows, err = db.Query(SQLCuratePosts)
		if err != nil {
			return nil, fmt.Errorf("SQL: %s\n%s\n%s, %s", SQLCuratePosts, dsn, err)
		}
	}
	if rows != nil {
		defer rows.Close()
	}

	i := 0
	posts := []map[string]string{}
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
		post := map[string]string{
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
		posts = append(posts, post)
	}
	if i == 0 {
		return nil, fmt.Errorf("no published posts")
	}
	return posts, nil
}


// addPost prompts for setting up a post in the collection from a Markdown document
// on local disk.
func addPost(scanner *bufio.Scanner, options []string, cfg *AppConfig, collection *Collection) error {
	// Clear the screen
	term.Clear()
	defer term.Clear()
	var fName string
	if len(options) > 0 {
		fName = options[0]
	}
	if fName == "" {
		term.Printf("Enter name: ")
		scanner.Scan()
		fName, _, _ = parseAnswer(scanner.Text())
	}
	if fName == "" {
		return fmt.Errorf("No filename entered")
	}
	cName := collection.File
	if err := cfg.Post(cName, fName); err != nil {
		return fmt.Errorf("failed to add post %q, %s", fName, err)
	}
	term.Printf("%s added to %s, press enter to return menu", fName, cName)
	scanner.Scan()
	scanner.Text()
	return nil
}

// setPublishPost will set the publication date from the front matter and status to published. If datePublished is misisng from front matter it'll update the Markdown post's front matter then set the item row values.
func setPublishPost(scanner *bufio.Scanner, options []string, posts []map[string]string, cfg *AppConfig, collection *Collection) error {
	// Clear the screen
	term.Clear()
	defer term.Clear()
	if len(options) == 0 {
		return fmt.Errorf("Missing post no or file path")
	}
	cName := collection.File
	var fName string
	for _, option := range options {
		itemNo := -1
		if val, err := extractInt(option); err == nil {
			itemNo = val - 1
			if itemNo >= 0 && itemNo < len(posts) {
				fName = getString(posts[itemNo], "postPath")
				if fName == "" {
					fName = getString(posts[itemNo], "postPath")
				}
			}
		} else {
			fName = option
		}
		if fName == "" {
			return fmt.Errorf("cannot find post to publish")
		}
		if err := cfg.PublishPost(cName, fName); err != nil {
			return err
		}
	}
	return nil
}


// delPost removes a post from the collection. Does not remove the Markdown or HTML renderings from local disk.
func delPost(scanner *bufio.Scanner, options []string, posts []map[string]string, cfg *AppConfig, collection *Collection) error {
	// Clear the screen
	term.Clear()
	defer term.Clear()
	if len(options) == 0 {
		return fmt.Errorf("Missing post no or file path")		
	}
	cName := collection.File
	var fName string
	for _, option := range options {
		fName = ""
		itemNo := -1
		if val, err := extractInt(option); err == nil {
			itemNo = val - 1
			term.Printf("DEBUG itemNo -> %d\n", itemNo)
			if itemNo >= 0 && itemNo < len(posts) {
				fName = getString(posts[itemNo], "postPath")
			}
		}
		if fName == "" {
			fName = option
		}
		if fName == "" {
			return fmt.Errorf("cannot find post to remove")
		}
		if err := cfg.Unpost(cName, fName); err != nil {
			return fmt.Errorf("unable to remove post %q for %q, %s", fName, cName, err)
		}
	}
	return nil
}



// helpCuratePosts explains how the options in the collection posts menu
func helpCuratePosts(scanner *bufio.Scanner) {
	term.Clear()
	defer term.Clear()
	term.ResetStyle()
	term.Printf(`
%sCurate posts. Command syntax.
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
: Add Markdown document as a post from the local file system

[d]el
: Removes a post from the collection. Does not remove the Markdown document from
the file system.

[p]ublish
: Set all posts status to a "published".

[v]iew NUMBER
: view the post identified by NUMBER

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

// curatePosts displays items in a collection so you can select items for publication
func curatePosts(scanner *bufio.Scanner, cfgName string, cfg *AppConfig, collection *Collection) error {
	// Clear the screen
	term.Clear()
	defer term.Clear()
	args := []string{}
	pageSize := int((term.GetTerminalHeight() - 5) / 2)
	pageWidth := term.GetTerminalHeight() - 12
	curPos := 0
	args = append(args, fmt.Sprintf("%d", pageSize))
	posts, err := listPosts(collection, args)
	if err != nil {
		displayErrorStatus("%s", err)
	}
	for quit := false; !quit; {
		tot := len(posts)
		term.Move(1, 1)
		term.ClrToEOL()
		term.Printf("Posts in %s\n\n", collection.File)
		for i := curPos; i < tot && i < (curPos+pageSize); i++ {
			title := constrainText(getString(posts[i], "title"), pageWidth)
			postPath := getString(posts[i], "postPath")
			status := getString(posts[i], "status")
			label := getString(posts[i], "label")
			pubDate := getString(posts[i], "pubDate")
			updated := getString(posts[i], "updated")
			term.ClrToEOL()
			term.Printf("%4d %s%s%s\n\t%q %s %s %s %s\n",
				i+1, termlib.Bold+termlib.Italic, title, termlib.Reset, status, label, postPath, pubDate, updated)
		}
		// Display prompt
		term.ResetStyle()
		term.Printf("\n(%d/%d, [h]elp or [q]uit): ", curPos + 1, tot)
		term.ClrToEOL()
		term.Refresh()
		if !scanner.Scan() {
			continue
		}
		answer, options, err := parseAnswer(scanner.Text())
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
		case strings.HasPrefix(answer, "a"):
			if err = addPost(scanner, options, cfg, collection); err != nil {
				displayErrorStatus("%s", err)
				continue
			}
			posts, err = listPosts(collection, []string{})
			if err != nil {
				displayErrorStatus("%s", err)
				continue
			}
			tot = len(posts)
		case strings.HasPrefix(answer, "d"):
			if err = delPost(scanner, options, posts, cfg, collection); err != nil {
				displayErrorStatus("%s", err)
				continue				
			}
			posts, err = listPosts(collection, []string{})
			if err != nil {
				displayErrorStatus("%s", err)
				continue
			}
			tot = len(posts)
		case strings.HasPrefix(answer, "p"):
			// FIXME: Make sure the Front Matter is updated and we
			if err := setPublishPost(scanner, options, posts, cfg, collection); err != nil {
				displayErrorStatus("%s", err)
				continue
			}
			posts, err = listPosts(collection, args)
			if err != nil {
				displayErrorStatus("%s", err)
				continue
			}
			tot = len(posts)
		case strings.HasPrefix(answer, "r"):
			if err := setItemStatus("review", options, posts, collection); err != nil {
				displayErrorStatus("%s", err)
				continue
			}
			posts, err = listPosts(collection, args)
			if err != nil {
				displayErrorStatus("%s", err)
				continue
			}
			tot = len(posts)
		case strings.HasPrefix(answer, "c"):
			if err := setItemStatus("", options, posts, collection); err != nil {
				displayErrorStatus("%s", err)
				continue
			}
			posts, err = listPosts(collection, args)
			if err != nil {
				displayErrorStatus("%s", err)
				continue
			}
			tot = len(posts)
		case strings.HasPrefix(answer, "v"):
			if len(options) == 0 {
				displayErrorStatus("missing item number to view")
				continue
			}
			if val, err := extractInt(options[0]); err != nil {
				displayErrorStatus("not a number, %s", err)
				continue
			} else {
				itemNo := val - 1
				if itemNo >= 0 && itemNo < tot {
					if err = viewPost(scanner, posts[itemNo]); err != nil {
						displayErrorStatus("%s", err)
						continue
					} 
				} else {
					displayErrorStatus("number %d should be between 1 and %d", itemNo, tot)
					continue
				}
			}
		
		case strings.HasPrefix(answer, "h"):
			helpCuratePosts(scanner)
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


/**
 * Item related methods
 */
 
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

// viewItem displays the full object retrieved from listItems
func viewItem(scanner *bufio.Scanner, item map[string]string) error {
	term.Clear()
	defer term.Clear()
	parts := []string{}
	title := getString(item, "title")
	link := getString(item, "link")
	description := getString(item, "description")
	sourceMarkdown := getString(item, "sourceMarkdown")
	pubDate := getString(item, "pubDate")
	status := getString(item, "status")
	//channel := getString(item, "channel")
	label := getString(item, "label")
	if title != "" {
		parts = append(parts, fmt.Sprintf("%s%s%s\n\n",
			termlib.Bold + termlib.Italic,
			title,
			termlib.Reset))
	}
	if sourceMarkdown != "" {
		parts = append(parts, fmt.Sprintf("%s\n\n", sourceMarkdown)) 
	} else if description != "" {
		parts = append(parts, fmt.Sprintf("%s\n\n", description))
	}
	if link != "" {
		parts = append(parts, fmt.Sprintf("\n\t%s\n\n", link))
	}
	if pubDate != "" || status != "" || label != "" {
		parts = append(parts, fmt.Sprintf("status: %s %s %s\n", status, label, pubDate))
	}

	pageSize := int((term.GetTerminalHeight() - 5) / 2)
	//pageWidth := term.GetTerminalHeight() - 12
	parts = strings.Split(strings.Join(parts, ""), "\n")
	tot := len(parts)
	curPos := 0
	for quit := false; ! quit; {
		if len(parts) < pageSize {
			term.Println(strings.Join(parts, "\n"))
			displayStatus("press enter to return to menu")
			scanner.Scan()
			quit = true;
		} else {
			for i := curPos; i < tot && i < (curPos+pageSize); i++ {
				term.ClrToEOL()
				term.Printf("%s\n", parts[i])
			}
			displayStatus("press enter for next page, q and enter to quit")
			scanner.Scan()
			answer, _, _ := parseAnswer(scanner.Text())
			switch {
			case answer == "":
				curPos = normalizePos(curPos+pageSize, pageSize, tot)
			case strings.HasPrefix(answer, "q"):
				quit = true
			}
		}
	}
	return nil
}


// helpCurateItems explains how the options in the collection menu
func helpCurateItems(scanner *bufio.Scanner) {
	term.Clear()
	defer term.Clear()
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

[v]iew NUMBER
: view the full item detail

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


// curateItems displays items in a collection so you can select items for publication
func curateItems(scanner *bufio.Scanner, collection *Collection) error {
	// Clear the screen
	term.Clear()
	defer term.Clear()
	args := []string{}
	pageSize := int((term.GetTerminalHeight() - 5) / 2)
	pageWidth := term.GetTerminalWidth() - 12
	curPos := 0
	args = append(args, fmt.Sprintf("%d", pageSize))
	items, err := listItems(collection, args)
	if err != nil {
		displayErrorStatus("%s", err)
	}
	tot := len(items)
	for quit := false; !quit; {
		term.Move(1, 1)
		term.ClrToEOL()

		term.Printf("Items in %s\n\n", collection.File)
		for i := curPos; i < tot && i < (curPos+pageSize); i++ {
			title := constrainText(getString(items[i], "title"), pageWidth)
			if title == "" {
				title = constrainText(getString(items[i], "description"), pageWidth)
			}
			status := getString(items[i], "status")
			label := getString(items[i], "label")
			pubDate := getString(items[i], "pubDate")
			term.ClrToEOL()
			term.Printf("%4d %s%s%s\n\tstatus: %s %s %s\n",
				i+1, termlib.Bold + termlib.Italic, title, termlib.Reset,
				status, label, pubDate)
		}
		// Display prompt
		term.Printf("\n(%d/%d, [h]elp or [q]uit): ", curPos + 1,tot)
		term.ClrToEOL()
		term.Refresh()
		if !scanner.Scan() {
			continue
		}
		answer, options, err := parseAnswer(scanner.Text())
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
				continue
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
				continue
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
				continue
			}
			tot = len(items)
		case strings.HasPrefix(answer, "v"):
			if len(options) == 0 {
				displayErrorStatus("missing item number to view")
				continue
			}
			if val, err := extractInt(options[0]); err != nil {
				displayErrorStatus("not a number, %s", err)
				continue
			} else {
				itemNo := val - 1
				if itemNo >= 0 && itemNo < tot {
					if err = viewItem(scanner, items[itemNo]); err != nil {
						displayErrorStatus("%s", err)
						continue
					} 
				} else {
					displayErrorStatus("number %d should be between 1 and %d", itemNo, tot)
					continue
				}
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

/**
 * Page methods
 */

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
		if err = rows.Scan(&inputPath, &outputPath, &updated); err != nil {
			displayErrorStatus("failed to read row (%d), %s\n", i, err)
			continue
		}
		if i == 0 {
			i++
		}
		page := map[string]string{
			"inputPath": inputPath,
			"outputPath": outputPath,
			"updated": updated,
		}
		pages = append(pages, page)
	}
	if i == 0 {
		return nil, fmt.Errorf("no pages found")
	}
	return pages, nil
}


// addPage prompts for setting up a page in the collection from a Markdown document
// on local disk.
func addPage(scanner *bufio.Scanner, options []string, pages []map[string]string, collection *Collection) error {
	// Clear the screen
	term.Clear()
	defer term.Clear()
	term.Println(`

DEBUG addPage not implemented yet.

press enter to return to previous menu
`)
	scanner.Scan()
	_, _, _ = parseAnswer(scanner.Text())
	return fmt.Errorf("addPage() not implemented.")
}

// delPage removes a page from the collection. Does not remove the Markdown or HTML renderings from local disk.
func delPage(scanner *bufio.Scanner, options []string, pages []map[string]string, collection *Collection) error {
	// Clear the screen
	term.Clear()
	defer term.Clear()
	term.Println(`

DEBUG delPage not implemented yet.

press enter to return to previous menu
`)
	scanner.Scan()
	_, _, _ = parseAnswer(scanner.Text())
	return fmt.Errorf("delPage() not implemented.")
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
: Add a Markdown document as a page to the collection

[d]el NUMBER|NAME
: Remove the page from the collection. Does not delete the file(s) on disk.

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


// curatePages displays items in a collection so you can select items for publication
func curatePages(scanner *bufio.Scanner, collection *Collection) error {
	// Clear the screen
	term.Clear()
	defer term.Clear()
	term.ResetStyle()
	args := []string{}
	pageSize := int((term.GetTerminalHeight() - 5) / 2)
	curPos := 0
	args = append(args, fmt.Sprintf("%d", pageSize))
	pages, err := listPages(collection, args)
	if err != nil {
		displayErrorStatus("%s", err)
	}
	tot := len(pages)
	for quit := false; !quit; {
		term.Move(1, 1)
		term.ClrToEOL()

		term.Printf("Pages in %s\n\n", collection.File)
		for i := curPos; i < tot && i < (curPos+pageSize); i++ {
			inputPath := getString(pages[i], "inputPath")
			outputPath := getString(pages[i], "outputPath")
			updated := getString(pages[i], "updated")
			term.ClrToEOL()
			term.Printf("%4d %s%s%s\n\t%q %s %s %s %s\n",
				i+1, termlib.Bold+termlib.Italic, inputPath, termlib.Reset, outputPath,updated)
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
		case strings.HasPrefix(answer, "a"):
			// Add a page
			if err = addPage(scanner, options, pages, collection); err != nil {
				displayErrorStatus("%s", err)
				continue
			}
			pages, err = listPages(collection, []string{})
			if err != nil {
				displayErrorStatus("%s", err)
				continue
			}
			tot = len(pages)
		case strings.HasPrefix(answer, "d"):
			// Remove a page
			if err = delPage(scanner, options, pages, collection); err != nil {
				displayErrorStatus("%s", err)
				continue
			}
			pages, err = listPages(collection, []string{})
			if err != nil {
				displayErrorStatus("%s", err)
				continue
			}
			tot = len(pages)			
		case strings.HasPrefix(answer, "h"):
			helpCuratePages(scanner)
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

/**
 * Manage a collection
 */

// helpCurateCollection explains how a collection works and what can be curated
func helpCurateCollection(scanner *bufio.Scanner) {
	term.Clear()
	defer term.Clear()
	term.ResetStyle()
	term.Printf(`
%sCurate collection. 

Collections contain three content types.

page
: Markdown expressing the contents of an HTML page. These are listed
in the generated sitemap but not list in any feeds.

post
: Markdown expressing a post (e.g. blog post) that will be rendered as
an HTML page and a feed item. These are also included in the sitemap.

item
: Items can hold a post or an RSS (feed) item harvested from the web. These are
listed in the aggregation pages but not explicitly included in the sitemap because
they may point to another web resource.

Each content type is managed in its own list. 
%s%s

Actions:

[pa]ges, curate pages or generates their HTML representations and
         updates the sitemap
[po]sts, curates posts in a collection or generates their HTML representations,
        updates the RSS 2.0 feeds and sitemap.
[i]tems, curated aggregated items and generates the aggregated HTML page,
         RSS 2.0 file, OPML file and updates the sitemap.

[h]elp
: Display this help

[q]uit
: To quit

Press enter to exit help.
`, termlib.Cyan, termlib.Italic, termlib.Reset)
	term.Refresh()
	scanner.Scan()
}

// curateCollection, curates a collections pages, posts and items
func curateCollection(scanner *bufio.Scanner, cfgName string, cfg *AppConfig, collection *Collection) error {
	term.Clear()
	defer term.Clear()
	for quit := false; quit == false; {
		term.Printf(`
Curate %s:

	[pa]ges
	[po]osts
	[i]tems

`, collection.File)
		term.ResetStyle()
		term.Printf("\n([h]elp or [q]uit): ")
		term.ClrToEOL()
		term.Refresh()
		// Read entry
		if !scanner.Scan() {
			continue
		}
		answer, _, _ := parseAnswer(scanner.Text())
		answer = strings.ToLower(answer)
		switch {
		case strings.HasPrefix(answer, "pa"):
			if err := curatePages(scanner, collection); err != nil {
				displayErrorStatus("%q", err)
				continue
			}
		case strings.HasPrefix(answer, "po"):
			// Curate posts
			if err := curatePosts(scanner, cfgName, cfg, collection); err != nil {
				displayErrorStatus("%q", err)
				continue
			}
		case strings.HasPrefix(answer, "i"):	
			// Curate items
			if err := curateItems(scanner, collection); err != nil {
				displayErrorStatus("%q", err)
				continue
			}
		case strings.HasPrefix(answer, "h"):
			helpCurateCollection(scanner)
			continue
		case strings.HasPrefix(answer, "q"):
			quit = true
		}
	
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

[ha]rvest [NAME|NUMBER]
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


// curateCollections provides the interaction loop for curating collections.
func curateCollections(scanner *bufio.Scanner, cfgName string, cfg *AppConfig) error {
	term.Clear()
	defer term.Clear()
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
		answer = strings.ToLower(answer)
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
		case strings.HasPrefix(answer, "ha"):
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
				if err := curateCollection(scanner, cfgName, cfg,  cfg.Collections[val - 1]); err != nil {
					displayErrorStatus("%s", err)
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

/**
 * Action method
 */


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

