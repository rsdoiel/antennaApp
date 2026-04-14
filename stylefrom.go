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
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rsdoiel/stylefrom"
)

/**
 * ExtractStyles implements the "stylefrom" action. It extracts CSS from a
 * LibreOffice Writer HTML export (.html, .htm) or ODF document (.odt, .ott)
 * and writes the result to a theme stylesheet.
 *
 * Usage:
 *
 *	antenna stylefrom INPUT_FILE [OUTPUT_PATH]
 *
 * INPUT_FILE must have a .html, .htm, .odt, or .ott extension.
 * OUTPUT_PATH defaults to "theme/style.css" when omitted.
 */
func (app *AntennaApp) ExtractStyles(out io.Writer, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing input file — usage: stylefrom INPUT_FILE [OUTPUT_PATH]")
	}
	inputFile := args[0]
	outputPath := filepath.Join("theme", "style.css")
	if len(args) > 1 {
		outputPath = args[1]
	}

	css, err := stylefrom.ExtractCSS(inputFile)
	if err != nil {
		return err
	}

	// Ensure the output directory exists.
	if dir := filepath.Dir(outputPath); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0775); err != nil {
			return fmt.Errorf("cannot create directory %q: %w", dir, err)
		}
	}

	if err := os.WriteFile(outputPath, []byte(css), 0664); err != nil {
		return fmt.Errorf("cannot write %q: %w", outputPath, err)
	}

	fmt.Fprintf(out, "wrote CSS to %s\n", outputPath)
	return nil
}
