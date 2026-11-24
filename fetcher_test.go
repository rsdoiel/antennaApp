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
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// hitLog records the paths a test server was asked for.
type hitLog struct {
	mu    sync.Mutex
	paths []string
}

func (h *hitLog) add(p string) { h.mu.Lock(); h.paths = append(h.paths, p); h.mu.Unlock() }
func (h *hitLog) count(p string) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	n := 0
	for _, x := range h.paths {
		if x == p {
			n++
		}
	}
	return n
}

// robotsServer answers robots.txt with status and body and every other path
// with 200 "ok", recording what was asked for.
func robotsServer(status int, robots string) (*httptest.Server, *hitLog) {
	hits := &hitLog{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.add(r.URL.Path)
		if r.URL.Path == "/robots.txt" {
			w.WriteHeader(status)
			w.Write([]byte(robots))
			return
		}
		w.Write([]byte("ok"))
	}))
	return srv, hits
}

// testFetcher is a fetcher whose clock is fake: sleep advances it and records
// the waits, so no test sleeps.
func testFetcher(pause time.Duration) (*fetcher, *bytes.Buffer, *[]time.Duration) {
	eout := &bytes.Buffer{}
	f := newFetcher("antenna-test", pause, eout)
	clock := time.Date(2026, 11, 23, 12, 0, 0, 0, time.UTC)
	waits := []time.Duration{}
	f.now = func() time.Time { return clock }
	f.sleep = func(d time.Duration) { waits = append(waits, d); clock = clock.Add(d) }
	return f, eout, &waits
}

func TestFetcher_DisallowedURLIsNeverRequested(t *testing.T) {
	srv, hits := robotsServer(200, "User-agent: *\nDisallow: /private\n")
	defer srv.Close()
	f, eout, _ := testFetcher(0)
	for i := 0; i < 2; i++ {
		_, err := f.get(srv.URL+"/private/feed.xml", "")
		if !errors.Is(err, errRobotsDisallowed) {
			t.Fatalf("pass %d: err = %v, want errRobotsDisallowed", i, err)
		}
	}
	if n := hits.count("/private/feed.xml"); n != 0 {
		t.Errorf("a disallowed URL was requested %d time(s)", n)
	}
	if n := strings.Count(eout.String(), "robots.txt"); n != 1 {
		t.Errorf("reported %d times, want once:\n%s", n, eout.String())
	}
	if _, err := f.get(srv.URL+"/public", ""); err != nil {
		t.Errorf("an allowed URL failed: %v", err)
	}
}

func TestFetcher_RobotsIsFetchedOncePerHost(t *testing.T) {
	srv, hits := robotsServer(200, "User-agent: *\nDisallow:\n")
	defer srv.Close()
	f, _, _ := testFetcher(0)
	for _, p := range []string{"/a", "/b", "/c"} {
		if _, err := f.get(srv.URL+p, ""); err != nil {
			t.Fatal(err)
		}
	}
	if n := hits.count("/robots.txt"); n != 1 {
		t.Errorf("robots.txt fetched %d times, want 1", n)
	}
}

func TestFetcher_MissingRobotsAllowsTheSite(t *testing.T) {
	for _, status := range []int{404, 403, 410} {
		srv, hits := robotsServer(status, "User-agent: *\nDisallow: /\n")
		f, _, _ := testFetcher(0)
		if _, err := f.get(srv.URL+"/feed", ""); err != nil {
			t.Errorf("robots.txt %d: %v, want the site harvested", status, err)
		}
		if hits.count("/feed") != 1 {
			t.Errorf("robots.txt %d: /feed not requested", status)
		}
		srv.Close()
	}
}

// A robots.txt that cannot be read (the server errors, or is not there) means
// "can't tell", and the host is not requested (RFC 9309). That is a failing
// host, not the host's choice, so it is an unavailable failure and not a quiet
// skip: a dead feed must not turn a harvest's exit code into 0.
func TestFetcher_UnreadableRobotsSkipsTheHostAndIsUnavailable(t *testing.T) {
	srv, hits := robotsServer(503, "")
	defer srv.Close()
	f, _, _ := testFetcher(0)
	for i := 0; i < 2; i++ {
		_, err := f.get(srv.URL+"/feed", "")
		if !errors.Is(err, errRobotsUnreadable) {
			t.Fatalf("pass %d: err = %v, want errRobotsUnreadable", i, err)
		}
		if errors.Is(err, errHostSkipped) {
			t.Errorf("pass %d: an unreadable robots.txt is a failure, not a quiet skip", i)
		}
		wantClass(t, "robots.txt 503", err, classUnavailable)
	}
	if hits.count("/feed") != 0 {
		t.Error("a host whose robots.txt answered 503 was still requested")
	}
	// Three tries (the first and two retries), once per run, not once per call.
	if hits.count("/robots.txt") != 3 {
		t.Errorf("robots.txt asked for %d times, want 3 (a try and two retries)", hits.count("/robots.txt"))
	}
}

