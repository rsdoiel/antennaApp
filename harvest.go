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
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	// 3rd Party pacakges
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
	for _, cName := range args {
		col, err := cfg.GetCollection(cName)
		if err != nil {
			return err
		}
		if err := col.Harvest(out, eout, cfg.UserAgent); err != nil {
			fmt.Fprintf(eout, "warning %s: %s\n", col.File, err)
		}
	}
	return nil
}

func (collection *Collection) Harvest(out io.Writer, eout io.Writer, userAgent string) error {
	src, err := os.ReadFile(collection.File)
	if err != nil {
		return err
	}
	doc := &CommonMark{}
	if err := doc.Parse(src); err != nil {
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
	// Retrieve is feed for links and save channel and items for feed
	for _, link := range links {
		// Retrieve feed data
		feed, err := webget(userAgent, link.URL)
		if err != nil {
			fmt.Fprintf(eout, "warning (%s %s): %s\n", link.Label, link.URL, err)
			continue
		}

		// Save the Channel data for the feed
		if err := saveChannel(db, link.URL, link.Label, feed); err != nil {
			fmt.Fprintf(eout, "failed to save chanel %q, %s\n", link.URL, err)
			continue
		}
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
	return nil
}

func redirectHandler (req *http.Request, via []*http.Request) error {
	if len(via) >= 5 {
		urlList := []string{}
		for _, redirect := range via {
			urlList = append(urlList, redirect.URL.String())
		}
		return fmt.Errorf("stopped after 5 redirectos: %s", strings.Join(urlList, ", "))
	}
	fmt.Fprintf(os.Stderr, "redirecting to %s because %s\n", req.URL.String(), http.ErrUseLastResponse)
 	return http.ErrUseLastResponse
}

// webget retrieves a feed and parses it.
// Uses mmcdole's gofeed, see docs at https://pkg.go.dev/github.com/mmcdole/gofeed
func webget(userAgent string, href string) (*gofeed.Feed, error) {
	// NOTE: I'm assuming only http, https at this point, later this will
	// need to be split up so I can handle Gopher, Gemini and sftp.
	client := &http.Client{
	 	CheckRedirect: redirectHandler,
	}
	req, err := http.NewRequest("GET", href, nil)
	if err != nil {
		return nil, err
	}
	if userAgent == "" {
		req.Header.Set("User-Agent", fmt.Sprintf("antenna/%s %s", Version, ReleaseHash))
	} else {
		req.Header.Set("User-Agent", userAgent)
	}
	// Set the accepted content types.
	req.Header.Set("accept", "application/rss+xml, application/atom+xml, application/feed+json, application/xml, application/json;q=0.9, */*;q=0.8")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http error: %s", res.Status)
	}
	// See if we can clean up some stuff that'll break feed parsing
	src, err := io.ReadAll(res.Body)
	if err != err {
		return nil, err
	}
	src = bytes.ReplaceAll(src , []byte(``), []byte(``))
	buf := bytes.NewBuffer(src)

	fp := gofeed.NewParser()
	feed, err := fp.Parse(buf)
	if err != nil {
		return nil, fmt.Errorf("feed error for %q, %s", href, err)
	}
	if feed.Link == "" || feed.Link == "/" {
		u, err := url.Parse(href)
		if err != nil {
			return nil, err
		}
		u.Path = "/"
		feed.Link = u.String()
	}
	return feed, nil
}

// saveChannel will write the Channel information to a skimmer channel table.
func saveChannel(db *sql.DB, link string, feedLabel string, channel *gofeed.Feed) error {
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
		categoriesStr, channel.FeedType, channel.FeedVersion)
	if err != nil {
		return fmt.Errorf("%s\nstmt: %s", err, stmt)
	}
	return nil
}

// saveItem saves a gofeed item to the item table in the skimmer database
func saveItem(db *sql.DB, feedLabel string, channel string, status string, item *gofeed.Item) error {
	var (
		pubDate string
		updated   string
	)
	if item.UpdatedParsed != nil {
		updated = item.UpdatedParsed.Format("2006-01-02 15:04:05")
	}
	if item.PublishedParsed != nil {
		pubDate = item.PublishedParsed.Format("2006-01-02 15:04:05")
	}
	var (
		authors []byte
		dcExt []byte
		enclosures []byte
		err error
	)
	if item.DublinCoreExt != nil {
		dcExt, err = json.Marshal(item.DublinCoreExt)
		if err != nil {
			return fmt.Errorf("failed to marshal item.DublinCoreExt, %s", err)
		}
	}
	if item.Enclosures != nil {
		enclosures, err = json.Marshal(item.Enclosures)
		if err != nil {
			return fmt.Errorf("failed to marshal item.Enclosures, %s", err)
		}
	}
	if item.Authors != nil {
		authors, err = json.Marshal(item.Authors)
		if err != nil {
			return fmt.Errorf("failed to marshal item.Authors, %s", err)
		}
	}
	// NOTE: postPath doens't exist for harvested items, only for posted ones but SQL statement
	// is used for both.
	postPath := "" 
	stmt := SQLUpdateItem
	if _, err := db.Exec(stmt,
		item.Link, item.Title, item.Description, string(authors),
		string(enclosures), item.GUID, pubDate, string(dcExt),
		channel, status, updated, feedLabel, postPath); err != nil {
		return fmt.Errorf("%s\nstmt: %s", err, stmt)
	}
	return nil
}
