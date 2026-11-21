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
	"path/filepath"
	"regexp"
	"strings"
)

const (
	// includeWarnDepth is the nesting level above which an include draws a
	// warning; the page is level 0 and the files it includes are level 1.
	includeWarnDepth = 3
	// includeMaxDepth is the nesting level above which an include is an error.
	includeMaxDepth = 5
)

// includeDirective matches an include directive that starts its line (up to
// three spaces of indent), with the blank lines before it: @include-text-block
// FILE or @include-code-block FILE [LANGUAGE]. The language is on the same
// line; the end of the line ends the directive.
var includeDirective2 = regexp.MustCompile(`(?m)(?:\A|\s*\n)[ ]{0,3}@include-(text|code)-block[ \t]+(\S+)(?:[ \t]+(\w+))?[ \t]*$`)

/** expandIncludesReport expands the @include-text-block and @include-code-block
 * directives in text (DR-0024 12). A text block is itself expanded, so its own
 * directives are followed; a code block is code and is never expanded. Paths
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
	st := &includeState{eout: eout, seen: map[string]bool{}}
	out := st.expand(text, nil)
	return out, st.files, st.err
}

type includeState struct {
	eout  io.Writer
	only  string // "text" or "code" to expand one kind; empty for both
	files []string
	seen  map[string]bool
	err   error
}

// expand expands the directives of one document. chain holds the included
// files above it (empty for the page itself).
func (st *includeState) expand(text string, chain []string) string {
	if st.err != nil || !strings.Contains(text, "@include-") {
		return text
	}
	level := len(chain) + 1 // the level of the files this document includes
	return includeDirective2.ReplaceAllStringFunc(text, func(match string) string {
		if st.err != nil {
			return match
		}
		sub := includeDirective2.FindStringSubmatch(match)
		kind, name, lang := sub[1], sub[2], sub[3]
		if st.only != "" && st.only != kind {
			return match
		}
		lead := "\n\n"
		if !strings.Contains(match, "\n") {
			lead = "" // the directive is the first line of the document
		}
		clean := filepath.Clean(name)
		if filepath.IsAbs(clean) || strings.HasPrefix(clean, "..") {
			fmt.Fprintf(st.eout, "include-%s-block: %s attempts to leave the site, left as written\n", kind, name)
			return match
		}
		if !st.seen[clean] {
			st.seen[clean] = true
			st.files = append(st.files, clean)
		}
		info, err := os.Stat(clean)
		if err != nil {
			fmt.Fprintf(st.eout, "include-%s-block: %s: %v, left as written\n", kind, clean, err)
			return match
		}
		if info.IsDir() {
			fmt.Fprintf(st.eout, "include-%s-block: %s is a directory, left as written\n", kind, clean)
			return match
		}
		src, err := os.ReadFile(clean)
		if err != nil {
			fmt.Fprintf(st.eout, "include-%s-block: %s: %v, left as written\n", kind, clean, err)
			return match
		}
		if kind == "code" {
			return lead + "~~~" + lang + "\n" + strings.TrimRight(string(src), "\n") + "\n~~~"
		}
		for _, c := range chain {
			if c == clean {
				st.err = dataErrorf("include cycle: %s", strings.Join(append(append([]string{}, chain...), clean), " -> "))
				return match
			}
		}
		next := append(append([]string{}, chain...), clean)
		if level > includeMaxDepth {
			st.err = dataErrorf("includes nested more than %d levels deep: %s", includeMaxDepth, strings.Join(next, " -> "))
			return match
		}
		if level > includeWarnDepth {
			fmt.Fprintf(st.eout, "include nested %d levels deep (more than %d is an error): %s\n", level, includeMaxDepth, strings.Join(next, " -> "))
		}
		_, body, perr := SplitFrontMatter(src)
		if perr != nil {
			body = string(src)
		}
		return lead + strings.TrimRight(st.expand(body, next), "\n")
	})
}
