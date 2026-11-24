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
*/
package antennaApp

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	// 3rd Party pacakges
	html2md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/mmcdole/gofeed"
)

func (app AntennaApp) Harvest(out io.Writer, eout io.Writer, cfgName string, args []string) error {
	cfg := &AppConfig{}
	if err := cfg.LoadConfig(cfgName); err != nil {
		return err
	}
	if len(args) == 0 {
		for _, col := range cfg.Collections {
			args = append(args, col.File)
		}
	}
	// One fetcher for the whole run: each host's robots.txt is read once and
	// the pause between requests to a host holds across collections.
	f := cfg.harvestFetcher(eout)
	// Harvest every collection asked for, then report: a failed collection
	// does not stop the rest, and does not let the command exit 0 either.
	tally := &failureTally{}
	for _, cName := range args {
		col, err := cfg.GetCollection(cName)
		if err != nil {
			return err
		}
		err = col.HarvestWith(out, eout, f)
		if err != nil {
			fmt.Fprintf(eout, "warning %s: %s\n", col.File, err)
		}
		tally.add(err)
	}
	if f.skipped > 0 {
		fmt.Fprintf(eout, "%d request(s) left alone because of robots.txt\n", f.skipped)
	}
	return tally.err("collections")
}

/** Harvest fetches every feed the collection lists, with its own fetcher
 * (one robots.txt read per host, the default pause).
 *
 * Parameters:
 *   out       (io.Writer) — progress
 *   eout      (io.Writer) — warnings
 *   userAgent (string)    — the User-Agent, empty for antenna/VERSION HASH
 *
 * Returns:
 *   error — the first failure's class when any feed failed, else nil
 *
 * Example:
 *   err := col.Harvest(os.Stdout, os.Stderr, cfg.UserAgent)
 */
func (collection *Collection) Harvest(out io.Writer, eout io.Writer, userAgent string) error {
	return collection.HarvestWith(out, eout, newFetcher(userAgent, defaultHarvestPause, eout))
}

/** HarvestWith fetches every feed the collection lists through f. A listed
 * address may be a feed or a site: discovery finds the feed (DR-0030 1) and
 * the address it found is stored in the channel's discovered_link, so later
 * runs go straight to it. A link that robots.txt forbids is skipped, not
 * failed; a site with no feed is a data error for that link alone.
 *
 * Parameters:
 *   out  (io.Writer) — progress
 *   eout (io.Writer) — warnings
 *   f    (*fetcher)  — makes the requests; shared across collections in a run
 *
 * Returns:
 *   error — the first failure's class when any link failed, else nil
 *
 * Example:
 *   err := col.HarvestWith(os.Stdout, os.Stderr, cfg.harvestFetcher(os.Stderr))
 */
