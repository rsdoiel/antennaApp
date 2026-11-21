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
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
)

// Sequence is one entry of the sequences list in antenna.yaml (DR-0026): a
// path that holds a sequence of pages or posts. A path ending in "/" is a
// directory whose Markdown files are the members, in Order (name, date or
// date-desc; name by default). A path ending in ".md" is a list file whose
// site-local links to published Markdown are the members, in document order.
type Sequence struct {
	Path  string `json:"path" yaml:"path"`
	Order string `json:"order,omitempty" yaml:"order,omitempty"`
}

// isDir reports whether the sequence names a directory.
func (s Sequence) isDir() bool { return strings.HasSuffix(s.Path, "/") }

// validateSequences checks the sequences list when antenna.yaml is read; each
// problem is a configuration error.
func (cfg *AppConfig) validateSequences() error {
	seen := map[string]bool{}
	for i, s := range cfg.Sequences {
		p := strings.TrimSpace(s.Path)
		if p == "" {
			return fmt.Errorf("sequences entry %d has no path", i+1)
		}
		isList := strings.HasSuffix(strings.ToLower(p), ".md")
		if !strings.HasSuffix(p, "/") && !isList {
			return fmt.Errorf("sequences entry %d: path %q must end in / (a directory) or .md (a list file)", i+1, p)
		}
		switch s.Order {
		case "":
		case "name", "date", "date-desc":
			if isList {
				return fmt.Errorf("sequences entry %d: a list file (%s) takes no order; its document order is the order", i+1, p)
			}
		default:
			return fmt.Errorf("sequences entry %d: order %q must be name, date or date-desc", i+1, s.Order)
		}
		clean := path.Clean(strings.TrimPrefix(filepath.ToSlash(p), "/"))
		if seen[clean] {
			return fmt.Errorf("sequences entry %d: path %q is listed twice", i+1, p)
		}
		seen[clean] = true
	}
	return nil
}

// sequenceInfo is what one sequence found: its members in order, whether its
// path exists, and the members another, earlier sequence already holds.
type sequenceInfo struct {
	Sequence Sequence
	Exists   bool
	Members  []string
	Overlaps []sequenceOverlap
}

// sequenceOverlap is a member that an earlier sequence already holds: it takes
// the neighbours of First (DR-0026 6).
type sequenceOverlap struct{ MD, First string }

/** sequenceMembers works out the members of every sequence in antenna.yaml, in
 * order, from the site's published Markdown (DR-0026 2, 3). A directory's
 * members are the pages and posts the table of contents lists directly in it,
 * other than index.md; a list file's are the site-local links in its body to
 * published Markdown, in document order, each once, leaving out the list file
 * itself and anything else.
 *
 * Parameters:
 *   htdocs (string) — the directory the site is published in
 *   listed ([]tocEntry) — the pages and posts the table of contents lists
 *
 * Returns:
 *   []sequenceInfo — one per sequence, in the order of antenna.yaml, each member
 *   in at most one sequence's Members (the first that holds it); Overlaps names
 *   what a later sequence also held, and the sequence that took it
 *
 * Example:
 *   infos := cfg.sequenceMembers("htdocs", append(pages, posts...))
 */
func (cfg *AppConfig) sequenceMembers(htdocs string, listed []tocEntry) []sequenceInfo {
	var out []sequenceInfo
	taken := map[string]string{} // member -> the first sequence that holds it
	for _, seq := range cfg.Sequences {
		info := sequenceInfo{Sequence: seq}
		root := filepath.Join(htdocs, filepath.FromSlash(strings.TrimPrefix(seq.Path, "/")))
		if st, err := os.Stat(root); err == nil && (st.IsDir() == seq.isDir()) {
			info.Exists = true
		}
		var members []string
		if info.Exists {
			if seq.isDir() {
				members = directoryMembers(htdocs, seq, listed)
			} else {
				members = listFileMembers(htdocs, path.Clean(strings.TrimPrefix(seq.Path, "/")))
			}
		}
		for _, m := range members {
			if first, ok := taken[m]; ok {
				info.Overlaps = append(info.Overlaps, sequenceOverlap{MD: m, First: first})
				continue
			}
			taken[m] = seq.Path
			info.Members = append(info.Members, m)
		}
		out = append(out, info)
	}
	return out
}

