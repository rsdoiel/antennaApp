posts — list posts in a collection

SYNOPSIS
  antenna posts COLLECTION_NAME [COUNT | FROM_DATE TO_DATE]

DESCRIPTION
  Prints a Markdown list of blog posts in COLLECTION_NAME in descending
  pubDate order. Items must have both a pubDate and a postPath to appear.

  Optional parameters constrain the list:
    COUNT           integer — return only the N most recent posts
    FROM_DATE       YYYY-MM-DD start of a date range
    TO_DATE         YYYY-MM-DD end of a date range

PARAMETERS
  COLLECTION_NAME  collection Markdown file
  COUNT            (optional) maximum number of posts to list
  FROM_DATE        (optional) start date (requires TO_DATE)
  TO_DATE          (optional) end date (requires FROM_DATE)

EXAMPLE
  antenna posts index.md
  antenna posts index.md 10
  antenna posts index.md 2026-01-01 2026-06-30

