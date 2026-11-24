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

page INPUT_PATH [OUTPUT_PATH...]
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

prune [-w] [-all]
: Remove the pages whose source file is gone: the pages record and the output
files it lists. A dry run listing them unless -w is given. See "{app_name} help prune".

migrate-links [-w] [PATH]
: In your Markdown, change links to .html pages that {app_name} renders from
Markdown into links to the .md source, which is what a Markdown reader needs. A
dry run showing a diff unless -w is given. See "{app_name} help migrate-links".

validate [-json] [-online] [-color WHEN] [PATH]
: Check the site, a directory or one Markdown file for problems, and for what
Markdown readers such as BonBon need. It only reads: it never changes a file.
With no PATH it checks the whole site named by the configuration. Exit 0 means no
errors or warnings, 1 means there are some. See "{app_name} help validate".

pages [-stale]
: List the page entries in the pages.md collection, newest first, one per line
with six tab-separated fields: input path, output path (deprecated), updated,
outputs, includes and stale. -stale lists only the pages that are out of date.


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

sitemap [-clean]
: This will generate a set of sitemap files for pages and posts found through the
{app_name}.yaml file. (e.g. sitemap_index.xml, sitemap_1.xml, sitemap_2.xml, ...)
With -clean, sitemap files left over from an earlier run are removed.

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

completion SHELL [-install]
: Write a shell completion script for SHELL (bash or powershell) to standard
output. With -install the script is installed for every new session instead:
for bash into $XDG_DATA_HOME/bash-completion/completions/{app_name} (replacing only
a file {app_name} wrote), for PowerShell beside the profile with one line added to
the profile to load it, once. Use '{app_name} help completion' for full details.

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

# EXIT STATUS

**{app_name}** exits with the codes the workspace uses for every
command-line tool (workspace DR-0003), so a script can tell what went
wrong from the number alone.

| Code | Class | Meaning |
|-----:|-------|---------|
| 0 | ok | Success. A listing that matches nothing (posts, items, pages) is still 0. |
| 1 | negative | The command ran and the answer is no: the named collection is not in the configuration, an action you cancelled. |
| 2 | usage | The command line is wrong and nothing was attempted: unknown action, flag or help topic, a missing or surplus argument, a bad date or count. |
| 65 | data | Content that was read is wrong: a document with unclosed front matter, a file that is not a valid ODT, a feed that does not parse, a damaged database. |
| 66 | no_input | A named file or the workspace is missing: no antenna.yaml, no such Markdown file. |
| 69 | unavailable | A feed or service could not be reached, or answered with an error. |
| 70 | internal | A bug, or an error nothing classified. Please report it. |
| 73 | cant_create | An output could not be created: the file or directory is in the way, or its directory cannot be made. |
| 74 | io | A read or write failed part way. |
| 75 | temp_fail | Trying again later may work: a locked database, a port already in use. |
| 77 | no_permission | The operating system refused access. |
| 78 | config | A configuration file is present but wrong: antenna.yaml or a generator YAML file that does not parse, or a setting with an invalid value. |

