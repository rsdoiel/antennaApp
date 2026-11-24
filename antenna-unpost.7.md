%antenna(7) user manual | version 0.1.2 24f4114
% R. S. Doiel
% 2026-11-23

# NAME

unpost

# SYNOPSIS

antenna unpost COLLECTION_NAME URL | POST_PATH

# DESCRIPTION

Removes an item from COLLECTION_NAME using either the public URL associated
with the post or the POST_PATH value. The source Markdown and generated HTML
files on disk are not deleted.

# PARAMETERS

COLLECTION_NAME
: collection Markdown file

URL or POST_PATH
: URL or postPath value identifying the item to remove

# EXAMPLES

antenna unpost index.md https://example.com/blog/2026/04/12/my-post.html
antenna unpost index.md blog/2026/04/12/my-post.html


