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
)

// CurateCollection displays items in a collection so you can select items for publication
func (app *AntennaApp) CurateItems(collection *Collection) error {
	return fmt.Errorf("CutateItems not implemented")
}

func clearScreen() {
	// ANSI escape code to clear the screen and move the cursor to the top-left
	fmt.Print("\033[H\033[2J")
}

// Curate provides a simple terminal interface to curating feed items for publication in your Antenna site.
func (app *AntennaApp) Curate(cfgName string, args []string) error {
	cfg := &AppConfig{}
	if err := cfg.LoadConfig(cfgName); err != nil {
		return err
	}
	if cfg.Collections == nil || len(cfg.Collections) == 0 {
		return fmt.Errorf("no collections found in  %s", cfgName)
	}

	colCount := len(cfg.Collections)
	// If we only have the default collection, jump to curate the items in the collection.
	if colCount == 1 {
		return CurateItems(cfg.Collections[0])
	}
	scanner := bufio.NewScanner(os.Stdin)
	clearScreen()
	for quit := false; quit == false; {
		// - A List collections
		fmt.Printf("Collections to curate\n\n")
		for i, col := range cfg.Collections {
			fmt.Printf("\t%2d: %s\n", i + 1, col.Title)
		}
		// FIXME: if there is only one collection, jump into it to curate items
		fmt.Printf("\nenter collection number or q to quit\n")
		// Read entry
		if ! scanner.Scan() {
			continue
		}
		answer := scanner.Text()
		switch answer {
		case "q":
			quit = true
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
			if err := app.CurateItems(cfg.Collections[i]); err != nil {
				fmt.Fprintf(os.Stderr, "%q\n", err)
				continue
			}
		}
		clearScreen()
	}
	return nil
}