// Some servers answer robots.txt with a 503 now and then while serving
// everything else (inveniosoftware.org, 2026-11-23). One flaky answer must not
// fail the link: robots.txt is asked for again, after a wait.
func TestFetcher_FlakyRobotsIsRetried(t *testing.T) {
	var robotsHits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			robotsHits++
			if robotsHits == 1 {
				http.Error(w, "busy", http.StatusServiceUnavailable)
				return
			}
			http.NotFound(w, r)
			return
		}
		w.Write([]byte("ok"))
	}))
	defer srv.Close()
	f, _, waits := testFetcher(0)
	if _, err := f.get(srv.URL+"/feed", ""); err != nil {
		t.Fatalf("one 503 from robots.txt failed the link: %v", err)
	}
	if robotsHits != 2 {
		t.Errorf("robots.txt asked for %d times, want 2", robotsHits)
	}
	if len(*waits) == 0 {
		t.Error("no wait before the retry")
	}
}

func TestFetcher_UnreachableRobotsSkipsTheHost(t *testing.T) {
	srv, _ := robotsServer(200, "")
	url := srv.URL
	srv.Close() // nothing is listening now
	f, _, _ := testFetcher(0)
	_, err := f.get(url+"/feed", "")
	if !errors.Is(err, errRobotsUnreadable) {
		t.Fatalf("err = %v, want errRobotsUnreadable", err)
	}
	wantClass(t, "unreachable host", err, classUnavailable)
}

func TestFetcher_PauseBetweenRequestsToOneHost(t *testing.T) {
	srv, _ := robotsServer(404, "")
	defer srv.Close()
	other, _ := robotsServer(404, "")
	defer other.Close()
	f, _, waits := testFetcher(500 * time.Millisecond)
	for _, u := range []string{srv.URL + "/a", srv.URL + "/b", other.URL + "/a"} {
		if _, err := f.get(u, ""); err != nil {
			t.Fatal(err)
		}
	}
	// srv: robots.txt, /a, /b are three requests (two waits); other: its
	// first request needs no wait, its second (/a after robots.txt) does.
	for _, w := range *waits {
		if w != 500*time.Millisecond {
			t.Errorf("waited %v, want 500ms", w)
		}
	}
	if len(*waits) != 3 {
		t.Errorf("%d waits %v, want 3 (two on the first host, one on the second)", len(*waits), *waits)
	}
}

func TestFetcher_CrawlDelayLongerThanThePauseWins(t *testing.T) {
	srv, _ := robotsServer(200, "User-agent: *\nCrawl-delay: 2\n")
	defer srv.Close()
	f, _, waits := testFetcher(500 * time.Millisecond)
	for _, p := range []string{"/a", "/b"} {
		if _, err := f.get(srv.URL+p, ""); err != nil {
			t.Fatal(err)
		}
	}
	for _, w := range *waits {
		if w != 2*time.Second {
			t.Errorf("waited %v, want the 2s Crawl-delay", w)
		}
	}
	if len(*waits) == 0 {
		t.Error("no wait between requests")
	}
}

func TestFetcher_CrawlDelayOverThirtySecondsSkipsTheHost(t *testing.T) {
	srv, hits := robotsServer(200, "User-agent: *\nCrawl-delay: 120\n")
	defer srv.Close()
	f, eout, _ := testFetcher(0)
	if _, err := f.get(srv.URL+"/feed", ""); !errors.Is(err, errHostSkipped) {
		t.Fatalf("err = %v, want errHostSkipped", err)
	}
	if hits.count("/feed") != 0 {
		t.Error("requested a host whose Crawl-delay is past the limit")
	}
	if !strings.Contains(eout.String(), "Crawl-delay") {
		t.Errorf("the warning does not name the Crawl-delay: %s", eout.String())
	}
}

func TestFetcher_UserAgentGroupsAreHonored(t *testing.T) {
	srv, hits := robotsServer(200, "User-agent: antenna-test\nDisallow: /blocked\n\nUser-agent: *\nDisallow:\n")
	defer srv.Close()
	f, _, _ := testFetcher(0)
	if _, err := f.get(srv.URL+"/blocked", ""); !errors.Is(err, errRobotsDisallowed) {
		t.Errorf("err = %v, want errRobotsDisallowed for our own group", err)
	}
	if hits.count("/blocked") != 0 {
		t.Error("requested a URL our own group disallows")
	}
}

