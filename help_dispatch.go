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
		text = AddHelpText
	case "apply":
		text = ApplyHelpText
	case "blogit":
		text = BlogitHelpText
	case "css":
		text = CssHelpText
	case "del":
		text = DelHelpText
	case "generate", "build":
		text = GenerateHelpText
	case "harvest", "fetch":
		text = HarvestHelpText
	case "init":
		text = InitHelpText
	case "interactive", "tui":
		text = InteractiveHelpText
	case "items":
		text = ItemsHelpText
	case "list":
		text = ListHelpText
	case "page":
		text = PageHelpText
	case "pages":
		text = PagesHelpText
	case "post":
		text = PostHelpText
	case "posts":
		text = PostsHelpText
	case "preview":
		text = PreviewHelpText
	case "quote", "reply":
		text = QuoteHelpText
	case "rss":
		text = RssHelpText
	case "sitemap":
		text = SitemapHelpText
	case "stylefrom":
		text = StylefromHelpText
	case "themes", "themes new":
		text = ThemeHelpText
	case "unpage":
		text = UnpageHelpText
	case "unpost":
		text = UnpostHelpText
	case "accessibility":
		text = AccessibilityHelpText
	case "configuration":
		text = ConfigurationHelpText
	case "metadata":
		text = MetadataHelpText
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
