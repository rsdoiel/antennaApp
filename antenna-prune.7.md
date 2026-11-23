%antenna(7) user manual | version 0.1.1 930fbba
% R. S. Doiel
% 2026-11-20

# NAME

prune

# SYNOPSIS

antenna prune [-w] [-all]

# DESCRIPTION

Finds the pages in the pages collection whose source file no longer exists, as
after an rm or mv, and removes them. With -w it deletes each page's record and
the output files the record lists (the HTML, and the Markdown copy when there is
one). Without -w it only lists them and changes nothing. unpage removes just
the record; prune is the clean-up for pages you have deleted.

The source file is never touched (it is gone), nor are files it included. An
output is kept, and reported on standard error, when another page still writes
it or when it is outside htdocs. The sitemap and toc.opml drop a pruned page the
next time you run antenna generate.

When the source of every page is missing, -w refuses, because that usually means
you are in the wrong directory or a drive is not mounted. Add -all to prune them
anyway.

# OPTIONS

-w
: delete the records and output files; without it prune is a dry run

-all
: with -w, prune even when every page's source is missing

# EXIT STATUS

0
: nothing to prune, or -w pruned the pages

1
: dry run found pages to prune, or -w refused because every source is missing

2
: unknown option or extra argument

66
: no pages collection or configuration found

74
: a file could not be deleted (the others are still pruned)

78
: the configuration is wrong

# EXAMPLES

Remove an old page, then see what prune would do, then do it:

~~~shell
rm old.md
antenna prune
antenna prune -w
antenna generate
~~~


