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
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mmcdole/gofeed"
)

// newTestItemsDB creates an in-memory SQLite DB with a minimal items table,
// suitable for passing to WriteHTML in tests.
func newTestItemsDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory db: %s", err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS items (
		link TEXT DEFAULT '',
		title TEXT DEFAULT '',
		description TEXT DEFAULT '',
		authors TEXT DEFAULT '',
		enclosures TEXT DEFAULT '',
		guid TEXT DEFAULT '',
		pubDate TEXT DEFAULT '',
		dcExt TEXT DEFAULT '',
		channel TEXT DEFAULT '',
		status TEXT DEFAULT '',
		updated TEXT DEFAULT '',
		label TEXT DEFAULT '',
		postPath TEXT DEFAULT '',
		sourceMarkdown TEXT DEFAULT '',
		categories TEXT DEFAULT ''
	)`)
	if err != nil {
		t.Fatalf("create items table: %s", err)
	}
	return db
}

// -------------------------------------------------------------------
// writeHeadElement tests
// -------------------------------------------------------------------

func newTestGenerator() *Generator {
	gen, _ := NewGenerator("antenna-test", "https://example.com")
	return gen
}

func TestWriteHeadElement_NilFrontMatter(t *testing.T) {
	gen := newTestGenerator()
	gen.Title = "Site Title"
	var buf bytes.Buffer
	gen.writeHeadElement(&buf, "", nil)
	out := buf.String()
	if !strings.Contains(out, `<title>Site Title</title>`) {
		t.Errorf("expected gen.Title in <title>, got:\n%s", out)
	}
	if strings.Contains(out, "data-pagefind-filter") {
		t.Errorf("nil frontMatter should produce no data-pagefind-filter, got:\n%s", out)
	}
}

func TestWriteHeadElement_StringField(t *testing.T) {
	gen := newTestGenerator()
	fm := map[string]interface{}{"author": "R. S. Doiel"}
	var buf bytes.Buffer
	gen.writeHeadElement(&buf, "", fm)
	out := buf.String()
	if !strings.Contains(out, `name="author"`) {
		t.Errorf("expected <meta name=\"author\">, got:\n%s", out)
	}
	if !strings.Contains(out, `data-pagefind-filter="author[content]"`) {
		t.Errorf("expected data-pagefind-filter for author, got:\n%s", out)
	}
	if !strings.Contains(out, `content="R. S. Doiel"`) {
		t.Errorf("expected content value, got:\n%s", out)
	}
}

func TestWriteHeadElement_SliceField(t *testing.T) {
	gen := newTestGenerator()
	fm := map[string]interface{}{"keywords": []interface{}{"Oberon", "programming"}}
	var buf bytes.Buffer
	gen.writeHeadElement(&buf, "", fm)
	out := buf.String()
	// Each value gets its own <meta> pair
	if count := strings.Count(out, `name="keywords"`); count != 2 {
		t.Errorf("expected 2 <meta name=\"keywords\">, got %d in:\n%s", count, out)
	}
	if count := strings.Count(out, `data-pagefind-filter="keywords[content]"`); count != 2 {
		t.Errorf("expected 2 data-pagefind-filter for keywords, got %d in:\n%s", count, out)
	}
	if !strings.Contains(out, `content="Oberon"`) || !strings.Contains(out, `content="programming"`) {
		t.Errorf("expected both keyword values, got:\n%s", out)
	}
}

func TestWriteHeadElement_TitleOverridesGenTitle(t *testing.T) {
	gen := newTestGenerator()
	gen.Title = "Site Title"
	fm := map[string]interface{}{"title": "Post Title"}
	var buf bytes.Buffer
	gen.writeHeadElement(&buf, "", fm)
	out := buf.String()
	if !strings.Contains(out, `<title>Post Title</title>`) {
		t.Errorf("expected front matter title in <title>, got:\n%s", out)
	}
	if strings.Contains(out, `<title>Site Title</title>`) {
		t.Errorf("gen.Title should be suppressed when front matter has title, got:\n%s", out)
	}
	// title key must not also appear as a <meta name="title"> element
	if strings.Contains(out, `name="title"`) {
		t.Errorf("title must not be emitted as <meta name=\"title\">, got:\n%s", out)
	}
}

func TestWriteHeadElement_TitleFallsBackToGenTitle(t *testing.T) {
	gen := newTestGenerator()
	gen.Title = "Site Title"
	fm := map[string]interface{}{"author": "Alice"}
	var buf bytes.Buffer
	gen.writeHeadElement(&buf, "", fm)
	out := buf.String()
	if !strings.Contains(out, `<title>Site Title</title>`) {
		t.Errorf("expected gen.Title fallback, got:\n%s", out)
	}
}

func TestWriteHeadElement_AllowedMetaFields(t *testing.T) {
	gen := newTestGenerator()
	gen.AllowedMetaFields = []string{"author"}
	fm := map[string]interface{}{
		"author":   "R. S. Doiel",
		"postPath": "posts/2020/test.md",
	}
	var buf bytes.Buffer
	gen.writeHeadElement(&buf, "", fm)
	out := buf.String()
	if !strings.Contains(out, `name="author"`) {
		t.Errorf("expected allowed field author, got:\n%s", out)
	}
	if strings.Contains(out, `name="postPath"`) {
		t.Errorf("postPath should be excluded by AllowedMetaFields, got:\n%s", out)
	}
}

// -------------------------------------------------------------------
// WriteItem tests
// -------------------------------------------------------------------

func TestWriteItem_NoMetadata(t *testing.T) {
	gen := newTestGenerator()
	var buf bytes.Buffer
	_, err := gen.WriteItem(&buf, "https://example.com", "Title", "desc",
		nil, "", nil, "guid1", "", "", "", "", "", "", "", ItemsConfig{})
	if err != nil {
		t.Fatalf("WriteItem: %s", err)
	}
	out := buf.String()
	if strings.Contains(out, "data-pagefind-filter") {
		t.Errorf("expected no data-pagefind-filter with empty metadata, got:\n%s", out)
	}
}

func TestWriteItem_SingleCategory(t *testing.T) {
	gen := newTestGenerator()
	var buf bytes.Buffer
	_, err := gen.WriteItem(&buf, "https://example.com", "Title", "desc",
		nil, "", nil, "guid1", "", "", "", "", "", "", `["Oberon"]`, ItemsConfig{})
	if err != nil {
		t.Fatalf("WriteItem: %s", err)
	}
	out := buf.String()
	if !strings.Contains(out, "category:Oberon") {
		t.Errorf("expected category:Oberon in filter, got:\n%s", out)
	}
}

func TestWriteItem_MultipleCategories(t *testing.T) {
	gen := newTestGenerator()
	var buf bytes.Buffer
	_, err := gen.WriteItem(&buf, "https://example.com", "Title", "desc",
		nil, "", nil, "guid1", "", "", "", "", "", "", `["a","b"]`, ItemsConfig{})
	if err != nil {
		t.Fatalf("WriteItem: %s", err)
	}
	out := buf.String()
	if !strings.Contains(out, "category:a") || !strings.Contains(out, "category:b") {
		t.Errorf("expected both categories in filter, got:\n%s", out)
	}
}

func TestWriteItem_DCSubjectAndCreator(t *testing.T) {
	gen := newTestGenerator()
	var buf bytes.Buffer
	dcExt := `{"subject":["Languages"],"creator":["Alice"]}`
	_, err := gen.WriteItem(&buf, "https://example.com", "Title", "desc",
		nil, "", nil, "guid1", "", dcExt, "", "", "", "", "", ItemsConfig{})
	if err != nil {
		t.Fatalf("WriteItem: %s", err)
	}
	out := buf.String()
	if !strings.Contains(out, "dc_subject:Languages") {
		t.Errorf("expected dc_subject:Languages, got:\n%s", out)
	}
	if !strings.Contains(out, "dc_creator:Alice") {
		t.Errorf("expected dc_creator:Alice, got:\n%s", out)
	}
}

func TestWriteItem_Author(t *testing.T) {
	gen := newTestGenerator()
	authors := []*gofeed.Person{{Name: "R. S. Doiel"}}
	var buf bytes.Buffer
	_, err := gen.WriteItem(&buf, "https://example.com", "Title", "desc",
		authors, "", nil, "guid1", "", "", "", "", "", "", "", ItemsConfig{})
	if err != nil {
		t.Fatalf("WriteItem: %s", err)
	}
	out := buf.String()
	if !strings.Contains(out, "author:R. S. Doiel") {
		t.Errorf("expected author filter, got:\n%s", out)
	}
}

func TestWriteItem_LabelAndChannel(t *testing.T) {
	gen := newTestGenerator()
	var buf bytes.Buffer
	_, err := gen.WriteItem(&buf, "https://example.com", "Title", "desc",
		nil, "", nil, "guid1", "", "", "https://example.com/feed.xml", "", "", "My Feed", "", ItemsConfig{})
	if err != nil {
		t.Fatalf("WriteItem: %s", err)
	}
	out := buf.String()
	if !strings.Contains(out, "label:My Feed") {
		t.Errorf("expected label filter, got:\n%s", out)
	}
	if !strings.Contains(out, "channel:https://example.com/feed.xml") {
		t.Errorf("expected channel filter, got:\n%s", out)
	}
}

func TestWriteItem_PubDate(t *testing.T) {
	gen := newTestGenerator()
	var buf bytes.Buffer
	_, err := gen.WriteItem(&buf, "https://example.com", "Title", "desc",
		nil, "", nil, "guid1", "2020-01-01", "", "", "", "", "", "", ItemsConfig{})
	if err != nil {
		t.Fatalf("WriteItem: %s", err)
	}
	out := buf.String()
	if !strings.Contains(out, "datePublished:2020-01-01") {
		t.Errorf("expected datePublished filter, got:\n%s", out)
	}
}

func TestWriteItem_AllCombined(t *testing.T) {
	gen := newTestGenerator()
	authors := []*gofeed.Person{{Name: "Alice"}}
	dcExt := `{"subject":["Science"]}`
	var buf bytes.Buffer
	_, err := gen.WriteItem(&buf, "https://example.com", "Title", "desc",
		authors, "", nil, "guid1", "2020-06-01", dcExt,
		"https://example.com/feed.xml", "", "", "Feed Label", `["tech"]`, ItemsConfig{})
	if err != nil {
		t.Fatalf("WriteItem: %s", err)
	}
	out := buf.String()
	for _, want := range []string{
		"category:tech",
		"dc_subject:Science",
		"author:Alice",
		"datePublished:2020-06-01",
		"label:Feed Label",
		"channel:https://example.com/feed.xml",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in filter output, got:\n%s", want, out)
		}
	}
}

// -------------------------------------------------------------------
// Phase 1: rel typo fix
// -------------------------------------------------------------------

func TestWriteHeadElement_AlternateRel(t *testing.T) {
	gen := newTestGenerator()
	var buf bytes.Buffer
	gen.writeHeadElement(&buf, "posts/test.md", nil)
	out := buf.String()
	if strings.Contains(out, `rel="altenate"`) {
		t.Errorf("found typo rel=\"altenate\" — should be rel=\"alternate\"")
	}
	if !strings.Contains(out, `rel="alternate"`) {
		t.Errorf("expected rel=\"alternate\" on markdown link, got:\n%s", out)
	}
}

// -------------------------------------------------------------------
// Phase 2: configurable lang
// -------------------------------------------------------------------

func TestWriteHTML_DefaultLang(t *testing.T) {
	gen := newTestGenerator()
	db := newTestItemsDB(t)
	defer db.Close()
	var buf bytes.Buffer
	if err := gen.WriteHTML(&buf, db, "", nil); err != nil {
		t.Fatalf("WriteHTML: %s", err)
	}
	out := buf.String()
	if !strings.Contains(out, `<html lang="en-US">`) {
		t.Errorf("expected default lang=\"en-US\", got:\n%s", out)
	}
}

func TestWriteHTML_CustomLang(t *testing.T) {
	gen := newTestGenerator()
	gen.Lang = "fr-FR"
	db := newTestItemsDB(t)
	defer db.Close()
	var buf bytes.Buffer
	if err := gen.WriteHTML(&buf, db, "", nil); err != nil {
		t.Fatalf("WriteHTML: %s", err)
	}
	out := buf.String()
	if !strings.Contains(out, `<html lang="fr-FR">`) {
		t.Errorf("expected lang=\"fr-FR\", got:\n%s", out)
	}
	if strings.Contains(out, `lang="en-US"`) {
		t.Errorf("should not contain en-US when Lang is fr-FR, got:\n%s", out)
	}
}

func TestWriteHtmlPage_DefaultLang(t *testing.T) {
	gen := newTestGenerator()
	tmpFile := filepath.Join(t.TempDir(), "test.html")
	if err := gen.WriteHtmlPage(tmpFile, "", "", "", "<p>hi</p>", nil); err != nil {
		t.Fatalf("WriteHtmlPage: %s", err)
	}
	content, _ := os.ReadFile(tmpFile)
	out := string(content)
	if !strings.Contains(out, `<html lang="en-US">`) {
		t.Errorf("expected default lang=\"en-US\", got:\n%s", out)
	}
}

func TestWriteHtmlPage_CustomLang(t *testing.T) {
	gen := newTestGenerator()
	gen.Lang = "ja"
	tmpFile := filepath.Join(t.TempDir(), "test.html")
	if err := gen.WriteHtmlPage(tmpFile, "", "", "", "<p>hi</p>", nil); err != nil {
		t.Fatalf("WriteHtmlPage: %s", err)
	}
	content, _ := os.ReadFile(tmpFile)
	out := string(content)
	if !strings.Contains(out, `<html lang="ja">`) {
		t.Errorf("expected lang=\"ja\", got:\n%s", out)
	}
}

// -------------------------------------------------------------------
// Phase 3: skip navigation link
// -------------------------------------------------------------------

func TestWriteHTML_SkipLink(t *testing.T) {
	gen := newTestGenerator()
	db := newTestItemsDB(t)
	defer db.Close()
	var buf bytes.Buffer
	if err := gen.WriteHTML(&buf, db, "", nil); err != nil {
		t.Fatalf("WriteHTML: %s", err)
	}
	out := buf.String()
	if !strings.Contains(out, `<a href="#main-content" class="skip-link">Skip to main content</a>`) {
		t.Errorf("expected skip link, got:\n%s", out)
	}
	// Skip link must appear before <nav
	skipIdx := strings.Index(out, `class="skip-link"`)
	navIdx := strings.Index(out, `<nav`)
	if skipIdx < 0 {
		t.Fatal("skip link not found")
	}
	if navIdx >= 0 && skipIdx > navIdx {
		t.Errorf("skip link must appear before <nav>, but skip@%d nav@%d", skipIdx, navIdx)
	}
}

func TestWriteHtmlPage_SkipLink(t *testing.T) {
	gen := newTestGenerator()
	tmpFile := filepath.Join(t.TempDir(), "test.html")
	if err := gen.WriteHtmlPage(tmpFile, "", "", "", "<p>hi</p>", nil); err != nil {
		t.Fatalf("WriteHtmlPage: %s", err)
	}
	content, _ := os.ReadFile(tmpFile)
	out := string(content)
	if !strings.Contains(out, `<a href="#main-content" class="skip-link">Skip to main content</a>`) {
		t.Errorf("expected skip link, got:\n%s", out)
	}
}

// -------------------------------------------------------------------
// Phase 4: WriteItem structure — <time>, <footer>, no <address>
// -------------------------------------------------------------------

func TestWriteItem_NoAddress(t *testing.T) {
	gen := newTestGenerator()
	var buf bytes.Buffer
	if _, err := gen.WriteItem(&buf, "https://example.com", "Title", "desc",
		nil, "", nil, "guid1", "2020-01-01", "", "", "", "", "", "", ItemsConfig{}); err != nil {
		t.Fatalf("WriteItem: %s", err)
	}
	out := buf.String()
	if strings.Contains(out, "<address>") {
		t.Errorf("WriteItem must not emit <address>, got:\n%s", out)
	}
}

func TestWriteItem_HasFooter(t *testing.T) {
	gen := newTestGenerator()
	var buf bytes.Buffer
	if _, err := gen.WriteItem(&buf, "https://example.com", "Title", "desc",
		nil, "", nil, "guid1", "2020-01-01", "", "", "", "", "", "", ItemsConfig{}); err != nil {
		t.Fatalf("WriteItem: %s", err)
	}
	out := buf.String()
	if !strings.Contains(out, "<footer>") {
		t.Errorf("expected <footer> in article, got:\n%s", out)
	}
}

func TestWriteItem_TimeElement(t *testing.T) {
	gen := newTestGenerator()
	var buf bytes.Buffer
	if _, err := gen.WriteItem(&buf, "https://example.com", "Title", "desc",
		nil, "", nil, "guid1", "2020-04-11", "", "", "", "", "", "", ItemsConfig{}); err != nil {
		t.Fatalf("WriteItem: %s", err)
	}
	out := buf.String()
	if !strings.Contains(out, `<time datetime="2020-04-11">`) {
		t.Errorf("expected <time datetime=\"2020-04-11\">, got:\n%s", out)
	}
}

func TestWriteItem_DateNotInHeading(t *testing.T) {
	gen := newTestGenerator()
	var buf bytes.Buffer
	if _, err := gen.WriteItem(&buf, "https://example.com", "My Post", "desc",
		nil, "", nil, "guid1", "2020-04-11", "", "", "", "", "", "", ItemsConfig{}); err != nil {
		t.Fatalf("WriteItem: %s", err)
	}
	out := buf.String()
	// Find the h2 element and check it does NOT contain the date
	h2Start := strings.Index(out, "<h2>")
	h2End := strings.Index(out, "</h2>")
	if h2Start < 0 || h2End < 0 {
		t.Fatalf("no <h2> found in:\n%s", out)
	}
	h2Content := out[h2Start : h2End+5]
	if strings.Contains(h2Content, "2020") {
		t.Errorf("date must not be inside <h2>, got heading: %s", h2Content)
	}
}

func TestWriteItem_UpdatedTimeElement(t *testing.T) {
	gen := newTestGenerator()
	var buf bytes.Buffer
	if _, err := gen.WriteItem(&buf, "https://example.com", "Title", "desc",
		nil, "", nil, "guid1", "2020-04-11", "", "", "", "2020-05-01", "", "", ItemsConfig{}); err != nil {
		t.Fatalf("WriteItem: %s", err)
	}
	out := buf.String()
	if strings.Count(out, "<time ") < 2 {
		t.Errorf("expected two <time> elements when updated differs from pubDate, got:\n%s", out)
	}
	if !strings.Contains(out, `datetime="2020-05-01"`) {
		t.Errorf("expected updated date in <time datetime=\"2020-05-01\">, got:\n%s", out)
	}
}

// -------------------------------------------------------------------
// Phase 5 (item formatting): WriteItem + ItemsConfig integration
// -------------------------------------------------------------------

func TestWriteItem_FieldsRestrictsSections(t *testing.T) {
	gen := newTestGenerator()
	var buf bytes.Buffer
	cfg := ItemsConfig{Fields: []string{"title", "content"}}
	if _, err := gen.WriteItem(&buf, "https://example.com", "Title", "desc",
		nil, "", nil, "guid1", "2020-01-01", "", "", "", "", "My Feed", "", cfg); err != nil {
		t.Fatalf("WriteItem: %s", err)
	}
	out := buf.String()
	if !strings.Contains(out, "<h2>Title</h2>") {
		t.Errorf("expected title section present, got:\n%s", out)
	}
	if strings.Contains(out, "<time ") {
		t.Errorf("expected pubDate section absent, got:\n%s", out)
	}
	if strings.Contains(out, `class="source"`) {
		t.Errorf("expected source section absent, got:\n%s", out)
	}
	if strings.Contains(out, "<p></p>") {
		t.Errorf("excluded fields must omit their wrapper element, not leave an empty <p></p>, got:\n%s", out)
	}
}

func TestWriteItem_ShowSourceFalse(t *testing.T) {
	gen := newTestGenerator()
	var buf bytes.Buffer
	f := false
	cfg := ItemsConfig{ShowSource: &f}
	if _, err := gen.WriteItem(&buf, "https://example.com", "Title", "desc",
		nil, "", nil, "guid1", "2020-01-01", "", "", "", "", "My Feed", "", cfg); err != nil {
		t.Fatalf("WriteItem: %s", err)
	}
	out := buf.String()
	if strings.Contains(out, `class="source"`) {
		t.Errorf("expected source block absent when show_source is false, got:\n%s", out)
	}
}

func TestWriteItem_OmittedItemWritesNothing(t *testing.T) {
	gen := newTestGenerator()
	var buf bytes.Buffer
	cfg := ItemsConfig{Link: LinkConfig{Missing: "omit"}}
	omitted, err := gen.WriteItem(&buf, "", "Title", "desc",
		nil, "", nil, "guid1", "2020-01-01", "", "", "", "", "", "", cfg)
	if err != nil {
		t.Fatalf("WriteItem: %s", err)
	}
	if !omitted {
		t.Error("expected omitted=true when link is empty and missing:omit")
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output written for an omitted item, got:\n%s", buf.String())
	}
}

func TestWriteItem_DefaultLabelIsAccessible(t *testing.T) {
	gen := newTestGenerator()
	var buf bytes.Buffer
	if _, err := gen.WriteItem(&buf, "https://example.com", "Title", "desc",
		nil, "", nil, "guid1", "2020-01-01", "", "", "", "", "", "", ItemsConfig{}); err != nil {
		t.Fatalf("WriteItem: %s", err)
	}
	out := buf.String()
	if !strings.Contains(out, `<a href="https://example.com">Continue reading</a>`) {
		t.Errorf("expected accessible default anchor label, got:\n%s", out)
	}
}

func TestWriteItem_LabelFieldLinkOptOut(t *testing.T) {
	gen := newTestGenerator()
	var buf bytes.Buffer
	cfg := ItemsConfig{Link: LinkConfig{LabelField: "link"}}
	if _, err := gen.WriteItem(&buf, "https://example.com", "Title", "desc",
		nil, "", nil, "guid1", "2020-01-01", "", "", "", "", "", "", cfg); err != nil {
		t.Fatalf("WriteItem: %s", err)
	}
	out := buf.String()
	if !strings.Contains(out, `<a href="https://example.com">https://example.com</a>`) {
		t.Errorf("expected URL-as-label with explicit opt-out, got:\n%s", out)
	}
}

