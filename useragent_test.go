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
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
)

// userAgentHarvest harvests one feed through the command, as the author does,
// with extra appended to antenna.yaml, and returns the User-Agents the feed
// server saw.
func userAgentHarvest(t *testing.T, extra string) []string {
	t.Helper()
	newWorkspace(t, true)
	var mu sync.Mutex
	var seen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		seen = append(seen, r.UserAgent())
		mu.Unlock()
		w.Header().Set("Content-Type", "application/rss+xml")
		w.Write([]byte(`<?xml version="1.0"?><rss version="2.0"><channel><title>T</title>` +
			`<link>http://example.com/</link><description>d</description></channel></rss>`))
	}))
	t.Cleanup(srv.Close)
	if extra != "" {
		f, err := os.OpenFile("antenna.yaml", os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			t.Fatal(err)
		}
		f.WriteString("\n" + extra + "\n")
		f.Close()
	}
	if err := os.WriteFile("feeds.md", []byte("# Feeds\n\n- [T]("+srv.URL+"/feed)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runIn(t, "add", "feeds.md"); err != nil {
		t.Fatalf("add: %s", err)
	}
	if err := runIn(t, "harvest", "feeds.md"); err != nil {
		t.Fatalf("harvest: %s", err)
	}
	mu.Lock()
	defer mu.Unlock()
	return append([]string(nil), seen...)
}

// antenna.yaml's userAgent is what harvest sends.
func TestHarvest_UserAgentFromAntennaYAML(t *testing.T) {
	seen := userAgentHarvest(t, "userAgent: my-planet/2.0 (+https://example.org/about)")
	if len(seen) == 0 {
		t.Fatal("the feed was never requested")
	}
	for _, ua := range seen {
		if ua != "my-planet/2.0 (+https://example.org/about)" {
			t.Errorf("User-Agent %q, want the configured one", ua)
		}
	}
}

// With no userAgent, harvest sends antenna's own.
func TestHarvest_UserAgentDefaultsToAntenna(t *testing.T) {
	seen := userAgentHarvest(t, "")
	if len(seen) == 0 {
		t.Fatal("the feed was never requested")
	}
	for _, ua := range seen {
		if !strings.HasPrefix(ua, "antenna/"+Version) {
			t.Errorf("User-Agent %q, want antenna/%s ...", ua, Version)
		}
	}
}

// An empty userAgent is the same as none.
func TestHarvest_EmptyUserAgentIsTheDefault(t *testing.T) {
	seen := userAgentHarvest(t, `userAgent: ""`)
	if len(seen) == 0 || !strings.HasPrefix(seen[0], "antenna/") {
		t.Fatalf("User-Agents %v", seen)
	}
}
