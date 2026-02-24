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
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	// 3rd Party Package
	"gopkg.in/yaml.v3"
	//"github.com/mmcdole/gofeed"
	ext "github.com/mmcdole/gofeed/extensions"
)

// AntennaApp configuration structure
type AppConfig struct {
	// Port holds the port number the preview web service will run on
	Port int `json:"port,omitempty" yaml:"port,omitempty"`

	// Host holds the IP address or machine name the preview service
	// will listen on. By default is is "localhost"
	Host string `json:"host,omitempty" yaml:"host,omitempty"`

	// Htdocs holds the path to directory that will recieve the generated content
	// It is also the directory used in the "preview" the static website.
	Htdocs string `json:"htdocs,omitempty" yaml:"htdocs,omitempty"`

	// UserAgent this holds a custom user agent string
	UserAgent string `json:"userAgent,omitempty" yaml:"userAgent,omitempty"`

	// BaseURL for the Antenna instance
	BaseURL string `json:"base_url,omitempty" yaml:"base_url,omitempty"`

	// Generator holds a YAML file that describes the HTML page structure.
	// This holds the default page generator description. Each collection can
	// use a custom one or the default one.
	Generator string `json:"generator,omitempty" yaml:"generator,omitempty"`

	// Collections holds a list of collections to curate
	Collections []*Collection `json:"collections,omitempty" yaml:"collections,omitempty"`

	// Sitemap settings, these should get sane defaults in the sitemap action
	ChunkSize   int
	DefaultFreq string
	DefaultPri  string
	FreqRules   map[string]string // outputPath prefix -> changefreq
	PriRules    map[string]string // outputPath prefix -> priority

}

// setupDatabase checks to see if anything needs to be setup (or fixed) for AntennaApp to run.
func setupDatabase(cName string, dbName string) error {
	// Check to see if we have an existing SQLite3 file
	if _, err := os.Stat(dbName); os.IsNotExist(err) {
		stmt := fmt.Sprintf(SQLCreateTables, cName, time.Now().Format("2006-01-02"))
		dsn := dbName
		db, err := sql.Open("sqlite", dsn)
		if err != nil {
			return err
		}
		defer db.Close()
		if db == nil {
			return fmt.Errorf("%s opened and returned nil", dbName)
		}
		_, err = db.Exec(stmt)
		if err != nil {
			return fmt.Errorf("%s\nstmt: %s", err, stmt)
		}
	}
	return nil
}

// AddCollection adds and saves a new collection to AppConfig
func (cfg *AppConfig) AddCollection(cfgName string, cName string) error {
	src, err := os.ReadFile(cName)
	if err != nil {
		return err
	}
	// Parse the Markdown collection of links and make sure it
	// makes sense.
	doc := &CommonMark{}
	if err := doc.Parse(src); err != nil {
		return err
	}
	// Now process the Markdown and pull out the list of links and label and description.
	if _, err := doc.GetLinks(); err != nil {
		return err
	}
	// If we get this parse we can update the configuration and include this link.
	if cfg.Collections == nil {
		cfg.Collections = []*Collection{}
	}
	collection := &Collection{
		File: cName,
	}
	if doc.FrontMatter != nil {
		collection.UpdateFrontMatter(doc.FrontMatter, cfg)
	}
	if collection.Title == "" {
		collection.Title = cName
	}
	// Add some default SQL filters
	if len(collection.Filters) == 0 {
		collection.Filters = []string{
			SQLUpdateStatusToReview,
			SQLSetStatusPublishedForRecentlyPublished,
		}
	}

	// Make sure we have a SQLite3 database name
	bName := filepath.Base(cName)
	xName := filepath.Ext(bName)
	if collection.DbName == "" {
		collection.DbName = strings.TrimSuffix(cName, xName) + ".db"
	}
	if collection.Link == "" {
		collection.Link = strings.TrimSuffix(cName, xName) + ".xml"
	}
	// Make sure we have a reasonable Generator YAML filename
	if collection.Generator == "" {
		collection.Generator = cfg.Generator
	}
	// Do I need to "add" the collection or replace the collection?
	appendCollection := true
	for i, col := range cfg.Collections {
		if filepath.Base(col.File) == filepath.Base(collection.File) {
			// update collection
			cfg.Collections[i] = collection
			appendCollection = false
		}
	}
	if appendCollection {
		// We're really adding the new collection
		cfg.Collections = append(cfg.Collections, collection)
	}
	// Create SQLite3 database for collection if not exists
	if err := setupDatabase(cName, collection.DbName); err != nil {
		return err
	}
	// Create generator YAML if not exists
	if _, err := os.Stat(collection.Generator); os.IsNotExist(err) {
		if err := InitPageGenerator(collection.Generator); err != nil {
			return err
		}
	}
	// Save all the updates
	if err := cfg.SaveConfig(cfgName); err != nil {
		return err
	}
	return nil
}