func TestWriteItem_CustomStaticLabel(t *testing.T) {
	gen := newTestGenerator()
	var buf bytes.Buffer
	cfg := ItemsConfig{Link: LinkConfig{LabelFallback: "read me"}}
	if _, err := gen.WriteItem(&buf, "https://example.com", "Title", "desc",
		nil, "", nil, "guid1", "2020-01-01", "", "", "", "", "", "", cfg); err != nil {
		t.Fatalf("WriteItem: %s", err)
	}
	out := buf.String()
	if !strings.Contains(out, `<a href="https://example.com">read me</a>`) {
		t.Errorf("expected custom static label via label_fallback only, got:\n%s", out)
	}
}

// -------------------------------------------------------------------
// Bug fixes found reviewing a real generated site: nested <p><p> from
// multi-paragraph sourceMarkdown, and unescaped title/href in HTML output.
// -------------------------------------------------------------------

func TestWriteItem_MultiParagraphContentNotNested(t *testing.T) {
	gen := newTestGenerator()
	var buf bytes.Buffer
	sourceMarkdown := "First paragraph.\n\nSecond paragraph."
	if _, err := gen.WriteItem(&buf, "https://example.com", "Title", "desc",
		nil, sourceMarkdown, nil, "guid1", "2020-01-01", "", "", "", "", "", "", ItemsConfig{}); err != nil {
		t.Fatalf("WriteItem: %s", err)
	}
	out := buf.String()
	if strings.Contains(out, "<p><p>") {
		t.Errorf("rendered multi-paragraph content must not be double-wrapped in <p>, got:\n%s", out)
	}
	if !strings.Contains(out, "<p>First paragraph.</p>") || !strings.Contains(out, "<p>Second paragraph.</p>") {
		t.Errorf("expected each source paragraph as its own <p>, got:\n%s", out)
	}
}

