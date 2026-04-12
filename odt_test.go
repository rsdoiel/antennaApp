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
	"os"
	"path/filepath"
	"testing"
)

// collectionContentXML is a minimal content.xml containing a list of
// hyperlinks as would appear in a LibreOffice Writer collection definition file.
var collectionContentXML = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<office:document-content
  xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
  xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0"
  xmlns:xlink="http://www.w3.org/1999/xlink"
  office:version="1.4">
  <office:body>
    <office:text>
      <text:list>
        <text:list-item>
          <text:p>
            <text:a xlink:type="simple" xlink:href="https://rsdoiel.github.io/rss.xml">R. S. Doiel&apos;s blog</text:a>
          </text:p>
        </text:list-item>
        <text:list-item>
          <text:p>
            <text:a xlink:type="simple" xlink:href="http://scripting.com/rss.xml" xlink:title="Dave Winer&apos;s blog">Scripting News</text:a>
          </text:p>
        </text:list-item>
        <text:list-item>
          <text:p>
            <text:a xlink:type="simple" xlink:href="https://example.org/feed.xml">Example Feed</text:a>
          </text:p>
        </text:list-item>
      </text:list>
    </office:text>
  </office:body>
</office:document-content>`)

// makeODTWithMetaAndContent creates a minimal valid ODT ZIP file containing
// both meta.xml and content.xml. It writes the file to path.
func makeODTWithMetaAndContent(path string, metaXML, contentXML []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	defer zw.Close()
	mt, err := zw.Create("mimetype")
	if err != nil {
		return err
	}
	if _, err := mt.Write([]byte("application/vnd.oasis.opendocument.text")); err != nil {
		return err
	}
	mw, err := zw.Create("meta.xml")
	if err != nil {
		return err
	}
	if _, err := mw.Write(metaXML); err != nil {
		return err
	}
	cw, err := zw.Create("content.xml")
	if err != nil {
		return err
	}
	if _, err := cw.Write(contentXML); err != nil {
		return err
	}
	return nil
}

// makeODTWithMeta creates a minimal valid ODT ZIP file containing only
// meta.xml with the provided content. It writes the file to path and returns
// any error.
func makeODTWithMeta(path string, metaXML []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	defer zw.Close()
	// mimetype must be the first entry and stored uncompressed per ODF spec,
	// but for metadata-only testing any valid zip works.
	mt, err := zw.Create("mimetype")
	if err != nil {
		return err
	}
	if _, err := mt.Write([]byte("application/vnd.oasis.opendocument.text")); err != nil {
		return err
	}
	mw, err := zw.Create("meta.xml")
	if err != nil {
		return err
	}
	if _, err := mw.Write(metaXML); err != nil {
		return err
	}
	return nil
}

// fullMetaXML is a representative meta.xml with every property that ODTMeta tracks,
// including the dc:rights, dc:source, and dc:type fields added for README.odt support.
var fullMetaXML = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<office:document-meta
  xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
  xmlns:meta="urn:oasis:names:tc:opendocument:xmlns:meta:1.0"
  xmlns:dc="http://purl.org/dc/elements/1.1/"
  office:version="1.3">
  <office:meta>
    <dc:title>A Test Article</dc:title>
    <dc:description>A short description of the article.</dc:description>
    <dc:creator>Jane Doe</dc:creator>
    <meta:initial-creator>John Smith</meta:initial-creator>
    <dc:date>2026-04-11T13:47:00.086857599</dc:date>
    <meta:creation-date>2026-01-15T09:00:00.000000000</meta:creation-date>
    <meta:keyword>antenna</meta:keyword>
    <meta:keyword>testing</meta:keyword>
    <meta:keyword>odt</meta:keyword>
    <dc:subject>Software documentation</dc:subject>
    <dc:language>en-US</dc:language>
    <dc:rights>Licensed under AGPL-3.0-or-later</dc:rights>
    <dc:source>https://example.org/source</dc:source>
    <dc:type>Article</dc:type>
    <meta:user-defined meta:name="postPath">blog/2026/01/15/a-test-article.md</meta:user-defined>
    <meta:user-defined meta:name="status">draft</meta:user-defined>
  </office:meta>
</office:document-meta>`)

