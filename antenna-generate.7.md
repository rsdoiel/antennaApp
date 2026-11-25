%antenna(7) user manual | version 0.2.1 c3213bd
% R. S. Doiel
% 2026-11-25

# NAME

generate

# SYNOPSIS

antenna generate [COLLECTION_NAME]

# DESCRIPTION

Processes all collections (or only COLLECTION_NAME if provided), rendering
an HTML page and RSS 2.0 feed for each. The HTML page of a collection that
aggregates items is STEM_items.html, beside its Markdown rendering
STEM_items.md; your own STEM.md, the list of feeds, has no HTML page. A
collection in page-index mode and the pages collection keep STEM.html. Output
files are written to the htdocs directory configured in antenna.yaml.

The HTML structure is controlled by the page generator YAML (page.yaml or
a per-collection override). Front matter from each item is emitted as
meta elements in the generated HTML.

A collection with items to show also gets a Markdown rendering of them,
STEM_items.md, beside STEM_items.html (not the pages collection, nor one with
markdownItems: false; markdownItems: true turns it on for pages), and the page
declares it in its head as its Markdown alternate. A collection with nothing to
show has neither file, and a file left by an earlier run is removed.

Earlier releases wrote the items page at STEM.html. The first generate that
writes STEM_items.html removes that old page, when antenna's generator tag is
in it and no pages row records it (a page you rendered from STEM.md with
antenna page is kept), and says so once on standard error: links to STEM.html
now fail, and the items are at STEM_items.html. A STEM.html that is not
antenna's is left alone with a warning. Removing the old page is not a failure.

A collection with html: false in antenna.yaml is Markdown-only: generate writes
its items Markdown, RSS and feed-list OPML and no HTML, neither the items page
nor its posts, and removes an items page left by an earlier run. Post
HTML from before the switch is left alone; generate says how many files remain.
See antenna help configuration.

A feed shows at most max_items_per_feed of its newest items (1000 unless
antenna.yaml says otherwise), or its own number in feed_limits. The older items
stay in the database and are only left off the pages; when a feed is cut,
generate warns once for the collection, for example "2 feeds over their cap:
Korea Times shows 1000 of 1771, ...". See antennaApp DR-0039 and
antenna help configuration.

A harvested item's text keeps the links and images as its own site wrote them,
and many are relative (/images/a.png, /2026/post/). When it is rendered, in the
HTML page, in STEM_items.md and in the RSS feed, each relative address is made
absolute against the item's link, so it points at the site the item came from.
Absolute addresses, mailto:, tel:, data: and #fragment are left as written, and
so are your own posts. The stored item is not changed. An item with no usable
link (not an absolute http or https address) has nothing to resolve against: it
is left as written and generate warns once for the collection with the number of
items. See antennaApp DR-0037.

Output is the same every time it is generated from the same data. A file is
rewritten only when its content changes. The lines that record when it was
built (the date meta element of an HTML page, dateModified in an items file,
lastBuildDate in an RSS feed and dateCreated in OPML) do not count, so
generating an unchanged site changes no file or modification time, and those
dates say when the content last changed, not when generate last ran.

When the whole site is generated (no COLLECTION_NAME), generate also writes a
table of contents for Markdown readers such as BonBon and cmarkbrowser,
htdocs/.well-known/toc.opml: one OPML outline with the site's pages, the items
pages of its collections and its posts (oldest first, by year and month). Only
pages and posts whose Markdown is published in htdocs are listed. The addresses
are made from base_url in antenna.yaml, which it needs: without one a warning
is given and the file is not written. Like the other files it is rewritten only
when its content changes.

Sites listed in the subsites setting of antenna.yaml are added to the outline as
include outlines (type="include", url=SITE/.well-known/toc.opml) after those
groups, so the table of contents at a host's root leads BonBon to its other
sites. Nothing is fetched. If the bonbon setting of antenna.yaml has a key,
generate also writes .well-known/bonbon.json from it; with none it writes no
file and leaves any file you wrote alone. See antenna help configuration.

The same outline gives previous and next links to the posts it lists that name
neither urlPrev nor urlNext in their front matter: the older post as previous
and the newer as next. The first and last have only one. A page gets none from
the table of contents, because pages are listed in file-name order, which is not
a sequence. A page or post in a sequence named in the sequences setting of
antenna.yaml gets that sequence's previous and next instead, and the generated
Markdown copy of such a page carries them as urlPrev and urlNext. The author's
own keys always win, and nothing is written into your Markdown.

# PARAMETERS

COLLECTION_NAME
: (optional) process only this collection

# EXIT STATUS

0
: every collection was generated

2
: unknown option or extra argument

65
: an item row could not be read; every file that could be written was written
  and the failure is reported after them

66
: the configuration or a collection file is missing

74
: a file could not be read or written part way, including an old page that
  could not be removed

78
: the configuration is wrong, for example items.html with an unknown value or
  html: false with markdownItems: false

# ALIASES

build

# EXAMPLES

antenna generate
antenna generate index.md


