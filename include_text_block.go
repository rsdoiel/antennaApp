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
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// IncludeTextBlock takes a text string and replaces the text blocks
// based on the file path included in the line.
//
// Parameters:
//
//	text (string): the text to be transformed
//
// Returns:
//
//	string: the transformed text
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
//
//	fullMatch (string): the full matched string
//
// Returns:
//
//	string: the replacement string
func replaceTextBlock(fullMatch string) string {
	// Extract filePath from the matched string
	matches := regexp.MustCompile(`\s+@include-text-block\s+([^\s]+)(?:\s+(\w+))?`).FindStringSubmatch(fullMatch)
	if len(matches) < 2 {
		return fullMatch // return original if no match
	}
	filePath := matches[1]

	// Security: Prevent path traversal attacks
	// Clean the path and ensure it's relative and safe
	cleanPath := filepath.Clean(filePath)
	if filepath.IsAbs(cleanPath) || strings.HasPrefix(cleanPath, "..") {
		fmt.Printf("Error: include-text-block path '%s' attempts directory traversal, using original text\n", filePath)
		return fullMatch
	}

	// Check if file exists and is a regular file (not directory or symlink)
	fileInfo, err := os.Stat(cleanPath)
	if err != nil {
		fmt.Printf("Error inserting text block from %s: %v\n", cleanPath, err)
		return fullMatch
	}
	if fileInfo.IsDir() {
		fmt.Printf("Error: include-text-block path '%s' is a directory, not a file\n", cleanPath)
		return fullMatch
	}

	src, err := ioutil.ReadFile(cleanPath)
	if err != nil {
		fmt.Printf("Error reading text block from %s: %v\n", cleanPath, err)
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