func (collection *Collection) HarvestWith(out io.Writer, eout io.Writer, f *fetcher) error {
	userAgent := f.userAgent
	// LoadCommonMark accepts Markdown (.md) or ODT/OTT collection files.
	doc, err := LoadCommonMark(collection.File)
	if err != nil {
		return err
	}
	links, err := doc.GetLinks()
	if err != nil {
		return err
	}
	// Open DB so we have a place to write data.
	dsn := collection.DbName
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	if err := ensureChannelColumns(db); err != nil {
		return err
	}
	// Retrieve is feed for links and save channel and items for feed. One
	// feed failing does not stop the others; the failures are counted and
	// returned at the end.
	tally := &failureTally{}
	for _, link := range links {
		// Retrieve feed data: the listed address may be the feed or a site
		// whose feed discovery finds (DR-0030 1).
		got, err := resolveFeed(f, link.URL, rememberedFeed(db, link.URL))
		if err != nil {
			if isSkip(err) {
				// Already reported once by the fetcher; not a failure.
				tally.add(nil)
				continue
			}
			fmt.Fprintf(eout, "warning (%s %s): %s\n", link.Label, link.URL, err)
			tally.add(fmt.Errorf("%s %s: %w", link.Label, link.URL, err))
			continue
		}
		feed := got.feed
		discovered := ""
		if got.address != link.URL {
			discovered = got.address
		}

		// Save the Channel data for the feed
		if err := saveChannel(db, link.URL, link.Label, feed, discovered); err != nil {
			fmt.Fprintf(eout, "failed to save chanel %q, %s\n", link.URL, err)
			tally.add(fmt.Errorf("saving channel %s: %w", link.URL, err))
			continue
		}
		tally.add(nil)
		// Setup a progress output
		t0 := time.Now()
		rptTime := time.Now()
		reportProgress := false
		tot := feed.Len()
		fmt.Fprintf(out, "processing %d items from %s %s\n", tot, link.Label, userAgent)
		i := 0
		// Save the item data for the feed
		for _, item := range feed.Items {
			if strings.HasPrefix(item.Link, "/") {
				item.Link = fmt.Sprintf("%s%s", strings.TrimSuffix(feed.Link, "/"), item.Link)
			}
			// Default to the feed dates if not set at item level
			if item.Updated == "" && item.UpdatedParsed == nil {
				item.Updated = feed.Updated
				item.UpdatedParsed = feed.UpdatedParsed
			}
			if item.Published == "" && item.PublishedParsed == nil {
				item.Published = feed.Published
				item.PublishedParsed = feed.PublishedParsed
			}
			// Add items from feed to database table
			if err := saveItem(db, link.Label, link.URL, "", item); err != nil {
				return err
			}
			if rptTime, reportProgress = CheckWaitInterval(rptTime, (20 * time.Second)); reportProgress {
				fmt.Fprintf(out, "(%d/%d) %s", i, tot, ProgressETA(t0, i, tot))
			}
			i++
		}
		fmt.Fprintf(out, "processed %d/%d from %s %s\n", i, tot, link.Label, userAgent)
	}
	return tally.err("feeds")
}

// webget retrieves the feed at href and parses it. It is the one-request
// helper: no discovery and no robots.txt (a harvest run uses resolveFeed on
// a shared fetcher). Redirects are followed.
// Uses mmcdole's gofeed, see docs at https://pkg.go.dev/github.com/mmcdole/gofeed
func webget(userAgent string, href string) (*gofeed.Feed, error) {
	f := newFetcher(userAgent, 0, io.Discard)
	res, err := f.fetch(href, feedAccept, false)
	if err != nil {
		return nil, err
	}
	if res.status != http.StatusOK {
		return nil, statusErr(res)
	}
	return parseFeedBody(res.body, href)
}

// ensureChannelColumns adds the columns a collection database made by an
// earlier release lacks (DR-0030 amendment: discovered_link) and keeps every
// row.
func ensureChannelColumns(db *sql.DB) error {
	rows, err := db.Query(`PRAGMA table_info(channels)`)
	if err != nil {
		return err
	}
	have := map[string]bool{}
	for rows.Next() {
		var (
			cid     int
			name    string
			ctype   string
			notnull int
			dflt    sql.NullString
			pk      int
		)
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			rows.Close()
			return err
		}
		have[name] = true
	}
	rows.Close()
	if len(have) == 0 {
		return nil // no channels table
	}
	if !have["discovered_link"] {
		if _, err := db.Exec(`ALTER TABLE channels ADD COLUMN discovered_link TEXT DEFAULT ''`); err != nil {
			return fmt.Errorf("adding channels.discovered_link: %w", err)
		}
	}
	return nil
}

// rememberedFeed returns the feed address discovery found for a listed link on
// an earlier run, or "".
func rememberedFeed(db *sql.DB, link string) string {
	var addr sql.NullString
	if err := db.QueryRow(`SELECT discovered_link FROM channels WHERE link = ?1`, link).Scan(&addr); err != nil {
		return ""
	}
	return addr.String
}

