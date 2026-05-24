
# TODO

Ideas, not quite a roadmap

## Bugs

- [ ] The RSS I'm producing isn't always valid for aggregated items
  - Look at converting item content into Markdown and using that to build the feed item
  - The problem is happening because the RSS items consumed aren't always strictly valid (often will enclude undefined embedded entities)

## Up Next

- [ ] Improve metadata in head of HTML pages by incorporating the front matter of the collections or blog posts
- [ ] Improve HTML accessibility and ARIA support in generated pages
  - [ ] Add per-page `<title>` support driven by collection title in `antenna.yaml`
- [ ] list posts and items
- [ ] list pages
- [ ] list themes
- [ ] generate default theme structure (new theme)
- [ ] apply theme
- [X] Consider a page action, it would make sure that the metadata is valid like post would be just render the HTML page next to the Markdown document, this could minimize the need to rely on Pandoc
  - [X] seems weird to reference a collection that will never hold anything but I need to find a YAML expression to build the page
  - [X] the YAML expression should allow full customization (or leaving out) the section, head, header, footer, etc.
  - [ ] The default rendering of the aggregation page doesn't make sense for an ad-hoc HTML page, I need a clean approach that will work for both
- [ ] Documenting simple to complex integrations
  - [X] A simple website of "pages"
  - [ ] A blog example using a collection to host posts
  - [ ] A link blog with embedded responses (do I need a different action than post?)
- [ ] It would be nice to beable to generate a configurable, sensible default CSS for working with Markdown content
