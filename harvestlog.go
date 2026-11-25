package antennaApp

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// DR-0038: harvest records every attempt to fetch a listed feed in the
// collection database's harvest_log table, and at the end of the run warns
// about the feeds that keep failing.

// harvestPolicy is what the log keeps and what counts as failing.
type harvestPolicy struct {
	// failingDays is the window a feed's attempts are judged in.
	failingDays int
	// failingAttempts is how many attempts in the window, all failed, make a
	// feed failing.
	failingAttempts int
	// logDays is how long log rows are kept.
	logDays int
}

// defaultHarvestPolicy is 3 failed runs in 7 days, kept 90 days (DR-0038 3, 4).
var defaultHarvestPolicy = harvestPolicy{failingDays: 7, failingAttempts: 3, logDays: 90}

// logTimeLayout is how the log stores a time: UTC, sortable as text.
const logTimeLayout = "2006-01-02 15:04:05"

/** ensureHarvestLog creates the harvest_log table when the database lacks it,
 * as a database made by an earlier release does. Every row already there is
 * kept.
 *
 * Parameters:
 *   db (*sql.DB) — a collection database
 *
 * Returns:
 *   error — the statement failed
 *
 * Example:
 *   err := ensureHarvestLog(db)
 */
func ensureHarvestLog(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS harvest_log (
	link TEXT,
	label TEXT,
	at TEXT,
	status INTEGER,
	ok INTEGER,
	error TEXT
);
CREATE INDEX IF NOT EXISTS harvest_log_link_at ON harvest_log (link, at);`)
	if err != nil {
		return fmt.Errorf("creating harvest_log: %w", err)
	}
	return nil
}

/** logAttempt records one attempt to fetch a listed feed: the HTTP status of
 * the response the outcome came from (0 when none), and the error when it
 * failed. A failure to write the log is a warning, never a reason to stop the
 * harvest.
 *
 * Parameters:
 *   db (*sql.DB) — a collection database with harvest_log
 *   eout (io.Writer) — receives the warning
 *   link (string) — the address as listed
 *   label (string) — the label as listed
 *   err (error) — nil when the feed was fetched and parsed
 *
 * Example:
 *   logAttempt(db, os.Stderr, "https://example.org/feed", "Example", err)
 */
func logAttempt(db *sql.DB, eout io.Writer, link, label string, err error) {
	status, ok, text := 200, 1, ""
	if err != nil {
		status, ok, text = statusOf(err), 0, err.Error()
	}
	_, werr := db.Exec(`INSERT INTO harvest_log (link, label, at, status, ok, error) VALUES (?1, ?2, ?3, ?4, ?5, ?6)`,
		link, label, clock().UTC().Format(logTimeLayout), status, ok, text)
	if werr != nil {
		fmt.Fprintf(eout, "warning (%s %s): the harvest log was not written: %s\n", label, link, werr)
	}
}

// statusWords shows an HTTP status as the number and its standard reason
// phrase, so nobody has to remember the numbers (DR-0038 5). A status the
// phrase table does not know shows the number alone; no status says so.
func statusWords(status int) string {
	switch {
	case status == 0:
		return "none (no answer)"
	case http.StatusText(status) == "":
		return fmt.Sprint(status)
	}
	return fmt.Sprintf("%d %s", status, http.StatusText(status))
}

// failingFeed is a listed feed that keeps failing.
type failingFeed struct {
	label, link string
	count       int
	since       string
	status      int
	errText     string
}

/** failingFeeds returns the listed feeds that are failing: every attempt in
 * the last failingDays days failed, and there were at least failingAttempts of
 * them (DR-0038 4). count and since cover the unbroken run of failures up to
 * the newest attempt, whatever the window.
 *
 * Parameters:
 *   db (*sql.DB) — a collection database with harvest_log
 *   links ([]string) — the addresses still listed; others are not reported
 *   p (harvestPolicy) — the window and the number of attempts
 *
 * Returns:
 *   []failingFeed — in the order listed
 *   error — the log could not be read
 *
 * Example:
 *   bad, err := failingFeeds(db, []string{"https://example.org/feed"}, defaultHarvestPolicy)
 */
func failingFeeds(db *sql.DB, links []string, p harvestPolicy) ([]failingFeed, error) {
	from := clock().UTC().AddDate(0, 0, -p.failingDays).Format(logTimeLayout)
	var out []failingFeed
	seen := map[string]bool{}
	for _, link := range links {
		if seen[link] {
			continue
		}
		seen[link] = true
		var inWindow, failed int
		if err := db.QueryRow(`SELECT COUNT(*), COALESCE(SUM(1 - ok), 0) FROM harvest_log WHERE link = ?1 AND at >= ?2`,
			link, from).Scan(&inWindow, &failed); err != nil {
			return nil, err
		}
		if inWindow < p.failingAttempts || failed != inWindow {
			continue
		}
		rows, err := db.Query(`SELECT label, at, status, ok, error FROM harvest_log WHERE link = ?1 ORDER BY at DESC, rowid DESC`, link)
		if err != nil {
			return nil, err
		}
		ff := failingFeed{link: link}
		first := true
		for rows.Next() {
			var label, at, text string
			var status, ok int
			if err := rows.Scan(&label, &at, &status, &ok, &text); err != nil {
				rows.Close()
				return nil, err
			}
			if ok == 1 {
				break
			}
			if first {
				ff.label, ff.status, ff.errText = label, status, text
				first = false
			}
			ff.count++
			ff.since = at
		}
		rows.Close()
		out = append(out, ff)
	}
	return out, nil
}

// warning is the line harvest writes for a failing feed.
func (ff failingFeed) warning(collection string) string {
	since := ff.since
	if len(since) > 10 {
		since = since[:10]
	}
	s := fmt.Sprintf("warning %s: %s %s has failed %d times since %s, last status %s",
		collection, ff.label, ff.link, ff.count, since, statusWords(ff.status))
	// The reason phrase says it all for an error status; an answer that was
	// not a feed, or no answer, needs the message.
	if ff.status == 0 || ff.status == http.StatusOK {
		s += ": " + strings.TrimSpace(ff.errText)
	}
	return s
}

/** reportFailingFeeds ends a collection's harvest: it deletes log rows older
 * than logDays, then writes one warning to eout for each listed feed that
 * keeps failing. It writes nothing when there is none, and never changes the
 * run's result: a log that cannot be read is itself only a warning.
 *
 * Parameters:
 *   db (*sql.DB) — the collection database
 *   eout (io.Writer) — receives the warnings
 *   collection (string) — the feed list's file name, for the warning
 *   links — the feed list's links (URL and Label)
 *   p (harvestPolicy) — the settings
 *
 * Example:
 *   reportFailingFeeds(db, os.Stderr, "news.md", links, defaultHarvestPolicy)
 */
func reportFailingFeeds(db *sql.DB, eout io.Writer, collection string, links []Link, p harvestPolicy) {
	cutoff := clock().UTC().AddDate(0, 0, -p.logDays).Format(logTimeLayout)
	if _, err := db.Exec(`DELETE FROM harvest_log WHERE at < ?1`, cutoff); err != nil {
		fmt.Fprintf(eout, "warning %s: the harvest log was not trimmed: %s\n", collection, err)
	}
	urls := make([]string, 0, len(links))
	for _, l := range links {
		urls = append(urls, l.URL)
	}
	bad, err := failingFeeds(db, urls, p)
	if err != nil {
		fmt.Fprintf(eout, "warning %s: the harvest log was not read: %s\n", collection, err)
		return
	}
	for _, ff := range bad {
		fmt.Fprintln(eout, ff.warning(collection))
	}
}
