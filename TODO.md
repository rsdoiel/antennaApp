
# TODO

Ideas, not quite a roadmap


## Release 0.0.28 blockers


## Bugs

- [x] `antenna unpage` exits 1 when no page record matches the path (was 0; DR-0003). Test `TestRun_UnpageNoMatchIsNegative`. (2026-11-20)

- [x] The skip link class is `skip-link` (code, CSS, tests and DR-0002 all agree); no rename needed. (2026-10-05)
- [x] On aggregated pages the skip link jumped to the footer. `<main id="main-content">` is now `tabindex="-1"` so focus moves into it; checked on the live-site copy. Fixed in cc28293. (2026-10-05)

## Up Next

- [ ] A way to remove a page completely. `unpage` only deletes the row in `pages.db`; the generated HTML and the published Markdown stay on disk, so a removed page lingers in the site, the sitemap and any hand-run `rm` is the only cleanup. Options: `unpage --files` (delete the output files too), or a `prune` verb that drops rows whose input file is gone. Found 2026-11-20 on the test blog, where a stale `index.md` row made `generate` exit 66. Also list `unpage` where authors look for it (`pages` help should point to it).

### BonBon and the Markdown web (built on branch `bonbon`, not yet merged)

BonBon (https://github.com/scripting/bonBon/) is Dave Winer's browser whose native format is Markdown. Design brief, plan and records are in agents/projects/antennaApp/ (DR-0007 to DR-0014). User documentation: `antenna help markdown-web`.

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
- [ ] `antenna generate` with no `base_url` warns and leaves an existing `.well-known/toc.opml` in place (a hand-written one would otherwise be deleted). Decide whether a stale generated one should go; DR-0013 only removes it when there is nothing to list. (code review 2026-11-20)
- [ ] RSS generation stops at the first unreadable item row (also true on main); HTML and the items Markdown now carry on and report the row at the end.
- [x] `blogit` adds `pubDate` (RSS form, local numeric offset) and a `title` from the first heading to the post's copy, so a blogit post is published and in the toc (DR-0016, built 2026-11-20).
- [x] `antenna validate [-json] [-online] [PATH]` built (DR-0017, 2026-11-20): 14 file checks, 7 site checks, `-online` for sub-sites; plain and JSON output, exit 0/1/2/66/78. Checked on the real blog repo: 306 warnings, 281 of them `link-html` (the migration backlog), 7 genuine `link-missing`, 16 `raw-html`, 2 `no-title`. Left: a way to silence a code (the codes are stable), `validate` of `_items.md` links, an ODT source reported as info.
- [ ] `antenna migrate-links [PATH]`: `.html` links to `.md` in the author's Markdown, dry run by default, `-w` writes (DR-0017, DR-0007 4). Build after `validate`.
- [x] A date with a zone keeps it when stored (RFC 3339 with the offset), so a post made at 21:30 Pacific stays on its day in `toc.opml`, the feeds and the page; `blogit` uses the process time zone, `TZ` overrides it (DR-0018, 2026-11-20).
- [ ] "Markdown-clean" check: warn when a post carries raw HTML, SVG images or other things BonBon strips (it allows PNG, JPEG, GIF and WebP only, and no page CSS or JavaScript).
- [ ] Harvest from BonBon sites: accept a site's `toc.opml` as a feed collection source.
- [ ] Open question: a Markdown-only collection mode with no theme (only `.md` and `toc.opml`), since the reader picks the look. This changes what a collection is, so write a decision record first.
- [ ] Textcasting alignment (https://textcasting.org/): audit what antenna passes through unchanged.
  - [ ] Titles optional on any post; no generated placeholder title.
  - [ ] Enclosures kept on inbound and outbound items (see #22, RSS enclosure escaping).
  - [ ] No length limit on posts; summaries may truncate, the full text stays.
  - [ ] Edits to a post reach the feed and the generated pages (changed items are republished, not left stale).
  - [ ] Pass through feed elements antenna does not itself use.
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