The harvest and generate actions do all their work first. When some feeds,
collections or pages failed, they print the counts (for example "1 of 2
feeds failed") and exit with the class of the first failure, in
configuration order, not 0. A cron job that harvests daily will therefore
see a non-zero status when a feed is down, and the other feeds will still
have been harvested.

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

Because blogit is for putting a post on the blog, it also makes the post
publishable. In the copy it makes (never in your original) it adds:

- pubDate, when the file has neither pubDate nor datePublished: the RSS 2.0 date
  form BonBon expects, with a numeric offset in your local time zone, for
  example "Thu, 19 Nov 2026 09:30:00 -0800". It is the current time, or, when
  POST_DATE is given, noon local on that day.
- title, when the file has none: the text of its first level-1 heading
  (# Heading). The body is not changed.

An existing pubDate, datePublished or title is never changed. If POST_DATE is
on a different day than the pubDate in the file, blogit warns and keeps the
file's date; the directory follows POST_DATE. It writes no author, description,
link, urlPrev or urlNext. A post with no pubDate stays a draft, so to keep one
as a draft use post, not blogit. See {app_name} help metadata.

The time and the day directory come from the time zone of the computer
{app_name} runs on, which the TZ environment variable overrides. If you work
over SSH on another machine or are travelling, say where you are:

    TZ=Asia/Tokyo {app_name} blogit my-post.md

A pubDate you wrote yourself, with its own offset, is kept as it is.

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

A collection with items to show also gets a Markdown rendering of them,
STEM_items.md, beside its HTML page (not the pages collection, nor one with
markdownItems: false; markdownItems: true turns it on for pages), and the page declares it in its head as
its Markdown alternate. A collection with nothing to show has neither, and a
file left by an earlier run is removed.

Output is the same every time it is generated from the same data. A file is
rewritten only when its content changes. The lines that record when it was
built (the date meta element of an HTML page, dateModified in an items file,
lastBuildDate in an RSS feed and dateCreated in OPML) do not count, so
generating an unchanged site changes no file or modification time, and those
dates say when the content last changed, not when generate last ran.

When the whole site is generated (no COLLECTION_NAME), generate also writes a
table of contents for Markdown readers such as BonBon and cmarkbrowser,
htdocs/.well-known/toc.opml: one OPML outline with the site's pages, the items
pages of its collections and its posts (oldest first, by year and month). Only
pages and posts whose Markdown is published in htdocs are listed. The addresses
are made from base_url in antenna.yaml, which it needs: without one a warning
is given and the file is not written. Like the other files it is rewritten only
when its content changes.

Sites listed in the subsites setting of antenna.yaml are added to the outline as
include outlines (type="include", url=SITE/.well-known/toc.opml) after those
groups, so the table of contents at a host's root leads BonBon to its other
sites. Nothing is fetched. If the bonbon setting of antenna.yaml has a key,
generate also writes .well-known/bonbon.json from it; with none it writes no
file and leaves any file you wrote alone. See {app_name} help configuration.

The same outline gives previous and next links to the posts it lists that name
neither urlPrev nor urlNext in their front matter: the older post as previous
and the newer as next. The first and last have only one. A page gets none from
the table of contents, because pages are listed in file-name order, which is not
a sequence. A page or post in a sequence named in the sequences setting of
antenna.yaml gets that sequence's previous and next instead, and the generated
Markdown copy of such a page carries them as urlPrev and urlNext. The author's
own keys always win, and nothing is written into your Markdown.

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

Retrieves RSS, Atom and JSON feed content for all collections (or only
COLLECTION_NAME) and stores harvested items in each collection's SQLite3
database. Run generate afterwards to rebuild the HTML pages.

A link in a collection's Markdown file may be the feed itself or the address
of a site. For a site, harvest finds the feed, trying in this order: the
address itself; an rss outline (xmlUrl) in a table of contents
(.well-known/toc.opml); a link tag in an HTML page that names a feed; and up
to eight common paths (feed, feed/, rss, rss.xml, feed.xml, atom.xml,
index.xml, feeds/posts/default). The address it finds is remembered in the
collection's database, so later runs go straight to it and make no guesses.
A site that has only a table of contents of pages and no feed cannot be
collected: harvest says so for that link and carries on with the others. It
never fetches the pages a table of contents lists.

Harvest reads each host's robots.txt first and obeys it for every request,
including the guesses. A URL it forbids is not requested and is reported once;
that is not a failure. A missing robots.txt means the site may be harvested,
and one with a mistake in it (a Crawl-delay line before the first User-agent
line, say) is read as far as it can be. A robots.txt that answers with a server
error or does not answer is asked for twice more, a second and two seconds
apart; if it still fails the host is not requested in that run and the link
fails as unavailable. A
Crawl-delay is honored up to 30 seconds; a host that asks for more is skipped
for the run. Requests to one host are made one at a time with a pause between
them, 500 ms unless the harvest setting in {app_name}.yaml says otherwise.

# PARAMETERS

COLLECTION_NAME
: (optional) harvest only this collection

# EXIT STATUS

0
: every feed was harvested, or left alone because robots.txt forbids it

2
: unknown option or extra argument

65
: a feed, table of contents or page could not be parsed, or a site has no
  feed to find

66
: the configuration or a collection file is missing

69
: a feed, or a host's robots.txt, could not be reached or answered with an error

74
: reading a response failed part way

78
: the configuration is wrong, for example a harvest pause that is not a duration

When some feeds fail the others are still harvested, and harvest exits with the
class of the first failure, after saying how many failed.

# ALIASES

fetch

# EXAMPLES

{app_name} harvest
{app_name} harvest feeds/tech.md

A collection file may list a site and let harvest find its feed:

  - [Dave Winer](https://dave.writersweb.org/)
  - [A WordPress blog](https://example.org/)

# SEE ALSO

{app_name} help configuration (the harvest setting), {app_name} help markdown-web

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

{app_name} page INPUT_PATH [OUTPUT_PATH...]

# DESCRIPTION

Renders INPUT_PATH to each OUTPUT_PATH named. The extension of an output says
what it is: .html is an HTML page made with the page generator YAML, and .md
is the Markdown of the source with its include directives
(@include-text-block and @include-code-block) expanded. The same expanded text
is used for both, so they are one document. With no OUTPUT_PATH one HTML page
is written next to INPUT_PATH with a .html extension.

The source can be a text file with directives (index.txt), which is how a
page that includes other files publishes both a page and a Markdown copy
that a reader such as BonBon can use:

    antenna page index.txt index.md index.html

In the Markdown copy the front matter is kept as written, with two keys added
when the source has none: a title from the first level-1 heading, and the
site-wide author from antenna.yaml. Links stay as the author wrote them (.md
links stay .md). The source file is never changed. The HTML page declares the Markdown copy
beside it (same folder, same stem) with <link rel="alternate"
type="text/markdown">, so a reader that cannot ask a server for Markdown can
find it. In the HTML, a link to a Markdown file
antenna renders becomes a link to its .html page.

Every run is recorded in the pages collection (pages.md / pages.db): the
outputs it wrote, in order, and the files the source includes, nested ones too.
{app_name} pages shows them and {app_name} generate rebuilds from them. A run
that writes only Markdown is recorded too. A run replaces what the last run
recorded for the same source, so name every output in one command. A file that
is only included into other pages (recent.txt) is never given to antenna page.

An include is expanded where its directive starts a line. An
@include-text-block file is expanded in turn, so it can include others; nesting
deeper than three levels draws a warning, and deeper than five levels, or a file
that includes itself, is an error (exit 65). An @include-code-block file is
shown as it is and never expanded.

Pages are excluded from RSS feeds. This action is for static content such
as About pages, contact pages, and search pages.

An output that is the source file is refused (exit 65) and nothing is written.
An output that does not end in .html or .md is a usage error (exit 2), and
nothing is written. A Markdown output needs a text source, not an ODT file.

WARNING: HTML in the Markdown source passes through unchanged (unsafe mode).
Only run this action on files you control and trust.

# PARAMETERS

INPUT_PATH
: path to the source file

OUTPUT_PATH
: (optional, any number) an output path ending in .html or .md

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

{app_name} pages [-stale]

# DESCRIPTION

Prints the page entries tracked in the pages collection (pages.md / pages.db)
in descending updated-timestamp order, one per line, with six fields separated
by tabs:

1. the input path, the source file
2. the output path, the first .html the page writes. Deprecated: it is kept
   for older tools and will stop being written; use field 4
3. updated, when {app_name} last built the page
4. outputs, a JSON array of every file the page writes, in the order given to
   {app_name} page, for example ["index.md","index.html"]
5. includes, a JSON array of every file the source includes with
   @include-text-block or @include-code-block, nested ones too
6. stale, empty for a current page, "stale" when the source or a file it
   includes was changed after the page was last built, "missing" when an output
   file is not there

The first three fields are what older versions printed. Fields 4 and 5 are
empty for a page built before they were recorded; run {app_name} generate to
fill them in. The staleness test compares file times, so it is only meaningful
in a working copy that was built there, and it does not see changes to the
generator YAML, the theme or antenna.yaml.

With -stale only the pages whose sixth field is not empty are printed, and the
exit status is 1 when there are any, 0 when there are none.

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

The preview also serves the Markdown web, so a Markdown reader such as BonBon
or cmarkbrowser can be tried before anything is published:

- Files ending .md and .markdown are sent as text/markdown; charset=utf-8 and
  .opml files as text/x-opml; charset=utf-8, whatever the host's own table says.
- The .well-known directory (in any folder, for toc.opml and bonbon.json) is
  served. Every other file or directory whose name starts with a dot is
  refused with 403, and so is anything inside .well-known that does.
- A client whose Accept header prefers text/markdown to text/html (BonBon sends
  "text/markdown, text/x-opml") that asks for an HTML page or a folder is sent
  Markdown instead: the page the HTML head declares with
  <link rel="alternate" type="text/markdown">, else the page's .md sibling (x.md,
  or index.md in a folder), else the HTML. Content-Location names the file sent
  and Vary: Accept is set. A browser gets the HTML as always.

Negotiation is a convenience of the preview. A static host cannot do it; what
works everywhere is the declared alternate and ordinary links to .md files.

# EXAMPLES

{app_name} preview
open http://localhost:8000
curl -H 'Accept: text/markdown' http://localhost:8000/blog/index.html

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

{app_name} sitemap [-clean]

# DESCRIPTION

Generates a set of sitemap files (sitemap_index.xml, sitemap_1.xml, ...)
in the htdocs directory so search engines can discover your content.

The sitemap lists the pages of your site, which come from the pages
collection (pages.md), and your local posts, which are the items with a
postPath in any collection. Items harvested from feeds are not pages of your
site and are never listed, so a site that only aggregates feeds has no
sitemap. When there is nothing to list, {app_name} says so on standard error,
writes no files and exits 0. A collection whose database file is missing is
reported and the others are still mapped, then the command exits 66.

# OPTIONS

-clean
: Remove the sitemap files an earlier run left behind. A site that shrank, or
  that stopped having anything to list, can have sitemap_N.xml files that the
  new sitemap_index.xml no longer names. With -clean, the regular files in the
  htdocs directory named exactly sitemap_NUMBER.xml that this run did not write
  are removed, and each removal is reported. When there is nothing to list,
  the old sitemap_index.xml is removed too, since nothing points at it. No
  other file is touched. If a collection could not be read nothing is removed,
  because the sitemap is then incomplete. Without -clean nothing is ever
  deleted.

# EXAMPLES

{app_name} sitemap
{app_name} sitemap -clean

`

	CompletionHelpText = `%{app_name}(7) user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}

# NAME

completion

# SYNOPSIS

{app_name} completion SHELL [-install]

# DESCRIPTION

Writes a shell completion script to standard output. SHELL is bash or
powershell. The script completes actions, the top level options, help
topics after "help", "new" after "themes", shell names after "completion",
and file names after -config and for other actions.

With -install the script is installed for every new session instead of
printed. For bash it is written to
$XDG_DATA_HOME/bash-completion/completions/{app_name} (default
~/.local/share/...), replacing only a file that {app_name} wrote. For
PowerShell it is written beside the profile and one line is added to the
profile to load it, once however often you install. The path written is
reported.

# PARAMETERS

SHELL
: bash or powershell

-install
: install the script instead of printing it

# EXAMPLES

Bash, for the current session:

    source <({app_name} completion bash)

Bash, for every session:

    {app_name} completion bash -install

PowerShell, for the current session:

    {app_name} completion powershell | Out-String | Invoke-Expression

PowerShell, for every session:

    {app_name} completion powershell -install

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
The HTML and Markdown files on disk are not deleted. When no page record
matches INPUT_PATH, {app_name} prints an error and exits 1.

# PARAMETERS

INPUT_PATH
: path to the source Markdown file as it was passed to page

# EXAMPLES

{app_name} unpage about.md

`

	PruneHelpText = `%{app_name}(7) user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}

# NAME

prune

# SYNOPSIS

{app_name} prune [-w] [-all]

# DESCRIPTION

Finds the pages in the pages collection whose source file no longer exists, as
after an rm or mv, and removes them. With -w it deletes each page's record and
the output files the record lists (the HTML, and the Markdown copy when there is
one). Without -w it only lists them and changes nothing. unpage removes just
the record; prune is the clean-up for pages you have deleted.

The source file is never touched (it is gone), nor are files it included. An
output is kept, and reported on standard error, when another page still writes
it or when it is outside htdocs. The sitemap and toc.opml drop a pruned page the
next time you run {app_name} generate.

When the source of every page is missing, -w refuses, because that usually means
you are in the wrong directory or a drive is not mounted. Add -all to prune them
anyway.

# OPTIONS

-w
: delete the records and output files; without it prune is a dry run

-all
: with -w, prune even when every page's source is missing

# EXIT STATUS

0
: nothing to prune, or -w pruned the pages

1
: dry run found pages to prune, or -w refused because every source is missing

2
: unknown option or extra argument

66
: no pages collection or configuration found

74
: a file could not be deleted (the others are still pruned)

78
: the configuration is wrong

# EXAMPLES

Remove an old page, then see what prune would do, then do it:

~~~shell
rm old.md
{app_name} prune
{app_name} prune -w
{app_name} generate
~~~

`

	MigrateLinksHelpText = `%{app_name}(7) user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}

