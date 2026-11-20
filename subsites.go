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
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

/** Subsite is another Markdown site on the same host, listed in antenna.yaml
 * so the site's table of contents can point at it with an include outline
 * (DR-0015).
 *
 * Fields:
 *   URL   (string) — the address of the sub-site's root, absolute http or https
 *   Title (string) — optional label for the include outline
 *
 * Example:
 *   sub := Subsite{URL: "https://example.org/opml", Title: "opml"}
 */
type Subsite struct {
	URL   string `json:"url" yaml:"url"`
	Title string `json:"title,omitempty" yaml:"title,omitempty"`
}

// root returns the sub-site's root address without a trailing slash.
func (s Subsite) root() string { return strings.TrimRight(strings.TrimSpace(s.URL), "/") }

// tocAddress is where BonBon's fixed table of contents of the sub-site is.
func (s Subsite) tocAddress() string { return s.root() + "/.well-known/toc.opml" }

// label is the include outline's text: the title, else the host and path.
func (s Subsite) label() string {
	if t := strings.TrimSpace(s.Title); t != "" {
		return t
	}
	return tocTitle(s.root())
}

// validateSubsites checks the subsites list: each url is present, absolute,
// http or https with a host, and has no query or fragment, and no two name the
// same site (DR-0015 2).
func (cfg *AppConfig) validateSubsites() error {
	seen := map[string]bool{}
	for i, sub := range cfg.Subsites {
		raw := strings.TrimSpace(sub.URL)
		if raw == "" {
			return fmt.Errorf("subsites entry %d has no url", i+1)
		}
		u, err := url.Parse(raw)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return fmt.Errorf("subsites entry %d: url %q must be an absolute http or https address with a host", i+1, raw)
		}
		if u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
			return fmt.Errorf("subsites entry %d: url %q must have no query or fragment", i+1, raw)
		}
		if key := sub.root(); seen[key] {
			return fmt.Errorf("subsites entry %d: url %q is listed twice", i+1, raw)
		} else {
			seen[key] = true
		}
	}
	return nil
}

// bonbonFileName is where BonBon looks for a site's settings, relative to htdocs.
var bonbonFileName = filepath.Join(".well-known", "bonbon.json")

// bonbonStringKeys are the settings BonBon (mdbrowser 0.6.54) reads; it keeps
// only a string for each.
var bonbonStringKeys = []string{"urlMirrorOf", "urlSearch", "urlUpdateSocket"}

/** GenerateBonbonJSON writes <htdocs>/.well-known/bonbon.json from the bonbon
 * mapping of antenna.yaml (DR-0015 5, 6). A mapping with no key writes
 * nothing and removes nothing, so a hand-written file is left alone. The
 * file is a JSON object with sorted keys, two-space indent and a final newline,
 * values as written, and it is rewritten only when its content changes. A
 * known key whose value is not a non-empty string is warned about once and
 * still written, because BonBon drops other types and keeps an empty string,
 * which makes a search address of "?q=word".
 *
 * Parameters:
 *   eout (io.Writer) — warnings
 *
 * Returns:
 *   error — a value that cannot be written as JSON (class config), or a write
 *           failure; nil when the file was written or there was nothing to write
 *
 * Example:
 *   err := cfg.GenerateBonbonJSON(os.Stderr)
 */
func (cfg *AppConfig) GenerateBonbonJSON(eout io.Writer) error {
	if len(cfg.Bonbon) == 0 {
		return nil
	}
	for _, key := range bonbonStringKeys {
		v, ok := cfg.Bonbon[key]
		if !ok {
			continue
		}
		if s, isString := v.(string); !isString || strings.TrimSpace(s) == "" {
			fmt.Fprintf(eout, "warning: bonbon.json: %s should be a non-empty string; BonBon %s\n", key,
				map[bool]string{true: "keeps an empty one, which breaks it", false: "ignores any other type"}[isString])
		}
	}
	// encoding/json sorts the keys of a map.
	src, err := json.MarshalIndent(cfg.Bonbon, "", "  ")
	if err != nil {
		return configErrorf("bonbon in antenna.yaml cannot be written as JSON: %w", err)
	}
	htdocs := cfg.Htdocs
	if htdocs == "" {
		htdocs = "."
	}
	path := filepath.Join(htdocs, bonbonFileName)
	if err := os.MkdirAll(filepath.Dir(path), 0o775); err != nil {
		return asCreate(err)
	}
	if _, err := writeStable(path, append(src, '\n'), 0o666); err != nil {
		return err
	}
	return nil
}