// TestWriteItem_ContentWrappedInSingleBodyContainer guards against a
// vertical-space bug found on a real "link-dump" post
// (inkdroid.org/2026/07/19/bookmarks/): its sourceMarkdown renders as many
// short elements (a heading + short paragraph per bookmark) rather than one
// long paragraph. The site's CSS collapses long articles by capping
// paragraph height, but capping each element individually never triggers
// when every individual element is already short — so the whole post (18
// headings) always renders in full. Wrapping all of "content" in one
// container lets a single max-height/overflow rule on that container clip
// the whole block together, regardless of how many elements it contains.
func TestWriteItem_ContentWrappedInSingleBodyContainer(t *testing.T) {
	gen := newTestGenerator()
	var buf bytes.Buffer
	sourceMarkdown := "## First bookmark\n\nShort note one.\n\n## Second bookmark\n\nShort note two."
	if _, err := gen.WriteItem(&buf, "https://example.com", "Title", "desc",
		nil, sourceMarkdown, nil, "guid1", "2020-01-01", "", "", "", "", "", "", ItemsConfig{}); err != nil {
		t.Fatalf("WriteItem: %s", err)
	}
	out := buf.String()
	start := strings.Index(out, `<div class="article-body">`)
	if start < 0 {
		t.Fatalf(`expected a <div class="article-body"> wrapper around content, got:\n%s`, out)
	}
	end := strings.Index(out, "</div>")
	if end < 0 || end < start {
		t.Fatalf("expected a closing </div> after the article-body wrapper, got:\n%s", out)
	}
	wrapped := out[start:end]
	if !strings.Contains(wrapped, "First bookmark") || !strings.Contains(wrapped, "Second bookmark") {
		t.Errorf("expected all rendered content elements (both headings) inside the single article-body wrapper, got:\n%s", out)
	}
}

