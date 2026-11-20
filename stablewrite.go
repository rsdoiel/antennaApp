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
	"bytes"
	"os"
	"regexp"
	"time"
)

// clock is where every build-time stamp in generated output comes from. It is
// time.Now; tests replace it so a stamp differs from run to run without anyone
// having to wait.
var clock = time.Now

// Build stamps are the lines a generator writes to say when it ran. They are
// the only part of a regenerated file that changes when nothing else did, so a
// file that differs from its old self by only these is left alone. Each matches
// the exact form antenna writes, not any similar text: a page's own `date:`
// front matter is a <meta name="date"> too, but written without the space
// before > that the build stamp has, so it is never mistaken for one. Only the
// first match in a file counts, which is where the generator put it.
var (
	// htmlStamps is <meta name="date" content="..." >, in the attribute order of
	// today's build and of the builds before attributes had a fixed order.
	htmlStamps = []*regexp.Regexp{
		regexp.MustCompile(`(?m)^[ \t]*<meta name="date" content="[^"]*" >[ \t]*\r?\n`),
		regexp.MustCompile(`(?m)^[ \t]*<meta content="[^"]*" name="date" >[ \t]*\r?\n`),
	}
	// itemsMDStamps is the dateModified line of an items file's front matter.
	itemsMDStamps = []*regexp.Regexp{regexp.MustCompile(`(?m)^dateModified: [^\n]*\n`)}
	// rssStamps is the channel's lastBuildDate.
	rssStamps = []*regexp.Regexp{regexp.MustCompile(`<lastBuildDate>[^<]*</lastBuildDate>`)}
	// opmlStamps is the head's dateCreated.
	opmlStamps = []*regexp.Regexp{regexp.MustCompile(`<dateCreated>[^<]*</dateCreated>`)}
	// tocStamps is the table of contents' dateModified: when its content last changed.
	tocStamps = []*regexp.Regexp{regexp.MustCompile(`<dateModified>[^<]*</dateModified>`)}
)

// stripStamps returns b without the first match of each stamp pattern.
func stripStamps(b []byte, stamps []*regexp.Regexp) []byte {
	out := b
	for _, re := range stamps {
		if loc := re.FindIndex(out); loc != nil {
			out = append(append([]byte{}, out[:loc[0]]...), out[loc[1]:]...)
		}
	}
	return out
}

/** writeStable writes generated output to path, unless the file already holds
 * the same thing. "The same" means identical once the build-stamp lines are
 * removed from both: regenerating a site then rewrites only the files whose
 * content really changed, so version control shows only real changes and a
 * deploy tool that goes by modification time copies only those. A file that
 * differs in anything else is rewritten, with its new stamp.
 *
 * Parameters:
 *   path (string) — the output file
 *   data ([]byte) — the newly generated content
 *   perm (os.FileMode) — the mode if the file has to be created
 *   stamps (...*regexp.Regexp) — what counts as a build stamp for this kind of file
 *
 * Returns:
 *   bool — true when the file was written, false when it was left as it was
 *   error — from reading or writing the file
 *
 * Example:
 *   written, err := writeStable("news.html", html, 0o666, htmlStamps...)
 */
func writeStable(path string, data []byte, perm os.FileMode, stamps ...*regexp.Regexp) (bool, error) {
	if old, err := os.ReadFile(path); err == nil {
		if bytes.Equal(stripStamps(old, stamps), stripStamps(data, stamps)) {
			return false, nil
		}
	}
	if err := os.WriteFile(path, data, perm); err != nil {
		return false, err
	}
	return true, nil
}
