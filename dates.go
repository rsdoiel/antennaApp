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
	"strings"
	"time"
)

// rssDateLayouts are the RSS 2.0 (RFC 822 and RFC 1123) date forms, such as
// "Fri, 23 Oct 2026 15:19:45 GMT". The day may have one or two digits, and the
// weekday is optional. Zone words are converted to numeric offsets before
// parsing (see rssZoneOffsets), so every layout here ends in a numeric zone.
var rssDateLayouts = []string{
	"Mon, 2 Jan 2006 15:04:05 -0700",
	"2 Jan 2006 15:04:05 -0700",
	"Mon, 2 Jan 2006 15:04 -0700",
	"2 Jan 2006 15:04 -0700",
	"Mon, 2 Jan 06 15:04:05 -0700",
}

// rssZoneOffsets are the zone words RFC 822 defines. Go would read an
// abbreviation it does not know as offset zero, which is wrong for these.
var rssZoneOffsets = map[string]string{
	"GMT": "+0000", "UT": "+0000", "UTC": "+0000", "Z": "+0000",
	"EST": "-0500", "EDT": "-0400", "CST": "-0600", "CDT": "-0500",
	"MST": "-0700", "MDT": "-0600", "PST": "-0800", "PDT": "-0700",
}

/** parseRSSDate reads an RSS 2.0 date such as "Fri, 23 Oct 2026 15:19:45 GMT"
 * (antennaApp DR-0009 4) and returns the instant.
 *
 * Parameters:
 *   raw (string) — the date as written
 *
 * Returns:
 *   time.Time — the instant, in UTC
 *   bool — false when raw is not an RSS date
 *
 * Example:
 *   t, ok := parseRSSDate("Fri, 23 Oct 2026 17:19:45 +0200") // 2026-10-23 15:19:45 UTC, true
 */
func parseRSSDate(raw string) (time.Time, bool) {
	s := strings.Join(strings.Fields(raw), " ")
	if i := strings.LastIndex(s, " "); i > 0 {
		if off, ok := rssZoneOffsets[strings.ToUpper(s[i+1:])]; ok {
			s = s[:i+1] + off
		}
	}
	for _, layout := range rssDateLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC(), true
		}
	}
	return time.Time{}, false
}

/** GetAttributeDate reads a date from the front matter in any form antenna
 * accepts (antennaApp DR-0009 4) and returns it in the form antenna stores,
 * so posts sort and publish whichever form the author wrote:
 *
 *   - a quoted date or datetime in a stored form ("2026-06-27",
 *     "2026-06-27 10:00:00", RFC 3339) is kept exactly as written;
 *   - an unquoted date or timestamp, which YAML reads as a time value, becomes
 *     "2006-01-02" at midnight UTC and "2006-01-02 15:04:05" (UTC) otherwise;
 *   - an RSS 2.0 date becomes "2006-01-02 15:04:05" in UTC;
 *   - anything else is returned as written with ok false, so the caller can
 *     warn and carry on as before.
 *
 * The author's file is never changed.
 *
 * Parameters:
 *   key (string) — the front matter key, such as "pubDate"
 *
 * Returns:
 *   string — the date in stored form (or as written when ok is false)
 *   bool — found: the key is present and not empty
 *   bool — ok: the value is a date antenna can read
 *
 * Example:
 *   stored, found, ok := doc.GetAttributeDate("pubDate") // "2026-10-23 15:19:45", true, true
 */
func (doc *CommonMark) GetAttributeDate(key string) (string, bool, bool) {
	switch v := doc.FrontMatter[key].(type) {
	case time.Time:
		v = v.UTC()
		if v.Hour() == 0 && v.Minute() == 0 && v.Second() == 0 && v.Nanosecond() == 0 {
			return v.Format("2006-01-02"), true, true
		}
		return v.Format("2006-01-02 15:04:05"), true, true
	case string:
		raw := strings.TrimSpace(v)
		if raw == "" {
			return "", false, false
		}
		for _, layout := range append(storedDateLayouts[:len(storedDateLayouts):len(storedDateLayouts)], "2006-01-02") {
			if _, err := time.Parse(layout, raw); err == nil {
				return raw, true, true
			}
		}
		if t, ok := parseRSSDate(raw); ok {
			return t.Format("2006-01-02 15:04:05"), true, true
		}
		return raw, true, false
	}
	return "", false, false
}

// postPubDate returns the publication date for a post or page from its
// front matter: pubDate, then datePublished, in stored form. A value antenna
// cannot read is returned as written (the behaviour before DR-0009) and
// reported on warn, once, naming from.
func postPubDate(doc *CommonMark, from string, warn io.Writer) string {
	for _, key := range []string{"pubDate", "datePublished"} {
		stored, found, ok := doc.GetAttributeDate(key)
		if !found {
			continue
		}
		if !ok && warn != nil {
			fmt.Fprintf(warn, "warning: %s: %s %q is not a date antenna can read (YYYY-MM-DD or an RSS date such as \"Fri, 23 Oct 2026 15:19:45 GMT\")\n", from, key, stored)
		}
		return stored
	}
	return ""
}
