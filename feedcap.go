package antennaApp

import (
	"database/sql"
	"fmt"
	"io"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// DR-0039: a feed shows at most a set number of its newest items in a
// collection's output, 1000 by default, set per feed in antenna.yaml. The
// items stay in the database; the cap only hides them when a page is made.

// initMaxItemsPerFeed is the cap antenna init writes into the antenna.yaml it
// creates. A file with no max_items_per_feed has no cap (DR-0039 3).
const initMaxItemsPerFeed = 1000

/** feedCaps is how many items each feed may show: a default and the feeds
 * that differ, by address as listed. 0 means no cap.
 *
 * Fields:
 *   def (int) — the number for a feed with no entry in limits
 *   limits (map[string]int) — feed address to its number
 */
type feedCaps struct {
	def    int
	limits map[string]int
}

// forFeed returns the number of items the feed at address may show; 0 means all.
func (c feedCaps) forFeed(address string) int {
	if n, ok := c.limits[address]; ok {
		return n
	}
	return c.def
}

/** feedCaps reads the cap settings of antenna.yaml: max_items_per_feed (no
 * cap when absent) and feed_limits.
 *
 * Returns:
 *   feedCaps — the settings, defaults filled in
 *
 * Example:
 *   caps := cfg.feedCaps()
 *   n := caps.forFeed("https://example.org/feed.xml")
 */
func (cfg *AppConfig) feedCaps() feedCaps {
	c := feedCaps{limits: cfg.FeedLimits}
	if cfg.MaxItemsPerFeed != nil {
		c.def = *cfg.MaxItemsPerFeed
	}
	return c
}

// validateFeedCaps checks the cap settings; a number below 0 is a
// configuration error found when the configuration loads (DR-0039 4). A value
// that is not a whole number is refused earlier, when the YAML is read.
func (cfg *AppConfig) validateFeedCaps() error {
	if cfg.MaxItemsPerFeed != nil && *cfg.MaxItemsPerFeed < 0 {
		return fmt.Errorf("max_items_per_feed is %d; it must be 0 or more (0 is no cap)", *cfg.MaxItemsPerFeed)
	}
	for address, n := range cfg.FeedLimits {
		if n < 0 {
			return fmt.Errorf("feed_limits for %s is %d; it must be 0 or more (0 is no cap)", address, n)
		}
	}
	return nil
}

// validateFeedCapSyntax refuses a cap written as anything but a whole number.
// The YAML decoder rounds 2.5 into an int without complaint, so the nodes are
// checked as they were written (DR-0039 4).
func validateFeedCapSyntax(src []byte) error {
	var raw struct {
		Max    yaml.Node            `yaml:"max_items_per_feed"`
		Limits map[string]yaml.Node `yaml:"feed_limits"`
	}
	if err := yaml.Unmarshal(src, &raw); err != nil {
		return nil // the real decode reports it
	}
	if raw.Max.Kind != 0 && raw.Max.Tag != "!!int" {
		return fmt.Errorf("max_items_per_feed is %q; it must be a whole number of 0 or more", raw.Max.Value)
	}
	for address, n := range raw.Limits {
		if n.Tag != "!!int" {
			return fmt.Errorf("feed_limits for %s is %q; it must be a whole number of 0 or more", address, n.Value)
		}
	}
	return nil
}

const displayItemColumns = `link, title, description, authors,
  enclosures, guid, pubDate, dcExt,
  channel, status, updated, label,
  postPath, ifnull(sourceMarkdown, '') as sourceMarkdown,
  ifnull(categories, '') as categories`

/** displayItems returns the query every rendering of a collection reads its
 * items with, and its arguments: SQLDisplayItems limited so that no feed shows
 * more than its cap (DR-0039 1, 5). Items with a postPath are the site's own and
 * are never limited. The HTML page, the items Markdown and the RSS feed all use
 * it, so they list the same items.
 *
 * Returns:
 *   string — the SQL
 *   []any — its arguments
 *
 * Example:
 *   stmt, args := gen.displayItems()
 *   rows, err := db.Query(stmt, args...)
 */
func (gen *Generator) displayItems() (string, []any) {
	var args []any
	var values []string
	addresses := make([]string, 0, len(gen.caps.limits))
	for a := range gen.caps.limits {
		addresses = append(addresses, a)
	}
	sort.Strings(addresses)
	for _, a := range addresses {
		values = append(values, "(?, ?)")
		args = append(args, a, gen.caps.limits[a])
	}
	caps := "SELECT '' AS channel, 0 AS n WHERE 0"
	if len(values) > 0 {
		caps = "VALUES " + strings.Join(values, ", ")
	}
	args = append(args, gen.caps.def)
	stmt := `WITH caps(channel, n) AS (` + caps + `),
ranked AS (
  SELECT rowid AS rid,
    ROW_NUMBER() OVER (PARTITION BY channel ORDER BY pubDate DESC, updated DESC) AS rn,
    COALESCE((SELECT n FROM caps WHERE caps.channel = items.channel), ?) AS cap
  FROM items
  WHERE (description != '' OR title = '') AND status = 'published' AND COALESCE(postPath, '') = ''
)
SELECT ` + displayItemColumns + `
FROM items WHERE (description != '' OR title = '') AND status = 'published'
  AND (COALESCE(postPath, '') != '' OR rowid IN (SELECT rid FROM ranked WHERE cap = 0 OR rn <= cap))
ORDER BY pubDate DESC, updated DESC;`
	return stmt, args
}

// cutFeed is a feed that has more items than its cap.
type cutFeed struct {
	address, label string
	shown, total   int
}

/** reportCutFeeds warns once for a collection when feeds had more items than
 * their cap, naming each with how many it shows and how many it has (DR-0039
 * 6). It writes nothing when no feed was cut and never changes the result of
 * the run.
 *
 * Parameters:
 *   db (*sql.DB) — the collection database
 *   eout (io.Writer) — receives the warning
 *   collection (string) — the feed list's file name
 *
 * Example:
 *   gen.reportCutFeeds(db, os.Stderr, "news.md")
 */
func (gen *Generator) reportCutFeeds(db *sql.DB, eout io.Writer, collection string) {
	if eout == nil {
		return
	}
	rows, err := db.Query(`SELECT channel, COALESCE(MAX(label), ''), COUNT(*) FROM items
WHERE (description != '' OR title = '') AND status = 'published' AND COALESCE(postPath, '') = ''
GROUP BY channel ORDER BY channel`)
	if err != nil {
		return
	}
	defer rows.Close()
	var cut []cutFeed
	for rows.Next() {
		var f cutFeed
		var ch sql.NullString
		if rows.Scan(&ch, &f.label, &f.total) != nil {
			return
		}
		f.address = ch.String
		if limit := gen.caps.forFeed(f.address); limit > 0 && f.total > limit {
			f.shown = limit
			if f.label == "" {
				f.label = f.address
			}
			cut = append(cut, f)
		}
	}
	if len(cut) == 0 {
		return
	}
	parts := make([]string, 0, len(cut))
	for _, f := range cut {
		parts = append(parts, fmt.Sprintf("%s shows %d of %d", f.label, f.shown, f.total))
	}
	noun := "1 feed over its cap"
	if len(cut) > 1 {
		noun = fmt.Sprintf("%d feeds over their cap", len(cut))
	}
	fmt.Fprintf(eout, "warning %s: %s: %s\n", collection, noun, strings.Join(parts, ", "))
}
