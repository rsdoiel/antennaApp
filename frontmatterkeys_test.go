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
	"strings"
	"testing"
)

// BonBon's eight front matter keys, in the order its documents list them
// (DR-0009). A change to BonBon's list is one edit to frontMatterKeys and here.
var bonbonEight = []string{"title", "description", "author", "pubDate", "link", "urlPrev", "urlNext", "rightMarginImage"}

func TestFrontMatterKeysHoldBonBonsEight(t *testing.T) {
	var got []string
	for _, k := range frontMatterKeys {
		if k.BonBon {
			got = append(got, k.Name)
		}
	}
	if strings.Join(got, ",") != strings.Join(bonbonEight, ",") {
		t.Errorf("BonBon keys = %v, want %v", got, bonbonEight)
	}
	for _, k := range frontMatterKeys {
		if k.Name == "" || k.Summary == "" {
			t.Errorf("key %+v needs a name and a summary", k)
		}
	}
}

func TestEveryKeyIsDocumentedInTheMetadataManual(t *testing.T) {
	for _, k := range frontMatterKeys {
		if !strings.Contains(MetadataHelpText, "\n"+k.Name+"\n") {
			t.Errorf("antenna help metadata does not document %s", k.Name)
		}
	}
}

func TestKeysTheHTMLRendersItselfAreNotEchoedAsMeta(t *testing.T) {
	for _, name := range []string{"title", "urlPrev", "urlNext", "rightMarginImage", "rightMarginImageAlt"} {
		if !frontMatterKeyRendered(name) {
			t.Errorf("%s is written by the renderer and must not also be a <meta>", name)
		}
	}
	for _, name := range []string{"description", "author", "pubDate", "link", "keywords", "series", "unknownKey"} {
		if frontMatterKeyRendered(name) {
			t.Errorf("%s is a <meta> key", name)
		}
	}
	// the navigation keys come from the table
	for _, nk := range pageNavKeys {
		if !frontMatterKeyRendered(nk.Key) {
			t.Errorf("%s is in pageNavKeys but not in frontMatterKeys", nk.Key)
		}
	}
}

func TestIsBonBonKey(t *testing.T) {
	for _, name := range bonbonEight {
		if !isBonBonKey(name) {
			t.Errorf("%s is one of BonBon's eight", name)
		}
	}
	for _, name := range []string{"keywords", "rightMarginImageAlt", "postPath", "source"} {
		if isBonBonKey(name) {
			t.Errorf("%s is not one of BonBon's eight", name)
		}
	}
}
