package antennaApp

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// ago is a time in the collection's default 21-day window, n hours old.
func ago(n int) time.Time { return time.Now().UTC().Add(-time.Duration(n) * time.Hour) }

// DR-0039: a feed shows at most a set number of its newest items in a
// collection's output, 1000 by default, set per feed in antenna.yaml. These
// tests harvest real feeds from a test server and run the real generate verb.

// numberedFeed is an RSS feed of n items, Item 1 oldest to Item n newest.
func numberedFeed(name string, n int) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0"?><rss version="2.0"><channel><title>` + name + `</title><link>https://example.org/` + name + `/</link>`)
	for i := 1; i <= n; i++ {
		fmt.Fprintf(&b, `<item><title>%s item %02d</title><link>https://example.org/%s/%02d</link><guid>%s-%02d</guid>`+
			`<pubDate>%s</pubDate><description>Body of %s %02d</description></item>`,
			name, i, name, i, name, i, ago(n-i).Format(time.RFC1123), name, i)
	}
	b.WriteString(`</channel></rss>`)
	return b.String()
}

// capSite harvests feeds a, b and c (12 items each unless sizes say otherwise)
// into news.db, and appends extra top-level lines to antenna.yaml. It returns
// the test server's address.
func capSite(t *testing.T, extraConfig string, sizes map[string]int) string {
	t.Helper()
	routes := map[string]route{}
	for _, name := range []string{"a", "b", "c"} {
		n := 12
		if v, ok := sizes[name]; ok {
			n = v
		}
		routes["/"+name+".xml"] = route{"application/rss+xml", numberedFeed(name, n)}
	}
	srv, _ := fakeSite(routes)
	t.Cleanup(srv.Close)
	harvestSite(t, "[A]("+srv.URL+"/a.xml)", "[B]("+srv.URL+"/b.xml)", "[C]("+srv.URL+"/c.xml)")
	if _, _, err := runCapture(t, "harvest", "news.md"); err != nil {
		t.Fatalf("harvest: %s", err)
	}
	// With no extra settings the file stays as init made it, with its cap of 1000.
	if extraConfig != "" {
		setCapConfig(t, strings.ReplaceAll(extraConfig, "{{SITE}}", srv.URL))
	}
	return srv.URL
}

