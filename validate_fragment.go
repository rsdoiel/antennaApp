package antennaApp

import (
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rsdoiel/markdownweb/slug"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
)

// explicitAnchor matches an anchor an author wrote by hand: an id or name
// attribute in raw HTML, or a {#id} heading attribute.
var explicitAnchor = regexp.MustCompile(`(?i)\b(?:id|name)\s*=\s*["']([^"']+)["']|\{#([A-Za-z][\w:.-]*)\}`)

/** headingAnchors returns the fragments that a Markdown body offers: the slug
 * of each heading (markdownweb's slug rule, repeats numbered) and any id or
 * name an author wrote in the text. Headings in code blocks are not headings.
 * Fragments are lower case, since a reader is not strict about case.
 *
 * Parameters:
 *   body (string) — the Markdown, front matter already removed
 *
 * Returns:
 *   map[string]bool — the set of lower-case fragments
 *
 * Example:
 *   a := headingAnchors("# Soil & Seeds\n") // {"soil--seeds": true}
 */
func headingAnchors(body string) map[string]bool {
	src := []byte(body)
	root := goldmark.New(goldmark.WithExtensions(extension.GFM, extension.Footnote, extension.DefinitionList)).Parser().Parse(text.NewReader(src))
	out := map[string]bool{}
	slugs := slug.New()
	_ = ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		h, ok := n.(*ast.Heading)
		if !ok || !entering {
			return ast.WalkContinue, nil
		}
		var b strings.Builder
		_ = ast.Walk(h, func(c ast.Node, in bool) (ast.WalkStatus, error) {
			if !in {
				return ast.WalkContinue, nil
			}
			switch t := c.(type) {
			case *ast.Text:
				b.Write(t.Segment.Value(src))
				if t.SoftLineBreak() || t.HardLineBreak() {
					b.WriteByte(' ')
				}
			case *ast.String:
				b.Write(t.Value)
			}
			return ast.WalkContinue, nil
		})
		if s := slugs.Next(b.String()); s != "" {
			out[strings.ToLower(s)] = true
		}
		return ast.WalkSkipChildren, nil
	})
	for _, m := range explicitAnchor.FindAllStringSubmatch(body, -1) {
		out[strings.ToLower(m[1]+m[2])] = true
	}
	return out
}

// anchorsOf reads the Markdown file at the disk path and returns its anchors,
// cached for the run. ok is false when the file cannot be read or its front
// matter does not parse, so there is nothing to check against.
func (v *validator) anchorsOf(disk string) (map[string]bool, bool) {
	if a, seen := v.anchorCache[disk]; seen {
		return a, a != nil
	}
	if v.anchorCache == nil {
		v.anchorCache = map[string]map[string]bool{}
	}
	var a map[string]bool
	if raw, err := os.ReadFile(disk); err == nil {
		if _, body, err := SplitFrontMatter(raw); err == nil {
			a = headingAnchors(body)
		}
	}
	v.anchorCache[disk] = a
	return a, a != nil
}

// checkFragment reports a link to a heading that the page it names does not
// have (link-fragment). It looks at a link to a place in the same file
// (#heading) and a relative link to another Markdown file that exists
// (other.md#heading); anything else, such as an HTML page, a missing file or
// an outside address, has no headings to read. #top always works.
func (v *validator) checkFragment(file, dest string, u *url.URL, line int) {
	frag := strings.ToLower(u.Fragment)
	if frag == "" || frag == "top" || u.Scheme != "" || u.Host != "" {
		return
	}
	disk := filepath.FromSlash(file)
	if u.Path != "" {
		var site string
		if strings.HasPrefix(u.Path, "/") {
			site = path.Clean(strings.TrimPrefix(u.Path, "/"))
		} else {
			site = path.Join(siteDir(file), u.Path)
		}
		if ext := strings.ToLower(path.Ext(site)); ext != ".md" && ext != ".markdown" || site == ".." || strings.HasPrefix(site, "../") {
			return
		}
		disk = v.diskPathInSite(site)
		if disk == "" {
			return // link-missing reports it
		}
	}
	anchors, ok := v.anchorsOf(disk)
	if !ok || anchors[frag] {
		return
	}
	v.add("link-fragment", SeverityWarning, file, line, "%s: no heading makes the fragment #%s", dest, u.Fragment)
}

// diskPathInSite returns where a site-relative path is on disk, in the working
// directory or htdocs, or "" when it is in neither.
func (v *validator) diskPathInSite(site string) string {
	roots := []string{"."}
	if v.cfg != nil && v.cfg.Htdocs != "" && v.cfg.Htdocs != "." {
		roots = append(roots, v.cfg.Htdocs)
	}
	for _, r := range roots {
		p := filepath.Join(r, filepath.FromSlash(site))
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p
		}
	}
	return ""
}
