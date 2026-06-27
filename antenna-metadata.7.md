metadata — front matter fields and the allowed_meta_fields allowlist

DESCRIPTION
  Every YAML front matter field in a Markdown post or page is emitted into
  the generated HTML <head> as a <meta name="KEY" content="VALUE"> element.
  The same key-value pair is also written as a data-pagefind-filter attribute
  on the enclosing <article> element for PageFind faceted search.

STANDARD FIELDS

  title          Sets <title> in the HTML head (not emitted as <meta>)
  description    Short summary for search engines and RSS
  author         Author name
  keywords       List of tags; each value gets its own <meta> pair
  pubDate        Publication date — required for posts (YYYY-MM-DD)
  postPath       Relative path to the generated HTML file — required for posts
  link           Public URL of the post
  series         Series name for multi-part posts
  seriesNumber   Position within the series
  dateCreated    Creation date
  dateModified   Last-modified date
  datePublished  Publication date (alias for pubDate in page context)

EXAMPLE FRONT MATTER:

  ---
  title: "My Post Title"
  description: "A short summary."
  author: "Your Name"
  keywords:
    - go
    - web
  pubDate: "2026-06-27"
  postPath: "blog/2026/06/27/my-post.html"
  series: "Building with Antenna"
  seriesNumber: 3
  ---

CONTROLLING WHICH FIELDS ARE PUBLISHED (allowed_meta_fields)

  By default all front matter keys are emitted as <meta> elements. To
  restrict which keys are published, set allowed_meta_fields in page.yaml:

    allowed_meta_fields:
      - title
      - author
      - description
      - keywords
      - series
      - seriesNumber

  When this list is set, only the listed keys appear in generated HTML.
  Internal workflow fields (build flags, notes) are silently omitted.

SEE ALSO
  antenna help configuration
  antenna help post

