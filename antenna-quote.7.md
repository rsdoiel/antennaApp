%antenna(7) user manual | version 0.2.1 c3213bd
% R. S. Doiel
% 2026-11-25

# NAME

quote

# SYNOPSIS

antenna quote TEXT_FRAGMENT_URL

# DESCRIPTION

Parses a TEXT_FRAGMENT_URL (a URL with a #:~:text= fragment) into a
Markdown quotation. Output is written to standard out. Redirect it into a
file to use as the basis for a response post.

# ALIASES

reply

# PARAMETERS

TEXT_FRAGMENT_URL
: a URL ending in a #:~:text= text-fragment selector

# EXAMPLES

antenna quote "https://example.com/article#:~:text=interesting%20passage"