func TestWriteItem_TitleAmpersandEscaped(t *testing.T) {
	gen := newTestGenerator()
	var buf bytes.Buffer
	if _, err := gen.WriteItem(&buf, "https://example.com", "Q&A: Test", "desc",
		nil, "", nil, "guid1", "2020-01-01", "", "", "", "", "", "", ItemsConfig{}); err != nil {
		t.Fatalf("WriteItem: %s", err)
	}
	out := buf.String()
	if strings.Contains(out, "<h2>Q&A: Test</h2>") {
		t.Errorf("expected title's & to be HTML-escaped, got raw ampersand:\n%s", out)
	}
	if !strings.Contains(out, "<h2>Q&amp;A: Test</h2>") {
		t.Errorf("expected escaped title in heading, got:\n%s", out)
	}
}

func TestWriteItem_LinkHrefEscaped(t *testing.T) {
	gen := newTestGenerator()
	var buf bytes.Buffer
	link := "https://example.com/?a=1&b=2"
	if _, err := gen.WriteItem(&buf, link, "Title", "desc",
		nil, "", nil, "guid1", "2020-01-01", "", "", "", "", "", "", ItemsConfig{}); err != nil {
		t.Fatalf("WriteItem: %s", err)
	}
	out := buf.String()
	if strings.Contains(out, `href="https://example.com/?a=1&b=2"`) {
		t.Errorf("expected href's & to be HTML-escaped, got raw ampersand:\n%s", out)
	}
	if !strings.Contains(out, `href="https://example.com/?a=1&amp;b=2"`) {
		t.Errorf("expected escaped href, got:\n%s", out)
	}
}

