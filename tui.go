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
	"unicode"

	// My packages
	"github.com/rsdoiel/termlib"

	// 3rd Party
	//_ "github.com/glebarez/go-sqlite"
)

/**
 * TUI
 */

const (
	Reset = termlib.Reset
	Bold = termlib.Bold
	Italic = termlib.Italic
	Black = termlib.Black
	Red        = termlib.Red
	Green      = termlib.Green
	Yellow     = termlib.Yellow
	Blue       = termlib.Blue
	Magenta    = termlib.Magenta
	Cyan       = termlib.Cyan
	White      = termlib.White
	BlackBg    = termlib.BlackBg
	RedBg      = termlib.RedBg
	GreenBg    = termlib.GreenBg
	YellowBg   = termlib.YellowBg
	BlueBg     = termlib.BlueBg
	MagentaBg  = termlib.MagentaBg
	CyanBg     = termlib.CyanBg
	WhiteBg    = termlib.WhiteBg
)

var (
 	// term holds the TUI object
 	term = termlib.New(os.Stdout) 
)

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
 * Action method
 */

// TUI provides a simple terminal interface to curating collections and 
// feed items for publication in your Antenna site.
func (app *AntennaApp) TUI(cfgName string, args []string) error {
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

