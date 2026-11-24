%antenna(7) user manual | version 0.1.2 24f4114
% R. S. Doiel
% 2026-11-23

# NAME

unpage

# SYNOPSIS

antenna unpage INPUT_PATH

# DESCRIPTION

Removes the page record associated with INPUT_PATH from the pages collection.
The HTML and Markdown files on disk are not deleted. When no page record
matches INPUT_PATH, antenna prints an error and exits 1.

# PARAMETERS

INPUT_PATH
: path to the source Markdown file as it was passed to page

# EXAMPLES

antenna unpage about.md


