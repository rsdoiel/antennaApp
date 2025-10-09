

# antennaApp

**antenna** is a command line feed oriented content management tool. It let's you create and curate micro blogs, blogs, link blogs, news and wiki websites using Markdown. Configuration and metadata are maintained using YAML. It is inspired by Dave Winer's [Textcasting](https://textcasting.org) and [FeedLand](https://feedland.org) and my own experimental website, [antenna](https://rsdoiel.github.io/antenna).

Antenna App's Features:

- support for multiple collections of feeds, each defined by a simple Markdown document
  - a feed collection is a Markdown document containing a list of links to feeds
- collections can be harvested individually or collectively
- harvested content is stored in a SQLite3 database
- harvested content in a collection can be aggregated and rendered as an HTML page for reading
- Markdown documents can be imported into a collection as a feed item (e.g. as a blog post or micro blog post)
- RSS 2.0, OPML and HTML documents are generated per collection
- HTML files are generated for each post or page 
- Posts and pages are written using Markdown with metadata expressed as YAML front matter
- A preview feature to view the render content in your web browser via a localhost URL
- Page generators are implemented as a YAML file
- Page generators can be "themed" using a directory with Markdown documents for each body element in the generator and a YAML file for head HTML element content

The ability to harvest feed items means we can read what others post on the web on our own website. Markdown content can be added to a feed allows us to comment on the items read (thus being social). This can all be run on localhost (our own computer) or staged for public Web consumption via a static host provider.

Through YAML configuration files you can customize the HTML rendered by **antenna** on a per collection basis. That means it is possible to recreate a "news paper" like experience. 

While Antenna App was initially conceived as a link blogging tool, it doesn't impose a directory structure on the site. It can be used for wiki like websites too (e.g. you could structure posts around topics expressed as paths).

A static website using **antenna** can grow through either enhancing the HTML markup defined in the YAML configuration, theming or through manipulation of the collection contents in the SQLite3 database. This provides opportunities to integrate with other static website tools like [PageFind](https://pagefind.app "A browser side search engine") and [FlatLake](https://flatlake.app "A static JSON API driven by front matter in Markdown documents"). You can even use **antenna** to augment your existing blog.

### Authors

- Doiel, R. S. Doiel

## Software Requirements

- Go >= 1.25.2
- CMTools >= 0.0.40

### Software Suggestions

- GNU Make >= 3.4
- Pandoc >= 3.1
- Bash or Powershell


## Related resources

- [Getting Help, Reporting bugs](https://github.com/rsdoiel/antennaApp/issues)
- [LICENSE](https://www.gnu.org/licenses/agpl-3.0.en.html)
- [Installation](INSTALL.md)
- [User Manual](user_manual.md)
- [About](about.md)

