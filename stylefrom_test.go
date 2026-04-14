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

func TestExtractStyles_HTML(t *testing.T) {
	app := NewAntennaApp("antenna")
	out := &bytes.Buffer{}
	outputPath := filepath.Join(t.TempDir(), "style.css")

	if err := app.ExtractStyles(out, []string{"testdata/index.html", outputPath}); err != nil {
		t.Fatalf("ExtractStyles(.html) error: %v", err)
	}

	css, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("reading output file: %v", err)
	}
	if strings.TrimSpace(string(css)) == "" {
		t.Fatal("expected non-empty CSS, got empty file")
	}
	if !strings.Contains(out.String(), outputPath) {
		t.Errorf("expected output message to contain %q, got %q", outputPath, out.String())
	}
}

func TestExtractStyles_ODT(t *testing.T) {
	app := NewAntennaApp("antenna")
	out := &bytes.Buffer{}
	outputPath := filepath.Join(t.TempDir(), "style.css")

	if err := app.ExtractStyles(out, []string{"testdata/index.odt", outputPath}); err != nil {
		t.Fatalf("ExtractStyles(.odt) error: %v", err)
	}

	css, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("reading output file: %v", err)
	}
	if strings.TrimSpace(string(css)) == "" {
		t.Fatal("expected non-empty CSS, got empty file")
	}
}

func TestExtractStyles_DefaultOutputPath(t *testing.T) {
	app := NewAntennaApp("antenna")
	out := &bytes.Buffer{}

	// Change to a temp directory so the default "theme/style.css" is written there.
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	// Use an absolute path to the fixture since we changed directories.
	fixture, err := filepath.Abs(filepath.Join(orig, "testdata/index.html"))
	if err != nil {
		t.Fatal(err)
	}

	if err := app.ExtractStyles(out, []string{fixture}); err != nil {
		t.Fatalf("ExtractStyles(default output) error: %v", err)
	}

	defaultOut := filepath.Join(tmp, "theme", "style.css")
	if _, err := os.Stat(defaultOut); err != nil {
		t.Errorf("expected output at %q, stat error: %v", defaultOut, err)
	}
}

func TestExtractStyles_MissingInputFile(t *testing.T) {
	app := NewAntennaApp("antenna")
	out := &bytes.Buffer{}
	err := app.ExtractStyles(out, []string{"testdata/does-not-exist.html"})
	if err == nil {
		t.Error("expected error for missing input file, got nil")
	}
}

func TestExtractStyles_NoArgs(t *testing.T) {
	app := NewAntennaApp("antenna")
	out := &bytes.Buffer{}
	err := app.ExtractStyles(out, []string{})
	if err == nil {
		t.Error("expected error when no args provided, got nil")
	}
}

func TestExtractStyles_UnknownExtension(t *testing.T) {
	app := NewAntennaApp("antenna")
	out := &bytes.Buffer{}
	err := app.ExtractStyles(out, []string{"testdata/index.html", filepath.Join(t.TempDir(), "style.css")})
	// Reuse a known-good call to confirm no false positive, then test the bad extension.
	if err != nil {
		t.Fatalf("setup call failed: %v", err)
	}

	err = app.ExtractStyles(out, []string{"testdata/index.xyz"})
	if err == nil {
		t.Error("expected error for unrecognised file extension, got nil")
	}
}
