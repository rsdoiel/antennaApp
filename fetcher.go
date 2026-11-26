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
	"errors"
	"fmt"
	"time"

	"github.com/rsdoiel/markdownweb/fetch"
)

const (
	// defaultHarvestPause is the wait between two requests to one host
	// (DR-0030 8). A guess to check on a real site.
	defaultHarvestPause = 500 * time.Millisecond

	// maxFeedBytes is the largest feed or table of contents a harvest reads.
	maxFeedBytes = 32 << 20
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

/** fetcher makes the HTTP requests of one harvest run. It is markdownweb's
 * fetch.Fetcher (markdownweb DR-0002) with antenna's User-Agent and pause, and
 * the error classes of workspace DR-0003. It does not read robots.txt: antenna
 * fetches feeds sites publish for subscription (DR-0035).
 *
 * Example:
 *   f := newFetcher("antenna/0.1.1", defaultHarvestPause)
 *   res, err := f.get("https://example.org/feed", "application/rss+xml")
 */
type fetcher struct {
	f         *fetch.Fetcher
	userAgent string

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
	return newFetcherClock(userAgent, pause, nil, nil)
}

/** newFetcherClock is newFetcher with its clock given, so tests never sleep.
 *
 * Parameters:
 *   userAgent (string)             — as newFetcher
 *   pause     (time.Duration)      — as newFetcher
 *   now       (func() time.Time)   — the clock; nil means time.Now
 *   sleep     (func(time.Duration)) — the wait; nil means time.Sleep
 *
 * Returns:
 *   *fetcher
 *
 * Example:
 *   f := newFetcherClock("", 0, fakeNow, fakeSleep)
 */
func newFetcherClock(userAgent string, pause time.Duration, now func() time.Time, sleep func(time.Duration)) *fetcher {
	if userAgent == "" {
		userAgent = fmt.Sprintf("antenna/%s %s", Version, ReleaseHash)
	}
	return &fetcher{
		userAgent: userAgent,
		policy:    defaultHarvestPolicy,
		f: fetch.New(fetch.Options{
			UserAgent: userAgent,
			Pause:     pause,
			MaxBody:   maxFeedBytes,
			Now:       now,
			Sleep:     sleep,
		}),
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
	res, err := f.f.Get(context.Background(), href, accept)
	if err != nil {
		return nil, classifyFetch(err)
	}
	return &fetchResult{
		status:      res.Status,
		contentType: res.ContentType,
		body:        res.Body,
		finalURL:    res.FinalURL,
	}, nil
}

// classifyFetch gives a markdownweb fetch error its DR-0003 class.
func classifyFetch(err error) error {
	switch {
	case errors.Is(err, fetch.ErrUnavailable):
		return unavailablef("%w", err)
	case errors.Is(err, fetch.ErrData):
		return dataErrorf("%w", err)
	case errors.Is(err, fetch.ErrForbidden):
		return negativef("%w", err)
	}
	return internalErrorf("%w", err)
}
