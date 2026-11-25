%antenna(7) user manual | version 0.2.1 8278d2c
% R. S. Doiel
% 2026-11-25

# NAME

post

# SYNOPSIS

antenna post [COLLECTION_NAME] FILEPATH

# DESCRIPTION

Adds FILEPATH to COLLECTION_NAME (default: pages.md). The file's YAML
front matter supplies the required fields. If a record with the same
postPath already exists, it is overwritten.

Required front matter fields:
  title or description  at least one must be present
  postPath              relative path to the generated HTML file
  pubDate               publication date (YYYY-MM-DD recommended)

Recommended additional fields:
  link        public URL to the post
  author      author name
  description summary for RSS and search engines
  keywords    list of tags

After posting, run generate to rebuild the collection HTML and RSS feed.

A post made by antenna always has a guid, which must be unique and a URL: a
guid in the front matter is kept, otherwise it is the post's link, and the feed
writes it as a permalink (isPermaLink="true"). A post stored before this with
no guid is given the same at output time, so readers that tell items apart by
link see no change.

When the collection has html: false (Markdown-only), post writes no HTML and
the item's link is the post's own address, base_url plus the postPath, which
ends .md. A post made before the collection was switched keeps the link it has.

WARNING: HTML in the Markdown source passes through unchanged (unsafe mode).
Only run this action on files you control and trust.

# PARAMETERS

COLLECTION_NAME
: (optional) collection Markdown file (default: pages.md)

FILEPATH
: path to the source Markdown document

# EXAMPLES

antenna post index.md blog/2026/04/12/my-post.md

# SEE ALSO

antenna help metadata
antenna help blogit


