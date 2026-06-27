generate — render HTML pages and RSS feeds

SYNOPSIS
  antenna generate [COLLECTION_NAME]

DESCRIPTION
  Processes all collections (or only COLLECTION_NAME if provided), rendering
  an HTML page and RSS 2.0 feed for each. Output files are written to the
  htdocs directory configured in antenna.yaml.

  The HTML structure is controlled by the page generator YAML (page.yaml or
  a per-collection override). Front matter from each item is emitted as
  <meta> elements in the generated HTML.

PARAMETERS
  COLLECTION_NAME  (optional) process only this collection

ALIASES
  build

EXAMPLE
  antenna generate
  antenna generate index.md

