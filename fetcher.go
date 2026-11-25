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
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	// defaultHarvestPause is the wait between two requests to one host
	// (DR-0030 8). A guess to check on a real site.
	defaultHarvestPause = 500 * time.Millisecond

	// maxRedirects is how many redirects one request follows.
	maxRedirects = 5
)

/** fetchResult is one HTTP response read in full.
 *
 * Fields:
 *   status      (int)    — the HTTP status code
 *   contentType (string) — the Content-Type header
 *   body        (bytes)  — the response body
 *   finalURL    (string) — the address after redirects
 */
type fetchResult struct {
	status      int
	contentType string
	body        []byte
	finalURL    string
}

/** fetcher makes the HTTP requests of one harvest run. It sends the User-Agent,
 * pauses between requests to a host and follows redirects. It does not read
 * robots.txt: antenna fetches feeds sites publish for subscription (DR-0035).
 *
 * Example:
 *   f := newFetcher("antenna/0.1.1", defaultHarvestPause)
 *   res, err := f.get("https://example.org/feed", "application/rss+xml")
 */
type fetcher struct {
	client    *http.Client
	userAgent string
	pause     time.Duration
	now       func() time.Time
	sleep     func(time.Duration)

	last map[string]time.Time

	// policy is what the run records and warns about (DR-0038).
	policy harvestPolicy
}

/** newFetcher returns a fetcher for one run.
 *
 * Parameters:
 *   userAgent (string)        — the User-Agent to send; empty means antenna/VERSION HASH
 *   pause     (time.Duration) — wait between requests to one host; 0 means none
 *
 * Returns:
 *   *fetcher
 *
 * Example:
 *   f := newFetcher("", defaultHarvestPause)
 */
func newFetcher(userAgent string, pause time.Duration) *fetcher {
	if userAgent == "" {
		userAgent = fmt.Sprintf("antenna/%s %s", Version, ReleaseHash)
	}
	f := &fetcher{
		userAgent: userAgent,
		pause:     pause,
		now:       time.Now,
		sleep:     time.Sleep,
		last:      map[string]time.Time{},
		policy:    defaultHarvestPolicy,
	}
	f.client = &http.Client{
		Timeout: 60 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return unavailablef("stopped after %d redirects: %s", maxRedirects, req.URL)
			}
			f.wait(req.URL.Host)
			return nil
		},
	}
	return f
}

// wait sleeps until the pause has passed since the last request to host.
func (f *fetcher) wait(host string) {
	delay := f.pause
	if last, ok := f.last[host]; ok && delay > 0 {
		if rem := delay - f.now().Sub(last); rem > 0 {
			f.sleep(rem)
		}
	}
}

/** get requests href, waiting out the pause first, and returns the response
 * whatever its status (the caller decides what a 404 means). A bad address is a
 * data error and a request that cannot be made is unavailable. DR-0035: no
 * robots.txt is read.
 *
 * Parameters:
 *   href   (string) — the address to request
 *   accept (string) — the Accept header, empty for none
 *
 * Returns:
 *   *fetchResult — the response
 *   error        — a classed failure
 *
 * Example:
 *   res, err := f.get("https://example.org/feed/", "application/rss+xml")
 */
func (f *fetcher) get(href, accept string) (*fetchResult, error) {
	u, err := url.Parse(href)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, dataErrorf("bad URL %q", href)
	}
	req, err := http.NewRequest("GET", href, nil)
	if err != nil {
		return nil, dataErrorf("bad URL %q: %w", href, err)
	}
	req.Header.Set("User-Agent", f.userAgent)
	if accept != "" {
		req.Header.Set("Accept", accept)
	}
	f.wait(u.Host)
	res, err := f.client.Do(req)
	f.last[u.Host] = f.now()
	if err != nil {
		return nil, unavailablef("%s: %w", href, err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	f.last[u.Host] = f.now()
	if err != nil {
		return nil, unavailablef("reading %s: %w", href, err)
	}
	return &fetchResult{
		status:      res.StatusCode,
		contentType: res.Header.Get("Content-Type"),
		body:        body,
		finalURL:    res.Request.URL.String(),
	}, nil
}
