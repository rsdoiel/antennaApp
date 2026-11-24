%antenna(7) user manual | version 0.1.2 24f4114
% R. S. Doiel
% 2026-11-23

# NAME

harvest

# SYNOPSIS

antenna harvest [COLLECTION_NAME]

# DESCRIPTION

Retrieves RSS, Atom and JSON feed content for all collections (or only
COLLECTION_NAME) and stores harvested items in each collection's SQLite3
database. Run generate afterwards to rebuild the HTML pages.

A link in a collection's Markdown file may be the feed itself or the address
of a site. For a site, harvest finds the feed, trying in this order: the
address itself; an rss outline (xmlUrl) in a table of contents
(.well-known/toc.opml); a link tag in an HTML page that names a feed; and up
to eight common paths (feed, feed/, rss, rss.xml, feed.xml, atom.xml,
index.xml, feeds/posts/default). The address it finds is remembered in the
collection's database, so later runs go straight to it and make no guesses.
A site that has only a table of contents of pages and no feed cannot be
collected: harvest says so for that link and carries on with the others. It
never fetches the pages a table of contents lists.

Harvest reads each host's robots.txt first and obeys it for every request,
including the guesses. A URL it forbids is not requested and is reported once;
that is not a failure. A missing robots.txt means the site may be harvested,
and one with a mistake in it (a Crawl-delay line before the first User-agent
line, say) is read as far as it can be. A robots.txt that answers with a server
error or does not answer is asked for twice more, a second and two seconds
apart; if it still fails the host is not requested in that run and the link
fails as unavailable. A
Crawl-delay is honored up to 30 seconds; a host that asks for more is skipped
for the run. Requests to one host are made one at a time with a pause between
them, 500 ms unless the harvest setting in antenna.yaml says otherwise.

# PARAMETERS

COLLECTION_NAME
: (optional) harvest only this collection

# EXIT STATUS

0
: every feed was harvested, or left alone because robots.txt forbids it

2
: unknown option or extra argument

65
: a feed, table of contents or page could not be parsed, or a site has no
  feed to find

66
: the configuration or a collection file is missing

69
: a feed, or a host's robots.txt, could not be reached or answered with an error

74
: reading a response failed part way

78
: the configuration is wrong, for example a harvest pause that is not a duration

When some feeds fail the others are still harvested, and harvest exits with the
class of the first failure, after saying how many failed.

# ALIASES

fetch

# EXAMPLES

antenna harvest
antenna harvest feeds/tech.md

A collection file may list a site and let harvest find its feed:

  - [Dave Winer](https://dave.writersweb.org/)
  - [A WordPress blog](https://example.org/)

# SEE ALSO

antenna help configuration (the harvest setting), antenna help markdown-web