// TestWriteItem_TitleWithPreEscapedEntityNotDoubleEscaped guards against a
// regression found reviewing a real koreaherald.com item (10812786): its
// feed title already contains the literal 6-character text "&quot;" (the
// source double-HTML-escaped it before publishing), not a real quote
// character. html.EscapeString(title) alone escapes the "&" in that literal
// text again, producing visible "&amp;quot;" in the heading instead of a
// quote mark. Titles must be entity-normalized (unescaped, then escaped)
// before embedding so pre-existing entities collapse to their real
// character first.
func TestWriteItem_TitleWithPreEscapedEntityNotDoubleEscaped(t *testing.T) {
	gen := newTestGenerator()
	var buf bytes.Buffer
	title := `Say &quot;hello&quot; to me`
	if _, err := gen.WriteItem(&buf, "https://example.com", title, "desc",
		nil, "", nil, "guid1", "2020-01-01", "", "", "", "", "", "", ItemsConfig{}); err != nil {
		t.Fatalf("WriteItem: %s", err)
	}
	out := buf.String()
	if strings.Contains(out, "&amp;quot;") {
		t.Errorf("pre-existing entity in title must not be double-escaped, got:\n%s", out)
	}
	if !strings.Contains(out, "<h2>Say &#34;hello&#34; to me</h2>") {
		t.Errorf("expected pre-existing entity normalized to a real quote then re-escaped once, got:\n%s", out)
	}
}

