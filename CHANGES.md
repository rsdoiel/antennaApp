# Changes

## Unreleased

- `antenna validate`: new check `link-fragment` (warning). A link to `#heading`, or to `other.md#heading`, is checked against the headings of that Markdown file, using markdownweb's `slug` package for the fragment rule. markdownweb is pinned at a commit after v0.0.5 until its next release.

## v0.2.2 — 2026-11-26

### Changed: markdownweb is v0.0.4

- `github.com/rsdoiel/markdownweb` v0.0.3 to v0.0.4. antenna's own previous and
  next arrows do not use `toc.Neighbours` and are unchanged by it. What v0.0.4
  changes for antenna: the table-of-contents `Finder` reads through `fetch`
  (the program's User-Agent and settings), and `toc.Finder.Pages` flattens an
  outline as BonBon does (markdownweb DR-0003). `toc.Neighbours` now follows
  BonBon's rule (markdownweb DR-0004), which is why the posts' arrows change
  below.

### Fixed: a post's previous and next wrap at the ends, as BonBon's do (antennaApp DR-0040)

- The oldest post's previous is now the newest post, and the newest post's next
  is the oldest, as in BonBon and cmarkbrowser (v0.0.3), which read the same
  table of contents and wrap. They had one arrow each before (DR-0025 3). The
  HTML of the oldest and newest post of every site with two or more posts
  changes on the next `generate`; one post still has none, and two point at each
  other.
- Unchanged: pages and collections get none (DR-0025 2); a declared sequence
  keeps its ends (DR-0026), which is left open in DR-0040; a document's own
  `urlPrev` and `urlNext` still win. antenna still wraps inside the posts, where
  BonBon would go on into the group above or below them.
- `antenna help generate` (and `antenna-generate.7.md`) say the ends wrap.

### Changed: harvest fetches through markdownweb's `fetch`

- The HTTP client of harvest is `github.com/rsdoiel/markdownweb/fetch`
  (markdownweb DR-0002); antenna keeps its pause, its User-Agent, its five
  redirects and its exit classes, and still does not read `robots.txt`
  (DR-0035). New: a 429 or 503 with a `Retry-After` of up to two minutes is
  waited out and tried once more, a longer one is unavailable (69), and a feed or
  table of contents over 32 MiB is a data error (65).
- Documented: `userAgent` in `antenna.yaml` sets the User-Agent harvest sends; empty
  or missing means `antenna/VERSION HASH`. It already worked; it was not in
  `antenna help configuration`. Tested through `antenna harvest`.

### Changed: front matter, macros and includes come from `frontmatter` v0.0.1

- `SplitFrontMatter`, the `{{page.NAME}}` expander and the include directives are
  now `github.com/rsdoiel/frontmatter` calls; the copies in antenna are gone.
  Checked on a copy of the blog site: every page (199) renders to the same HTML,
  Markdown copy, warnings and exit status as before, and `generate` gives the
  same files, twice.
- New: a dotted key reaches into a nested table, `{{page.source.title}}`. A
  leading UTF-8 byte-order mark in a source file is skipped.

## v0.2.1 — 2026-11-25

### Added: a cap on the items each feed shows (DR-0039)

- A feed shows at most its newest `max_items_per_feed` items (no cap when the
  setting is absent or 0; `antenna init` writes 1000 into a new `antenna.yaml` and
  leaves an existing one alone) in a collection's HTML page, `STEM_items.md` and `STEM.xml`; the
  top-level `feed_limits` mapping sets a number for a single feed by its address.
  Items are hidden when the pages are made, not deleted or re-statused. The
  site's own posts are never capped. A negative or non-whole number is a
  configuration error (exit 78).
- `generate` warns once for a collection when feeds were cut, naming each with
  how many it shows and how many it has.

### Added: harvest records each feed's status and warns about failing feeds (DR-0038)

- Every attempt to fetch a listed feed is written to a new `harvest_log` table
  in the collection database (made on the first harvest of an older database;
  rows kept 90 days).
- A feed whose attempts all failed over the last 7 days, at least 3 of them,
  gets a warning at the end of the harvest on standard error, with the HTTP
  status in words (`last status 404 Not Found`). 429 and 503 count as failures.
  The exit status is unchanged and nothing is skipped or removed.
- Settings under `harvest:` in `antenna.yaml`: `failing_days`, `failing_attempts`
  and `log_days`; a value below 1 is a configuration error (exit 78).

### Fixed: relative links and images in harvested items (DR-0037)

- A harvested item's relative `href` and `src` (root-relative, page-relative,
  protocol-relative) are made absolute against the item's link when it is
  rendered: in `STEM_items.html`, `STEM_items.md` and `STEM.xml` (description and
  `source:markdown`). They used to resolve against the site that aggregates the
  items and break. Absolute, `mailto:`, `tel:`, `data:` and `#fragment` values
  and the site's own posts are left. The stored item is not changed.
