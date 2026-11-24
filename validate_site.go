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
	"context"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// checkSite runs the checks that look at the whole site (DR-0017 5). Each
// finding names the file it is about.
func (v *validator) checkSite() {
	cfg := v.cfg
	cfgFile := filepath.ToSlash(v.cfgName)
	if strings.TrimSpace(cfg.BaseURL) == "" {
		v.add("no-base-url", SeverityWarning, cfgFile, 0,
			"no base_url: the table of contents (.well-known/toc.opml) and the links in the feeds need the address the site is published at")
	}
	htdocs := cfg.Htdocs
	if htdocs == "" {
		htdocs = "."
	}
	if cfg.Htdocs != "" {
		if info, err := os.Stat(cfg.Htdocs); err != nil || !info.IsDir() {
			v.add("htdocs-missing", SeverityError, cfgFile, 0, "htdocs %q is not a directory", cfg.Htdocs)
		}
	}
	v.checkPageRows()
	v.checkPageOutputs(htdocs)
	v.checkSequences(htdocs)
	v.checkStaticMarkdownAlternates()
	v.checkMarkdownOnlyCollections(cfgFile)
	v.checkTOC(htdocs)
	v.checkBonbonJSON(htdocs)
	v.checkNoJekyll(htdocs)
	if v.online {
		v.checkSubsites(cfgFile)
	}
}

// checkPageRows reports a pages-table row whose input file is gone: the stale
// row that made generate exit 66 on 2026-11-20.
func (v *validator) checkPageRows() {
	pages, err := v.cfg.GetPages()
	if err != nil {
		return
	}
	for _, page := range pages {
		in := page["inputPath"]
		if in == "" {
			continue
		}
		if _, err := os.Stat(filepath.FromSlash(in)); err != nil {
			v.add("row-source-missing", SeverityError, filepath.ToSlash(in), 0,
				"the pages table has a row for this file but it does not exist; remove the row with antenna unpage %s", in)
		}
	}
}

// checkTOC reports the Markdown that is not published (what generate warns
// about while it builds the table of contents) and a table of contents that is
// absent or differs from what generate would write.
func (v *validator) checkTOC(htdocs string) {
	var unpublished []Finding
	warnf := func(format string, args ...any) {
		switch {
		case strings.Contains(format, "is not published in") && len(args) >= 2:
			md := toString(args[1])
			unpublished = append(unpublished, Finding{Code: "not-published", Severity: SeverityWarning, File: md,
				Message: "its Markdown is not in " + toString(args[len(args)-1]) + ", so it is left out of the table of contents; publish the .md beside the .html"})
		case strings.Contains(format, "its date") && len(args) >= 2:
			// A post whose date cannot be read is reported with its file when the
			// file exists; here only when there is no file to check.
			if _, err := os.Stat(filepath.FromSlash(toString(args[0]))); err != nil {
				unpublished = append(unpublished, Finding{Code: "date-unreadable", Severity: SeverityError, File: toString(args[0]),
					Message: "its stored date " + toString(args[1]) + " cannot be read"})
			}
		}
	}
	data, state, _, err := v.cfg.buildTOC(warnf)
	if err != nil {
		return
	}
	v.findings = append(v.findings, unpublished...)
	tocFile := filepath.ToSlash(tocFileName)
	existing, readErr := os.ReadFile(filepath.Join(htdocs, tocFileName))
	switch state {
	case tocBuilt:
		switch {
		case readErr != nil:
			v.add("toc-stale", SeverityWarning, tocFile, 0, "the table of contents is missing; run antenna generate")
		case string(stripStamps(existing, tocStamps)) != string(stripStamps(data, tocStamps)):
			v.add("toc-stale", SeverityWarning, tocFile, 0, "the table of contents is out of date; run antenna generate")
		}
	case tocNothing:
		if readErr == nil {
			v.add("toc-stale", SeverityWarning, tocFile, 0, "the site lists nothing now, so generate removes this table of contents; run antenna generate")
		}
	}
}

// toString returns a printable form of a warning argument.
func toString(a any) string {
	if s, ok := a.(string); ok {
		return s
	}
	return strings.TrimSpace(strings.ReplaceAll(strings.Trim(fmt.Sprint(a), "\""), "\n", " "))
}