# NAME

migrate-links

# SYNOPSIS

{app_name} migrate-links [-w] [PATH]

# DESCRIPTION

Changes the links in your Markdown that point at an .html page {app_name}
renders from Markdown into links to the .md source. A Markdown reader such as
BonBon follows an .html link to the HTML page (BonBon opens it in the system
browser); the .md link keeps the reader in Markdown. This is the change
{app_name} validate reports as link-html.

By default it is a dry run: it prints each change as a unified diff and
changes nothing. Add -w to write the files. Review the diff first, and keep your
site in git so you can look at what was written.

With no PATH it works on the Markdown the site's own data names: the pages, the
collection source files and the posts. A PATH limits it to one file or the
Markdown files under one directory.

What it changes, and what it does not:

- Only a relative link to an .html or .htm page that {app_name} renders from a
  Markdown file. The fragment (#section) and query (?x=1) are kept. A link
  written from the site root (/blog/post.html) stays rooted (/blog/post.md).
- A link to a collection's page (news.html) becomes its items Markdown
  (news_items.md) when that file exists. If it does not exist yet the link is
  left, and a line says so.
- Inline links, links in angle brackets and reference definitions
  ([label]: destination). A title after the destination is kept.
- Never: code spans and code blocks, images, absolute addresses (including
  links to this site's own address), links to files that are not .html, and
  links to .html files {app_name} does not render, such as pages you wrote by
  hand.
- Only the bytes of each destination change. The front matter, the line endings
  and everything else are kept exactly. Each changed file is replaced by writing
  a new file beside it and renaming it.
- It does not run generate. The HTML {app_name} generates writes the .md links
  back as .html, so your pages look the same after the change. A post's stored
  copy is refreshed the next time you run {app_name} post on it.

Running it again changes nothing.

# OPTIONS

-w
: Write the changes. Without it nothing is written.

# EXIT STATUS

0
: It ran: a dry run, a write, or nothing to migrate.

2
: The command line is wrong (an unknown option, more than one PATH).

65
: A file could not be parsed. It is reported and the other files are still done;
the exit status is the class of the first failure.

66
: PATH does not exist, or there is no configuration file.

74
: A file could not be written.

# EXAMPLES

{app_name} migrate-links
{app_name} migrate-links blog
{app_name} migrate-links -w blog/index.md
{app_name} validate | grep link-html

# SEE ALSO

{app_name} help validate
{app_name} help markdown-web

`

	ValidateHelpText = `%{app_name}(7) user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}

# NAME

validate

# SYNOPSIS

{app_name} validate [-json] [-online] [-color WHEN] [PATH]

# DESCRIPTION

Checks Markdown for mistakes and for what Markdown readers such as BonBon
need, and reports each finding as one line:

    PATH:LINE: SEVERITY: CODE: message

validate only reads. It changes no file, and makes no network request unless
-online is given. With no PATH it checks the whole site: the pages, the
collection source files and the posts the site's own data names, plus the site
itself (the configuration, the table of contents, bonbon.json). A directory is
walked for .md and .markdown files, skipping folders whose name begins with a
dot and the items pages {app_name} generates. A single file is checked alone.
Checks that need to know what {app_name} renders are skipped when there is no
configuration file.

A finding has a severity: error, warning or info. The last line counts them.
Only errors and warnings affect the exit status.

# OPTIONS

-json
: Write one JSON object per finding (code, severity, file, line, message) and
a last object {"summary": {...}}. A failure of the command itself is also
written as an object with error, class and code.

-online
: Also ask each sub-site named in the subsites setting for its table of
contents, the way BonBon does, and warn about one that does not answer.

-color WHEN
: Color the report: the severity word (error red, warning yellow, info dim) and
the PATH:LINE. WHEN is auto (the default: color when standard output is a
terminal and NO_COLOR is not set), always or never. The text is the same with
or without color. -json is never colored.

# CHECKS

Each file:

front-matter
: The front matter cannot be read (error).

post-no-pubdate
: A file with a postPath and no pubDate is a draft; readers do not see it
(warning). {app_name} blogit adds the date.

date-unreadable
: A pubDate or datePublished {app_name} cannot read (error).

date-form
: pubDate is not the RSS 2.0 form BonBon expects (info).

no-title
: No title in the front matter and no level-1 heading (warning).

title-from-heading
: No title in the front matter, but a level-1 heading the title can come from
(info). {app_name} blogit and {app_name} page use that heading.

no-description
: No description, or an empty one; the table of contents and search results
use it (info).

description-long
: A description longer than descriptionMax characters (info). descriptionMax
is set in antenna.yaml and is 300 by default; 0 turns the check off.

no-author
: No author in the front matter and no author set in antenna.yaml (info).

postpath-mismatch
: postPath names a different file than the one it is in (warning).

nav-address
: urlPrev, urlNext or rightMarginImage is not an http, https, mailto or
relative address (error).

link-html
: A relative link to a .html page that {app_name} renders from Markdown. A
Markdown reader needs the .md link. The message says which (warning). A link to
a collection's page is named as its items Markdown when that file exists.

link-unrendered
: A relative .md link to a file that exists but that {app_name} does not render
(warning).

link-missing
: A relative link to a file that does not exist (warning).

link-absolute-self
: An absolute link to a page or file of this site, where a relative link works
for every reader (info).

raw-html
: Raw HTML, script or SVG, which Markdown readers strip (warning).

include-directive
: An @include-text-block or @include-code-block line in a Markdown file, which
a Markdown reader shows as text. Keep the source as a .txt file and write the
Markdown with {app_name} page (warning).

image-type
: An image that is not PNG, JPEG, GIF or WebP (warning).

macro-missing
: A {{page.NAME}} that names a key the page does not have (warning).

Code spans and code blocks are never looked into.

The whole site:

no-base-url
: There is no base_url (warning).

htdocs-missing
: htdocs is not a directory (error).

row-source-missing
: The pages table has a row for a file that does not exist (error). Remove it
with {app_name} unpage, or all of them with {app_name} prune.

alternate-static
: The generator YAML's link list holds a Markdown alternate (text/markdown or
application/markdown). The list is copied into every page, so the entry names
the same Markdown for all of them; {app_name} writes the right one for each page
(warning).

