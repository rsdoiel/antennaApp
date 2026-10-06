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
  - pages nice with other static site tools like  [PageFind](https://pagefind.app "A browser side search engine") and
  [FlatLake](https://flatlake.app "A static JSON API driven by front matter in Markdown documents")
authors:
  - family_name: Doiel
    given_name: R. S. Doiel
    id: https://orcid.org/0000-0003-0900-6903



repository_code: https://github.com/rsdoiel/antennaApp
version: 0.0.27
license_url: https://www.gnu.org/licenses/agpl-3.0.en.html

programming_language:
  - Go

keywords:
  - RSS
  - Feeds
  - Linkblog
  - website generator

date_released: 2026-10-05
---

About this software
===================

## antennaApp 0.0.27

- Added `antenna completion bash|powershell`, with `-install` to set up completion for every new session
- Exit codes now follow the workspace convention (DR-0003): 0 ok, 1 no, 2 usage, 65 data, 66 no input, 69 unavailable, 70 internal, 73 cannot create, 74 I/O, 75 try again, 77 permission, 78 config; they are documented under EXIT STATUS in the manual
- Breaking for scripts: `harvest` and `generate` finish all their work, then exit non-zero when any feed, collection, post or page failed (a dead feed now exits 69); `posts`, `items` and `pages` that find nothing exit 0; a surplus argument is refused with exit 2 (`preview extra` used to start the server); `del` of an unknown collection exits 1
- `antenna sitemap` now lists only the site's pages (from the pages collection) and its local posts, never harvested feed items; numbers its files across collections, writes them to htdocs, warns and exits 0 when there is nothing to list, and has `-clean` to remove stale `sitemap_N.xml` files
- Fixed RSS feeds: enclosure URLs are XML-escaped (feeds with query strings were not well-formed), an unknown enclosure length is written as 0, and items now carry their own `pubDate` in RFC 1123 form (#22)
- Fixed the skip link: `<main>` is focusable, so following it moves keyboard focus into the main area
- Fixed `items` dropping rows that had NULL columns, `blogit` treating hyphenated file names as dates, and a feed whose body was cut off being parsed as if complete

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
- pages nice with other static site tools like  [PageFind](https://pagefind.app "A browser side search engine") and
[FlatLake](https://flatlake.app "A static JSON API driven by front matter in Markdown documents")

- [License](https://www.gnu.org/licenses/agpl-3.0.en.html)
- [Code Repository](https://github.com/rsdoiel/antennaApp)
  - [Issue Tracker](https://github.com/rsdoiel/antennaApp/issues)

## Programming languages

- Go




## Software Requirements

- Go >= 1.26.2
- CMTools >= 0.0.45b


## Software Suggestions

- GNU Make >= 3.4
- Pandoc >= 3.9
- Bash or PowerShell


