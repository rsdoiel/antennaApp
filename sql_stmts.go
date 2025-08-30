package antennaApp

var (
	// SQLCreateTables provides the statements that are use to create our tables
	// It has two percent s, first is feed list name, second is datetime scheme
	// was generated.
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
	title TEXT,
	description TEXT,
	authors JSON,
	enclosures JSON DEFAULT '',
	guid TEXT,
	pubDate DATETIME,
	dcExt JSON,
	channel TEXT,
	status TEXT DEFAULT '',
	label TEXT DEFAULT '',
	updated DATETIME
);
`
	// SQLResetChannels clear the channels talbe
	SQLResetChannels = `DELETE FROM channels;`

	// Update the channels in the skimmer file
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

	// Update a feed item in the items table
	SQLUpdateItem = `INSERT INTO items (
	link, title, description, authors,
	enclosures, guid, pubDate, dcExt,
	channel, status, updated, label
) VALUES (
	?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8, ?9, ?10, ?11, ?12
) ON CONFLICT (link) DO
  UPDATE SET 
  	title = ?2, description = ?3, authors = ?4,
	enclosures = ?5, guid = ?6, pubDate = ?7, dcExt = ?8,
	channel = ?9, status = ?10, updated = ?11, label = ?12;`

	// SQLItemCount returns a list of items in the items table
	SQLItemCount = `SELECT COUNT(*) FROM items;`

	// SQLDisplayItems returns a list of items in decending chronological order.
	SQLDisplayItems = `SELECT 
  link, title, description, authors,
  enclosures, guid, pubDate, dcExt,
  channel, status, updated, label
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


