add — add a feed collection

SYNOPSIS
  antenna add COLLECTION_FILE [NAME DESCRIPTION]

DESCRIPTION
  Registers a feed collection with antenna. COLLECTION_FILE is a Markdown
  document (.md) whose body lists RSS/Atom feed URLs as hyperlinks.  NAME and
  DESCRIPTION are optional; if omitted, antenna reads them from the
  document's YAML front matter.

  antenna creates a matching SQLite3 database (same basename, .db extension)
  and records the collection in antenna.yaml.

PARAMETERS
  COLLECTION_FILE  path to the Markdown document defining the collection
  NAME             (optional) override the collection name
  DESCRIPTION      (optional) override the collection description

EXAMPLE
  antenna add feeds/tech.md
  antenna add feeds/tech.md "Tech Feeds" "My technology reading list"

