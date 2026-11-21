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
version: 0.1.1
license_url: https://www.gnu.org/licenses/agpl-3.0.en.html

programming_language:
  - Go >= 1.27.4

keywords:
  - RSS
  - Feeds
  - Linkblog
  - website generator

date_released: 2026-11-20
---

About this software
===================

## antennaApp 0.1.1

- `antenna page SOURCE [OUTPUT...]` writes `.html` and `.md` by extension; the `.md` is the expanded Markdown a Markdown reader gets, and the page declares it as its Markdown alternate
- New `sequences` setting in `antenna.yaml`: directories or list files whose members get previous and next links, in the HTML and the Markdown copy; the table of contents gives previous and next to posts only, and pages only through front matter or a sequence
- The `pages` table records each page's outputs and includes (upgraded in place); `antenna pages` shows six fields and `-stale` lists pages older than their source (exit 1 when there are any); `generate` rewrites every recorded output
- Includes nest: a warning past 3 levels, exit 65 past 5 or for a cycle
- `antenna.yaml` `author` is the default author of page Markdown and your own posts; `descriptionMax` sets the length `validate` flags
- `antenna validate` has new findings (`include-directive`, `no-description`, `no-author`, `title-from-heading`, `description-long`, `output-stale`, `row-output-missing`, `row-duplicate-output`, `include-missing`, `sequence-path-missing`, `sequence-empty`, `sequence-overlap`) and `-color WHEN`
- Front matter is only the block at the top of a file
- Fixed: a directive's language no longer swallows the next paragraph's first word; a code block keeps its blank line before it; `validate` no longer calls an e-mail autolink a missing file
- The documentation site is published for Markdown readers too: `index.md`, `.well-known/toc.opml` and a Markdown alternate on every page

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


