---
title: antennaApp
abstract: |-
  **antenna** is a tool for building feed oriented websites using Markdown.
  If you can edit a Markdown list of links you can generate a static site
  RSS Reader with little effort beyond curating your list of links.
  You can create simple websites make from the content of Markdown pages.
  You can create blog by posting the Markdown documents to a collection.
  The goal of **antenna** is to put you in control of the web content
  you read or write through using Markdown.

  **antenna** automates must of the process of creating a site using Markdown
  files. It handles the creation of HTML, RSS, OPML and sitemap.xml for you. This
  let's you focus on the content written in Markdown.

  **antenna** supports a simple theme system defined Markdown files used to
  describe page elements, a file to define page metadata and CSS files to define
  the page's layout and presentation. Each collection defined for your site
  may have it's own theme.

  Features:

  - Makes it easy to generate a website only using Markdown
  - Makes it trivial to generate a blog, link blog or feed reading site using Markdown
  - supports as multiple feed collections per site
  - provides actions to automate most of your site curation leaving you time to focus on Markdown content
  - HTML, RSS 2.0 XML, OPML and sitemap.xml are generated automatically
  - A preview web server is provided so you can read the curated content on your computer
  - Publishes for Markdown readers such as Dave Winer's BonBon as well as for web browsers: `.md` links, Markdown renderings of collections, a `.well-known/toc.opml` table of contents, and BonBon's front matter
  - `antenna validate` checks a site for mistakes and for what Markdown readers need, and `antenna migrate-links` brings older links up to date
  - pages nice with other static site tools like  [PageFind](https://pagefind.app "A browser side search engine") and
  [FlatLake](https://flatlake.app "A static JSON API driven by front matter in Markdown documents")
authors:
  - family_name: Doiel
    given_name: R. S. Doiel
    id: https://orcid.org/0000-0003-0900-6903



repository_code: https://github.com/rsdoiel/antennaApp
version: 0.2.1
license_url: https://www.gnu.org/licenses/agpl-3.0.en.html

programming_language:
  - Go >= 1.27.4

keywords:
  - RSS
  - Feeds
  - Linkblog
  - website generator

date_released: 2026-11-25
---

About this software
===================

## antennaApp 0.2.1

- Fixed: relative links and images in a harvested item (`/images/a.png`, `/2026/post/`, `c.html`) are made absolute against the item's link when it is rendered, in `STEM_items.html`, `STEM_items.md` and `STEM.xml`; they used to resolve against the aggregating site and break. The stored item is not changed. An item with no usable link is left as written and `generate` warns once for the collection with the count
- Added: `harvest` records every attempt to fetch a listed feed (HTTP status, error) in a `harvest_log` table and, at the end of the run, warns about a feed whose attempts all failed over 7 days (at least 3), with the status in words (`last status 404 Not Found`); 429 and 503 count as failures. Settings `failing_days`, `failing_attempts` and `log_days` under `harvest:`. The exit status is unchanged
- Added: a cap on the items each feed shows, `max_items_per_feed` (default 1000, 0 for none) and `feed_limits` by feed address in `antenna.yaml`. Older items are hidden, not deleted, and your own posts are never capped. `generate` warns when a feed is cut. A feed with more than 1000 items in its window (for example Korea Times on the antenna site) now shows its newest 1000

## Authors

- [R. S. Doiel Doiel](https://orcid.org/0000-0003-0900-6903)






**antenna** is a tool for building feed oriented websites using Markdown.
If you can edit a Markdown list of links you can generate a static site
RSS Reader with little effort beyond curating your list of links.
You can create simple websites make from the content of Markdown pages.
You can create blog by posting the Markdown documents to a collection.
The goal of **antenna** is to put you in control of the web content
you read or write through using Markdown.

**antenna** automates must of the process of creating a site using Markdown
files. It handles the creation of HTML, RSS, OPML and sitemap.xml for you. This
let's you focus on the content written in Markdown.

**antenna** supports a simple theme system defined Markdown files used to
describe page elements, a file to define page metadata and CSS files to define
the page's layout and presentation. Each collection defined for your site
may have it's own theme.

Features:

- Makes it easy to generate a website only using Markdown
- Makes it trivial to generate a blog, link blog or feed reading site using Markdown
- supports as multiple feed collections per site
- provides actions to automate most of your site curation leaving you time to focus on Markdown content
- HTML, RSS 2.0 XML, OPML and sitemap.xml are generated automatically
- A preview web server is provided so you can read the curated content on your computer
- Publishes for Markdown readers such as Dave Winer's BonBon as well as for web browsers: `.md` links, Markdown renderings of collections, a `.well-known/toc.opml` table of contents, and BonBon's front matter
- `antenna validate` checks a site for mistakes and for what Markdown readers need, and `antenna migrate-links` brings older links up to date
- pages nice with other static site tools like  [PageFind](https://pagefind.app "A browser side search engine") and
[FlatLake](https://flatlake.app "A static JSON API driven by front matter in Markdown documents")

- [License](https://www.gnu.org/licenses/agpl-3.0.en.html)
- [Code Repository](https://github.com/rsdoiel/antennaApp)
  - [Issue Tracker](https://github.com/rsdoiel/antennaApp/issues)

## Programming languages

- Go >= 1.27.4




## Software Requirements

- Go >= 1.27.4
- github.com/rsdoiel/opml >= 0.0.12
- github.com/rsdoiel/stylefrom >= 0.0.2
- github.com/rsdoiel/termlib >= 0.0.10
- CMTools >= 0.0.45b


## Software Suggestions

- GNU Make >= 3.4
- Pandoc >= 3.9
- Bash or PowerShell


