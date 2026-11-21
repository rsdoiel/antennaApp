%antenna(7) user manual | version 0.1.0 f9d2bc2
% R. S. Doiel
% 2026-11-20

# NAME

page

# SYNOPSIS

antenna page INPUT_PATH [OUTPUT_PATH...]

# DESCRIPTION

Renders INPUT_PATH to each OUTPUT_PATH named. The extension of an output says
what it is: .html is an HTML page made with the page generator YAML, and .md
is the Markdown of the source with its include directives
(@include-text-block and @include-code-block) expanded. The same expanded text
is used for both, so they are one document. With no OUTPUT_PATH one HTML page
is written next to INPUT_PATH with a .html extension.

The source can be a text file with directives (index.txt), which is how a
page that includes other files publishes both a page and a Markdown copy
that a reader such as BonBon can use:

    antenna page index.txt index.md index.html

In the Markdown copy the front matter is kept as written, with two keys added
when the source has none: a title from the first level-1 heading, and the
site-wide author from antenna.yaml. Links stay as the author wrote them (.md
links stay .md). The source file is never changed. The HTML page declares the Markdown copy
beside it (same folder, same stem) with <link rel="alternate"
type="text/markdown">, so a reader that cannot ask a server for Markdown can
find it. In the HTML, a link to a Markdown file
antenna renders becomes a link to its .html page.

Every run is recorded in the pages collection (pages.md / pages.db): the
outputs it wrote, in order, and the files the source includes, nested ones too.
antenna pages shows them and antenna generate rebuilds from them. A run
that writes only Markdown is recorded too. A run replaces what the last run
recorded for the same source, so name every output in one command. A file that
is only included into other pages (recent.txt) is never given to antenna page.

An include is expanded where its directive starts a line. An
@include-text-block file is expanded in turn, so it can include others; nesting
deeper than three levels draws a warning, and deeper than five levels, or a file
that includes itself, is an error (exit 65). An @include-code-block file is
shown as it is and never expanded.

Pages are excluded from RSS feeds. This action is for static content such
as About pages, contact pages, and search pages.

An output that is the source file is refused (exit 65) and nothing is written.
An output that does not end in .html or .md is a usage error (exit 2), and
nothing is written. A Markdown output needs a text source, not an ODT file.

WARNING: HTML in the Markdown source passes through unchanged (unsafe mode).
Only run this action on files you control and trust.

# PARAMETERS

INPUT_PATH
: path to the source file

OUTPUT_PATH
: (optional, any number) an output path ending in .html or .md

# EXAMPLES

antenna page about.md
antenna page about.md htdocs/about.html


