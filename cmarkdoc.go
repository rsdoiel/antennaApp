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
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	// 3rd party packages
	mathjax "github.com/litao91/goldmark-mathjax"
	"github.com/mangoumbrella/goldmark-figure"
	"github.com/mmcdole/gofeed"
	"github.com/stefanfritsch/goldmark-fences"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark-emoji"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v3"
)

// CommonMark holds the structure of front matter and the CommonMark
// text.
type CommonMark struct {
	// FrontMatter holds the object of the parsed FrontMatter if available
	// in the document.
	FrontMatter map[string]interface{} `json:"frontMatter,omitempty" yaml:"frontMatter,omitempty"`
	// Text holds the CommonMark text that comes after any front matter
	Text string `json:"text,omitempty" yaml:"text,omitempty"`
	// useMathJax controls if MathJax support is included when rendering HTML
	useMathJax bool `json:"-"`
	// Links, when set, rewrites relative links to Markdown files antenna
	// renders so the HTML links to the HTML page (DR-0007). Nil leaves links
	// as written.
	Links *LinkRewriter `json:"-" yaml:"-"`
	// Macros, when set, expands {{page.NAME}} macros from the front matter
	// while rendering (DR-0009 6). Nil expands nothing.
	Macros *MacroExpander `json:"-" yaml:"-"`
	// Nav and NavFrom, when set, let the HTML take urlPrev and urlNext from the
	// site's table of contents when the front matter gives none (DR-0013 5).
	// NavFrom is this document's own published Markdown path.
	Nav     tocNav `json:"-" yaml:"-"`
	NavFrom string `json:"-" yaml:"-"`
}


// ParseMarkdownLinks parses a Markdown text for links in the format `- [LABEL](URL "DESCRIPTION")`
// and returns a slice of Link structures.
func ParseMarkdownLinks(markdownText string) ([]Link, error) {
	// Regular expression to match Markdown links with optional description
	re := regexp.MustCompile(`-\s+\[([^\]]+)\]\(([^)\s]+)(?:\s+"([^"]+)")?\)`)

	lines := strings.Split(markdownText, "\n")
	var links []Link

	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if len(matches) == 0 {
			continue // Skip lines that don't match the pattern
		}

		label := matches[1]
		url := matches[2]
		description := ""
		if len(matches) > 3 {
			description = matches[3]
		}

		links = append(links, Link{
			Label:       label,
			URL:         url,
			Description: description,
		})
	}

	return links, nil
}

/** frontMatterBlock finds the front matter block at the top of a file
 * (DR-0022): the first line is "---" and the block ends at the next line that
 * is "---" or "...". A "\r" before a newline is allowed. Any later "---" line
 * is body. A file whose first line is not "---" has no block.
 *
 * Parameters:
 *   src ([]byte) — the whole file
 *
 * Returns:
 *   string — the text between the two lines, lines joined with "\n"
 *   int — the offset in src where the body starts, after the closing line
 *   bool — true when a block was found
 *   error — a data error when the first line is "---" and there is no closing line
 *
 * Example:
 *   block, bodyAt, found, err := frontMatterBlock([]byte("---\ntitle: T\n---\nBody.\n"))
 */
func frontMatterBlock(src []byte) (string, int, bool, error) {
	var lines []string
	pos, first := 0, true
	for pos < len(src) {
		end := len(src)
		if nl := bytes.IndexByte(src[pos:], '\n'); nl >= 0 {
			end = pos + nl
		}
		line := strings.TrimRight(string(src[pos:end]), "\r")
		next := end + 1
		if next > len(src) {
			next = len(src)
		}
		if first {
			if line != "---" {
				return "", 0, false, nil
			}
			first = false
		} else if line == "---" || line == "..." {
			return strings.Join(lines, "\n"), next, true, nil
		} else {
			lines = append(lines, line)
		}
		pos = next
	}
	if first {
		return "", 0, false, nil
	}
	return "", 0, false, dataErrorf("unclosed FrontMatter")
}

