%antenna(7) user manual | version 0.1.1 930fbba
% R. S. Doiel
% 2026-11-20

# NAME

metadata

# DESCRIPTION

Every YAML front matter field in a Markdown post or page is emitted into
the generated HTML <head> as a <meta name="KEY" content="VALUE"> element.
The same key-value pair is also written as a data-pagefind-filter attribute
on the enclosing <article> element for PageFind faceted search.

# STANDARD FIELDS

Keys are camelCase. The eight marked (BonBon) are the ones Dave Winer's BonBon
reads, so one Markdown file serves both BonBon and the HTML site. Any other key
passes through unchanged.

title
: (BonBon) Sets <title> in the HTML head (not emitted as <meta>). Where the
  text says {{page.title}} it is expanded to this value (see MACROS).

description
: (BonBon) Short summary for search engines and RSS

author
: (BonBon) Author name

pubDate
: (BonBon) Publication date, required for posts. YYYY-MM-DD or an RSS 2.0 date
  such as Fri, 23 Oct 2026 15:19:45 GMT (see DATES)

link
: (BonBon) Public URL of the post. Where the title is shown through
  {{page.title}} it is a link to this address (see MACROS)

urlPrev
: (BonBon) The previous page. Written as <link rel="prev"> in the head and a
  Previous link on the page (see PREVIOUS AND NEXT)

urlNext
: (BonBon) The next page. Written as <link rel="next"> in the head and a Next
  link on the page (see PREVIOUS AND NEXT)

rightMarginImage
: (BonBon) An image floated to the right at the top of the text (see MARGIN
  IMAGE)

rightMarginImageAlt
: The alt text for rightMarginImage. antenna's own key, BonBon ignores it

keywords
: List of tags; each value gets its own <meta> pair

postPath
: Relative path to the generated HTML file required for posts

series
: Series name for multi-part posts

seriesNumber
: Position within the series

dateCreated
: Creation date

dateModified
: Last-modified date

datePublished
: Publication date (alias for pubDate in page context)

# DATES

pubDate may be written as 2026-06-27, as a timestamp such as
2026-06-27T10:00:00Z (quoted or not), or in the RSS 2.0 form BonBon uses,
Fri, 23 Oct 2026 15:19:45 GMT. The weekday may be left out, the day may have
one or two digits, and the zone may be GMT, UT, a US zone such as EST or PDT,
or an offset such as +0200. A date with a zone keeps it: antenna stores
"Mon, 30 Nov 2026 21:30:00 -0800" as 2026-11-30T21:30:00-08:00, so the post is
on November 30 in the table of contents, the feeds and the page, the day you
wrote it, wherever the reader is. A date in GMT, or with no zone, is stored as
YYYY-MM-DD HH:MM:SS. Posts sort by the moment they were made. Your Markdown
file is never changed. A date antenna cannot read is kept as written and
reported with a warning; it does not stop the command.

# MACROS

In the text of a page or post, {{page.NAME}} is replaced by the front matter
value NAME (spaces inside the braces are allowed), so a page can start with a
first-level heading that reads {{page.title}}. The value is inserted as plain
text: it never becomes
Markdown or HTML. Macros are not expanded inside code spans or code blocks, so
a page can show one. Expansion is a single pass; a value that itself holds a
macro is shown as written. A name the page does not have leaves the macro
visible and prints one warning. When the page has a link, the expanded title
links to it (http, https, mailto or a relative address). Your published
Markdown keeps the macros: BonBon expands them for its readers.

# PREVIOUS AND NEXT

urlPrev and urlNext name the previous and next page. A same-site .md address is
written as the .html page, like any other link. The page gets <link rel="prev">
and <link rel="next"> in the head and, after its text, a labelled navigation
with a Previous and a Next link; only the directions you name are shown. An
address that is not http, https, mailto or relative is ignored with a warning.

A post that gives neither key gets them when the site is generated, from the
site's table of contents (see antenna help generate): the older post as previous
and the newer as next. A page gets none, since the table of contents lists pages
in file-name order, which is not a sequence; a page or post in a sequence named
in the sequences setting of antenna.yaml gets that sequence's neighbours
instead (see antenna help configuration). A key you write always wins over
these.

# MARGIN IMAGE

rightMarginImage is the address of an image (http, https or a relative path).
It is shown floated to the right at the top of the text, and on a narrow screen
above the text. BonBon's key has no alt text, so set rightMarginImageAlt to
describe the image; without it the image is treated as decoration and has an
empty alt. Neither key is also written as a <meta> element.

# EXAMPLE FRONT MATTER

---
title: "My Post Title"
description: "A short summary."
author: "Your Name"
keywords:
  - go
  - web
pubDate: "2026-06-27"
postPath: "blog/2026/06/27/my-post.html"
series: "Building with Antenna"
seriesNumber: 3
urlPrev: "my-earlier-post.md"
rightMarginImage: "images/cover.jpg"
rightMarginImageAlt: "The book cover"
---

# CONTROLLING WHICH FIELDS ARE PUBLISHED (allowed_meta_fields)

By default all front matter keys are emitted as <meta> elements. To
restrict which keys are published, set allowed_meta_fields in page.yaml:

  allowed_meta_fields:
    - title
    - author
    - description
    - keywords
    - series
    - seriesNumber

When this list is set, only the listed keys appear in generated HTML.
Internal workflow fields (build flags, notes) are silently omitted.

# SEE ALSO

antenna help configuration
antenna help post


