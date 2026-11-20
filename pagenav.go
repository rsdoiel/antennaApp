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
	"html"
	"strings"
)

// pageNavKeys are the front matter keys that name the previous and next page
// (BonBon's urlPrev and urlNext, antennaApp DR-0009 3), with the link relation
// and the visible text for each, in the order they are shown.
var pageNavKeys = []struct{ Key, Rel, Text string }{
	{"urlPrev", "prev", "Previous"},
	{"urlNext", "next", "Next"},
}

/** htmlFrontMatter returns the front matter to hand to the HTML renderers: a
 * copy of the document's in which urlPrev and urlNext are rewritten like any
 * link in the page (a same-site .md address becomes the .html page, DR-0007),
 * and an address that is not http, https, mailto or relative is dropped with a
 * warning. The document's own front matter, and so the published Markdown, is
 * not changed.
 *
 * Returns:
 *   map[string]interface{} — front matter for WriteHtmlPage, nil when the
 *     document has none
 *
 * Example:
 *   err := gen.WriteHtmlPage(htmlName, link, postPath, pubDate, innerHTML, doc.htmlFrontMatter())
 */
func (doc *CommonMark) htmlFrontMatter() map[string]interface{} {
	if doc.FrontMatter == nil {
		return nil
	}
	out := make(map[string]interface{}, len(doc.FrontMatter))
	for k, v := range doc.FrontMatter {
		out[k] = v
	}
	for _, nk := range pageNavKeys {
		raw, ok := doc.FrontMatter[nk.Key].(string)
		if !ok {
			continue
		}
		raw = strings.TrimSpace(raw)
		if raw == "" {
			delete(out, nk.Key)
			continue
		}
		from := ""
		if doc.Links != nil {
			from = doc.Links.From
		}
		if !isSafePageAddress(raw) {
			if doc.Links != nil && doc.Links.Warn != nil {
				fmt.Fprintf(doc.Links.Warn, "warning: %s: ignoring %s %q: only http, https, mailto and relative addresses are used\n", from, nk.Key, raw)
			}
			delete(out, nk.Key)
			continue
		}
		if doc.Links != nil {
			dest, warn := RewriteDestination(doc.Links.From, raw, doc.Links.Targets)
			if warn != "" && doc.Links.Warn != nil {
				fmt.Fprintf(doc.Links.Warn, "warning: %s\n", warn)
			}
			raw = dest
		}
		out[nk.Key] = raw
	}
	return out
}

// pageNavAddresses returns the usable previous and next addresses in
// front matter, by key, skipping any that are not text.
func pageNavAddresses(frontMatter map[string]interface{}) map[string]string {
	found := map[string]string{}
	for _, nk := range pageNavKeys {
		if s, ok := frontMatter[nk.Key].(string); ok && strings.TrimSpace(s) != "" {
			found[nk.Key] = strings.TrimSpace(s)
		}
	}
	return found
}

// pageNavHTML returns the visible previous and next navigation for a page, or
// "" when the page names neither. It is a labelled navigation landmark (DR-0002):
// the links are text, with rel set, in reading order.
func pageNavHTML(frontMatter map[string]interface{}) string {
	addrs := pageNavAddresses(frontMatter)
	if len(addrs) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("<nav class=\"page-nav\" aria-label=\"Previous and next pages\">\n")
	for _, nk := range pageNavKeys {
		if addr, ok := addrs[nk.Key]; ok {
			fmt.Fprintf(&sb, "  <a rel=%q href=\"%s\">%s</a>\n", nk.Rel, html.EscapeString(addr), nk.Text)
		}
	}
	sb.WriteString("</nav>")
	return sb.String()
}
