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
	"os"

	"github.com/rsdoiel/frontmatter"
)

/** IncludeCodeBlock expands the @include-code-block directives in text, and only
 * those. It is kept for callers of the package; antenna itself uses
 * expandIncludesReport, which expands both kinds, follows nested text includes
 * and reports what was included (DR-0024 12). A cycle, or nesting deeper than
 * five levels, leaves the text as it was and reports the problem on standard
 * error.
 *
 * Parameters:
 *   text (string) — the text to be transformed
 *
 * Returns:
 *   string — the transformed text
 *
 * Example:
 *   out := IncludeCodeBlock("Top.\n\n@include-code-block file.txt\n")
 */
func IncludeCodeBlock(text string) string {
	f := frontmatter.Markdown
	f.TextDirective = ""
	out, _, err := expandIncludesWith(text, os.Stderr, f)
	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		return text
	}
	return out
}