// webget used to stop at every redirect (its handler always returned
// http.ErrUseLastResponse), so a feed that moved from http to https, or from
// /feed to /feed/, failed with "http error: 301".
func TestFetcher_FollowsRedirectsAndChecksTheirTargets(t *testing.T) {
	hits := &hitLog{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.add(r.URL.Path)
		switch r.URL.Path {
		case "/robots.txt":
			w.Write([]byte("User-agent: *\nDisallow: /secret\n"))
		case "/old":
			http.Redirect(w, r, "/new", http.StatusMovedPermanently)
		case "/sneaky":
			http.Redirect(w, r, "/secret/feed", http.StatusFound)
		default:
			w.Write([]byte("ok"))
		}
	}))
	defer srv.Close()
	f, _, _ := testFetcher(0)
	res, err := f.get(srv.URL+"/old", "")
	if err != nil || string(res.body) != "ok" || res.status != 200 {
		t.Fatalf("redirect not followed: %+v %v", res, err)
	}
	if _, err := f.get(srv.URL+"/sneaky", ""); !errors.Is(err, errRobotsDisallowed) {
		t.Errorf("err = %v, want the redirect target refused", err)
	}
	if hits.count("/secret/feed") != 0 {
		t.Error("followed a redirect into a disallowed path")
	}
}

func TestFetcher_TooManyRedirectsIsUnavailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, r.URL.Path+"x", http.StatusFound)
	}))
	defer srv.Close()
	f, _, _ := testFetcher(0)
	_, err := f.get(srv.URL+"/loop", "")
	if err == nil {
		t.Fatal("a redirect loop was accepted")
	}
	wantClass(t, "redirect loop", err, classUnavailable)
}

func TestFetcher_SendsUserAgentAndAccept(t *testing.T) {
	var ua, accept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			http.NotFound(w, r)
			return
		}
		ua, accept = r.UserAgent(), r.Header.Get("Accept")
		w.Write([]byte("ok"))
	}))
	defer srv.Close()
	f, _, _ := testFetcher(0)
	if _, err := f.get(srv.URL+"/feed", "application/rss+xml"); err != nil {
		t.Fatal(err)
	}
	if ua != "antenna-test" || accept != "application/rss+xml" {
		t.Errorf("User-Agent %q, Accept %q", ua, accept)
	}
}

// A robots.txt that is itself a redirect (http to https, or /robots.txt to a
// canonical path) is followed; the redirect check must not ask for the file
// that is still being read.
func TestFetcher_RobotsTxtMayRedirect(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/robots.txt":
			http.Redirect(w, r, "/real-robots.txt", http.StatusMovedPermanently)
		case "/real-robots.txt":
			w.Write([]byte("User-agent: *\nDisallow: /no\n"))
		default:
			w.Write([]byte("ok"))
		}
	}))
	defer srv.Close()
	f, _, _ := testFetcher(0)
	if _, err := f.get(srv.URL+"/no/feed", ""); !errors.Is(err, errRobotsDisallowed) {
		t.Fatalf("err = %v, want the rules from the redirected robots.txt applied", err)
	}
}

// Real robots.txt files are often a little wrong. The Scholarly Kitchen's
// begins with "Crawl-delay: 10" before any User-agent line, which the parser
// rejects outright. A crawler that is lenient reads the rest: it must not turn
// a misplaced line into a failed host.
func TestFetcher_MisplacedLineBeforeUserAgentIsIgnored(t *testing.T) {
	srv, hits := robotsServer(200, "Crawl-delay: 10\n# START\nUser-agent: *\nDisallow: /private\n\nSitemap: https://example.org/s.xml\n")
	defer srv.Close()
	f, _, _ := testFetcher(0)
	if _, err := f.get(srv.URL+"/public", ""); err != nil {
		t.Fatalf("a robots.txt with a misplaced first line failed the host: %v", err)
	}
	if _, err := f.get(srv.URL+"/private/x", ""); !errors.Is(err, errRobotsDisallowed) {
		t.Errorf("err = %v, want the rules after the stray line still applied", err)
	}
	if hits.count("/private/x") != 0 {
		t.Error("requested a path the readable rules disallow")
	}
}

func TestFetcher_UnparseableRobotsAllowsTheSiteWithAWarning(t *testing.T) {
	srv, hits := robotsServer(200, "Disallow: /\nAllow: /\nCrawl-delay: x\n")
	defer srv.Close()
	f, eout, _ := testFetcher(0)
	if _, err := f.get(srv.URL+"/feed", ""); err != nil {
		t.Fatalf("a robots.txt nothing can be read from failed the host: %v", err)
	}
	if hits.count("/feed") != 1 {
		t.Error("the feed was not requested")
	}
	if !strings.Contains(eout.String(), "robots.txt") {
		t.Errorf("no warning about the unreadable robots.txt:\n%s", eout.String())
	}
}