// -------------------------------------------------------------------
// Phase 5: no-header warning
// -------------------------------------------------------------------

func TestWriteHTML_NoHeaderWarning(t *testing.T) {
	gen := newTestGenerator()
	// gen.Header is "" by default; redirect eout to capture the warning
	var errBuf bytes.Buffer
	gen.eout = &errBuf
	db := newTestItemsDB(t)
	defer db.Close()
	var outBuf bytes.Buffer
	if err := gen.WriteHTML(&outBuf, db, "", nil); err != nil {
		t.Fatalf("WriteHTML: %s", err)
	}
	warn := errBuf.String()
	if !strings.Contains(warn, "warning: aggregate page has no <h1>") {
		t.Errorf("expected no-h1 warning on stderr, got: %q", warn)
	}
}

func TestWriteHTML_NoWarningWhenHeaderSet(t *testing.T) {
	gen := newTestGenerator()
	gen.Header = "<h1>My Site</h1>"
	var errBuf bytes.Buffer
	gen.eout = &errBuf
	db := newTestItemsDB(t)
	defer db.Close()
	var outBuf bytes.Buffer
	if err := gen.WriteHTML(&outBuf, db, "", nil); err != nil {
		t.Fatalf("WriteHTML: %s", err)
	}
	warn := errBuf.String()
	if strings.Contains(warn, "warning: aggregate page has no <h1>") {
		t.Errorf("must not warn when Header is set, got: %q", warn)
	}
}