- An item with no usable link keeps its body as written, and `generate` warns
  once for the collection with the count.

## v0.2.0 — 2026-11-23

### Changed: generated file names (DR-0034, amends DR-0033)

This is a breaking change to generated file names, so the next release is a
minor one at least.

- **The aggregated items page is `STEM_items.html`**, beside `STEM_items.md`
  (it was `STEM.html`). The author's `STEM.md`, the feed list, has no HTML
  page. A collection in `mode: page-index` and the pages collection keep
  `STEM.html`; a Markdown-only collection still writes no HTML.
- **A collection with nothing to show writes no items page**, and a stale
  `STEM_items.html` is removed (it used to write an empty `STEM.html`).
- **The `STEM.html` an earlier release wrote is removed** on the first
  `generate`, with one note on standard error, when it carries antenna's
  generator tag and no pages row records it. A page rendered from `STEM.md`
  with `antenna page` is kept; a `STEM.html` that is not antenna's is kept with
  a warning. Links and bookmarks to `STEM.html` now fail; the RSS and OPML
  addresses do not move.
- **The link rewriter** maps `STEM_items.md` to `STEM_items.html`. A link to
  `STEM.md` of an aggregate collection is left as Markdown with the warning for
  a file antenna does not render. `validate` and `migrate-links` follow it.
  `antenna help generate`, `configuration`, `markdown-web` and `sitemap` say so,
  and `antenna help generate` now has an EXIT STATUS section.

### Fixed: the sitemap

- **The sitemap lists every page, post and items page written as HTML**:
  `STEM_items.html` of each aggregate collection and `STEM.html` of a
  page-index collection (once generated), and every HTML output of a page (a
  page that writes `a.html` and `b.html` lists both). A pages row with no HTML
  output (`antenna page c.txt c.md`) was listed as the site root and no longer
  is.

### Changed (DR-0035)

- **Harvest no longer reads `robots.txt`.** antenna fetches feeds that sites
  publish for subscription, so it does not fetch a host's `robots.txt`, skips
  no feed because of it and ignores `Crawl-delay`. The dependency
  `github.com/temoto/robotstxt` is removed. The pause between requests to one
  host (500 ms, `harvest: pause:`) and the User-Agent stay. This changes
  behavior: feeds that a `robots.txt` used to forbid (for example
  `news.un.org`, whose `Disallow: */news/` matched its own feed addresses) are
  now harvested. A site that wants antenna gone has to block its User-Agent.

## v0.1.2 — 2026-11-23

### Added

- **`antenna prune [-w] [-all]`** removes the pages whose source file is gone,
  with the output files their records list. A dry run unless `-w` is given
  (exit 1 when there is something to prune). It keeps an output another page
  still writes or one outside htdocs, and refuses to prune every page unless
  `-all` is given (DR-0028).
