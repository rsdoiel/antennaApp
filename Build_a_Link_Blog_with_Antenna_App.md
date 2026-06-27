---
title: Build a Link Blog with Antenna
dateCreated: "2026-06-27"
dateModified: "2026-06-27"
datePublished: "2026-06-27"
author: R. S. Doiel
keywords:
  - Antenna
  - link blog
  - tutorial
postPath: "blog/2026/06/27/Build_a_Link_Blog_with_Antenna.html"
---

# Build a Link Blog with Antenna

By R. S. Doiel, 2026-06-27

A link blog is a website where you share links to things you find interesting on the web — and optionally respond to them with commentary of your own. This tutorial shows how to build one with Antenna: subscribing to feeds, harvesting their content, and writing responses that appear alongside the harvested items.

For a simpler starting point, see [Build a Blog with Antenna](Build_a_Blog_with_Antenna_App.md), which covers posting your own content without feed subscriptions.

## Setting up

Create a directory for the site and change into it.

~~~shell
mkdir mylinkblog
cd mylinkblog
~~~

## Initializing

~~~shell
antenna init
antenna css
~~~

`init` creates `antenna.yaml` and `page.yaml`. `css` writes `css/site.css` and links it in `page.yaml`.

## Defining your reading list

A link blog is built around a *collection* — a Markdown file whose body lists the RSS/Atom feeds you want to follow. The collection's name determines the output filenames: `links.md` produces `links.html` (the aggregated reading page) and `links.xml` (your own RSS feed).

Create `links.md`:

~~~markdown
---
title: My Link Blog
description: Links and commentary from around the web.
---

# My Link Blog

Reading list below. Posts appear in reverse chronological order.

- [Dave Winer's Scripting News](http://scripting.com/rss.xml)
- [R. S. Doiel's blog](https://rsdoiel.github.io/rss.xml)
~~~

The list in the body is the subscription list. You can add or remove feeds by editing this file at any time.

Register the collection with Antenna:

~~~shell
antenna add links.md
~~~

This creates `links.db` (the SQLite3 database for the collection) and adds it to `antenna.yaml`.

## Harvesting content

Pull content from all subscribed feeds into `links.db`:

~~~shell
antenna harvest
~~~

Antenna fetches each feed in the list and stores the items locally. On a slow connection or with many feeds, the first harvest may take a moment. Subsequent harvests are incremental.

Now render the HTML page and RSS feed:

~~~shell
antenna generate
antenna preview
~~~

Open `http://localhost:8000/links.html` to see the aggregated content. Each item shows the title, publication date, excerpt, and a link back to the original source.

## Reviewing what was harvested

Check what is in the collection at any time:

~~~shell
antenna items links.md
~~~

This prints a Markdown list of every item in `links.db` — harvested feed entries and any posts you have added — in descending date order.

## Writing a response

The most distinctive feature of a link blog is the ability to respond to something you read. Antenna's `quote` action converts a [text-fragment URL](https://web.dev/text-fragments/) into a Markdown blockquote with source attribution, which you can use as the start of a response post.

### Getting a text-fragment URL

In a modern browser, select the passage you want to quote, right-click it, and choose **Copy link to highlight** (Chrome/Edge) or equivalent. This copies a URL ending in `#:~:text=the+selected+text`.

### Generating the blockquote

~~~shell
antenna quote "https://scripting.com/2026/06/27/example.html#:~:text=the+selected+passage" > response.md
~~~

`antenna quote` writes the blockquote and attribution to standard output. Redirecting to `response.md` gives you a file to edit.

The generated content looks like:

~~~markdown
> the selected passage

([scripting.com](https://scripting.com/2026/06/27/example.html#:~:text=the+selected+passage), accessed 2026-06-27)
~~~

### Adding your commentary

Open `response.md` in your editor and add a YAML front matter block and your own text:

~~~markdown
---
title: Thoughts on Dave's point about RSS
description: My take on Dave Winer's argument for RSS as social infrastructure.
author: Your Name
keywords:
  - RSS
  - social web
link: https://scripting.com/2026/06/27/example.html
pubDate: "2026-06-27"
postPath: "blog/2026/06/27/thoughts-on-dave.html"
---

# Thoughts on Dave's point about RSS

> the selected passage

([scripting.com](https://scripting.com/2026/06/27/example.html#:~:text=the+selected+passage), accessed 2026-06-27)

My commentary goes here. I agree with Dave that...
~~~

The `link` field in the front matter points to the original article. The `postPath` determines where Antenna writes the rendered HTML.

Create the output directory and post the response to the `links.md` collection:

~~~shell
mkdir -p blog/2026/06/27
antenna post links.md response.md
~~~

Antenna renders `blog/2026/06/27/thoughts-on-dave.html` and records the post in `links.db`. Regenerate and preview:

~~~shell
antenna generate
antenna preview
~~~

Your response now appears in `links.html` alongside the harvested items, with the `link` field providing a connection back to the original article.

## Listing your responses

To see only posts you have written (items with a `postPath`), use:

~~~shell
antenna posts links.md
~~~

To see everything in the collection — harvested items and your own posts:

~~~shell
antenna items links.md
~~~

## The daily update cycle

A link blog stays current by re-harvesting regularly. The typical cycle is:

~~~shell
antenna harvest
antenna generate
antenna preview
~~~

Run this whenever you want to pull in new content. You can automate it with a cron job or a task scheduler:

~~~shell
# Example cron entry — harvest and generate every hour
0 * * * * cd /path/to/mylinkblog && antenna harvest && antenna generate
~~~

## Customizing the page

Open `page.yaml` to personalise the HTML shell. A fuller example for a link blog:

~~~yaml
lang: en-US

link:
  - rel: stylesheet
    type: text/css
    href: /css/site.css
  - rel: alternate
    type: application/rss+xml
    title: My Link Blog RSS
    href: /links.xml

header: |
  <h1>My Link Blog</h1>

nav: |
  <ul>
    <li><a href="/links.html">Links</a></li>
    <li><a href="/about.html">About</a></li>
  </ul>

footer: |
  <p>© 2026 Your Name — <a href="/links.xml">RSS</a></p>
~~~

The `link` element pointing to `/links.xml` adds an autodiscovery tag so feed readers can find your RSS feed automatically.

### Adding an About page

~~~shell
antenna page about.md
~~~

See [Build a Blog with Antenna](Build_a_Blog_with_Antenna_App.md) for front matter metadata fields and the `allowed_meta_fields` allowlist.

## Publishing

When the site looks right locally:

~~~shell
git add -A
git commit -m "Initial link blog"
git push origin main
~~~

After publishing, share your feed URL (`https://yourdomain.example/links.xml`) so others can subscribe to your link blog.
