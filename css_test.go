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
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateCSS_WritesFile(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &AppConfig{Htdocs: tmpDir}
	var buf bytes.Buffer
	if err := cfg.GenerateCSS(&buf, "css/site.css"); err != nil {
		t.Fatalf("GenerateCSS: %s", err)
	}
	cssPath := filepath.Join(tmpDir, "css", "site.css")
	if _, err := os.Stat(cssPath); err != nil {
		t.Errorf("expected %s to exist: %s", cssPath, err)
	}
}

func TestGenerateCSS_ContentHasSkipLink(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &AppConfig{Htdocs: tmpDir}
	var buf bytes.Buffer
	if err := cfg.GenerateCSS(&buf, "css/site.css"); err != nil {
		t.Fatalf("GenerateCSS: %s", err)
	}
	content, _ := os.ReadFile(filepath.Join(tmpDir, "css", "site.css"))
	out := string(content)
	if !strings.Contains(out, ".skip-link") {
		t.Errorf("expected .skip-link rule in generated CSS")
	}
	if !strings.Contains(out, ".skip-link:focus") {
		t.Errorf("expected .skip-link:focus rule in generated CSS")
	}
}

func TestGenerateCSS_ContentHasCustomProperties(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &AppConfig{Htdocs: tmpDir}
	var buf bytes.Buffer
	if err := cfg.GenerateCSS(&buf, "css/site.css"); err != nil {
		t.Fatalf("GenerateCSS: %s", err)
	}
	content, _ := os.ReadFile(filepath.Join(tmpDir, "css", "site.css"))
	out := string(content)
	if !strings.Contains(out, ":root") {
		t.Errorf("expected :root block with custom properties")
	}
	if !strings.Contains(out, "article footer") {
		t.Errorf("expected article footer rule (replaces old article address)")
	}
}

func TestGenerateCSS_BacksUpExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	cssDir := filepath.Join(tmpDir, "css")
	os.MkdirAll(cssDir, 0775)
	cssPath := filepath.Join(cssDir, "site.css")
	os.WriteFile(cssPath, []byte("/* old content */"), 0664)

	cfg := &AppConfig{Htdocs: tmpDir}
	var buf bytes.Buffer
	if err := cfg.GenerateCSS(&buf, "css/site.css"); err != nil {
		t.Fatalf("GenerateCSS: %s", err)
	}

	bakPath := cssPath + ".bak"
	bakContent, err := os.ReadFile(bakPath)
	if err != nil {
		t.Fatalf("expected backup at %s: %s", bakPath, err)
	}
	if string(bakContent) != "/* old content */" {
		t.Errorf("backup should contain old content, got: %s", string(bakContent))
	}
	newContent, _ := os.ReadFile(cssPath)
	if string(newContent) == "/* old content */" {
		t.Error("css/site.css should have new content after overwrite")
	}
}

func TestGenerateCSS_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &AppConfig{Htdocs: tmpDir}
	var buf bytes.Buffer
	if err := cfg.GenerateCSS(&buf, "css/sub/site.css"); err != nil {
		t.Fatalf("GenerateCSS with nested path: %s", err)
	}
	cssPath := filepath.Join(tmpDir, "css", "sub", "site.css")
	if _, err := os.Stat(cssPath); err != nil {
		t.Errorf("expected nested path %s to exist: %s", cssPath, err)
	}
}

func TestPatchGeneratorYAML_AppendsLinkBlock(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "page.yaml")
	os.WriteFile(yamlPath, []byte("title: My Site\n"), 0664)

	patched, _, err := patchGeneratorYAML(yamlPath, "/css/site.css")
	if err != nil {
		t.Fatalf("patchGeneratorYAML: %s", err)
	}
	if !patched {
		t.Error("expected file to be patched when no link: section exists")
	}
	content, _ := os.ReadFile(yamlPath)
	if !strings.Contains(string(content), "link:") {
		t.Errorf("expected link: section in patched YAML:\n%s", content)
	}
	if !strings.Contains(string(content), "/css/site.css") {
		t.Errorf("expected CSS href in patched YAML:\n%s", content)
	}
}

func TestPatchGeneratorYAML_SkipsIfPresent(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "page.yaml")
	os.WriteFile(yamlPath, []byte("title: My Site\nlink:\n  - rel: stylesheet\n    href: /css/site.css\n"), 0664)

	patched, msg, err := patchGeneratorYAML(yamlPath, "/css/site.css")
	if err != nil {
		t.Fatalf("patchGeneratorYAML: %s", err)
	}
	if patched {
		t.Error("should not patch when CSS href already present")
	}
	if !strings.Contains(msg, "already") {
		t.Errorf("expected 'already' in message, got: %q", msg)
	}
}

func TestPatchGeneratorYAML_InstructsWhenLinkSectionExists(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "page.yaml")
	os.WriteFile(yamlPath, []byte("title: My Site\nlink:\n  - rel: alternate\n    href: /feed.xml\n"), 0664)

	patched, msg, err := patchGeneratorYAML(yamlPath, "/css/site.css")
	if err != nil {
		t.Fatalf("patchGeneratorYAML: %s", err)
	}
	if patched {
		t.Error("should not rewrite file when link: section already exists")
	}
	if !strings.Contains(msg, "/css/site.css") {
		t.Errorf("expected CSS href in instruction message, got: %q", msg)
	}
}
