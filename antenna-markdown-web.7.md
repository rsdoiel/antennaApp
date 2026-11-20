%antenna(7) user manual | version 0.0.28 1e5ac36
% R. S. Doiel
% 2026-10-05

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

# WHAT GENERATE PUBLISHES

(see: antenna help generate)

- A Markdown rendering of each collection's items, STEM_items.md, beside its
  HTML page, with front matter saying what it is. The HTML page declares it in
  its head as <link rel="alternate" type="text/markdown">.
- A table of contents, .well-known/toc.opml under htdocs, listing the site's
  pages, collection items pages and posts. A reader finds it at that fixed
  address and uses it for previous and next. It needs base_url in antenna.yaml.
- Previous and next links on pages and posts, taken from that table of contents
  unless the page sets urlPrev and urlNext itself.
- The same files, byte for byte, when nothing changed: generate rewrites a file
  only when its content changes, so a repository kept in git shows only real
  changes.

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

# WHAT IS NOT DONE

antenna does not write .well-known/bonbon.json. Its keys (urlSearch,
urlUpdateSocket, urlMirrorOf) need a search service, a websocket or a mirror,
none of which a static site has. If you run one, write the file by hand in the
htdocs directory.

Raw HTML, CSS, scripts and SVG images are ignored by Markdown readers.
The HTML pages keep them; the Markdown you write is what a reader shows.

# SEE ALSO

antenna help generate
antenna help preview
antenna help metadata


