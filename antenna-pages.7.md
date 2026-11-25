%antenna(7) user manual | version 0.2.1 c3213bd
% R. S. Doiel
% 2026-11-25

# NAME

pages

# SYNOPSIS

antenna pages [-stale]

# DESCRIPTION

Prints the page entries tracked in the pages collection (pages.md / pages.db)
in descending updated-timestamp order, one per line, with six fields separated
by tabs:

1. the input path, the source file
2. the output path, the first .html the page writes. Deprecated: it is kept
   for older tools and will stop being written; use field 4
3. updated, when antenna last built the page
4. outputs, a JSON array of every file the page writes, in the order given to
   antenna page, for example ["index.md","index.html"]
5. includes, a JSON array of every file the source includes with
   @include-text-block or @include-code-block, nested ones too
6. stale, empty for a current page, "stale" when the source or a file it
   includes was changed after the page was last built, "missing" when an output
   file is not there

The first three fields are what older versions printed. Fields 4 and 5 are
empty for a page built before they were recorded; run antenna generate to
fill them in. The staleness test compares file times, so it is only meaningful
in a working copy that was built there, and it does not see changes to the
generator YAML, the theme or antenna.yaml.

With -stale only the pages whose sixth field is not empty are printed, and the
exit status is 1 when there are any, 0 when there are none.

Compare with posts (blog posts in a collection) and items (all items
in any collection database).

# EXAMPLES

antenna pages


