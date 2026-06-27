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
	"testing"
)

func TestGetAttributeStringSlice(t *testing.T) {
	cases := []struct {
		name     string
		fm       map[string]interface{}
		key      string
		expected []string
	}{
		{
			name:     "key absent",
			fm:       map[string]interface{}{},
			key:      "keywords",
			expected: nil,
		},
		{
			name:     "empty string value",
			fm:       map[string]interface{}{"keywords": ""},
			key:      "keywords",
			expected: nil,
		},
		{
			name:     "non-empty string",
			fm:       map[string]interface{}{"keywords": "Oberon"},
			key:      "keywords",
			expected: []string{"Oberon"},
		},
		{
			name:     "sequence of strings",
			fm:       map[string]interface{}{"keywords": []interface{}{"a", "b"}},
			key:      "keywords",
			expected: []string{"a", "b"},
		},
		{
			name:     "sequence with empty entries",
			fm:       map[string]interface{}{"keywords": []interface{}{"a", "", "b"}},
			key:      "keywords",
			expected: []string{"a", "b"},
		},
		{
			name:     "non-string non-slice type",
			fm:       map[string]interface{}{"keywords": 42},
			key:      "keywords",
			expected: nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			doc := &CommonMark{FrontMatter: tc.fm}
			got := doc.GetAttributeStringSlice(tc.key)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("got %v, want %v", got, tc.expected)
			}
		})
	}
}