// directoryMembers are the listed Markdown files directly in the directory,
// except index.md, in the sequence's order.
func directoryMembers(htdocs string, seq Sequence, listed []tocEntry) []string {
	dir := path.Clean(strings.TrimPrefix(seq.Path, "/"))
	type member struct {
		md   string
		when time.Time
	}
	var ms []member
	seen := map[string]bool{}
	for _, e := range listed {
		if path.Dir(e.md) != dir || path.Base(e.md) == "index.md" || seen[e.md] {
			continue
		}
		seen[e.md] = true
		ms = append(ms, member{md: e.md, when: markdownDate(htdocs, e)})
	}
	sort.SliceStable(ms, func(i, j int) bool {
		switch seq.Order {
		case "date":
			if !ms[i].when.Equal(ms[j].when) {
				return ms[i].when.Before(ms[j].when)
			}
		case "date-desc":
			if !ms[i].when.Equal(ms[j].when) {
				return ms[i].when.After(ms[j].when)
			}
		}
		return ms[i].md < ms[j].md
	})
	out := make([]string, 0, len(ms))
	for _, m := range ms {
		out = append(out, m.md)
	}
	return out
}

// markdownDate is the date of a listed document: pubDate, else datePublished
// from its published front matter, else the date the listing carries.
func markdownDate(htdocs string, e tocEntry) time.Time {
	if b, err := os.ReadFile(filepath.Join(htdocs, filepath.FromSlash(e.md))); err == nil {
		if fm, _, err := SplitFrontMatter(b); err == nil {
			doc := &CommonMark{FrontMatter: fm}
			for _, key := range []string{"pubDate", "datePublished"} {
				if stored, found, ok := doc.GetAttributeDate(key); found && ok {
					if t, ok := parseStoredDate(stored); ok {
						return t
					}
				}
			}
		}
	}
	return e.when
}

// listFileMembers are the site-local links in a list file's body to published
// Markdown, in document order, once each, without the list file itself.
func listFileMembers(htdocs, listPath string) []string {
	raw, err := os.ReadFile(filepath.Join(htdocs, filepath.FromSlash(listPath)))
	if err != nil {
		return nil
	}
	_, body, err := SplitFrontMatter(raw)
	if err != nil {
		body = string(raw)
	}
	src := []byte(body)
	root := goldmark.New(goldmark.WithExtensions(extension.GFM, extension.Footnote, extension.DefinitionList)).Parser().Parse(text.NewReader(src))
	var out []string
	seen := map[string]bool{listPath: true}
	_ = ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		link, ok := n.(*ast.Link)
		if !ok || !entering {
			return ast.WalkContinue, nil
		}
		u, err := url.Parse(string(link.Destination))
		if err != nil || u.Scheme != "" || u.Host != "" || u.Path == "" {
			return ast.WalkContinue, nil
		}
		var md string
		if strings.HasPrefix(u.Path, "/") {
			md = path.Clean(strings.TrimPrefix(u.Path, "/"))
		} else {
			md = path.Join(path.Dir(listPath), u.Path)
		}
		if md == ".." || strings.HasPrefix(md, "../") || seen[md] || markdownPathFor(md) != md || !published(htdocs, md) {
			return ast.WalkContinue, nil
		}
		seen[md] = true
		out = append(out, md)
		return ast.WalkContinue, nil
	})
	return out
}

// sequenceNeighbours turns the members of every sequence into previous and
// next, by the Markdown path of each member (DR-0026 4, 5).
func (cfg *AppConfig) sequenceNeighbours(htdocs string, listed []tocEntry) tocNav {
	nav := tocNav{}
	for _, info := range cfg.sequenceMembers(htdocs, listed) {
		for i, md := range info.Members {
			var n tocNeighbours
			if i > 0 {
				n.prev = info.Members[i-1]
			}
			if i+1 < len(info.Members) {
				n.next = info.Members[i+1]
			}
			nav[md] = n
		}
	}
	return nav
}
