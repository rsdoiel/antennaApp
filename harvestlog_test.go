package antennaApp

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// DR-0038: harvest records each feed's HTTP status in harvest_log and warns,
// at the end of the run, about feeds that keep failing. These tests run the
// real harvest verb against a test server, move the clock between runs and
// read the real SQLite file.

const goodFeed = `<?xml version="1.0"?><rss version="2.0"><channel><title>Good</title><link>https://example.org/</link>
<item><title>One</title><link>https://example.org/one</link><guid>one</guid><description>First</description></item></channel></rss>`

// statusServer answers each path with a settable status and body.
type statusServer struct {
	*httptest.Server
	mu     sync.Mutex
	status map[string]int
}

func newStatusServer(t *testing.T) *statusServer {
	t.Helper()
	s := &statusServer{status: map[string]int{}}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		code, ok := s.status[r.URL.Path]
		s.mu.Unlock()
		if !ok {
			http.NotFound(w, r)
			return
		}
		switch r.URL.Path {
		case "/notfeed":
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, "<html><body>no feed here</body></html>")
			return
		}
		if code != http.StatusOK {
			w.WriteHeader(code)
			return
		}
		w.Header().Set("Content-Type", "application/rss+xml")
		fmt.Fprint(w, goodFeed)
	}))
	t.Cleanup(s.Close)
	return s
}

func (s *statusServer) set(path string, code int) {
	s.mu.Lock()
	s.status[path] = code
	s.mu.Unlock()
}

