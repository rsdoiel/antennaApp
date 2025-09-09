

# antennaApp

**antenna** is a tool for working with RSS feeds and rendering a link blog.
It is inspired by Dave Winer's [Textcasting](https://textcasting.org) and [FeedLand](https://github.com/scripting/feedLand/).

My approach is focused on curating feeds for static website generation. Currently 
**antenna** is a command line tool. It runs in a "terminal" under Rapsberry Pi OS,
macOS and Windows. From the terminal you can create, curate and stage in bound and collections
of RSS feeds using simlpe Markdown files. You can even post new Markdown documents to a
feed collection. 

I believe that the link blog where you both consume and generate RSS can be a basis for a truely 
distributed social web with out the complexity of many of the current (2015 - 2025)
proposed solutions.

Features:

- support for multiple collections of feeds with
- Outbound feeds aggregated by feed collection
- a collection is defined by a Markdown document containing a list of links to feeds
  - This means configuration of the feeds you follow are written using a simple Markdown file
- collections can be harvested (in bound feeds)
  - harvested content is stored in a SQLite3 database
- Collections are aggregated and rendered as an HTML page
- Collections are aggregated and rendered as RSS 2.0 XML
- Markdown documents can be imported into a collection as a feed item
  - example a Markdown blog post or micro blog feed only post
- You staged website can be previewed on localhost using your web browser

The ability to harvest feed items means we can read what others post on the web. The Markdown content
can be added to a collection allows us to comment on the items read (thus being social).

Through YAML configuration files you can customize the HTML rendered by **antenna** on a per
collection basis. That means it is possible to recreate a "news reader" like experience. 

A static website using **antenna** can grow and be enhanced.  

- enhancing the HTML markup rendered by update the YAML configuration for a collection
- add CSS and JavaScript to run browser side to enhance usability
- integrate [PageFind](https://pagefind.app) to provide static site search
- integrate [FlatLake](https://flatlake.app) to provide a JSON API

While **antenna** was designed to be a link blog or news site generator it can also
function as blog.

## Release Notes

**antenna** is a prototype. It is intended to be used to generate the [Antenna website](https://rsdoiel.github.io/antenna).


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

