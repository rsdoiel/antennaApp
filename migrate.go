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
	"github.com/rsdoiel/frontmatter"
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
)

// linkEdit replaces the bytes [start, end) of a file with repl: the
// destination of one link.
type linkEdit struct {
	start, end int
	repl       string
}

// leftLink is a link to a page antenna renders that could not be migrated.
type leftLink struct {
	line int
	dest string
}

// definitionRE finds a link reference definition at the start of a line:
// "[label]: destination". Footnote labels ([^1]) are excluded by the caller.
var definitionRE = regexp.MustCompile(`(?m)^( {0,3}\[)([^\]\n]+)(\]:[ \t]*)(<[^>\n]*>|\S+)`)

/** MigrateLinks implements `antenna migrate-links [-w] [PATH]` (DR-0017 7-9):
 * in the author's Markdown it rewrites a relative link to an .html page that
 * antenna renders from Markdown into the link to the .md source, the migration
 * `validate` reports as link-html. By default it is a dry run that prints a
 * unified diff and writes nothing; -w writes. Only the bytes of the destinations
 * change. Code spans, code blocks, images, absolute addresses and other
 * extensions are never touched. It does not run generate.
 *
 * Parameters:
 *   out     (io.Writer) — the diff and the summary
 *   eout    (io.Writer) — unused
 *   cfgName (string)    — the configuration file
 *   args    ([]string)  — -w and at most one PATH
 *
 * Returns:
 *   error — nil when it ran (a dry run that finds changes, a write, or nothing
 *           to migrate); usage (2), no input (66), data (65) for a file it
 *           cannot parse, io (74) for a failed write
 *
 * Example:
 *   err := app.MigrateLinks(os.Stdout, os.Stderr, "antenna.yaml", []string{"-w", "blog"})
 */
func (app *AntennaApp) MigrateLinks(out io.Writer, eout io.Writer, cfgName string, args []string) error {
	writeFiles, target := false, ""
	paths := []string{}
	for _, a := range args {
		switch a {
		case "-w", "--write":
			writeFiles = true
		default:
			if strings.HasPrefix(a, "-") && len(a) > 1 {
				return usageErrorf("unknown option %q; try 'antenna help migrate-links'", a)
			}
			paths = append(paths, a)
		}
	}
	if len(paths) > 1 {
		return usageErrorf("expected at most one PATH, got %d; try 'antenna help migrate-links'", len(paths))
	}
	if len(paths) == 1 {
		target = paths[0]
	}
	var files []string
	isDir := false
	if target != "" {
		info, err := os.Stat(target)
		if err != nil {
			return noInputf("%s: %w", target, err)
		}
		isDir = info.IsDir()
	}
	cfg := &AppConfig{}
	if err := cfg.LoadConfig(cfgName); err != nil {
		return err
	}
	targets, err := cfg.LinkTargets()
	if err != nil {
		return err
	}
	if target == "" {
		files = siteMarkdownFiles(targets)
	} else if files, err = markdownUnder(target, isDir); err != nil {
		return err
	}
	v := &validator{cfg: cfg, targets: targets, cfgName: cfgName}
	byHTML := v.htmlToMarkdown()

	var firstErr error
	fail := func(err error) {
		if firstErr == nil {
			firstErr = err
		}
	}
	links, changedFiles := 0, 0
	for _, file := range files {
		raw, err := os.ReadFile(filepath.FromSlash(file))
		if err != nil {
			fail(ioErrorf("%s: %w", file, err))
			continue
		}
		fm, _, err := SplitFrontMatter(raw)
		if err != nil {
			fail(dataErrorf("%s: cannot be parsed: %w", file, err))
			continue
		}
		if _, generated := fm["aggregateOf"]; generated {
			continue
		}
		edits, lefts := v.linkEdits(file, raw, byHTML)
		for _, l := range lefts {
			fmt.Fprintf(out, "%s:%d: left: %s is a page antenna renders, but it has no Markdown to link to yet\n", file, l.line, l.dest)
		}
		if len(edits) == 0 {
			continue
		}
		links += len(edits)
		changedFiles++
		updated := applyEdits(raw, edits)
		if !writeFiles {
			writeDiff(out, file, raw, updated)
			continue
		}
		if err := replaceFile(filepath.FromSlash(file), updated); err != nil {
			fail(ioErrorf("%s: %w", file, err))
			changedFiles--
			links -= len(edits)
		}
	}
	switch {
	case links == 0:
		fmt.Fprintln(out, "nothing to migrate")
	case writeFiles:
		fmt.Fprintf(out, "%s changed\n", countLinks(links, changedFiles))
	default:
		fmt.Fprintf(out, "%s would change (run with -w to write)\n", countLinks(links, changedFiles))
	}
	return firstErr
}

