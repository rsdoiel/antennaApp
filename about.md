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
version: 0.1.0
license_url: https://www.gnu.org/licenses/agpl-3.0.en.html

programming_language:
  - Go >= 1.27.3

keywords:
  - RSS
  - Feeds
  - Linkblog
  - website generator

date_released: 2026-11-20
---

About this software
===================

## antennaApp 0.1.0

- Antenna sites now publish for the Markdown web (Dave Winer's BonBon browser and readers like it) as well as for HTML browsers, from the same Markdown. See `antenna help markdown-web`
- Links between your pages are written to the `.md` file; in the HTML output a link to a page, collection or post that antenna renders becomes the `.html` page, with a warning for a link to a file it does not render
- Each collection with items also gets a Markdown rendering of them, `STEM_items.md`, declared in the HTML head as its Markdown alternate
- `antenna generate` writes `.well-known/toc.opml`: the site's pages, collections and posts (oldest first, by year and month), with `include` outlines for the other sites on the host named by the new `subsites` list in `antenna.yaml`; it gives pages and posts previous and next links from it when their front matter names none
- New `bonbon` setting in `antenna.yaml` writes `.well-known/bonbon.json`; with no setting no file is written and none is removed
- Front matter keys BonBon reads: `link`, `urlPrev`, `urlNext`, `rightMarginImage` (with `rightMarginImageAlt`), `pubDate` in the RSS 2.0 form, and `{{page.NAME}}` macros in pages and posts
- New `antenna validate [-json] [-online] [PATH]`: a read-only check of a file, a directory or the whole site, with 21 stable finding codes (front matter, dates, links, raw HTML, images, macros, stale table of contents, `.nojekyll` on GitHub Pages and more); exit 0 when clean, 1 when there are errors or warnings
- New `antenna migrate-links [-w] [PATH]`: rewrites links to `.html` pages that antenna renders from Markdown into `.md` links in your Markdown; a dry-run diff by default, `-w` writes, and only the link destinations change
- `antenna blogit` now makes the post publishable: it adds `pubDate` (the RSS form, in the zone of your computer, `TZ` overrides it) and a `title` from the first heading when the file has none, and never changes what is there; before this a blogit post stayed a draft
- A date with a zone keeps it: an RSS `pubDate` such as `Mon, 30 Nov 2026 21:30:00 -0800` is stored with its offset, so a late evening post stays on its own day and month in the table of contents, the feeds and the page
- `antenna preview` serves `.md` as `text/markdown`, serves `.well-known`, and sends Markdown to a client that asks for it with `Accept`
- Generated files are deterministic: a file is rewritten only when its content changes, so regenerating an unchanged site changes no file; expect one round of changes in a site you regenerate for the first time (attribute order, `.md` links written as `.html`, previous and next links)
- Breaking for scripts: `antenna unpage` exits 1 when no page record matches the path
- Fixed: one unreadable item row no longer stops a collection's HTML page and items Markdown from being written; `{{page.title}}` inside a link you wrote no longer makes a link inside a link; an items file left by a collection that became a page index is removed
- Every help topic's manual page is regenerated from `antenna help TOPIC` (the Makefile does it), so the `.7.md` pages are current; new topics `markdown-web`, `validate` and `migrate-links`
- Requires Go 1.27.3 and the `opml` module 0.0.12

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

- Go >= 1.27.3




## Software Requirements

- Go >= 1.27.3
- github.com/rsdoiel/opml >= 0.0.12
- github.com/rsdoiel/stylefrom >= 0.0.2
- github.com/rsdoiel/termlib >= 0.0.10
- CMTools >= 0.0.45b


## Software Suggestions

- GNU Make >= 3.4
- Pandoc >= 3.9
- Bash or PowerShell