// DelCollection removes a collection from the configuration, saving it.
func (cfg *AppConfig) DelCollection(cfgName string, cName string) error {
	i := cfg.CollectionIndex(cName)
	if i > -1 {
		cfg.Collections = append(cfg.Collections[:i], cfg.Collections[i+1:]...)
	}
	// Save all the updates
	if err := cfg.SaveConfig(cfgName); err != nil {
		return err
	}
	return nil
}

// ListCollectionFiles returns a list of collections defined in the configuration
func (cfg *AppConfig) ListCollectionFiles(cfgName string) ([]string, error) {
	if cfg.Collections == nil {
		return nil, fmt.Errorf("not properly initialized")
	}
	names := []string{}
	for _, col := range cfg.Collections {
		if col.File != "" {
			names = append(names, col.File)
		}
	}
	return names, nil
}

// Collection describes the metadata about a collection of feeds.
// A collection can also be used to generate an RSS 2.0 feed of items
// harvested and related to the collection forming an aggregated item view
// of the collection of feeds.
//
// Some of the fields from the RSS 2.0 Channel can be set from the
// Markdown document's front matter.
//
// See https://cyber.harvard.edu/rss/rss.html#optionalChannelElements
type Collection struct {
	// Title of the collection
	Title string `json:"title,omitempty" yaml:"title,omitempty"`

	// Links holds the Link element used in the published RSS 2.0 output.
	Link string `json:"link,omitempty" yaml:"link,omitempty"`

	// Description holds the description of the collection
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// The language the collect is written in
	Language string `json:"language,omitempty" yaml:"language,omitempty"`

	// Copyright notice for content in the collection
	Copyright string `json:"copyright,omitempty" yaml:"copyright,omitempty"`

	// ManagingEditor holds an Email address for person responsible for editorial content
	// of the collection.
	ManagingEditor string `json:"managingEditor,omitempty" yaml:"managingEditor,omitempty"`

	// WebMaster holds an Email address for person responsible for technical issues relating to collection
	WebMaster string `json:"webMaster,omitempty" yaml:"webMaster,omitempty"`

	// PubDate holds the publication date for the content in the collection
	PubDate string `json:"pubDate,omitempty" yaml:"pubDate,omitempty"`

	// TTL is the time to live, the number of seconds to wait before trying a refresh
	TTL int `json:"ttl,omitempty" yaml:"ttl,omitempty"`

	// File holds the filepath to the Markdown document used to
	// define the collection.
	File string `json:"file,omitempty" yaml:"file,omitempty"`

	// Generator points to the YAML file to use when generating
	// a collection's HTML page.
	Generator string `json:"generator,omitempty" yaml:"generator,omitempty"`

	// Filters holds custom SQL that will be run against the Source to
	// determine which items to include and had off to the Generator.
	Filters []string `json:"filters,omitempty" yaml:"filters,omitempty"`

	// DbName holds the SQLite3 database filename
	DbName string `json:"dbName,omitempty" yaml:"dbName,omitempty"`
}

// Name returns the collection basename used for the collection
func (col *Collection) Name() string {
	return strings.TrimSuffix(col.File, ".md")
}

