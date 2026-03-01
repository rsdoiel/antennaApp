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
	"strings"
)

var (
	// Help and Menus for related actions
	collectionMenuHelp = fmt.Sprintf(`
Collection Menu Help

   (optional parameters can be completed using resulting prompts)

%sitems%s
: List the item in collection (aggregated items, not posts)

%sposts%s
: List posts in collection (this menu will let you add and delete posts)

%spages%s
: List pages in collection (this menu will let you add and delete posts)

With in each of these menu actions you can add, remove or curate the items.

`, 
	Yellow+Bold, Reset,
	Yellow+Bold, Reset,
	Yellow+Bold, Reset)

	collectionMenu = fmt.Sprintf(`
Collection Menu - %%s

    %sitems%s
        (list collection items)

    %sposts%s
        (list collection posts)

    %spages%s
        (list collection pages)

`,	
	Yellow+Bold, Reset,
	Yellow+Bold, Reset,
	Yellow+Bold, Reset)

 )


/**
 * Collection methods
 */

// helpCurateCollection explains how the options in the collection menu
func helpCurateCollection(scanner *bufio.Scanner) {
	term.Clear()
	defer term.Clear()
	term.ResetStyle()
	term.Printf(collectionMenuHelp)
	term.Printf("\n(%sq%suit, to exit help): ", Green, Reset)
	term.Refresh()
	scanner.Scan()
}

// curateCollection provides the interaction loop for curating a single collection.
func curateCollection(scanner *bufio.Scanner, cName string, cfgName string, cfg *AppConfig) error {
	term.Clear()
	defer term.Clear()
	for quit := false; quit == false; {
		term.Move(1, 1)
		term.ClrToEOL()
		// - list next actions
		term.Printf(collectionMenu, cName)
		term.Printf("\n(%sh%selp or %sq%suit): ", Green+Bold, Reset, Green+Bold, Reset)
		term.ClrToEOL()
		term.Refresh()
		
		// Read entry
		if !scanner.Scan() {
			continue
		}
		answer, _, err := parseAnswer(scanner.Text())
		if err != nil {
			displayErrorStatus("%s", err)
		}
		answer = strings.ToLower(answer)
		switch {
		case strings.HasPrefix(answer, "i"):
			// curate an item
			displayErrorStatus("%q menu not implemented", answer)
			continue
		case strings.HasPrefix(answer,"po"):
			displayErrorStatus("%q menu not implemented", answer)
			continue
		case strings.HasPrefix(answer,"pa"):
			displayErrorStatus("%q menu not implemented", answer)
			continue
		case strings.HasPrefix(answer, "h"):
			helpCurateCollection(scanner)
			continue
		case strings.HasPrefix(answer, "q"):
			quit = true
		}
		term.Clear()
	}
	return nil
}

