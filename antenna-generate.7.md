%antenna(7) user manual | version 0.1.1 930fbba
% R. S. Doiel
% 2026-11-20

# NAME

generate

# SYNOPSIS

antenna generate [COLLECTION_NAME]

# DESCRIPTION

Processes all collections (or only COLLECTION_NAME if provided), rendering
an HTML page and RSS 2.0 feed for each. Output files are written to the
htdocs directory configured in antenna.yaml.

The HTML structure is controlled by the page generator YAML (page.yaml or
a per-collection override). Front matter from each item is emitted as
meta elements in the generated HTML.

A collection with items to show also gets a Markdown rendering of them,
STEM_items.md, beside its HTML page (not the pages collection, nor one with
markdownItems: false; markdownItems: true turns it on for pages), and the page declares it in its head as
its Markdown alternate. A collection with nothing to show has neither, and a
file left by an earlier run is removed.

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

# ALIASES

build

# EXAMPLES

antenna generate
antenna generate index.md


