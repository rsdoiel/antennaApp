

# antennaApp

**antenna** is a command line feed oriented content management tool. It let's you create and curate micro blogs, blogs, link blogs, news sites using Markdown and a sprinkling of YAML[^1]. It is inspired by Dave Winer's [Textcasting](https://textcasting.org) and [FeedLand](https://feedland.org) and my own experimental website, [antenna](https://rsdoiel.github.io/antenna).

[^1]: Configuration and metadata (front matter in Markdown) are maintained using YAML.

Antenna App's Features:

- Pages and posts are written using Markdown
  - Pages are excluded from feeds
- Posts use metadata expressed as YAML front matter
  - Posts are included in feeds
- Support for multiple collections of feeds
  - A feed collection is defined by a Markdown document containing a list of links to feeds
- Collections can be harvested individually or collectively
  - harvested content is stored in a SQLite3 database(s)
  - harvested content in a collection can be aggregated and rendered as an HTML page for reading
  - Posts are associated with a collection and are part of the output feeds
- HTML, RSS 2.0, OPML and sitemap.xml documents are automatically generated without additional configuration
  - HTML files are generated for each post or page 
- A preview feature to view the render content in your web browser via a localhost URL
- Page generators YAML files can composed using a "themes"
  - themes are a directory populated with Markdown files for other theme elements[^2]
  - Page generators are implemented as a YAML file, they are used to frame the Markdown content added a pages and posts
- Antenna plays well with other static website tools (e.g. PageFind, FlatLake)

[^2]: A YAML file called head.yaml is used to describe custom meta, link and script elements in the page. There is a style.css file that can be included in a theme to be included in the head element of rendered HTML.

The ability to harvest feed items means we can read what others post on the web on our own website. Markdown content can be added to a feed allows us to comment on the items read (thus being social). This can all be run on localhost (our own computer) or staged for public Web consumption via a static host provider.

While Antenna App was initially conceived as a link blogging tool, it doesn't impose a directory structure on your site. It can be used for general purpose sites, blogs, microblogs and linkblogs too.

A static website using **antenna** can grow through either enhancing the HTML markup defined in the YAML configuration, through themes or through manipulation of the SQLite3 databases holding collection metadata. This provides opportunities to integrate with other static website tools like [PageFind](https://pagefind.app "A browser side search engine") and [FlatLake](https://flatlake.app "A static JSON API driven by front matter in Markdown documents"). You can even use **antenna** to augment your existing blog.

### Authors

- Doiel, R. S. Doiel

## Software Requirements

Requirements top compile the Antenna App.

- Go >= 1.25.3
- CMTools >= 0.0.40

### Software Suggestions

Helpful if you are developing the Antenna App.

- GNU Make >= 3.4
- Pandoc >= 3.1
- Bash or Powershell


## Related resources

- [Getting Help, Reporting bugs](https://github.com/rsdoiel/antennaApp/issues)
- [LICENSE](https://www.gnu.org/licenses/agpl-3.0.en.html)
- [Installation](INSTALL.md)
- [User Manual](user_manual.md)
- [About](about.md)

