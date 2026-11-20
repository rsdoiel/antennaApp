%antenna(7) user manual | version 0.1.0 f9d2bc2
% R. S. Doiel
% 2026-11-20

# NAME

items

# SYNOPSIS

antenna items [COLLECTION_NAME]

# DESCRIPTION

Prints a Markdown list of all items stored in the named collection's
SQLite3 database. Items include blog posts added via post or blogit,
and feed entries harvested via harvest. The list is in descending
publication-date order.

Each line shows: - [Title](link), YYYY-MM-DD, status (label)

If COLLECTION_NAME is omitted it defaults to pages.md.

See posts for a list restricted to blog posts (items with a postPath),
and pages for a list of static page entries.

# PARAMETERS

COLLECTION_NAME
: (optional) collection Markdown file (default: pages.md)

# EXAMPLES

antenna items
antenna items index.md


