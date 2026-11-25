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
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	// 3rd Party packages
	"github.com/mmcdole/gofeed"
	"github.com/rsdoiel/opml"
	"golang.org/x/net/html"
)

// feedAccept is the Accept header sent when asking for a feed.
const feedAccept = "application/rss+xml, application/atom+xml, application/feed+json, application/xml, application/json;q=0.9, */*;q=0.8"

// maxFeedProbes is how many guessed feed addresses are tried for one listed
// address (DR-0030 1).
const maxFeedProbes = 8

// feedProbePaths are the common places a site keeps its feed, in the order
// they are tried: WordPress and Substack (feed), static site generators
// (rss.xml, feed.xml, atom.xml, index.xml), Blogger (feeds/posts/default).
var feedProbePaths = []string{
	"feed", "feed/", "rss", "rss.xml", "feed.xml", "atom.xml", "index.xml", "feeds/posts/default",
}

// feedLinkTypes are the types of a <link rel="alternate"> that name a feed.
var feedLinkTypes = map[string]bool{
	"application/rss+xml":   true,
	"application/atom+xml":  true,
	"application/feed+json": true,
}

/** resolvedFeed is a feed that discovery found and read.
 *
 * Fields:
 *   feed    (*gofeed.Feed) — the parsed feed
 *   address (string)       — the address it was read from; differs from the
 *                            listed one when discovery found it elsewhere
 */
type resolvedFeed struct {
	feed    *gofeed.Feed
	address string
}

// noFeedErr is the data error for a link whose site has no feed that
// discovery can find (DR-0030 1.5).
func noFeedErr(listed, why string) error {
	return dataErrorf("%s: no feed found (%s)", listed, why)
}

// parseFeedBody parses body as an RSS, Atom or JSON feed.
func parseFeedBody(body []byte, href string) (*gofeed.Feed, error) {
	feed, err := gofeed.NewParser().Parse(bytes.NewReader(body))
	if err != nil {
		return nil, dataErrorf("feed error for %q, %w", href, err)
	}
	if feed.Link == "" || feed.Link == "/" {
		u, err := url.Parse(href)
		if err != nil {
			return nil, dataErrorf("bad feed URL %q: %w", href, err)
		}
		u.Path = "/"
		feed.Link = u.String()
	}
	return feed, nil
}

// statusErr is the failure for a response that is not 200.
func statusErr(res *fetchResult) error {
	return &statusError{status: res.status, err: unavailablef("http error: %d %s", res.status, http.StatusText(res.status))}
}

/** statusError carries the HTTP status of the response a failure came from
 * (DR-0038 1), so harvest can log it. It keeps the class of the error it wraps.
 *
 * Fields:
 *   status (int) — the HTTP status of the response
 *   err (error) — the failure
 */
type statusError struct {
	status int
	err    error
}

func (e *statusError) Error() string { return e.err.Error() }
func (e *statusError) Unwrap() error { return e.err }

// withStatus marks err as coming from a response with the given status. An
// error that already carries a status (a feed page that answered 404) keeps
// it, since that is the more specific answer.
func withStatus(status int, err error) error {
	var se *statusError
	if err == nil || errors.As(err, &se) {
		return err
	}
	return &statusError{status: status, err: err}
}

/** statusOf returns the HTTP status a failure came from, or 0 when no response
 * was read (the address did not answer).
 *
 * Parameters:
 *   err (error) — a failure from resolveFeed
 *
 * Returns:
 *   int — the status, or 0
 *
 * Example:
 *   n := statusOf(err) // 404
 */
func statusOf(err error) int {
	var se *statusError
	if errors.As(err, &se) {
		return se.status
	}
	return 0
}

// readFeed requests addr and parses the response as a feed.
func readFeed(f *fetcher, addr string) (*resolvedFeed, error) {
	res, err := f.get(addr, feedAccept)
	if err != nil {
		return nil, err
	}
	if res.status != http.StatusOK {
		return nil, statusErr(res)
	}
	feed, err := parseFeedBody(res.body, addr)
	if err != nil {
		return nil, withStatus(res.status, err)
	}
	return &resolvedFeed{feed: feed, address: addr}, nil
}

