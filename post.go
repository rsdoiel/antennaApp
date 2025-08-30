package antennaApp

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	 // 3rd Party
	_ "github.com/glebarez/go-sqlite"
	//"github.com/mmcdole/gofeed"
	ext "github.com/mmcdole/gofeed/extensions"
)

// Post will add a CommonMark document as a feed item and if the postPath and link
// are provided it will convert the CommonMark document to HTML and save it in the postPath.
func (app *AntennaApp) Post(cfgName string, args []string) error {
	fmt.Fprintf(os.Stderr, "DEBUG cfgName: %q, args: %+v\n", cfgName, args)

	if len(args) != 2 {
		return fmt.Errorf("expected a collection name and filepath in the collection")
	}
	cfg := &AppConfig{}
	if err := cfg.LoadConfig(cfgName); err != nil {
		return err
	}
	cName, fName := strings.TrimSpace(args[0]), strings.TrimSpace(args[1])
	fmt.Fprintf(os.Stderr, "DEBUG cName: %q, fName: %q\n", cName, fName)
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

	src, err := os.ReadFile(fName)
	if err != nil {
		return err
	}
	doc := &CommonMark{}
	if err := doc.Parse(src); err != nil {
		return err
	}
	// Convert our document text to HTML
	innerHTML, err := doc.ToHTML()
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "DEBUG innerHTML\n%s\n\n", innerHTML)
	title := doc.GetAttributeString("title", "")
	authors := doc.GetAttributeString("author", "")
	description := doc.GetAttributeString("description", "")
	link := doc.GetAttributeString("link", "")
	postPath := doc.GetAttributeString("postPath", "")
	pubDate := doc.GetAttributeString("pubDate", "")
	draft := doc.GetAttributeBool("draft", false)
	channel := doc.GetAttributeString("channel", collection.Link)
	guid := doc.GetAttributeString("guid", link)
	// FIXME: Need to handle getting enclosures and publishing them to posts tree
	status := "review"
	if draft || pubDate == ""{
		return fmt.Errorf("%s is set to draft or is missing pubDate", fName)
	} else {
		status = "published" 
	}

	// When no description is provided the description is set with the body text
	if description == "" {
		description = innerHTML
	}
	if title == "" && description == "" {
		return fmt.Errorf("missing both title and description")
	}
	if postPath != "" {
		if link == "" {
			return fmt.Errorf("missing link to used with postPath %q", postPath)
		}
		// Write out an HTML page to the postPath
		htmlName := filepath.Join(cfg.Htdocs, postPath)
		fmt.Fprintf(os.Stderr, "DEBUG htmlName: %q\n", htmlName)
		dName := filepath.Dir(htmlName)
		fmt.Fprintf(os.Stderr, "DEBUG dName: %q\n", dName)
		if _, err := os.Stat(dName); err != nil {
			if err := os.MkdirAll(dName, 0775); err != nil {
				return err
			}
		}
		fmt.Fprintf(os.Stderr, "DEBUG htmlName: %q\n", htmlName)
		gen, err := NewGenerator(app.appName)
		if err != nil {
			return err
		}
		collection, err := cfg.GetCollection(cName)
		if err != nil {
			return err
		}
		if err := gen.LoadConfig(collection.Generator); err != nil {
			return err
		}
		if err := gen.WriteHtmlPage(htmlName, link, pubDate, innerHTML); err != nil {
			return err
		} 
	}
	// NOTE: Insert/update item in collection
	// FIXME: need to populate the enclosures
	enclosures := []*Enclosure{}
	// FIXME: need to populate the Dublin Core extension
	dcExt := &ext.DublinCoreExtension{}
	updated := time.Now().Format(time.RFC3339)
	label := collection.Title
	return updateItem(db, link, title, description, authors,
		enclosures, guid, pubDate, dcExt, channel, status, updated, label)
}

func (app *AntennaApp) Unpost(cfgName string, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("expected a collection name and url in the collection")
	}
	cfg := &AppConfig{}
	if err := cfg.LoadConfig(cfgName); err != nil {
		return err
	}
	cName, link := strings.TrimSpace(args[0]), strings.TrimSpace(args[1])
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
	return removeItem(db, link)
}

// updateItem will perform an "upsert" to insert or update the row
func updateItem(db *sql.DB, link string, title string, description string, authors string,
	enclosures []*Enclosure, guid string, pubDate string, dcExt *ext.DublinCoreExtension,
	channel, status string, updated string, label string) error {
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
		channel, status, updated, label)
	if err != nil {
		return err
	}
	return nil
}

func removeItem(db *sql.DB, link string) error {
	_, err := db.Exec(SQLDeleteItemByLink, link)
	if err != nil {
		return err
	}
	return nil
}
