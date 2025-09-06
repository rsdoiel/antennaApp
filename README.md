

# antennaApp

**antenna** is a tool for working with RSS feeds and rendering a link blog.
It is inspired by Dave Winer's [Textcasting](https://textcasting.org) concept.

The approach I am taking is to make it easy to curate feeds and generated a static
website using a simple command line tool. I believe that the link blog where you
both consume and generate RSS can be a basis for a truely distributed social web
with out the complexity of many of the current (2015 - 2025) proposed solutions.

Features:

- support for multiple collections of feeds
- a collection is defined by a Markdown document containing a list of links to feeds
- collections can be harvested, meaning content retrieved from the feeds listed in the Markdown document
- harvested content is stored in a SQLite3 database
- harvested content in a collection can be aggregated and rendered as an HTML page for reading
- Markdown documents can be imported into a collection as a feed item
- RSS 2.0 XML can be generated from a collection
- A preview feature to view the render content in your web browser via a localhost URL
- You can manage your collections via a localhost URL too.

The ability to harvest feed items means we can read what others post on the web. The Markdown content
can be added to a feed allows us to comment on the items read (thus being social).

Through YAML configuration files you can customize the HTML rendered by **antenna** on a per
collection basis. That means it is possible to recreate a "news paper" like experience. 

A statis website using **antenna** can grow through either enhancing the HTML markup defined
in the YAML configuration or through manipulation of the collection contents in the SQLite3 database.
This provides opporutinies to integrate with other static website tools like
[PageFind](https://pagefind.app "A browser side search engine") and
[FlatLake](https://flatlake.app "A static JSON API driven by front matter in Markdown documents").
You can even use **antenna** to augment your existing blog.

## Release Notes

- version: 0.0.1
- status: concept


prototype of a simple tool to create and curate an Antenna like website.


### Authors

- Doiel, R. S. Doiel



## Software Requirements

- Go >= 1.25.0
- CMTools >= 0.0.40

### Software Suggestions

- GNU Make >= 3.4
- Pandoc >= 3.1
- Bash or Powershell



## Related resources



- [Getting Help, Reporting bugs](https://github.com/rsdoiel/AntennaApp/issues)
- [LICENSE](https://www.gnu.org/licenses/agpl-3.0.en.html)
- [Installation](INSTALL.md)
- [About](about.md)

