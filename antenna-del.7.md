del — remove a collection from the configuration

SYNOPSIS
  antenna del COLLECTION_FILE

DESCRIPTION
  Removes the collection identified by COLLECTION_FILE from antenna.yaml.
  The Markdown file and its SQLite3 database are not deleted from disk.

PARAMETERS
  COLLECTION_FILE  path to the Markdown document as registered with 'add'

EXAMPLE
  antenna del feeds/tech.md

