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
	"net/url"
	"strings"
)

// isSafeImageAddress reports whether raw can be the source of an image that
// front matter supplies: an http or https address or a relative one. Script,
// data and mailto addresses are refused.
func isSafeImageAddress(raw string) bool {
	u, err := url.Parse(strings.TrimSpace(raw))
	return err == nil && (u.Scheme == "" || u.Scheme == "http" || u.Scheme == "https")
}

/** marginImageHTML returns the image that front matter's rightMarginImage
 * names, for the top of a page's text (antennaApp DR-0009 3), or "" when the
 * page has none. The CSS floats it to the right. BonBon's key carries no alt
 * text, so the alt comes from antenna's optional rightMarginImageAlt; without
 * it the image is decorative: an empty alt, nothing invented, no warning.
 *
 * Parameters:
 *   frontMatter (map[string]interface{}) — the page's front matter as
 *     returned by CommonMark.htmlFrontMatter (the address already checked)
 *
 * Returns:
 *   string — an <img> element, or ""
 *
 * Example:
 *   marginImageHTML(map[string]interface{}{"rightMarginImage": "a.jpg", "rightMarginImageAlt": "A cat"})
 *   // `<img class="right-margin-image" src="a.jpg" alt="A cat">`
 */
func marginImageHTML(frontMatter map[string]interface{}) string {
	src, ok := frontMatter["rightMarginImage"].(string)
	if !ok || strings.TrimSpace(src) == "" {
		return ""
	}
	alt, _ := frontMatter["rightMarginImageAlt"].(string)
	return fmt.Sprintf(`<img class="right-margin-image" src="%s" alt="%s">`,
		html.EscapeString(strings.TrimSpace(src)), html.EscapeString(strings.TrimSpace(alt)))
}
