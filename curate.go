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
	"strconv"
	"strings"
)

// CurateCollection displays items in a collection so you can select items for publication
func curateItems(collection *Collection) error {
	return fmt.Errorf("curateItems not implemented")
}

func clearScreen() {
	// ANSI escape code to clear the screen and move the cursor to the top-left
	fmt.Print("\033[H\033[2J")
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
func helpCollectionMenu() {
	clearScreen()
	fmt.Printf(`

Collection menu options. Commands have the following form.

  MENU_NUMBER ENTER_KEY
  ACTION [PARAMETERS] ENTER_KEY

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
`)
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
	clearScreen()
	for quit := false; quit == false; {
		// - A List collections
		fmt.Printf("Curate Collections\n\n")
		for i, col := range cfg.Collections {
			fmt.Printf("\t%2d: %s, %s\n", i + 1, col.File, col.Title)
		}
		// FIXME: if there is only one collection, jump into it to curate items
		fmt.Printf("\n(q to quit, h for help): \n")
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
		case (answer == "q" || answer == "quit"):
			quit = true
		case (answer == "n" || answer == "new"):
			if err := createCollection(params); err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				continue
			}
		case (answer == "e" || answer == "edit"):
			if err := editCollection(params); err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				continue
			}
		case (answer == "r" || answer == "remove" ):
			if err := removeCollection(params); err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				continue
			}
		case (answer == "h" || answer == "help"):
			helpCollectionMenu()
			continue
		default:
			val, err := strconv.Atoi(answer)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%q\n", err)
				continue
			}
			if (val < 1) || (val > colCount) {
				fmt.Fprintf(os.Stderr, "enter a number between 1 and %d or enter q to quit\n", colCount)
				continue
			}
			// calc collection number to curate
			i := val - 1
			if err := curateItems(cfg.Collections[i]); err != nil {
				fmt.Fprintf(os.Stderr, "%q\n", err)
				continue
			}
		}
		clearScreen()
	}
	return nil
}