sequence-path-missing
: A path in the sequences setting of antenna.yaml is not there (warning).

sequence-empty
: A sequence has no members (info).

sequence-overlap
: A document is in more than one sequence and takes the neighbours of the
first (info).

row-output-missing
: The pages table records an output that is not there (warning). Run
{app_name} generate.

row-duplicate-output
: Two rows of the pages table write the same output, as after a rename
(warning). Remove the stale one with {app_name} unpage, or all of them with {app_name} prune.

include-missing
: A page's source includes a file that does not exist (warning). The directive
is left in the page as written.

output-stale
: An output is older than its source or a file it includes (info). Run
{app_name} generate.

not-published
: A page's or post's Markdown is not in htdocs, so the table of contents leaves
it out (warning).

toc-stale
: .well-known/toc.opml is missing or is not what generate would write (warning).

bonbon-json
: .well-known/bonbon.json differs from the bonbon setting, is not a JSON
object, or holds a known key BonBon would drop or break (warning).

nojekyll
: A github.io site served from the repository root has no .nojekyll file, so
GitHub Pages hides .well-known (warning).

subsite-unreachable
: With -online, a sub-site's table of contents does not answer 200 (warning).

# EXIT STATUS

0
: No errors or warnings (info does not count).

1
: At least one error or warning was found.

