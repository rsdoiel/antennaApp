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
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ODTMeta holds the document properties extracted from the meta.xml inside an
// ODT or OTT file. Fields map directly to Dublin Core and ODF meta namespace
// elements.
type ODTMeta struct {
	// Title is the document title (dc:title).
	Title string
	// Description is the document description or abstract (dc:description).
	Description string
	// Creator is the name of the person who last modified the document (dc:creator).
	Creator string
	// InitialCreator is the name of the original author (meta:initial-creator).
	InitialCreator string
	// Date is the date and time of the last modification in ISO 8601 format (dc:date).
	Date string
	// CreationDate is the date and time the document was first created in ISO 8601
	// format (meta:creation-date).
	CreationDate string
	// Keywords is the list of document keywords (one meta:keyword element per entry).
	Keywords []string
	// Subject is the document subject (dc:subject).
	Subject string
	// Language is the document language tag, e.g. "en-US" (dc:language).
	Language string
	// Rights is the rights or license statement for the document (dc:rights).
	Rights string
	// Source is the URL or identifier of the resource from which this document
	// is derived (dc:source).
	Source string
	// Type is the nature or genre of the document, e.g. "Documentation" (dc:type).
	Type string
	// UserDefined holds any meta:user-defined properties keyed by their meta:name
	// attribute value.
	UserDefined map[string]string
}

// ODTMetaToFrontMatter converts an ODTMeta into a map[string]interface{} ready
// for use as YAML front matter in a CommonMark document. The mapping is:
//
//   - Title          → "title"
//   - Description    → "description"
//   - Creator (or InitialCreator if Creator is empty) → "author"
//   - CreationDate   → "pubDate"      (sub-second precision stripped)
//   - Date           → "dateModified" (sub-second precision stripped)
//   - Keywords       → "keywords"     ([]string)
//   - Subject        → "subject"
//   - Language       → "language"
//   - Rights         → "copyright"
//   - Source         → "source"
//   - Type           → "type"
//   - UserDefined    → each key/value passed through (existing keys are not overwritten)
//
// Example:
//
//	m, _ := ParseODTMeta("document.odt")
//	fm := ODTMetaToFrontMatter(m)
//	// fm["title"], fm["author"], fm["pubDate"] …
func ODTMetaToFrontMatter(m *ODTMeta) map[string]interface{} {
	fm := map[string]interface{}{}
	if m.Title != "" {
		fm["title"] = m.Title
	}
	if m.Description != "" {
		fm["description"] = m.Description
	}
	author := m.Creator
	if author == "" {
		author = m.InitialCreator
	}
	if author != "" {
		fm["author"] = author
	}
	if m.CreationDate != "" {
		fm["pubDate"] = normalizeODTDate(m.CreationDate)
	}
	if m.Date != "" {
		fm["dateModified"] = normalizeODTDate(m.Date)
	}
	if len(m.Keywords) > 0 {
		fm["keywords"] = m.Keywords
	}
	if m.Subject != "" {
		fm["subject"] = m.Subject
	}
	if m.Language != "" {
		fm["language"] = m.Language
	}
	if m.Rights != "" {
		fm["copyright"] = m.Rights
	}
	if m.Source != "" {
		fm["source"] = m.Source
	}
	if m.Type != "" {
		fm["type"] = m.Type
	}
	for k, v := range m.UserDefined {
		if _, exists := fm[k]; !exists {
			fm[k] = v
		}
	}
	return fm
}

// normalizeODTDate strips sub-second precision from an ODF ISO 8601 datetime
// string. LibreOffice Writer emits timestamps like "2026-04-11T12:24:44.869517200";
// this function trims the fractional seconds so the result is a standard
// RFC3339-compatible string that can be stored in the database.
//
// Example:
//
//	normalizeODTDate("2026-04-11T12:24:44.869517200") // "2026-04-11T12:24:44"
//	normalizeODTDate("2026-04-11T12:24:44")           // "2026-04-11T12:24:44"
func normalizeODTDate(s string) string {
	if i := strings.IndexByte(s, '.'); i != -1 {
		return s[:i]
	}
	return s
}

// isODTFile reports whether fName has an ODT or OTT file extension.
func isODTFile(fName string) bool {
	switch strings.ToLower(filepath.Ext(fName)) {
	case ".odt", ".ott":
		return true
	}
	return false
}

