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
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/
package antennaApp

import (
	"fmt"
	"io"
	"os"

	"github.com/rsdoiel/frontmatter"
)

/** expandIncludesReport expands the @include-text-block and @include-code-block
 * directives in text (DR-0024 12). A text block is itself expanded, so its own
 * directives are followed; a code block is code and is never expanded. The work is frontmatter.Includes with
 * frontmatter.Markdown (frontmatter DR-0002). Paths
 * are read from the working directory (the site root); an absolute path or one
 * with .. is refused and left as written. A file that cannot be read is left
 * as written and reported to eout. An included file's front matter is dropped.
 *
 * Nesting deeper than three levels draws a warning on eout and the expansion
 * goes on; deeper than five levels, or a file that includes itself directly or
 * through others, is a data error naming the chain.
 *
 * Parameters:
 *   text (string) — the document text
 *   eout (io.Writer) — warnings and the reasons a directive was left as written
 *
 * Returns:
 *   string — text with the directives expanded
 *   []string — every file a directive named, nested ones too, once each, in the
 *     order they were first met; a missing file is listed, a refused path is not
 *   error — a data error for a cycle or for nesting deeper than five levels
 *
 * Example:
 *   out, files, err := expandIncludesReport("Top.\n\n@include-text-block a.txt\n", os.Stderr)
 */
func expandIncludesReport(text string, eout io.Writer) (string, []string, error) {
	return expandIncludesWith(text, eout, frontmatter.Markdown)
}

// expandIncludesWith is expandIncludesReport for a format that may know fewer
// directives: IncludeTextBlock and IncludeCodeBlock use it to expand one kind.
func expandIncludesWith(text string, eout io.Writer, f frontmatter.Format) (string, []string, error) {
	out, inc, err := frontmatter.Includes(text, os.DirFS("."), f)
	for _, w := range inc.Warnings {
		fmt.Fprintln(eout, w)
	}
	if err != nil {
		return text, inc.Files, dataErrorf("%w", err)
	}
	return out, inc.Files, nil
}
