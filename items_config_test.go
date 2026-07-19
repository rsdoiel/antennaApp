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
	"reflect"
	"strings"
	"testing"
)

func TestItemsConfigDefaults(t *testing.T) {
	wantFields := []string{"title", "source", "pubDate", "content"}

	t.Run("empty block", func(t *testing.T) {
		cfg := ItemsConfig{}
		cfg.applyDefaults()
		if !reflect.DeepEqual(cfg.Fields, wantFields) {
			t.Errorf("Fields = %#v, want %#v", cfg.Fields, wantFields)
		}
		if cfg.Link.LabelField != "static" {
			t.Errorf("Link.LabelField = %q, want %q", cfg.Link.LabelField, "static")
		}
		if cfg.Link.LabelFallback != "Continue reading" {
			t.Errorf("Link.LabelFallback = %q, want %q", cfg.Link.LabelFallback, "Continue reading")
		}
		if cfg.Link.Missing != "unlinked" {
			t.Errorf("Link.Missing = %q, want %q", cfg.Link.Missing, "unlinked")
		}
		if cfg.DateFormat != "2006-01-02" {
			t.Errorf("DateFormat = %q, want %q", cfg.DateFormat, "2006-01-02")
		}
		if cfg.ShowSource == nil || !*cfg.ShowSource {
			t.Errorf("ShowSource = %v, want true", cfg.ShowSource)
		}
		if cfg.HTML != "strip" {
			t.Errorf("HTML = %q, want %q", cfg.HTML, "strip")
		}
	})

	t.Run("partial override", func(t *testing.T) {
		cfg := ItemsConfig{HTML: "unsafe"}
		cfg.applyDefaults()
		if cfg.HTML != "unsafe" {
			t.Errorf("HTML = %q, want %q (explicit value must survive defaulting)", cfg.HTML, "unsafe")
		}
		if cfg.Link.LabelField != "static" {
			t.Errorf("Link.LabelField = %q, want %q (unset fields still default)", cfg.Link.LabelField, "static")
		}
		if cfg.DateFormat != "2006-01-02" {
			t.Errorf("DateFormat = %q, want %q", cfg.DateFormat, "2006-01-02")
		}
	})

	t.Run("explicit show_source false is not overwritten", func(t *testing.T) {
		f := false
		cfg := ItemsConfig{ShowSource: &f}
		cfg.applyDefaults()
		if cfg.ShowSource == nil || *cfg.ShowSource {
			t.Errorf("ShowSource = %v, want false (explicit false must survive defaulting)", cfg.ShowSource)
		}
	})

	t.Run("zero value produces same defaults as empty block", func(t *testing.T) {
		var cfg ItemsConfig // no items: key in page.yaml at all
		cfg.applyDefaults()
		if !reflect.DeepEqual(cfg.Fields, wantFields) {
			t.Errorf("Fields = %#v, want %#v", cfg.Fields, wantFields)
		}
		if cfg.Link.LabelField != "static" || cfg.Link.LabelFallback != "Continue reading" {
			t.Errorf("Link = %#v, want static/Continue reading defaults", cfg.Link)
		}
	})
}

func TestItemsConfigValidate(t *testing.T) {
	cases := []struct {
		name    string
		cfg     ItemsConfig
		wantErr string // substring expected in error, "" means no error
	}{
		{
			name:    "invalid html",
			cfg:     ItemsConfig{HTML: "bogus", Link: LinkConfig{Missing: "unlinked"}},
			wantErr: "items.html",
		},
		{
			name:    "invalid missing",
			cfg:     ItemsConfig{HTML: "strip", Link: LinkConfig{Missing: "bogus"}},
			wantErr: "items.link.missing",
		},
		{
			name:    "valid strip/unlinked",
			cfg:     ItemsConfig{HTML: "strip", Link: LinkConfig{Missing: "unlinked"}},
			wantErr: "",
		},
		{
			name:    "valid escape/omit",
			cfg:     ItemsConfig{HTML: "escape", Link: LinkConfig{Missing: "omit"}},
			wantErr: "",
		},
		{
			name:    "valid unsafe/source_link",
			cfg:     ItemsConfig{HTML: "unsafe", Link: LinkConfig{Missing: "source_link"}},
			wantErr: "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.validate()
			if tc.wantErr == "" {
				if err != nil {
					t.Errorf("validate() = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("validate() = nil, want error containing %q", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("validate() = %q, want it to contain %q", err.Error(), tc.wantErr)
			}
		})
	}
}
