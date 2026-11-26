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
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Accept headers: BonBon's, a browser's, and the ones in between.
const (
	acceptBonBon  = "text/markdown, text/x-opml"
	acceptBrowser = "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8"
)

// previewDocRoot builds a small site: the antenna collection case (the .md
// sibling is the feed list, the news is the declared alternate), a page with
// no declaration, a folder, relative and hostile declarations, and the
// dot paths that must stay closed.
func previewDocRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"food.html":                         `<html><head><link href="food_items.md" type="text/markdown" rel="alternate"></head><body>HTML food</body></html>`,
		"food.md":                           "# Feed list\n",
		"food_items.md":                     "# The news\n",
		"about.html":                        `<html><head><title>About</title></head><body>HTML about</body></html>`,
		"about.md":                          "# About (md)\n",
		"plain.html":                        `<html><head></head><body>HTML only</body></html>`,
		"dir/index.html":                    `<html><head></head><body>HTML dir</body></html>`,
		"dir/index.md":                      "# Dir index (md)\n",
		"index.html":                        `<html><head><link rel="alternate" type="text/markdown" href="home.md"></head><body>HTML home</body></html>`,
		"home.md":                           "# Home (md)\n",
		"sub/page.html":                     `<html><head><link rel="alternate" type="text/markdown" href="../other/items.md"></head><body>HTML sub</body></html>`,
		"other/items.md":                    "# Relative alternate\n",
		"evil.html":                         `<html><head><link rel="alternate" type="text/markdown" href=".secret.md"></head><body>HTML evil</body></html>`,
		"evil.md":                           "# Evil sibling\n",
		".secret.md":                        "# DOT SECRET\n",
		"escape.html":                       `<html><head><link rel="alternate" type="text/markdown" href="../../../../etc/passwd.md"></head><body>HTML escape</body></html>`,
		"escape.md":                         "# Escape sibling\n",
		"cross.html":                        `<html><head><link rel="alternate" type="text/markdown" href="http://other.example/x.md"></head><body>HTML cross</body></html>`,
		"cross.md":                          "# Cross sibling\n",
		"missingalt.html":                   `<html><head><link rel="alternate" type="text/markdown" href="gone.md"></head><body>HTML missingalt</body></html>`,
		"missingalt.md":                     "# Missingalt sibling\n",
		"legacy.html":                       `<html><head><link rel="alternate" type="application/markdown" href="wrong.md"></head><body>HTML legacy</body></html>`,
		"legacy.md":                         "# Legacy sibling\n",
		"wrong.md":                          "# Wrong\n",
		".well-known/toc.opml":              `<opml version="2.0"><head/><body/></opml>`,
		".well-known/bonbon.json":           `{"cname":"example.org"}`,
		".well-known/.htaccess":             "deny from all",
		".well-known/.git/config":           "[core] SECRET",
		"scripting/.well-known/bonbon.json": `{"urlMirrorOf":"http://scripting.com/"}`,
		".git/config":                       "[core] SECRET",
		".htaccess":                         "deny from all",
		"page.markdown":                     "# markdown ext\n",
		"list.opml":                         `<opml version="2.0"><head/><body/></opml>`,
	}
	for name, body := range files {
		p := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// previewHandlers are the two ways the preview is built: antenna preview (a
// SafeFileSystem) and the interactive preview (StaticRouter over a plain
// directory). The Markdown web must behave the same in both.
func previewHandlers(root string) map[string]http.Handler {
	mk := func(h http.Handler) http.Handler {
		mux := http.NewServeMux() // as in production: the mux cleans paths and redirects
		mux.Handle("/", h)
		return mux
	}
	return map[string]http.Handler{
		"antenna preview": mk(NewPreviewHandler(SafeFileSystem{http.Dir(root)})),
		"interactive":     mk(StaticRouter(NewPreviewHandler(http.Dir(root)))),
	}
}

// get makes a real HTTP request to the server and returns the response and body.
func get(t *testing.T, srv *httptest.Server, method, path, accept string) (*http.Response, string) {
	t.Helper()
	req, err := http.NewRequest(method, srv.URL+path, nil)
	if err != nil {
		t.Fatal(err)
	}
	if accept != "" {
		req.Header.Set("Accept", accept)
	}
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp, string(b)
}

func eachPreview(t *testing.T, fn func(t *testing.T, srv *httptest.Server)) {
	t.Helper()
	root := previewDocRoot(t)
	for name, h := range previewHandlers(root) {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(h)
			defer srv.Close()
			fn(t, srv)
		})
	}
}

