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
version: 0.2.2
license_url: https://www.gnu.org/licenses/agpl-3.0.en.html

programming_language:
  - Go >= 1.27.4

keywords:
  - RSS
  - Feeds
  - Linkblog
  - website generator

date_released: 2026-11-26
---

About this software
===================

## antennaApp 0.2.2

- Fixed: a post's previous and next now follow BonBon's rule at the ends (antennaApp DR-0040): the oldest post's previous is the newest and the newest post's next is the oldest, as in BonBon and cmarkbrowser. The HTML of the oldest and newest post changes on the next generate; one post has none, two point at each other. Pages, collections and declared sequences are unchanged
- Changed: harvest fetches through markdownweb's fetch package (v0.0.4, markdownweb DR-0002): a 429 or 503 with a Retry-After of up to two minutes is waited out and tried once more, a longer one is unavailable (69), and a feed or table of contents over 32 MiB is a data error (65). antenna keeps its pause, User-Agent, redirects and exit classes, and still does not read robots.txt. userAgent in antenna.yaml is documented
- Changed: front matter, macros and includes come from github.com/rsdoiel/frontmatter v0.0.1. A dotted key reaches into a nested table ({{page.source.title}}) and a leading UTF-8 byte-order mark in a source file is skipped
- Needs Go 1.27.4

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
- github.com/rsdoiel/frontmatter >= 0.0.1
- github.com/rsdoiel/markdownweb >= 0.0.4
- github.com/rsdoiel/opml >= 0.0.12
- github.com/rsdoiel/stylefrom >= 0.0.2
- github.com/rsdoiel/termlib >= 0.0.10
- CMTools >= 0.0.45b


## Software Suggestions

- GNU Make >= 3.4
- Pandoc >= 3.9
- Bash or PowerShell