// normalizeToHTMLExt replaces a source document extension (.md, .odt, .ott)
// with .html. If the extension is not a recognised source type the name is
// returned unchanged.
//
// Example:
//
//	normalizeToHTMLExt("blog/2026/04/12/post.odt") // "blog/2026/04/12/post.html"
//	normalizeToHTMLExt("blog/2026/04/12/post.md")  // "blog/2026/04/12/post.html"
func normalizeToHTMLExt(name string) string {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".md", ".odt", ".ott":
		return strings.TrimSuffix(name, filepath.Ext(name)) + ".html"
	}
	return name
}

// xmlUserDefined is used internally when decoding meta:user-defined elements.
type xmlUserDefined struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

// xmlOfficeMeta mirrors the office:meta element inside a meta.xml file.
// Go's encoding/xml matches element names by local name, so namespace prefixes
// (dc:, meta:) are ignored during decoding.
type xmlOfficeMeta struct {
	Title          string           `xml:"title"`
	Description    string           `xml:"description"`
	Creator        string           `xml:"creator"`
	InitialCreator string           `xml:"initial-creator"`
	Date           string           `xml:"date"`
	CreationDate   string           `xml:"creation-date"`
	Keywords       []string         `xml:"keyword"`
	Subject        string           `xml:"subject"`
	Language       string           `xml:"language"`
	Rights         string           `xml:"rights"`
	Source         string           `xml:"source"`
	Type           string           `xml:"type"`
	UserDefined    []xmlUserDefined `xml:"user-defined"`
}

// xmlDocumentMeta is the root element of a meta.xml file.
type xmlDocumentMeta struct {
	Meta xmlOfficeMeta `xml:"meta"`
}

// parseODTMetaXML decodes the content of a meta.xml byte slice into an ODTMeta
// struct. It is the inner implementation used by ParseODTMeta and is exported
// for testing with synthetic XML.
//
// Example:
//
//	data, _ := os.ReadFile("meta.xml")
//	m, err := parseODTMetaXML(data)
func parseODTMetaXML(data []byte) (*ODTMeta, error) {
	var doc xmlDocumentMeta
	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing ODT meta.xml: %w", err)
	}
	m := &ODTMeta{
		Title:          doc.Meta.Title,
		Description:    doc.Meta.Description,
		Creator:        doc.Meta.Creator,
		InitialCreator: doc.Meta.InitialCreator,
		Date:           doc.Meta.Date,
		CreationDate:   doc.Meta.CreationDate,
		Keywords:       doc.Meta.Keywords,
		Subject:        doc.Meta.Subject,
		Language:       doc.Meta.Language,
		Rights:         doc.Meta.Rights,
		Source:         doc.Meta.Source,
		Type:           doc.Meta.Type,
	}
	if len(doc.Meta.UserDefined) > 0 {
		m.UserDefined = make(map[string]string, len(doc.Meta.UserDefined))
		for _, ud := range doc.Meta.UserDefined {
			if ud.Name != "" {
				m.UserDefined[ud.Name] = ud.Value
			}
		}
	}
	return m, nil
}

// ParseODTMeta opens the ODT or OTT file at path, locates meta.xml inside the
// ZIP archive, and returns the document properties as an ODTMeta struct.
//
// Example:
//
//	m, err := ParseODTMeta("article.odt")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println("title:", m.Title, "author:", m.Creator)
func ParseODTMeta(path string) (*ODTMeta, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open %q: %w", path, err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("cannot stat %q: %w", path, err)
	}

	zr, err := zip.NewReader(f, info.Size())
	if err != nil {
		return nil, fmt.Errorf("cannot read ZIP archive %q: %w", path, err)
	}

	for _, zf := range zr.File {
		if zf.Name != "meta.xml" {
			continue
		}
		rc, err := zf.Open()
		if err != nil {
			return nil, fmt.Errorf("cannot open meta.xml in %q: %w", path, err)
		}
		defer rc.Close()
		data, err := io.ReadAll(rc)
		if err != nil {
			return nil, fmt.Errorf("cannot read meta.xml in %q: %w", path, err)
		}
		return parseODTMetaXML(data)
	}
	return nil, fmt.Errorf("meta.xml not found in %q", path)
}

// parseODTContentXML scans content.xml byte data and returns every hyperlink
// found in the document as a Link slice. It uses a token-based XML decoder so
// that nested markup inside link text (e.g. text:span for bold) is handled
// correctly — the visible text of each link is accumulated from CharData
// tokens. The xlink:href attribute becomes Link.URL and xlink:title (if
// present) becomes Link.Description.
//
// Example:
//
//	links, err := parseODTContentXML(data)
func parseODTContentXML(data []byte) ([]Link, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	var links []Link
	var inLink bool
	var href, title, label string

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("parsing ODT content.xml: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "a" {
				inLink = true
				href, title, label = "", "", ""
				for _, attr := range t.Attr {
					switch attr.Name.Local {
					case "href":
						href = attr.Value
					case "title":
						title = attr.Value
					}
				}
			}
		case xml.EndElement:
			if inLink && t.Name.Local == "a" {
				if href != "" {
					links = append(links, Link{
						Label:       strings.TrimSpace(label),
						URL:         href,
						Description: title,
					})
				}
				inLink = false
			}
		case xml.CharData:
			if inLink {
				label += string(t)
			}
		}
	}
	return links, nil
}