// DR-0010 1: .well-known is served, at any depth, and nothing else with a dot.
func TestWellKnownIsServedAndOtherDotPathsStayClosed(t *testing.T) {
	eachPreview(t, func(t *testing.T, srv *httptest.Server) {
		for path, want := range map[string]int{
			"/.well-known/toc.opml":              200,
			"/.well-known/bonbon.json":           200,
			"/scripting/.well-known/bonbon.json": 200, // a folder's own, as Dave's sites have
			"/.well-known/.htaccess":             403,
			"/.well-known/.git/config":           403,
			"/.git/config":                       403,
			"/.htaccess":                         403,
			"/.secret.md":                        403,
		} {
			resp, body := get(t, srv, "GET", path, acceptBrowser)
			if resp.StatusCode != want {
				t.Errorf("GET %s: %d, want %d", path, resp.StatusCode, want)
			}
			if want == 403 && strings.Contains(body, "SECRET") {
				t.Errorf("GET %s leaked: %q", path, body)
			}
		}
		// ways around the rule
		for _, path := range []string{"/.well-known/../.git/config", "/%2e%67it/config", "/.well-known/%2e%67it/config", "/.WELL-KNOWN/toc.opml/../../.git/config"} {
			resp, body := get(t, srv, "GET", path, acceptBrowser)
			if resp.StatusCode == 200 || strings.Contains(body, "SECRET") {
				t.Errorf("GET %s: %d %q, want it refused", path, resp.StatusCode, body)
			}
		}
	})
}

// DR-0010 2: the type comes from the extension, not from the host's table.
func TestMarkdownAndOPMLTypesAreSetByTheServer(t *testing.T) {
	eachPreview(t, func(t *testing.T, srv *httptest.Server) {
		for path, want := range map[string]string{
			"/food.md":              "text/markdown; charset=utf-8",
			"/page.markdown":        "text/markdown; charset=utf-8",
			"/list.opml":            "text/x-opml; charset=utf-8",
			"/.well-known/toc.opml": "text/x-opml; charset=utf-8",
			"/about.html":           "text/html; charset=utf-8",
		} {
			resp, _ := get(t, srv, "GET", path, "")
			if got := resp.Header.Get("Content-Type"); got != want {
				t.Errorf("GET %s: Content-Type %q, want %q", path, got, want)
			}
		}
	})
}

// DR-0010 3: with BonBon's Accept a page gets the Markdown its head declares,
// not the .md sibling by name, and says where it came from.
func TestBonBonGetsTheDeclaredAlternateNotTheSibling(t *testing.T) {
	eachPreview(t, func(t *testing.T, srv *httptest.Server) {
		resp, body := get(t, srv, "GET", "/food.html", acceptBonBon)
		if resp.StatusCode != 200 || !strings.Contains(body, "The news") {
			t.Fatalf("status %d body %q: want the declared alternate food_items.md, not the feed list", resp.StatusCode, body)
		}
		if got := resp.Header.Get("Content-Type"); got != "text/markdown; charset=utf-8" {
			t.Errorf("Content-Type %q", got)
		}
		if got := resp.Header.Get("Content-Location"); got != "/food_items.md" {
			t.Errorf("Content-Location %q, want /food_items.md", got)
		}
		if !strings.Contains(resp.Header.Get("Vary"), "Accept") {
			t.Errorf("Vary %q, want Accept", resp.Header.Get("Vary"))
		}
	})
}

