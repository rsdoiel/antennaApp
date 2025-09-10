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

# DESCRIPTION

**{app_name}** is a tool for working with RSS feeds and rendering a link blog.
It is inspired by Dave Winer's [Textcasting](https://textcasting.org) and
<https://news.scripting.com>.

The approach I am taking is to make it easy to curate feeds and generated a static
website using a simple command line tool. I believe that the link blog where you
both consume and generate RSS can be a basis for a truely distributed social web
with out the complexity of many of the current (2015 - 2025) proposed solutions.

Features:

- support for multiple collections of feeds
- a collection is defined by a Markdown document containing a list of links to feeds
  - fromt matter in the Makrdown document is used to enhance content and RSS feed
- collections can be harvested, meaning content retrieved from the feeds listed in the Markdown document
- harvested content is stored in a SQLite3 database
- harvested content in a collection can be aggregated and rendered as an HTML page for reading
- generate HTML pages along side RSS 2.0 XML for collections
- Markdown documents can be imported into a collection as a "post" or removed with "unpost"
- HTML, CSS and JavaScript can be customized per collection using a YAML configuration file
- Custom SQL filters can be written for each collection stored in the YAML configuration 
- A preview feature to view the render content in your web browser via a localhost URL

The ability to harvest feed items means we can read what others post on the web. The Markdown content
can be added to a feed allows us to comment on the items read (thus being social).

Through YAML configuration files you can customize the HTML rendered by **{app_name}** on a per
collection basis. That means it is possible to recreate a "news paper" like experience. 

A statis website using **{app_name}** can grow through either enhancing the HTML markup defined
in the YAML configuration or through manipulation of the collection contents in the SQLite3 database.
This provides opporutinies to integrate with other static website tools like
[PageFind](https://pagefind.app "A browser side search engine"). You can even use
**{app_name}** to augment your existing static website blog.

# OPTIONS

--help, help
: display help

--license, license
: display license

--version, version
: display version

-config FILENAME
: Use specified YAML file for configution

# ACTION

Actions form the primary way you use Antenna and manage a link blog through its life cycle. Below
is a list of supported ACTION and their purpose. ACTION can be split between two general purposes
. The following commands are related to curating your Antenna's collections of feeds and feed items.

init [FILENAME]
: Initialize an Antenna instances by creating a YAML configuration file or validating one. If
FILENAME is provided that name will be used otherwise it will be called "antenna.yaml".

add COLLECTION_FILE [NAME DESCRIPTION]
: Add the feed collection name by COLLECTION_FILE to your Antenna configuration.
A COLLECTION_FILE is a Markdown document containing one or more links in a list. You 
can include a short name that will be displayed when the HTML was generated. You may
also supply a DESCRIPTION associated with the colleciton. These can also be set in
the Front Matter of the Markdown document.

del COLLECTION_FILE
: Remove a collection from the Antenna configuration.

post COLLECTION_FILE FILEPATH
: Add a Markdown document to a feed collection. The front matter is used to 
specify things like the link to the post, guid, description, etc. If these are not
provided then the post will display and error and not write the content to the
post directory location or add it to the collections. Required front matter
**title** or **description**, see
RSS 2.0 defined at <https://cyber.harvard.edu/rss/rss.html#hrelementsOfLtitemgt>.
To include a the file in the posts directory tree you need to provide a **postPath**.
In that case it is also recommended you provide a value for **link** that reflects the
public URL to where the post can be viewed.

unpost COLLECTION_FILE URL
: Remove an item from a collection using the URL associated with the item.

The following commands are related to producing a link blog static website.

harvest [COLLECTION_FILE]
: The harvest retrieves feed content. If COLLECTION_FILE is provided then only the 
the single collection will be harvested otherwise all collections defined in your
Antenna YAML configuration are harvested.

generate [COLLECTION_FILE]
: This process the collections rendering HTML pages and RSS 2.0 feeds for each collection.
If the collection name is provided then only that HTML page will be generated.

preview
: Let's your preview the rendered your Antenna instance as a localhost website using
your favorite web browser.

# CONFIGURATION

**{app_name}** uses a YAML configuration file. Below the the primary attributes you can
set in the YAML file.

port
: (optional, default: 8000) The localhost port for the "view" action

htdocs
: (optional, default: ".") The directory that rendered CommonMark, assets and HTML writen to. This is the directory
that will be served out using the "view" action.

generator
: (optional) This holds the default YAML configuration filename to use when
rendering the Antenna collections as HTML pages.

collections
: (required) This holds a list of collections managed by your Antenna instance

The collections attribute holds a collection objects. Each collection object has
the following attributes.

title
: (optional, defaults to the file name minus extension) name of the collection

file
: (required), the path to the Markdown file defining the collection

generator
: (optional) The YAML configuration filename to used to render the colllection
as HTML.

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

`
)
