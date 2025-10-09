---
title: antennaApp
abstract: "**antenna** is a tool for working with RSS feeds and rendering a link blog.
It is inspired by Dave Winer&#x27;s [Textcasting](https://textcasting.org) and
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
collection basis. That means it is possible to recreate a &quot;news paper&quot; like experience. 

A static website using **antenna** can grow through either enhancing the HTML markup defined
in the YAML configuration or through manipulation of the collection contents in the SQLite3 database.
This provides opportunities to integrate with other static website tools like
[PageFind](https://pagefind.app &quot;A browser side search engine&quot;) and
[FlatLake](https://flatlake.app &quot;A static JSON API driven by front matter in Markdown documents&quot;).
You can even use **{app_name}** to augment your existing blog."
authors:
  - family_name: Doiel
    given_name: R. S. Doiel
    id: https://orcid.org/0000-0003-0900-6903



repository_code: https://github.com/rsdoiel/antennaApp
version: 0.0.11
license_url: https://www.gnu.org/licenses/agpl-3.0.en.html

programming_language:
  - Go

keywords:
  - RSS
  - Feeds
  - Linkblog
  - website generator

date_released: 2025-10-08
---

About this software
===================

## antennaApp 0.0.11

- adding theme support to make it easy to customize generator YAML files

### Authors

- R. S. Doiel Doiel, <https://orcid.org/0000-0003-0900-6903>






**antenna** is a tool for working with RSS feeds and rendering a link blog.
It is inspired by Dave Winer&#x27;s [Textcasting](https://textcasting.org) and
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
collection basis. That means it is possible to recreate a &quot;news paper&quot; like experience. 

A static website using **antenna** can grow through either enhancing the HTML markup defined
in the YAML configuration or through manipulation of the collection contents in the SQLite3 database.
This provides opportunities to integrate with other static website tools like
[PageFind](https://pagefind.app &quot;A browser side search engine&quot;) and
[FlatLake](https://flatlake.app &quot;A static JSON API driven by front matter in Markdown documents&quot;).
You can even use **{app_name}** to augment your existing blog.

- License: <https://www.gnu.org/licenses/agpl-3.0.en.html>
- GitHub: <https://github.com/rsdoiel/antennaApp>
- Issues: <https://github.com/rsdoiel/antennaApp/issues>

### Programming languages

- Go




### Software Requirements

- Go >= 1.25.1
- CMTools >= 0.0.40


### Software Suggestions

- GNU Make &gt;&#x3D; 3.4
- Pandoc &gt;&#x3D; 3.1
- Bash or Powershell


