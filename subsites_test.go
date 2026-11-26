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
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// runInBuf is runIn that also returns what the action wrote to standard error.
func runInBuf(t *testing.T, eout *bytes.Buffer, action string, args ...string) error {
	t.Helper()
	old := os.Stdout
	if null, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0); err == nil {
		os.Stdout = null
		defer func() { os.Stdout = old; null.Close() }()
	}
	app := &AntennaApp{appName: "antenna"}
	var out bytes.Buffer
	return app.Run(strings.NewReader(""), &out, eout, "antenna.yaml", action, args)
}

func mustTime(s string) time.Time {
	ts, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return ts
}

// subsiteSite is a workspace made by antenna init whose antenna.yaml has a
// base_url, an htdocs folder and the extra YAML given (DR-0015). It returns
// htdocs.
func subsiteSite(t *testing.T, extraYAML string) (htdocs string) {
	t.Helper()
	newWorkspace(t, true)
	src, _ := os.ReadFile("antenna.yaml")
	edited := strings.Replace(string(src), "base_url: http://localhost:8000", "base_url: https://example.org\nhtdocs: htdocs", 1)
	if edited == string(src) {
		t.Fatalf("antenna.yaml is not what the test expects:\n%s", src)
	}
	if err := os.WriteFile("antenna.yaml", []byte(edited+"\n"+extraYAML+"\n"), 0o664); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll("htdocs", 0o775); err != nil {
		t.Fatal(err)
	}
	return "htdocs"
}

const twoSubsites = `subsites:
  - url: https://example.org/antenna
  - url: https://example.org/opml/
    title: opml
`

// DR-0015 1: the lists are read from a real antenna.yaml.
func TestSubsitesAndBonbonAreReadFromAntennaYAML(t *testing.T) {
	subsiteSite(t, twoSubsites+"bonbon:\n  urlSearch: https://example.org/search\n")
	cfg := &AppConfig{}
	if err := cfg.LoadConfig("antenna.yaml"); err != nil {
		t.Fatalf("LoadConfig: %s", err)
	}
	if len(cfg.Subsites) != 2 || cfg.Subsites[0].URL != "https://example.org/antenna" ||
		cfg.Subsites[1].Title != "opml" {
		t.Errorf("subsites %+v", cfg.Subsites)
	}
	if cfg.Bonbon["urlSearch"] != "https://example.org/search" {
		t.Errorf("bonbon %+v", cfg.Bonbon)
	}
}

// DR-0015 2, 5: every mistake in the new settings is a configuration error
// (78), found when the configuration loads, whatever the command.
func TestBadSubsitesAndBonbonAreConfigErrors(t *testing.T) {
	for label, yaml := range map[string]string{
		"no url":              "subsites:\n  - title: x\n",
		"relative url":        "subsites:\n  - url: /opml\n",
		"no host":             "subsites:\n  - url: https:///opml\n",
		"ftp":                 "subsites:\n  - url: ftp://example.org/opml\n",
		"query":               "subsites:\n  - url: https://example.org/opml?a=b\n",
		"fragment":            "subsites:\n  - url: https://example.org/opml#top\n",
		"duplicate":           "subsites:\n  - url: https://example.org/opml\n  - url: https://example.org/opml/\n",
		"not a list":          "subsites: https://example.org/opml\n",
		"entry is not a map":  "subsites:\n  - https://example.org/opml\n",
		"bonbon is not a map": "bonbon: https://example.org/search\n",
		"bonbon is a list":    "bonbon:\n  - urlSearch\n",
	} {
		subsiteSite(t, yaml)
		wantClass(t, label, runIn(t, "list"), classConfig)
	}
}

// DR-0015 3: includes follow the groups, in the order listed, with type, text
// and the table of contents address, and the file parses back.
func TestSubsitesAreWrittenAsIncludeOutlinesAfterTheGroups(t *testing.T) {
	cfg, htdocs := tocSite(t)
	cfg.Subsites = []Subsite{{URL: "https://example.org/antenna"}, {URL: "https://example.org/opml/", Title: "opml"}}
	var eout bytes.Buffer
	if err := cfg.GenerateTOC(&eout); err != nil {
		t.Fatalf("GenerateTOC: %s\n%s", err, eout.String())
	}
	o, raw := readTOC(t, htdocs)
	if got := strings.Join(texts(o.Body.Outline), ","); got != "Pages,Collections,Posts,example.org/antenna,opml" {
		t.Fatalf("top level %q\n%s", got, raw)
	}
	for i, want := range []struct{ text, url string }{
		{"example.org/antenna", "https://example.org/antenna/.well-known/toc.opml"},
		{"opml", "https://example.org/opml/.well-known/toc.opml"},
	} {
		inc := o.Body.Outline[3+i]
		if inc.Type != "include" || inc.Text != want.text || inc.URL != want.url || len(inc.Outline) != 0 {
			t.Errorf("include %d: %+v, want include %q %q", i, inc, want.text, want.url)
		}
	}
	// the derived previous and next ignore includes: the newest post's next is
	// the oldest post (the ends wrap, DR-0040), never an include
	if _, next := neighboursFromNav(t, cfg, "blog/2026/11/19/c.md"); next != "blog/2025/12/25/old.md" {
		t.Errorf("an include became the next page: %q", next)
	}
}

