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
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	// 3rd Party packages
	"github.com/temoto/robotstxt"
)

const (
	// defaultHarvestPause is the wait between two requests to one host
	// (DR-0030 8). A guess to check on a real site.
	defaultHarvestPause = 500 * time.Millisecond

	// maxCrawlDelay is the longest Crawl-delay antenna waits out; a host that
	// asks for more is skipped for the run (DR-0030 8).
	maxCrawlDelay = 30 * time.Second

	// robotsRetries is how many more times a robots.txt that answers 5xx or
	// cannot be reached is asked for before the host is skipped.
	robotsRetries = 2

	// maxRedirects is how many redirects one request follows.
	maxRedirects = 5
)

var (
	// errRobotsDisallowed marks a URL the host's robots.txt does not allow
	// antenna to request. It is a skip, not a failure (DR-0030 5).
	errRobotsDisallowed = errors.New("disallowed by robots.txt")

	// errHostSkipped marks a host antenna does not request in this run
	// because its Crawl-delay is longer than antenna waits out (DR-0030 8).
	// The host's choice: a skip, not a failure.
	errHostSkipped = errors.New("host skipped")

	// errRobotsUnreadable marks a host whose robots.txt could not be read (a
	// server error, a timeout, nothing listening). The host is not requested,
	// as RFC 9309 prescribes, and the link fails as unavailable (69): a dead
	// host is a failure, never a quiet exit 0 (workspace DR-0003).
	errRobotsUnreadable = errors.New("robots.txt unreadable")
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

// robotsRules is what one host's robots.txt says, read once per run.
type robotsRules struct {
	data  *robotstxt.RobotsData
	skip  string        // non-empty: the host is not requested, and why
	fail  bool          // the skip is a failure (robots.txt unreadable)
	delay time.Duration // Crawl-delay for antenna's group, 0 when none
}

/** fetcher makes the HTTP requests of one harvest run. It sends the User-Agent,
 * reads each host's robots.txt before anything else and honors it, pauses
 * between requests to a host, follows redirects (checking each target against
 * robots.txt), and reports skipped hosts and URLs once. DR-0030 5, 8.
 *
 * Example:
 *   f := newFetcher("antenna/0.1.1", defaultHarvestPause, os.Stderr)
 *   res, err := f.get("https://example.org/feed", "application/rss+xml")
 */
type fetcher struct {
	client    *http.Client
	userAgent string
	pause     time.Duration
	eout      io.Writer
	now       func() time.Time
	sleep     func(time.Duration)

	robots   map[string]*robotsRules
	last     map[string]time.Time
	reported map[string]bool

	// skipped counts URLs and hosts left alone because of robots.txt.
	skipped int
}

/** newFetcher returns a fetcher for one run.
 *
 * Parameters:
 *   userAgent (string)        — the User-Agent to send; empty means antenna/VERSION HASH
 *   pause     (time.Duration) — wait between requests to one host; 0 means none
 *   eout      (io.Writer)     — where warnings go
 *
 * Returns:
 *   *fetcher
 *
 * Example:
 *   f := newFetcher("", defaultHarvestPause, os.Stderr)
 */
func newFetcher(userAgent string, pause time.Duration, eout io.Writer) *fetcher {
	if userAgent == "" {
		userAgent = fmt.Sprintf("antenna/%s %s", Version, ReleaseHash)
	}
	f := &fetcher{
		userAgent: userAgent,
		pause:     pause,
		eout:      eout,
		now:       time.Now,
		sleep:     time.Sleep,
		robots:    map[string]*robotsRules{},
		last:      map[string]time.Time{},
		reported:  map[string]bool{},
	}
	f.client = &http.Client{
		Timeout: 60 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return unavailablef("stopped after %d redirects: %s", maxRedirects, req.URL)
			}
			// A redirect may leave the host or enter a path robots.txt
			// forbids: the target is checked like any other request.
			if err := f.permit(req.URL); err != nil {
				return err
			}
			f.wait(req.URL.Host)
			return nil
		},
	}
	return f
}

// warnOnce writes one line to eout the first time key is seen.
func (f *fetcher) warnOnce(key, format string, a ...any) {
	if f.reported[key] {
		return
	}
	f.reported[key] = true
	fmt.Fprintf(f.eout, format+"\n", a...)
}

// agentTokens returns the names robots.txt groups are matched against: the
// configured User-Agent, and "antenna" when the configured one does not
// already start with it (DR-0030 5).
func (f *fetcher) agentTokens() []string {
	tokens := []string{f.userAgent}
	if !strings.HasPrefix(strings.ToLower(f.userAgent), "antenna") {
		tokens = append(tokens, "antenna")
	}
	return tokens
}