// bodyKind says what a response body is, by its content.
type bodyKind int

const (
	kindOther bodyKind = iota
	kindFeed
	kindOPML
	kindHTML
)

// sniff looks at the start of body and the Content-Type to tell an OPML
// document and HTML apart from everything else; a feed is found by trying to
// parse it.
func sniff(res *fetchResult) bodyKind {
	head := res.body
	if len(head) > 4096 {
		head = head[:4096]
	}
	lower := strings.ToLower(string(head))
	switch {
	case strings.Contains(lower, "<opml"):
		return kindOPML
	case strings.Contains(lower, "<!doctype html") || strings.Contains(lower, "<html"),
		strings.Contains(strings.ToLower(res.contentType), "text/html"):
		return kindHTML
	}
	return kindOther
}

// tocResult is what reading an OPML document found.
type tocResult struct {
	feed    string // address of an rss outline, empty when none
	isToc   bool   // outlines name pages or include other documents
	feeds   int    // outlines that name a feed (xmlUrl)
	hasPage bool
}

// readToc parses an OPML document and looks for a feed in it without
// requesting any page it lists (DR-0030 1.2).
func readToc(body []byte, base string) (*tocResult, error) {
	o, err := opml.Parse(body)
	if err != nil {
		return nil, dataErrorf("%s: OPML error: %w", base, err)
	}
	r := &tocResult{}
	var first string
	_ = o.Walk(func(ol *opml.Outline) bool {
		if ol.XMLURL != "" {
			r.feeds++
			if first == "" {
				first = ol.XMLURL
			}
		}
		return true
	})
	for _, p := range o.Flatten() {
		if p.XMLURL == "" {
			r.hasPage = true
			break
		}
	}
	r.isToc = r.hasPage || len(o.Includes()) > 0
	if first != "" {
		b, err := url.Parse(base)
		if err == nil {
			if ref, err := url.Parse(first); err == nil {
				first = b.ResolveReference(ref).String()
			}
		}
		r.feed = first
	}
	return r, nil
}

// autodiscover returns the feed addresses a page names in its <head>
// (<link rel="alternate" type="application/rss+xml" href="...">), in
// document order, resolved against the page's address.
func autodiscover(body []byte, pageURL string) []string {
	base, err := url.Parse(pageURL)
	if err != nil {
		return nil
	}
	var found []string
	z := html.NewTokenizer(bytes.NewReader(body))
	for {
		switch z.Next() {
		case html.ErrorToken:
			return found
		case html.StartTagToken, html.SelfClosingTagToken:
			name, hasAttr := z.TagName()
			tag := string(name)
			if tag == "body" {
				return found
			}
			if tag != "link" || !hasAttr {
				continue
			}
			attrs := map[string]string{}
			for {
				k, v, more := z.TagAttr()
				attrs[strings.ToLower(string(k))] = string(v)
				if !more {
					break
				}
			}
			if !strings.EqualFold(strings.TrimSpace(attrs["rel"]), "alternate") {
				continue
			}
			if !feedLinkTypes[strings.ToLower(strings.TrimSpace(attrs["type"]))] || attrs["href"] == "" {
				continue
			}
			ref, err := url.Parse(strings.TrimSpace(attrs["href"]))
			if err != nil {
				continue
			}
			found = append(found, base.ResolveReference(ref).String())
		}
	}
}

// probeBases returns the addresses guessed feed paths are tried under: the
// folder of the listed address (the part before /.well-known/ for a table of
// contents) and the site's origin.
func probeBases(listed string) []string {
	u, err := url.Parse(listed)
	if err != nil {
		return nil
	}
	origin := u.Scheme + "://" + u.Host + "/"
	dir := u.Path
	if i := strings.Index(dir, "/.well-known/"); i >= 0 {
		dir = dir[:i+1]
	} else if !strings.HasSuffix(dir, "/") {
		dir = dir[:strings.LastIndex(dir, "/")+1]
	}
	if dir == "" {
		dir = "/"
	}
	base := u.Scheme + "://" + u.Host + dir
	if base == origin {
		return []string{origin}
	}
	return []string{base, origin}
}

