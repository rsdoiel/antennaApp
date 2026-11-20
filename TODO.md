
# TODO

Ideas, not quite a roadmap


## Release 0.0.28 blockers


## Bugs

- [ ] `antenna unpage` exits 0 when no page record matches the path (a typo, or a stale row already gone). DR-0003 says not found is exit 1; print a plain error and return the negative class. Test-first through the real `unpage` dispatch.

- [x] The skip link class is `skip-link` (code, CSS, tests and DR-0002 all agree); no rename needed. (2026-10-05)
- [x] On aggregated pages the skip link jumped to the footer. `<main id="main-content">` is now `tabindex="-1"` so focus moves into it; checked on the live-site copy. Fixed in cc28293. (2026-10-05)

## Up Next

- [ ] A way to remove a page completely. `unpage` only deletes the row in `pages.db`; the generated HTML and the published Markdown stay on disk, so a removed page lingers in the site, the sitemap and any hand-run `rm` is the only cleanup. Options: `unpage --files` (delete the output files too), or a `prune` verb that drops rows whose input file is gone. Found 2026-11-20 on the test blog, where a stale `index.md` row made `generate` exit 66. Also list `unpage` where authors look for it (`pages` help should point to it).

### BonBon and the Markdown web (proposed 2026-11-19)

BonBon (https://github.com/scripting/bonBon/) is Dave Winer's browser whose native format is Markdown. Charter: https://github.com/scripting/bonBon/blob/main/docs/charter.md; skill note: https://github.com/scripting/bonBon/blob/main/docs/bonbonSkillFromClaude.md (written for BonBon 0.6.46). Antenna already shares much of the model (Markdown with YAML front matter, OPML, RSS 2.0 with the Markdown source namespace, Textcasting as inspiration). Not yet checked against a running BonBon or the generator code. Design brief, plan and records are in agents/projects/antennaApp/ (design/bonbon-markdown-web-design.md, plans/bonbon-markdown-web-plan.md, DR-0007 and DR-0008, both proposed 2026-11-19: Markdown is the source of truth for links, `.md` links become `.html` in HTML output; collections also get a Markdown rendering of their items). Build in the plan's phase order once the records are accepted. Items are ordered smallest first; the first three only add output files.

- [x] Phase 1 (DR-0007): `.md` links to pages, collections and posts become `.html` in HTML output, with a warning for a link to a file antenna does not render (linkrewrite.go; 2026-11-19). Left: author documentation, and a migration command for existing `.html` links (plan Phase 2).
- [ ] Publish the Markdown source beside each generated HTML page and post (`post.md` next to `post.html`).
  - [ ] Check the MIME type the publishing host gives `.md`. BonBon accepts `text/markdown`, or `text/plain` when the path ends in `.md` or `.markdown`.
  - [ ] Add `<link rel="alternate" type="text/markdown" href="...">` to generated HTML heads so other readers can find the source.
  - [ ] Decide whether the published `.md` keeps the front matter (BonBon reads and strips it) and whether antenna-only fields are removed.
- [ ] Generate `/.well-known/toc.opml` per collection (and for the site).
  - [ ] Outline lines in page and post order; BonBon derives prev/next arrows from neighboring lines.
  - [ ] Use `<outline type="include" url="..."/>` to join a site's collections; a subfolder may have its own `toc.opml`.
  - [ ] Decide how this relates to the OPML antenna already writes (feed lists).
- [ ] Add the BonBon front matter fields: `link`, `urlPrev`, `urlNext`, `rightMarginImage`.
  - [ ] Fill `urlPrev` and `urlNext` from collection order when the author does not set them.
  - [ ] Check `pubDate` is written in the RSS 2.0 date form BonBon expects (`Fri, 23 Oct 2026 15:19:45 GMT`).
  - [ ] Decide what antenna does with the `{{page.title}}` style macros BonBon expands (leave alone, or expand when rendering HTML).
- [ ] Generate `/.well-known/bonbon.json` (all fields optional).
  - [ ] `urlMirrorOf`: the HTML site this Markdown site mirrors.
  - [ ] `urlSearch`: BonBon appends `?q=word` and expects a Markdown results page; serve it from `antenna preview` using SQLite.
  - [ ] `cname`: preferred host name for a site served from a GitHub repository.
  - [ ] `urlUpdateSocket`: websocket for live reload (`watch <url>` in, `update` out); later, probably preview only.
- [ ] `antenna preview`: negotiate content.
  - [ ] Read `Accept: text/markdown, text/x-opml` and send Markdown to BonBon, HTML to other browsers, from one URL.
  - [ ] Folder URLs: send a `Content-Location` header naming the file so BonBon can find the page in the table of contents.
  - [ ] Test through a real HTTP client, not only handler calls.
- [ ] "Markdown-clean" check: warn when a post carries raw HTML, SVG images or other things BonBon strips (it allows PNG, JPEG, GIF and WebP only, and no page CSS or JavaScript).
- [ ] Harvest from BonBon sites: accept a site's `toc.opml` as a feed collection source.
- [ ] Open question: a Markdown-only collection mode with no theme (only `.md`, `toc.opml`, `bonbon.json`), since the reader picks the look. This changes what a collection is, so write a decision record first.
- [ ] Textcasting alignment (https://textcasting.org/, updated 2026-10-31): audit what antenna passes through unchanged.
  - [ ] Titles optional on any post; no generated placeholder title.
  - [ ] Enclosures kept on inbound and outbound items (see #22, RSS enclosure escaping).
  - [ ] No length limit on posts; summaries may truncate, the full text stays.
  - [ ] Edits to a post reach the feed and the generated pages (changed items are republished, not left stale).
  - [ ] Pass through feed elements antenna does not itself use, so it can peer with apps that support other parts of the spec.
- [ ] Re-read https://this.how/standards/ (rules for standards-makers) when writing the design brief; it is general guidance, not BonBon specific, and its text was not available when this list was written.


## Modernization (2026-11-16)

To match the workspace standard (knowledge and opml are the models; Go 1.27.3 across repositories). Do one repository at a time; run the tests and `make -n release` after each step.

- [x] Go 1.27.3: change `go 1.26.4` in go.mod, the `Go >=` lines in codemeta.json, README.md and INSTALL.md; run `go mod tidy`, `go vet ./...` and the tests.
- [x] Makefile: add `generate` (cmt for version.go, CITATION.cff, about.md, installers); make `build` stop running cmt. Model: opml/Makefile.
- [x] Makefile: add `check-release` and make `release` run `generate` then `check-release`.
- [x] Makefile: add `check-go-version` (go.mod, codemeta.json and the Go in use agree); `check-release` calls it.
- [x] Drop the armv7l target and 32-bit Pi wording (dropped 2026-10-27).
- [x] publish.bash/publish.ps1: take knowledge's checked versions (a failed pull no longer publishes nothing silently).
- [x] release.bash: matches fountain's apart from the generated-by header.
- [x] release.ps1: take knowledge's (existing-draft check, checksums).
- [x] website.mak, page.tmpl, add-col-scope.lua, website.ps1: take knowledge's (check `NO_PUBLISH` and nav links).
