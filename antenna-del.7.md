%antenna(7) user manual | version 0.2.1 1280a8e
% R. S. Doiel
% 2026-11-25

# NAME

del

# SYNOPSIS

antenna del COLLECTION_FILE

# DESCRIPTION

Removes the collection identified by COLLECTION_FILE from antenna.yaml.
The Markdown file and its SQLite3 database are not deleted from disk.

# PARAMETERS

COLLECTION_FILE
: path to the Markdown document as registered with add

# EXAMPLES

antenna del feeds/tech.md