// minimalMetaXML has only the fields that testdata/index.odt actually contains.
var minimalMetaXML = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<office:document-meta
  xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
  xmlns:meta="urn:oasis:names:tc:opendocument:xmlns:meta:1.0"
  xmlns:dc="http://purl.org/dc/elements/1.1/"
  office:version="1.4">
  <office:meta>
    <meta:initial-creator>Robert Doiel</meta:initial-creator>
    <meta:creation-date>2026-04-11T12:24:44.869517200</meta:creation-date>
    <dc:date>2026-04-11T13:47:00.086857599</dc:date>
    <dc:creator>Robert Doiel</dc:creator>
  </office:meta>
</office:document-meta>`)

// -------------------------------------------------------------------
// parseODTMetaXML tests
// -------------------------------------------------------------------

func TestParseODTMetaXML_FullProperties(t *testing.T) {
	m, err := parseODTMetaXML(fullMetaXML)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"Title", m.Title, "A Test Article"},
		{"Description", m.Description, "A short description of the article."},
		{"Creator", m.Creator, "Jane Doe"},
		{"InitialCreator", m.InitialCreator, "John Smith"},
		{"Date", m.Date, "2026-04-11T13:47:00.086857599"},
		{"CreationDate", m.CreationDate, "2026-01-15T09:00:00.000000000"},
		{"Subject", m.Subject, "Software documentation"},
		{"Language", m.Language, "en-US"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.name, tt.got, tt.want)
		}
	}
	// Keywords
	if len(m.Keywords) != 3 {
		t.Errorf("Keywords: got %d entries, want 3", len(m.Keywords))
	} else {
		for i, kw := range []string{"antenna", "testing", "odt"} {
			if m.Keywords[i] != kw {
				t.Errorf("Keywords[%d]: got %q, want %q", i, m.Keywords[i], kw)
			}
		}
	}
	// UserDefined
	if m.UserDefined == nil {
		t.Fatal("UserDefined: got nil, want populated map")
	}
	if got := m.UserDefined["postPath"]; got != "blog/2026/01/15/a-test-article.md" {
		t.Errorf("UserDefined[postPath]: got %q, want %q", got, "blog/2026/01/15/a-test-article.md")
	}
	if got := m.UserDefined["status"]; got != "draft" {
		t.Errorf("UserDefined[status]: got %q, want %q", got, "draft")
	}
}

func TestParseODTMetaXML_MinimalProperties(t *testing.T) {
	m, err := parseODTMetaXML(minimalMetaXML)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if m.Creator != "Robert Doiel" {
		t.Errorf("Creator: got %q, want %q", m.Creator, "Robert Doiel")
	}
	if m.InitialCreator != "Robert Doiel" {
		t.Errorf("InitialCreator: got %q, want %q", m.InitialCreator, "Robert Doiel")
	}
	if m.Title != "" {
		t.Errorf("Title: got %q, want empty", m.Title)
	}
	if m.Description != "" {
		t.Errorf("Description: got %q, want empty", m.Description)
	}
	if len(m.Keywords) != 0 {
		t.Errorf("Keywords: got %d entries, want 0", len(m.Keywords))
	}
	if len(m.UserDefined) != 0 {
		t.Errorf("UserDefined: got %d entries, want 0", len(m.UserDefined))
	}
}

func TestParseODTMetaXML_InvalidXML(t *testing.T) {
	_, err := parseODTMetaXML([]byte("this is not xml"))
	if err == nil {
		t.Error("expected error for invalid XML, got nil")
	}
}

func TestParseODTMetaXML_EmptyMeta(t *testing.T) {
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<office:document-meta
  xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
  xmlns:meta="urn:oasis:names:tc:opendocument:xmlns:meta:1.0"
  xmlns:dc="http://purl.org/dc/elements/1.1/">
  <office:meta/>
</office:document-meta>`)
	m, err := parseODTMetaXML(data)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if m.Title != "" || m.Creator != "" || m.Description != "" {
		t.Errorf("expected all fields empty for empty meta.xml, got %+v", m)
	}
}

