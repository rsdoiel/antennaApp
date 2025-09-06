package antennaApp

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	// 3rd Party
	_ "github.com/glebarez/go-sqlite"
)

var (
	DefaultGeneratorYaml = `###
## Example YAML used to custom a collection's Generator YAML.
##
## Use this CSS file for the CSS link in the head element. To
## pull an other CSS, use a CSS @import statement. (uncomment line
## and update location)
#css: css/site.css
## You list of ES6 modules goes here, (uncomment lines, update locations)
#modules:
#  - modules/date_filter.js
##
## The next set of elements let you include HTML in the page
## area decribed.
##
header: |
  <header><!-- your custom header inner HTML goes here --></header>

nav: |
  <nav><!-- your custom nav innert HTML goes here --></nav>

topContent: |
  <!-- your custom HTML content before the section element goes here -->

bottomContent: |
  <!-- your custom HTML content after the section element goes here -->

footer: |
  <footer><!-- your custom footer innert HTML goes here --></footer>
`
)

func (app *AntennaApp) Add(cfgName string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing collection name")
	}
	// create a cfg object
	cfg := &AppConfig{}
	// Load configuration
	if err := cfg.LoadConfig(cfgName); err != nil {
		return err
	}
	// Get the collection name
	cName := args[0]
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
	// Make sure we have a reasonable Generator YAML filename
	if collection.Generator == "" {
		collection.Generator = strings.TrimSuffix(cName, xName) + ".yaml"
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
		if err := os.WriteFile(collection.Generator, []byte(DefaultGeneratorYaml), 0664); err != nil {
			return err
		}
	}
	// Save all the updates
	if err := cfg.SaveConfig(cfgName); err != nil {
		return err
	}
	return nil
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

// Del removes one or more collections from your Antenna instance.
// NOTE: It does not remove generated files, e.g. SQLite3 database,
// YAML, HTML or RSS files.
func (app *AntennaApp) Del(cfgName string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing collection name")
	}

	// create a cfg object
	cfg := &AppConfig{}
	// Load configuration
	if err := cfg.LoadConfig(cfgName); err != nil {
		return err
	}
	for _, cName := range args {
		i := cfg.CollectionIndex(cName)
		if i > -1 {
			cfg.Collections = append(cfg.Collections[:i], cfg.Collections[i+1:]...)	
		}
	}
	// Save all the updates
	if err := cfg.SaveConfig(cfgName); err != nil {
		return err
	}
	return nil
}
