%antenna(7) user manual | version 0.2.0 d1de725
% R. S. Doiel
% 2026-11-23

# NAME

migrate-links

# SYNOPSIS

antenna migrate-links [-w] [PATH]

# DESCRIPTION

Changes the links in your Markdown that point at an .html page antenna
renders from Markdown into links to the .md source. A Markdown reader such as
BonBon follows an .html link to the HTML page (BonBon opens it in the system
browser); the .md link keeps the reader in Markdown. This is the change
antenna validate reports as link-html.

By default it is a dry run: it prints each change as a unified diff and
changes nothing. Add -w to write the files. Review the diff first, and keep your
site in git so you can look at what was written.

With no PATH it works on the Markdown the site's own data names: the pages, the
collection source files and the posts. A PATH limits it to one file or the
Markdown files under one directory.

What it changes, and what it does not:

- Only a relative link to an .html or .htm page that antenna renders from a
  Markdown file. The fragment (#section) and query (?x=1) are kept. A link
  written from the site root (/blog/post.html) stays rooted (/blog/post.md).
- A link to a collection's page (news.html) becomes its items Markdown
  (news_items.md) when that file exists. If it does not exist yet the link is
  left, and a line says so.
- Inline links, links in angle brackets and reference definitions
  ([label]: destination). A title after the destination is kept.
- Never: code spans and code blocks, images, absolute addresses (including
  links to this site's own address), links to files that are not .html, and
  links to .html files antenna does not render, such as pages you wrote by
  hand.
- Only the bytes of each destination change. The front matter, the line endings
  and everything else are kept exactly. Each changed file is replaced by writing
  a new file beside it and renaming it.
- It does not run generate. The HTML antenna generates writes the .md links
  back as .html, so your pages look the same after the change. A post's stored
  copy is refreshed the next time you run antenna post on it.

Running it again changes nothing.

# OPTIONS

-w
: Write the changes. Without it nothing is written.

# EXIT STATUS

0
: It ran: a dry run, a write, or nothing to migrate.

2
: The command line is wrong (an unknown option, more than one PATH).

65
: A file could not be parsed. It is reported and the other files are still done;
the exit status is the class of the first failure.

66
: PATH does not exist, or there is no configuration file.

74
: A file could not be written.

# EXAMPLES

antenna migrate-links
antenna migrate-links blog
antenna migrate-links -w blog/index.md
antenna validate | grep link-html

# SEE ALSO

antenna help validate
antenna help markdown-web


