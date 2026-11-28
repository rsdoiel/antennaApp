
# TODO

Ideas, not quite a roadmap


## Release 0.1.0 blockers


## Bugs

- [ ] Writing a pages.html file doesn't make sense if there is no content. This is appearing all the time now in the antenna website.
- [x] Built 2026-11-25 (antennaApp DR-0037, uncommitted); checked on a copy of the antenna site: root-relative `href`/`src` in the four `_items.html` files 0 after, stored items unchanged. Original report: root-relative links and images in harvested item bodies are kept as written, so they resolve against the antenna site, not the site the item came from, and break. Found 2026-11-23 on the antenna site (v0.2.0): `columns_items.html` has 94 such links (`/images/bluecore-editor.png`, `/2026/08/02/games-at-the-press-of-a-button-the-rip-o-bot/`), `pacific_items.html` 21 (`/ascc-career-fair`, `/behavioral-health`), `retro_computing_items.html` 12, `writing_items.html` 2. The same holds in `*_items.md` and, if the feed carries the HTML, in `STEM.xml`. Wanted: when rendering an item, resolve a relative `href` or `src` (root-relative, and page-relative) against the item's `link` (or the feed's `xml:base` when it has one), in the HTML and the items Markdown; leave absolute, `mailto:`, `#fragment` and `data:` URLs alone. Decide whether the stored item keeps the original text and only the rendering rewrites it (as the `source:markdown` front matter fix does). Needs a record first, since it changes what readers see for every collection. Reproduce through `harvest` and `generate` with a feed whose item body has `<img src="/images/x.png">` and a link `/a/b`, not by calling the renderer alone.

- [x] Front matter of a harvested post's `source:markdown` showed in the item text of `columns.html` and `columns_items.md`; stripped at render time (2d58c51, 2026-11-23). `columns.xml` keeps the source as harvested.
- [x] `antenna unpage` exits 1 when no page record matches the path (was 0; DR-0003). Test `TestRun_UnpageNoMatchIsNegative`. (2026-11-20)

- [x] The skip link class is `skip-link` (code, CSS, tests and DR-0002 all agree); no rename needed. (2026-10-05)
- [x] On aggregated pages the skip link jumped to the footer. `<main id="main-content">` is now `tabindex="-1"` so focus moves into it; checked on the live-site copy. Fixed in cc28293. (2026-10-05)

## Up Next