// neighboursFromNav returns the previous and next the derived navigation
// holds for a published Markdown page.
func neighboursFromNav(t *testing.T, cfg *AppConfig, md string) (prev, next string) {
	t.Helper()
	nav := cfg.computeTOCNav()
	n, ok := nav[md]
	if !ok {
		t.Fatalf("%s is not in the navigation", md)
	}
	return n.prev, n.next
}

// DR-0015 3: sub-sites alone make a file worth writing; the file still needs
// base_url.
func TestASiteWithOnlySubsitesStillGetsATOC(t *testing.T) {
	htdocs := subsiteSite(t, twoSubsites)
	if err := runIn(t, "generate"); err != nil {
		t.Fatalf("generate: %s", err)
	}
	o, _ := readTOC(t, htdocs)
	if got := strings.Join(texts(o.Body.Outline), ","); got != "example.org/antenna,opml" {
		t.Errorf("top level %q", got)
	}
}

func TestSubsitesNeedABaseURLLikeTheRestOfTheTOC(t *testing.T) {
	cfg := sitemapSite(t, "pages.md")
	cfg.Htdocs = filepath.Join(t.TempDir(), "htdocs")
	cfg.BaseURL = ""
	cfg.Subsites = []Subsite{{URL: "https://example.org/opml"}}
	var eout bytes.Buffer
	if err := cfg.GenerateTOC(&eout); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(cfg.Htdocs, ".well-known", "toc.opml")); err == nil {
		t.Error("toc.opml written with no base_url")
	}
	if !strings.Contains(eout.String(), "base_url") {
		t.Errorf("no warning: %q", eout.String())
	}
}

// DR-0015 4: generate makes no network request. A sub-site that does not
// exist is listed all the same, and the run exits 0.
func TestGenerateDoesNotTouchTheNetworkForSubsites(t *testing.T) {
	htdocs := subsiteSite(t, "subsites:\n  - url: http://127.0.0.1:1/nothing-listens-here\n")
	if err := runIn(t, "generate"); err != nil {
		t.Fatalf("generate: %s", err)
	}
	if _, raw := readTOC(t, htdocs); !strings.Contains(raw, "http://127.0.0.1:1/nothing-listens-here/.well-known/toc.opml") {
		t.Errorf("include missing:\n%s", raw)
	}
}

// DR-0015 5: bonbon.json is a sorted, indented JSON object with a final
// newline, values as written, unknown keys passed through.
func TestBonbonJSONIsWrittenFromTheBonbonMapping(t *testing.T) {
	htdocs := subsiteSite(t, "bonbon:\n  urlSearch: https://example.org/search\n  urlMirrorOf: http://example.org/\n  someFutureKey: 3\n")
	var eout bytes.Buffer
	if err := runInBuf(t, &eout, "generate"); err != nil {
		t.Fatalf("generate: %s\n%s", err, eout.String())
	}
	b, err := os.ReadFile(filepath.Join(htdocs, ".well-known", "bonbon.json"))
	if err != nil {
		t.Fatalf("bonbon.json not written: %s", err)
	}
	want := "{\n  \"someFutureKey\": 3,\n  \"urlMirrorOf\": \"http://example.org/\",\n  \"urlSearch\": \"https://example.org/search\"\n}\n"
	if string(b) != want {
		t.Errorf("bonbon.json:\n%s\nwant:\n%s", b, want)
	}
	if !json.Valid(b) {
		t.Error("not valid JSON")
	}
	if strings.Contains(eout.String(), "bonbon") {
		t.Errorf("warning for a good block: %s", eout.String())
	}
}