// Link represents a Markdown link with Label, URL, and optional Description.
type Link struct {
	// Label holds the text label that will be used when displaying the feed
	Label string `json:"label,omitempty" yaml:"label,omitempty"`
	// The URL holds the link text to the feed
	URL string `json:"url,omitempty" yaml:"url,omitempty"`
	// The optional description holds any description text associated with the link
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// LoadConfig process the AntennaApp YAML file and sets the
// attributes of the AntennaApp structure.
func (cfg *AppConfig) LoadConfig(cfgName string) error {
	src, err := os.ReadFile(cfgName)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(src, &cfg)
}

// SaveConfig save the current configuration of the AntennaApp
func (cfg *AppConfig) SaveConfig(cfgName string) error {
	src, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := os.WriteFile(cfgName, src, 0664); err != nil {
		return err
	}
	return nil
}

func (cfg *AppConfig) CollectionIndex(cName string) int {
	for i, col := range cfg.Collections {
		if filepath.Base(col.File) == filepath.Base(cName) {
			return i
		}
	}
	return -1
}

func (cfg *AppConfig) GetCollection(cName string) (*Collection, error) {
	i := cfg.CollectionIndex(cName)
	if i > -1 {
		return cfg.Collections[i], nil
	}
	return nil, fmt.Errorf("%s not in collection", cName)
}

func (collection *Collection) UpdateFrontMatter(frontMatter map[string]interface{}, cfg *AppConfig) error {
	rssFile := strings.TrimSuffix(collection.File, ".md") + ".xml"
	collection.Title = ""
	if title, ok := frontMatter["title"].(string); ok {
		collection.Title = title
	}
	if link, ok := frontMatter["link"].(string); ok {
		collection.Link = link
	} else if collection.Link == "" {
		if cfg.BaseURL != "" {
			collection.Link = fmt.Sprintf(`%s/%s`, cfg.BaseURL, rssFile)
		} else if cfg.Host != "" {
			collection.Link = fmt.Sprintf(`http://%s:%d/%s`, cfg.Host, cfg.Port, rssFile)
		} else {
			collection.Link = rssFile
		}

	}
	collection.Description = ""
	if description, ok := frontMatter["description"].(string); ok {
		collection.Description = description
	}
	collection.Language = ""
	if language, ok := frontMatter["language"].(string); ok {
		collection.Language = language
	}
	collection.Copyright = ""
	if copyright, ok := frontMatter["copyright"].(string); ok {
		collection.Copyright = copyright
	}
	collection.ManagingEditor = ""
	if managingEditor, ok := frontMatter["managingEditor"].(string); ok {
		collection.ManagingEditor = managingEditor
	}
	collection.WebMaster = ""
	if webMaster, ok := frontMatter["webMaster"].(string); ok {
		collection.WebMaster = webMaster
	}
	collection.PubDate = ""
	if pubDate, ok := frontMatter["pubDate"].(string); ok {
		collection.PubDate = pubDate
	}
	collection.TTL = 0
	if val, ok := frontMatter["ttl"].(int); ok {
		collection.TTL = val
	}
	collection.Generator = ""
	if generator, ok := frontMatter["generator"].(string); ok {
		collection.Generator = generator
	}
	collection.Filters = []string{}
	if filters, ok := frontMatter["filters"].([]string); ok {
		collection.Filters = append([]string{}, filters...)
	}
	collection.DbName = ""
	if dbName, ok := frontMatter["dbName"].(string); ok {
		collection.DbName = dbName
	}
	return nil
}

// Posts lists posts in a collection
func (cfg *AppConfig) Posts(cName string, options []string) error {
	collection, err := cfg.GetCollection(cName)
	if err != nil {
		return fmt.Errorf("%s, %s", cName, err)
	}
	dsn := collection.DbName
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	// NOTES: posts action supports three different SQL statements
	var (
		rows *sql.Rows
	)
	switch {
	case len(options) == 3:
		fromDate, toDate := options[1], options[2]
		rows, err = db.Query(SQLListDateRangePosts, fromDate, toDate)
		if err != nil {
			return fmt.Errorf("%s\n%s, %s", SQLListDateRangePosts, dsn, err)
		}
	case len(options) == 2:
		count, err := strconv.Atoi(options[1])
		if err != nil {
			return err
		}
		rows, err = db.Query(SQLListRecentPosts, count)
		if err != nil {
			return fmt.Errorf("%s\n%s, %s", SQLListRecentPosts, dsn, err)
		}
	default:
		rows, err = db.Query(SQLListPosts)
		if err != nil {
			return fmt.Errorf("%s\n%s, %s", SQLListPosts, dsn, err)
		}
	}
	if rows != nil {
		defer rows.Close()
	}

	i := 0
	for rows.Next() {
		var (
			link     string
			title    string
			pubDate  string
			postPath string
		)
		if err := rows.Scan(&link, &title, &pubDate, &postPath); err != nil {
			fmt.Fprintf(os.Stderr, "failed to read row, %s\n", err)
			continue
		}
		if strings.Contains(pubDate, "T") {
			parts := strings.SplitN(pubDate, "T", 2)
			pubDate = parts[0]
		}
		if i == 0 {
			fmt.Println("")
			i++
		}
		fmt.Printf("- [%s](%s), %s\n",
			title, postPath, pubDate)
	}
	if i == 0 {
		return fmt.Errorf("no published posts")
	}
	fmt.Println("")

	return nil
}

// updateItem will perform an "upsert" to insert or update the row
func updateItem(db *sql.DB, link string, title string, description string, authors string,
	enclosures []*Enclosure, guid string, pubDate string, dcExt *ext.DublinCoreExtension,
	channel, status string, updated string, label string, postPath string, sourceMarkdown string) error {
	enclosuresSrc, err := json.Marshal(enclosures)
	if err != nil {
		return nil
	}
	dcExtSrc, err := json.Marshal(dcExt)
	if err != nil {
		return nil
	}
	_, err = db.Exec(SQLUpdateItem, link, title, description, authors,
		enclosuresSrc, guid, pubDate, dcExtSrc,
		channel, status, updated, label, postPath, sourceMarkdown)
	if err != nil {
		return err
	}
	return nil
}

// publishPost publishes an item in the items table using link
func publishPost(db *sql.DB, postPath string, pubDate string, status string, updated string) error {
	_, err := db.Exec(SQLPublishPost, pubDate, status, postPath)
	if err != nil {
		return err
	}
	return nil
}

// PublishPost sets the publication date if not set in the front matter
// of the Markdown document, then sets the publication date and status to published
// in the items table.
func (cfg *AppConfig) PublishPost(cName string, fName string) error {
	collection, err := cfg.GetCollection(cName)
	if err != nil {
		return err
	}

	src, err := os.ReadFile(fName)
	if err != nil {
		return fmt.Errorf("failed to read %q, %s", fName, err)
	}
	doc := &CommonMark{}
	if err := doc.Parse(src); err != nil {
		return fmt.Errorf("failed to parse %q, %s", fName, err)
	}
	postPath := doc.GetAttributeString("postPath", "")
	if postPath == "" {
		return fmt.Errorf("missing postPath")
	}
	updateMarkdownDoc := false
	if doc.FrontMatter == nil {
		doc.FrontMatter = map[string]interface{}{}
	}
	today := time.Now().Format("2006-01-02")
	pubDate := doc.GetAttributeString("pubDate", "")
	if pubDate == "" {
		pubDate = doc.GetAttributeString("datePublished", today)
		doc.FrontMatter["datePublished"] = pubDate
		updateMarkdownDoc = true
	}
	if updateMarkdownDoc {
		doc.FrontMatter["dateModified"] = today
		if err = saveMarkdown(fName, doc); err != nil {
			return fmt.Errorf("unable to save %s, %s", fName, err)
		}
	}
	dsn := collection.DbName
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	status := "published"
	updated := time.Now().Format(time.RFC3339)
	return publishPost(db, postPath, pubDate, status, updated)
}

// Post adds a post to a collection
func (cfg *AppConfig) Post(cName string, fName string) error {
	collection, err := cfg.GetCollection(cName)
	if err != nil {
		return err
	}

	src, err := os.ReadFile(fName)
	if err != nil {
		return fmt.Errorf("failed to read %q, %s", fName, err)
	}
	doc := &CommonMark{}
	if err := doc.Parse(src); err != nil {
		return fmt.Errorf("failed to parse %q, %s", fName, err)
	}
	updateMarkdownDoc := false
	if doc.FrontMatter == nil {
		doc.FrontMatter = map[string]interface{}{}
	}
	postPath := doc.GetAttributeString("postPath", "")
	if postPath == "" {
		postPath = fName
		doc.FrontMatter["postPath"] = fName
		updateMarkdownDoc = true
	}
	today := time.Now().Format("2006-01-02")
	dateModified := doc.GetAttributeString("dateModified", "")
	if dateModified == "" {
		doc.FrontMatter["dateModified"] = today
		updateMarkdownDoc = true
	}
	if updateMarkdownDoc {
		if err = saveMarkdown(fName, doc); err != nil {
			return fmt.Errorf("unable to save %s, %s", fName, err)
		}
	}

	// NOTE: This is trusted content so I can support commonMarkDoc
	// processor extensions safely.
	if strings.Contains(doc.Text, "@include-text-block") {
		doc.Text = IncludeTextBlock(doc.Text)
	}
	if strings.Contains(doc.Text, "@include-code-block") {
		doc.Text = IncludeCodeBlock(doc.Text)
	}

	// Convert our document text to HTML
	innerHTML, err := doc.ToUnsafeHTML()
	if err != nil {
		return err
	}
	title := doc.GetAttributeString("title", "")

	authors, err := doc.GetPersons("author", false)
	if err != nil {
		return err
	}
	description := doc.GetAttributeString("description", "")
	if description == "" {
		description = doc.GetAttributeString("abstract", "")
	}
	link := doc.GetAttributeString("link", "")

	pubDate := doc.GetAttributeString("pubDate", "")
	if pubDate == "" {
		pubDate = doc.GetAttributeString("datePublished", "")
	}
	status := ""
	if postPath != "" && pubDate != "" {
		status = "published"
	} else if postPath != "" {
		status = "draft"
	}
	sourceMarkdown := doc.String()

	channel := doc.GetAttributeString("channel", collection.Link)
	guid := doc.GetAttributeString("guid", link)

	// When no description is provided the description is set with the body text
	if description == "" {
		description = innerHTML
	}
	if title == "" && description == "" {
		return fmt.Errorf("missing both title and description")
	}

	if postPath != "" {
		if link == "" {
			if cfg.BaseURL != "" {
				if strings.HasSuffix(postPath, ".md") {
					link = cfg.BaseURL + "/" + strings.TrimSuffix(postPath, ".md") + ".html"
				} else {
					link = cfg.BaseURL + "/" + postPath
				}
			} else {
				return fmt.Errorf("missing base_url in antenna YAML, could not form link using postPath %q", postPath)
			}
		}
		// Write out an HTML page to the postPath, if Markdown doc, normalize .html
		htmlName := filepath.Join(cfg.Htdocs, postPath)
		if strings.HasSuffix(htmlName, ".md") {
			htmlName = strings.TrimSuffix(htmlName, ".md") + ".html"
		}
		gen, err := NewGenerator(path.Base(os.Args[0]), cfg.BaseURL)
		if err != nil {
			return err
		}
		if err := gen.LoadConfig(collection.Generator); err != nil {
			return err
		}
		if err := gen.WriteHtmlPage(htmlName, link, postPath, pubDate, innerHTML); err != nil {
			return err
		}
	}
	// FIXME: Need to handle getting enclosures and publishing them to posts tree
	// NOTE: Insert/update item in collection
	// FIXME: need to populate the enclosures
	enclosures := []*Enclosure{}
	// FIXME: need to populate the Dublin Core extension
	dcExt := &ext.DublinCoreExtension{}
	updated := time.Now().Format(time.RFC3339)
	if dateModified != "" {
		d, err := time.Parse("2006-01-02", dateModified)
		if err != nil {
			return fmt.Errorf("failed to parse dateModified: %q, %s", dateModified, err)
		}
		updated = d.Format(time.RFC3339)
	}

	label := collection.Title
	authorsSrc := []byte{}
	if authors != nil {
		authorsSrc, err = json.Marshal(authors)
		if err != nil {
			return fmt.Errorf("failed to marshal author, %s", err)
		}
	}
	dsn := collection.DbName
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	return updateItem(db, link, title, description, fmt.Sprintf("%s", authorsSrc), enclosures, guid, pubDate, dcExt, channel, status, updated, label, postPath, sourceMarkdown)
}

// removePost removes an item from the items table using postPath
func removePost(db *sql.DB, postPath string) error {
	_, err := db.Exec(SQLDeletePost, postPath)
	if err != nil {
		return err
	}
	return nil
}

// Unpost deletes a post from the items table. Does not remove content on disk
func (cfg *AppConfig) Unpost(cName string, fName string) error {
	collection, err := cfg.GetCollection(cName)
	if err != nil {
		return err
	}
	dsn := collection.DbName
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	return removePost(db, fName)
}

// Unpage delete a page from the pages table (does not remove files from disk)
// The deletion will happen for pages with either an inputPath or outputPath matching
// the fName parameter.
func (cfg *AppConfig) Unpage(fName string) error {
	// NOTE: remove the page from pages table.
	collection, err := cfg.GetCollection("pages.md")
	if err != nil {
		return err
	}
	dsn := collection.DbName
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("failed to open %s, %s", dsn, err)
	}
	defer db.Close()

	if _, err := db.Exec(SQLDeletePageByPath, fName, fName); err != nil {
		return fmt.Errorf("%s, %s", dsn, err)
	}
	return nil
}