// checkBonbonJSON reports a bonbon.json that is not what the bonbon setting of
// antenna.yaml would write, is not a JSON object, or holds a known key BonBon
// would drop or break.
func (v *validator) checkBonbonJSON(htdocs string) {
	file := filepath.ToSlash(bonbonFileName)
	existing, readErr := os.ReadFile(filepath.Join(htdocs, bonbonFileName))
	var content map[string]any
	if len(v.cfg.Bonbon) > 0 {
		want, err := v.cfg.bonbonJSONBytes()
		if err != nil {
			return
		}
		switch {
		case readErr != nil:
			v.add("bonbon-json", SeverityWarning, file, 0, "antenna.yaml has a bonbon setting but the file is missing; run antenna generate")
		case string(existing) != string(want):
			v.add("bonbon-json", SeverityWarning, file, 0, "the file differs from the bonbon setting of antenna.yaml; run antenna generate")
		}
		content = v.cfg.Bonbon
	} else if readErr == nil {
		if err := json.Unmarshal(existing, &content); err != nil || content == nil {
			v.add("bonbon-json", SeverityWarning, file, 0, "the file is not a JSON object, so BonBon ignores it")
			return
		}
	}
	for _, key := range bonbonStringKeys {
		if val, ok := content[key]; ok {
			if s, isString := val.(string); !isString || strings.TrimSpace(s) == "" {
				v.add("bonbon-json", SeverityWarning, file, 0,
					"%s should be a non-empty string: BonBon ignores any other type and keeps an empty one, which breaks it", key)
			}
		}
	}
}

// checkNoJekyll reports a GitHub Pages site served from the repository root
// with no .nojekyll file, because GitHub Pages then hides .well-known.
func (v *validator) checkNoJekyll(htdocs string) {
	u, err := url.Parse(strings.TrimSpace(v.cfg.BaseURL))
	if err != nil || !strings.HasSuffix(strings.ToLower(u.Hostname()), ".github.io") {
		return
	}
	if v.cfg.Htdocs != "" && v.cfg.Htdocs != "." {
		return
	}
	if _, err := os.Stat(filepath.Join(htdocs, ".nojekyll")); err != nil {
		v.add("nojekyll", SeverityWarning, ".nojekyll", 0,
			"base_url is a github.io address and the site is the repository root, so GitHub Pages hides .well-known (toc.opml, bonbon.json) without an empty .nojekyll file")
	}
}

// checkSubsites asks each sub-site for its table of contents the way BonBon
// does (an Accept header for OPML) and reports one that does not answer 200
// (DR-0017 5). It runs only with -online: validate makes no request otherwise.
func (v *validator) checkSubsites(cfgFile string) {
	client := &http.Client{Timeout: 10 * time.Second}
	for _, sub := range v.cfg.Subsites {
		address := sub.tocAddress()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
		if err != nil {
			cancel()
			continue
		}
		req.Header.Set("Accept", "text/x-opml")
		req.Header.Set("User-Agent", "antenna validate")
		resp, err := client.Do(req)
		if err != nil {
			v.add("subsite-unreachable", SeverityWarning, cfgFile, 0, "sub-site %s: %s could not be fetched: %s", sub.root(), address, err)
			cancel()
			continue
		}
		resp.Body.Close()
		cancel()
		if resp.StatusCode != http.StatusOK {
			v.add("subsite-unreachable", SeverityWarning, cfgFile, 0,
				"sub-site %s: %s answered %d; BonBon leaves an include it cannot read as an empty node", sub.root(), address, resp.StatusCode)
		}
	}
}

// checkStaticMarkdownAlternates reports a Markdown alternate written in a
// generator YAML's link list (DR-0023): the list is copied into every page, so
// the entry names the same Markdown for all of them, and antenna already writes
// the right one per page.
func (v *validator) checkStaticMarkdownAlternates() {
	seen := map[string]bool{}
	names := []string{v.cfg.Generator}
	for _, col := range v.cfg.Collections {
		names = append(names, col.Generator)
	}
	for _, name := range names {
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		src, err := os.ReadFile(name)
		if err != nil {
			continue
		}
		var gen struct {
			Link []map[string]string `yaml:"link"`
		}
		if err := yaml.Unmarshal(src, &gen); err != nil {
			continue
		}
		for _, l := range gen.Link {
			typ := strings.ToLower(strings.TrimSpace(l["type"]))
			isAlt := false
			for _, w := range strings.Fields(strings.ToLower(l["rel"])) {
				isAlt = isAlt || w == "alternate"
			}
			if !isAlt || (typ != "text/markdown" && typ != "application/markdown") {
				continue
			}
			v.add("alternate-static", SeverityWarning, filepath.ToSlash(name), keyLine(append([]byte("---\n"), src...), "link"),
				"a Markdown alternate (%s, href %q) in the link list is copied into every page, so it names the same Markdown for all of them; remove it, antenna writes the right one for each page", typ, l["href"])
		}
	}
}