// SplitFrontMatter splits a CommonMark document into FrontMatter and the rest
// of the content. Front matter is only the block at the top of the file
// (DR-0022); the rest is returned as written, line endings and final newline
// included.
func SplitFrontMatter(src []byte) (map[string]interface{}, string, error) {
	block, bodyAt, found, err := frontMatterBlock(src)
	if err != nil {
		return nil, string(src), err
	}
	if !found {
		return nil, string(src), nil
	}
	body := string(src[bodyAt:])
	if strings.TrimSpace(block) == "" {
		return nil, body, nil // an empty block
	}
	var result map[string]interface{}
	if err := yaml.Unmarshal([]byte(block), &result); err != nil {
		return nil, "", dataErrorf("front matter: %w", err)
	}
	return result, body, nil
}

// UseMathJax will set whether or not the MathJax extension is loaded when
// converting to HTML
func (doc *CommonMark) UseMathJax(value bool) {
	doc.useMathJax = value
}

// Parse will read a byte slice and populate any FrontMatter found
// and set the remaining text as the Text element of CommonMark structure.
func (doc *CommonMark) Parse(src []byte) error {
	frontMatter, text, err := SplitFrontMatter(src)
	if err != nil {
		return err
	}
	doc.FrontMatter = frontMatter
	doc.Text = text
	return nil
}

// GetLinks process the Text of a CommonMark struct and returns
// a list of Link objects if found.
func (doc *CommonMark) GetLinks() ([]Link, error) {
	return ParseMarkdownLinks(doc.Text)
}

// GetAttributeString returns a string attribute from
// the front matter the document
func (doc *CommonMark) GetAttributeString(key string, defaultValue string) string {
	if val, ok := doc.FrontMatter[key].(string); ok {
		return val
	}
	return defaultValue
}

// emailAddressGetName transform an email addres like "jane.doe@example.edi (Jane Doe)" to
// "Jane Doe".
func emailAddressGetName(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, ")") {
		parts := strings.SplitN(s, "(", 2)
		if len(parts) == 2 && strings.Contains(parts[1], "@") == false {
			return parts[1]
		}
		return ""
	}
	if !strings.Contains(s, "@") {
		return s
	}
	return ""
}

func emailAddressTrimName(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, ")") {
		parts := strings.SplitN(s, "(", 2)
		if len(parts) > 0 && strings.Contains(parts[0], "@") {
			return parts[0]
		}
		return ""
	}
	if strings.Contains(s, "@") {
		return s
	}
	return ""
}

func mapToPerson(val map[string]interface{}) *gofeed.Person {
	person := &gofeed.Person{}
	if v, ok := val["name"].(string); ok {
		person.Name = emailAddressGetName(v)
	}
	if v, ok := val["email"].(string); ok {
		person.Email = emailAddressTrimName(v)
	}
	if person.Name != "" || person.Email != "" {
		return person
	}
	return nil
}

// GetPersons returns a list of `*gofeed.Person{}`
// from the front matter in the document document
func (doc *CommonMark) GetPersons(key string, isRequired bool) ([]*gofeed.Person, error) {
	peopleList := []*gofeed.Person{}
	if val, ok := doc.FrontMatter[key].(interface{}); ok {
		switch val.(type) {
		case string:
			person := &gofeed.Person{}
			person.Name = emailAddressGetName(val.(string))
			person.Email = emailAddressTrimName(val.(string))
			if person.Name != "" || person.Email != "" {
				peopleList = append(peopleList, person)
			}
		case []interface{}:
			for i, v := range val.([]interface{}) {
				person := &gofeed.Person{}
				switch v.(type) {
				case string:
					person.Name = emailAddressGetName(v.(string))
					person.Email = emailAddressTrimName(v.(string))
					if person.Name != "" || person.Email != "" {
						peopleList = append(peopleList, person)
					}
				case map[string]interface{}:
					person = mapToPerson(v.(map[string]interface{}))
					if person != nil {
						peopleList = append(peopleList, person)
					}
				default:
					return nil, dataErrorf("failed to parse %q (%d) -> %T %+v", key, i, v, v)
				}
			}
		case map[string]interface{}:
			person := mapToPerson(val.(map[string]interface{}))
			if person != nil {
				peopleList = append(peopleList, person)
			}
		default:
			return nil, dataErrorf("unable to parse %q", key)
		}
	}
	// If we have a populated peopleList return it.
	if len(peopleList) > 0 {
		return peopleList, nil
	}
	// Do we required a populated peopleList?
	if isRequired {
		return nil, dataErrorf("no persons found for %q", key)
	}
	// An empty peopleList is OK, field is optional
	return nil, nil
}