// countLinks says "1 link in 1 file" or "8 links in 2 files".
func countLinks(links, files int) string {
	l, f := "links", "files"
	if links == 1 {
		l = "link"
	}
	if files == 1 {
		f = "file"
	}
	return fmt.Sprintf("%d %s in %d %s", links, l, files, f)
}

// bodyStartOffset returns the byte offset in raw where the Markdown body
// starts: after the closing --- line of the front matter, or 0 when there is
// none. A \r before the newline is allowed, so CRLF files work.
func bodyStartOffset(raw []byte) int {
	_, at, found, err := frontmatter.Block(raw)
	if err != nil || !found {
		return 0
	}
	return at
}

// protectedRanges returns the byte ranges of the body where a "](" is not a
// link: code blocks, code spans and raw HTML.
func protectedRanges(body []byte) [][2]int {
	md := goldmark.New(goldmark.WithExtensions(extension.GFM, extension.Footnote, extension.DefinitionList))
	root := md.Parser().Parse(text.NewReader(body))
	var ranges [][2]int
	lines := func(n ast.Node) {
		l := n.Lines()
		for i := 0; i < l.Len(); i++ {
			seg := l.At(i)
			ranges = append(ranges, [2]int{seg.Start, seg.Stop})
		}
	}
	_ = ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch v := n.(type) {
		case *ast.FencedCodeBlock, *ast.CodeBlock, *ast.HTMLBlock:
			lines(n)
			return ast.WalkSkipChildren, nil
		case *ast.CodeSpan:
			lo, hi := -1, -1
			for c := v.FirstChild(); c != nil; c = c.NextSibling() {
				if t, ok := c.(*ast.Text); ok {
					if lo < 0 || t.Segment.Start < lo {
						lo = t.Segment.Start
					}
					if t.Segment.Stop > hi {
						hi = t.Segment.Stop
					}
				}
			}
			if lo >= 0 {
				ranges = append(ranges, [2]int{lo, hi})
			}
			return ast.WalkSkipChildren, nil
		case *ast.RawHTML:
			for i := 0; i < v.Segments.Len(); i++ {
				seg := v.Segments.At(i)
				ranges = append(ranges, [2]int{seg.Start, seg.Stop})
			}
		}
		return ast.WalkContinue, nil
	})
	return ranges
}

// inRanges reports whether offset is inside any range.
func inRanges(ranges [][2]int, offset int) bool {
	for _, r := range ranges {
		if offset >= r[0] && offset < r[1] {
			return true
		}
	}
	return false
}

// escaped reports whether the byte at i is preceded by an odd number of backslashes.
func escaped(b []byte, i int) bool {
	n := 0
	for i--; i >= 0 && b[i] == '\\'; i-- {
		n++
	}
	return n%2 == 1
}

