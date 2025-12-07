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

var (
	// SQLCreateTables provides the statements that are use to create our tables
	// It has two percent s, first is feed list name, second is datetime scheme
	// was generated.
	//
	// See <https://source.scripting.com> when I am ready to support more the
	// source namespace.
	SQLCreateTables = `-- This is the scheme used for %s's SQLite 3 database
-- %s
CREATE TABLE IF NOT EXISTS channels (
	link PRIMARY KEY,
	title TEXT,
	description TEXT,
	feed_link TEXT,
	links JSON,
	updated DATETIME,
	published DATETIME,
	authors JSON,
	language TEXT,
	copyright TEXT,
	generator TEXT,
	categories JSON,
	feed_type TEXT,
	feed_version TEXT
);

CREATE TABLE IF NOT EXISTS items (
	link PRIMARY KEY,
	postPath TEXT DEFAULT '',
	title TEXT,
	description TEXT,
	authors JSON,
	enclosures JSON DEFAULT '',
	guid TEXT,
	pubDate DATETIME,
	dcExt JSON,
	channel TEXT,
	sourceMarkdown TEXT DEFUALT '',
	status TEXT DEFAULT '',
	label TEXT DEFAULT '',
	updated DATETIME
);

CREATE TABLE IF NOT EXISTS pages (
  inputPath PRIMARY KEY,
  outputPath TEXT DEFAULT '',
  updated DATETIME
);
`

	// SQLResetChannels clear the channels table
	SQLResetChannels = `DELETE FROM channels;`

	// Update the channels table
	SQLUpdateChannel = `REPLACE INTO channels (
link, title, description, feed_link, links,
updated, published, 
authors, language, copyright, generator,
categories, feed_type, feed_version
) VALUES (
?, ?, ?, ?, ?, 
?, ?,
?, ?, ?, ?,
?, ?, ?
);`
	// Display the channels in the table
	SQLDisplayChannels = `SELECT 
link, title, description, feed_link, links,
updated, published, 
authors, language, copyright, generator,
categories, feed_type, feed_version
FROM channels
ORDER by title, link
;`

	// Update a feed item in the items table
	SQLUpdateItem = `INSERT INTO items (
	link, title, description, authors,
	enclosures, guid, pubDate, dcExt,
	channel, status, updated, label, postPath, sourceMarkdown
) VALUES (
	?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8, ?9, ?10, ?11, ?12, ?13, ?14
) ON CONFLICT (link) DO
  UPDATE SET 
  	title = ?2, description = ?3, authors = ?4,
	enclosures = ?5, guid = ?6, pubDate = ?7, dcExt = ?8,
	channel = ?9, status = ?10, updated = ?11, label = ?12,
	postPath = ?13, sourceMarkdown = ?14;`

	// SQLItemCount returns a list of items in the items table
	SQLItemCount = `SELECT COUNT(*) FROM items;`

	// SQLDisplayItems returns a list of items in decending chronological order.
	SQLDisplayItems = `SELECT 
  link, title, description, authors,
  enclosures, guid, pubDate, dcExt,
  channel, status, updated, label,
  postPath, ifnull(sourceMarkdown, '') as sourceMarkdown
FROM items WHERE (description != '' OR title = '') AND status = 'published'
ORDER BY pubDate DESC, updated DESC;`

	// SQLRssPosts generate an RSS feed for all posts
	SQLRssPosts = `SELECT link, title, description, authors,
  enclosures, guid, pubDate, dcExt,
  channel, status, updated, label,
  postPath, ifnull(sourceMarkdown, '') as sourceMarkdown
FROM items
WHERE (pubDate IS NOT NULL) AND 
   (pubDate != "") AND
   (postPath != "")
ORDER BY pubDate DESC;`

	// SQLRssRecentPosts generate an RSS feed for recent posts
	SQLRssRecentPosts = `SELECT link, title, description, authors,
  enclosures, guid, pubDate, dcExt,
  channel, status, updated, label,
  postPath, ifnull(sourceMarkdown, '') as sourceMarkdown
FROM items
WHERE (pubDate IS NOT NULL) AND 
   (pubDate != '') AND
   (postPath != '')
ORDER BY pubDate DESC
LIMIT %d;`

	// SQLRssDateRangePosts generate an RSS feed for recent posts
	SQLRssDateRangePosts = `SELECT link, title, description, authors,
  enclosures, guid, pubDate, dcExt,
  channel, status, updated, label,
  postPath, ifnull(sourceMarkdown, '') as sourceMarkdown
FROM items
WHERE (pubDate IS NOT NULL) AND 
   (postPath != '') AND
   (pubDate >= '%s') AND
   (pubDate <= '%s')
ORDER BY pubDate DESC;`

	// SQLListPosts will list all published posts with a postPath by their descending pubDate
	SQLListPosts = `SELECT link, title, pubDate, postPath
FROM items
WHERE (pubDate IS NOT NULL) AND 
   (pubDate != '') AND
   (postPath != '')
ORDER BY pubDate DESC;`

	// SQLListRecentPosts will list recent published posts with a postPath by their descending pubDate
	SQLListRecentPosts = `SELECT link, title, pubDate, postPath
FROM items
WHERE (pubDate IS NOT NULL) AND 
   (pubDate != '') AND
   (postPath != '')
ORDER BY pubDate DESC
LIMIT ?;`

	// SQLListDateRangePosts will list all published posts with a postPath by their descending pubDate
	SQLListDateRangePosts = `SELECT link, title, pubDate, postPath
FROM items
WHERE (pubDate IS NOT NULL) AND 
   (postPath != '') AND
   (pubDate >= ?) AND
   (pubDate <= ?)
ORDER BY pubDate DESC;`

	// SQLSetStatusToReview
	SQLUpdateStatusToReview = `UPDATE items SET status = 'review';`

	// SQLSetStatusPublishedForRecentlyPublished will the will set status to "published" where
	// Items have a published date greater than or equal to the date provided.
	SQLSetStatusPublishedForRecentlyPublished = `UPDATE items SET status = 'published' WHERE pubDate >= date('now', '-21 days');`

	// SQLDeleteItemByLinkOrPostPath removes an item in the items table with provided link
	SQLDeleteItemByLinkOrPostPath = `DELETE FROM items WHERE link = ? OR postPath = ?`

	// Update a feed item in the items table
	SQLUpdatePage = `INSERT INTO pages (
	inputPath, outputPath, updated
) VALUES (
	?1, ?2, ?3
) ON CONFLICT (inputPath) DO
  UPDATE SET 
  	outputPath = ?2, updated = ?3;`

	// SQLCountPage returns a list of items in the items table
	SQLCountPage = `SELECT COUNT(*) FROM pages;`

	// SQLDisplayPage returns a list of pages in the pages table
	SQLDisplayPage = `SELECT inputPath, outputPath, updated
FROM pages
ORDER BY updated desc
;`

	// SQLSitemapListPages returns a list of pages by outputPath
	SQLSitemapListPages = `SELECT outputPath, updated
FROM pages
ORDER BY outputPath
;`

	// SQLListPosts will list all posts by post path
	SQLSitemapListPosts = `SELECT 
	 CASE
        WHEN postPath LIKE '%.md' THEN REPLACE(postPath, '.md', '.html')
        ELSE postPath
    END as outputPath,
	updated
FROM items
WHERE pubDate IS NOT NULL AND pubDate != ""
ORDER BY postPath;`

	// SQLDeletePageByPath removes a page by either input or output paths.
	SQLDeletePageByPath = `DELETE
FROM pages
WHERE inputPath = ?1 or outputPath = ?2
;`
)
