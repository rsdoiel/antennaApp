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
	"io/ioutil"
	"regexp"
)

// IncludeCodeBlock takes a text string and replaces the code blocks
// based on the file path included in the line and the language name.
// The generated code block uses the `~~~` sequence to delimit the block
// with the language name provided in the opening delimiter.
//
// Parameters:
//
//	text (string): the text to be transformed
//
// Returns:
//
//	string: the transformed text
func IncludeCodeBlock(text string) string {
	// Find the include-code-block directive in the page.
	insertBlockRegExp := regexp.MustCompile(`\s+@include-code-block\s+([^\s]+)(?:\s+(\w+))?`)
	// Insert the code blocks
	return insertBlockRegExp.ReplaceAllStringFunc(text, replaceCodeBlock)
}

// replaceCodeBlock does the actual replacement work with the result
// of the matched RegExp.
//
// Parameters:
//
//	fullMatch (string): the full matched string
//
// Returns:
//
//	string: the replacement string
func replaceCodeBlock(fullMatch string) string {
	// Extract filePath and language from the matched string
	matches := regexp.MustCompile(`@include-code-block\s+([^\s]+)(?:\s+(\w+))?`).FindStringSubmatch(fullMatch)
	if len(matches) < 2 {
		return fullMatch // return original if no match
	}
	filePath := matches[1]
	language := ""
	if len(matches) > 2 {
		language = matches[2]
	}

	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error inserting code block from %s: %v\n", filePath, err)
		return fullMatch
	}

	return fmt.Sprintf("~~~%s\n%s\n~~~", language, string(fileContent))
}