// openingBracket finds the "[" that opens the link text closed by the "]" at
// close, staying within the paragraph, or -1.
func openingBracket(b []byte, close int) int {
	depth := 1
	for i := close - 1; i >= 0; i-- {
		if b[i] == '\n' && i > 0 && b[i-1] == '\n' {
			return -1
		}
		if escaped(b, i) {
			continue
		}
		switch b[i] {
		case ']':
			depth++
		case '[':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// linkEdits finds the destinations in file that should change, and the links it
// leaves for want of a Markdown page to point at (DR-0017 7). Offsets are into
// raw.
func (v *validator) linkEdits(file string, raw []byte, byHTML map[string][]string) ([]linkEdit, []leftLink) {
	base := bodyStartOffset(raw)
	body := raw[base:]
	protected := protectedRanges(body)
	var edits []linkEdit
	var lefts []leftLink
	consider := func(start, end int) {
		dest := string(body[start:end])
		u, err := url.Parse(dest)
		if err != nil || u.Scheme != "" || u.Host != "" || u.Path == "" || strings.HasSuffix(u.Path, "/") {
			return
		}
		var site string
		if strings.HasPrefix(u.Path, "/") {
			site = path.Clean(strings.TrimPrefix(u.Path, "/"))
		} else {
			site = path.Join(siteDir(file), u.Path)
		}
		if site == ".." || strings.HasPrefix(site, "../") {
			return
		}
		if ext := strings.ToLower(path.Ext(site)); ext != ".html" && ext != ".htm" {
			return
		}
		repl, status := v.migrationFor(relativeToWorkDir(file), u, site, byHTML)
		line := bytes.Count(raw[:base+start], []byte("\n")) + 1
		switch status {
		case migRewrite:
			edits = append(edits, linkEdit{start: base + start, end: base + end, repl: repl})
		case migLeft:
			lefts = append(lefts, leftLink{line: line, dest: dest})
		}
	}
	// inline links: [text](destination "title")
	for i := 0; i+1 < len(body); i++ {
		if body[i] != ']' || body[i+1] != '(' || escaped(body, i) || inRanges(protected, i) {
			continue
		}
		open := openingBracket(body, i)
		if open < 0 || (open > 0 && body[open-1] == '!') {
			continue // no link text found, or an image
		}
		j := i + 2
		for j < len(body) && (body[j] == ' ' || body[j] == '\t') {
			j++
		}
		start, end := j, j
		if j < len(body) && body[j] == '<' {
			k := bytes.IndexAny(body[j+1:], ">\n")
			if k < 0 || body[j+1+k] != '>' {
				continue
			}
			start, end = j+1, j+1+k
		} else {
			depth := 0
			for end < len(body) {
				c := body[end]
				if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
					break
				}
				if c == '(' {
					depth++
				} else if c == ')' {
					if depth == 0 {
						break
					}
					depth--
				}
				end++
			}
		}
		if end > start {
			consider(start, end)
		}
	}
	// reference definitions: [label]: destination
	for _, m := range definitionRE.FindAllSubmatchIndex(body, -1) {
		if inRanges(protected, m[0]) || strings.HasPrefix(string(body[m[4]:m[5]]), "^") {
			continue
		}
		start, end := m[8], m[9]
		if start < end && body[start] == '<' && body[end-1] == '>' {
			start, end = start+1, end-1
		}
		if end > start {
			consider(start, end)
		}
	}
	sort.Slice(edits, func(i, j int) bool { return edits[i].start < edits[j].start })
	sort.Slice(lefts, func(i, j int) bool { return lefts[i].line < lefts[j].line })
	return edits, lefts
}

// applyEdits returns raw with each edit's bytes replaced; edits do not overlap.
func applyEdits(raw []byte, edits []linkEdit) []byte {
	var b bytes.Buffer
	pos := 0
	for _, e := range edits {
		b.Write(raw[pos:e.start])
		b.WriteString(e.repl)
		pos = e.end
	}
	b.Write(raw[pos:])
	return b.Bytes()
}

// writeDiff prints a unified diff with no context: one hunk for each changed line.
func writeDiff(out io.Writer, file string, oldRaw, newRaw []byte) {
	oldLines := strings.Split(string(oldRaw), "\n")
	newLines := strings.Split(string(newRaw), "\n")
	fmt.Fprintf(out, "--- %s\n+++ %s\n", file, file)
	for i := range oldLines {
		if i < len(newLines) && oldLines[i] != newLines[i] {
			fmt.Fprintf(out, "@@ -%d +%d @@\n-%s\n+%s\n", i+1, i+1,
				strings.TrimRight(oldLines[i], "\r"), strings.TrimRight(newLines[i], "\r"))
		}
	}
}

// replaceFile writes data over name by writing a temporary file beside it and
// renaming it, keeping the file's mode.
func replaceFile(name string, data []byte) error {
	info, err := os.Stat(name)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(name), ".migrate-links-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Chmod(tmpName, info.Mode().Perm()); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, name); err != nil {
		os.Remove(tmpName)
		return err
	}
	return nil
}