// saveChannel will write the Channel information to a skimmer channel table.
func saveChannel(db *sql.DB, link string, feedLabel string, channel *gofeed.Feed, discovered string) error {
	/*
	   link, title, description, feed_link, links,
	   updated, published,
	   authors, language, copyright, generator,
	   categories, feed_type, feed_version, enclosure
	*/
	var (
		err           error
		src           []byte
		title         string
		linksStr      string
		authorsStr    string
		categoriesStr string
	)
	linksStr = link
	title = feedLabel
	if feedLabel == "" {
		title = channel.Title
	}
	if channel.Links != nil {
		src, err = JSONMarshal(channel.Links)
		if err != nil {
			return err
		}
		linksStr = fmt.Sprintf("%s", src)
	}
	if channel.Authors != nil {
		src, err = JSONMarshal(channel.Authors)
		if err != nil {
			return err
		}
		authorsStr = fmt.Sprintf("%s", src)
	}
	if channel.Categories != nil {
		src, err = JSONMarshal(channel.Categories)
		if err != nil {
			return err
		}
		categoriesStr = fmt.Sprintf("%s", src)
	}

	stmt := SQLUpdateChannel
	_, err = db.Exec(stmt,
		&link, &title, channel.Description, channel.FeedLink, linksStr,
		channel.Updated, channel.Published,
		authorsStr, channel.Language, channel.Copyright, channel.Generator,
		categoriesStr, channel.FeedType, channel.FeedVersion, discovered)
	if err != nil {
		return fmt.Errorf("%w\nstmt: %s", err, stmt)
	}
	return nil
}

// saveItem saves a gofeed item to the item table in the skimmer database
func saveItem(db *sql.DB, feedLabel string, channel string, status string, item *gofeed.Item) error {
	var (
		pubDate string
		updated string
	)

	if item.UpdatedParsed != nil {
		updated = item.UpdatedParsed.Format("2006-01-02 15:04:05")
	} else {
		updated = item.Updated
	}

	if item.PublishedParsed != nil {
		pubDate = item.PublishedParsed.Format("2006-01-02 15:04:05")
	} else {
		pubDate = item.Published
	}
	var (
		authors    []byte
		dcExt      []byte
		enclosures []byte
		categories []byte
		err        error
	)
	if item.DublinCoreExt != nil {
		dcExt, err = json.Marshal(item.DublinCoreExt)
		if err != nil {
			return fmt.Errorf("failed to marshal item.DublinCoreExt, %w", err)
		}
	}
	if item.Enclosures != nil {
		enclosures, err = json.Marshal(item.Enclosures)
		if err != nil {
			return fmt.Errorf("failed to marshal item.Enclosures, %w", err)
		}
	}
	if item.Authors != nil {
		authors, err = json.Marshal(item.Authors)
		if err != nil {
			return fmt.Errorf("failed to marshal item.Authors, %w", err)
		}
	}
	if item.Categories != nil {
		categories, err = json.Marshal(item.Categories)
		if err != nil {
			return fmt.Errorf("failed to marshal item.Categories, %w", err)
		}
	}
	// FIXME: Need to find a feed that uses the source:markdown name space to verify this is how gofeed
	// handled it!
	sourceMarkdown := ""
	// NOTE: Need to look for source:markdown and get it populated.
	if item.Extensions != nil {
		if source, ok := item.Extensions["source"]; ok && len(source) > 0 {
			if data, ok := source["markdown"]; ok {
				for _, val := range data {
					if val.Name == "markdown" {
						sourceMarkdown = val.Value
					}
				}
			}
		}
	}
	// Check to see if we have sourceMarkdown set or if I want to try converting the description field.
	converter := html2md.NewConverter("", true, nil)
	if item.Description != "" && sourceMarkdown == "" {
		src, err := converter.ConvertString(item.Description)
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: %s failed to convert description %q, %s, skipping\n", feedLabel, item.Description, err)
		} else {
			sourceMarkdown = src
		}
	}
	// NOTE: postPath doens't exist for harvested items, only for posted ones but SQL statement
	// is used for both.
	postPath := ""
	stmt := SQLUpdateItem
	if _, err := db.Exec(stmt,
		item.Link, item.Title, item.Description, string(authors),
		string(enclosures), item.GUID, pubDate, string(dcExt),
		channel, status, updated, feedLabel, postPath, sourceMarkdown,
		string(categories)); err != nil {
		return fmt.Errorf("%w\nstmt: %s", err, stmt)
	}
	return nil
}