func (cfg *AppConfig) Page(fName string, oName string) error {
	src, err := os.ReadFile(fName)
	if err != nil {
		return err
	}
	doc := &CommonMark{}
	if err := doc.Parse(src); err != nil {
		return err
	}
	// NOTE: This is trusted content so I can support commonMarkDoc
	// processor extensions safely.
	if strings.Contains(doc.Text, "@include-text-block ") {
		doc.Text = IncludeTextBlock(doc.Text)
	}
	if strings.Contains(doc.Text, "@include-code-block ") {
		doc.Text = IncludeCodeBlock(doc.Text)
	}

	// Convert our document text to HTML
	// NOTE: Pages are allowed to have "unsafe" embedded HTML because they are
	// not reading from a feed, they are being read from your file system.
	innerHTML, err := doc.ToUnsafeHTML()
	if err != nil {
		return err
	}
	postPath := doc.GetAttributeString("postPath", fName)
	htmlName := filepath.Join(cfg.Htdocs, postPath)
	if oName != "" {
		htmlName = filepath.Join(cfg.Htdocs, oName)
	}
	if strings.HasSuffix(htmlName, ".md") {
		htmlName = strings.TrimSuffix(htmlName, ".md") + ".html"
	}
	dName := filepath.Dir(htmlName)
	if _, err := os.Stat(dName); err != nil {
		if err := os.MkdirAll(dName, 0775); err != nil {
			return err
		}
	}
	gen, err := NewGenerator(path.Base(os.Args[0]), cfg.BaseURL)
	if err != nil {
		return err
	}
	if err := gen.LoadConfig(cfg.Generator); err != nil {
		return err
	}
	if err := gen.WriteHtmlPage(htmlName, "", postPath, "", innerHTML); err != nil {
		return err
	}
	// NOTE: I need to add the page to pages.db
	// NOTE: remove the page from pages table.
	// NOTE: remove the page from pages table.
	collection, err := cfg.GetCollection("pages.md")
	if err != nil {
		return err
	}
	dsn := collection.DbName
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("failed to open %s, %s", dsn, err)
	}
	defer db.Close()
	timestamp := time.Now().Format(time.RFC3339)
	if _, err := db.Exec(SQLUpdatePage, fName, oName, timestamp); err != nil {
		return fmt.Errorf("%s, %s", dsn, err)
	}
	return nil
}

