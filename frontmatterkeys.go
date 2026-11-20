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

/** FrontMatterKey describes a front matter key that antenna gives a meaning to
 * (antennaApp DR-0009). Keys are camelCase, as BonBon's and antenna's are.
 *
 * Fields:
 *   Name (string) — the key, spelled exactly
 *   BonBon (bool) — true for the eight keys BonBon reads
 *   Rendered (bool) — true when the HTML renderer writes the key's content
 *     itself, so it is not also emitted as a <meta> element
 *   Summary (string) — what the key does, for the manual and tests
 */
type FrontMatterKey struct {
	Name     string
	BonBon   bool
	Rendered bool
	Summary  string
}

/** frontMatterKeys is the one table of the front matter keys BonBon reads, in
 * the order its documents list them, plus antenna's own rightMarginImageAlt.
 * If BonBon's list changes, this table, the metadata manual (helptext.go) and
 * the test beside it are the places to edit. Other keys pass through unchanged
 * (DR-0009 2).
 */
var frontMatterKeys = []FrontMatterKey{
	{"title", true, true, "the page title; <title> in the head, and {{page.title}} in the text"},
	{"description", true, false, "a short summary"},
	{"author", true, false, "the author"},
	{"pubDate", true, false, "the publication date, YYYY-MM-DD or an RSS 2.0 date"},
	{"link", true, false, "the address the page's title links to"},
	{"urlPrev", true, true, "the previous page: <link rel=prev> and a Previous link"},
	{"urlNext", true, true, "the next page: <link rel=next> and a Next link"},
	{"rightMarginImage", true, true, "an image floated to the right at the top of the text"},
	{"rightMarginImageAlt", false, true, "the alt text for rightMarginImage; antenna's own, BonBon ignores it"},
}

// frontMatterKeyRendered reports whether the HTML renderer writes name's
// content itself, so name is not also emitted as a <meta> element.
func frontMatterKeyRendered(name string) bool {
	for _, k := range frontMatterKeys {
		if k.Name == name {
			return k.Rendered
		}
	}
	return false
}

// isBonBonKey reports whether name is one of the eight keys BonBon reads.
func isBonBonKey(name string) bool {
	for _, k := range frontMatterKeys {
		if k.Name == name {
			return k.BonBon
		}
	}
	return false
}
