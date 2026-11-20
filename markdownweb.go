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
	"io"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// maxHeadRead is how much of an HTML page is read to find the Markdown
// alternate it declares: a head is small, and the page is never loaded whole.
const maxHeadRead = 128 << 10

/** IsForbiddenDotPath reports whether a request path names a dot file or
 * directory that the preview must not serve (antennaApp DR-0010 1). It is
 * IsDotPath with one exception: a path segment that is exactly ".well-known" is
 * allowed, at any depth, because a Markdown-web site keeps its toc.opml and
 * bonbon.json there. Every other segment that starts with a dot, before or
 * after it, is still forbidden (.git, .htaccess, .well-known/.htaccess,
 * .wellknown, .well-known.bak).
 *
 * Parameters:
 *   p (string) — a request path, already decoded
 *
 * Returns:
 *   bool — true when the path must be refused
 *
 * Example:
 *   IsForbiddenDotPath("/.well-known/toc.opml") // false
 *   IsForbiddenDotPath("/.git/config")          // true
 */
func IsForbiddenDotPath(p string) bool {
	for _, part := range strings.Split(path.Clean(p), "/") {
		if part == "" || part == "." || part == ".well-known" {
			continue
		}
		if strings.HasPrefix(part, ".") {
			return true
		}
	}
	return false
}

/** NewPreviewHandler returns the handler the preview server uses over a file
 * system: static files, plus what the Markdown web needs (antennaApp DR-0010).
 *
 * The type of .md, .markdown and .opml files is set from the extension, not left
 * to the host's mime table. When the request's Accept header prefers
 * text/markdown to text/html (BonBon sends "text/markdown, text/x-opml"), a
 * request for an HTML page or a folder is answered with the Markdown the page's
 * head declares (<link rel="alternate" type="text/markdown">), else its .md
 * sibling (x.md, or index.md in a folder), else the HTML as before. The answer
 * names the file served in Content-Location, and every negotiable response
 * carries Vary: Accept. A browser, whose Accept puts text/html first, gets the
 * HTML it always did. Static hosts cannot do this; it is a preview convenience,
 * and the declared alternate is what works everywhere.
 *
 * The file system is expected to keep dot files out (SafeFileSystem does);
 * negotiation also refuses to serve a dot file, whatever a page declares.
 *
 * Parameters:
 *   fsys (http.FileSystem) — the site's files
 *
 * Returns:
 *   http.Handler — the handler to mount at "/"
 *
 * Example:
 *   mux.Handle("/", NewPreviewHandler(SafeFileSystem{http.Dir(htdocs)}))
 */
func NewPreviewHandler(fsys http.FileSystem) http.Handler {
	return negotiateMarkdown(fsys, setMarkdownWebTypes(http.FileServer(fsys)))
}

// setMarkdownWebTypes sets the Content-Type of Markdown and OPML files by
// extension. http.FileServer keeps a type that is already set.
func setMarkdownWebTypes(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch strings.ToLower(path.Ext(r.URL.Path)) {
		case ".md", ".markdown":
			w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		case ".opml":
			w.Header().Set("Content-Type", "text/x-opml; charset=utf-8")
		}
		next.ServeHTTP(w, r)
	})
}

// negotiateMarkdown answers a Markdown-preferring request for an HTML page or a
// folder with Markdown, as NewPreviewHandler describes.
func negotiateMarkdown(fsys http.FileSystem, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			next.ServeHTTP(w, r)
			return
		}
		target, ok := negotiableTarget(fsys, r.URL.Path)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}
		// What is sent for this address depends on Accept.
		w.Header().Add("Vary", "Accept")
		if !prefersMarkdown(r.Header.Get("Accept")) {
			next.ServeHTTP(w, r)
			return
		}
		alt := markdownFor(fsys, target, r.Host)
		if alt == "" {
			next.ServeHTTP(w, r)
			return
		}
		r2 := r.Clone(r.Context())
		r2.URL.Path, r2.URL.RawPath = alt, ""
		w.Header().Set("Content-Location", (&url.URL{Path: alt}).EscapedPath())
		next.ServeHTTP(w, r2)
	})
}

// target is what a request for an HTML page or a folder stands for.
type target struct {
	html    string // the HTML file ("" for a folder with none)
	sibling string // the .md file that sits where the page does
}

// negotiableTarget says whether urlPath names an HTML page that exists or a
// folder, and what its HTML file and .md sibling would be.
func negotiableTarget(fsys http.FileSystem, urlPath string) (target, bool) {
	if IsForbiddenDotPath(urlPath) {
		return target{}, false
	}
	p := path.Clean("/" + urlPath)
	ext := strings.ToLower(path.Ext(p))
	switch {
	case ext == ".html" || ext == ".htm":
		if !isFile(fsys, p) {
			return target{}, false
		}
		return target{html: p, sibling: strings.TrimSuffix(p, path.Ext(p)) + ".md"}, true
	case strings.HasSuffix(urlPath, "/") && isDir(fsys, p):
		t := target{sibling: path.Join(p, "index.md")}
		if idx := path.Join(p, "index.html"); isFile(fsys, idx) {
			t.html = idx
		}
		return t, true
	}
	return target{}, false
}

// markdownFor returns the path of the Markdown to serve for t: the file its
// HTML declares, else its sibling, else "". host is the request's, so an
// alternate on another host is not followed.
func markdownFor(fsys http.FileSystem, t target, host string) string {
	if t.html != "" {
		if alt := declaredMarkdown(fsys, t.html, host); alt != "" {
			return alt
		}
	}
	if isMarkdownFile(fsys, t.sibling) {
		return t.sibling
	}
	return ""
}

