---
title: Build a Blog with Antenna
dateCreated: "2025-09-05"
dateModified: "2026-06-27"
datePublished: "2025-09-05"
author: R. S. Doiel
keywords:
  - Antenna
  - blog
  - tutorial
postPath: "blog/2025/09/05/Build_a_Blog_with_Antenna.html"
---

# Build a Blog with Antenna

By R. S. Doiel, 2025-09-05

Antenna is a feed-oriented content management tool. This tutorial walks through building a simple blog: configuring the blog, posting content in Markdown, generating HTML pages and an RSS feed, and adding supporting static pages.

## Setting up

Create a directory for the website and change into it.

~~~shell
mkdir myblog
cd myblog
~~~

## Initializing the blog

The `antenna init` action creates the two configuration files the blog needs.

~~~shell
antenna init
~~~

This produces:

- **`antenna.yaml`** — the main Antenna configuration (collections, base URL, htdocs path)
- **`page.yaml`** — the HTML page generator (stylesheet links, nav, header, footer, scripts)

After initializing, generate a default stylesheet. This writes `css/site.css` with sensible typography, dark-mode support, navigation styles, and the skip-navigation link required for keyboard accessibility. It also adds the stylesheet reference to `page.yaml`.

~~~shell
antenna css
~~~

## Defining the blog collection

A blog is built from three things:

- Markdown files holding individual posts
- An HTML page listing posts in reverse chronological order
- An RSS feed of recent posts

These are managed inside an Antenna *collection*. A collection is defined by a Markdown file whose front matter and body describe the collection itself. The collection's name determines the output filenames — `index.md` produces `index.html`, `index.xml`, and `index.opml`.

Create `index.md` with this content:

~~~markdown
---
title: My Blog
description: A simple blog built with Antenna App.
---

# Welcome to My Blog

Posts appear below in reverse chronological order.
~~~

Add it to Antenna once:

~~~shell
antenna add index.md
~~~

This creates `index.db` (the SQLite3 database that tracks posts) and, if needed, updates `page.yaml`.

## Adding the first post

Posts live in a date-based directory tree under a `blog/` folder. For a post on September 5, 2025:

~~~shell
mkdir -p blog/2025/09/05
~~~

On Windows:

~~~shell
New-Item -ItemType Directory -Force -Path blog\2025\09\05
~~~

Create `blog/2025/09/05/welcome.md`. The required front matter fields are `title`, `postPath`, and `pubDate`:

~~~markdown
---
title: Welcome
description: The first post on my new blog.
author: Your Name
keywords:
  - welcome
  - announcement
pubDate: "2025-09-05"
postPath: "blog/2025/09/05/welcome.html"
---

# Welcome

This is a demonstration of blogging with Antenna App.
~~~

The `title`, `description`, `author`, and `keywords` fields are written into the generated HTML as `<meta>` elements, which search engines and the [PageFind](https://pagefind.app) site-search index can use.

Post it to the `index.md` collection:

~~~shell
antenna post index.md blog/2025/09/05/welcome.md
~~~

On Windows, use backslash paths for the Markdown file argument.

Antenna generates `blog/2025/09/05/welcome.html` and records the post in `index.db`.

Now rebuild the collection page and RSS feed:

~~~shell
antenna generate
~~~

Preview the result at `http://localhost:8000`:

~~~shell
antenna preview
~~~

## Updating a post

Re-run `antenna post` on the same file to update it. Antenna matches by `postPath` and overwrites the record:

~~~shell
antenna post index.md blog/2025/09/05/welcome.md
antenna generate
antenna preview
~~~

## Listing posts, pages, and items

Check what Antenna has recorded at any time:

~~~shell
# Blog posts in the index.md collection
antenna posts index.md

# Static pages (tracked separately in pages.db)
antenna pages

# All items in a collection, including harvested feed items
antenna items index.md

# All defined collections
antenna list
~~~

## Adding static pages

Static pages — an About page, a contact page, a series index — are Markdown files that are not part of any post collection. Use `antenna page` to render them:

~~~markdown
---
title: About
description: About the author of this blog.
author: Your Name
---

# About

I write about technology and other things I find interesting.
~~~

Save as `about.md` and render it:

~~~shell
antenna page about.md
~~~

This generates `about.html` next to `about.md` and records it in the pages database. Re-run `antenna page about.md` whenever `about.md` changes.

## Front matter metadata

Every front matter field in a post or page is written into the generated HTML `<head>` as a standard `<meta name="…" content="…">` element and a matching `data-pagefind-filter` attribute for [PageFind](https://pagefind.app) faceted search. Common useful fields:

| Field | Purpose |
|-------|---------|
| `title` | Sets the `<title>` element; not emitted as a `<meta>` |
| `description` | Short summary for search engines |
| `author` | Author name |
| `keywords` | List of tags; each value gets its own `<meta>` pair |
| `series` | Series name for multi-part posts |
| `seriesNumber` | Position in the series |
| `datePublished` | Publication date (`YYYY-MM-DD`) |
| `dateModified` | Last-modified date |

### Controlling which fields are published

By default all front matter fields are emitted as metadata. If some fields are internal (build flags, workflow notes) you can restrict publication to an explicit allowlist in `page.yaml`:

~~~yaml
allowed_meta_fields:
  - title
  - author
  - keywords
  - description
  - series
  - seriesNumber
~~~

When `allowed_meta_fields` is set, only those keys appear in the generated HTML; all other front matter fields are silently omitted.

## Enhancing the blog

Open `page.yaml` and customise the HTML shell. The `antenna css` command already added the stylesheet link; here is a fuller example:

~~~yaml
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
  <p>© 2025 Your Name</p>
~~~

`lang` sets the `lang` attribute on the `<html>` element. Change it for non-English sites (e.g. `lang: fr-FR`, `lang: ja`).

The `nav` HTML is wrapped in `<nav aria-label="Site navigation">` automatically. The page already includes a *skip to main content* link before the nav for keyboard accessibility — the `css/site.css` generated by `antenna css` hides it off-screen until it receives keyboard focus.

After editing `page.yaml`, re-post and regenerate:

~~~shell
antenna post index.md blog/2025/09/05/welcome.md
antenna generate
antenna preview
~~~

## Publishing

When the blog looks right in preview, publish to your host using whatever tool it provides. For GitHub Pages:

~~~shell
git add -A
git commit -m "Initial blog posts"
git push origin main
~~~

Happy blogging!
