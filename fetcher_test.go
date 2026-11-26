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
func testFetcher(pause time.Duration) (*fetcher, *[]time.Duration) {
	clock := time.Date(2026, 11, 23, 12, 0, 0, 0, time.UTC)
	waits := []time.Duration{}
	now := func() time.Time { return clock }
	sleep := func(d time.Duration) { waits = append(waits, d); clock = clock.Add(d) }
	return newFetcherClock("antenna-test", pause, now, sleep), &waits
}

func TestFetcher_PauseBetweenRequestsToOneHost(t *testing.T) {
	srv, _ := robotsServer(200, "")
	defer srv.Close()
	other, _ := robotsServer(200, "")
	defer other.Close()
	f, waits := testFetcher(500 * time.Millisecond)
	for _, u := range []string{srv.URL + "/a", srv.URL + "/b", other.URL + "/a"} {
		if _, err := f.get(u, ""); err != nil {
			t.Fatal(err)
		}
	}
	// srv: /a then /b is one wait; other: its first request needs none.
	for _, w := range *waits {
		if w != 500*time.Millisecond {
			t.Errorf("waited %v, want 500ms", w)
		}
	}
	if len(*waits) != 1 {
		t.Errorf("%d waits %v, want 1 (between the two requests to the first host)", len(*waits), *waits)
	}
}

// webget used to stop at every redirect (its handler always returned
// http.ErrUseLastResponse), so a feed that moved from http to https, or from
// /feed to /feed/, failed with "http error: 301".
func TestFetcher_FollowsRedirects(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/old" {
			http.Redirect(w, r, "/new", http.StatusMovedPermanently)
			return
		}
		w.Write([]byte("ok"))
	}))
	defer srv.Close()
	f, _ := testFetcher(0)
	res, err := f.get(srv.URL+"/old", "")
	if err != nil || string(res.body) != "ok" || res.status != 200 || !strings.HasSuffix(res.finalURL, "/new") {
		t.Fatalf("redirect not followed: %+v %v", res, err)
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
	f, _ := testFetcher(0)
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
	f, _ := testFetcher(0)
	if _, err := f.get(srv.URL+"/feed", "application/rss+xml"); err != nil {
		t.Fatal(err)
	}
	if ua != "antenna-test" || accept != "application/rss+xml" {
		t.Errorf("User-Agent %q, Accept %q", ua, accept)
	}
}

// DR-0035: harvest does not read robots.txt. A site whose robots.txt forbids
// everything, or answers 503, or demands a long Crawl-delay, is still fetched,
// and robots.txt is never requested.
func TestFetcher_IgnoresRobotsTxtEntirely(t *testing.T) {
	cases := map[string]struct {
		status int
		body   string
	}{
		"forbids all":      {200, "User-agent: *\nDisallow: /\n"},
		"forbids antenna":  {200, "User-agent: antenna\nDisallow: /\n"},
		"server error":     {503, ""},
		"forbidden status": {403, ""},
		"long crawl-delay": {200, "User-agent: *\nCrawl-delay: 120\n"},
	}
	for name, c := range cases {
		srv, hits := robotsServer(c.status, c.body)
		f, _ := testFetcher(0)
		res, err := f.get(srv.URL+"/feed.xml", "")
		srv.Close()
		if err != nil || res == nil || res.status != 200 {
			t.Errorf("%s: get = %v, %v; want the feed fetched", name, res, err)
		}
		if n := hits.count("/robots.txt"); n != 0 {
			t.Errorf("%s: robots.txt requested %d time(s), want never", name, n)
		}
	}
}

// markdownweb DR-0002: the fetcher is markdownweb's. A 429 with Retry-After is
// waited out and tried once more, which antenna's own fetcher never did.
func TestFetcher_WaitsOutRetryAfter(t *testing.T) {
	n := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n++
		if n == 1 {
			w.Header().Set("Retry-After", "3")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Write([]byte("ok"))
	}))
	defer srv.Close()
	f, waits := testFetcher(0)
	res, err := f.get(srv.URL+"/feed", "")
	if err != nil || res.status != 200 || n != 2 {
		t.Fatalf("res %+v err %v requests %d", res, err, n)
	}
	if len(*waits) != 1 || (*waits)[0] != 3*time.Second {
		t.Fatalf("waits %v, want [3s]", *waits)
	}
}

// A Retry-After far beyond a harvest's patience is unavailable (69), not a
// long sleep.
func TestFetcher_LongRetryAfterIsUnavailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "86400")
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()
	f, waits := testFetcher(0)
	_, err := f.get(srv.URL+"/feed", "")
	wantClass(t, "long Retry-After", err, classUnavailable)
	if len(*waits) != 0 {
		t.Fatalf("slept %v", *waits)
	}
}

func TestFetcher_BadURLIsDataClass(t *testing.T) {
	f, _ := testFetcher(0)
	for _, u := range []string{"", "ftp://example.org/x", "::bad"} {
		_, err := f.get(u, "")
		wantClass(t, "bad URL "+u, err, classData)
	}
}

func TestFetcher_UnreachableIsUnavailableClass(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	u := srv.URL
	srv.Close()
	f, _ := testFetcher(0)
	_, err := f.get(u, "")
	wantClass(t, "closed port", err, classUnavailable)
}

// The default User-Agent is still antenna's, not markdownweb's.
func TestFetcher_DefaultUserAgentNamesAntenna(t *testing.T) {
	var ua string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { ua = r.UserAgent() }))
	defer srv.Close()
	f := newFetcher("", 0)
	if _, err := f.get(srv.URL, ""); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(ua, "antenna/") || f.userAgent != ua {
		t.Fatalf("User-Agent sent %q, held %q", ua, f.userAgent)
	}
}