// probeCandidates lists the guessed addresses for listed, at most
// maxFeedProbes, folder first and origin second for each path.
func probeCandidates(listed string) []string {
	var out []string
	for _, p := range feedProbePaths {
		for _, b := range probeBases(listed) {
			out = append(out, b+p)
		}
	}
	if len(out) > maxFeedProbes {
		out = out[:maxFeedProbes]
	}
	return out
}

// probe tries the guessed feed addresses and returns the first that parses
// as a feed, or nil. A miss or a failed request goes on to the next.
func probe(f *fetcher, listed string) (*resolvedFeed, error) {
	for _, cand := range probeCandidates(listed) {
		res, err := f.get(cand, feedAccept)
		if err != nil {
			continue
		}
		if res.status != http.StatusOK {
			continue
		}
		if feed, err := parseFeedBody(res.body, cand); err == nil {
			return &resolvedFeed{feed: feed, address: cand}, nil
		}
	}
	return nil, nil
}

/** resolveFeed reads the feed behind a listed address (DR-0030 1). The
 * address may be the feed itself, an OPML table of contents with an rss
 * outline, a page that names its feed in a <link> tag, or a site whose feed
 * is at a common path. A remembered address from an earlier run is tried
 * first. No page that a table of contents lists is ever requested.
 *
 * Parameters:
 *   f          (*fetcher) — makes the requests
 *   listed     (string)   — the address the author listed
 *   remembered (string)   — the feed address found on an earlier run, or ""
 *
 * Returns:
 *   *resolvedFeed — the feed and the address it was read from
 *   error         — a data error naming the link when no feed is found, else
 *                   the failure of the request
 *
 * Example:
 *   got, err := resolveFeed(f, "https://example.org/", "")
 *   if err == nil { fmt.Println(got.address, got.feed.Title) }
 */
func resolveFeed(f *fetcher, listed, remembered string) (_ *resolvedFeed, err error) {
	if remembered != "" && remembered != listed {
		got, err := readFeed(f, remembered)
		if err == nil {
			return got, nil
		}
	}
	res, err := f.get(listed, feedAccept)
	if err != nil {
		return nil, err
	}
	if res.status != http.StatusOK {
		return nil, statusErr(res)
	}
	// From here the address answered: a failure is that answer's (DR-0038 1).
	defer func() { err = withStatus(res.status, err) }()
	kind := sniff(res)
	if kind != kindOPML && kind != kindHTML {
		if feed, err := parseFeedBody(res.body, listed); err == nil {
			return &resolvedFeed{feed: feed, address: listed}, nil
		}
	}
	why := "the address is not a feed and names none"
	switch kind {
	case kindOPML:
		toc, err := readToc(res.body, listed)
		if err != nil {
			return nil, err
		}
		if !toc.isToc {
			return nil, dataErrorf("%s: this OPML is a feed list of %d feeds, not a site; list the feeds themselves", listed, toc.feeds)
		}
		if toc.feed != "" {
			return readFeed(f, toc.feed)
		}
		why = "its table of contents lists pages and no feed"
	case kindHTML:
		for _, addr := range autodiscover(res.body, res.finalURL) {
			got, err := readFeed(f, addr)
			if err == nil {
				return got, nil
			}
		}
		why = "the page names no feed"
	default:
		// A Markdown site has no <head>; its table of contents may name a feed.
		for _, base := range probeBases(listed) {
			tocURL := base + ".well-known/toc.opml"
			tres, err := f.get(tocURL, "text/x-opml, text/xml, application/xml, */*;q=0.5")
			if err != nil {
				continue
			}
			if tres.status != http.StatusOK || sniff(tres) != kindOPML {
				continue
			}
			toc, err := readToc(tres.body, tocURL)
			if err != nil || toc.feed == "" {
				if err == nil {
					why = "its table of contents lists pages and no feed"
				}
				continue
			}
			return readFeed(f, toc.feed)
		}
	}
	got, err := probe(f, listed)
	if err != nil {
		return nil, err
	}
	if got != nil {
		return got, nil
	}
	return nil, noFeedErr(listed, fmt.Sprintf("%s; %d common feed paths tried", why, len(probeCandidates(listed))))
}
