package antennaApp

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// URL represents a single URL entry in the sitemap.
type URL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

// URLSet represents the root of the sitemap XML.
type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	Xmlns   string   `xml:"xmlns,attr"`
	URLs    []URL    `xml:"url"`
}

// SitemapIndex represents the root of the sitemap index XML.
type SitemapIndex struct {
	XMLName  xml.Name `xml:"sitemapindex"`
	Xmlns    string   `xml:"xmlns,attr"`
	Sitemaps []struct {
		Loc string `xml:"loc"`
	} `xml:"sitemap"`
}

// Sitemap implements the antenna sitemap action.
func (app *AntennaApp) Sitemap(cfgName string, args []string) error {
	cfg := &AppConfig{}
	if err := cfg.LoadConfig(cfgName); err != nil {
		return err
	}

	if cfg.BaseURL == "" {
		if cfg.Port != 0 {
			cfg.BaseURL = fmt.Sprintf("http://%s:%d", cfg.Host, cfg.Port)
		} else {
			cfg.BaseURL = fmt.Sprintf("http://%s", cfg.Host)
		}
	}

	// Setup some sain defaults.
	cfg.ChunkSize = 100
	/* FIXME: need to come up with some good guesses for defaults here.
	cfg.DefaultFreq = "weekly"
	cfg.DefaultPri = "0.5"
	cfg.FreqRules = map[string]string{"blog/": "daily", "news/": "hourly"}
	cfg.PriRules = map[string]string{"": "1.0", "blog/": "0.8", "about/": "0.7"}
	*/
	return generateSitemaps(cfg)
}

// generateSitemaps iterates over all the collections pages and posts and
// generates the needed sitemaps
func generateSitemaps(cfg *AppConfig) error {
	sitemapFiles := []string{}
	for _, col := range cfg.Collections {
		if col.DbName == "" {
			fmt.Fprintf(os.Stderr, "%q is missing SQLite3 db name\n", col.File)
			continue
		}
		l, err := sitemap(cfg, col.DbName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%q (%q) sitemap error, %s\n", col.File, col.DbName)
		}
		sitemapFiles = append(sitemapFiles, l...)
	}

	// Create the sitemap index
	var sitemaps []struct {
		Loc string `xml:"loc"`
	}
	for _, file := range sitemapFiles {
		sitemaps = append(sitemaps, struct {
			Loc string `xml:"loc"`
		}{
			Loc: fmt.Sprintf("%s/%s", cfg.BaseURL, file),
		})
	}
	index := SitemapIndex{
		Xmlns:    "http://www.sitemaps.org/schemas/sitemap/0.9",
		Sitemaps: sitemaps,
	}

	// Marshal the index to XML
	indexData, err := xml.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal sitemap index: %s", err)
	}
	indexData = []byte(xml.Header + string(indexData))

	// Write the index to file
	if err := os.WriteFile("sitemap_index.xml", indexData, 0644); err != nil {
		return fmt.Errorf("failed to write sitemap_index.xml: %s", err)
	}

	log.Println("Sitemap files and index generated successfully!")
	return nil
}

// startsWith checks if a string starts with a prefix.
func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// sitemap generates the needed sitemaps for a given collection
func sitemap(cfg *AppConfig, dbName string) ([]string, error) {
	var (
		urls         []URL
		sitemapFiles []string
	)

	// Open the SQLite database
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return sitemapFiles, fmt.Errorf("failed to open database (%s): %s", dbName, err)
	}
	defer db.Close()

	// Process pages and posts for collection
	if pageUrls, err := processSitemapRows(cfg, dbName, db, SQLSitemapListPages); err != nil {
		return sitemapFiles, err
	} else {
		urls = append(urls, pageUrls...)
	}
	if postUrls, err := processSitemapRows(cfg, dbName, db, SQLSitemapListPosts); err != nil {
		return sitemapFiles, err
	} else {
		urls = append(urls, postUrls...)
	}

	if len(urls) == 0 {
		fmt.Fprintf(os.Stderr, "No pages or posts found in %s database", dbName)
		return nil, nil
	}

	// Split URLs into chunks
	for i := 0; i < len(urls); i += cfg.ChunkSize {
		end := i + cfg.ChunkSize
		if end > len(urls) {
			end = len(urls)
		}
		chunk := urls[i:end]

		// Create the URLSet for this chunk"
		urlSet := URLSet{
			Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
			URLs:  chunk,
		}

		// Marshal to XML
		xmlData, err := xml.MarshalIndent(urlSet, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to marshal XML for chunk %d: %v", i/cfg.ChunkSize+1, err)
			continue
		}
		xmlData = []byte(xml.Header + string(xmlData))

		// Write to file
		filename := fmt.Sprintf("sitemap_%d.xml", i/cfg.ChunkSize+1)
		if err := os.WriteFile(filename, xmlData, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write sitemap file %s: %v", filename, err)
			continue
		}
		sitemapFiles = append(sitemapFiles, filename)
		log.Printf("Generated %s with %d URLs", filename, len(chunk))
	}
	return sitemapFiles, nil
}

func processSitemapRows(cfg *AppConfig, dbName string, db *sql.DB, sqlStmt string) ([]URL, error) {
	var urls []URL
	// Query the pages table for in the collection.
	rows, err := db.Query(sqlStmt)
	if err != nil {
		return urls, fmt.Errorf("failed to query %s collection, %s", dbName, err)
	}
	defer rows.Close()

	for rows.Next() {
		var outputPath string
		var updated time.Time
		if err := rows.Scan(&outputPath, &updated); err != nil {
			fmt.Fprintf(os.Stderr, "failed to scan row (%s): %s", dbName, err)
			continue
		}

		// Determine changefreq and priority based on rules
		changeFreq := cfg.DefaultFreq
		priority := cfg.DefaultPri
		for prefix, freq := range cfg.FreqRules {
			if startsWith(outputPath, prefix) {
				changeFreq = freq
				break
			}
		}
		for prefix, pri := range cfg.PriRules {
			if startsWith(outputPath, prefix) {
				priority = pri
				break
			}
		}
		u := URL{
			Loc:     fmt.Sprintf("%s/%s", cfg.BaseURL, outputPath),
			LastMod: updated.Format("2006-01-02"),
		}
		if changeFreq != "" {
			u.ChangeFreq = changeFreq
		}
		if priority != "" {
			u.Priority = priority
		}
		urls = append(urls, u)
	}

	return urls, nil
}