- **`markdownItems`** per collection in `antenna.yaml` switches a collection's
  `STEM_items.md` on or off; the pages collection gets none by default. A
  `.md` link into a `subsites` path becomes `.html` in the HTML (DR-0027).
- **`html: false` on a collection makes it Markdown-only** (DR-0031). `generate`
  writes the collection's items Markdown, RSS and feed-list OPML and no HTML,
  neither its page nor its posts, and removes a collection page left from before;
  post HTML from before is left and counted in a warning. `antenna post` into it
  writes no HTML and stores the post's `.md` address as its link (earlier posts
  keep theirs); the sitemap lists such posts by that address; a link to them stays
  a `.md` link; `validate` warns when the collection names a generator
  (`markdown-only-generator`). `html: false` with `markdownItems: false` or
  `mode: page-index` is a configuration error (78).
- **Harvest finds a site's feed** (DR-0030). A link in a collection may be a
  site address as well as a feed: harvest tries the address, an `rss` outline
  in the site's `toc.opml`, the page's `<link rel="alternate">`, and up to
  eight common paths (`feed`, `feed/`, `rss`, `rss.xml`, `feed.xml`,
  `atom.xml`, `index.xml`, `feeds/posts/default`), remembers the answer in the
  new `channels.discovered_link` column (added to an existing database on the
  next harvest, rows kept) and uses it first next time. A Markdown site with
  only a table of contents of pages and no feed is reported for that link; harvest
  never fetches the pages a table of contents lists.
- **Harvest obeys `robots.txt`** (new dependency `github.com/temoto/robotstxt`).
  Each host's file is read once per run and applies to every request,
  including the guesses. A forbidden URL is skipped and reported, not failed;
  a missing file allows the site and one with a mistake in it is read as far as
  it can be; one that answers 5xx or cannot be reached is asked for twice more,
  then skips the host for the run and fails the link as unavailable (69); a
  `Crawl-delay` is honored up to 30 s. Requests to a host are one at a time
  with a pause (500 ms), set with `harvest: pause:` in `antenna.yaml` (a bad
  value is exit 78).

### Changed (Textcasting alignment, DR-0029)

- **An item with no title has no heading.** The HTML page no longer shows a
  placeholder `@Source` heading (the `<article>` gets an `aria-label` from the
  source and date) and `_items.md` has no `## [@Source](link)` line; the item
  starts at a thematic break and its link is on a line of its own.
- **Links, bold and italic reach every item.** A feed item with an HTML
  description and no stored Markdown is converted to Markdown when it is
  rendered, so its links and emphasis are kept. Nothing is written to the
  database, and `items.html: escape` and `unsafe` keep their meaning.
