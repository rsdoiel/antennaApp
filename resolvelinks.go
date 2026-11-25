package antennaApp

import (
	"bytes"
	"fmt"
	"html"
	"net/url"
	"sort"
	"strings"

	xhtml "golang.org/x/net/html"
)

// DR-0037: a relative href or src in a harvested item's body is resolved
// against the item's link when the item is rendered. The stored item is not
// changed.

/** resolveRef resolves one reference against an item's address.
 *
 * A reference is left as written when it is empty, a #fragment only, has a
 * scheme (absolute URLs, mailto:, tel:, data:, javascript:), does not parse,
 * or when base is not an absolute http(s) address.
 *
 * Parameters:
 *   base (string) — the item's link
 *   ref (string) — an href or src value as written
 *
 * Returns:
 *   string — the absolute address, or ref unchanged
 *   bool — true when ref was resolved (changed)
 *
 * Example:
 *   abs, ok := resolveRef("https://example.org/2026/a/", "/images/x.png")
 *   // "https://example.org/images/x.png", true
 */
func resolveRef(base, ref string) (string, bool) {
	if !isRelativeRef(ref) {
		return ref, false
	}
	b, err := url.Parse(strings.TrimSpace(base))
	if err != nil || (b.Scheme != "http" && b.Scheme != "https") || b.Host == "" {
		return ref, false
	}
	u, err := url.Parse(strings.TrimSpace(ref))
	if err != nil {
		return ref, false
	}
	return b.ResolveReference(u).String(), true
}

// isRelativeRef reports whether ref is a reference that needs a base: not
// empty, not a #fragment alone, and with no scheme.
func isRelativeRef(ref string) bool {
	ref = strings.TrimSpace(ref)
	if ref == "" || strings.HasPrefix(ref, "#") {
		return false
	}
	u, err := url.Parse(ref)
	return err == nil && u.Scheme == ""
}

// refAttrs are the attributes of an HTML element that hold an address.
var refAttrs = map[string]bool{"href": true, "src": true, "poster": true}

// rewriteHTMLRefs walks the tags of s and passes each href, src and poster
// value to fn. A tag is rewritten only when fn changes a value; every other
// byte is kept as it was. It returns the new text and the number of values
// changed.
func rewriteHTMLRefs(s string, fn func(ref string) (string, bool)) (string, int) {
	z := xhtml.NewTokenizer(strings.NewReader(s))
	var out bytes.Buffer
	n := 0
	for {
		tt := z.Next()
		if tt == xhtml.ErrorToken {
			break
		}
		raw := z.Raw()
		if tt != xhtml.StartTagToken && tt != xhtml.SelfClosingTagToken {
			out.Write(raw)
			continue
		}
		tok := z.Token()
		changed := false
		for i, a := range tok.Attr {
			if a.Namespace != "" || !refAttrs[a.Key] {
				continue
			}
			if v, ok := fn(a.Val); ok {
				tok.Attr[i].Val = v
				changed = true
				n++
			}
		}
		if !changed {
			out.Write(raw)
			continue
		}
		out.WriteString("<" + tok.Data)
		for _, a := range tok.Attr {
			if a.Val == "" {
				out.WriteString(" " + a.Key)
				continue
			}
			out.WriteString(" " + a.Key + `="` + html.EscapeString(a.Val) + `"`)
		}
		if tt == xhtml.SelfClosingTagToken {
			out.WriteString("/")
		}
		out.WriteString(">")
	}
	return out.String(), n
}

/** resolveHTMLRefs resolves the relative href, src and poster values in HTML.
 *
 * Parameters:
 *   s (string) — HTML text
 *   base (string) — the item's link
 *
 * Returns:
 *   string — s with its relative references made absolute
 *   int — how many references were changed
 *
 * Example:
 *   out, n := resolveHTMLRefs(`<img src="/a.png">`, "https://example.org/p/")
 *   // `<img src="https://example.org/a.png">`, 1
 */