2
: The command line is wrong (an unknown option, more than one PATH).

66
: PATH does not exist, or there is no configuration file for a whole-site check.

78
: The configuration file is present but wrong.

# EXAMPLES

{app_name} validate
{app_name} validate blog
{app_name} validate blog/2026/11/19/my-post.md
{app_name} validate -online
{app_name} validate -json | jq 'select(.severity == "error")'

# SEE ALSO

{app_name} help markdown-web
{app_name} help metadata
{app_name} help blogit

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

The target is <main id="main-content" tabindex="-1">; tabindex="-1" makes
the link move keyboard focus into the main area, not just scroll to it.

If you maintain your own stylesheet, add these rules:

  main:focus { outline: none; }
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

base_url
: (optional) the address the site is published at, with no trailing slash.
  The feeds, the sitemap and the table of contents (.well-known/toc.opml) are
  made from it; without it the table of contents is not written

author
: (optional) the site-wide author. A page's generated Markdown (antenna page
  SOURCE OUTPUT.md) and the site's own posts in the collection page get it
  when they have no author of their own; an author in the document wins, and
  harvested items never get it. {app_name} validate does not suggest an author
  for a file when this is set

descriptionMax
: (optional, default: 300) the longest description {app_name} validate accepts
  before it suggests a shorter one; 0 turns that check off. A negative or
  non-numeric value is a configuration error (exit 78)

