%antenna(1) user manual | version 0.0.21 a35859b
% R. S. Doiel
% 2026-01-04

# NAME

antenna

# SYNOPSIS

antenna [OPTIONS] ACTION [PARAMETERS]

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
You can even use **antenna** to augment your existing blog.

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
generated is called "antenna.yaml". It'll create a Markdown document called pages.md (if none exists)
and a pages.db SQLite3 file to hold the default collection. Collections hold posts and pages metadata.
Posts will be be included in RSS feed output while pages will be excluded from feed output.

add COLLECTION_FILE [NAME DESCRIPTION]
: Add the feed collection name by COLLECTION_FILE to your Antenna configuration.
A COLLECTION_FILE is a Markdown document containing one or more links in a list. You 
can include a short name that will be displayed when the HTML was generated. You may
also supply a DESCRIPTION associated with the collection. These can also be set in
the Front Matter of the Markdown document.

del COLLECTION_FILE
: Remove a collection from the Antenna configuration.

page INPUT_PATH [OUTPUT_PATH]
: This will create a standalone HTML page in a collection called pages.md. The pages.md
is the only collection that has pages (hense the name). It is also the default collection
(created by the init action). It uses the default page generator defined in the
antenna.yaml if one is not specifically set for the pages.db collection. The pages
command reads in the Markdown document from INPUT_PATH and writes it an HTML file
using the  same basename. If OUTPUT_PATH is set it uses that name for the HTML file
generated. (NOTE: the pages are not shown in the RSS feed. The page action is useful for
pages like an about page, home page, and search page. __The Markdown processed via page
action will allow "unsafe" HTML to pass through. Only use page if you trust the
Markdown document!!!__)

unpage INPUT_PATH
: This will will remove a page's from a collection based in the input filepath provided.
It does not remove the page on disk, just from the collection so that it will no longer
be used to create a corresponding HTML page when the generate action is run.

pages
: List the page entries in the pages.md collection. Pages are ordered by
descending updated timestamp.


post [COLLECTION_NAME] FILEPATH
: Add a Markdown document to a feed collection (default is pages.md). The front matter is
used to specify things like the link to the post, guid, description, etc. If these are not
provided then the post will display and error and not write the content to the
post directory location or add it to the collections. Required front matter
**title** or **description**, see RSS 2.0 defined at
<https://cyber.harvard.edu/rss/rss.html#hrelementsOfLtitemgt>. To include a the file in
the posts directory tree you need to provide a **postPath**. In that case it is also
recommended you provide a value for **link** that reflects the public URL to where the
post can be viewed. Post like page will allow unsafe HTML to be used in the Markdown
document unlike aggregated content.

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
antenna.yaml file. (e.g. sitemap_index.xml, sitemap_1.xml, sitemap_2.xml, ...)

curate
: Interactive collection and item curation.

preview
: Let's your preview the rendered your Antenna instance as a localhost website using
your favorite web browser.

apply THEME_PATH [YAML_FILE_PATH]
: This will apply the content THEME_PATH and update the YAML generator file described
by YAML_FILE_PATH. If YAML_FILE_PATH is not provided then that YAML generator file
will be replaced by the theme.

# CONFIGURATION

**antenna** uses a YAML configuration file. Below the the primary attributes you can
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
: (required) This holds a list of collections managed by **antenna**. For **antenna**
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

# EXAMPLES

Here's an example of using **antenna** to create a single collection static site.

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
antenna init
antenna add example.md
antenna harvest
antenna generate
antenna preview
~~~

The "preview" action runs a localhost web server so you can read the
contents in the generate HTML page called "example.html". You'd open
your web browsers to <http://localhost:8000/example.html> to review
the harvested content.

To update your static site you'd just do the following

~~~shell
antenna harvest
antenna generate
antenna preview
open http://localhost:8000
~~~

When you run the "generate" action HTML files and RSS feeds will
be written to the directory designated in the **antenna** YAML
configuration file (defaults to to the current working directory, "").
The "preview" action  serves that out over localhost (default port 8000)
so you can read your static site with your favorite web browser.

# Also 

- [antenna-themes (5)](antenna-themes.5.md)


