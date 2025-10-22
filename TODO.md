
# TODO

## Bugs

- [ ] The RSS I'm producing isn't always valid for aggregated items
  - [X] I need to verify RSS feeds with the source:markdown namespace is still valid, the XML render in Firefox gripes about it
- [X] The link element for RSS feeds isn't showing up in aggregated HTML or post HTML pages
- [ ] generate is not generating everything for the collection
  - [ ] re-rendering post lists correctly
  - [ ] pages are not being rendered for pages collection

## Up Next

- [ ] When I pass the collection on the command I should not require the ".md" filename
- [ ] Figure out how I want to handle a list of links to posts, e.g. recent posts and archive of posts instead of feed reading posts
- [X] init needs to create the default pages.md collection, that way Antenna will be able to manage
      collections pages and posts.
- [ ] pages always go in the pages collection, this will let me generate a sitemap by taking the pages
      and posts from the pages collection along with the items from other collections and rendering out
      the appropriate sitemap.xml file(s).
- [ ] pages should be rendered to HTML when you use the generate command without specifying collection
- [ ] When I have sitemaps debugged the sitemaps should be create/updated whenever generate action is taken
- [ ] Added PWA generation support. If Antenna app had an option of generating websites as a PWA you could use a Pi Access Point/Website to distribute the "app" and the content would be available "off line" when people are away from the access point
  - [ ] The "app" name should come from the antenna.yaml file
  - [ ] The "page" action should track which pages are added to the site, needed when regenerating the cache page list
  - [ ] The antenna.yaml should have a boolean to indicate if the site should be configured as a pwa
  - [ ] There needs to be a command line easy way to set things up or turn them off for pwa support
  - [ ] If a page list is included these could be automatically regenerated from the "generate" action.
- [ ] Explore a "reply" action, this would take a link or guide, find the markdown translation in a feed, then pop it into an editor as a a quoted Markdown content. The reply link should be tracked some how and displayed in relation to the item in the aggregated feed.  Enough metadata for threading will need to be tracked. Look at prior art to see what is easy to integrate without recreating ActivityPub or AT Proto
- [X] cmarkdoc.go should support @include-text-block and @include-code-block like I implemented in my commonMarkDoc processor, this will let me remove remaining Pandoc requirements from building my website
- [X] The Markdown document defining the feed should get rendered as a standard OPML file along side the HTML and RSS aggregated feed. This could then be linked shared with other people or systems
- [X] Double check the ordering of my head element children, make sure the meta element for character encoding comes before title

## Thinking about

- [X] Consider a page action, it would make sure that the metadata is valid like post would be just render the HTML page next to the Markdown document, this could minimize the need to rely on Pandoc
  - [X] seems weird to reference a collection that will never hold anything but I need to find a YAML expression to build the page
  - [X] the YAML expression should allow full customization (or leaving out) the section, head, header, footer, etc.
  - [ ] The default rendering of the aggregation page doesn't make sense for an ad-hoc HTML page, I need a clean approach that will work for both
- [ ] Documenting simple to complex integrations
  - [X] A simple website of "pages"
  - [ ] A blog example using a collection to host posts
  - [ ] A link blog with embedded responses (do I need a different action than post?)
  - [ ] An example should show how to integrated with with Pandoc, PageFind and FlatLake
- [ ] Should antenna init also generate some sample CSS and JavaScript modules?
  - [ ] Should it setup for integrations with Pandoc, PageFind or FlatLake?
- [ ] To make Antenna app more interesting  I need to include some sample themes other people would use. 
  - I could look over the websites for feeds followed in the Antenna website and see if I could recreate similar styles
  - I could look at the WP theme gallery for ideas and see what could be implemented
- [ ] In the front matter defining a collection, should the link element require full URL or a partial one?
  - full is simpler in rendering implementation but makes previewing and proof reading tricky
  - relative makes previewing and proof reading much easier as the preview becomes functions like published sites
  - the antenna YAML could make a distriction between preview URL (localhost) and published URL (staged for production)
- [ ] Should postPath point at Markdown document or HTML?  The link element is pointing at HTML any
- [ ] Should I allow relative link elements in the post's front matter? this would limit typos of long urls, might handle the preview verus publication issues around full URL versus relative paths.
- [ ] Shouldn't generate handle all HTML/RSS 2.0 rendering?  The post action renders the HTML in the target path, this feels right in the sense it is immediately visible to proof read but when the YAML page definitions are updated you have to "post" the items again to pickup the changes, so this doesn't make sense as the post wasn't really updated in terms of content.
  - As the items are processed, items that have `postPath` set could queued be re-rendered as HTML
- [ ] Think about generator.go and what they are generating, might make sense to split out the HTML and RSS 2.0 generation into separate files and have the wrapping functions or interfaces defined in generator.go
- [ ] Improve default YAML for rendering colllections
  - Nav could be autogenerated for the defined collections in the antenna.yaml
  - header and footer elements could be formed such that customization is easily seen
  - The head element needs work, I could include an example of a style element with sensible generic CSS for correctly sizing H elements, this could go after the automated stuff like setting character encoding correctly before the title element 
- [ ] It'd be nice to have full enclusure support so podcasting using antenna's post worked seemlessly. Need to look at existing Go packages to see what has already be implemented or what might suggest the right path forward
  - [github.com/eduncan911/podcast](https://github.com/eduncan911/podcast)
- [ ] Evaluate how to "post" to specific platforms, e.g. BlueSky and Mastodon since they do not handle inbound RSS yet
  - See [github.com/bitesinbyte/ferret](https://github.com/bitesinbyte/ferret) as an example
- [ ] Do I need a publish action that would present the website using the published base URL?

## Someday, Maybe

- [ ] A Web GUI for managing collections, feeds and items. Could work like preview action, `antenna manage`