// logSite makes a workspace whose news.md lists the links, with the pause
// off and extra lines added to the harvest: mapping of antenna.yaml.
func logSite(t *testing.T, harvestYAML string, links ...string) {
	t.Helper()
	newWorkspace(t, true)
	cfg, err := os.ReadFile("antenna.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("antenna.yaml", append(cfg, []byte("\nharvest:\n  pause: 0s\n"+harvestYAML)...), 0o644); err != nil {
		t.Fatal(err)
	}
	md := "# News\n\n"
	for _, l := range links {
		md += "- " + l + "\n"
	}
	if err := os.WriteFile("news.md", []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runIn(t, "add", "news.md"); err != nil {
		t.Fatalf("add: %s", err)
	}
}

// day sets the clock to the given day of November 2026, noon UTC.
func day(t *testing.T, d int) {
	t.Helper()
	old := clock
	clock = func() time.Time { return time.Date(2026, 11, d, 12, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { clock = old })
}

type logRow struct {
	link, label, at string
	status, ok      int
	errText         string
}

func logRows(t *testing.T) []logRow {
	t.Helper()
	db, err := sql.Open("sqlite", "news.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows, err := db.Query(`SELECT link, label, at, status, ok, error FROM harvest_log ORDER BY at, link`)
	if err != nil {
		t.Fatalf("reading harvest_log: %s", err)
	}
	defer rows.Close()
	var out []logRow
	for rows.Next() {
		var r logRow
		if err := rows.Scan(&r.link, &r.label, &r.at, &r.status, &r.ok, &r.errText); err != nil {
			t.Fatal(err)
		}
		out = append(out, r)
	}
	return out
}

func harvestOn(t *testing.T, d int) (stdout, stderr string, err error) {
	t.Helper()
	day(t, d)
	return runCapture(t, "harvest", "news.md")
}

func TestHarvestLogRecordsEachAttemptWithItsStatus(t *testing.T) {
	srv := newStatusServer(t)
	srv.set("/good", 200)
	srv.set("/gone", 404)
	srv.set("/broken", 500)
	srv.set("/busy", 503)
	srv.set("/limited", 429)
	srv.set("/notfeed", 200)
	dead := httptest.NewServer(http.NotFoundHandler())
	deadURL := dead.URL
	dead.Close()
	logSite(t, "",
		"[Good]("+srv.URL+"/good)", "[Gone]("+srv.URL+"/gone)", "[Broken]("+srv.URL+"/broken)",
		"[Busy]("+srv.URL+"/busy)", "[Limited]("+srv.URL+"/limited)", "[NotFeed]("+srv.URL+"/notfeed)",
		"[Dead]("+deadURL+"/feed)")
	harvestOn(t, 18)
	want := map[string]struct{ status, ok int }{
		"Good": {200, 1}, "Gone": {404, 0}, "Broken": {500, 0}, "Busy": {503, 0},
		"Limited": {429, 0}, "NotFeed": {200, 0}, "Dead": {0, 0},
	}
	rows := logRows(t)
	if len(rows) != len(want) {
		t.Fatalf("%d rows, want %d: %+v", len(rows), len(want), rows)
	}
	for _, r := range rows {
		w, ok := want[r.label]
		if !ok {
			t.Errorf("unexpected row %+v", r)
			continue
		}
		if r.status != w.status || r.ok != w.ok {
			t.Errorf("%s: status %d ok %d, want %d %d", r.label, r.status, r.ok, w.status, w.ok)
		}
		if r.ok == 0 && r.errText == "" {
			t.Errorf("%s: a failure with no error text", r.label)
		}
		if r.ok == 1 && r.errText != "" {
			t.Errorf("%s: error text %q on a success", r.label, r.errText)
		}
		if r.at != "2026-11-18 12:00:00" {
			t.Errorf("%s: at %q, want the clock's time in UTC", r.label, r.at)
		}
	}
}

func TestHarvestWarnsOnTheThirdFailureNotTheSecond(t *testing.T) {
	srv := newStatusServer(t)
	srv.set("/gone", 404)
	logSite(t, "", "[Gone]("+srv.URL+"/gone)")
	for d := 18; d <= 19; d++ {
		out, eout, _ := harvestOn(t, d)
		if strings.Contains(eout, "has failed") || strings.Contains(out, "has failed") {
			t.Errorf("day %d: warned before the third failure:\n%s", d, eout)
		}
	}
	out, eout, err := harvestOn(t, 20)
	if !strings.Contains(eout, "warning news.md: Gone "+srv.URL+"/gone has failed 3 times since 2026-11-18, last status 404 Not Found") {
		t.Errorf("the third failure gave no warning:\n%s", eout)
	}
	if strings.Contains(out, "has failed") {
		t.Errorf("the warning is on standard output:\n%s", out)
	}
	if err == nil {
		t.Errorf("the failed feed no longer fails the run")
	} else {
		wantClass(t, "a feed that answered 404", err, classUnavailable)
	}
}

func TestHarvestWarningNamesTheStatusInWords(t *testing.T) {
	srv := newStatusServer(t)
	srv.set("/busy", 503)
	srv.set("/limited", 429)
	srv.set("/broken", 500)
	logSite(t, "", "[Busy]("+srv.URL+"/busy)", "[Limited]("+srv.URL+"/limited)", "[Broken]("+srv.URL+"/broken)")
	var eout string
	for d := 18; d <= 20; d++ {
		_, eout, _ = harvestOn(t, d)
	}
	for _, want := range []string{"last status 503 Service Unavailable", "last status 429 Too Many Requests", "last status 500 Internal Server Error"} {
		if !strings.Contains(eout, want) {
			t.Errorf("warning lacks %q:\n%s", want, eout)
		}
	}
}

func TestHarvestWarningForNoAnswerAndForNotAFeed(t *testing.T) {
	srv := newStatusServer(t)
	srv.set("/notfeed", 200)
	dead := httptest.NewServer(http.NotFoundHandler())
	deadURL := dead.URL
	dead.Close()
	logSite(t, "", "[Dead]("+deadURL+"/feed)", "[NotFeed]("+srv.URL+"/notfeed)")
	var eout string
	for d := 18; d <= 20; d++ {
		_, eout, _ = harvestOn(t, d)
	}
	if !strings.Contains(eout, "Dead "+deadURL+"/feed has failed 3 times since 2026-11-18, last status none (no answer)") {
		t.Errorf("no-answer warning wrong:\n%s", eout)
	}
	if !strings.Contains(eout, "NotFeed "+srv.URL+"/notfeed has failed 3 times") || !strings.Contains(eout, "200 OK") {
		t.Errorf("not-a-feed warning wrong:\n%s", eout)
	}
}

func TestHarvestOneGoodAnswerClearsIt(t *testing.T) {
	srv := newStatusServer(t)
	srv.set("/flaky", 503)
	logSite(t, "", "[Flaky]("+srv.URL+"/flaky)")
	harvestOn(t, 18)
	harvestOn(t, 19)
	srv.set("/flaky", 200)
	if _, _, err := harvestOn(t, 20); err != nil {
		t.Fatalf("harvest: %s", err)
	}
	srv.set("/flaky", 503)
	// The good answer on day 20 is inside the window until day 27.
	for d := 21; d <= 23; d++ {
		if _, eout, _ := harvestOn(t, d); strings.Contains(eout, "has failed") {
			t.Errorf("day %d: warned with a good answer inside the window:\n%s", d, eout)
		}
	}
	// Eight days after it, only the failures are in the window; the count is
	// the unbroken run since the good answer.
	_, eout, _ := harvestOn(t, 28)
	if !strings.Contains(eout, "has failed 4 times since 2026-11-21") {
		t.Errorf("want the count to restart after the good answer:\n%s", eout)
	}
}

func TestHarvestAttemptsOutsideTheWindowDoNotCount(t *testing.T) {
	srv := newStatusServer(t)
	srv.set("/gone", 404)
	logSite(t, "", "[Gone]("+srv.URL+"/gone)")
	harvestOn(t, 1)
	harvestOn(t, 2)
	// Eight and nine days on, only these two are inside seven days.
	harvestOn(t, 10)
	_, eout, _ := harvestOn(t, 11)
	if strings.Contains(eout, "has failed") {
		t.Errorf("attempts older than failing_days counted:\n%s", eout)
	}
}

func TestHarvestThresholdsComeFromTheConfiguration(t *testing.T) {
	srv := newStatusServer(t)
	srv.set("/gone", 404)
	logSite(t, "  failing_attempts: 2\n  failing_days: 30\n", "[Gone]("+srv.URL+"/gone)")
	harvestOn(t, 1)
	_, eout, _ := harvestOn(t, 20)
	if !strings.Contains(eout, "has failed 2 times since 2026-11-01") {
		t.Errorf("failing_attempts: 2 and failing_days: 30 not honoured:\n%s", eout)
	}
}

func TestHarvestOldLogRowsAreDeleted(t *testing.T) {
	srv := newStatusServer(t)
	srv.set("/good", 200)
	logSite(t, "  log_days: 10\n", "[Good]("+srv.URL+"/good)")
	harvestOn(t, 1)
	harvestOn(t, 5)
	harvestOn(t, 14)
	rows := logRows(t)
	if len(rows) != 2 || rows[0].at != "2026-11-05 12:00:00" {
		t.Errorf("want the day-1 row deleted and the others kept: %+v", rows)
	}
}

func TestHarvestDoesNotReportAFeedRemovedFromTheList(t *testing.T) {
	srv := newStatusServer(t)
	srv.set("/gone", 404)
	srv.set("/good", 200)
	logSite(t, "", "[Gone]("+srv.URL+"/gone)", "[Good]("+srv.URL+"/good)")
	harvestOn(t, 18)
	harvestOn(t, 19)
	if err := os.WriteFile("news.md", []byte("# News\n\n- [Good]("+srv.URL+"/good)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, eout, err := harvestOn(t, 20)
	if err != nil {
		t.Fatalf("harvest: %s", err)
	}
	if strings.Contains(eout, "has failed") {
		t.Errorf("a feed no longer listed was reported:\n%s", eout)
	}
}

func TestHarvestAddsTheLogToADatabaseFromAnEarlierRelease(t *testing.T) {
	srv := newStatusServer(t)
	srv.set("/good", 200)
	logSite(t, "", "[Good]("+srv.URL+"/good)")
	harvestOn(t, 18)
	db, err := sql.Open("sqlite", "news.db")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`DROP TABLE harvest_log`); err != nil {
		t.Fatalf("drop: %s (the table should exist after a harvest)", err)
	}
	var before int
	db.QueryRow(`SELECT COUNT(*) FROM items`).Scan(&before)
	db.Close()
	harvestOn(t, 19)
	if n := len(logRows(t)); n != 1 {
		t.Errorf("rows after re-adding the table: %d, want 1", n)
	}
	if itemCount(t) != before || before == 0 {
		t.Errorf("items %d, want %d kept", itemCount(t), before)
	}
}

func TestHarvestBadFailingSettingsAreConfigurationErrors(t *testing.T) {
	for _, bad := range []string{"failing_days: 0", "failing_attempts: -1", "log_days: 0"} {
		newWorkspace(t, true)
		cfg, _ := os.ReadFile("antenna.yaml")
		os.WriteFile("antenna.yaml", append(cfg, []byte("\nharvest:\n  "+bad+"\n")...), 0o644)
		_, _, err := runCapture(t, "harvest")
		wantClass(t, "harvest "+bad, err, classConfig)
	}
}

func TestHarvestWithOnlyGoodFeedsStaysQuietAndExitsZero(t *testing.T) {
	srv := newStatusServer(t)
	srv.set("/good", 200)
	logSite(t, "", "[Good]("+srv.URL+"/good)")
	for d := 18; d <= 21; d++ {
		out, eout, err := harvestOn(t, d)
		if err != nil || strings.Contains(eout, "has failed") || strings.Contains(out, "has failed") {
			t.Errorf("day %d: err %v, out %q, eout %q", d, err, out, eout)
		}
	}
}