// ParseODTLinks opens the ODT or OTT file at path, reads content.xml from the
// ZIP archive, and returns every hyperlink in the document as a Link slice.
// This is used to treat an ODT document as a collection definition file, where
// the hyperlinks represent RSS/Atom feed URLs — the same role that Markdown
// link-list entries play in a .md collection file.
//
// Example:
//
//	links, err := ParseODTLinks("my-feeds.odt")
//	for _, l := range links {
//		fmt.Println(l.URL, l.Label)
//	}
func ParseODTLinks(path string) ([]Link, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open %q: %w", path, err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("cannot stat %q: %w", path, err)
	}

	zr, err := zip.NewReader(f, info.Size())
	if err != nil {
		return nil, fmt.Errorf("cannot read ZIP archive %q: %w", path, err)
	}

	for _, zf := range zr.File {
		if zf.Name != "content.xml" {
			continue
		}
		rc, err := zf.Open()
		if err != nil {
			return nil, fmt.Errorf("cannot open content.xml in %q: %w", path, err)
		}
		defer rc.Close()
		data, err := io.ReadAll(rc)
		if err != nil {
			return nil, fmt.Errorf("cannot read content.xml in %q: %w", path, err)
		}
		return parseODTContentXML(data)
	}
	// No content.xml — treat as a document with no links (not an error)
	return nil, nil
}

// linksToMarkdown converts a slice of Link values into a Markdown unordered
// list suitable for use as doc.Text in a collection CommonMark document.
// Each entry becomes "- [Label](URL)" or "- [Label](URL "Description")".
// If a link has no label the URL is used in its place.
//
// Example:
//
//	md := linksToMarkdown([]Link{{Label: "My Feed", URL: "https://example.org/rss.xml"}})
//	// "- [My Feed](https://example.org/rss.xml)\n"
func linksToMarkdown(links []Link) string {
	if len(links) == 0 {
		return ""
	}
	var sb strings.Builder
	for _, link := range links {
		label := link.Label
		if label == "" {
			label = link.URL
		}
		if link.Description != "" {
			fmt.Fprintf(&sb, "- [%s](%s %q)\n", label, link.URL, link.Description)
		} else {
			fmt.Fprintf(&sb, "- [%s](%s)\n", label, link.URL)
		}
	}
	return sb.String()
}

// ODTToCommonMark reads an ODT or OTT file and returns a CommonMark document
// whose FrontMatter is populated from the document properties (meta.xml) and
// whose Text contains a Markdown link list built from every hyperlink found in
// the document body (content.xml). The Text field can therefore be used
// directly with GetLinks(), making an ODT file a valid collection definition.
//
// Example:
//
//	doc, err := ODTToCommonMark("my-feeds.odt")
//	if err != nil {
//		log.Fatal(err)
//	}
//	links, _ := doc.GetLinks()  // returns the feed URLs
func ODTToCommonMark(path string) (*CommonMark, error) {
	m, err := ParseODTMeta(path)
	if err != nil {
		return nil, err
	}
	links, err := ParseODTLinks(path)
	if err != nil {
		return nil, err
	}
	doc := &CommonMark{
		FrontMatter: ODTMetaToFrontMatter(m),
		Text:        linksToMarkdown(links),
	}
	return doc, nil
}

// LoadCommonMark reads a source document file and returns a *CommonMark.
// For Markdown files (.md) it reads and parses the file normally including
// any YAML front matter. For ODT and OTT files it extracts document properties
// from meta.xml and returns them as front matter with an empty body.
//
// Example:
//
//	doc, err := LoadCommonMark("post.md")
//	doc, err := LoadCommonMark("post.odt")
func LoadCommonMark(fName string) (*CommonMark, error) {
	if isODTFile(fName) {
		return ODTToCommonMark(fName)
	}
	src, err := os.ReadFile(fName)
	if err != nil {
		return nil, fmt.Errorf("failed to read %q: %w", fName, err)
	}
	doc := &CommonMark{}
	if err := doc.Parse(src); err != nil {
		return nil, fmt.Errorf("failed to parse %q: %w", fName, err)
	}
	return doc, nil
}