func TestNegotiationFallsBackThroughTheSiblingToTheHTML(t *testing.T) {
	eachPreview(t, func(t *testing.T, srv *httptest.Server) {
		cases := []struct{ path, want, location string }{
			{"/about.html", "About (md)", "/about.md"},                   // no declaration: the sibling
			{"/missingalt.html", "Missingalt sibling", "/missingalt.md"}, // declared file is missing: the sibling
			{"/legacy.html", "Legacy sibling", "/legacy.md"},             // application/markdown is not a declaration
			{"/cross.html", "Cross sibling", "/cross.md"},                // another host's alternate is not followed
			{"/evil.html", "Evil sibling", "/evil.md"},                   // a dot file is never served, declared or not
			{"/escape.html", "Escape sibling", "/escape.md"},             // a declaration cannot climb out of the site
			{"/sub/page.html", "Relative alternate", "/other/items.md"},  // relative to the page, not the root
			{"/plain.html", "HTML only", ""},                             // nothing Markdown at all: the HTML, as today
		}
		for _, c := range cases {
			resp, body := get(t, srv, "GET", c.path, acceptBonBon)
			if resp.StatusCode != 200 || !strings.Contains(body, c.want) {
				t.Errorf("%s: %d %q, want %q", c.path, resp.StatusCode, body, c.want)
			}
			if got := resp.Header.Get("Content-Location"); got != c.location {
				t.Errorf("%s: Content-Location %q, want %q", c.path, got, c.location)
			}
			if strings.Contains(body, "DOT SECRET") || strings.Contains(body, "root:") {
				t.Errorf("%s leaked a file it must not serve: %q", c.path, body)
			}
		}
	})
}

// A browser, whose Accept puts text/html first, gets the HTML it always did.
func TestBrowsersAndPlainClientsStillGetHTML(t *testing.T) {
	eachPreview(t, func(t *testing.T, srv *httptest.Server) {
		for _, accept := range []string{acceptBrowser, "", "*/*", "text/html", "text/html, text/markdown", "text/*", "text/markdown;q=0.5, text/html"} {
			resp, body := get(t, srv, "GET", "/food.html", accept)
			if !strings.Contains(body, "HTML food") || resp.Header.Get("Content-Location") != "" {
				t.Errorf("Accept %q: got %q (Content-Location %q), want the HTML", accept, body, resp.Header.Get("Content-Location"))
			}
			if !strings.Contains(resp.Header.Get("Vary"), "Accept") {
				t.Errorf("Accept %q: Vary %q, want Accept, so a cache keeps the two apart", accept, resp.Header.Get("Vary"))
			}
		}
	})
}

func TestQualityValuesDecideWhichIsPreferred(t *testing.T) {
	eachPreview(t, func(t *testing.T, srv *httptest.Server) {
		for accept, wantMarkdown := range map[string]bool{
			"text/markdown":                        true, // html absent
			"text/markdown, text/html;q=0.9":       true,
			"text/markdown;q=0.9, text/html;q=0.5": true,
			"text/html, text/markdown;q=0.9":       false,
			"text/markdown;q=0.5, text/html":       false,
			"text/markdown;q=0, text/x-opml":       false, // q=0 means not acceptable
			"text/markdown, */*;q=0.1":             true,
			"*/*, text/markdown;q=0.4":             false, // */* matches html at q=1
		} {
			_, body := get(t, srv, "GET", "/about.html", accept)
			if got := strings.Contains(body, "About (md)"); got != wantMarkdown {
				t.Errorf("Accept %q: markdown %v, want %v (body %q)", accept, got, wantMarkdown, body)
			}
		}
	})
}

// DR-0010 4: a folder is answered the same way, with Content-Location.
func TestFoldersAreNegotiatedAndNameTheFileServed(t *testing.T) {
	eachPreview(t, func(t *testing.T, srv *httptest.Server) {
		resp, body := get(t, srv, "GET", "/dir/", acceptBonBon)
		if !strings.Contains(body, "Dir index (md)") || resp.Header.Get("Content-Location") != "/dir/index.md" {
			t.Errorf("/dir/: %q Content-Location %q", body, resp.Header.Get("Content-Location"))
		}
		// the site root: index.html declares home.md
		resp, body = get(t, srv, "GET", "/", acceptBonBon)
		if !strings.Contains(body, "Home (md)") || resp.Header.Get("Content-Location") != "/home.md" {
			t.Errorf("/: %q Content-Location %q", body, resp.Header.Get("Content-Location"))
		}
		_, body = get(t, srv, "GET", "/dir/", acceptBrowser)
		if !strings.Contains(body, "HTML dir") {
			t.Errorf("a browser's folder request: %q", body)
		}
	})
}

func TestMarkdownAddressesAreServedAsTheyAre(t *testing.T) {
	eachPreview(t, func(t *testing.T, srv *httptest.Server) {
		for _, accept := range []string{acceptBonBon, acceptBrowser, ""} {
			resp, body := get(t, srv, "GET", "/food.md", accept)
			if !strings.Contains(body, "Feed list") || resp.Header.Get("Content-Location") != "" {
				t.Errorf("Accept %q: /food.md gave %q (Content-Location %q)", accept, body, resp.Header.Get("Content-Location"))
			}
		}
	})
}

