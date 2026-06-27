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
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/
package antennaApp

import (
	"fmt"
	"io"
	"strings"
)

/** HelpTopicsText returns a formatted index of all available help topics with
 * one-line descriptions, suitable for printing to an io.Writer.
 *
 * Returns:
 *   string — formatted topic listing.
 *
 * Example:
 *   fmt.Print(HelpTopicsText())
 */
func HelpTopicsText() string {
	return `Available help topics — type 'antenna help TOPIC' for full details:

Commands:
  add          Add a feed collection to the configuration
  apply        Apply a theme to the page generator YAML
  blogit       Add a post using an automatic date-based directory path
  css          Generate a default CSS stylesheet and patch page.yaml
  del          Remove a collection from the configuration
  generate     Render HTML pages and RSS feeds for all (or one) collection
  harvest      Fetch content from remote feeds into collection databases
  init         Initialize antenna configuration files
  interactive  Guided action wizard — menu-driven help for any action
  items        List all items stored in a collection database
  list         List all defined collections
  page         Render a Markdown file as a standalone HTML page
  pages        List static pages tracked in the pages collection
  post         Add or update a blog post in a collection
  posts        List posts in a collection (with optional count or date range)
  preview      Serve the site on localhost for browser review
  quote        Convert a text-fragment URL into a Markdown excerpt
  rss          Generate an RSS feed file from posts in a collection
  sitemap      Generate sitemap XML index files
  stylefrom    Extract CSS from a LibreOffice HTML export
  themes       List available themes; 'themes new [NAME]' creates a skeleton
  unpage       Remove a page record from the pages collection
  unpost       Remove a post record from a collection

Reference:
  accessibility  Skip navigation link, lang attribute, and CSS requirements
  configuration  antenna.yaml and page.yaml settings reference
  metadata       Front matter fields and the allowed_meta_fields allowlist

  topics       Show this list of all available help topics
`
}

/** PrintHelpTopic writes the help guide for the named topic to w, substituting
 * {app_name}, {version}, {release_date}, and {release_hash} tokens.
 *
 * Parameters:
 *   w           (io.Writer) — destination for help output
 *   topic       (string)    — topic name or alias (case-insensitive)
 *   appName     (string)    — binary name to substitute for {app_name}
 *   version     (string)    — version string
 *   releaseDate (string)    — release date string
 *   releaseHash (string)    — release commit hash
 *
 * Returns:
 *   bool — true if the topic was recognized, false if unknown.
 *
 * Example:
 *   ok := PrintHelpTopic(os.Stdout, "css", "antenna", Version, ReleaseDate, ReleaseHash)
 *   if !ok {
 *       fmt.Fprintln(os.Stderr, "unknown topic")
 *   }
 */
