%antenna(7) user manual | version 0.0.28 1e5ac36
% R. S. Doiel
% 2026-10-05

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