func TestMissingPagesStayNotFoundAndOnlyGetAndHeadAreNegotiated(t *testing.T) {
	eachPreview(t, func(t *testing.T, srv *httptest.Server) {
		if resp, _ := get(t, srv, "GET", "/missing.html", acceptBonBon); resp.StatusCode != 404 {
			t.Errorf("missing page: %d, want 404", resp.StatusCode)
		}
		resp, body := get(t, srv, "HEAD", "/food.html", acceptBonBon)
		if resp.StatusCode != 200 || body != "" || resp.Header.Get("Content-Location") != "/food_items.md" || resp.Header.Get("Content-Type") != "text/markdown; charset=utf-8" {
			t.Errorf("HEAD: %d body %q location %q type %q", resp.StatusCode, body, resp.Header.Get("Content-Location"), resp.Header.Get("Content-Type"))
		}
		resp, _ = get(t, srv, "POST", "/food.html", acceptBonBon)
		if resp.Header.Get("Content-Location") != "" {
			t.Errorf("POST was negotiated")
		}
	})
}

// The helper functions the handler is built from.
func TestForbiddenDotPathAllowsOnlyWellKnown(t *testing.T) {
	for p, want := range map[string]bool{
		"/":                          false,
		"/a/b.html":                  false,
		"/.well-known/toc.opml":      false,
		"/a/.well-known/bonbon.json": false,
		"/.well-known/x/y.json":      false,
		"/.git/config":               true,
		"/a/.git/config":             true,
		"/.well-known/.htaccess":     true,
		"/.well-known/.git/config":   true,
		"/.htaccess":                 true,
		"/.well-known.bak/x":         true,
		"/.wellknown/x":              true,
	} {
		if got := IsForbiddenDotPath(p); got != want {
			t.Errorf("IsForbiddenDotPath(%q) = %v, want %v", p, got, want)
		}
	}
}

// antenna preview builds its handler in previewMux, in both places that start a
// server. The tests above build handlers themselves; this one goes through the
// function the servers really use, so the Markdown web cannot be left out of
// the preview by a change to how it is wired (it was, once: the types and the
// negotiation worked in tests and not in the running server).
func TestThePreviewMuxServesTheMarkdownWeb(t *testing.T) {
	root := previewDocRoot(t)
	srv := httptest.NewServer(previewMux(SafeFileSystem{http.Dir(root)}))
	defer srv.Close()
	resp, body := get(t, srv, "GET", "/food.html", acceptBonBon)
	if !strings.Contains(body, "The news") || resp.Header.Get("Content-Location") != "/food_items.md" {
		t.Errorf("BonBon's Accept: %q (Content-Location %q)", body, resp.Header.Get("Content-Location"))
	}
	if resp, _ := get(t, srv, "GET", "/list.opml", ""); resp.Header.Get("Content-Type") != "text/x-opml; charset=utf-8" {
		t.Errorf("opml type %q", resp.Header.Get("Content-Type"))
	}
	if resp, _ := get(t, srv, "GET", "/.well-known/toc.opml", ""); resp.StatusCode != 200 {
		t.Errorf(".well-known: %d", resp.StatusCode)
	}
	if resp, _ := get(t, srv, "GET", "/.git/config", ""); resp.StatusCode != 403 {
		t.Errorf(".git: %d", resp.StatusCode)
	}
}

// Both servers, the one WebService.Run starts and antenna preview's, must use it.
func TestBothPreviewServersUsePreviewMux(t *testing.T) {
	src, err := os.ReadFile("webserver.go")
	if err != nil {
		t.Fatal(err)
	}
	s := string(src)
	if n := strings.Count(s, "previewMux(fs)"); n < 2 {
		t.Errorf("webserver.go calls previewMux %d times, want 2 (Run and Preview)", n)
	}
	if strings.Contains(s, `mux.Handle("/", http.FileServer(fs))`) {
		t.Error("a server in webserver.go still serves files without the Markdown web")
	}
}

func TestThePreviewManualDescribesTheMarkdownWeb(t *testing.T) {
	for _, want := range []string{".well-known", "text/markdown; charset=utf-8", "Content-Location", "Vary: Accept", "declared alternate"} {
		if !strings.Contains(PreviewHelpText, want) {
			t.Errorf("antenna help preview lacks %q", want)
		}
	}
}
