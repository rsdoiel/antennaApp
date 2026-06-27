quote — convert a text-fragment URL into a Markdown excerpt

SYNOPSIS
  antenna quote TEXT_FRAGMENT_URL

DESCRIPTION
  Parses a TEXT_FRAGMENT_URL (a URL with a #:~:text= fragment) into a
  Markdown quotation. Output is written to standard out. Redirect it into a
  file to use as the basis for a response post.

ALIASES
  reply

PARAMETERS
  TEXT_FRAGMENT_URL  a URL ending in a #:~:text= text-fragment selector

EXAMPLE
  antenna quote "https://example.com/article#:~:text=interesting%20passage"