func resolveHTMLRefs(s, base string) (string, int) {
	return rewriteHTMLRefs(s, func(ref string) (string, bool) { return resolveRef(base, ref) })
}

// scanMarkdownRefs finds the destination of every inline link and image and
// every reference definition in s, outside code and raw HTML, and passes each
// to fn. It returns s with the destinations fn changed, and how many.
func scanMarkdownRefs(s string, fn func(ref string) (string, bool)) (string, int) {
	body := []byte(s)
	protected := protectedRanges(body)
	var edits []linkEdit
	consider := func(start, end int) {
		if v, ok := fn(string(body[start:end])); ok {
			edits = append(edits, linkEdit{start: start, end: end, repl: v})
		}
	}
	for i := 0; i+1 < len(body); i++ {
		if body[i] != ']' || body[i+1] != '(' || escaped(body, i) || inRanges(protected, i) {
			continue
		}
		if openingBracket(body, i) < 0 {
			continue
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
	if len(edits) == 0 {
		return s, 0
	}
	sort.Slice(edits, func(i, j int) bool { return edits[i].start < edits[j].start })
	return string(applyEdits(body, edits)), len(edits)
}

/** resolveMarkdownRefs resolves the relative destinations of inline links,
 * images and reference definitions in Markdown. Code spans, code blocks and
 * raw HTML are not touched.
 *
 * Parameters:
 *   s (string) — Markdown text
 *   base (string) — the item's link
 *
 * Returns:
 *   string — s with its relative destinations made absolute
 *   int — how many destinations were changed
 *
 * Example:
 *   out, n := resolveMarkdownRefs("![x](/a.png)", "https://example.org/p/")
 *   // "![x](https://example.org/a.png)", 1
 */
func resolveMarkdownRefs(s, base string) (string, int) {
	return scanMarkdownRefs(s, func(ref string) (string, bool) { return resolveRef(base, ref) })
}

// countRelative counts the relative references in s, as HTML or Markdown.
func countRelative(s string, markdown bool) int {
	n := 0
	count := func(ref string) (string, bool) {
		if isRelativeRef(ref) {
			n++
		}
		return ref, false
	}
	if markdown {
		scanMarkdownRefs(s, count)
	} else {
		rewriteHTMLRefs(s, count)
	}
	return n
}

// usableBase reports whether link can resolve a relative reference.
func usableBase(link string) bool {
	_, ok := resolveRef(link, "x")
	return ok
}

// resolveItemRefs resolves the relative references in an item's body, as HTML
// or Markdown, against the item's link. When the link cannot serve as a base
// and the body has relative references, the body is left as written and the
// item is remembered for the warning generate gives at the end (DR-0037 2).
// key names the item across the three outputs so it is counted once.
func (gen *Generator) resolveItemRefs(body, link, key string, markdown bool) string {
	if body == "" {
		return body
	}
	if usableBase(link) {
		if markdown {
			out, _ := resolveMarkdownRefs(body, link)
			return out
		}
		out, _ := resolveHTMLRefs(body, link)
		return out
	}
	if countRelative(body, markdown) > 0 {
		if gen.unresolved == nil {
			gen.unresolved = map[string]struct{}{}
		}
		gen.unresolved[key] = struct{}{}
	}
	return body
}

// itemKey names an item across outputs for the unresolved count.
func itemKey(link, title, pubDate, label string) string {
	return link + "\x00" + title + "\x00" + pubDate + "\x00" + label
}

// reportUnresolved warns once for the collection when items had relative
// references and no address to resolve them against.
func (gen *Generator) reportUnresolved(collection string) {
	if len(gen.unresolved) == 0 || gen.eout == nil {
		return
	}
	fmt.Fprintf(gen.eout, "warning %s: %d item(s) have relative links or images and no usable link to resolve them against; left as written\n",
		collection, len(gen.unresolved))
	gen.unresolved = nil
}
