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

// IncludeTextBlock takes a text string and replaces the text blocks
// based on the file path included in the line.
//
// Parameters:
//   text (string): the text to be transformed
// Returns:
//   string: the transformed text
func IncludeTextBlock(text string) string {
	// Find the include-text-block directive in the page.
	insertBlockRegExp := regexp.MustCompile(`\s+@include-text-block\s+([^\s]+)(?:\s+(\w+))?`)
	// Insert the text blocks
	return insertBlockRegExp.ReplaceAllStringFunc(text, replaceTextBlock)
}

// replaceTextBlock does the actual replacement work with the result
// of the matched RegExp.
//
// Parameters:
//   fullMatch (string): the full matched string
// Returns:
//   string: the replacement string
func replaceTextBlock(fullMatch string) string {
	// Extract filePath from the matched string
	matches := regexp.MustCompile(`\s+@include-text-block\s+([^\s]+)(?:\s+(\w+))?`).FindStringSubmatch(fullMatch)
	if len(matches) < 2 {
		return fullMatch // retur(src); errn original if no match
	}
	filePath := matches[1]

	src, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error inserting text block from %s: %v\n", filePath, err)
		return fullMatch
	}
	var fileContent string
	// Split off front matter before including content
	doc := &CommonMark{}
	if err := doc.Parse(src); err == nil {
		fileContent = "\n" + doc.Text + "\n"
	} else {
		fileContent = fmt.Sprintf("\n%s\n", src)
	}
	return fileContent
}
