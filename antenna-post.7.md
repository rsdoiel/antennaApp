post — add or update a blog post in a collection

SYNOPSIS
  antenna post [COLLECTION_NAME] FILEPATH

DESCRIPTION
  Adds FILEPATH to COLLECTION_NAME (default: pages.md). The file's YAML
  front matter supplies the required fields. If a record with the same
  postPath already exists, it is overwritten.

  Required front matter fields:
    title or description  — at least one must be present
    postPath              — relative path to the generated HTML file
    pubDate               — publication date (YYYY-MM-DD recommended)

  Recommended additional fields:
    link        public URL to the post
    author      author name
    description summary for RSS and search engines
    keywords    list of tags

  After posting, run 'generate' to rebuild the collection HTML and RSS feed.

  WARNING: HTML in the Markdown source passes through unchanged (unsafe mode).
  Only run this action on files you control and trust.

PARAMETERS
  COLLECTION_NAME  (optional) collection Markdown file (default: pages.md)
  FILEPATH         path to the source Markdown document

EXAMPLE
  antenna post index.md blog/2026/04/12/my-post.md

SEE ALSO
  antenna help metadata
  antenna help blogit

