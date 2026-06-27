rss — generate an RSS feed file from posts

SYNOPSIS
  antenna rss COLLECTION_NAME RSS_FILENAME [COUNT | FROM_DATE TO_DATE]

DESCRIPTION
  Writes an RSS 2.0 feed for COLLECTION_NAME to RSS_FILENAME. The optional
  parameters work the same way as for 'posts': a COUNT or a FROM_DATE/TO_DATE
  range to limit the items included.

PARAMETERS
  COLLECTION_NAME  collection Markdown file
  RSS_FILENAME     output path for the RSS file
  COUNT            (optional) maximum number of items
  FROM_DATE        (optional) start date (requires TO_DATE)
  TO_DATE          (optional) end date (requires FROM_DATE)

EXAMPLE
  antenna rss index.md index.xml
  antenna rss index.md archive.xml 2026-01-01 2026-06-30

