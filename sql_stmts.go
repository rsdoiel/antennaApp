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
FROM items
WHERE (description != '' OR title != '') AND status = 'published'
ORDER BY pubDate DESC, updated DESC;`

	// SQLSetStatusToReview
	SQLUpdateStatusToReview = `UPDATE items SET status = 'review';`

	// SQLSetStatusPublishedForRecentlyPublished will the will set status to "published" where
	// Items have a published date greater than or equal to the date provided.
	SQLSetStatusPublishedForRecentlyPublished = `UPDATE items SET status = 'published' WHERE pubDate >= date('now', '-21 days');`

	// SQLDeleteItemByLink removes an item in the items table with provided link
	SQLDeleteItemByLink = `DELETE FROM items WHERE link = ?`
)