// -------------------------------------------------------------------
// Phase 2 (item formatting): truncateWords, stripTags, resolveItemContent
// -------------------------------------------------------------------

func TestTruncateWords(t *testing.T) {
	cases := []struct {
		name   string
		s      string
		maxLen int
		want   string
	}{
		{"shorter than max", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"mid-word cut avoided", "the quick brown fox", 12, "the quick"},
		{"no whitespace before limit", "supercalifragilistic", 5, "super"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := truncateWords(tc.s, tc.maxLen)
			if got != tc.want {
				t.Errorf("truncateWords(%q, %d) = %q, want %q", tc.s, tc.maxLen, got, tc.want)
			}
		})
	}
}

func TestStripTags(t *testing.T) {
	cases := []struct {
		name string
		s    string
		want string
	}{
		{"plain text", "hello", "hello"},
		{"simple tags", "<p>hello</p>", "hello"},
		{"nested tags", "<div><b>hi</b> there</div>", "hi there"},
		{"unclosed tag", "<p>hi", "hi"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := stripTags(tc.s)
			if got != tc.want {
				t.Errorf("stripTags(%q) = %q, want %q", tc.s, got, tc.want)
			}
		})
	}
}

func TestResolveItemContent(t *testing.T) {
	t.Run("markdown present, default strip", func(t *testing.T) {
		cfg := ItemsConfig{HTML: "strip"}
		got, isBlockHTML, err := resolveItemContent("<p>raw</p>", "**bold**", cfg)
		if err != nil {
			t.Fatalf("resolveItemContent: %s", err)
		}
		if !strings.Contains(got, "<strong>bold</strong>") {
			t.Errorf("expected rendered markdown, got %q", got)
		}
		if strings.Contains(got, "raw") {
			t.Errorf("raw description must not be used when sourceMarkdown is present, got %q", got)
		}
		if !isBlockHTML {
			t.Errorf("rendered markdown must be reported as block HTML")
		}
	})

	t.Run("markdown present, unsafe", func(t *testing.T) {
		cfg := ItemsConfig{HTML: "unsafe"}
		got, isBlockHTML, err := resolveItemContent("ignored", "before <script>x</script> after", cfg)
		if err != nil {
			t.Fatalf("resolveItemContent: %s", err)
		}
		if !strings.Contains(got, "<script>x</script>") {
			t.Errorf("expected raw <script> to pass through in unsafe mode, got %q", got)
		}
		if !isBlockHTML {
			t.Errorf("rendered markdown must be reported as block HTML")
		}
	})

	t.Run("no markdown, default strip", func(t *testing.T) {
		cfg := ItemsConfig{HTML: "strip"}
		got, isBlockHTML, err := resolveItemContent("<p>raw &amp; unsafe</p>", "", cfg)
		if err != nil {
			t.Fatalf("resolveItemContent: %s", err)
		}
		if strings.Contains(got, "<p>") {
			t.Errorf("expected tags stripped, got %q", got)
		}
		if isBlockHTML {
			t.Errorf("stripped plain text must not be reported as block HTML")
		}
	})

	t.Run("no markdown, escape", func(t *testing.T) {
		cfg := ItemsConfig{HTML: "escape"}
		got, isBlockHTML, err := resolveItemContent("<b>hi</b>", "", cfg)
		if err != nil {
			t.Fatalf("resolveItemContent: %s", err)
		}
		if got != "&lt;b&gt;hi&lt;/b&gt;" {
			t.Errorf("resolveItemContent escape = %q, want %q", got, "&lt;b&gt;hi&lt;/b&gt;")
		}
		if isBlockHTML {
			t.Errorf("escaped plain text must not be reported as block HTML")
		}
	})

	t.Run("no markdown, unsafe", func(t *testing.T) {
		cfg := ItemsConfig{HTML: "unsafe"}
		got, isBlockHTML, err := resolveItemContent("<b>hi</b>", "", cfg)
		if err != nil {
			t.Fatalf("resolveItemContent: %s", err)
		}
		if got != "<b>hi</b>" {
			t.Errorf("resolveItemContent unsafe = %q, want unchanged %q", got, "<b>hi</b>")
		}
		if !isBlockHTML {
			t.Errorf("unsafe passthrough must be reported as block HTML")
		}
	})

	t.Run("truncation applied pre-render", func(t *testing.T) {
		cfg := ItemsConfig{HTML: "strip", ContentMaxLength: 5}
		got, _, err := resolveItemContent("ignored", "one two three four five", cfg)
		if err != nil {
			t.Fatalf("resolveItemContent: %s", err)
		}
		if strings.Contains(got, "three") {
			t.Errorf("expected source truncated to ~5 chars before rendering, got %q", got)
		}
	})
}

// -------------------------------------------------------------------
// Phase 3 (item formatting): resolveItemLink
// -------------------------------------------------------------------