// DR-0015 5: a known key BonBon would drop, or keep and break, is warned about
// and still written as the author wrote it.
func TestBonbonKnownKeysWithBadValuesWarn(t *testing.T) {
	htdocs := subsiteSite(t, "bonbon:\n  urlSearch: \"\"\n  urlUpdateSocket: 12\n  urlMirrorOf: http://example.org/\n")
	var eout bytes.Buffer
	if err := runInBuf(t, &eout, "generate"); err != nil {
		t.Fatalf("generate: %s", err)
	}
	for _, key := range []string{"urlSearch", "urlUpdateSocket"} {
		if !strings.Contains(eout.String(), key) {
			t.Errorf("no warning for %s: %q", key, eout.String())
		}
	}
	if strings.Contains(eout.String(), "urlMirrorOf") {
		t.Errorf("warned about a good key: %q", eout.String())
	}
	b, _ := os.ReadFile(filepath.Join(htdocs, ".well-known", "bonbon.json"))
	if !strings.Contains(string(b), `"urlSearch": ""`) || !strings.Contains(string(b), `"urlUpdateSocket": 12`) {
		t.Errorf("values were changed:\n%s", b)
	}
}

// DR-0015 6: no block, or an empty one, writes no file and never removes one,
// since it may be hand-written.
func TestAnAbsentOrEmptyBonbonBlockLeavesBonbonJSONAlone(t *testing.T) {
	for label, yaml := range map[string]string{"absent": "", "empty": "bonbon: {}\n"} {
		htdocs := subsiteSite(t, yaml)
		path := filepath.Join(htdocs, ".well-known", "bonbon.json")
		if err := runIn(t, "generate"); err != nil {
			t.Fatalf("%s: generate: %s", label, err)
		}
		if _, err := os.Stat(path); err == nil {
			t.Errorf("%s: bonbon.json written", label)
		}
		os.MkdirAll(filepath.Dir(path), 0o775)
		hand := []byte("{\"urlSearch\": \"https://hand.example/\"}\n")
		os.WriteFile(path, hand, 0o664)
		if err := runIn(t, "generate"); err != nil {
			t.Fatalf("%s: generate again: %s", label, err)
		}
		if got, _ := os.ReadFile(path); !bytes.Equal(got, hand) {
			t.Errorf("%s: the hand-written file was changed: %q", label, got)
		}
	}
}

// DR-0015 5 and 9: a key taken out of the block is gone from the rewritten
// file, and an unchanged site leaves both files untouched.
func TestBonbonJSONAndTOCAreStable(t *testing.T) {
	htdocs := subsiteSite(t, twoSubsites+"bonbon:\n  urlSearch: https://example.org/search\n  urlMirrorOf: http://example.org/\n")
	if err := runIn(t, "generate"); err != nil {
		t.Fatal(err)
	}
	files := []string{filepath.Join(htdocs, ".well-known", "bonbon.json"), filepath.Join(htdocs, ".well-known", "toc.opml")}
	var times []int64
	for _, f := range files {
		os.Chtimes(f, mustTime("2020-01-01T00:00:00Z"), mustTime("2020-01-01T00:00:00Z"))
		info, _ := os.Stat(f)
		times = append(times, info.ModTime().UnixNano())
	}
	if err := runIn(t, "generate"); err != nil {
		t.Fatal(err)
	}
	for i, f := range files {
		if info, _ := os.Stat(f); info.ModTime().UnixNano() != times[i] {
			t.Errorf("%s was rewritten for an unchanged site", filepath.Base(f))
		}
	}
	// now drop a key
	src, _ := os.ReadFile("antenna.yaml")
	os.WriteFile("antenna.yaml", []byte(strings.Replace(string(src), "  urlMirrorOf: http://example.org/\n", "", 1)), 0o664)
	if err := runIn(t, "generate"); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(files[0]); strings.Contains(string(b), "urlMirrorOf") || !strings.Contains(string(b), "urlSearch") {
		t.Errorf("bonbon.json after dropping a key:\n%s", b)
	}
}

// DR-0015 9: like the table of contents, bonbon.json is for the whole site.
func TestGenerateOfOneCollectionWritesNeitherFile(t *testing.T) {
	htdocs := subsiteSite(t, twoSubsites+"bonbon:\n  urlSearch: https://example.org/search\n")
	if err := runIn(t, "generate", "pages.md"); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"toc.opml", "bonbon.json"} {
		if _, err := os.Stat(filepath.Join(htdocs, ".well-known", f)); err == nil {
			t.Errorf("%s written by generate pages.md", f)
		}
	}
}