// dropInitCap removes the max_items_per_feed line antenna init wrote.
func dropInitCap(t *testing.T) {
	t.Helper()
	b, err := os.ReadFile("antenna.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var keep []string
	for _, line := range strings.Split(string(b), "\n") {
		if !strings.HasPrefix(line, "max_items_per_feed:") {
			keep = append(keep, line)
		}
	}
	if err := os.WriteFile("antenna.yaml", []byte(strings.Join(keep, "\n")), 0o644); err != nil {
		t.Fatal(err)
	}
}

// setCapConfig removes any earlier cap settings, including the one init
// wrote, and appends extra to antenna.yaml.
func setCapConfig(t *testing.T, extra string) {
	t.Helper()
	dropInitCap(t)
	b, err := os.ReadFile("antenna.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var keep []string
	skipping := false
	for _, line := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(line, "# cap-begin") {
			skipping = true
		}
		if !skipping {
			keep = append(keep, line)
		}
	}
	text := strings.Join(keep, "\n")
	if extra != "" {
		text += "\n# cap-begin\n" + extra + "\n"
	}
	if err := os.WriteFile("antenna.yaml", []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func generateCap(t *testing.T) (eout string) {
	t.Helper()
	_, eout, err := runCapture(t, "generate")
	if err != nil {
		t.Fatalf("generate: %s\n%s", err, eout)
	}
	return eout
}

// shown counts the items of a feed in text by their link.
func shown(text, feed string) int { return strings.Count(text, "https://example.org/"+feed+"/") }

func itemStatuses(t *testing.T) string {
	t.Helper()
	db, err := sql.Open("sqlite", "news.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows, err := db.Query(`SELECT link, status FROM items ORDER BY link`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var b strings.Builder
	for rows.Next() {
		var l, s string
		rows.Scan(&l, &s)
		b.WriteString(l + "=" + s + "\n")
	}
	return b.String()
}

// The default cap with a setting of 5: the five newest of each feed, the same
// five in every output, and the database untouched.
func TestGenerateCapsEachFeedToItsNewestItems(t *testing.T) {
	capSite(t, "max_items_per_feed: 5", nil)
	generateCap(t)
	for _, file := range []string{"news_items.html", "news_items.md", "news.xml"} {
		text := findOutput(t, file)
		for _, feed := range []string{"a", "b", "c"} {
			// A link appears more than once in an item (heading, footer), so
			// look for each item's own title instead.
			for i := 1; i <= 12; i++ {
				title := fmt.Sprintf("%s item %02d", feed, i)
				has := strings.Contains(text, title)
				if want := i > 7; has != want {
					t.Errorf("%s: %q present=%v, want %v", file, title, has, want)
				}
			}
		}
	}
}

func TestGenerateCapHidesItemsAndDoesNotChangeTheDatabase(t *testing.T) {
	capSite(t, "", nil)
	generateCap(t)
	before := itemStatuses(t)
	setCapConfig(t, "max_items_per_feed: 3")
	generateCap(t)
	if after := itemStatuses(t); after != before {
		t.Errorf("generate with a cap changed items.status:\n%s\n--- to ---\n%s", before, after)
	}
	if got := findOutput(t, "news_items.md"); strings.Contains(got, "a item 01") || !strings.Contains(got, "a item 12") {
		t.Errorf("cap of 3 not applied:\n%s", got)
	}
	// Taking the cap off brings the items back.
	setCapConfig(t, "")
	generateCap(t)
	if got := findOutput(t, "news_items.md"); !strings.Contains(got, "a item 01") {
		t.Errorf("an item hidden by the cap did not return when the cap was removed")
	}
}

func TestGenerateFeedLimitsSetEachFeedByAddress(t *testing.T) {
	capSite(t, "max_items_per_feed: 5\nfeed_limits:\n  {{SITE}}/a.xml: 2\n  {{SITE}}/c.xml: 0\n", nil)
	generateCap(t)
	for _, file := range []string{"news_items.html", "news_items.md", "news.xml"} {
		text := findOutput(t, file)
		count := func(feed string) int {
			n := 0
			for i := 1; i <= 12; i++ {
				if strings.Contains(text, fmt.Sprintf("%s item %02d", feed, i)) {
					n++
				}
			}
			return n
		}
		if a, b, c := count("a"), count("b"), count("c"); a != 2 || b != 5 || c != 12 {
			t.Errorf("%s: a=%d b=%d c=%d, want 2 5 12", file, a, b, c)
		}
	}
}

func TestGenerateZeroMeansNoCap(t *testing.T) {
	capSite(t, "max_items_per_feed: 0", nil)
	generateCap(t)
	if got := findOutput(t, "news_items.md"); !strings.Contains(got, "a item 01") || strings.Contains(got, "feeds over") {
		t.Errorf("max_items_per_feed: 0 still capped:\n%s", got)
	}
}

// addBigFeed adds 1005 published items to feed a's channel.
func addBigFeed(t *testing.T, site string) {
	t.Helper()
	db, err := sql.Open("sqlite", "news.db")
	if err != nil {
		t.Fatal(err)
	}
	tx, _ := db.Begin()
	for i := 1; i <= 1005; i++ {
		_, err := tx.Exec(`INSERT INTO items (link, title, description, authors, enclosures, guid, pubDate, dcExt, channel, status, updated, label, postPath, sourceMarkdown, categories)
			VALUES (?1, ?2, 'body', '', '', ?1, ?3, '', ?4, 'published', ?3, 'A', '', '', '')`,
			fmt.Sprintf("https://example.org/big/%04d", i), fmt.Sprintf("big item %04d", i),
			ago(1+i/60).Add(-time.Duration(i%60)*time.Second).Format("2006-01-02 15:04:05"), site+"/a.xml")
		if err != nil {
			t.Fatal(err)
		}
	}
	tx.Commit()
	db.Close()
}

func itemHeadings(text string) int {
	return strings.Count(text, "big item ") + strings.Count(text, "a item ")
}

// A site made by antenna init has a cap of 1000 (DR-0039 3).
func TestInitWritesACapOfOneThousand(t *testing.T) {
	newWorkspace(t, true)
	if got := readText(t, "antenna.yaml"); !strings.Contains(got, "max_items_per_feed: 1000\n") {
		t.Errorf("init did not write the cap:\n%s", got)
	}
}

// An antenna.yaml that exists is left as it is: running init again adds no cap.
func TestInitAddsNoCapToAnExistingFile(t *testing.T) {
	newWorkspace(t, true)
	dropInitCap(t)
	if err := runIn(t, "init"); err != nil {
		t.Fatalf("init: %s", err)
	}
	if got := readText(t, "antenna.yaml"); strings.Contains(got, "max_items_per_feed") {
		t.Errorf("init added a cap to an existing antenna.yaml:\n%s", got)
	}
}

func TestGenerateOnANewSiteCapsAtOneThousand(t *testing.T) {
	site := capSite(t, "", map[string]int{"a": 1, "b": 1, "c": 1})
	if got := readText(t, "antenna.yaml"); !strings.Contains(got, "max_items_per_feed: 1000") {
		t.Fatalf("test setup: init wrote no cap:\n%s", got)
	}
	addBigFeed(t, site)
	eout := generateCap(t)
	if n := itemHeadings(findOutput(t, "news_items.md")); n != 1000 {
		t.Errorf("feed shows %d items, want 1000", n)
	}
	if !strings.Contains(eout, "1 feed over its cap: A shows 1000 of 1006") {
		t.Errorf("warning wrong:\n%s", eout)
	}
}

// No setting in antenna.yaml means no cap (DR-0039 3).
func TestGenerateWithNoSettingHasNoCap(t *testing.T) {
	site := capSite(t, "", map[string]int{"a": 1, "b": 1, "c": 1})
	dropInitCap(t)
	addBigFeed(t, site)
	eout := generateCap(t)
	if n := itemHeadings(findOutput(t, "news_items.md")); n != 1006 {
		t.Errorf("feed shows %d items, want all 1006", n)
	}
	if strings.Contains(eout, "over its cap") || strings.Contains(eout, "over their cap") {
		t.Errorf("warned with no cap set:\n%s", eout)
	}
}

func TestGenerateNeverCapsTheSitesOwnPosts(t *testing.T) {
	capSite(t, "max_items_per_feed: 1", map[string]int{"a": 1, "b": 1, "c": 1})
	for i := 1; i <= 6; i++ {
		sitemapExec(t, "news.db", fmt.Sprintf(`INSERT INTO items (link, postPath, title, description, authors, enclosures, guid, pubDate, dcExt, channel, status, updated, label, sourceMarkdown, categories)
			VALUES ('https://example.com/blog/p%d', 'blog/p%d.md', 'own post %d', 'body', '', '', 'g%d', '%s', '', '', 'published', '%s', '', '', '')`, i, i, i, i, ago(i).Format("2006-01-02 15:04:05"), ago(i).Format("2006-01-02 15:04:05")))
	}
	generateCap(t)
	got := findOutput(t, "news_items.md")
	for i := 1; i <= 6; i++ {
		if !strings.Contains(got, fmt.Sprintf("own post %d", i)) {
			t.Errorf("own post %d was capped", i)
		}
	}
}

func TestGenerateWarnsWhichFeedsWereCut(t *testing.T) {
	capSite(t, "max_items_per_feed: 5\nfeed_limits:\n  {{SITE}}/c.xml: 0\n", map[string]int{"a": 12, "b": 12, "c": 12})
	eout := generateCap(t)
	if !strings.Contains(eout, "warning news.md: 2 feeds over their cap: A shows 5 of 12, B shows 5 of 12") {
		t.Errorf("warning wrong:\n%s", eout)
	}
	if strings.Contains(eout, "C shows") {
		t.Errorf("an uncapped feed was named:\n%s", eout)
	}
}

func TestGenerateWithNothingOverTheCapIsQuiet(t *testing.T) {
	capSite(t, "max_items_per_feed: 12", nil)
	if eout := generateCap(t); strings.Contains(eout, "over their cap") || strings.Contains(eout, "over its cap") {
		t.Errorf("warned with nothing cut:\n%s", eout)
	}
}

func TestBadCapSettingsAreConfigurationErrors(t *testing.T) {
	for _, bad := range []string{
		"max_items_per_feed: -1", "max_items_per_feed: many", "max_items_per_feed: 2.5",
		"feed_limits:\n  https://example.org/a.xml: -3", "feed_limits:\n  https://example.org/a.xml: lots",
	} {
		newWorkspace(t, true)
		dropInitCap(t)
		cfg, _ := os.ReadFile("antenna.yaml")
		os.WriteFile("antenna.yaml", append(cfg, []byte("\n"+bad+"\n")...), 0o644)
		_, _, err := runCapture(t, "generate")
		wantClass(t, bad, err, classConfig)
	}
}

// Found on a real site, 2026-11-25: feed_limits written as a YAML list gave
// "line 7: cannot unmarshal !!seq into map[string]int", which names neither the
// setting nor what it wants. It must say which setting and what shape.
func TestFeedLimitsAsAListIsExplained(t *testing.T) {
	for name, bad := range map[string]string{
		"list of single-key maps": "feed_limits:\n  - https://example.org/a.xml: 200\n",
		"list of url and limit":   "feed_limits:\n  - url: https://example.org/a.xml\n    limit: 200\n",
		"a single number":         "feed_limits: 200\n",
	} {
		newWorkspace(t, true)
		dropInitCap(t)
		cfg, _ := os.ReadFile("antenna.yaml")
		os.WriteFile("antenna.yaml", append(cfg, []byte("\n"+bad)...), 0o644)
		_, _, err := runCapture(t, "generate")
		wantClass(t, name, err, classConfig)
		if err != nil && (!strings.Contains(err.Error(), "feed_limits") || !strings.Contains(err.Error(), "address")) {
			t.Errorf("%s: the message does not name feed_limits and say what it wants:\n%s", name, err)
		}
	}
}

func TestMaxItemsPerFeedAsAListIsExplained(t *testing.T) {
	newWorkspace(t, true)
	dropInitCap(t)
	cfg, _ := os.ReadFile("antenna.yaml")
	os.WriteFile("antenna.yaml", append(cfg, []byte("\nmax_items_per_feed:\n  - 5\n")...), 0o644)
	_, _, err := runCapture(t, "generate")
	wantClass(t, "list", err, classConfig)
	if err != nil && !strings.Contains(err.Error(), "max_items_per_feed") {
		t.Errorf("the message does not name max_items_per_feed:\n%s", err)
	}
}