// declaredMarkdown reads the head of the HTML file at htmlPath and returns the
// path of the Markdown file it declares with <link rel="alternate"
// type="text/markdown" href>, resolved against the page's own path, or "". The
// file must exist, end in .md or .markdown, and not be a dot file; an address
// on another host, or one that climbs above the site root, is not followed.
func declaredMarkdown(fsys http.FileSystem, htmlPath, host string) string {
	f, err := fsys.Open(htmlPath)
	if err != nil {
		return ""
	}
	defer f.Close()
	head, _ := io.ReadAll(io.LimitReader(f, maxHeadRead))
	href := declaredAlternateHref(head)
	if href == "" {
		return ""
	}
	ref, err := url.Parse(href)
	if err != nil || (ref.Scheme != "" && ref.Scheme != "http" && ref.Scheme != "https") || (ref.Host != "" && ref.Host != host) {
		return ""
	}
	base := &url.URL{Path: htmlPath}
	p := path.Clean("/" + base.ResolveReference(&url.URL{Path: ref.Path}).Path)
	if IsForbiddenDotPath(p) || !isMarkdownFile(fsys, p) {
		return ""
	}
	return p
}

// declaredAlternateHref finds the href of the first <link rel="alternate"
// type="text/markdown"> in the head of an HTML page. rel may hold several words
// and type may carry parameters, both compared without case; an alternate
// stylesheet is not an alternate page, a link in the body is not a declaration,
// and application/markdown is not text/markdown.
func declaredAlternateHref(src []byte) string {
	z := html.NewTokenizer(bytes.NewReader(src))
	for {
		switch z.Next() {
		case html.ErrorToken:
			return ""
		case html.EndTagToken:
			if name, _ := z.TagName(); string(name) == "head" {
				return ""
			}
		case html.StartTagToken, html.SelfClosingTagToken:
			name, hasAttr := z.TagName()
			switch string(name) {
			case "body":
				return ""
			case "link":
				if !hasAttr {
					continue
				}
				var rel, typ, href string
				for {
					key, val, more := z.TagAttr()
					switch string(key) {
					case "rel":
						rel = string(val)
					case "type":
						typ = string(val)
					case "href":
						href = string(val)
					}
					if !more {
						break
					}
				}
				words := strings.Fields(strings.ToLower(rel))
				isAlternate, isStyle := false, false
				for _, w := range words {
					isAlternate = isAlternate || w == "alternate"
					isStyle = isStyle || w == "stylesheet"
				}
				if !isAlternate || isStyle || strings.TrimSpace(href) == "" {
					continue
				}
				if mt, _, err := mime.ParseMediaType(typ); err == nil && mt == "text/markdown" {
					return strings.TrimSpace(href)
				}
			}
		}
	}
}

// isFile reports whether p names a regular file in the file system.
func isFile(fsys http.FileSystem, p string) bool {
	f, err := fsys.Open(p)
	if err != nil {
		return false
	}
	defer f.Close()
	info, err := f.Stat()
	return err == nil && !info.IsDir()
}

// isDir reports whether p names a directory in the file system.
func isDir(fsys http.FileSystem, p string) bool {
	f, err := fsys.Open(p)
	if err != nil {
		return false
	}
	defer f.Close()
	info, err := f.Stat()
	return err == nil && info.IsDir()
}

// isMarkdownFile reports whether p is an existing .md or .markdown file that is
// not a dot file.
func isMarkdownFile(fsys http.FileSystem, p string) bool {
	if IsForbiddenDotPath(p) {
		return false
	}
	ext := strings.ToLower(path.Ext(p))
	return (ext == ".md" || ext == ".markdown") && isFile(fsys, p)
}

/** prefersMarkdown reports whether an Accept header asks for Markdown in
 * preference to HTML: text/markdown is named with a quality above zero, and
 * either no HTML is acceptable (text/html and the two wildcards that would match
 * it, "text/" followed by a star and a lone star type, are all absent or q=0) or
 * Markdown has the higher quality. A tie goes to HTML, and a wildcard never
 * counts as asking for Markdown, so a browser (text/html first, then a wildcard
 * at q=0.8) and a plain wildcard-only Accept get HTML.
 *
 * Parameters:
 *   accept (string) — the Accept request header
 *
 * Returns:
 *   bool — true when Markdown is preferred
 *
 * Example:
 *   prefersMarkdown("text/markdown, text/x-opml") // true (BonBon)
 *   prefersMarkdown("text/html, text/markdown")   // false (a tie)
 */
func prefersMarkdown(accept string) bool {
	var qMarkdown, qHTML float64
	for _, entry := range strings.Split(accept, ",") {
		parts := strings.Split(entry, ";")
		mediaType := strings.ToLower(strings.TrimSpace(parts[0]))
		q := 1.0
		for _, param := range parts[1:] {
			k, v, ok := strings.Cut(strings.TrimSpace(param), "=")
			if ok && strings.EqualFold(strings.TrimSpace(k), "q") {
				if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
					q = f
				}
			}
		}
		switch mediaType {
		case "text/markdown":
			qMarkdown = max(qMarkdown, q)
		case "text/html", "text/*", "*/*":
			qHTML = max(qHTML, q)
		}
	}
	return qMarkdown > 0 && (qHTML <= 0 || qMarkdown > qHTML)
}
