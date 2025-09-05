package antennaApp

import (
	"database/sql"
    "encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	// 3rd Party Packages
	"gopkg.in/yaml.v3"
    "github.com/mmcdole/gofeed"
)

// Enclosure holds the data for RSS enclusure support
type Enclosure struct {
	Url    string `json:"url,omitempty" yaml:"url,omitempty"`
	Length string    `json:"length,omitempty" yaml:"length,omitempty"`
	Type   string `json:"type,omitempty" yaml:"type,omitempty"`
}

// Generator supports the generation of HTML pages from a YAML configuration
type Generator struct {
	// AppName holds the name of application running the generator
	AppName string `json:"appName,omitempty" yaml:"appName,omitempty"`

	// Version holds the version of the genliction
	// used when generating the "generator" metadata
	Version string `json:"version,omitempty" yaml:"version,omitempty"`

	// DbName holds the path to the SQLite3 database
	DBName string `json:"dbName,omitempty" yaml:"dbName,omitempty"`

	// Title if this is set the title will be included
	// when generating the markdown of saved items
	Title string `json:"title,omitempty" yaml:"title,omitempty"`

	// Description, included as metadata in head element
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// CSS is the path to a CSS file
	CSS string `json:"css,omitempty" yaml:"css,omitempty"`

	// Modules is a list for ES6 diles
	Modules []string `json:"modules,omitempty" yaml:"modules,omitempty"`

	// Header hold the HTML markdup of the Header element. If not included
	// then it will be generated using the Title and timestamp
	Header string `json:"header,omitempty" yaml:"header,omitempty"`

	// Nav holds the HTML markup for navigation
	Nav string `json:"nav,omitempty" yaml:"nav,omitempty"`

	// TopContent holds HTML that comes before the selecton element
	// containing articles
	TopContent string `json:"topContent,omitempty" yaml:"topContent,omitempty"`

	// BottomContent holds HTML that comes before the selecton element
	// containing articles
	BottomContent string `json:"bottomContent,omitempty" yaml:"bottomContent,omitempty"`

	// Footer holds the HTML markup for the footer
	Footer string `json:"footer,omitempty" yaml:"footer,omitempty"`

	out  io.Writer
	eout io.Writer
}

// NewGenerator initialized a new Generator struct
func NewGenerator(appName string) (*Generator, error) {
	gen := new(Generator)
	gen.AppName = appName
	gen.Version = Version
	gen.out = os.Stdout
	gen.eout = os.Stderr
	return gen, nil
}

func (gen *Generator) WriteItem(out io.Writer, link string, title string, description string, authors []*gofeed.Person,
	enclosures []*Enclosure, guid string, pubDate string, dcExt string,
	channel string, status string, updated string, label string) error {
	// Setup expressing update time.
	pressTime := pubDate
	if len(pressTime) > 10 {
		pressTime = pressTime[0:10]
	}
	if updated != "" {
		if len(updated) > 10 {
			updated = updated[0:10]
		}
		pressTime += ", updated: " + updated
	}

	// Setup the Title
	if title == "" {
		title = fmt.Sprintf("<h1>@%s</h1>\n\n(date: %s, from: <a href=%q>%s</a>)", label, pressTime, link, label)
	} else {
		title = fmt.Sprintf("<h1>%s</h1>\n\n(date: %s, from: <a href=%q>%s</a>)", title, pressTime, link, link)
	}

	fmt.Fprintf(out, `
    <article data-published=%q data-link=%q>
      %s
      <p>
      %s
      <address>
        <a href=%q>%s</a>
      </address>
    </article>
`, pubDate, link, title, description, link, link)
	return nil
}

func (gen *Generator) writeHeadElement(out io.Writer) {
	fmt.Fprintln(out, "<head>")
	defer fmt.Fprintln(out, "</head>")
	// Write out charset
	fmt.Fprintln(out, "  <meta charset=\"UTF-8\" />")
	// Write title
	if gen.Title != "" {
		fmt.Fprintf(out, "  <title>%s</title>\n", gen.Title)
	}
	fmt.Fprintln(out, "  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\" />")
	if gen.CSS != "" {
		fmt.Fprintf(out, "  <link rel=\"stylesheet\" href=\"%s\" />\n", gen.CSS)
	}
	if gen.Modules != nil {
		for _, module := range gen.Modules {
			fmt.Fprintf(out, "  <script type=\"module\" src=\"%s\"></script>\n", module)
		}
	}
	// Get the current date
	currentDate := time.Now()

	// Format the date
	formattedDate := currentDate.Format(time.RFC3339)
	fmt.Fprintf(out, `  <meta name="generator" content="%s/%s">
  <meta name="date" content="%s">
`, gen.AppName, gen.Version, formattedDate)
}

