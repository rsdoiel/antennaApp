
# TODO

Ideas, not quite a roadmap

## Bugs

- [ ] The RSS I'm producing isn't always valid for aggregated items
  - Look at converting item content into Markdown and using that to build the feed item
  - The problem is happening because the RSS items consumed aren't always strictly valid (often will enclude undefined embedded entities)

## Up Next

- [X] Improve metadata in head of HTML pages by incorporating the front matter of the collections or blog posts
- [X] Improve HTML accessibility and ARIA support in generated pages
  - [X] Add per-page `<title>` support driven by front matter title
  - [X] Skip navigation link (WCAG 2.4.1)
  - [X] Fix `<address>` misuse in feed item cards → `<footer>`
  - [X] `<time>` element for dates in feed item cards
  - [X] Configurable `lang` attribute on `<html>`
  - [X] Warn when aggregate page has no `<h1>`
- [X] list posts and items (`antenna posts`, `antenna items`)
- [X] list pages (`antenna pages`)
- [X] list themes (`antenna themes`)
- [X] generate default theme structure (`antenna themes new [NAME]`)
- [X] apply theme (`antenna apply THEME_PATH [YAML_FILE_PATH]`)
- [X] Consider a page action, it would make sure that the metadata is valid like post would be just render the HTML page next to the Markdown document, this could minimize the need to rely on Pandoc
  - [X] seems weird to reference a collection that will never hold anything but I need to find a YAML expression to build the page
  - [X] the YAML expression should allow full customization (or leaving out) the section, head, header, footer, etc.
  - [ ] The default rendering of the aggregation page doesn't make sense for an ad-hoc HTML page, I need a clean approach that will work for both
- [X] generate a configurable, sensible default CSS for working with Markdown content (`antenna css`)
- [ ] Documenting simple to complex integrations
  - [X] A simple website of "pages"
  - [X] A blog example using a collection to host posts
  - [X] A link blog with embedded responses — `post` works fine; `quote` generates the blockquote stub