// GetPages returns a list of page maps for a collection
func (cfg *AppConfig) GetPages() ([]map[string]string, error) {
	collection, err := cfg.GetCollection("pages.md")
	if err != nil {
		return nil, err
	}
	dsn := collection.DbName
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var (
		rows *sql.Rows
	)
	rows, err = db.Query(SQLListPages)
	if err != nil {
		return nil, fmt.Errorf("%s\n%s, %s", SQLListItems, dsn, err)
	}
	if rows != nil {
		defer rows.Close()
	}

	i := 0
	pages := []map[string]string{}
	for rows.Next() {
		var (
			inputPath  string
			outputPath string
			updated    string
		)
		if err = rows.Scan(&inputPath, &outputPath, &updated); err != nil {
			displayErrorStatus("failed to read row (%d), %s\n", i, err)
			continue
		}
		if i == 0 {
			i++
		}
		page := map[string]string{
			"inputPath":  inputPath,
			"outputPath": outputPath,
			"updated":    updated,
		}
		pages = append(pages, page)
	}
	if i == 0 {
		return nil, fmt.Errorf("no pages found")
	}
	return pages, nil
}

// Pages diplays a page information to standard output
func (cfg *AppConfig) Pages() error {
	// NOTE: remove the page from pages table.
	collection, err := cfg.GetCollection("pages.md")
	if err != nil {
		return err
	}
	dsn := collection.DbName
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("failed to open %s, %s", dsn, err)
	}
	defer db.Close()

	rows, err := db.Query(SQLDisplayPage)
	if err != nil {
		return fmt.Errorf("%s, %s", dsn, err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			iName   string
			oName   string
			updated string
		)
		if err := rows.Scan(&iName, &oName, &updated); err != nil {
			fmt.Fprintf(os.Stderr, "failed to read row, %s\n", err)
			continue
		}
		fmt.Printf("%s\t%s\t%s\n", iName, oName, updated)
	}
	return nil
}