- **RSS passes more through.** A name-only author is written as
  `<dc:creator>`, and `<guid>` says `isPermaLink="false"` (its text is
  unchanged, so no reader sees a new identity; closes #23). Every feed is
  rewritten once.

### Fixed

- A post made by antenna now always has a guid: the one in its front matter, else
  its link (unique, and a URL), written as `<guid isPermaLink="true">`. Before,
  a post with none had no `<guid>` in the RSS and the code that meant to default
  it ran before the link existed. A post stored earlier with no guid gets its
  link at output time, so a reader that tells items apart by link sees the same
  identity. Harvested items are untouched.
- Harvest follows redirects. `webget` stopped at every one (its redirect
  handler always refused), so a feed that moved from http to https or from
  `/feed` to `/feed/` failed with `http error: 301`.
- `antenna generate` writes the RSS feed and OPML from the readable rows when an
  item row cannot be read, and exits 65 with the counts, as the HTML page and
  the items Markdown already did. Before, the feed was not written at all.
- A harvested post whose feed carries its whole source (with YAML front matter)
  in `source:markdown` no longer shows the front matter in the item text of the
  collection's HTML page or its `_items.md`. The aggregated feed still carries
  the source as harvested.

### Removed

- **ODT support** (DR-0032). antenna no longer reads ODT or OTT files: not as a
  collection, a page source, a post or `blogit` input. Such a file is refused
  with a message to save it as Markdown (exit 65, or 2 for `antenna page`).
  `antenna stylefrom` is unchanged: its ODT input belongs to the separate
  `stylefrom` package.

### Deprecated

- **The `pages.outputPath` column will stop being written** in a later release.
  Use `outputs` (field 4 of `antenna pages`), which lists every file a page
  writes. Field 2 of `antenna pages` will then be filled from `outputs`, so the
  listing does not change. Scripts that read the column from `pages.db`
  directly should switch to `outputs` now (DR-0024 10).

## v0.1.1 — 2026-11-20

Page Markdown, previous and next links, and validation grow from the first
Markdown-web release. Everything here is from the same day-to-day use of
`antenna` on a BonBon site (decisions DR-0019 to DR-0026).

### Added

- **`antenna page SOURCE [OUTPUT...]` writes `.html` and `.md` by extension.**
  The `.md` output is the expanded Markdown (macros and includes resolved),
  the copy a Markdown reader gets (DR-0019).
- **Pages declare their Markdown alternate.** A page made from a text source
  names the Markdown copy beside it in its head; `validate` warns about a
  static Markdown alternate left in the generator YAML (DR-0023).
- **The `pages` table records what a page wrote and includes.** New
  `outputs` and `includes` columns (JSON), upgraded in place. `antenna pages`
  shows six fields and takes `-stale` to list pages whose output is older than
  their source or includes (exit 1 when there are any, 0 when none). Includes may nest: a warning past 3 levels, exit 65
  past 5 or for a cycle. `generate` rewrites every recorded output (DR-0024).
- **Sequences in `antenna.yaml`.** Name the directories or list files whose
  members are a sequence and each member gets previous and next links, in the
  HTML and in the Markdown copy (DR-0026).
- **`antenna.yaml` `author` and `descriptionMax`.** The site author is the
  default author of page Markdown and of your own posts (DR-0021).
- **New `validate` findings.** Info: `include-directive`,
  `no-description`, `no-author`, `title-from-heading`, `description-long`
  (DR-0020), `output-stale` (DR-0024), `sequence-empty` and `sequence-overlap`
  (DR-0026). Warning: `row-output-missing`, `row-duplicate-output`,
  `include-missing` (DR-0024) and `sequence-path-missing` (DR-0026).
  `validate -color WHEN` (`auto`, `always` or `never`) colors the report (`NO_COLOR` is honoured under `auto`; JSON is never colored).
- The documentation site is published for Markdown readers too: `index.md`,
  `.well-known/toc.opml` and a Markdown alternate on every page.

### Changed

- **Previous and next links come from the table of contents for posts only**;
  pages get none unless their front matter names them or they are in a
  sequence (DR-0025). A page is listed by the Markdown it was made from (for
  example `bio.md` rendered to `about.html`), and the neighbours use the same
  name.
- **Front matter is only the block at the top of a file.** A `---` block later
  in a document is text (DR-0022).
- The minimum Go is 1.27.4.

### Fixed

- `validate` no longer calls an e-mail autolink a missing file.
- A directive's language no longer swallows the first word of the next
  paragraph, and a code block keeps its blank line before it.

## v0.1.0 — 2026-11-20

A larger step than the v0.0.x releases: antenna sites now publish for the
Markdown web (Dave Winer's BonBon browser and readers like it) from the same
Markdown they publish as HTML, and two new commands check a site and bring
older sites up to date. Because the change is this large the minor version
moves from 0.0 to 0.1.

### Added

- **Publishing for Markdown readers.** Write links between your pages to the
  `.md` file. In the HTML output a link to a page, collection or post that
  antenna renders becomes the `.html` page, and a link to a file it does not
  render is left as written with a warning. `antenna help markdown-web` is the
  overview.
- **Collections as Markdown.** A collection with items also gets
  `STEM_items.md` beside its HTML page, with front matter saying what it is, and
  the HTML page declares it in its head as its Markdown alternate.
- **A table of contents for readers.** `antenna generate` writes
  `.well-known/toc.opml`: the site's pages, the collections' items pages and
  the posts (oldest first, by year and month). Pages and posts whose front
  matter names no `urlPrev` or `urlNext` get previous and next links from it.
- **`subsites` in `antenna.yaml`.** The other Markdown sites on the same host
  (for example `/opml` and `/cmarkbrowser`) become `include` outlines in the
  table of contents, which is how BonBon, asking only the host root, finds
  them. Nothing is fetched.
- **`bonbon` in `antenna.yaml`.** The content of `.well-known/bonbon.json`.
  With no setting no file is written and none is removed.
- **BonBon front matter.** `link`, `urlPrev`, `urlNext`, `rightMarginImage`
  (with `rightMarginImageAlt` for the HTML) and `pubDate` in the RSS 2.0 form
  are understood, and `{{page.NAME}}` macros are expanded in pages and posts.
- **`antenna validate [-json] [-online] [PATH]`.** A read-only check of a file,
  a directory or the whole site, with stable finding codes: front matter, dates,
  titles, links to `.html` pages, links to missing or unrendered files, raw HTML
  and SVG, image types, missing macro keys, a missing `base_url`, stale page
  rows, Markdown not published, an out-of-date table of contents, a bad
  `bonbon.json`, a missing `.nojekyll` on a GitHub Pages site, and with
  `-online` a sub-site that does not answer. Plain or JSON output; exit 0 when
  clean, 1 when there are errors or warnings.
- **`antenna migrate-links [-w] [PATH]`.** Rewrites links to `.html` pages that
  antenna renders from Markdown into `.md` links in your Markdown. A dry run
  showing a diff unless `-w` is given; only the link destinations change.
- **Help topics** `markdown-web`, `validate` and `migrate-links`, with manual
  pages and entries in the completion scripts and the interactive menu.
- `antenna preview` serves `.md` and `.markdown` as `text/markdown`, serves
  `.well-known`, and sends Markdown to a client that prefers it in `Accept`.

### Changed

- **`antenna blogit` makes the post publishable.** It adds `pubDate` (the RSS
  form, with an offset, in the time zone of the computer; set `TZ` to post from
  another zone) and a `title` from the first heading when the file has none. It
  never changes a value that is there. Before this a blogit post had no date
  and stayed a draft.
- **A date with a zone keeps it.** An RSS `pubDate` such as
  `Mon, 30 Nov 2026 21:30:00 -0800`, and an unquoted timestamp with an offset,
  are stored with the offset, so a post made late in the evening stays on its
  own day and month in the table of contents, the feeds and the page. Dates in
  GMT, with no zone, or quoted are stored as before.
- **Generated files are deterministic.** A file is rewritten only when its
  content changes, so regenerating an unchanged site touches no file. The first
  time you regenerate a site made with an earlier version, expect one round of
  differences (attribute and meta order, `.md` links written as `.html`,
  previous and next links, the Markdown alternate links).
- The minimum Go is 1.27.3, and antenna requires the `opml` module 0.0.12.
- The manual pages (`*.7.md`) are regenerated from `antenna help TOPIC` by the
  Makefile, so every topic's page is current.

### Fixed

- **`antenna unpage` exits 1** when no page record matches the path (it exited
  0). Scripts that relied on 0 for a stale path need to expect 1.
- One unreadable item row no longer stops a collection's HTML page and items
  Markdown from being written; the row is reported and the command exits with
  its class.
- A `{{page.title}}` inside a link you wrote no longer produces a link inside a
  link.
- An items file left by a collection that became a page index is removed.

## v0.0.28 and earlier

See the release notes of each release on
[GitHub](https://github.com/rsdoiel/antennaApp/releases).