func PrintHelpTopic(w io.Writer, topic, appName, version, releaseDate, releaseHash string) bool {
	topic = strings.ToLower(strings.TrimSpace(topic))
	var text string
	switch topic {
	case "topics", "index":
		text = HelpTopicsText()
	case "add":
		text = `add — add a feed collection

SYNOPSIS
  {app_name} add COLLECTION_FILE [NAME DESCRIPTION]

DESCRIPTION
  Registers a feed collection with {app_name}. COLLECTION_FILE is a Markdown
  document (.md) whose body lists RSS/Atom feed URLs as hyperlinks.  NAME and
  DESCRIPTION are optional; if omitted, {app_name} reads them from the
  document's YAML front matter.

  {app_name} creates a matching SQLite3 database (same basename, .db extension)
  and records the collection in antenna.yaml.

PARAMETERS
  COLLECTION_FILE  path to the Markdown document defining the collection
  NAME             (optional) override the collection name
  DESCRIPTION      (optional) override the collection description

EXAMPLE
  {app_name} add feeds/tech.md
  {app_name} add feeds/tech.md "Tech Feeds" "My technology reading list"
`
	case "apply":
		text = `apply — apply a theme

SYNOPSIS
  {app_name} apply THEME_PATH [YAML_FILE_PATH]

DESCRIPTION
  Applies the theme at THEME_PATH to the page generator YAML file at
  YAML_FILE_PATH. If YAML_FILE_PATH is omitted, the default generator YAML
  (page.yaml) is replaced.

  A theme directory contains Markdown and YAML files whose names map to
  generator YAML attributes: header.md, nav.md, footer.md, head.yaml, etc.
  Run 'antenna help themes' for the full theme directory layout.

PARAMETERS
  THEME_PATH      path to the theme directory
  YAML_FILE_PATH  (optional) path to the generator YAML to update

EXAMPLE
  {app_name} apply theme/my-theme
  {app_name} apply theme/my-theme collection/feeds.yaml
`
	case "blogit":
		text = `blogit — add a post using a date-based directory path

SYNOPSIS
  {app_name} blogit [COLLECTION_NAME] FILEPATH [POST_DATE]

DESCRIPTION
  A variation of 'post' that places the source Markdown document in a
  blog-style date directory tree (e.g. blog/2026/04/12/my-post.md) before
  adding it to the collection.  If POST_DATE is omitted, the current date is
  used.  The collection defaults to pages.md when COLLECTION_NAME is omitted.

PARAMETERS
  COLLECTION_NAME  (optional) path to the collection Markdown file
  FILEPATH         source Markdown document
  POST_DATE        (optional) date in YYYY-MM-DD format

EXAMPLE
  {app_name} blogit index.md notes/my-idea.md
  {app_name} blogit index.md notes/my-idea.md 2026-04-01
`
	case "css":
		text = `css — generate a default CSS stylesheet

SYNOPSIS
  {app_name} css [CSS_PATH]

DESCRIPTION
  Writes a comprehensive starter stylesheet to CSS_PATH within the htdocs
  directory configured in antenna.yaml. If CSS_PATH is omitted it defaults to
  css/site.css. Directory levels are created automatically.

  If a stylesheet already exists at the target path it is backed up to
  CSS_PATH.bak before being overwritten.

  After writing the CSS, {app_name} patches the generator YAML (page.yaml) to
  add a <link rel="stylesheet"> entry pointing to the new file. If page.yaml
  already has a link: section, {app_name} prints instructions for adding the
  entry by hand instead of modifying the file automatically.

  The generated stylesheet includes:
    • CSS custom properties for colors, fonts, and layout (easy to override)
    • Dark-mode overrides via @media (prefers-color-scheme: dark)
    • Skip-navigation link (WCAG 2.4.1 — required by default HTML output)
    • Navigation bar, article cards, standalone pages, and site footer
    • Typography for headings, code blocks, blockquotes, and tables

PARAMETERS
  CSS_PATH  (optional) path relative to htdocs (default: css/site.css)

EXAMPLE
  {app_name} css
  {app_name} css css/custom/theme.css

SEE ALSO
  antenna help accessibility
  antenna help themes
`
	case "del":
		text = `del — remove a collection from the configuration

SYNOPSIS
  {app_name} del COLLECTION_FILE

DESCRIPTION
  Removes the collection identified by COLLECTION_FILE from antenna.yaml.
  The Markdown file and its SQLite3 database are not deleted from disk.

PARAMETERS
  COLLECTION_FILE  path to the Markdown document as registered with 'add'

EXAMPLE
  {app_name} del feeds/tech.md
`
	case "generate", "build":
		text = `generate — render HTML pages and RSS feeds

SYNOPSIS
  {app_name} generate [COLLECTION_NAME]

DESCRIPTION
  Processes all collections (or only COLLECTION_NAME if provided), rendering
  an HTML page and RSS 2.0 feed for each. Output files are written to the
  htdocs directory configured in antenna.yaml.

  The HTML structure is controlled by the page generator YAML (page.yaml or
  a per-collection override). Front matter from each item is emitted as
  <meta> elements in the generated HTML.

PARAMETERS
  COLLECTION_NAME  (optional) process only this collection

ALIASES
  build

EXAMPLE
  {app_name} generate
  {app_name} generate index.md
`
	case "harvest", "fetch":
		text = `harvest — fetch content from remote feeds

SYNOPSIS
  {app_name} harvest [COLLECTION_NAME]

DESCRIPTION
  Retrieves RSS/Atom feed content for all collections (or only
  COLLECTION_NAME) and stores harvested items in each collection's SQLite3
  database. Run 'generate' afterwards to rebuild the HTML pages.

PARAMETERS
  COLLECTION_NAME  (optional) harvest only this collection

ALIASES
  fetch

EXAMPLE
  {app_name} harvest
  {app_name} harvest feeds/tech.md
`
	case "init":
		text = `init — initialize Antenna configuration

SYNOPSIS
  {app_name} init

DESCRIPTION
  Creates the two configuration files {app_name} needs if they do not exist:

    antenna.yaml  main configuration (htdocs path, port, collections list)
    page.yaml     page generator (link, meta, nav, header, footer, scripts)

  Also creates a default pages.md collection and pages.db database.

  After running init, run '{app_name} css' to generate a starter stylesheet
  and automatically link it in page.yaml.

EXAMPLE
  mkdir myblog && cd myblog
  {app_name} init
  {app_name} css
`
	case "interactive", "tui":
		text = `interactive — guided action wizard

SYNOPSIS
  {app_name} interactive [ACTION [PARAMETERS]]

DESCRIPTION
  Starts a menu-driven session that walks through any {app_name} action step
  by step. Each parameter is explained and pre-filled from arguments already
  on the command line. The complete command is shown before it runs.

  If ACTION is omitted a menu of all available actions is presented.
  Useful for learning the command syntax interactively.

ALIASES
  tui

EXAMPLE
  {app_name} interactive
  {app_name} interactive post
`
	case "items":
		text = `items — list all items in a collection database

SYNOPSIS
  {app_name} items [COLLECTION_NAME]

DESCRIPTION
  Prints a Markdown list of all items stored in the named collection's
  SQLite3 database.  Items include blog posts added via 'post' or 'blogit',
  and feed entries harvested via 'harvest'.  The list is in descending
  publication-date order.

  Each line shows: - [Title](link), YYYY-MM-DD, status (label)

  If COLLECTION_NAME is omitted it defaults to pages.md.

  See 'posts' for a list restricted to blog posts (items with a postPath),
  and 'pages' for a list of static page entries.

PARAMETERS
  COLLECTION_NAME  (optional) collection Markdown file (default: pages.md)

EXAMPLE
  {app_name} items
  {app_name} items index.md
`
	case "list":
		text = `list — list all defined collections

SYNOPSIS
  {app_name} list

DESCRIPTION
  Prints the collection Markdown filenames registered in antenna.yaml.
  Each entry is the path passed to 'add'.

EXAMPLE
  {app_name} list
`
	case "page":
		text = `page — render a Markdown file as a standalone HTML page

SYNOPSIS
  {app_name} page INPUT_PATH [OUTPUT_PATH]

DESCRIPTION
  Renders INPUT_PATH as an HTML file using the page generator YAML.  The
  output file is placed next to INPUT_PATH with a .html extension unless
  OUTPUT_PATH is specified.  The page is recorded in the pages collection
  (pages.md / pages.db) so it appears in 'antenna pages' output.

  Pages are excluded from RSS feeds. This action is for static content such
  as About pages, contact pages, and search pages.

  WARNING: HTML in the Markdown source passes through unchanged (unsafe mode).
  Only run this action on files you control and trust.

PARAMETERS
  INPUT_PATH   path to the source Markdown file
  OUTPUT_PATH  (optional) explicit output HTML path

EXAMPLE
  {app_name} page about.md
  {app_name} page about.md htdocs/about.html
`
	case "pages":
		text = `pages — list static pages

SYNOPSIS
  {app_name} pages

DESCRIPTION
  Prints the page entries tracked in the pages collection (pages.md / pages.db)
  in descending updated-timestamp order.

  Compare with 'posts' (blog posts in a collection) and 'items' (all items
  in any collection database).

EXAMPLE
  {app_name} pages
`
	case "post":
		text = `post — add or update a blog post in a collection

SYNOPSIS
  {app_name} post [COLLECTION_NAME] FILEPATH

DESCRIPTION
  Adds FILEPATH to COLLECTION_NAME (default: pages.md). The file's YAML
  front matter supplies the required fields. If a record with the same
  postPath already exists, it is overwritten.

  Required front matter fields:
    title or description  — at least one must be present
    postPath              — relative path to the generated HTML file
    pubDate               — publication date (YYYY-MM-DD recommended)

  Recommended additional fields:
    link        public URL to the post
    author      author name
    description summary for RSS and search engines
    keywords    list of tags

  After posting, run 'generate' to rebuild the collection HTML and RSS feed.

  WARNING: HTML in the Markdown source passes through unchanged (unsafe mode).
  Only run this action on files you control and trust.

PARAMETERS
  COLLECTION_NAME  (optional) collection Markdown file (default: pages.md)
  FILEPATH         path to the source Markdown document

EXAMPLE
  {app_name} post index.md blog/2026/04/12/my-post.md

SEE ALSO
  antenna help metadata
  antenna help blogit
`
	case "posts":
		text = `posts — list posts in a collection

SYNOPSIS
  {app_name} posts COLLECTION_NAME [COUNT | FROM_DATE TO_DATE]

DESCRIPTION
  Prints a Markdown list of blog posts in COLLECTION_NAME in descending
  pubDate order. Items must have both a pubDate and a postPath to appear.

  Optional parameters constrain the list:
    COUNT           integer — return only the N most recent posts
    FROM_DATE       YYYY-MM-DD start of a date range
    TO_DATE         YYYY-MM-DD end of a date range

PARAMETERS
  COLLECTION_NAME  collection Markdown file
  COUNT            (optional) maximum number of posts to list
  FROM_DATE        (optional) start date (requires TO_DATE)
  TO_DATE          (optional) end date (requires FROM_DATE)

EXAMPLE
  {app_name} posts index.md
  {app_name} posts index.md 10
  {app_name} posts index.md 2026-01-01 2026-06-30
`
	case "preview":
		text = `preview — serve the site on localhost

SYNOPSIS
  {app_name} preview

DESCRIPTION
  Starts a local HTTP server serving the htdocs directory so you can review
  the generated site in a browser. The host and port are set in antenna.yaml
  (defaults: localhost:8000).

  Press Ctrl-C to stop the server.

EXAMPLE
  {app_name} preview
  open http://localhost:8000
`
	case "quote", "reply":
		text = `quote — convert a text-fragment URL into a Markdown excerpt

SYNOPSIS
  {app_name} quote TEXT_FRAGMENT_URL

DESCRIPTION
  Parses a TEXT_FRAGMENT_URL (a URL with a #:~:text= fragment) into a
  Markdown quotation. Output is written to standard out. Redirect it into a
  file to use as the basis for a response post.

ALIASES
  reply

PARAMETERS
  TEXT_FRAGMENT_URL  a URL ending in a #:~:text= text-fragment selector

EXAMPLE
  {app_name} quote "https://example.com/article#:~:text=interesting%20passage"
`
	case "rss":
		text = `rss — generate an RSS feed file from posts

SYNOPSIS
  {app_name} rss COLLECTION_NAME RSS_FILENAME [COUNT | FROM_DATE TO_DATE]

DESCRIPTION
  Writes an RSS 2.0 feed for COLLECTION_NAME to RSS_FILENAME. The optional
  parameters work the same way as for 'posts': a COUNT or a FROM_DATE/TO_DATE
  range to limit the items included.

PARAMETERS
  COLLECTION_NAME  collection Markdown file
  RSS_FILENAME     output path for the RSS file
  COUNT            (optional) maximum number of items
  FROM_DATE        (optional) start date (requires TO_DATE)
  TO_DATE          (optional) end date (requires FROM_DATE)

EXAMPLE
  {app_name} rss index.md index.xml
  {app_name} rss index.md archive.xml 2026-01-01 2026-06-30
`
	case "sitemap":
		text = `sitemap — generate sitemap XML files

SYNOPSIS
  {app_name} sitemap

DESCRIPTION
  Generates a set of sitemap files (sitemap_index.xml, sitemap_1.xml, …)
  for all pages and posts found via antenna.yaml. Place these in the root of
  your htdocs directory so search engines can discover your content.

EXAMPLE
  {app_name} sitemap
`
	case "stylefrom":
		text = `stylefrom — extract CSS from a LibreOffice HTML export

SYNOPSIS
  {app_name} stylefrom INPUT_FILE [OUTPUT_PATH]

DESCRIPTION
  Extracts the embedded CSS from a LibreOffice Writer HTML export
  (INPUT_FILE, .html or .htm) and writes it to OUTPUT_PATH. OUTPUT_PATH
  defaults to theme/style.css; the directory is created if needed.

  Use this action to seed a theme stylesheet from a styled Writer document.

PARAMETERS
  INPUT_FILE   path to the LibreOffice-exported HTML file
  OUTPUT_PATH  (optional) output CSS path (default: theme/style.css)

EXAMPLE
  {app_name} stylefrom my-doc.html
  {app_name} stylefrom my-doc.html css/libreoffice.css
`
	case "themes", "themes new":
		text = ThemeHelpText
	case "unpage":
		text = `unpage — remove a page record from the pages collection

SYNOPSIS
  {app_name} unpage INPUT_PATH

DESCRIPTION
  Removes the page record associated with INPUT_PATH from the pages collection.
  The HTML and Markdown files on disk are not deleted.

PARAMETERS
  INPUT_PATH  path to the source Markdown file as it was passed to 'page'

EXAMPLE
  {app_name} unpage about.md
`
	case "unpost":
		text = `unpost — remove a post record from a collection

SYNOPSIS
  {app_name} unpost COLLECTION_NAME URL | POST_PATH

DESCRIPTION
  Removes an item from COLLECTION_NAME using either the public URL associated
  with the post or the POST_PATH value. The source Markdown and generated HTML
  files on disk are not deleted.

PARAMETERS
  COLLECTION_NAME  collection Markdown file
  URL or POST_PATH URL or postPath value identifying the item to remove

EXAMPLE
  {app_name} unpost index.md https://example.com/blog/2026/04/12/my-post.html
  {app_name} unpost index.md blog/2026/04/12/my-post.html
`
	case "accessibility":
		text = `accessibility — skip navigation, lang attribute, and CSS requirements

DESCRIPTION
  {app_name} generates HTML that meets WCAG 2.1 Level A success criterion
  2.4.1 (Bypass Blocks) and uses semantic HTML5 markup throughout.

SKIP NAVIGATION LINK
  Every generated page includes a skip-navigation link immediately after
  <body> that lets keyboard users jump past the site navigation directly to
  the main content:

    <a href="#main-content" class="skip-link">Skip to main content</a>

  The skip link is visually hidden off-screen until it receives keyboard
  focus, at which point it becomes visible. The CSS generated by
  'antenna css' includes the required .skip-link and .skip-link:focus rules.

  If you maintain your own stylesheet, add these rules:

    .skip-link {
      position: absolute;
      top: -999px;
      left: -999px;
      padding: 0.5rem 1rem;
      background: #333;
      color: #fff;
      text-decoration: none;
      z-index: 9999;
    }
    .skip-link:focus {
      position: static;
      top: auto;
      left: auto;
      display: block;
    }

LANG ATTRIBUTE
  The <html> element includes a lang attribute from page.yaml. The default
  is "en-US". Change it for non-English sites:

    lang: ja        # Japanese
    lang: fr-FR     # French (France)

ARTICLE FOOTER FOR SOURCE LINKS
  Feed item cards use <article> > <footer> (not <address>) for the source
  URL at the bottom of each card. CSS selectors targeting the old pattern
  "article address" must be updated to "article footer".

  The default CSS from 'antenna css' already uses the correct selector.

SEMANTIC TIME ELEMENTS
  Publication and update dates are wrapped in <time datetime="YYYY-MM-DD">
  elements and placed outside the <h2> heading so screen readers do not
  announce the date as part of the article title.

SEE ALSO
  antenna help css
  antenna help configuration
`
	case "configuration":
		text = `configuration — antenna.yaml and page.yaml settings reference

ANTENNA.YAML  (main configuration)

  port         (optional, default: 8000)   localhost port for 'preview'
  host         (optional, default: localhost) host name for 'preview'
  htdocs       (optional, default: ".")    directory for generated HTML/RSS
  generator    (optional, default: page.yaml) default page generator YAML
  collections  (required) list of collection objects

  Each collection object:
    file       (required) path to the collection Markdown document
    title      (optional, default: filename) display name
    generator  (optional) per-collection page generator YAML override
    mode       (optional) rendering mode: "aggregate" (default) or "page-index"
               "aggregate"  — feed-item cards from the items table (default)
               "page-index" — simple <ul> link list from the pages table

EXAMPLE antenna.yaml:

  htdocs: htdocs
  port: 8000
  collections:
    - file: index.md                 # aggregate (default)
    - file: links.md
      generator: links-page.yaml
    - file: pages.md
      mode: page-index               # renders a simple link list

PAGE.YAML  (page generator)

  lang               (optional, default: en-US) lang= attribute on <html>
  title              (optional) page <title> override
  meta               (optional) list of <meta> element attribute maps
  link               (optional) list of <link> element attribute maps
  script             (optional) list of <script> element attribute maps
  style              (optional) inline CSS injected at end of <head>
  header             (optional) innerHTML of <header>
  nav                (optional) innerHTML of <nav aria-label="Site navigation">
  top_content        (optional) content between <nav> and <main>
  bottom_content     (optional) content between </main> and <footer>
  footer             (optional) innerHTML of <footer>
  allowed_meta_fields (optional) allowlist of front matter keys to emit as <meta>

EXAMPLE page.yaml:

  lang: en-US
  link:
    - rel: stylesheet
      type: text/css
      href: /css/site.css
  header: |
    <h1>My Blog</h1>
  nav: |
    <ul>
      <li><a href="/">Home</a></li>
      <li><a href="/about.html">About</a></li>
    </ul>
  footer: |
    <p>© 2026 Your Name</p>
  allowed_meta_fields:
    - title
    - author
    - description
    - keywords

SEE ALSO
  antenna help metadata
  antenna help accessibility
`
	case "metadata":
		text = `metadata — front matter fields and the allowed_meta_fields allowlist

DESCRIPTION
  Every YAML front matter field in a Markdown post or page is emitted into
  the generated HTML <head> as a <meta name="KEY" content="VALUE"> element.
  The same key-value pair is also written as a data-pagefind-filter attribute
  on the enclosing <article> element for PageFind faceted search.

STANDARD FIELDS

  title          Sets <title> in the HTML head (not emitted as <meta>)
  description    Short summary for search engines and RSS
  author         Author name
  keywords       List of tags; each value gets its own <meta> pair
  pubDate        Publication date — required for posts (YYYY-MM-DD)
  postPath       Relative path to the generated HTML file — required for posts
  link           Public URL of the post
  series         Series name for multi-part posts
  seriesNumber   Position within the series
  dateCreated    Creation date
  dateModified   Last-modified date
  datePublished  Publication date (alias for pubDate in page context)

EXAMPLE FRONT MATTER:

  ---
  title: "My Post Title"
  description: "A short summary."
  author: "Your Name"
  keywords:
    - go
    - web
  pubDate: "2026-06-27"
  postPath: "blog/2026/06/27/my-post.html"
  series: "Building with Antenna"
  seriesNumber: 3
  ---

CONTROLLING WHICH FIELDS ARE PUBLISHED (allowed_meta_fields)

  By default all front matter keys are emitted as <meta> elements. To
  restrict which keys are published, set allowed_meta_fields in page.yaml:

    allowed_meta_fields:
      - title
      - author
      - description
      - keywords
      - series
      - seriesNumber

  When this list is set, only the listed keys appear in generated HTML.
  Internal workflow fields (build flags, notes) are silently omitted.

SEE ALSO
  antenna help configuration
  antenna help post
`
	default:
		return false
	}

	r := strings.NewReplacer(
		"{app_name}", appName,
		"{version}", version,
		"{release_date}", releaseDate,
		"{release_hash}", releaseHash,
	)
	fmt.Fprintln(w, r.Replace(text))
	return true
}