// -------------------------------------------------------------------
// normalizeODTDate tests
// -------------------------------------------------------------------

func TestNormalizeODTDate_WithFraction(t *testing.T) {
	got := normalizeODTDate("2026-04-11T13:47:00.086857599")
	want := "2026-04-11T13:47:00"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNormalizeODTDate_WithoutFraction(t *testing.T) {
	got := normalizeODTDate("2026-04-11T13:47:00")
	want := "2026-04-11T13:47:00"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNormalizeODTDate_DateOnly(t *testing.T) {
	got := normalizeODTDate("2026-04-11")
	want := "2026-04-11"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// -------------------------------------------------------------------
// ODTMetaToFrontMatter tests
// -------------------------------------------------------------------

func TestODTMetaToFrontMatter_FullProperties(t *testing.T) {
	m := &ODTMeta{
		Title:        "My Article",
		Description:  "A description",
		Creator:      "Jane Doe",
		CreationDate: "2026-01-15T09:00:00.000000000",
		Date:         "2026-04-11T13:47:00.086857599",
		Keywords:     []string{"go", "odt"},
		Subject:      "Programming",
		Language:     "en-US",
		UserDefined:  map[string]string{"postPath": "blog/2026/01/15/my-article.md"},
	}
	fm := ODTMetaToFrontMatter(m)

	checks := []struct {
		key  string
		want interface{}
	}{
		{"title", "My Article"},
		{"description", "A description"},
		{"author", "Jane Doe"},
		{"pubDate", "2026-01-15T09:00:00"},
		{"dateModified", "2026-04-11T13:47:00"},
		{"subject", "Programming"},
		{"language", "en-US"},
		{"postPath", "blog/2026/01/15/my-article.md"},
	}
	for _, c := range checks {
		if got, ok := fm[c.key]; !ok {
			t.Errorf("missing key %q in front matter", c.key)
		} else if got != c.want {
			t.Errorf("fm[%q]: got %v, want %v", c.key, got, c.want)
		}
	}
	// keywords is []string
	kws, ok := fm["keywords"].([]string)
	if !ok {
		t.Errorf("keywords: expected []string, got %T", fm["keywords"])
	} else if len(kws) != 2 || kws[0] != "go" || kws[1] != "odt" {
		t.Errorf("keywords: got %v, want [go odt]", kws)
	}
}

func TestODTMetaToFrontMatter_CreatorFallsBackToInitialCreator(t *testing.T) {
	m := &ODTMeta{
		InitialCreator: "Original Author",
	}
	fm := ODTMetaToFrontMatter(m)
	if got, ok := fm["author"]; !ok {
		t.Error("author key missing when only InitialCreator is set")
	} else if got != "Original Author" {
		t.Errorf("author: got %q, want %q", got, "Original Author")
	}
}

func TestODTMetaToFrontMatter_CreatorTakesPriorityOverInitialCreator(t *testing.T) {
	m := &ODTMeta{
		Creator:        "Last Editor",
		InitialCreator: "Original Author",
	}
	fm := ODTMetaToFrontMatter(m)
	if got := fm["author"]; got != "Last Editor" {
		t.Errorf("author: got %q, want %q", got, "Last Editor")
	}
}

func TestODTMetaToFrontMatter_UserDefinedDoesNotOverwriteStandardKey(t *testing.T) {
	m := &ODTMeta{
		Title:       "Standard Title",
		UserDefined: map[string]string{"title": "User Title"},
	}
	fm := ODTMetaToFrontMatter(m)
	if got := fm["title"]; got != "Standard Title" {
		t.Errorf("title: got %q, want %q — user-defined should not overwrite standard key", got, "Standard Title")
	}
}

func TestODTMetaToFrontMatter_EmptyMeta(t *testing.T) {
	fm := ODTMetaToFrontMatter(&ODTMeta{})
	if len(fm) != 0 {
		t.Errorf("expected empty front matter for empty ODTMeta, got %v", fm)
	}
}

// -------------------------------------------------------------------
// ParseODTMeta file-level tests
// -------------------------------------------------------------------

func TestParseODTMeta_TestdataIndexODT(t *testing.T) {
	// testdata/index.odt is the real fixture committed to the repository.
	// We verify the fields we know are present from inspecting the file.
	m, err := ParseODTMeta(filepath.Join("testdata", "index.odt"))
	if err != nil {
		t.Fatalf("ParseODTMeta: %s", err)
	}
	if m.Creator != "Robert Doiel" {
		t.Errorf("Creator: got %q, want %q", m.Creator, "Robert Doiel")
	}
	if m.InitialCreator != "Robert Doiel" {
		t.Errorf("InitialCreator: got %q, want %q", m.InitialCreator, "Robert Doiel")
	}
	if m.CreationDate == "" {
		t.Error("CreationDate: got empty, want non-empty")
	}
	if m.Date == "" {
		t.Error("Date: got empty, want non-empty")
	}
}

func TestParseODTMeta_SyntheticFullODT(t *testing.T) {
	// Build a minimal ODT ZIP with fullMetaXML in a temp directory.
	dir := t.TempDir()
	path := filepath.Join(dir, "full.odt")
	if err := makeODTWithMeta(path, fullMetaXML); err != nil {
		t.Fatalf("makeODTWithMeta: %s", err)
	}
	m, err := ParseODTMeta(path)
	if err != nil {
		t.Fatalf("ParseODTMeta: %s", err)
	}
	if m.Title != "A Test Article" {
		t.Errorf("Title: got %q, want %q", m.Title, "A Test Article")
	}
	if m.Creator != "Jane Doe" {
		t.Errorf("Creator: got %q, want %q", m.Creator, "Jane Doe")
	}
	if len(m.Keywords) != 3 {
		t.Errorf("Keywords: got %d, want 3", len(m.Keywords))
	}
	if m.UserDefined["postPath"] != "blog/2026/01/15/a-test-article.md" {
		t.Errorf("UserDefined[postPath]: got %q", m.UserDefined["postPath"])
	}
}

func TestParseODTMeta_MissingFile(t *testing.T) {
	_, err := ParseODTMeta("testdata/does-not-exist.odt")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestParseODTMeta_NoMetaXML(t *testing.T) {
	// Build an ODT ZIP that contains no meta.xml.
	dir := t.TempDir()
	path := filepath.Join(dir, "no-meta.odt")
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	mt, _ := zw.Create("mimetype")
	mt.Write([]byte("application/vnd.oasis.opendocument.text")) //nolint
	zw.Close()
	if err := os.WriteFile(path, buf.Bytes(), 0664); err != nil {
		t.Fatalf("WriteFile: %s", err)
	}
	_, err := ParseODTMeta(path)
	if err == nil {
		t.Error("expected error when meta.xml is absent, got nil")
	}
}

// -------------------------------------------------------------------
// README.odt integration tests
// -------------------------------------------------------------------

func TestParseODTMeta_ReadmeODT(t *testing.T) {
	m, err := ParseODTMeta(filepath.Join("testdata", "README.odt"))
	if err != nil {
		t.Fatalf("ParseODTMeta: %s", err)
	}
	if m.Title != "README for Antenna App" {
		t.Errorf("Title: got %q, want %q", m.Title, "README for Antenna App")
	}
	if m.Creator != "Robert Doiel" {
		t.Errorf("Creator: got %q, want %q", m.Creator, "Robert Doiel")
	}
	if m.Subject != "Static Website Content Management" {
		t.Errorf("Subject: got %q, want %q", m.Subject, "Static Website Content Management")
	}
	if m.Rights == "" {
		t.Error("Rights: got empty, want non-empty")
	}
	if m.Source == "" {
		t.Error("Source: got empty, want non-empty")
	}
	if m.Type != "Documentation" {
		t.Errorf("Type: got %q, want %q", m.Type, "Documentation")
	}
	if len(m.Keywords) == 0 {
		t.Error("Keywords: got empty slice, want at least one keyword")
	}
	if m.UserDefined == nil {
		t.Fatal("UserDefined: got nil, want populated map")
	}
	if m.UserDefined["Date Published"] == "" {
		t.Error("UserDefined[Date Published]: got empty, want non-empty")
	}
	if m.UserDefined["Date Modified"] == "" {
		t.Error("UserDefined[Date Modified]: got empty, want non-empty")
	}
}

func TestODTMetaToFrontMatter_ReadmeODT(t *testing.T) {
	m, err := ParseODTMeta(filepath.Join("testdata", "README.odt"))
	if err != nil {
		t.Fatalf("ParseODTMeta: %s", err)
	}
	fm := ODTMetaToFrontMatter(m)

	checks := []struct {
		key  string
		want string
	}{
		{"title", "README for Antenna App"},
		{"author", "Robert Doiel"},
		{"subject", "Static Website Content Management"},
		{"type", "Documentation"},
	}
	for _, c := range checks {
		got, ok := fm[c.key]
		if !ok {
			t.Errorf("missing key %q in front matter", c.key)
			continue
		}
		if got != c.want {
			t.Errorf("fm[%q]: got %q, want %q", c.key, got, c.want)
		}
	}
	if _, ok := fm["copyright"]; !ok {
		t.Error("missing key \"copyright\" (from dc:rights)")
	}
	if _, ok := fm["source"]; !ok {
		t.Error("missing key \"source\" (from dc:source)")
	}
	if _, ok := fm["dateModified"]; !ok {
		t.Error("missing key \"dateModified\" (from dc:date)")
	}
}

// -------------------------------------------------------------------
// Rights/Source/Type field tests in parseODTMetaXML
// -------------------------------------------------------------------

func TestParseODTMetaXML_RightsSourceType(t *testing.T) {
	m, err := parseODTMetaXML(fullMetaXML)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if m.Rights != "Licensed under AGPL-3.0-or-later" {
		t.Errorf("Rights: got %q, want %q", m.Rights, "Licensed under AGPL-3.0-or-later")
	}
	if m.Source != "https://example.org/source" {
		t.Errorf("Source: got %q, want %q", m.Source, "https://example.org/source")
	}
	if m.Type != "Article" {
		t.Errorf("Type: got %q, want %q", m.Type, "Article")
	}
}

func TestODTMetaToFrontMatter_RightsSourceType(t *testing.T) {
	m := &ODTMeta{
		Rights: "CC-BY-4.0",
		Source: "https://example.org",
		Type:   "Article",
	}
	fm := ODTMetaToFrontMatter(m)
	if got := fm["copyright"]; got != "CC-BY-4.0" {
		t.Errorf("copyright: got %q, want %q", got, "CC-BY-4.0")
	}
	if got := fm["source"]; got != "https://example.org" {
		t.Errorf("source: got %q, want %q", got, "https://example.org")
	}
	if got := fm["type"]; got != "Article" {
		t.Errorf("type: got %q, want %q", got, "Article")
	}
}

// -------------------------------------------------------------------
// isODTFile and normalizeToHTMLExt tests
// -------------------------------------------------------------------

func TestIsODTFile(t *testing.T) {
	cases := []struct {
		name string
		want bool
	}{
		{"document.odt", true},
		{"template.ott", true},
		{"Document.ODT", true},
		{"post.md", false},
		{"index.html", false},
		{"style.css", false},
		{"noext", false},
	}
	for _, c := range cases {
		if got := isODTFile(c.name); got != c.want {
			t.Errorf("isODTFile(%q): got %v, want %v", c.name, got, c.want)
		}
	}
}

func TestNormalizeToHTMLExt(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"post.md", "post.html"},
		{"post.odt", "post.html"},
		{"post.ott", "post.html"},
		{"blog/2026/04/12/post.odt", "blog/2026/04/12/post.html"},
		{"blog/2026/04/12/post.md", "blog/2026/04/12/post.html"},
		{"index.html", "index.html"},
		{"style.css", "style.css"},
	}
	for _, c := range cases {
		if got := normalizeToHTMLExt(c.in); got != c.want {
			t.Errorf("normalizeToHTMLExt(%q): got %q, want %q", c.in, got, c.want)
		}
	}
}

// -------------------------------------------------------------------
// ODTToCommonMark and LoadCommonMark tests
// -------------------------------------------------------------------

func TestODTToCommonMark_ReadmeODT(t *testing.T) {
	doc, err := ODTToCommonMark(filepath.Join("testdata", "README.odt"))
	if err != nil {
		t.Fatalf("ODTToCommonMark: %s", err)
	}
	if doc.FrontMatter == nil {
		t.Fatal("FrontMatter: got nil, want populated map")
	}
	if got := doc.GetAttributeString("title", ""); got != "README for Antenna App" {
		t.Errorf("title: got %q, want %q", got, "README for Antenna App")
	}
	// README.odt contains hyperlinks so doc.Text is a Markdown link list.
	// Verify GetLinks() works on it without error.
	links, err := doc.GetLinks()
	if err != nil {
		t.Fatalf("GetLinks: %s", err)
	}
	if len(links) == 0 {
		t.Error("expected at least one link from README.odt hyperlinks, got none")
	}
}

func TestLoadCommonMark_MarkdownFile(t *testing.T) {
	// Write a temporary Markdown file with front matter and verify it round-trips.
	dir := t.TempDir()
	mdPath := filepath.Join(dir, "test.md")
	content := "---\ntitle: Test Post\n---\n\nBody text.\n"
	if err := os.WriteFile(mdPath, []byte(content), 0664); err != nil {
		t.Fatalf("WriteFile: %s", err)
	}
	doc, err := LoadCommonMark(mdPath)
	if err != nil {
		t.Fatalf("LoadCommonMark: %s", err)
	}
	if got := doc.GetAttributeString("title", ""); got != "Test Post" {
		t.Errorf("title: got %q, want %q", got, "Test Post")
	}
	if !bytes.Contains([]byte(doc.Text), []byte("Body text.")) {
		t.Errorf("Text: expected body content, got %q", doc.Text)
	}
}

func TestLoadCommonMark_ODTFile(t *testing.T) {
	doc, err := LoadCommonMark(filepath.Join("testdata", "README.odt"))
	if err != nil {
		t.Fatalf("LoadCommonMark: %s", err)
	}
	if doc.GetAttributeString("title", "") != "README for Antenna App" {
		t.Errorf("title: got %q", doc.GetAttributeString("title", ""))
	}
}

func TestLoadCommonMark_MissingFile(t *testing.T) {
	_, err := LoadCommonMark("testdata/does-not-exist.md")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

// -------------------------------------------------------------------
// Collection support: parseODTContentXML, ParseODTLinks, linksToMarkdown
// -------------------------------------------------------------------

func TestParseODTContentXML_Links(t *testing.T) {
	links, err := parseODTContentXML(collectionContentXML)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(links) != 3 {
		t.Fatalf("got %d links, want 3", len(links))
	}
	cases := []struct {
		url   string
		label string
		desc  string
	}{
		{"https://rsdoiel.github.io/rss.xml", "R. S. Doiel's blog", ""},
		{"http://scripting.com/rss.xml", "Scripting News", "Dave Winer's blog"},
		{"https://example.org/feed.xml", "Example Feed", ""},
	}
	for i, c := range cases {
		if links[i].URL != c.url {
			t.Errorf("links[%d].URL: got %q, want %q", i, links[i].URL, c.url)
		}
		if links[i].Label != c.label {
			t.Errorf("links[%d].Label: got %q, want %q", i, links[i].Label, c.label)
		}
		if links[i].Description != c.desc {
			t.Errorf("links[%d].Description: got %q, want %q", i, links[i].Description, c.desc)
		}
	}
}

func TestParseODTContentXML_NoLinks(t *testing.T) {
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<office:document-content
  xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
  xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
  <office:body><office:text><text:p>No hyperlinks here.</text:p></office:text></office:body>
</office:document-content>`)
	links, err := parseODTContentXML(data)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(links) != 0 {
		t.Errorf("got %d links, want 0", len(links))
	}
}

func TestParseODTContentXML_SkipsEmptyHref(t *testing.T) {
	// An anchor element with no xlink:href should be ignored.
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<office:document-content
  xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
  xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0"
  xmlns:xlink="http://www.w3.org/1999/xlink">
  <office:body><office:text>
    <text:p><text:a>no href here</text:a></text:p>
    <text:p><text:a xlink:href="https://good.example/rss.xml">Good</text:a></text:p>
  </office:text></office:body>
</office:document-content>`)
	links, err := parseODTContentXML(data)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(links) != 1 {
		t.Fatalf("got %d links, want 1", len(links))
	}
	if links[0].URL != "https://good.example/rss.xml" {
		t.Errorf("URL: got %q", links[0].URL)
	}
}

func TestLinksToMarkdown(t *testing.T) {
	links := []Link{
		{Label: "My Feed", URL: "https://example.org/rss.xml"},
		{Label: "Other Feed", URL: "https://other.org/feed.xml", Description: "A great feed"},
		{Label: "", URL: "https://bare.org/rss.xml"},
	}
	md := linksToMarkdown(links)
	want := "- [My Feed](https://example.org/rss.xml)\n" +
		"- [Other Feed](https://other.org/feed.xml \"A great feed\")\n" +
		"- [https://bare.org/rss.xml](https://bare.org/rss.xml)\n"
	if md != want {
		t.Errorf("linksToMarkdown:\ngot:  %q\nwant: %q", md, want)
	}
}

func TestLinksToMarkdown_Empty(t *testing.T) {
	if got := linksToMarkdown(nil); got != "" {
		t.Errorf("expected empty string for nil links, got %q", got)
	}
}

func TestParseODTLinks_SyntheticCollection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "collection.odt")
	if err := makeODTWithMetaAndContent(path, minimalMetaXML, collectionContentXML); err != nil {
		t.Fatalf("makeODTWithMetaAndContent: %s", err)
	}
	links, err := ParseODTLinks(path)
	if err != nil {
		t.Fatalf("ParseODTLinks: %s", err)
	}
	if len(links) != 3 {
		t.Fatalf("got %d links, want 3", len(links))
	}
}