/** GetAttributeStringSlice returns a []string for a front matter key whose
 * value is either a YAML sequence or a plain string. Returns nil if absent,
 * empty, or an unsupported type.
 *
 * Parameters:
 *   key (string) — the front matter key to look up
 *
 * Returns:
 *   []string — the values, or nil
 *
 * Example:
 *   // front matter: keywords: [Oberon, programming]
 *   kw := doc.GetAttributeStringSlice("keywords")
 *   // kw == []string{"Oberon", "programming"}
 */
func (doc *CommonMark) GetAttributeStringSlice(key string) []string {
	val, ok := doc.FrontMatter[key]
	if !ok {
		return nil
	}
	switch v := val.(type) {
	case string:
		if v == "" {
			return nil
		}
		return []string{v}
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		if len(out) == 0 {
			return nil
		}
		return out
	}
	return nil
}

// GetAttributeBool returns a boolean attribute from
// the front matter the document
func (doc *CommonMark) GetAttributeBool(key string, defaultValue bool) bool {
	if val, ok := doc.FrontMatter[key].(bool); ok {
		return val
	}
	return defaultValue
}

func (doc *CommonMark) String() string {
	if len(doc.FrontMatter) > 0 {
		src, _ := yaml.Marshal(doc.FrontMatter)
		return fmt.Sprintf(`---
%s
---

%s
`, src, doc.Text)
	}
	return doc.Text
}

func (doc *CommonMark) ToHTML() (string, error) {
	extenders := []goldmark.Extender{
			extension.GFM,
			extension.DefinitionList,
			extension.Footnote,
			emoji.Emoji,
			figure.Figure,
			figure.Figure.WithImageLink(),
			figure.Figure.WithSkipNoCaption(),
			extension.CJK,
			&fences.Extender{},
	}
	if doc.useMathJax {
			extenders = append(extenders, mathjax.MathJax)
	}
	md := goldmark.New(
		goldmark.WithExtensions(
			extenders...
		),
		goldmark.WithParserOptions(doc.parserOptions()...),
	)
	var buf bytes.Buffer
	if err := md.Convert([]byte(doc.Text), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (doc *CommonMark) ToUnsafeHTML() (string, error) {
	extenders := []goldmark.Extender{
			extension.GFM,
			extension.DefinitionList,
			extension.Footnote,
			emoji.Emoji,
			figure.Figure,
			figure.Figure.WithImageLink(),
			figure.Figure.WithSkipNoCaption(),
			extension.CJK,
			&fences.Extender{},
	}
	if doc.useMathJax {
			extenders = append(extenders, mathjax.MathJax)
	}
	md := goldmark.New(
		goldmark.WithExtensions(
			extenders...
		),
		goldmark.WithParserOptions(doc.parserOptions()...),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)
	var buf bytes.Buffer
	if err := md.Convert([]byte(doc.Text), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// isODTFile reports whether fName has an ODT or OTT file extension. antenna
// does not read them (DR-0032); the name is only used to refuse one.
func isODTFile(fName string) bool {
	switch strings.ToLower(filepath.Ext(fName)) {
	case ".odt", ".ott":
		return true
	}
	return false
}

// odtUnsupported is the message for an ODT or OTT file (DR-0032).
const odtUnsupported = "ODT is not supported; save the document as Markdown"

// normalizeToHTMLExt replaces the .md extension of a source document with
// .html. A name with any other extension is returned unchanged.
//
// Example:
//
//	normalizeToHTMLExt("blog/2026/04/12/post.md") // "blog/2026/04/12/post.html"
func normalizeToHTMLExt(name string) string {
	if strings.ToLower(filepath.Ext(name)) == ".md" {
		return strings.TrimSuffix(name, filepath.Ext(name)) + ".html"
	}
	return name
}

// LoadCommonMark reads a Markdown source document file and returns a
// *CommonMark, with its YAML front matter parsed. An ODT or OTT file is
// refused with a data error (DR-0032).
//
// Example:
//
//	doc, err := LoadCommonMark("post.md")
func LoadCommonMark(fName string) (*CommonMark, error) {
	if isODTFile(fName) {
		return nil, dataErrorf("%s: %s", fName, odtUnsupported)
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
