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

The above menu choice will let you curate the choosen type collection
elements.

%sblogit%s %s[PATH_TO_MARKDOWN_FILE]%s
: Blogit will bring in an Markdown to use as a post and set it up in a 
blog style path with in this collection. Example PATH_TO_MARKDOWN_FILE
"~/Documents/on_the_march.md" and this would be copied into
"blog/2026/03/01/on_the_march.md" if it is March first of 2026. Then
when you choose the posts actions you should find it included.

`, 
	Yellow+Bold, Reset,
	Yellow+Bold, Reset,
	Yellow+Bold, Reset,
	Yellow+Bold, Reset, Cyan, Reset)

	collectionMenu = fmt.Sprintf(`
Collection Menu - %%s

    %sitems%s
        (list collection items)

    %sposts%s
        (list collection posts)

    %spages%s
        (list collection pages)

    %sblogit%s %s[PATH_TO_MARKDOWN_FILE]%s
`,	
	Yellow+Bold, Reset,
	Yellow+Bold, Reset,
	Yellow+Bold, Reset,
	Yellow+Bold, Reset,
	Cyan, Reset)

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
			continue
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
		case strings.HasPrefix(answer,"b"):
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

