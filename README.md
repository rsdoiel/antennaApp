

# Antenna App

**antenna** is a command line feed oriented content management tool. It let's you create and curate simple websites, micro blogs, blogs, link blogs, and person news sites using Markdown. It was inspired by Dave Winer's [Textcasting](https://textcasting.org) concept, [FeedLand](https://feedland.org) and my own experimental website, [antenna](https://rsdoiel.github.io/antenna). The goal of the Antenna App is to make it easy for writers create web sites without being required to be web developers or software engineers. Ideally all you need to know Markdown and that would be enough to create writer oriented websites you can host publically or privately.

Antenna App's Features:

- Pages and posts are written using Markdown
- Posts are automatically included in an RSS feed for your site
- Aggregations of feeds can be included by providing Markdown list of links to RSS, Atom and JSON feeds
- Pages, posts and aggregations can include metadata expressed as YAML front matter[^1]
- Posts can be added to curated feed collection too
- You can have multiple collections of feeds
- Collections can be harvested individually or collectively
- The managed content is stored in SQLite3 database(s)
- HTML, RSS 2.0, OPML and sitemap.xml documents are automatically generated without additional configuration
  - HTML files are generated for each post, page and aggregation collection 
- A preview feature to view the render content in your web browser via a localhost URL
- Site customization is possible through the use of page generators
  - Page generators YAML files describing basic elements of a page, header, navigation, before/after content section, footer
  - Custom page generators are composed by apply "themes"
  - Themes are a directory populated with Markdown files describing each element of the theme[^2]
  - Page generators are implemented as a YAML file, they are used to frame the Markdown content added a pages and posts
  - Each collection may contain its own page generator
- Antenna App plays well with other static website tools (e.g. PageFind)
- Antenna App includes an experimental interactive terminal interface for curating your collections and posts

[^1]: YAML is a simple notation describing structed data. Front matter is a block of text hodling metadata about the document. Using YAML Front Matter is a common practice in the Markdown community. I first encountered it in the data science community and other scholarly science communities.

[^2]: A YAML file called head.yaml is used to describe custom meta, link and script elements in the page. There is a style.css file that can be included in a theme to be included in the head element of rendered HTML.

The combined features of Antenna App go beyond most static site generators. It provides a means of making collaboritive website through the ability to import feeds from one or more people. You can add your own Markdown content to a feed allowing for commentary in a specific feed collection. Feed orientation allows Antenna App sites to be social without turning over atonomy of our content to another site. 

I believe Antenna App is well suited for creating local community sites where each contributed produces a feed of articles that are then linked from a common website. This could be used for aggregating content of special interest groups or for local neighborhood news sites. The site doing the aggregation can control those feed items to be published allowing for editorial control. Since this runs on your computer you don't need to run specialized services hosted on the public web. You get the flexibity of a dynamic system with the low overhead and high availability provided by static websites.

### Authors

- Doiel, R. S. DoielHelpful if you are developing the Antenna App.


## Software Requirements

To run Antenna App you just need to [install](INSTALL.md) it on a Raspberry Pi, Linux, macOS or Windows computer.

Requirements top compile the Antenna App from source code.

- Go >= 1.25.3
- CMTools >= 0.0.40

### Software Suggestions

The following are helpful if you want to develop or customize the Antenna App software.

- GNU Make >= 3.4
- Pandoc >= 3.1
- Bash or Powershell


## Related resources

- [Getting Help, Reporting bugs](https://github.com/rsdoiel/antennaApp/issues)
- [LICENSE](https://www.gnu.org/licenses/agpl-3.0.en.html)
- [Installation](INSTALL.md)
- [User Manual](user_manual.md)
- [About](about.md)