// rulesFor reads (once) the robots.txt of the host of u.
func (f *fetcher) rulesFor(u *url.URL) *robotsRules {
	key := u.Scheme + "://" + u.Host
	if r, ok := f.robots[key]; ok {
		return r
	}
	r := &robotsRules{}
	f.robots[key] = r
	// A 5xx may be a passing fault: ask again, after a wait, up to
	// robotsRetries more times before the host is given up for this run.
	res, err := f.fetch(key+"/robots.txt", "", false)
	for attempt := 1; attempt <= robotsRetries && (err != nil || res.status >= 500); attempt++ {
		f.sleep(time.Duration(attempt) * time.Second)
		res, err = f.fetch(key+"/robots.txt", "", false)
	}
	if err != nil {
		r.skip, r.fail = fmt.Sprintf("robots.txt could not be read (%s)", err), true
		return r
	}
	data, err := robotstxt.FromStatusAndBytes(res.status, res.body)
	if err != nil {
		// Real robots.txt files are often a little wrong. Read what can be
		// read; a file nothing can be read from allows the site, with a
		// warning (antenna is a lenient crawler, DR-0030 5).
		data = parseLeniently(res.body)
		if data == nil {
			data, _ = robotstxt.FromStatusAndBytes(404, nil) // allow all
			f.warnOnce("robots "+u.Host, "warning: %s: robots.txt cannot be parsed, treating the site as allowed", u.Host)
		}
	}
	if res.status >= 500 {
		r.skip, r.fail = fmt.Sprintf("robots.txt answered %d", res.status), true
		return r
	}
	r.data = data
	for _, tok := range f.agentTokens() {
		if d := data.FindGroup(tok).CrawlDelay; d > r.delay {
			r.delay = d
		}
	}
	if r.delay > maxCrawlDelay {
		r.skip = fmt.Sprintf("its Crawl-delay is %s, more than the %s antenna waits", r.delay, maxCrawlDelay)
	}
	return r
}

// parseLeniently reads a robots.txt the parser rejected. The usual fault is a
// directive such as Crawl-delay or Sitemap before the first User-agent line;
// those lines are dropped and the rest is read. Nil when nothing is usable.
func parseLeniently(body []byte) *robotstxt.RobotsData {
	lines := strings.Split(string(body), "\n")
	for i, line := range lines {
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), "user-agent") {
			data, err := robotstxt.FromString(strings.Join(lines[i:], "\n"))
			if err != nil {
				return nil
			}
			return data
		}
	}
	return nil
}

// permit says whether u may be requested: nil, errHostSkipped or
// errRobotsDisallowed. A host or URL that is refused is reported once.
func (f *fetcher) permit(u *url.URL) error {
	r := f.rulesFor(u)
	if r.skip != "" {
		if r.fail {
			// Reported by the caller as the link's failure, once per link.
			return unavailablef("%s: skipping the host, %s: %w", u.Host, r.skip, errRobotsUnreadable)
		}
		f.warnOnce("host "+u.Host, "warning: skipping %s: %s", u.Host, r.skip)
		f.skipped++
		return fmt.Errorf("%s: %s: %w", u.Host, r.skip, errHostSkipped)
	}
	if r.data == nil {
		// This host's robots.txt is being read right now (it redirected):
		// nothing to test against yet.
		return nil
	}
	path := u.EscapedPath()
	if path == "" {
		path = "/"
	}
	if u.RawQuery != "" {
		path += "?" + u.RawQuery
	}
	for _, tok := range f.agentTokens() {
		if !r.data.TestAgent(path, tok) {
			f.warnOnce("url "+u.String(), "warning: %s not requested: robots.txt disallows it", u)
			f.skipped++
			return fmt.Errorf("%s: %w", u, errRobotsDisallowed)
		}
	}
	return nil
}

// wait sleeps until the pause (or the host's Crawl-delay, when longer) has
// passed since the last request to host.
func (f *fetcher) wait(host string) {
	delay := f.pause
	for key, r := range f.robots {
		if strings.HasSuffix(key, "://"+host) && r.delay > delay {
			delay = r.delay
		}
	}
	if last, ok := f.last[host]; ok && delay > 0 {
		if rem := delay - f.now().Sub(last); rem > 0 {
			f.sleep(rem)
		}
	}
}

// fetch makes one request. With check true the host's robots.txt is
// consulted first; it is false for robots.txt itself.
func (f *fetcher) fetch(href, accept string, check bool) (*fetchResult, error) {
	u, err := url.Parse(href)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, dataErrorf("bad URL %q", href)
	}
	if check {
		if err := f.permit(u); err != nil {
			return nil, err
		}
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
		if errors.Is(err, errRobotsDisallowed) || errors.Is(err, errHostSkipped) || errors.Is(err, errRobotsUnreadable) {
			return nil, err
		}
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

/** get requests href, honoring robots.txt and the pause, and returns the
 * response whatever its status (the caller decides what a 404 means). A URL
 * robots.txt forbids returns errRobotsDisallowed and a host with a Crawl-delay
 * past the limit returns errHostSkipped (both quiet skips); a host whose
 * robots.txt cannot be read returns errRobotsUnreadable, a failure. None
 * makes a request.
 *
 * Parameters:
 *   href   (string) — the address to request
 *   accept (string) — the Accept header, empty for none
 *
 * Returns:
 *   *fetchResult — the response
 *   error        — errRobotsDisallowed, errHostSkipped, or a classed failure
 *
 * Example:
 *   res, err := f.get("https://example.org/feed/", "application/rss+xml")
 */
func (f *fetcher) get(href, accept string) (*fetchResult, error) {
	return f.fetch(href, accept, true)
}
