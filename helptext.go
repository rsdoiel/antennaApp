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
*/
package antennaApp

var (
	HelpText = `%{app_name}(1) user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}

# NAME

{app_name}

# SYNOPSIS

{app_name} [OPTIONS] ACTION [PARAMETERS]

{app_name} help [TOPIC]

# DESCRIPTION

**antenna** is a tool for working with RSS feeds and rendering a link blog.
It is inspired by Dave Winer's [Textcasting](https://textcasting.org) and
[FeedLand](https://feedland.org) and my own experimental website,
[antenna](https://rsdoiel.github.io/antenna).

The approach I am taking is to make it easy to curate feeds and generated a static
website using a simple command line tool. I believe that a link blog which can
consume and generate RSS can be a basis for a truly distributed social web.
It avoids the complexity of solutions like ATProto and ActivityPub.

Features:

- support for multiple collections of feeds
- a collection is defined by a Markdown document containing a list of links to feeds
- collections can be harvested, meaning content retrieved from the feeds listed in the Markdown document
- harvested content is stored in a SQLite3 database
- harvested content in a collection can be aggregated and rendered as an HTML page for reading
- Markdown documents can be imported into a collection as a feed item
- RSS 2.0 XML and HTML are generated per collection
- A preview feature to view the render content in your web browser via a localhost URL
- You can manage your collections via a localhost URL too.

The ability to harvest feed items means we can read what others post on the web. The Markdown content
can be added to a feed allows us to comment on the items read (thus being social).

Through YAML configuration files you can customize the HTML rendered by **antenna** on a per
collection basis. That means it is possible to recreate a "news paper" like experience. 

A static website using **antenna** can grow through either enhancing the HTML markup defined
in the YAML configuration or through manipulation of the collection contents in the SQLite3 database.
This provides opportunities to integrate with other static website tools like
[PageFind](https://pagefind.app "A browser side search engine") and
[FlatLake](https://flatlake.app "A static JSON API driven by front matter in Markdown documents").
You can even use **{app_name}** to augment your existing blog.

# OPTIONS

--help, help
: display help

--license, license
: display license

--version, version
: display version

-config FILENAME
: Use specified YAML file for configuration

# ACTION

Actions form the primary way you use Antenna and manage a link blog through its life cycle. Below
is a list of supported ACTION and their purpose. ACTION can be split between two general purposes
. The following commands are related to curating your Antenna's collections of feeds and feed items.

init
: Initialize an Antenna instances by creating a YAML configuration file or validating one. The file
generated is called "{app_name}.yaml". It'll create a Markdown document called pages.md (if none exists)
and a pages.db SQLite3 file to hold the default collection. Collections hold posts and pages metadata.
Posts will be be included in RSS feed output while pages will be excluded from feed output.

add COLLECTION_FILE [NAME DESCRIPTION]
: Add the feed collection defined by COLLECTION_FILE to your Antenna configuration.
COLLECTION_FILE is a Markdown document (.md) whose body contains a list of hyperlinks
to RSS/Atom feeds. You can supply an optional NAME and DESCRIPTION, or define them
in the document's YAML front matter.

del COLLECTION_FILE
: Remove a collection from the Antenna configuration. COLLECTION_FILE is the
Markdown (.md) filename as registered with the add action.

list
: List the collection filenames defined in the "{app_name}.yaml".

page INPUT_PATH [OUTPUT_PATH]
: This will create a standalone HTML page in a collection called pages.md. The pages.md
is the only collection that has pages (hence the name). It is also the default collection
(created by the init action). It uses the default page generator defined in the
{app_name}.yaml if one is not specifically set for the pages.db collection. INPUT_PATH
is a Markdown document (.md). The resulting HTML file uses the same basename with a
.html extension. If OUTPUT_PATH is set it uses that name for the HTML file generated.
(NOTE: pages are not shown in the RSS feed. The page action is useful for pages like an
about page, home page, and search page. __The Markdown processed via the page action
will allow "unsafe" HTML to pass through. Only use page with files you trust
completely!!!__)

unpage INPUT_PATH
: This will will remove a page's from a collection based in the input filepath provided.
It does not remove the page on disk, just from the collection so that it will no longer
be used to create a corresponding HTML page when the generate action is run.

pages
: List the page entries in the pages.md collection. Pages are ordered by
descending updated timestamp.


post [COLLECTION_NAME] FILEPATH
: Add a Markdown document (.md) to a feed collection (default is pages.md). The front
matter is used to specify things like the link to the post, guid, description, etc. If
these are not provided then the post action will display an error and not write the
content to the post directory location or add it to the collection. Required front
matter: **title** or **description**, see RSS 2.0 at
<https://cyber.harvard.edu/rss/rss.html#hrelementsOfLtitemgt>. To include the file in
the posts directory tree you need to provide a **postPath**. In that case it is also
recommended you provide a value for **link** that reflects the public URL to where the
post can be viewed. Like the page action, unsafe HTML passes through unchanged; only
use post with files you trust completely.

blogit [COLLECTION_NAME] FILEPATH [POST_DATE]
: This is a variation of post where FILEPATH is used as the source Markdown document (.md)
to be placed in a blog-style date directory path (e.g. blog/2026/04/12/my-post.md).
After calculating the target path and copying the file there it uses post to finish
adding it to the collection.

unpost COLLECTION_NAME URL | POST_PATH
: Remove an item from a collection using the URL associated with the item. You can
provide either the full URL or the POST_PATH value to trigger the removal.

posts COLLECTION_NAME [COUNT | FROM_DATE TO_DATE]
: List the posts in a collection expressed as a Markdown list. The post
must contain both a pubDate value and postPath value. The list is in descending
pubDate order. The list includes a Markdown link made of the title and postPath
followed by the pubDate. If you provide the optional elements then the list will
be constrained by a count or time range. COUNT  is an integer, FROM_DATE and TO_DATE
are dates in the YYYY-MM-DD format.

rss COLLECTION_NAME RSS_FILENAME [COUNT | FROM_DATE TO_DATE]
: Generate an RSS feed from posts. The optional parameters are applied
like the posts action.

quote TEXT_FRAGMENT_URL
: This will parse a TEXT_FRAGMENT_URL into a Markdown text. The text is
written to standard out. You can redirect this into a file. The purpose of
the "quote" action is to simply quoting another site for use in a post.

The following commands are related to producing a link blog static website.

harvest [COLLECTION_NAME]
: The harvest retrieves feed content. If COLLECTION_NAME is provided then only the 
the single collection will be harvested otherwise all collections defined in your
Antenna YAML configuration are harvested.

generate [COLLECTION_NAME]
: This process the collections rendering HTML pages and RSS 2.0 feeds for each collection.
If the collection name is provided then only that HTML page will be generated.

sitemap
: This will generate a set of sitemap files for pages and posts found through the
{app_name}.yaml file. (e.g. sitemap_index.xml, sitemap_1.xml, sitemap_2.xml, ...)

preview
: Let's your preview the rendered your Antenna instance as a localhost website using
your favorite web browser.

themes [new [THEME_NAME]]
: Without arguments, lists the theme directories detected in the project. With "new",
creates a skeleton theme directory named THEME_NAME (default: "theme") containing
header.md, nav.md, footer.md, and head.yaml stub files. Existing files are never
overwritten. After editing the skeleton files, apply the theme with the "apply" action.

apply THEME_PATH [YAML_FILE_PATH]
: This will apply the content THEME_PATH and update the YAML generator file described
by YAML_FILE_PATH. If YAML_FILE_PATH is not provided then that YAML generator file
will be replaced by the theme.

interactive [ACTION [PARAMETERS]]
: Start a guided conversation that walks you through any antenna action step by
step. Each parameter is explained and pre-filled from arguments already on the
command line. The complete command is shown before it runs, making this a useful
way to learn the antenna command syntax. If ACTION is omitted you are presented
with a menu of all available actions to choose from.

stylefrom INPUT_FILE [OUTPUT_PATH]
: Extract CSS from a LibreOffice Writer HTML export (.html, .htm) and write it to
OUTPUT_PATH. OUTPUT_PATH defaults to "theme/style.css" when omitted. The directory
is created if it does not exist. This makes it easy to seed a theme stylesheet
directly from a styled LibreOffice Writer document.

css [CSS_PATH]
: Write a comprehensive default stylesheet to CSS_PATH within the htdocs
directory (default: css/site.css). The stylesheet includes CSS custom properties,
dark-mode support, a skip-navigation link (WCAG 2.4.1), and styles for all HTML
structures generated by {app_name}. An existing file is backed up to CSS_PATH.bak.
After writing the CSS the configured page generator YAML (page.yaml) is patched to
add a link: entry referencing the new stylesheet. Use 'antenna help css' for full
details.

items [COLLECTION_NAME]
: List all items stored in the named collection's SQLite3 database as a Markdown
list in descending publication-date order. Items include blog posts, static pages,
and feed entries harvested from remote feeds. COLLECTION_NAME defaults to pages.md.
Compare with 'posts' (items with a postPath only) and 'pages' (static pages only).

help [TOPIC]
: Display help. Without TOPIC the full manual page is shown. With TOPIC the guide
for that specific command or concept is shown. Run 'antenna help topics' for a list
of all available topics.

# CONFIGURATION

**{app_name}** uses a YAML configuration file. Below the the primary attributes you can
set in the YAML file.

port
: (optional, default: 8000) The on host to listen on when running the "preview" action

host
: (optional, default is localhost) The host name to use with port when running the "preview" action.

htdocs
: (optional, default: ".") The directory that will hold  the HTML and RSS files rendered
by the "generate" action or viewed with "preview" action in your web browser.

generator
; (optional, default: "page.yaml")  This names the YAML used to describe an HTML page
structure. It is created on initialization. You can use a custom one per collection.

collections
: (required) This holds a list of collections managed by **{app_name}**. For **{app_name}**
to be useful you need to define at least one collection.

The collections attribute holds a collection objects. Each collection object has
the following attributes.

title
: (optional, defaults to the file name minus extension) name of the collection

file
: (required), the path to the Markdown file defining the collection

generator
: (optional) The YAML configuration filename to used to render the collection
as HTML.

mode
: (optional, default: "aggregate") Controls how the collection is rendered as HTML.
"aggregate" renders feed-item cards from the items table (the default behaviour for
link blogs and feed aggregators). "page-index" renders a simple link list from the
pages table, suitable for collections like pages.md that track static HTML pages
rather than harvested feed items.

# EXAMPLES

Here's an example of using **{app_name}** to create a single collection static site.

Step 1. Create a markdown document called "example.md" with the following text.

~~~markdown
---
title: Example Collection
---

This is a an example of a Markdown document used to defined a 
collection of feeds. The title in the front matter will be 
used as the collection's title. The collection itself will be
populated from the feed list below.

- [R. S. Doiel's blog](https://rsdoiel.github.io/rss.xml)
- Dave Winer's [scripting.com](http://scripting.com/rss.xml)
~~~

Once have defined a collection we can create an Antenna instance.

Steps are init, add our collection, harvest, generate then preview.

~~~shell
{app_name} init
{app_name} add example.md
{app_name} harvest
{app_name} generate
{app_name} preview
~~~

The "preview" action runs a localhost web server so you can read the
contents in the generate HTML page called "example.html". You'd open
your web browsers to <http://localhost:8000/example.html> to review
the harvested content.

To update your static site you'd just do the following

~~~shell
{app_name} harvest
{app_name} generate
{app_name} preview
open http://localhost:8000
~~~

When you run the "generate" action HTML files and RSS feeds will
be written to the directory designated in the **{app_name}** YAML
configuration file (defaults to to the current working directory, "").
The "preview" action  serves that out over localhost (default port 8000)
so you can read your static site with your favorite web browser.

# Also 

- [antenna-themes (7)](antenna-themes.7.md)

`

	ThemeHelpText = `%{app_name}(7) user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}

# NAME

{app_name} themes

# SYNOPSIS

A directory with page elements in Markdown or YAML

# DESCRIPTION

A directory with files that can be used to generate an {app_name} page generator
description. The {app_name} uses a page generator description YAML file to render
HTML pages. The YAML structure is organized around those elements that are in the
HTML head element as well as the body elements of HTML pages.

A theme is held in a directory. The directory name is used as the theme's name.
Inside the directory are zero or more files where their names map the YAML attribute
names in a page generator YAML file. Here is an example of a theme called "theme"
that can be applied to generate a generator YAML file.

~~~
theme\header.md
theme\nav.md
theme\top_content.md
theme\bottom_content.md
theme\footer.md
theme\head.yaml
theme\style.css
~~~

The following Markdown documents are used to express their related attributes in the
page generator YAML files. Markdown is used to express the HTML values that will be
used in the page generator file for these attributes. The elements describe form
the innerHTML of the body element in an HTML document. They are rendered in the
order presented if they are present.

header.md
: (optional, used when present) This Markdown document contains a Markdown
expressing of the innerHTML of a header HTML element.

nav.md
: (optional, used when present) This Markdown document contains a Markdown
expressing of the innerHTML of a nav HTML element.

top_content.md
: (optional, used when present) This Markdown document contains a Markdown
expressing the HTML that will appear after the nav element and before a section
element if present.

bottom_content.md
: (optional, used when present) This Markdown document contains a Markdown
expressing the innerHTML that will appear after section element and before
the footer element.

footer.md
: (optional, used when present) This Markdown document contains a Markdown
expressing of the innerHTML of a footer HTML element. It is rendered before
closing the body element.

The head element's content may also be included in a theme. It is expressed as a
YAML file called "head.yaml". YAML is used because there 
is not a direct relationship between the element attributes and how they could be expressed
using Markdown. Most of the time the head.yaml isn't necessary in the theme because 
{app_name} generates most of the head elements' content automatically. There are times when
my wish to enhance the generated content (e.g. include link elements pointing to files or
include script elements JavaScript). The head element's innerHTML is populated in the order of
meta elements, link elements and script elements if they are defined in the YAML as the 
attributes meta, link and script. Each of these top level YAML elements are list and the
individual items in the list express the attribute names and values that form that element.

title
: (optional, used when present) A page title represented as a string.

meta
: (optional, used when present) A list of objects expressing a sequence of meta 
HTML elements attributes. Each item in the list is formed from the attribute names
and values that are define in a meta element. See 
<https://developer.mozilla.org/en-US/docs/Web/HTML/Reference/Elements/meta>

link
: (optional, used when present) A list of objects expressing a sequence of link 
HTML elements attributes. Each item in the list is formed from the attribute names
and values that are defined in a link element. See
<https://developer.mozilla.org/en-US/docs/Web/HTML/Reference/Elements/link>

script
: (optional, used when present) A list of objects expressing a sequence of script 
HTML elements attributes. Each item in the list is formed from the attribute names
and values that are defined in a script element. See
https://developer.mozilla.org/en-US/docs/Web/HTML/Reference/Elements/script

style
: (optional, used when present) A string holding CSS to be injected as the last
element of the head when rendering HTML. If you have styled a LibreOffice Writer document you can use the {app_name} stylefrom action to extract that into a stylesheet.

Here is an example "head.yaml"

~~~yaml
title: My theme based title
meta:
  - charset: utf-8
  - name: language
    content: en-US
link:
  - rel: alternate
    type: application/rss+xml
	href: archive.xml
  - rel: stylesheet
    href: /css/site.css
script:
  - type: module
    src: modules/myscript.js
style: |+
  /* This CSS will turn headings vertical */
  h1 {
    writing-mode: vertical-rl;
    transform: rotate(180deg);
    text-orientation: mixed;
  }

~~~

NOTE: In this example the last style element will override the H1 definitions
previously included in the CSS files using the link attributes.

# Also 

- [antenna (1)](antenna.1.md)

`

	// Command help topics

	AddHelpText = `%{app_name}(7) user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}

# NAME

add

# SYNOPSIS

{app_name} add COLLECTION_FILE [NAME DESCRIPTION]

# DESCRIPTION

Registers a feed collection with {app_name}. COLLECTION_FILE is a Markdown
document (.md) whose body lists RSS/Atom feed URLs as hyperlinks. NAME and
DESCRIPTION are optional; if omitted, {app_name} reads them from the
document's YAML front matter.

{app_name} creates a matching SQLite3 database (same basename, .db extension)
and records the collection in antenna.yaml.

# PARAMETERS

COLLECTION_FILE
: path to the Markdown document defining the collection

NAME
: (optional) override the collection name

DESCRIPTION
: (optional) override the collection description

# EXAMPLES

{app_name} add feeds/tech.md
{app_name} add feeds/tech.md "Tech Feeds" "My technology reading list"

`

	ApplyHelpText = `%{app_name}(7) user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}

# NAME

apply

# SYNOPSIS

{app_name} apply THEME_PATH [YAML_FILE_PATH]

# DESCRIPTION

Applies the theme at THEME_PATH to the page generator YAML file at
YAML_FILE_PATH. If YAML_FILE_PATH is omitted, the default generator YAML
(page.yaml) is replaced.

A theme directory contains Markdown and YAML files whose names map to
generator YAML attributes: header.md, nav.md, footer.md, head.yaml, etc.
Run '{app_name} help themes' for the full theme directory layout.

# PARAMETERS

THEME_PATH
: path to the theme directory

YAML_FILE_PATH
: (optional) path to the generator YAML to update

# EXAMPLES

{app_name} apply theme/my-theme
{app_name} apply theme/my-theme collection/feeds.yaml

`

	BlogitHelpText = `%{app_name}(7) user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}

# NAME

blogit

# SYNOPSIS

{app_name} blogit [COLLECTION_NAME] FILEPATH [POST_DATE]

# DESCRIPTION

A variation of 'post' that places the source Markdown document in a
blog-style date directory tree (e.g. blog/2026/04/12/my-post.md) before
adding it to the collection. If POST_DATE is omitted, the current date is
used. The collection defaults to pages.md when COLLECTION_NAME is omitted.

# PARAMETERS

COLLECTION_NAME
: (optional) path to the collection Markdown file

FILEPATH
: source Markdown document

POST_DATE
: (optional) date in YYYY-MM-DD format

# EXAMPLES

{app_name} blogit index.md notes/my-idea.md
{app_name} blogit index.md notes/my-idea.md 2026-04-01

`

	CssHelpText = `%{app_name}(7) user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}

# NAME

css

# SYNOPSIS

{app_name} css [CSS_PATH]

# DESCRIPTION

Writes a comprehensive starter stylesheet to CSS_PATH within the htdocs
directory configured in antenna.yaml. If CSS_PATH is omitted it defaults to
css/site.css. Directory levels are created automatically.

If a stylesheet already exists at the target path it is backed up to
CSS_PATH.bak before being overwritten.

After writing the CSS, {app_name} patches the generator YAML (page.yaml) to
add a link entry referencing the new file. If page.yaml already has a link
section, {app_name} prints instructions for adding the entry by hand.

The generated stylesheet includes CSS custom properties for colors, fonts,
and layout, dark-mode support, skip-navigation link, navigation bar, article
cards, standalone pages, site footer, and typography.

# PARAMETERS

CSS_PATH
: (optional) path relative to htdocs (default: css/site.css)

# SEE ALSO

{app_name} help accessibility
{app_name} help themes

`

	DelHelpText = `%{app_name}(7) user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}

# NAME

del

# SYNOPSIS

{app_name} del COLLECTION_FILE

# DESCRIPTION

Removes the collection identified by COLLECTION_FILE from antenna.yaml.
The Markdown file and its SQLite3 database are not deleted from disk.

# PARAMETERS

COLLECTION_FILE
: path to the Markdown document as registered with add

# EXAMPLES

{app_name} del feeds/tech.md

`

	GenerateHelpText = `%{app_name}(7) user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}

# NAME

generate

# SYNOPSIS

{app_name} generate [COLLECTION_NAME]

# DESCRIPTION

Processes all collections (or only COLLECTION_NAME if provided), rendering
an HTML page and RSS 2.0 feed for each. Output files are written to the
htdocs directory configured in antenna.yaml.

The HTML structure is controlled by the page generator YAML (page.yaml or
a per-collection override). Front matter from each item is emitted as
meta elements in the generated HTML.

# PARAMETERS

COLLECTION_NAME
: (optional) process only this collection

# ALIASES

build

# EXAMPLES

{app_name} generate
{app_name} generate index.md

`

	HarvestHelpText = `%{app_name}(7) user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}

# NAME

harvest

# SYNOPSIS

{app_name} harvest [COLLECTION_NAME]

# DESCRIPTION

Retrieves RSS/Atom feed content for all collections (or only
COLLECTION_NAME) and stores harvested items in each collection's SQLite3
database. Run generate afterwards to rebuild the HTML pages.

# PARAMETERS

COLLECTION_NAME
: (optional) harvest only this collection

# ALIASES

fetch

# EXAMPLES

{app_name} harvest
{app_name} harvest feeds/tech.md

`

	InitHelpText = `%{app_name}(7) user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}

# NAME

init

# SYNOPSIS

{app_name} init

# DESCRIPTION

Creates the two configuration files {app_name} needs if they do not exist:

  antenna.yaml  main configuration (htdocs path, port, collections list)
  page.yaml     page generator (link, meta, nav, header, footer, scripts)

Also creates a default pages.md collection and pages.db database.

After running init, run '{app_name} css' to generate a starter stylesheet
and automatically link it in page.yaml.

# EXAMPLES

mkdir myblog && cd myblog
{app_name} init
{app_name} css

`

	InteractiveHelpText = `%{app_name}(7) user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}

# NAME

interactive

# SYNOPSIS

{app_name} interactive [ACTION [PARAMETERS]]

# DESCRIPTION

Starts a menu-driven session that walks through any {app_name} action step
by step. Each parameter is explained and pre-filled from arguments already
on the command line. The complete command is shown before it runs.

If ACTION is omitted a menu of all available actions is presented.
Useful for learning the command syntax interactively.

# ALIASES

tui

# EXAMPLES

{app_name} interactive
{app_name} interactive post

`

	ItemsHelpText = `%{app_name}(7) user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}

# NAME

items

# SYNOPSIS

{app_name} items [COLLECTION_NAME]

# DESCRIPTION

Prints a Markdown list of all items stored in the named collection's
SQLite3 database. Items include blog posts added via post or blogit,
and feed entries harvested via harvest. The list is in descending
publication-date order.

Each line shows: - [Title](link), YYYY-MM-DD, status (label)

If COLLECTION_NAME is omitted it defaults to pages.md.

See posts for a list restricted to blog posts (items with a postPath),
and pages for a list of static page entries.

# PARAMETERS

COLLECTION_NAME
: (optional) collection Markdown file (default: pages.md)

# EXAMPLES

{app_name} items
{app_name} items index.md

`

	ListHelpText = `%{app_name}(7) user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}

# NAME

list

# SYNOPSIS

{app_name} list

# DESCRIPTION

Prints the collection Markdown filenames registered in antenna.yaml.
Each entry is the path passed to add.

# EXAMPLES

{app_name} list

`

	PageHelpText = `%{app_name}(7) user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}

# NAME

page

# SYNOPSIS

{app_name} page INPUT_PATH [OUTPUT_PATH]

# DESCRIPTION

Renders INPUT_PATH as an HTML file using the page generator YAML. The
output file is placed next to INPUT_PATH with a .html extension unless
OUTPUT_PATH is specified. The page is recorded in the pages collection
(pages.md / pages.db) so it appears in pages output.

Pages are excluded from RSS feeds. This action is for static content such
as About pages, contact pages, and search pages.

WARNING: HTML in the Markdown source passes through unchanged (unsafe mode).
Only run this action on files you control and trust.

# PARAMETERS

INPUT_PATH
: path to the source Markdown file

OUTPUT_PATH
: (optional) explicit output HTML path

# EXAMPLES

{app_name} page about.md
{app_name} page about.md htdocs/about.html

`

	PagesHelpText = `%{app_name}(7) user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}

# NAME

pages

# SYNOPSIS

{app_name} pages

# DESCRIPTION

Prints the page entries tracked in the pages collection (pages.md / pages.db)
in descending updated-timestamp order.

Compare with posts (blog posts in a collection) and items (all items
in any collection database).

# EXAMPLES

{app_name} pages

`

	PostHelpText = `%{app_name}(7) user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}

# NAME

post

# SYNOPSIS

{app_name} post [COLLECTION_NAME] FILEPATH

# DESCRIPTION

Adds FILEPATH to COLLECTION_NAME (default: pages.md). The file's YAML
front matter supplies the required fields. If a record with the same
postPath already exists, it is overwritten.

Required front matter fields:
  title or description  at least one must be present
  postPath              relative path to the generated HTML file
  pubDate               publication date (YYYY-MM-DD recommended)

Recommended additional fields:
  link        public URL to the post
  author      author name
  description summary for RSS and search engines
  keywords    list of tags

After posting, run generate to rebuild the collection HTML and RSS feed.

WARNING: HTML in the Markdown source passes through unchanged (unsafe mode).
Only run this action on files you control and trust.

# PARAMETERS

COLLECTION_NAME
: (optional) collection Markdown file (default: pages.md)

FILEPATH
: path to the source Markdown document

# EXAMPLES

{app_name} post index.md blog/2026/04/12/my-post.md

# SEE ALSO

{app_name} help metadata
{app_name} help blogit

`

	PostsHelpText = `%{app_name}(7) user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}

# NAME

posts

# SYNOPSIS

{app_name} posts COLLECTION_NAME [COUNT | FROM_DATE TO_DATE]

# DESCRIPTION

Prints a Markdown list of blog posts in COLLECTION_NAME in descending
pubDate order. Items must have both a pubDate and a postPath to appear.

Optional parameters constrain the list:
  COUNT           integer  return only the N most recent posts
  FROM_DATE       YYYY-MM-DD start of a date range
  TO_DATE         YYYY-MM-DD end of a date range

# PARAMETERS

COLLECTION_NAME
: collection Markdown file

COUNT
: (optional) maximum number of posts to list

FROM_DATE
: (optional) start date (requires TO_DATE)

TO_DATE
: (optional) end date (requires FROM_DATE)

# EXAMPLES

{app_name} posts index.md
{app_name} posts index.md 10
{app_name} posts index.md 2026-01-01 2026-06-30

`

	PreviewHelpText = `%{app_name}(7) user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}

# NAME

preview

# SYNOPSIS

{app_name} preview

# DESCRIPTION

Starts a local HTTP server serving the htdocs directory so you can review
the generated site in a browser. The host and port are set in antenna.yaml
(defaults: localhost:8000).

Press Ctrl-C to stop the server.

# EXAMPLES

{app_name} preview
open http://localhost:8000

`

	QuoteHelpText = `%{app_name}(7) user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}

# NAME

quote

# SYNOPSIS

{app_name} quote TEXT_FRAGMENT_URL

# DESCRIPTION

Parses a TEXT_FRAGMENT_URL (a URL with a #:~:text= fragment) into a
Markdown quotation. Output is written to standard out. Redirect it into a
file to use as the basis for a response post.

# ALIASES

reply

# PARAMETERS

TEXT_FRAGMENT_URL
: a URL ending in a #:~:text= text-fragment selector

# EXAMPLES

{app_name} quote "https://example.com/article#:~:text=interesting%20passage"

`

	RssHelpText = `%{app_name}(7) user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}

# NAME

rss

# SYNOPSIS

{app_name} rss COLLECTION_NAME RSS_FILENAME [COUNT | FROM_DATE TO_DATE]

# DESCRIPTION

Writes an RSS 2.0 feed for COLLECTION_NAME to RSS_FILENAME. The optional
parameters work the same way as for posts: a COUNT or a FROM_DATE/TO_DATE
range to limit the items included.

# PARAMETERS

COLLECTION_NAME
: collection Markdown file

RSS_FILENAME
: output path for the RSS file

COUNT
: (optional) maximum number of items

FROM_DATE
: (optional) start date (requires TO_DATE)

TO_DATE
: (optional) end date (requires FROM_DATE)

# EXAMPLES

{app_name} rss index.md index.xml
{app_name} rss index.md archive.xml 2026-01-01 2026-06-30

`

	SitemapHelpText = `%{app_name}(7) user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}

# NAME

sitemap

# SYNOPSIS

{app_name} sitemap

# DESCRIPTION

Generates a set of sitemap files (sitemap_index.xml, sitemap_1.xml, ...)
for all pages and posts found via antenna.yaml. Place these in the root of
your htdocs directory so search engines can discover your content.

# EXAMPLES

{app_name} sitemap

`

	StylefromHelpText = `%{app_name}(7) user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}

# NAME

stylefrom

# SYNOPSIS

{app_name} stylefrom INPUT_FILE [OUTPUT_PATH]

# DESCRIPTION

Extracts the embedded CSS from a LibreOffice Writer HTML export
(INPUT_FILE, .html or .htm) and writes it to OUTPUT_PATH. OUTPUT_PATH
defaults to theme/style.css; the directory is created if needed.

Use this action to seed a theme stylesheet from a styled Writer document.

# PARAMETERS

INPUT_FILE
: path to the LibreOffice-exported HTML file

OUTPUT_PATH
: (optional) output CSS path (default: theme/style.css)

# EXAMPLES

{app_name} stylefrom my-doc.html
{app_name} stylefrom my-doc.html css/libreoffice.css

`

	UnpageHelpText = `%{app_name}(7) user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}

# NAME

unpage

# SYNOPSIS

{app_name} unpage INPUT_PATH

# DESCRIPTION

Removes the page record associated with INPUT_PATH from the pages collection.
The HTML and Markdown files on disk are not deleted.

# PARAMETERS

INPUT_PATH
: path to the source Markdown file as it was passed to page

# EXAMPLES

{app_name} unpage about.md

`

	UnpostHelpText = `%{app_name}(7) user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}

# NAME

unpost

# SYNOPSIS

{app_name} unpost COLLECTION_NAME URL | POST_PATH

# DESCRIPTION

Removes an item from COLLECTION_NAME using either the public URL associated
with the post or the POST_PATH value. The source Markdown and generated HTML
files on disk are not deleted.

# PARAMETERS

COLLECTION_NAME
: collection Markdown file

URL or POST_PATH
: URL or postPath value identifying the item to remove

# EXAMPLES

{app_name} unpost index.md https://example.com/blog/2026/04/12/my-post.html
{app_name} unpost index.md blog/2026/04/12/my-post.html

`

	// Reference help topics

	AccessibilityHelpText = `%{app_name}(7) user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}

# NAME

accessibility

# DESCRIPTION

{app_name} generates HTML that meets WCAG 2.1 Level A success criterion
2.4.1 (Bypass Blocks) and uses semantic HTML5 markup throughout.

# SKIP NAVIGATION LINK

Every generated page includes a skip-navigation link immediately after
<body> that lets keyboard users jump past the site navigation directly to
the main content:

  <a href="#main-content" class="skip-link">Skip to main content</a>

The skip link is visually hidden off-screen until it receives keyboard
focus, at which point it becomes visible.

If you maintain your own stylesheet, add these rules:

  .skip-link {
    position: absolute;
    top: -999px;
    left: -999px;
    padding: 0.5rem 1rem;
    background: #333;
    color: #fff;
    text-decoration: none;
    z-index: 9999;
  }
  .skip-link:focus {
    position: static;
    top: auto;
    left: auto;
    display: block;
  }

# LANG ATTRIBUTE

The <html> element includes a lang attribute from page.yaml. The default
is "en-US". Change it for non-English sites:

  lang: ja        # Japanese
  lang: fr-FR     # French (France)

# ARTICLE FOOTER FOR SOURCE LINKS

Feed item cards use <article> > <footer> (not <address>) for the source
URL at the bottom of each card.

# SEMANTIC TIME ELEMENTS

Publication and update dates are wrapped in <time datetime="YYYY-MM-DD">
elements and placed outside the <h2> heading so screen readers do not
announce the date as part of the article title.

# SEE ALSO

{app_name} help css
{app_name} help configuration

`

	ConfigurationHelpText = `%{app_name}(7) user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}

# NAME

configuration

# DESCRIPTION

{app_name} uses configuration files to control its behavior.

# ANTENNA.YAML (main configuration)

port
: (optional, default: 8000) localhost port for preview

host
: (optional, default: localhost) host name for preview

htdocs
: (optional, default: ".") directory for generated HTML/RSS

generator
: (optional, default: page.yaml) default page generator YAML

collections
: (required) list of collection objects

Each collection object:
  file
  : (required) path to the collection Markdown document

  title
  : (optional, default: filename) display name

  generator
  : (optional) per-collection page generator YAML override

  mode
  : (optional) rendering mode: "aggregate" (default) or "page-index"
     "aggregate"  feed-item cards from the items table (default)
     "page-index" simple <ul> link list from the pages table

Example antenna.yaml:

  htdocs: htdocs
  port: 8000
  collections:
    - file: index.md                 # aggregate (default)
    - file: links.md
      generator: links-page.yaml
    - file: pages.md
      mode: page-index               # renders a simple link list

# PAGE.YAML (page generator)

lang
: (optional, default: en-US) lang= attribute on <html>

title
: (optional) page <title> override

meta
: (optional) list of <meta> element attribute maps

link
: (optional) list of <link> element attribute maps

script
: (optional) list of <script> element attribute maps

style
: (optional) inline CSS injected at end of <head>

header
: (optional) innerHTML of <header>

nav
: (optional) innerHTML of <nav aria-label="Site navigation">

top_content
: (optional) content between <nav> and <main>

bottom_content
: (optional) content between </main> and <footer>

footer
: (optional) innerHTML of <footer>

allowed_meta_fields
: (optional) allowlist of front matter keys to emit as <meta>

Example page.yaml:

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
    <p>&copy; 2026 Your Name</p>
  allowed_meta_fields:
    - title
    - author
    - description
    - keywords

# SEE ALSO

{app_name} help metadata
{app_name} help accessibility

`

	MetadataHelpText = `%{app_name}(7) user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}

# NAME

metadata

# DESCRIPTION

Every YAML front matter field in a Markdown post or page is emitted into
the generated HTML <head> as a <meta name="KEY" content="VALUE"> element.
The same key-value pair is also written as a data-pagefind-filter attribute
on the enclosing <article> element for PageFind faceted search.

# STANDARD FIELDS

title
: Sets <title> in the HTML head (not emitted as <meta>)

description
: Short summary for search engines and RSS

author
: Author name

keywords
: List of tags; each value gets its own <meta> pair

pubDate
: Publication date required for posts (YYYY-MM-DD)

postPath
: Relative path to the generated HTML file required for posts

link
: Public URL of the post

series
: Series name for multi-part posts

seriesNumber
: Position within the series

dateCreated
: Creation date

dateModified
: Last-modified date

datePublished
: Publication date (alias for pubDate in page context)

# EXAMPLE FRONT MATTER

---
title: "My Post Title"
description: "A short summary."
author: "Your Name"
keywords:
  - go
  - web
pubDate: "2026-06-27"
postPath: "blog/2026/06/27/my-post.html"
series: "Building with Antenna"
seriesNumber: 3
---

# CONTROLLING WHICH FIELDS ARE PUBLISHED (allowed_meta_fields)

By default all front matter keys are emitted as <meta> elements. To
restrict which keys are published, set allowed_meta_fields in page.yaml:

  allowed_meta_fields:
    - title
    - author
    - description
    - keywords
    - series
    - seriesNumber

When this list is set, only the listed keys appear in generated HTML.
Internal workflow fields (build flags, notes) are silently omitted.

# SEE ALSO

{app_name} help configuration
{app_name} help post

`
)
