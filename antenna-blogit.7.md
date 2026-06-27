blogit — add a post using a date-based directory path

SYNOPSIS
  antenna blogit [COLLECTION_NAME] FILEPATH [POST_DATE]

DESCRIPTION
  A variation of 'post' that places the source Markdown document in a
  blog-style date directory tree (e.g. blog/2026/04/12/my-post.md) before
  adding it to the collection.  If POST_DATE is omitted, the current date is
  used.  The collection defaults to pages.md when COLLECTION_NAME is omitted.

PARAMETERS
  COLLECTION_NAME  (optional) path to the collection Markdown file
  FILEPATH         source Markdown document
  POST_DATE        (optional) date in YYYY-MM-DD format

EXAMPLE
  antenna blogit index.md notes/my-idea.md
  antenna blogit index.md notes/my-idea.md 2026-04-01