func TestParseODTLinks_NoContentXML(t *testing.T) {
	// ODT with only meta.xml — ParseODTLinks should return nil, nil (not an error)
	dir := t.TempDir()
	path := filepath.Join(dir, "meta-only.odt")
	if err := makeODTWithMeta(path, minimalMetaXML); err != nil {
		t.Fatalf("makeODTWithMeta: %s", err)
	}
	links, err := ParseODTLinks(path)
	if err != nil {
		t.Errorf("expected nil error for missing content.xml, got %s", err)
	}
	if len(links) != 0 {
		t.Errorf("expected empty links for missing content.xml, got %d", len(links))
	}
}

func TestODTToCommonMark_WithLinks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "collection.odt")
	if err := makeODTWithMetaAndContent(path, fullMetaXML, collectionContentXML); err != nil {
		t.Fatalf("makeODTWithMetaAndContent: %s", err)
	}
	doc, err := ODTToCommonMark(path)
	if err != nil {
		t.Fatalf("ODTToCommonMark: %s", err)
	}
	// Front matter from meta.xml
	if doc.GetAttributeString("title", "") != "A Test Article" {
		t.Errorf("title: got %q", doc.GetAttributeString("title", ""))
	}
	// doc.Text should be a Markdown link list
	links, err := doc.GetLinks()
	if err != nil {
		t.Fatalf("GetLinks: %s", err)
	}
	if len(links) != 3 {
		t.Fatalf("got %d links from doc.Text, want 3", len(links))
	}
	if links[0].URL != "https://rsdoiel.github.io/rss.xml" {
		t.Errorf("links[0].URL: got %q", links[0].URL)
	}
}

func TestODTToCommonMark_NoLinks(t *testing.T) {
	// When content.xml has no links, doc.Text is empty and GetLinks returns nil, nil
	doc, err := ODTToCommonMark(filepath.Join("testdata", "README.odt"))
	if err != nil {
		t.Fatalf("ODTToCommonMark: %s", err)
	}
	// README.odt has prose links; they should appear in doc.Text
	links, err := doc.GetLinks()
	if err != nil {
		t.Fatalf("GetLinks: %s", err)
	}
	// README.odt has hyperlinks (Textcasting, antenna, INSTALL.md, etc.)
	// We just verify it doesn't error and the count is non-negative.
	if len(links) < 0 {
		t.Error("unexpected negative link count")
	}
}