sequences
: (optional) the paths that hold a sequence of pages or posts, so previous and
  next are derived for their members. A list of entries, each with path
  (required) and order (optional). A path ending in / is a directory: its
  members are the published Markdown files directly in it, other than
  index.md, in order name (the default, by file name), date (oldest first, by
  pubDate or datePublished) or date-desc. A path ending in .md is a list file:
  its members are the site-local links in its body to published Markdown, in
  document order, each once; it takes no order. The list file is not a member
  of its own sequence. A member's previous and next are its neighbours in the
  sequence; the first has only a next and the last only a previous. A document
  in two sequences takes the first one listed. The previous and next are
  written to the HTML and to the generated Markdown copy of a page, never to
  your own Markdown file. A missing path, an unknown order, an order on a list
  file, an entry with no path or a path listed twice is a configuration error
  (exit 78); {app_name} validate warns about a path that is not there.

subsites
: (optional) other Markdown sites on the same host, so this site's table of
  contents can point at them. A list of objects, each with url (required: the
  sub-site's root, an absolute http or https address with a host, no query or
  fragment) and title (optional label). generate writes one include outline
  per entry, after the pages, collections and posts, in the order listed, at
  url/.well-known/toc.opml. BonBon asks only the root of a host for a table of
  contents and follows an include to reach a sub-site, so a host with several
  sites lists them here in the root site's antenna.yaml. Write the address as
  the sub-site's own pages are published (same scheme and host). Nothing is
  fetched: generate does not check that a sub-site exists. A missing or
  relative url, a duplicate, or an entry that is not a list of objects is a
  configuration error (exit 78).
  A relative .md link in a page or post that points into a sub-site on this
  host (one whose url has a path and the same host as base_url) is written
  into the HTML as the sub-site's .html page of the same name, for example
  /antenna/index.md becomes /antenna/index.html, with no warning. A sub-site
  is taken to publish an .html page beside each .md page.

bonbon
: (optional) the content of .well-known/bonbon.json, a mapping of keys to
  values. When it has at least one key, generate writes the file (a JSON object,
  keys sorted), and rewrites it only when it changes. BonBon reads three keys
  and only as strings: urlSearch, urlUpdateSocket and urlMirrorOf; any other key
  is written as given and ignored by BonBon today. A known key that is empty or
  not a string draws a warning, because BonBon keeps an empty one and a search
  then goes to a broken address: leave a key out instead of blanking it. With
  no bonbon setting {app_name} writes no file and never removes one, so a file
  you wrote by hand is safe; if you take the setting out, delete the file
  yourself. {app_name} runs no search or update service; see
  {app_name} help markdown-web