// indentText splits  the string into lines, then prefixes the number of
// spaces specified to each line returning updated text
func indentText(src string, spaces int) string {
	lines := strings.Split(src, "\n")
	return strings.Join(lines, "\n"+strings.Repeat(" ", spaces))
}

// WriteHTML writes aggregated items into an HTML page from the contents of the database
func (gen *Generator) WriteHTML(out io.Writer, db *sql.DB, cfgName string, collection *Collection) error {
	// Create the outer elements of a page.
	fmt.Fprintln(out, `<!doctype html>
<html lang="en-US">`)
	defer fmt.Fprintln(out, "</html>")
	// Setup the metadata in the head element
	gen.writeHeadElement(out)
	// Setup body element
	fmt.Fprintln(out, "<body>")
	defer fmt.Fprintln(out, "</body>")
	// Setup header element
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	if gen.Header != "" {
		fmt.Fprintf(out, "  <header>\n    %s\n  </header>\n", indentText(strings.TrimSpace(gen.Header), 4))
	} else if gen.Title != "" {
		fmt.Fprintf(out, `  <header>
    <h1>%s</h1>

    (date: %s)

  </header>
`, gen.Title, timestamp)
	} else {
		fmt.Fprintf(out, `  <header>
    (date: %s)
  </header>
`, timestamp)
	}
	// Setup nav element
	if gen.Nav != "" {
		fmt.Fprintf(out, `  <nav>
    %s
  </nav>
`, indentText(strings.TrimSpace(gen.Nav), 4))
	}
	if gen.TopContent != "" {
		fmt.Fprintf(out, `
    %s
`, indentText(strings.TrimSpace(gen.TopContent), 4))
	}
	// Setup section
	fmt.Fprintln(out, "  <section>")
	stmt := SQLDisplayItems
	rows, err := db.Query(stmt)
	if err != nil {
		return err
	}
	defer rows.Close()
	// Setup and write out the body
	for rows.Next() {
		var (
			link          string
			title         string
			description   string
			authorsSrc    string
            authors       []*gofeed.Person
			enclosuresSrc string
			enclosures    []*Enclosure
			guid          string
			pubDate       string
			dcExt         string
			channel       string
			status        string
			updated       string
            label         string
		)
		if err := rows.Scan(&link, &title, &description, &authorsSrc,
			&enclosuresSrc, &guid, &pubDate, &dcExt,
			&channel, &status, &updated, &label); err != nil {
			fmt.Fprintf(gen.eout, "error (%s): s\n", stmt, err)
			continue
		}
        if authorsSrc != "" {
            authors = []*gofeed.Person{}
            if err := json.Unmarshal([]byte(authorsSrc), &authors); err != nil {
                fmt.Fprintf(gen.eout, "error (authors: %s): %s\n", authorsSrc, err)
                authors = nil
            }            
        }
		if enclosuresSrc != "" {
			enclosures = []*Enclosure{}
			if err := json.Unmarshal([]byte(enclosuresSrc), &enclosures); err != nil {
				fmt.Fprintf(gen.eout, "error (enclosures: %s): %s\n", err, enclosuresSrc)
				enclosures = nil
			}
		}

		if err := gen.WriteItem(out, link, title, description, authors,
			enclosures, guid, pubDate, dcExt,
			channel, status, updated, label); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	fmt.Fprintln(out, "  </section>")
	if gen.Footer != "" {
		fmt.Fprintf(out, "  <footer>\n    %s\n  </footer>\n", indentText(strings.TrimSpace(gen.Footer), 4))
	}
	if gen.BottomContent != "" {
		fmt.Fprintf(out, `
    %s
`, indentText(strings.TrimSpace(gen.BottomContent), 4))
	}
	// close the body
	return nil
}



func toXMLString(input string) string {
	const (
		XML_AMP = "&#38;"
		XML_APOS = "&#39;"
		XML_GT = "&#62;"
		XML_LT = "&#60;"
		XML_QUOT = "&#34;"
	)
	input = strings.ReplaceAll(input, "&", XML_AMP) // Encode ampersand first to avoid double encoding
	input = strings.ReplaceAll(input, "<", XML_LT)  // Less than sign
	input = strings.ReplaceAll(input, ">", XML_GT)  // Greater than sign
	input = strings.ReplaceAll(input, "\"", XML_QUOT) // Double quote
	input = strings.ReplaceAll(input, "'", XML_APOS)  // Apostrophe
	return input
}

func (gen *Generator) WriteItemRSS(out io.Writer, link string, title string, description string, authors []*gofeed.Person,
	enclosures []*Enclosure, guid string, pubDate string, dcExt string,
	channel string, status string, updated string, label string) error {
	// Setup expressing update time.
	pressTime := pubDate
	if len(pressTime) > 10 {
		pressTime = pressTime[0:10]
	}
	if updated != "" {
		if len(updated) > 10 {
			updated = updated[0:10]
		}
		pressTime += ", updated: " + updated
	}
	// Wrap the Item
	fmt.Fprintf(out, `    <item>
`)
	defer fmt.Fprintf(out, `    </item>`)
	// Setup the Title
	if title != "" {
		fmt.Fprintf(out, "      <title>%s</title>\n", strings.TrimSpace(toXMLString(title)))
	}
	if link != "" {
		fmt.Fprintf(out, "      <link>%s</link>\n", strings.TrimSpace(link))
	}
	if description != "" {
		fmt.Fprintf(out, `      <description>
        <![CDATA[%s]]>
      </description>
`, indentText(strings.TrimSpace(description), 8))
	}
	if authors != nil {
        for _, author := range authors  {
            if author.Email != "" && author.Name != "" {
        		fmt.Fprintf(out, "      <author>%s (%s)</author>\n", author.Email, author.Name)
            }
        }
	}
	if enclosures != nil {
		for _, enclosure := range enclosures {
			fmt.Fprintf(out, `      <enclosure url=%q length=%q type=%q />
`, strings.TrimSpace(enclosure.Url), enclosure.Length, strings.TrimSpace(enclosure.Type))
		}
	}
	if guid != "" {
		fmt.Fprintf(out, "      <guid>%s</guid>\n", strings.TrimSpace(guid))
	}
	if pubDate != "" {
		fmt.Fprintf(out, "      <pubDate>%s</pubDate>\n", strings.TrimSpace(pubDate))
	}
	return nil
}

// WriteRSS writes aggregated items into an HTML page from the contents of the database
func (gen *Generator) WriteRSS(out io.Writer, db *sql.DB, appName string, collection *Collection) error {

	// Create the outer elements of a page.
	fmt.Fprintln(out, `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
  <channel>`)
	defer fmt.Fprintln(out, `
  </channel>
</rss>`)
	// Channel Metadata
	if collection.Title != "" {
		fmt.Fprintf(out, `    <title>%s</title>
`, collection.Title)
	}
	if collection.Description != "" {
		fmt.Fprintf(out, `    <description>
      %s
    </description>
`, indentText(strings.TrimSpace(collection.Description), 6))
	}
	if collection.Link != "" {
		fmt.Fprintf(out, `    <link>%s</link>
`, strings.TrimSpace(collection.Link))
	}
	if collection.Copyright != "" {
		fmt.Fprintf(out, `    <copyright>%s</copyright>
`, strings.TrimSpace(collection.Copyright))
	}
	if collection.ManagingEditor != "" {
		fmt.Fprintf(out, `    <managingEditor>%s</managingEditor>
`, strings.TrimSpace(collection.ManagingEditor))
	}
	if collection.WebMaster != "" {
		fmt.Fprintf(out, `    <webMaster>%s</webMaster>
`, strings.TrimSpace(collection.WebMaster))
	}
	if collection.PubDate != "" {
		fmt.Fprintf(out, `    <pubDate>%s</pubDate>
`, strings.TrimSpace(collection.PubDate))
	}
	// The following are hardcode because they are dependent on the generator and
	// when it executed.
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(out, `    <lastBuildDate>%s</lastBuildDate>
`, timestamp)
	fmt.Fprintf(out, `    <generator>%s/%s</generator>
    <docs>https://cyber.harvard.edu/rss/rss.html</docs>
`, appName, Version)


	// Setup  items
	stmt := SQLDisplayItems
	rows, err := db.Query(stmt)
	if err != nil {
		return err
	}
	defer rows.Close()
	// Setup and write out the body
	for rows.Next() {
		var (
			link          string
			title         string
			description   string
            authorsSrc    string
			authors       []*gofeed.Person
			enclosuresSrc string
			enclosures    []*Enclosure
			guid          string
			pubDate       string
			dcExt         string
			channel       string
			status        string
			updated       string
            label         string
		)
		if err := rows.Scan(&link, &title, &description, &authorsSrc,
              &enclosuresSrc, &guid, &pubDate, &dcExt,
              &channel, &status, &updated, &label); err != nil {
            return err
		}
        if authorsSrc != "" {
            authors = []*gofeed.Person{}
            if  err := json.Unmarshal([]byte(authorsSrc), &authors); err != nil {
                fmt.Fprintf(gen.eout, "error (%s): %s\n", authorsSrc, err)
                authors = nil
            }
        }
        if enclosuresSrc != "" {
            enclosures = []*Enclosure{}
            if err := json.Unmarshal([]byte(enclosuresSrc), &enclosures); err != nil {
                fmt.Fprintf(gen.eout, "error (%s): %s\n", enclosuresSrc, err)
                enclosures = nil
            }
        }
		if err := gen.WriteItemRSS(out, link, title, description, authors,
			enclosures, guid, pubDate, dcExt,
			channel, status, updated, label); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	// Close the body via defer
	return nil
}

func getDsnAndCfgName(args []string) (string, string) {
	dsn := args[0]
	if len(args) == 2 {
		return args[0], args[1]
	}
	// Figure out if we have a YAML config or not
	cfgName := strings.TrimSuffix(dsn, ".db") + ".yaml"
	if _, err := os.Stat(cfgName); err != nil {
		return dsn, ""
	}
	return dsn, cfgName
}

// LoadConfig read in the generator configuration (not AppConfig)
// and map the settings into the Generator object.
func (gen *Generator) LoadConfig(cfgName string) error {
	src, err := os.ReadFile(cfgName)
	if err != nil {
		return err
	}
	obj := Generator{}
	if err := yaml.Unmarshal(src, &obj); err != nil {
		return err
	}
	// Pull in the configuration elements
	if obj.AppName != "" {
		gen.AppName = obj.AppName
	}
	if obj.Version != "" {
		gen.Version = obj.Version
	}
	if obj.Title != "" {
		gen.Title = obj.Title
	}
	if obj.Description != "" {
		gen.Description = obj.Description
	}
	if obj.CSS != "" {
		gen.CSS = obj.CSS
	}
	if obj.Modules != nil && len(obj.Modules) > 0 {
		gen.Modules = obj.Modules[:]
	}
	if obj.Header != "" {
		gen.Header = obj.Header
	}
	if obj.Nav != "" {
		gen.Nav = obj.Nav
	}
	if obj.TopContent != "" {
		gen.TopContent = obj.TopContent
	}
	if obj.BottomContent != "" {
		gen.BottomContent = obj.BottomContent
	}
	if obj.Footer != "" {
		gen.Footer = obj.Footer
	}
	return nil
}

func (app AntennaApp) Generate(out io.Writer, eout io.Writer, cfgName string, args []string) error {
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
        if col == nil {
            fmt.Fprintf(eout, "warning could not retrieve %q, skipping\n", cName)
            continue
        }
		if err := col.Generate(out, eout, app.appName, cfg); err != nil {
			fmt.Fprintf(eout, "warning %s: %s\n", col.File, err)
		}
	}
	return nil
}

func (collection *Collection) ApplyFilters(db *sql.DB) error {
	if len(collection.Filters) == 0 {
		return nil
	}
	for _, stmt := range collection.Filters {
        if strings.TrimSpace(stmt) != "" {
            _, err := db.Exec(stmt)
            if err != nil {
                return fmt.Errorf("%s\nstmt: %s", err, stmt)
            }
        }
	}
	return nil
}

func (collection *Collection) Generate(out io.Writer, eout io.Writer, appName string, cfg *AppConfig) error {
	gen, err := NewGenerator(appName)
	if err != nil {
		return err
	}
	if collection.Generator == "" {
		bName := filepath.Base(collection.File)
		xName := filepath.Ext(bName)
		collection.Generator = filepath.Join(cfg.Htdocs, strings.TrimSuffix(bName, xName)+".yaml")
	}
	if _, err := os.Stat(collection.Generator); err == nil {
		src, err := os.ReadFile(collection.Generator)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(src, &gen); err != nil {
			return err
		}
	} else {
		if err := yaml.Unmarshal([]byte(DefaultGeneratorYaml), &gen); err != nil {
			return err
		}
	}
	return gen.Generate(eout, appName, cfg, collection)
}

func (gen *Generator) Generate(eout io.Writer, appName string, cfg *AppConfig, collection *Collection) error {
	// Open DB so we have a place to write data.
	dsn := collection.DbName
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

    // Run the collection filter to determine which items to publish
	if err := collection.ApplyFilters(db); err != nil {
		return err
	}

	// figure out the name and path to write the HTML file to.
	bName := filepath.Base(collection.File)
	xName := filepath.Ext(bName)
	htmlName := filepath.Join(cfg.Htdocs, strings.TrimSuffix(bName, xName)+".html")
	rssName := filepath.Join(cfg.Htdocs, strings.TrimSuffix(bName, xName)+".xml")

	// clear existing page
	if _, err := os.Stat(htmlName); err == nil {
		if err := os.Remove(htmlName); err != nil {
			return nil
		}
	}
	// Create the HTML file
	out, err := os.Create(htmlName)
	if err != nil {
		return err
	}

	// Write out HTML page
	if err := gen.WriteHTML(out, db, appName, collection); err != nil {
		return err
	}
	out.Close()
 
    // clear existing page
	if _, err := os.Stat(rssName); err == nil {
		if err := os.Remove(rssName); err != nil {
			return nil
		}
	}


    // Create the RSS file
	out, err = os.Create(rssName)
	if err != nil {
		return err
	}
    defer out.Close()

    // Write out RSS page
	if err := gen.WriteRSS(out, db, appName, collection); err != nil {
		return err
	}
	return nil
}


// WriteHtmlPage renders an HTML Page using HTML connent and wrapping it based on the 
// generator configuration.
func (gen *Generator) WriteHtmlPage(htmlName string, link string, pubDate string, innerHTML string) error {
	// clear existing page
	if _, err := os.Stat(htmlName); err == nil {
		if err := os.Remove(htmlName); err != nil {
			return nil
		}
	}
	// Create the HTML file
	out, err := os.Create(htmlName)
	if err != nil {
		return err
	}
	defer out.Close()

	// Create the outer elements of a page.
	fmt.Fprintln(out, `<!doctype html>
<html lang="en-US">`)
	defer fmt.Fprintln(out, "</html>")
	// Setup the metadata in the head element
	gen.writeHeadElement(out)
	// Setup body element
	fmt.Fprintln(out, "<body>")
	defer fmt.Fprintln(out, "</body>")
	// Setup header element
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	if gen.Header != "" {
		fmt.Fprintf(out, "  <header>\n    %s\n  </header>\n", indentText(strings.TrimSpace(gen.Header), 4))
	} else if gen.Title != "" {
		fmt.Fprintf(out, `  <header>
    <h1>%s</h1>

    (date: %s)

  </header>
`, gen.Title, timestamp)
	} else {
		fmt.Fprintf(out, `  <header>
    (date: %s)
  </header>
`, timestamp)
	}
	// Setup nav element
	if gen.Nav != "" {
		fmt.Fprintf(out, `  <nav>
    %s
  </nav>
`, indentText(strings.TrimSpace(gen.Nav), 4))
	}
	if gen.TopContent != "" {
		fmt.Fprintf(out, `
    %s
`, indentText(strings.TrimSpace(gen.TopContent), 4))
	}

	// Now render our innerHTML
	fmt.Fprintf(out, `
  <section>
    <article data-published=%q data-link=%q>
      %s
    </article>
  </section>
`, pubDate, link, indentText(innerHTML, 6))

	// Wrap up the page
	if gen.Footer != "" {
		fmt.Fprintf(out, "  <footer>\n    %s\n  </footer>\n", indentText(strings.TrimSpace(gen.Footer), 4))
	}
	if gen.BottomContent != "" {
		fmt.Fprintf(out, `
    %s
`, indentText(strings.TrimSpace(gen.BottomContent), 4))
	}
	return nil
}
