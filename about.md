---
title: antennaApp
abstract: "**antenna** is a tool for building feed oriented websites using Markdown.
If you can edit a Markdown list of links you can generate a static site
RSS Reader with little effort beyond curating your list of links.
You can create simple websites make from the content of Markdown pages.
You can create blog by posting the Markdown documents to a collection.
The goal of **antenna** is to put you in control of the web content
you read or write through using Markdown.

**antenna** automates must of the process of creating a site using Markdown
files. It handles the creation of HTML, RSS, OPML and sitemap.xml for you. This
let&#x27;s you focus on the content written in Markdown.

**antenna** supports a simple theme system defined Markdown files used to
describe page elements, a file to define page metadata and CSS files to define
the page&#x27;s layout and presentation. Each collection defined for your site
may have it&#x27;s own theme.

Features:

- Makes it easy to generate a website only using Markdown
- Makes it trivial to generate a blog, linkblog or feed reading site using Markdown
- supports as multiple feed collections per site
- provides actions to automate most of your site curation leaving you time to focus on Markdown content
- HTML, RSS 2.0 XML, OPML and sitemap.xml are generated automatically
- A preview web server is provided so you can read the curated content on your computer
- pages nice with other static site tools like  [PageFind](https://pagefind.app &quot;A browser side search engine&quot;) and
[FlatLake](https://flatlake.app &quot;A static JSON API driven by front matter in Markdown documents&quot;)"
authors:
  - family_name: Doiel
    given_name: R. S. Doiel
    id: https://orcid.org/0000-0003-0900-6903



repository_code: https://github.com/rsdoiel/antennaApp
version: 0.0.16-beta
license_url: https://www.gnu.org/licenses/agpl-3.0.en.html

programming_language:
  - Go

keywords:
  - RSS
  - Feeds
  - Linkblog
  - website generator

date_released: 2025-10-20
---

About this software
===================

## antennaApp 0.0.16-beta

- posted support for including text and code blocks based on commonMarkDoc processor, includes are processed when parsing CommonMark documents
- pages are now tracked in the default pages.md collection
- page actions are now page, pages and unpage
- sitemap files are generated for pages and posts
- posts by default go into the pages collection unless collection name is provided.

### Authors

- R. S. Doiel Doiel, <https://orcid.org/0000-0003-0900-6903>






**antenna** is a tool for building feed oriented websites using Markdown.
If you can edit a Markdown list of links you can generate a static site
RSS Reader with little effort beyond curating your list of links.
You can create simple websites make from the content of Markdown pages.
You can create blog by posting the Markdown documents to a collection.
The goal of **antenna** is to put you in control of the web content
you read or write through using Markdown.

**antenna** automates must of the process of creating a site using Markdown
files. It handles the creation of HTML, RSS, OPML and sitemap.xml for you. This
let&#x27;s you focus on the content written in Markdown.

**antenna** supports a simple theme system defined Markdown files used to
describe page elements, a file to define page metadata and CSS files to define
the page&#x27;s layout and presentation. Each collection defined for your site
may have it&#x27;s own theme.

Features:

- Makes it easy to generate a website only using Markdown
- Makes it trivial to generate a blog, linkblog or feed reading site using Markdown
- supports as multiple feed collections per site
- provides actions to automate most of your site curation leaving you time to focus on Markdown content
- HTML, RSS 2.0 XML, OPML and sitemap.xml are generated automatically
- A preview web server is provided so you can read the curated content on your computer
- pages nice with other static site tools like  [PageFind](https://pagefind.app &quot;A browser side search engine&quot;) and
[FlatLake](https://flatlake.app &quot;A static JSON API driven by front matter in Markdown documents&quot;)

- License: <https://www.gnu.org/licenses/agpl-3.0.en.html>
- GitHub: <https://github.com/rsdoiel/antennaApp>
- Issues: <https://github.com/rsdoiel/antennaApp/issues>

### Programming languages

- Go




### Software Requirements

- Go >= 1.25.3
- CMTools >= 0.0.40


### Software Suggestions

- GNU Make &gt;&#x3D; 3.4
- Pandoc &gt;&#x3D; 3.1
- Bash or Powershell


