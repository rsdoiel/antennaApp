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
