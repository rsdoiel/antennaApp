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
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// pubDateLayout is the RSS 2.0 date form BonBon expects, with a numeric zone
// offset (DR-0016).
const pubDateLayout = "Mon, 02 Jan 2006 15:04:05 -0700"

/** firstHeadingText returns the plain text of the first level-1 heading of a
 * Markdown body, with inline markup removed, or "" when there is none.
 * Headings inside code blocks are not headings.
 *
 * Parameters:
 *   body (string) — Markdown without front matter
 *
 * Returns:
 *   string — the heading text, white space collapsed
 *
 * Example:
 *   firstHeadingText("# Hello *there*\n") // "Hello there"
 */
func firstHeadingText(body string) string {
	src := []byte(body)
	root := goldmark.New().Parser().Parse(text.NewReader(src))
	var found string
	_ = ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		h, ok := n.(*ast.Heading)
		if !ok || !entering || h.Level != 1 || found != "" {
			return ast.WalkContinue, nil
		}
		var b strings.Builder
		_ = ast.Walk(h, func(c ast.Node, in bool) (ast.WalkStatus, error) {
			if !in {
				return ast.WalkContinue, nil
			}
			switch v := c.(type) {
			case *ast.Text:
				b.Write(v.Segment.Value(src))
				if v.SoftLineBreak() || v.HardLineBreak() {
					b.WriteByte(' ')
				}
			case *ast.String:
				b.Write(v.Value)
			}
			return ast.WalkContinue, nil
		})
		found = strings.Join(strings.Fields(b.String()), " ")
		return ast.WalkSkipChildren, nil
	})
	return found
}

/** blogitFrontMatter adds what a published blog post needs to the copy of a
 * file that blogit made (DR-0016): a pubDate in the RSS 2.0 form with a numeric
 * offset when the file has neither pubDate nor datePublished (now, or noon
 * local on the given day), and a title from the first level-1 heading when it
 * has none. Existing values are never changed, nothing else is written, and a
 * file that needs nothing is not rewritten. A file that is not Markdown (.md or
 * .markdown) is left alone. When a
 * day was given and the file's own date is on another day, one warning is
 * written and the file's date stays.
 *
 * Parameters:
 *   fName    (string)    — the post's copy under blog/
 *   postDay  (time.Time) — the day the post is filed under, or now
 *   dayGiven (bool)      — true when POST_DATE was given
 *   now      (time.Time) — the current time
 *   warn     (io.Writer) — warnings
 *
 * Returns:
 *   error — the file could not be read, parsed or written
 *
 * Example:
 *   err := blogitFrontMatter("blog/2026/11/19/a.md", day, true, time.Now(), os.Stderr)
 */
func blogitFrontMatter(fName string, postDay time.Time, dayGiven bool, now time.Time, warn io.Writer) error {
	// Only Markdown text is edited; any other file is copied as it is.
	if ext := strings.ToLower(filepath.Ext(fName)); ext != ".md" && ext != ".markdown" {
		return nil
	}
	doc, err := LoadCommonMark(fName)
	if err != nil {
		return fmt.Errorf("failed to load %q, %w", fName, err)
	}
	if doc.FrontMatter == nil {
		doc.FrontMatter = map[string]interface{}{}
	}
	changed := false
	_, hasPub, _ := doc.GetAttributeDate("pubDate")
	_, hasDatePublished, _ := doc.GetAttributeDate("datePublished")
	switch {
	case !hasPub && !hasDatePublished:
		when := now
		if dayGiven {
			when = time.Date(postDay.Year(), postDay.Month(), postDay.Day(), 12, 0, 0, 0, time.Local)
		}
		doc.FrontMatter["pubDate"] = when.Format(pubDateLayout)
		changed = true
	case dayGiven && warn != nil:
		key := "pubDate"
		if !hasPub {
			key = "datePublished"
		}
		stored, _, _ := doc.GetAttributeDate(key)
		if t, ok := parseStoredDate(stored); ok {
			want := postDay.Format("2006-01-02")
			if t.UTC().Format("2006-01-02") != want && t.Local().Format("2006-01-02") != want {
				fmt.Fprintf(warn, "warning: %s: %s %q is not on POST_DATE %s; the file's date is kept and the post is filed under %s\n",
					fName, key, stored, want, want)
			}
		}
	}
	if strings.TrimSpace(doc.GetAttributeString("title", "")) == "" {
		if h := firstHeadingText(doc.Text); h != "" {
			doc.FrontMatter["title"] = h
			changed = true
		}
	}
	if !changed {
		return nil
	}
	return os.WriteFile(fName, []byte(doc.String()), 0o644)
}
