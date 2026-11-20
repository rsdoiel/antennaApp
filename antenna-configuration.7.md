%antenna(7) user manual | version 0.1.0 f9d2bc2
% R. S. Doiel
% 2026-11-20

# NAME

configuration

# DESCRIPTION

antenna uses configuration files to control its behavior.

# ANTENNA.YAML (main configuration)

port
: (optional, default: 8000) localhost port for preview

host
: (optional, default: localhost) host name for preview

htdocs
: (optional, default: ".") directory for generated HTML/RSS

generator
: (optional, default: page.yaml) default page generator YAML

base_url
: (optional) the address the site is published at, with no trailing slash.
  The feeds, the sitemap and the table of contents (.well-known/toc.opml) are
  made from it; without it the table of contents is not written

subsites
: (optional) other Markdown sites on the same host, so this site's table of
  contents can point at them. A list of objects, each with url (required: the
  sub-site's root, an absolute http or https address with a host, no query or
  fragment) and title (optional label). generate writes one include outline
  per entry, after the pages, collections and posts, in the order listed, at
  url/.well-known/toc.opml. BonBon asks only the root of a host for a table of
  contents and follows an include to reach a sub-site, so a host with several
  sites lists them here in the root site's antenna.yaml. Write the address as
  the sub-site's own pages are published (same scheme and host). Nothing is
  fetched: generate does not check that a sub-site exists. A missing or
  relative url, a duplicate, or an entry that is not a list of objects is a
  configuration error (exit 78)

bonbon
: (optional) the content of .well-known/bonbon.json, a mapping of keys to
  values. When it has at least one key, generate writes the file (a JSON object,
  keys sorted), and rewrites it only when it changes. BonBon reads three keys
  and only as strings: urlSearch, urlUpdateSocket and urlMirrorOf; any other key
  is written as given and ignored by BonBon today. A known key that is empty or
  not a string draws a warning, because BonBon keeps an empty one and a search
  then goes to a broken address: leave a key out instead of blanking it. With
  no bonbon setting antenna writes no file and never removes one, so a file
  you wrote by hand is safe; if you take the setting out, delete the file
  yourself. antenna runs no search or update service; see
  antenna help markdown-web

collections
: (required) list of collection objects

Each collection object:
  file
  : (required) path to the collection Markdown document

  title
  : (optional, default: filename) display name

  generator
  : (optional) per-collection page generator YAML override

  mode
  : (optional) rendering mode: "aggregate" (default) or "page-index"
     "aggregate"  feed-item cards from the items table (default)
     "page-index" simple <ul> link list from the pages table

Example antenna.yaml:

  htdocs: htdocs
  port: 8000
  base_url: https://example.org
  subsites:
    - url: https://example.org/opml
      title: opml
    - url: https://example.org/antenna
  bonbon:
    urlSearch: https://example.org/search
  collections:
    - file: index.md                 # aggregate (default)
    - file: links.md
      generator: links-page.yaml
    - file: pages.md
      mode: page-index               # renders a simple link list

# PAGE.YAML (page generator)

lang
: (optional, default: en-US) lang= attribute on <html>

title
: (optional) page <title> override

meta
: (optional) list of <meta> element attribute maps

link
: (optional) list of <link> element attribute maps

script
: (optional) list of <script> element attribute maps

style
: (optional) inline CSS injected at end of <head>

header
: (optional) innerHTML of <header>

nav
: (optional) innerHTML of <nav aria-label="Site navigation">

top_content
: (optional) content between <nav> and <main>

bottom_content
: (optional) content between </main> and <footer>

footer
: (optional) innerHTML of <footer>

allowed_meta_fields
: (optional) allowlist of front matter keys to emit as <meta>

Example page.yaml:

  lang: en-US
  link:
    - rel: stylesheet
      type: text/css
      href: /css/site.css
  header: |
    <h1>My Blog</h1>
  nav: |
    <ul>
      <li><a href="/">Home</a></li>
      <li><a href="/about.html">About</a></li>
    </ul>
  footer: |
    <p>&copy; 2026 Your Name</p>
  allowed_meta_fields:
    - title
    - author
    - description
    - keywords

# SEE ALSO

antenna help metadata
antenna help accessibility