func TestResolveItemLink(t *testing.T) {
	defaults := func() LinkConfig {
		cfg := ItemsConfig{}
		cfg.applyDefaults()
		return cfg.Link
	}

	t.Run("normal, default label is accessible (DEC-026)", func(t *testing.T) {
		got, err := resolveItemLink("https://x", "T", "", defaults())
		if err != nil {
			t.Fatalf("resolveItemLink: %s", err)
		}
		want := LinkResolution{Href: "https://x", Label: "Continue reading"}
		if got != want {
			t.Errorf("got %#v, want %#v", got, want)
		}
	})

	t.Run("label_field: link restores pre-existing behavior", func(t *testing.T) {
		cfg := defaults()
		cfg.LabelField = "link"
		got, err := resolveItemLink("https://x", "T", "", cfg)
		if err != nil {
			t.Fatalf("resolveItemLink: %s", err)
		}
		if got.Label != "https://x" {
			t.Errorf("Label = %q, want %q", got.Label, "https://x")
		}
	})

	t.Run("label_field: title", func(t *testing.T) {
		cfg := defaults()
		cfg.LabelField = "title"
		got, err := resolveItemLink("https://x", "T", "", cfg)
		if err != nil {
			t.Fatalf("resolveItemLink: %s", err)
		}
		if got.Label != "T" {
			t.Errorf("Label = %q, want %q", got.Label, "T")
		}
	})

	t.Run("label_field: static with custom fallback", func(t *testing.T) {
		cfg := defaults()
		cfg.LabelFallback = "read me"
		got, err := resolveItemLink("https://x", "T", "", cfg)
		if err != nil {
			t.Fatalf("resolveItemLink: %s", err)
		}
		if got.Label != "read me" {
			t.Errorf("Label = %q, want %q", got.Label, "read me")
		}
	})

	t.Run("empty field falls back to label_fallback", func(t *testing.T) {
		cfg := defaults()
		cfg.LabelField = "title"
		cfg.LabelFallback = "Read more"
		got, err := resolveItemLink("https://x", "", "", cfg)
		if err != nil {
			t.Fatalf("resolveItemLink: %s", err)
		}
		if got.Label != "Read more" {
			t.Errorf("Label = %q, want %q", got.Label, "Read more")
		}
	})

	t.Run("missing link, default missing mode", func(t *testing.T) {
		got, err := resolveItemLink("", "T", "", defaults())
		if err != nil {
			t.Fatalf("resolveItemLink: %s", err)
		}
		if !got.AsPlainText || got.Omit {
			t.Errorf("got %#v, want AsPlainText=true, Omit=false", got)
		}
		if got.Label != "Continue reading" {
			t.Errorf("Label = %q, want %q", got.Label, "Continue reading")
		}
	})

	t.Run("missing link, omit", func(t *testing.T) {
		cfg := defaults()
		cfg.Missing = "omit"
		got, err := resolveItemLink("", "T", "", cfg)
		if err != nil {
			t.Fatalf("resolveItemLink: %s", err)
		}
		if !got.Omit {
			t.Errorf("got %#v, want Omit=true", got)
		}
	})

	t.Run("missing link, source_link with channel", func(t *testing.T) {
		cfg := defaults()
		cfg.Missing = "source_link"
		got, err := resolveItemLink("", "T", "https://chan", cfg)
		if err != nil {
			t.Fatalf("resolveItemLink: %s", err)
		}
		if got.Href != "https://chan" || got.AsPlainText {
			t.Errorf("got %#v, want Href=https://chan, AsPlainText=false", got)
		}
	})

	t.Run("missing link, source_link, channel also empty falls back to unlinked (DEC-027)", func(t *testing.T) {
		cfg := defaults()
		cfg.Missing = "source_link"
		got, err := resolveItemLink("", "T", "", cfg)
		if err != nil {
			t.Fatalf("resolveItemLink: %s", err)
		}
		if !got.AsPlainText || got.Href != "" {
			t.Errorf("got %#v, want AsPlainText=true, Href empty", got)
		}
	})

	t.Run("missing link, required returns error", func(t *testing.T) {
		cfg := defaults()
		cfg.Required = true
		_, err := resolveItemLink("", "T", "", cfg)
		if err == nil {
			t.Error("expected error when required link is missing, got nil")
		}
	})
}

// -------------------------------------------------------------------
// Phase 4 (item formatting): formatItemDate
// -------------------------------------------------------------------

func TestFormatItemDate(t *testing.T) {
	cases := []struct {
		name   string
		raw    string
		layout string
		want   string
	}{
		{"parseable, custom layout", "2020-04-11 00:00:00", "Jan 2, 2006", "Apr 11, 2020"},
		{"parseable, default layout", "2020-04-11 00:00:00", "2006-01-02", "2020-04-11"},
		{"parseable, RFC3339 (production sqlite driver's actual return shape)", "2020-04-11T00:00:00Z", "Jan 2, 2006", "Apr 11, 2020"},
		{"unparseable, longer than 10 chars (DEC-028)", "Sat, 11 Apr 2020", "Jan 2, 2006", "Sat, 11 Ap"},
		{"unparseable, 10 chars or shorter", "bad-date", "Jan 2, 2006", "bad-date"},
		{"empty", "", "Jan 2, 2006", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := formatItemDate(tc.raw, tc.layout)
			if got != tc.want {
				t.Errorf("formatItemDate(%q, %q) = %q, want %q", tc.raw, tc.layout, got, tc.want)
			}
		})
	}
}
