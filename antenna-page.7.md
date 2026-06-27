page — render a Markdown file as a standalone HTML page

SYNOPSIS
  antenna page INPUT_PATH [OUTPUT_PATH]

DESCRIPTION
  Renders INPUT_PATH as an HTML file using the page generator YAML.  The
  output file is placed next to INPUT_PATH with a .html extension unless
  OUTPUT_PATH is specified.  The page is recorded in the pages collection
  (pages.md / pages.db) so it appears in 'antenna pages' output.

  Pages are excluded from RSS feeds. This action is for static content such
  as About pages, contact pages, and search pages.

  WARNING: HTML in the Markdown source passes through unchanged (unsafe mode).
  Only run this action on files you control and trust.

PARAMETERS
  INPUT_PATH   path to the source Markdown file
  OUTPUT_PATH  (optional) explicit output HTML path

EXAMPLE
  antenna page about.md
  antenna page about.md htdocs/about.html