- [ ] Write `/.well-known/cmarkfind.json` when the site publishes a cmarkfind index: a small JSON object, `{"index": "/cmarkfind/manifest.json"}`, a same-origin path (or address) of the index's `manifest.json`, beside `toc.opml` and `bonbon.json`. cmarkfind's `indexer discover` reads it and, when it is absent, tries `/cmarkfind/manifest.json` at the site root, so a site that publishes there needs no file. Nothing goes in `toc.opml` (BonBon's author moved site settings out of the head). Needs an antennaApp record when built; cmarkfind DR-0017 (proposed) has the file's rules. Not before cmarkfind's Phase 7 is built.

- [ ] Review Dave Winer's RSS Chat, https://github.com/scripting/rss.chat, see what features could be encorporated into AntennaApp to support the RSS oriented chat that Dave has talked about. This would be targeted at an eventual v0.3.x release. It can start as an experimental feature in v0.2.x.

- [x] Built 2026-11-25 (antennaApp DR-0038, uncommitted): the status table `harvest_log` and the end-of-harvest warning. Left, for a later record: a command that lists failing feeds, a flag that skips them, and a command that prunes them from the Markdown feed list (dry run by default). Original request: when we harvest, we need a table that keeps track of the feeds' HTTP status response. Then when a feed isn't available for some threshhold over a defined time period a warning at the end of the harvest can be shown for the feeds that fall into that category. There could be an automatic flag to stop trying to havest the failing feed or another command to auto-prune the feed from the related Markdown feed list.
- [x] A way to remove a page completely: `antenna prune [-w] [-all]` removes the pages whose source is gone, with their recorded outputs; dry run by default (antennaApp DR-0028, built 2026-11-23). `unpage` is unchanged. Left: interactive-mode guide for `prune`.

- [x] A cap on the items each feed shows: `max_items_per_feed` (absent or 0 is no cap; `antenna init` writes 1000) and `feed_limits` by feed address in antenna.yaml; hides, never deletes; `generate` warns when a feed is cut (antennaApp DR-0039, built 2026-11-25, uncommitted). Left: `validate` reporting a `feed_limits` address no feed list holds.
- [ ] Leaner item markup on aggregated pages. Measured on the Pacific page (5.9 MB, 4,374 items): `data-link` repeats the footer `href` (524 KB; `modules/article_collapse.js` of the antenna site reads it for the domain badge) and `data-pagefind-filter` is 716 KB (pagefind is commented out in the antenna site's `website.bash`). Needs a record; the site's scripts change with it. Wanted by the author, not yet decided.

- [ ] Deprecate the `pages.outputPath` column (antennaApp DR-0024 10), in steps. Done: `outputs` authoritative and `outputPath` a mirror (v0.1.1); help and `pages` manual say it will stop being written; the deprecation is in the Unreleased notes in CHANGES.md (copy it into codemeta.json `releaseNotes` at release prep). Left: a later release stops writing it and fills field 2 of `antenna pages` from `outputs`; dropping the column is a major-version decision.

### BonBon and the Markdown web (built on branch `bonbon`, not yet merged)

BonBon (https://github.com/scripting/bonBon/) is Dave Winer's browser whose native format is Markdown. Design brief, plan and records are in agents/projects/antennaApp/ (DR-0007 to DR-0014). User documentation: `antenna help markdown-web`.

- [ ] Does the page verb need to format/add front matter links for supporting BonBon?

- [x] Phase 1 (DR-0007): `.md` links become `.html` in HTML output, with a warning for a link to a file antenna does not render.
- [x] Collections get a Markdown rendering of their items, `STEM_items.md`, declared as the HTML page's alternate (DR-0008).
- [x] BonBon front matter: `link`, `urlPrev`, `urlNext`, `rightMarginImage`, RSS-form dates, `{{page.NAME}}` macros (DR-0009).
- [x] `antenna preview` serves `.well-known`, the Markdown and OPML types, Accept negotiation and Content-Location (DR-0010, DR-0012).
- [x] Deterministic output: a generated file is rewritten only when its content changes (DR-0011).
- [x] `antenna generate` writes `.well-known/toc.opml` (Pages, Collections, Posts oldest first); previous and next come from it (DR-0013).
- [x] `bonbon.json` is not written by default (DR-0014); an optional `bonbon` mapping in `antenna.yaml` writes it (DR-0015, built 2026-11-20).
- [x] `subsites` in `antenna.yaml`: include outlines in the site's `toc.opml` for other Markdown sites on the host (`/antenna`, `/opml`, `/cmarkbrowser`), one level, absolute http(s) addresses, no network (DR-0015, built 2026-11-20). Left: a check command that asks whether each sub-site's toc answers; finding sub-sites by scanning htdocs; sub-sites' sitemaps in the sitemap index.
- [x] `subsites` for opml and cmarkbrowser added to the root rsdoiel.github.io `antenna.yaml` (blog repo, branch bonbon, e8242624, 2026-11-20). [ ] Add `antenna` once its own site publishes `.well-known/toc.opml` (it 404s today: generate it with the BonBon build, commit, push).
- [x] The `bonbon` branch is merged to `main` (fast-forward, 2026-11-20); not released, v0.0.28 stays the last release. The blog repo's `bonbon` branch is not merged to its `main` yet.
- [ ] Before merging `bonbon` to `main`: code review of the branch; restore the blog test repo's `base_url: https://rsdoiel.github.io` and drop `host:`; the blog's toc needs an include (antenna writes none).
- [x] `antenna generate` with no `base_url` warns and leaves an existing `.well-known/toc.opml` in place; decided 2026-11-23 to keep that (a hand-written file would otherwise be deleted, and nothing can regenerate it without a `base_url`). DR-0013 still removes the file only when there is nothing to list.
- [x] RSS generation no longer stops at the first unreadable item row: `generate` writes the feed (and the OPML) from the readable rows and exits 65 with the counts, as HTML and items.md do (2026-11-23). Left: `antenna rss` / `post` (post.go) still return the error and write no feed for a bad row.
- [x] `blogit` adds `pubDate` (RSS form, local numeric offset) and a `title` from the first heading to the post's copy, so a blogit post is published and in the toc (DR-0016, built 2026-11-20).
- [x] `antenna validate [-json] [-online] [PATH]` built (DR-0017, 2026-11-20): 14 file checks, 7 site checks, `-online` for sub-sites; plain and JSON output, exit 0/1/2/66/78. Checked on the real blog repo: 306 warnings, 281 of them `link-html` (the migration backlog), 7 genuine `link-missing`, 16 `raw-html`, 2 `no-title`. Left: a way to silence a code (the codes are stable), `validate` of `_items.md` links, an ODT source reported as info.
- [x] `antenna migrate-links [-w] [PATH]` built (DR-0017 7-9, 2026-11-20): dry-run diff by default, `-w` writes; only destinations change (CRLF and front matter kept); inline, angle-bracket and reference-definition links; a collection names its items Markdown or is left and reported. Trial on a copy of the blog: 281 links in 33 files (equals `validate`'s link-html), `validate` then finds none, a second run says nothing to migrate, and regenerating gives the same HTML (one cosmetic `./series/index.html` to `series/index.html`). Not done: destinations containing spaces or unbalanced parentheses unless in angle brackets, links written as raw `<a href>`, an absolute link to the site's own address (reported by `validate`, not rewritten).
- [x] A date with a zone keeps it when stored (RFC 3339 with the offset), so a post made at 21:30 Pacific stays on its day in `toc.opml`, the feeds and the page; `blogit` uses the process time zone, `TZ` overrides it (DR-0018, 2026-11-20).
- [x] "Markdown-clean" check: already built in DR-0017 as `validate` findings `raw-html` (script, style, svg, iframe, any raw tag) and `image-type` (not PNG, JPEG, GIF or WebP); tests in validate_body_test.go. Checked 2026-11-23. Not covered: `javascript:` links, CSS in front matter.
- [x] Harvest from BonBon sites: discover a site's feed from the address listed (feed, toc `rss` outline, HTML autodiscovery, common paths), store it in `channels.discovered_link`, and respect `robots.txt` (antennaApp DR-0030, built 2026-11-23, uncommitted). A feed made from a `toc.opml`, a sitemap or a page of links, and search, are not antenna's job in this release (a later tool; cmarkfind); notes in agents/projects/antennaApp/design/harvest-for-search-note.md.
- [ ] Harvest: a site whose root `toc.opml` is only an `include` (https://dave.writersweb.org/) has no feed to find at the root; its feed is in the included folder (`/scripting/rss.xml`). Listing the folder works. Whether discovery should probe each include's folder (up to eight requests per include) is deferred to a later release (decided 2026-11-23).
- [ ] Harvest: `antenna help harvest` in interactive mode (`guideHarvest`) does not mention site addresses.
- [x] A Markdown-only collection mode: `html: false` on a collection (antennaApp DR-0031, built 2026-11-23, uncommitted). Left: a Markdown-only sub-site (DR-0027 rewrites `.md` links into a `subsites` path to `.html`, which such a sub-site does not have); a site-wide default for `html`.
- [x] Textcasting alignment (https://textcasting.org/, antennaApp DR-0029, built 2026-11-23): titles optional (no invented heading; thematic break and link line in `_items.md`), links and styling for description-only items, name-only authors as `dc:creator`, `isPermaLink="false"`. Enclosures, length and edits needed no change.
  - [ ] Pass through feed elements antenna does not itself use (an item's other extension elements are not stored). Needs a new column and a rule per namespace; count what the live feeds carry first, then a record.
  - [ ] DR-0029 consequences should say that an untitled item in `_items.md` starts at a `---` line (found when reading the real output; the record is accepted, so the author decides whether to amend it).
- [ ] Re-read https://this.how/standards/ (rules for standards-makers) when writing the next design brief.


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

- [x] ODT support removed (antennaApp DR-0032, 2026-11-23, uncommitted). Open: whether `antenna stylefrom` should stop accepting `.odt` (it uses the separate `stylefrom` package).
