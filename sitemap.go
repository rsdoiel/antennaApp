package antennaApp

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// URL represents a single URL entry in the sitemap.
type URL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
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
	return generateSitemaps(cfg, os.Stderr)
}

// Sitemap implements the antenna sitemap action for TUI.
func (cfg *AppConfig) Sitemap() error {
	if cfg.BaseURL == "" {
		if cfg.Port != 0 {
			cfg.BaseURL = fmt.Sprintf("http://%s:%d", cfg.Host, cfg.Port)
		} else {
			cfg.BaseURL = fmt.Sprintf("http://%s", cfg.Host)
		}
	}

	// Setup some sain defaults.
	cfg.ChunkSize = 100
	return generateSitemaps(cfg, os.Stderr)
}

/** generateSitemaps writes the site's sitemap files and index into the
 * document root (cfg.Htdocs).
 *
 * What is listed: the pages of the site, which come only from the pages
 * collection (pages.md), and the local posts, which are the items with a
 * postPath in any collection. Harvested feed items are not pages of this site
 * and are never listed. A collection whose database has no pages table
 * (an older database) has no pages, which is not a failure; its posts still
 * count.
 *
 * Every collection is processed even when one fails; the failures are then
 * reported with the class of the first (DR-0003). When there is nothing to
 * list the function warns on eout, writes no file, and succeeds: an
 * aggregation-only site has no sitemap, and an empty one is not valid per the
 * sitemaps.org schema.
 *
 * Parameters:
 *   cfg  (*AppConfig) — the loaded antenna.yaml, with BaseURL and ChunkSize set
 *   eout (io.Writer)  — where warnings and progress are written
 *
 * Returns:
 *   error — nil, or the failures of the collections that could not be read,
 *           or an error writing a sitemap file
 *
 * Example:
 *   err := generateSitemaps(cfg, os.Stderr)
 */
func generateSitemaps(cfg *AppConfig, eout io.Writer) error {
	tally := &failureTally{}
	urls := []URL{}
	seen := map[string]bool{}
	for _, col := range cfg.Collections {
		if col.DbName == "" {
			fmt.Fprintf(eout, "%q is missing SQLite3 db name\n", col.File)
			continue
		}
		colURLs, err := collectionURLs(cfg, col)
		if err != nil {
			fmt.Fprintf(eout, "%q (%q) sitemap error, %s\n", col.File, col.DbName, err)
			tally.add(fmt.Errorf("%s: %w", col.File, err))
		} else {
			tally.add(nil)
		}
		for _, u := range colURLs {
			if !seen[u.Loc] {
				seen[u.Loc] = true
				urls = append(urls, u)
			}
		}
	}
	if len(urls) == 0 {
		fmt.Fprintln(eout, "warning: no pages or posts found in any collection, no sitemap written")
		return tally.err("collections")
	}

	// Chunk across the whole run so file names are unique and each file is
	// named once in the index.
	chunk := cfg.ChunkSize
	if chunk <= 0 {
		chunk = 100
	}
	var sitemaps []struct {
		Loc string `xml:"loc"`
	}
	for i := 0; i < len(urls); i += chunk {
		end := i + chunk
		if end > len(urls) {
			end = len(urls)
		}
		urlSet := URLSet{Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9", URLs: urls[i:end]}
		xmlData, err := xml.MarshalIndent(urlSet, "", "  ")
		if err != nil {
			return internalErrorf("failed to marshal sitemap chunk %d: %w", i/chunk+1, err)
		}
		name := fmt.Sprintf("sitemap_%d.xml", i/chunk+1)
		if err := os.WriteFile(filepath.Join(cfg.Htdocs, name), []byte(xml.Header+string(xmlData)), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", name, asCreate(err))
		}
		sitemaps = append(sitemaps, struct {
			Loc string `xml:"loc"`
		}{Loc: fmt.Sprintf("%s/%s", cfg.BaseURL, name)})
		fmt.Fprintf(eout, "Generated %s with %d URLs\n", name, end-i)
	}

	index := SitemapIndex{Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9", Sitemaps: sitemaps}
	indexData, err := xml.MarshalIndent(index, "", "  ")
	if err != nil {
		return internalErrorf("failed to marshal sitemap index: %w", err)
	}
	if err := os.WriteFile(filepath.Join(cfg.Htdocs, "sitemap_index.xml"), []byte(xml.Header+string(indexData)), 0644); err != nil {
		return fmt.Errorf("failed to write sitemap_index.xml: %w", asCreate(err))
	}
	fmt.Fprintln(eout, "Sitemap files and index generated successfully!")
	return tally.err("collections")
}

// startsWith checks if a string starts with a prefix.
func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// isPagesCollection reports whether col is the pages collection, the one that
// holds the site's pages table (antenna page adds to pages.md).
func isPagesCollection(col *Collection) bool {
	base := filepath.Base(col.File)
	return strings.TrimSuffix(base, filepath.Ext(base)) == "pages"
}

// collectionURLs returns the sitemap URLs for one collection: its pages (the
// pages collection only) and its local posts. A database file that does not
// exist is reported as an error and is never created by looking at it.
func collectionURLs(cfg *AppConfig, col *Collection) ([]URL, error) {
	if _, err := os.Stat(col.DbName); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite3", col.DbName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database (%s): %w", col.DbName, err)
	}
	defer db.Close()
	urls := []URL{}
	if isPagesCollection(col) {
		var n int
		if err := db.QueryRow(SQLHasPagesTable).Scan(&n); err != nil {
			return nil, fmt.Errorf("failed to inspect %s: %w", col.DbName, err)
		}
		if n > 0 {
			pageURLs, err := processSitemapRows(cfg, col.DbName, db, SQLSitemapListPages)
			if err != nil {
				return urls, err
			}
			urls = append(urls, pageURLs...)
		}
	}
	postURLs, err := processSitemapRows(cfg, col.DbName, db, SQLSitemapListPosts)
	if err != nil {
		return urls, err
	}
	return append(urls, postURLs...), nil
}

func processSitemapRows(cfg *AppConfig, dbName string, db *sql.DB, sqlStmt string) ([]URL, error) {
	var urls []URL
	// Query the pages table for in the collection.
	rows, err := db.Query(sqlStmt)
	if err != nil {
		return urls, fmt.Errorf("failed to query %s collection, %w", dbName, err)
	}
	defer rows.Close()

	tally := &failureTally{}
	for rows.Next() {
		var outputPath string
		var updated sql.NullTime
		if err := rows.Scan(&outputPath, &updated); err != nil {
			fmt.Fprintf(os.Stderr, "failed to scan row (%s): %s\n", dbName, err)
			tally.add(dataErrorf("failed to scan row (%s): %w", dbName, err))
			continue
		}
		tally.add(nil)

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
		}
		// A row with no updated date is still listed, without lastmod.
		if updated.Valid {
			u.LastMod = updated.Time.Format("2006-01-02")
		}
		if changeFreq != "" {
			u.ChangeFreq = changeFreq
		}
		if priority != "" {
			u.Priority = priority
		}
		urls = append(urls, u)
	}

	return urls, tally.err("rows")
}
