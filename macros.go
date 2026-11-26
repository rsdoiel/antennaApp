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
	"io"
	"net/url"
	"strings"

	"github.com/rsdoiel/frontmatter"
	"github.com/yuin/goldmark/parser"
)

/** MacroExpander expands BonBon's {{page.NAME}} macros while a page or post is
 * rendered to HTML (antennaApp DR-0009 6). Set it as CommonMark.Macros; nil
 * (the default, and what harvested feed items get) expands nothing.
 *
 * The work is frontmatter.MacroOption's (frontmatter DR-0002). Expansion works on the parsed document's text nodes, never on the source
 * text, so a value is inserted as escaped text and can never become Markdown,
 * a link or HTML. Code spans and code blocks are never touched. It makes one
 * pass over the page's own text, so a value holding a macro is shown as
 * written. A key the page does not have leaves the macro visible and is
 * reported once per page. When the page has a `link` value, the expanded
 * `title` is a link to it (BonBon's rule); an address that is not http, https,
 * mailto or relative is ignored with a warning.
 *
 * Fields:
 *   From (string) — the page's output path, named in warnings
 *   Warn (io.Writer) — receives one "warning: ..." line per problem; nil
 *     discards them
 */
type MacroExpander struct {
	From   string
	Warn   io.Writer
	report frontmatter.Report
}

// warnf writes one warning line.
func (m *MacroExpander) warnf(format string, args ...any) {
	if m.Warn != nil {
		fmt.Fprintf(m.Warn, "warning: %s: %s\n", m.From, fmt.Sprintf(format, args...))
	}
}

// flush writes what the last render could not expand, as warnings, and clears
// the report. Call it after the document has been converted.
func (m *MacroExpander) flush() {
	for _, key := range m.report.Missing {
		m.warnf("{{page.%s}}: the page has no front matter value for %s, so the macro is left as written", key, key)
	}
	for _, w := range m.report.Warnings {
		m.warnf("%s", w)
	}
	m.report = frontmatter.Report{}
}

// isSafePageAddress reports whether raw can be used as the address of a link
// that front matter supplies: an http, https or mailto address, or a relative
// one. Script and data schemes are refused.
func isSafePageAddress(raw string) bool {
	u, err := url.Parse(strings.TrimSpace(raw))
	return err == nil && (u.Scheme == "" || u.Scheme == "http" || u.Scheme == "https" || u.Scheme == "mailto")
}

// macroOption returns the parser option for doc's MacroExpander, or nil.
func (doc *CommonMark) macroOption() parser.Option {
	if doc.Macros == nil {
		return nil
	}
	return frontmatter.MacroOption(doc.FrontMatter, &doc.Macros.report)
}
