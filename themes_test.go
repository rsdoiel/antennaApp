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
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestNewTheme_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &AppConfig{Htdocs: tmpDir}
	var buf bytes.Buffer
	if err := cfg.NewTheme(&buf, "mytheme"); err != nil {
		t.Fatalf("NewTheme: %s", err)
	}
	themePath := filepath.Join(tmpDir, "mytheme")
	if _, err := os.Stat(themePath); err != nil {
		t.Errorf("expected theme directory %s to exist: %s", themePath, err)
	}
}

func TestNewTheme_CreatesSkeletonFiles(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &AppConfig{Htdocs: tmpDir}
	var buf bytes.Buffer
	if err := cfg.NewTheme(&buf, "mytheme"); err != nil {
		t.Fatalf("NewTheme: %s", err)
	}
	themePath := filepath.Join(tmpDir, "mytheme")
	expected := []string{
		"header.md",
		"nav.md",
		"footer.md",
		"head.yaml",
	}
	for _, name := range expected {
		p := filepath.Join(themePath, name)
		if _, err := os.Stat(p); err != nil {
			t.Errorf("expected skeleton file %s to exist: %s", p, err)
		}
	}
}

func TestNewTheme_DefaultNameIsTheme(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &AppConfig{Htdocs: tmpDir}
	var buf bytes.Buffer
	if err := cfg.NewTheme(&buf, ""); err != nil {
		t.Fatalf("NewTheme with empty name: %s", err)
	}
	themePath := filepath.Join(tmpDir, "theme")
	if _, err := os.Stat(themePath); err != nil {
		t.Errorf("expected default theme directory 'theme' to exist: %s", err)
	}
}

func TestNewTheme_DoesNotOverwriteExistingFiles(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &AppConfig{Htdocs: tmpDir}

	// First creation
	var buf bytes.Buffer
	if err := cfg.NewTheme(&buf, "mytheme"); err != nil {
		t.Fatalf("first NewTheme: %s", err)
	}

	// Overwrite header.md with custom content
	headerPath := filepath.Join(tmpDir, "mytheme", "header.md")
	customContent := "# My Custom Header\n"
	if err := os.WriteFile(headerPath, []byte(customContent), 0664); err != nil {
		t.Fatalf("write custom header: %s", err)
	}

	// Second creation — must not overwrite
	buf.Reset()
	if err := cfg.NewTheme(&buf, "mytheme"); err != nil {
		t.Fatalf("second NewTheme: %s", err)
	}
	got, _ := os.ReadFile(headerPath)
	if string(got) != customContent {
		t.Errorf("expected custom header to be preserved, got: %q", string(got))
	}
	// Progress output should mention "skipped" or "exists"
	if !strings.Contains(buf.String(), "exist") && !strings.Contains(buf.String(), "skip") {
		t.Errorf("expected 'exists' or 'skip' in output, got: %q", buf.String())
	}
}

func TestNewTheme_HeaderMdContainsSiteTitle(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &AppConfig{Htdocs: tmpDir}
	var buf bytes.Buffer
	if err := cfg.NewTheme(&buf, "mytheme"); err != nil {
		t.Fatalf("NewTheme: %s", err)
	}
	headerPath := filepath.Join(tmpDir, "mytheme", "header.md")
	content, _ := os.ReadFile(headerPath)
	if !strings.Contains(string(content), "#") {
		t.Errorf("expected header.md to contain a heading, got: %q", string(content))
	}
}

func TestNewTheme_HeadYamlContainsLinkEntry(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &AppConfig{Htdocs: tmpDir}
	var buf bytes.Buffer
	if err := cfg.NewTheme(&buf, "mytheme"); err != nil {
		t.Fatalf("NewTheme: %s", err)
	}
	headPath := filepath.Join(tmpDir, "mytheme", "head.yaml")
	content, _ := os.ReadFile(headPath)
	if !strings.Contains(string(content), "link:") {
		t.Errorf("expected head.yaml to contain a link: section, got: %q", string(content))
	}
}

func TestNewTheme_ReportsCreatedFiles(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &AppConfig{Htdocs: tmpDir}
	var buf bytes.Buffer
	if err := cfg.NewTheme(&buf, "mytheme"); err != nil {
		t.Fatalf("NewTheme: %s", err)
	}
	out := buf.String()
	if !strings.Contains(out, "mytheme") {
		t.Errorf("expected output to mention theme name, got: %q", out)
	}
}

// -------------------------------------------------------------------
// Phase 6 (item formatting): items.yaml theme integration
// -------------------------------------------------------------------

func TestIsTargetFile_ItemsYAML(t *testing.T) {
	if !isTargetFile("items.yaml") {
		t.Error(`isTargetFile("items.yaml") = false, want true`)
	}
}

func TestUpdateItemsElement_NoFilePresent(t *testing.T) {
	tmpDir := t.TempDir()
	gen := &Generator{}
	changed, err := updateItemsElement(gen, tmpDir)
	if err != nil {
		t.Fatalf("updateItemsElement: %s", err)
	}
	if changed {
		t.Error("expected changed=false when items.yaml is absent")
	}
	if !reflect.DeepEqual(gen.Items, ItemsConfig{}) {
		t.Errorf("gen.Items = %#v, want zero value unchanged", gen.Items)
	}
}

func TestUpdateItemsElement_FilePresent(t *testing.T) {
	tmpDir := t.TempDir()
	src := `fields:
  - title
  - content
link:
  label_fallback: "read me"
date_format: "Jan 2, 2006"
content_max_length: 320
show_source: true
html: strip
`
	if err := os.WriteFile(filepath.Join(tmpDir, "items.yaml"), []byte(src), 0664); err != nil {
		t.Fatalf("write items.yaml: %s", err)
	}
	gen := &Generator{}
	changed, err := updateItemsElement(gen, tmpDir)
	if err != nil {
		t.Fatalf("updateItemsElement: %s", err)
	}
	if !changed {
		t.Error("expected changed=true when items.yaml is present")
	}
	if len(gen.Items.Fields) != 2 || gen.Items.Fields[0] != "title" || gen.Items.Fields[1] != "content" {
		t.Errorf("gen.Items.Fields = %#v, want [title content]", gen.Items.Fields)
	}
	if gen.Items.Link.LabelFallback != "read me" {
		t.Errorf("gen.Items.Link.LabelFallback = %q, want %q", gen.Items.Link.LabelFallback, "read me")
	}
	if gen.Items.DateFormat != "Jan 2, 2006" {
		t.Errorf("gen.Items.DateFormat = %q, want %q", gen.Items.DateFormat, "Jan 2, 2006")
	}
	if gen.Items.ContentMaxLength != 320 {
		t.Errorf("gen.Items.ContentMaxLength = %d, want 320", gen.Items.ContentMaxLength)
	}
}

func TestUpdateItemsElement_MalformedYAML(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "items.yaml"), []byte("fields: [unterminated"), 0664); err != nil {
		t.Fatalf("write items.yaml: %s", err)
	}
	gen := &Generator{}
	_, err := updateItemsElement(gen, tmpDir)
	if err == nil {
		t.Error("expected error for malformed items.yaml, got nil")
	}
}
