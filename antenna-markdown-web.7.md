%antenna(7) user manual | version 0.2.0 d1de725
% R. S. Doiel
% 2026-11-23

# NAME

markdown-web

# DESCRIPTION

The Markdown web is a web whose pages are Markdown. A reader built for it, such
as BonBon or cmarkbrowser, asks a site for Markdown and shows that, instead of
HTML. A site made with antenna can serve both from the same address, with no
extra work beyond writing Markdown the way you already do.

This page is the overview. Each part is described in full in the topic named
in parentheses.

# WRITING LINKS

Link to the Markdown file, as you would in any Markdown tool:

    [Reading list](reading.md)

In the HTML antenna generates, a relative link to a .md file the site
renders (a page, a collection or a post) is written as the .html page. The
Markdown you publish is not changed, so a Markdown reader follows the .md link
and a browser the .html one. Absolute URLs, other file types and links in code
are left alone. A relative .md link to a file antenna does not render stays
as it is, and generate warns about it.

A link to a collection's own file (feeds/tech.md) means the collection, and
becomes the collection's items page.

A site written before this can have many links to .html pages. antenna
validate finds them (link-html) and antenna migrate-links rewrites them to
.md, as a dry run first.

# WHAT GENERATE PUBLISHES

(see: antenna help generate)

- A Markdown rendering of each collection's items, STEM_items.md, beside its
  items page STEM_items.html, with front matter saying what it is. The HTML page declares it in
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

antenna blogit FILE [POST_DATE] copies a Markdown file into
blog/YYYY/MM/DD/ and posts it, adding a pubDate in the RSS form BonBon expects
and a title from the first heading when the file has none (see
antenna help blogit). A post with no pubDate is a draft and is not in the
table of contents.

# CHECKING A SITE

antenna validate reads the whole site, a directory or one file and reports
what a Markdown reader would trip over: links that point at .html pages instead
of the .md source, raw HTML and SVG, images BonBon does not show, posts with no
date, a table of contents that is out of date, a missing .nojekyll on GitHub
Pages and more. It never changes a file. See antenna help validate.

# FRONT MATTER AND MACROS

(see: antenna help metadata)

The front matter fields BonBon reads (link, urlPrev, urlNext, rightMarginImage
and pubDate in the RSS 2.0 date form) are understood. In the text, {{page.NAME}}
is replaced by the front matter value NAME in the HTML, and left as written in
your Markdown, which the reader expands.

# TRYING IT

(see: antenna help preview)

antenna preview serves .md files as text/markdown, serves the .well-known
directory and, for a client that asks for Markdown (Accept: text/markdown),
sends the page's Markdown instead of its HTML. Point a Markdown reader at the
preview address to see the site as it will read once published. A static host
cannot negotiate content, so on the published site a reader depends on the
declared alternate, the table of contents and the .md links themselves.

# A SITE WITH NO HTML

A collection with html: false in antenna.yaml publishes only Markdown: the items
file, the RSS and OPML feeds and, with the table of contents, what a Markdown
reader needs. antenna page writes no HTML when you name only .md outputs. See
antenna help configuration for the setting and what it changes.

# COLLECTING A MARKDOWN SITE

A Markdown site may publish no feed. Put its address in a collection's list of
links and antenna harvest looks for the feed: in the site's table of
contents (an outline with type rss and an xmlUrl), in the page's link tags,
and at common paths. When it finds one it uses the feed alone and remembers
where it was. A site that has only a table of contents of pages is not
collected, because that would mean one request for every page it lists; harvest
reports the link and goes on. See antenna help harvest.

# WHAT IS NOT DONE

antenna runs no search service, update socket or mirror. Those need a
server, which a static site does not have. By default it writes no
.well-known/bonbon.json. If you run such a service, put its address in the
bonbon setting of antenna.yaml (urlSearch, urlUpdateSocket, urlMirrorOf) and
generate writes the file; see antenna help configuration.

Raw HTML, CSS, scripts and SVG images are ignored by Markdown readers.
The HTML pages keep them; the Markdown you write is what a reader shows.

# SEE ALSO

antenna help generate
antenna help preview
antenna help metadata
antenna help validate


