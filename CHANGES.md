# Changes

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
