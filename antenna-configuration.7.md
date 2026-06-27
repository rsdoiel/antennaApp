configuration — antenna.yaml and page.yaml settings reference

ANTENNA.YAML  (main configuration)

  port         (optional, default: 8000)   localhost port for 'preview'
  host         (optional, default: localhost) host name for 'preview'
  htdocs       (optional, default: ".")    directory for generated HTML/RSS
  generator    (optional, default: page.yaml) default page generator YAML
  collections  (required) list of collection objects

  Each collection object:
    file       (required) path to the collection Markdown document
    title      (optional, default: filename) display name
    generator  (optional) per-collection page generator YAML override
    mode       (optional) rendering mode: "aggregate" (default) or "page-index"
               "aggregate"  — feed-item cards from the items table (default)
               "page-index" — simple <ul> link list from the pages table

EXAMPLE antenna.yaml:

  htdocs: htdocs
  port: 8000
  collections:
    - file: index.md                 # aggregate (default)
    - file: links.md
      generator: links-page.yaml
    - file: pages.md
      mode: page-index               # renders a simple link list

PAGE.YAML  (page generator)

  lang               (optional, default: en-US) lang= attribute on <html>
  title              (optional) page <title> override
  meta               (optional) list of <meta> element attribute maps
  link               (optional) list of <link> element attribute maps
  script             (optional) list of <script> element attribute maps
  style              (optional) inline CSS injected at end of <head>
  header             (optional) innerHTML of <header>
  nav                (optional) innerHTML of <nav aria-label="Site navigation">
  top_content        (optional) content between <nav> and <main>
  bottom_content     (optional) content between </main> and <footer>
  footer             (optional) innerHTML of <footer>
  allowed_meta_fields (optional) allowlist of front matter keys to emit as <meta>

EXAMPLE page.yaml:

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
    <p>© 2026 Your Name</p>
  allowed_meta_fields:
    - title
    - author
    - description
    - keywords

SEE ALSO
  antenna help metadata
  antenna help accessibility

