%antenna(7) user manual | version 0.1.1 930fbba
% R. S. Doiel
% 2026-11-20

# NAME

preview

# SYNOPSIS

antenna preview

# DESCRIPTION

Starts a local HTTP server serving the htdocs directory so you can review
the generated site in a browser. The host and port are set in antenna.yaml
(defaults: localhost:8000).

Press Ctrl-C to stop the server.

The preview also serves the Markdown web, so a Markdown reader such as BonBon
or cmarkbrowser can be tried before anything is published:

- Files ending .md and .markdown are sent as text/markdown; charset=utf-8 and
  .opml files as text/x-opml; charset=utf-8, whatever the host's own table says.
- The .well-known directory (in any folder, for toc.opml and bonbon.json) is
  served. Every other file or directory whose name starts with a dot is
  refused with 403, and so is anything inside .well-known that does.
- A client whose Accept header prefers text/markdown to text/html (BonBon sends
  "text/markdown, text/x-opml") that asks for an HTML page or a folder is sent
  Markdown instead: the page the HTML head declares with
  <link rel="alternate" type="text/markdown">, else the page's .md sibling (x.md,
  or index.md in a folder), else the HTML. Content-Location names the file sent
  and Vary: Accept is set. A browser gets the HTML as always.

Negotiation is a convenience of the preview. A static host cannot do it; what
works everywhere is the declared alternate and ordinary links to .md files.

# EXAMPLES

antenna preview
open http://localhost:8000
curl -H 'Accept: text/markdown' http://localhost:8000/blog/index.html


