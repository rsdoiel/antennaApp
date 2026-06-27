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
	"strings"
	"testing"

	"github.com/mmcdole/gofeed"
)

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
	err := gen.WriteItem(&buf, "https://example.com", "Title", "desc",
		nil, "", nil, "guid1", "", "", "", "", "", "", "")
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
	err := gen.WriteItem(&buf, "https://example.com", "Title", "desc",
		nil, "", nil, "guid1", "", "", "", "", "", "", `["Oberon"]`)
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
	err := gen.WriteItem(&buf, "https://example.com", "Title", "desc",
		nil, "", nil, "guid1", "", "", "", "", "", "", `["a","b"]`)
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
	err := gen.WriteItem(&buf, "https://example.com", "Title", "desc",
		nil, "", nil, "guid1", "", dcExt, "", "", "", "", "")
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
	err := gen.WriteItem(&buf, "https://example.com", "Title", "desc",
		authors, "", nil, "guid1", "", "", "", "", "", "", "")
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
	err := gen.WriteItem(&buf, "https://example.com", "Title", "desc",
		nil, "", nil, "guid1", "", "", "https://example.com/feed.xml", "", "", "My Feed", "")
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
	err := gen.WriteItem(&buf, "https://example.com", "Title", "desc",
		nil, "", nil, "guid1", "2020-01-01", "", "", "", "", "", "")
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
	err := gen.WriteItem(&buf, "https://example.com", "Title", "desc",
		authors, "", nil, "guid1", "2020-06-01", dcExt,
		"https://example.com/feed.xml", "", "", "Feed Label", `["tech"]`)
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
