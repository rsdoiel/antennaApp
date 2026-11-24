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
version: 0.1.2
license_url: https://www.gnu.org/licenses/agpl-3.0.en.html

programming_language:
  - Go >= 1.27.4

keywords:
  - RSS
  - Feeds
  - Linkblog
  - website generator

date_released: 2026-11-23
---

About this software
===================

## antennaApp 0.1.2

- `antenna harvest` finds a site's feed from the address you list (the address itself, an `rss` outline in its `toc.opml`, its `<link rel="alternate">`, or a common path such as `feed/`), remembers it in `channels.discovered_link`, and obeys each host's `robots.txt` with a polite pause (`harvest: pause:` in `antenna.yaml`); redirects are now followed
- `html: false` on a collection makes it Markdown-only: items Markdown, RSS and posts, no HTML; `markdownItems` switches a collection's `_items.md` on or off
- `antenna prune [-w] [-all]` removes the pages whose source file is gone, with their recorded outputs
- Textcasting alignment: an untitled item has no placeholder heading, links and emphasis reach every item, a name-only author is `<dc:creator>`, every `<guid>` says `isPermaLink`, and a post made by antenna always has a guid (its front matter's, else its link)
- Fixed: front matter in a harvested `source:markdown` no longer shows in the collection page or `_items.md`; `generate` writes the RSS and OPML from the readable rows when one cannot be read
- Removed: ODT and OTT support (such a file is refused with a message to save it as Markdown)
- Deprecated: the `pages.outputPath` column will stop being written; use `outputs`

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


