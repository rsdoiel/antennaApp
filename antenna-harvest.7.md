harvest — fetch content from remote feeds

SYNOPSIS
  antenna harvest [COLLECTION_NAME]

DESCRIPTION
  Retrieves RSS/Atom feed content for all collections (or only
  COLLECTION_NAME) and stores harvested items in each collection's SQLite3
  database. Run 'generate' afterwards to rebuild the HTML pages.

PARAMETERS
  COLLECTION_NAME  (optional) harvest only this collection

ALIASES
  fetch

EXAMPLE
  antenna harvest
  antenna harvest feeds/tech.md