harvest
: (optional) settings for the harvest verb. pause is the wait between two
  requests to one host, a duration such as 500ms or 2s (default 500ms; 0s turns
  it off). A host's Crawl-delay in robots.txt is honored as well and wins when
  longer. A value that is not a duration, or is negative, is a configuration
  error (exit 78).

  harvest:
    pause: 1s

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

  markdownItems
  : (optional) true or false: whether generate writes STEM_items.md, the
     Markdown rendering of the collection's items. Default: true for every
     collection except pages.md, whose items are the site's own pages. false
     turns it off (the HTML, RSS and OPML are unchanged and a stale file is
     removed); true turns it on for pages.md.

Example antenna.yaml:

  htdocs: htdocs
  port: 8000
  base_url: https://example.org
  subsites:
    - url: https://example.org/opml
      title: opml
    - url: https://example.org/antenna
  bonbon:
    urlSearch: https://example.org/search
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

	MarkdownWebHelpText = `%{app_name}(7) user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}

# NAME

markdown-web

# DESCRIPTION

The Markdown web is a web whose pages are Markdown. A reader built for it, such
as BonBon or cmarkbrowser, asks a site for Markdown and shows that, instead of
HTML. A site made with {app_name} can serve both from the same address, with no
extra work beyond writing Markdown the way you already do.

This page is the overview. Each part is described in full in the topic named
in parentheses.

# WRITING LINKS

Link to the Markdown file, as you would in any Markdown tool:

    [Reading list](reading.md)

In the HTML {app_name} generates, a relative link to a .md file the site
renders (a page, a collection or a post) is written as the .html page. The
Markdown you publish is not changed, so a Markdown reader follows the .md link
and a browser the .html one. Absolute URLs, other file types and links in code
are left alone. A relative .md link to a file {app_name} does not render stays
as it is, and generate warns about it.

A link to a collection's own file (feeds/tech.md) means the collection, and
becomes the collection's items page.

A site written before this can have many links to .html pages. {app_name}
validate finds them (link-html) and {app_name} migrate-links rewrites them to
.md, as a dry run first.

# WHAT GENERATE PUBLISHES

(see: {app_name} help generate)

- A Markdown rendering of each collection's items, STEM_items.md, beside its
  HTML page, with front matter saying what it is. The HTML page declares it in
  its head as <link rel="alternate" type="text/markdown">.
- A table of contents, .well-known/toc.opml under htdocs, listing the site's
  pages, collection items pages and posts. A reader finds it at that fixed
  address and uses it for previous and next. It needs base_url in antenna.yaml.
- Include outlines for the sub-sites listed in the subsites setting of
  antenna.yaml, so a host with several Markdown sites (this one, /opml, /antenna)
  has one table of contents at its root that BonBon can follow. Only the root
  site of a host needs them; BonBon reads one level.
- .well-known/bonbon.json, only when the bonbon setting of antenna.yaml has a
  key (see WHAT IS NOT DONE).
- Previous and next links on posts, taken from that table of contents (the older
  post and the newer one), and on the pages and posts of a sequence named in the
  sequences setting of antenna.yaml, unless the document sets urlPrev and
  urlNext itself. A page in no sequence gets none.
- The same files, byte for byte, when nothing changed: generate rewrites a file
  only when its content changes, so a repository kept in git shows only real
  changes.

# ADDING A POST

{app_name} blogit FILE [POST_DATE] copies a Markdown file into
blog/YYYY/MM/DD/ and posts it, adding a pubDate in the RSS form BonBon expects
and a title from the first heading when the file has none (see
{app_name} help blogit). A post with no pubDate is a draft and is not in the
table of contents.

# CHECKING A SITE

{app_name} validate reads the whole site, a directory or one file and reports
what a Markdown reader would trip over: links that point at .html pages instead
of the .md source, raw HTML and SVG, images BonBon does not show, posts with no
date, a table of contents that is out of date, a missing .nojekyll on GitHub
Pages and more. It never changes a file. See {app_name} help validate.

# FRONT MATTER AND MACROS

(see: {app_name} help metadata)

The front matter fields BonBon reads (link, urlPrev, urlNext, rightMarginImage
and pubDate in the RSS 2.0 date form) are understood. In the text, {{page.NAME}}
is replaced by the front matter value NAME in the HTML, and left as written in
your Markdown, which the reader expands.

# TRYING IT

(see: {app_name} help preview)

{app_name} preview serves .md files as text/markdown, serves the .well-known
directory and, for a client that asks for Markdown (Accept: text/markdown),
sends the page's Markdown instead of its HTML. Point a Markdown reader at the
preview address to see the site as it will read once published. A static host
cannot negotiate content, so on the published site a reader depends on the
declared alternate, the table of contents and the .md links themselves.

# COLLECTING A MARKDOWN SITE

A Markdown site may publish no feed. Put its address in a collection's list of
links and {app_name} harvest looks for the feed: in the site's table of
contents (an outline with type rss and an xmlUrl), in the page's link tags,
and at common paths. When it finds one it uses the feed alone and remembers
where it was. A site that has only a table of contents of pages is not
collected, because that would mean one request for every page it lists; harvest
reports the link and goes on. {app_name} harvest respects robots.txt; see
{app_name} help harvest.

# WHAT IS NOT DONE

{app_name} runs no search service, update socket or mirror. Those need a
server, which a static site does not have. By default it writes no
.well-known/bonbon.json. If you run such a service, put its address in the
bonbon setting of antenna.yaml (urlSearch, urlUpdateSocket, urlMirrorOf) and
generate writes the file; see {app_name} help configuration.

Raw HTML, CSS, scripts and SVG images are ignored by Markdown readers.
The HTML pages keep them; the Markdown you write is what a reader shows.

# SEE ALSO

{app_name} help generate
{app_name} help preview
{app_name} help metadata
{app_name} help validate

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

Keys are camelCase. The eight marked (BonBon) are the ones Dave Winer's BonBon
reads, so one Markdown file serves both BonBon and the HTML site. Any other key
passes through unchanged.

title
: (BonBon) Sets <title> in the HTML head (not emitted as <meta>). Where the
  text says {{page.title}} it is expanded to this value (see MACROS).

description
: (BonBon) Short summary for search engines and RSS

author
: (BonBon) Author name

pubDate
: (BonBon) Publication date, required for posts. YYYY-MM-DD or an RSS 2.0 date
  such as Fri, 23 Oct 2026 15:19:45 GMT (see DATES)

link
: (BonBon) Public URL of the post. Where the title is shown through
  {{page.title}} it is a link to this address (see MACROS)

urlPrev
: (BonBon) The previous page. Written as <link rel="prev"> in the head and a
  Previous link on the page (see PREVIOUS AND NEXT)

urlNext
: (BonBon) The next page. Written as <link rel="next"> in the head and a Next
  link on the page (see PREVIOUS AND NEXT)

rightMarginImage
: (BonBon) An image floated to the right at the top of the text (see MARGIN
  IMAGE)

rightMarginImageAlt
: The alt text for rightMarginImage. antenna's own key, BonBon ignores it

keywords
: List of tags; each value gets its own <meta> pair

postPath
: Relative path to the generated HTML file required for posts

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

# DATES

pubDate may be written as 2026-06-27, as a timestamp such as
2026-06-27T10:00:00Z (quoted or not), or in the RSS 2.0 form BonBon uses,
Fri, 23 Oct 2026 15:19:45 GMT. The weekday may be left out, the day may have
one or two digits, and the zone may be GMT, UT, a US zone such as EST or PDT,
or an offset such as +0200. A date with a zone keeps it: antenna stores
"Mon, 30 Nov 2026 21:30:00 -0800" as 2026-11-30T21:30:00-08:00, so the post is
on November 30 in the table of contents, the feeds and the page, the day you
wrote it, wherever the reader is. A date in GMT, or with no zone, is stored as
YYYY-MM-DD HH:MM:SS. Posts sort by the moment they were made. Your Markdown
file is never changed. A date antenna cannot read is kept as written and
reported with a warning; it does not stop the command.

# MACROS

In the text of a page or post, {{page.NAME}} is replaced by the front matter
value NAME (spaces inside the braces are allowed), so a page can start with a
first-level heading that reads {{page.title}}. The value is inserted as plain
text: it never becomes
Markdown or HTML. Macros are not expanded inside code spans or code blocks, so
a page can show one. Expansion is a single pass; a value that itself holds a
macro is shown as written. A name the page does not have leaves the macro
visible and prints one warning. When the page has a link, the expanded title
links to it (http, https, mailto or a relative address). Your published
Markdown keeps the macros: BonBon expands them for its readers.

# PREVIOUS AND NEXT

urlPrev and urlNext name the previous and next page. A same-site .md address is
written as the .html page, like any other link. The page gets <link rel="prev">
and <link rel="next"> in the head and, after its text, a labelled navigation
with a Previous and a Next link; only the directions you name are shown. An
address that is not http, https, mailto or relative is ignored with a warning.

A post that gives neither key gets them when the site is generated, from the
site's table of contents (see antenna help generate): the older post as previous
and the newer as next. A page gets none, since the table of contents lists pages
in file-name order, which is not a sequence; a page or post in a sequence named
in the sequences setting of antenna.yaml gets that sequence's neighbours
instead (see antenna help configuration). A key you write always wins over
these.

# MARGIN IMAGE

rightMarginImage is the address of an image (http, https or a relative path).
It is shown floated to the right at the top of the text, and on a narrow screen
above the text. BonBon's key has no alt text, so set rightMarginImageAlt to
describe the image; without it the image is treated as decoration and has an
empty alt. Neither key is also written as a <meta> element.

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
urlPrev: "my-earlier-post.md"
rightMarginImage: "images/cover.jpg"
rightMarginImageAlt: "The book cover"
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