// checkPageOutputs runs the checks on what the pages table records (DR-0024 9):
// an output that is not there, two rows that write the same output (as after a
// rename), a file a source includes that does not exist, and an output that is
// older than its source or an include.
func (v *validator) checkPageOutputs(htdocs string) {
	rows, err := v.cfg.GetPageRows()
	if err != nil {
		return
	}
	writers := map[string][]string{}
	var order []string
	for _, r := range rows {
		if r.Input == "" {
			continue
		}
		in := filepath.ToSlash(r.Input)
		for _, out := range r.Outputs {
			if _, err := os.Stat(filepath.Join(htdocs, filepath.FromSlash(out))); err != nil {
				v.add("row-output-missing", SeverityWarning, in, 0,
					"the pages table records the output %s but it is not there; run antenna generate", out)
			}
			if _, seen := writers[out]; !seen {
				order = append(order, out)
			}
			writers[out] = append(writers[out], in)
		}
		for _, inc := range r.Includes {
			if _, err := os.Stat(filepath.FromSlash(inc)); err != nil {
				v.add("include-missing", SeverityWarning, in, 0,
					"includes %s, which does not exist; the directive is left in the page as written", inc)
			}
		}
		if pageStale(r, htdocs) == "stale" {
			v.add("output-stale", SeverityInfo, in, 0,
				"an output is older than this source or a file it includes; run antenna generate")
		}
	}
	for _, out := range order {
		if srcs := writers[out]; len(srcs) > 1 {
			v.add("row-duplicate-output", SeverityWarning, filepath.ToSlash(out), 0,
				"written by more than one row of the pages table (%s); remove the stale one with antenna unpage", strings.Join(srcs, ", "))
		}
	}
}

// checkSequences reports the sequences of antenna.yaml (DR-0026 9): a path
// that is not there, a sequence with no members, and a document that two
// sequences both hold (it takes the first).
func (v *validator) checkSequences(htdocs string) {
	if len(v.cfg.Sequences) == 0 {
		return
	}
	quiet := func(string, ...any) {}
	tally := &failureTally{}
	pages := v.cfg.tocPages(htdocs, quiet, tally)
	posts := v.cfg.tocPosts(htdocs, quiet, tally)
	cfgFile := filepath.ToSlash(v.cfgName)
	for _, info := range v.cfg.sequenceMembers(htdocs, append(append([]tocEntry{}, pages...), posts...)) {
		p := info.Sequence.Path
		switch {
		case !info.Exists:
			what := "a published list file"
			if info.Sequence.isDir() {
				what = "a directory"
			}
			v.add("sequence-path-missing", SeverityWarning, cfgFile, 0, "the sequence path %s is not there; it should be %s under %s", p, what, htdocs)
		case len(info.Members) == 0 && len(info.Overlaps) == 0:
			v.add("sequence-empty", SeverityInfo, cfgFile, 0, "the sequence %s has no members", p)
		}
		for _, o := range info.Overlaps {
			v.add("sequence-overlap", SeverityInfo, cfgFile, 0,
				"%s is in the sequences %s and %s; it takes the neighbours from %s", o.MD, o.First, p, o.First)
		}
	}
}

// checkMarkdownOnlyCollections warns about a Markdown-only collection (html:
// false, DR-0031) that names a generator: the theme is not used because no HTML
// is written for the collection.
func (v *validator) checkMarkdownOnlyCollections(cfgFile string) {
	for _, col := range v.cfg.Collections {
		if col != nil && !col.WantsHTML() && col.Generator != "" {
			v.add("markdown-only-generator", SeverityWarning, cfgFile, 0,
				"the collection %s is Markdown-only (html: false), so its generator %s is not used; remove the generator line", col.File, col.Generator)
		}
	}
}
