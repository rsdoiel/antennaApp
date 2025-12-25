

# Antenna App

**antenna** is a text (terminal) based feed oriented content management tool. It let's you create, curate and render simple websites, micro blogs, blogs, link blogs and personal news sites using Markdown. It runs interactively by default but is also easily scripted for command line use. A novel feature is its support for both inbound and outbound RSS feeds with integrated Markdown content via the RSS 2.0 source namespace. It was inspired by Dave Winer's [Textcasting](https://textcasting.org) concept.It grew out of my experimental personal  news website, [antenna](https://rsdoiel.github.io/antenna). Antenna App is intended to make it possible to return to a simple web. The goal is to make it easy for writers create web sites with only a knowledge of Markdown. This can enable writers to create publically or privately available websites with little more than their favorite text editor and Antenna App.

Antenna App's Features:

- An interactive text and scriptable command line interfaces
- Pages and posts are written using Markdown
- Posts are automatically included in an RSS feed for your site
- Aggregations of feeds can be included by providing Markdown list of links to RSS, Atom and JSON feeds
- Pages, posts and aggregations __can include__ metadata expressed as YAML front matter[^1]
- Local posts can be added to curated feed collection (introducing a distributed option for __opt in__ conversations)
- You can have multiple collections of feeds
- Feed collections can be harvested individually or collectively
- The managed content is stored in SQLite3 database(s)
- HTML, RSS 2.0, OPML and sitemap.xml documents are automatically generated without additional configuration
  - HTML files are generated for each post, page and aggregation in each collection 
- A preview web service is provided to view the render content in your web browser via a localhost URL
- Site customization is possible through the use of page generators
  - Page generators are YAML files describing basic elements of a page like  header, navigation, before/after content section, footer
  - Custom page generators are composed by apply "themes" allowing customized Markdown and CSS to be shared
    - Themes are a directory populated with Markdown files describing each element of the theme[^2]
  - Each collection can have it's own page generator and theme
- Antenna App plays well with other static website tools (e.g. PageFind)

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

