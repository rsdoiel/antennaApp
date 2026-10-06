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
	"bytes"
	"encoding/xml"
	"io"
	"strings"
	"testing"
)

// validateXML parses the given XML bytes and returns any parse error.
func validateXML(t *testing.T, src []byte) error {
	t.Helper()
	d := xml.NewDecoder(bytes.NewReader(src))
	for {
		_, err := d.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

// wrapCDATA wraps sanitized content in a CDATA section for XML validation.
func wrapCDATA(s string) string {
	return `<?xml version="1.0" encoding="UTF-8"?><r><![CDATA[` + s + `]]></r>`
}

func TestSanitizeCDATA_LeavesSafeContent(t *testing.T) {
	in := "Hello, world! &mdash; some text"
	got := sanitizeCDATA(in)
	if got != in {
		t.Errorf("expected unchanged content, got %q", got)
	}
}

func TestSanitizeCDATA_ProducesValidXMLWithTerminator(t *testing.T) {
	// A description containing ]]> must produce valid XML when CDATA-wrapped.
	in := "before ]]> after"
	xml := wrapCDATA(sanitizeCDATA(in))
	if err := validateXML(t, []byte(xml)); err != nil {
		t.Errorf("sanitized CDATA is not valid XML: %s\nXML: %s", err, xml)
	}
}

func TestSanitizeCDATA_ProducesValidXMLWithMultipleTerminators(t *testing.T) {
	in := "a ]]> b ]]> c"
	xml := wrapCDATA(sanitizeCDATA(in))
	if err := validateXML(t, []byte(xml)); err != nil {
		t.Errorf("sanitized CDATA is not valid XML: %s\nXML: %s", err, xml)
	}
}

func TestWriteItemRSS_ValidXMLWithHTMLEntities(t *testing.T) {
	// Harvested descriptions often contain named HTML entities like &mdash;, &nbsp;
	// The generated RSS must parse as valid XML.
	gen := &Generator{eout: io.Discard}
	var buf bytes.Buffer
	// Wrap in a minimal valid RSS shell so the XML parser is happy
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?><rss version="2.0"><channel>`)
	err := gen.WriteItemRSS(&buf, "http://example.com/", "Title &mdash; Subtitle",
		"<p>A post&mdash;with entities &amp; more &nbsp; content</p>",
		nil, nil, "guid-1", "2026-06-27", "", "", "published", "", "", "", "")
	if err != nil {
		t.Fatalf("WriteItemRSS: %s", err)
	}
	buf.WriteString(`</channel></rss>`)
	if err := validateXML(t, buf.Bytes()); err != nil {
		t.Errorf("RSS output is not valid XML: %s\nOutput:\n%s", err, buf.String())
	}
}

func TestWriteItemRSS_ValidXMLWithCDATATerminatorInDescription(t *testing.T) {
	// If description contains ]]>, the CDATA section must not break.
	gen := &Generator{eout: io.Discard}
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?><rss version="2.0"><channel>`)
	err := gen.WriteItemRSS(&buf, "http://example.com/", "JS Example",
		"Use <![CDATA[ ]]> in scripts to embed data",
		nil, nil, "guid-2", "2026-06-27", "", "", "published", "", "", "", "")
	if err != nil {
		t.Fatalf("WriteItemRSS: %s", err)
	}
	buf.WriteString(`</channel></rss>`)
	if err := validateXML(t, buf.Bytes()); err != nil {
		t.Errorf("RSS output is not valid XML when description has CDATA terminator: %s\nOutput:\n%s", err, buf.String())
	}
}

func TestWriteCustomRSS_ValidXMLWithProblematicContent(t *testing.T) {
	// Full WriteCustomRSS path with items that have HTML entities.
	db := newTestItemsDB(t)
	defer db.Close()

	// Insert an item with a description containing named entities
	_, err := db.Exec(`INSERT INTO items
		(link, title, description, authors, enclosures, guid, pubDate,
		 dcExt, channel, status, updated, label, postPath, sourceMarkdown, categories)
		VALUES (?, ?, ?, '', '', ?, ?, '', '', 'published', '', '', '', '', '')`,
		"http://example.com/1",
		"Post &mdash; Title",
		"<p>Content with &mdash; dash and &nbsp; space</p>",
		"guid-rss-1",
		"2026-06-27",
	)
	if err != nil {
		t.Fatalf("insert item: %s", err)
	}

	gen := &Generator{eout: io.Discard}
	col := &Collection{
		Title:       "Test Feed",
		Description: "A feed with & special characters",
		File:        "test.md",
	}
	var buf bytes.Buffer
	if err := gen.WriteCustomRSS(&buf, db, SQLDisplayItems, "http://example.com/test.xml", "antenna", col); err != nil {
		t.Fatalf("WriteCustomRSS: %s", err)
	}
	if err := validateXML(t, buf.Bytes()); err != nil {
		t.Errorf("RSS feed is not valid XML: %s\nOutput:\n%s", err, buf.String())
	}
}

// rssItemFor renders one item inside a minimal RSS shell and returns the
// text and a parse error, if any.
func rssItemFor(t *testing.T, enclosures []*Enclosure, pubDate string) (string, error) {
	t.Helper()
	gen := &Generator{eout: io.Discard}
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?><rss version="2.0"><channel>`)
	if err := gen.WriteItemRSS(&buf, "http://example.com/", "T", "d", nil,
		enclosures, "g", pubDate, "", "", "published", "", "", "", ""); err != nil {
		t.Fatalf("WriteItemRSS: %s", err)
	}
	buf.WriteString(`</channel></rss>`)
	return buf.String(), validateXML(t, buf.Bytes())
}

func TestWriteItemRSS_EnclosureURLWithAmpersandIsValidXML(t *testing.T) {
	// Podcast enclosure URLs routinely carry query strings (issue #22).
	enc := []*Enclosure{{
		Url:    "https://cdn.example.com/a.mp3?sid=1&source=rss&x=<y>",
		Length: "123", Type: "audio/mpeg",
	}}
	out, err := rssItemFor(t, enc, "2026-06-27")
	if err != nil {
		t.Fatalf("not valid XML: %s\n%s", err, out)
	}
	// The decoded attribute must round-trip to the original URL.
	d := xml.NewDecoder(strings.NewReader(out))
	for {
		tok, err := d.Token()
		if err != nil {
			t.Fatalf("no enclosure element found: %s", err)
		}
		if se, ok := tok.(xml.StartElement); ok && se.Name.Local == "enclosure" {
			for _, a := range se.Attr {
				if a.Name.Local == "url" && a.Value != enc[0].Url {
					t.Errorf("url round-trip: got %q want %q", a.Value, enc[0].Url)
				}
			}
			break
		}
	}
}

func TestWriteItemRSS_EnclosureEmptyLengthIsZero(t *testing.T) {
	// RSS 2.0 requires length; unknown is conventionally 0.
	out, _ := rssItemFor(t, []*Enclosure{{Url: "https://e.com/a.png", Type: "image/png"}}, "2026-06-27")
	if !strings.Contains(out, `length="0"`) {
		t.Errorf("expected length=\"0\" for unknown length, got:\n%s", out)
	}
}

func TestWriteItemRSS_PubDateFromStoredLayouts(t *testing.T) {
	// harvest.go stores "2006-01-02 15:04:05"; the production sqlite driver
	// returns RFC3339; posts store a bare date. All must yield a pubDate
	// carrying the item's own date (issue #22).
	for _, in := range []string{"2025-08-23 23:34:08", "2025-08-23T23:34:08Z", "2025-08-23"} {
		out, err := rssItemFor(t, nil, in)
		if err != nil {
			t.Fatalf("%q: not valid XML: %s", in, err)
		}
		if !strings.Contains(out, "<pubDate>Sat, 23 Aug 2025 ") {
			t.Errorf("%q: expected RFC 1123Z pubDate for 23 Aug 2025, got:\n%s", in, out)
		}
	}
}

func TestWriteItemRSS_UnparsablePubDateIsOmitted(t *testing.T) {
	out, err := rssItemFor(t, nil, "sometime last week")
	if err != nil {
		t.Fatalf("not valid XML: %s", err)
	}
	if strings.Contains(out, "<pubDate>") {
		t.Errorf("expected no pubDate for unparsable value, got:\n%s", out)
	}
}
