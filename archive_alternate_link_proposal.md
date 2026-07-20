# Proposal: Per-Collection Archiving to archive.org with an Alternate Reading Link

Status: Proposal / Draft
Author: (your name)
Related project: [antennaApp](https://github.com/rsdoiel/antennaApp)
Related files: `antenna.yaml` (new per-collection flag); new `archive.go`
(new `antenna archive` action); `sql_stmts.go`/`schema.go` (new `items`
column); `html.go` (`WriteItem`, new alternate link); `antenna.go`/
`help_dispatch.go` (new CLI verb).

## 1. Problem Statement

Some harvested feeds link to websites that are unpleasant or worth avoiding
on their merits, even though the feed content itself is worth reading. The
motivating case: Yap State Government's blog (`yapstate.gov.fm`, a Wix
site) publishes a genuinely useful RSS feed for the Pacific collection, but
visiting the live site trips Wix's stock analytics/app-install-nag/
notification-prompt bundle. Given the politics involved, asking a Pacific
island government to change its web vendor's default behavior isn't
appropriate — the right move is technical, not diplomatic: read the content
without loading the live site at all.

`archive.org`'s Wayback Machine already solves "read this without touching
the origin server." What's missing is automation: a way to (a) proactively
submit an item's URL for archiving, scoped to specific flagged feeds rather
than every item Antenna harvests, and (b) surface the resulting archived
copy as a second, optional link on the generated aggregation page, so a
reader can choose it instead of "Continue reading."

This is deliberately **not** a general-purpose archiving tool for the whole
site. Most feeds don't need this — it's for the small number of channels
where you specifically want an alternate, script-free reading path.

## 2. Design Goals

1. **Opt-in per collection**, not global. A new boolean flag on a
   collection in `antenna.yaml`, following the same shape as the existing
   per-collection `filters:` list — collections that don't set it behave
   exactly as they do today.
2. **A separate pipeline step**, not folded into `harvest` or `generate`.
   A live Wayback Machine capture takes anywhere from a few seconds to over
   a minute per URL; running it inline during `generate` (currently a fast,
   local, offline operation) would make `generate` slow and network-
   dependent for no benefit to collections that don't use this feature.
   Matches your own framing: "This could happen between the harvest run and
   the generate run."
3. **Idempotent.** Re-running the archive step must not resubmit URLs that
   already have a stored archive link. Both because it's pointless (the
   Wayback Machine already has a copy) and because archive.org's Save Page
   Now API is rate-limited — resubmitting unchanged items on every run
   would burn that budget for nothing.
4. **Graceful degradation.** If a submission fails (rate limit, the target
   site blocks the crawler, a network error), the item simply has no
   archive link and `generate` renders it exactly as it does today — no
   hard failure of the whole run over one bad URL.
5. **No secrets in `antenna.yaml`.** The Wayback Machine's Save Page Now
   API requires an archive.org account and an S3-style access key/secret
   pair from account settings. That credential is a secret and must not
   live in a file that gets committed to a site's git repo — an environment
   variable (or a separate, gitignored credentials file) is required,
   consistent with this workspace's existing "never commit `.env` files"
   convention.
6. **Reuse the existing link-resolution shape**, not a new templating
   mechanism. The `items.link` config added by the item-formatting feature
   (`item_formatting_proposal.md` §5.2, implemented in `resolveItemLink`)
   already models "an optional link with a fallback." The archive link is
   a second, independent link alongside it, not a replacement.

## 3. Non-Goals

- **Not a general "archive everything Antenna harvests" tool.** Only items
  belonging to collections that opt in are ever submitted.
- **Not a replacement for `Continue reading`.** The archive link is
  additive — an alternate, not a substitute. The original link/label
  behavior from the item-formatting feature (§5.2 of that proposal) is
  unchanged.
- **Not a full Wayback Machine client.** No browsing of historical
  snapshots, no snapshot-selection UI, no diffing between captures. One
  capture is requested per item; the resulting single URL is stored and
  shown.
- **Does not change `harvest.go`'s feed-fetching behavior.** This is a
  separate step that reads the same `items` table `harvest` already
  populates; it does not touch feed parsing or `saveItem`.
- **Does not attempt to solve credential storage generally** (e.g. no new
  secrets-management subsystem). Just enough to keep one API credential out
  of version control.

## 4. Current State (confirmed by reading the code)

- **`Collection` struct** (`schema.go:230-278`) already has a `Filters
  []string` field populated from `antenna.yaml`'s per-collection `filters:`
  list, and a `DbName` field naming that collection's SQLite database. A
  new `Archive bool` field (or similarly named) fits directly alongside
  these — same struct, same YAML-tag pattern
  (`` `yaml:"archive,omitempty"` ``).
- **CLI dispatch** (`antenna.go:83-86`) is a flat `switch action` on the
  first CLI argument: `"harvest", "fetch"` → `app.Harvest(...)`;
  `"generate", "build"` → `app.Generate(...)`. A new `"archive"` case
  calling a new `app.Archive(out, eout, cfgName, args)` follows the exact
  same shape. `help_dispatch.go:108-110` has a matching switch for help
  text that would need the same new case.
- **`Harvest`'s top-level shape** (`harvest.go:36-56`) is the direct model
  for `Archive`'s top-level shape: load `AppConfig`, default `args` to
  every collection's file if none given, resolve each named arg to a
  `*Collection` via `cfg.GetCollection`, and call a per-collection method
  — warning to `eout` on a single collection's failure rather than aborting
  the whole run. `Archive` would do the same, additionally skipping any
  collection where `col.Archive` is false (or unset).
- **`items` table schema** (`sql_stmts.go:45-61`): `link` (primary key),
  `postPath`, `title`, `description`, `authors`, `enclosures`, `guid`,
  `pubDate`, `dcExt`, `channel`, `sourceMarkdown`, `status`, `label`,
  `updated`, `categories`. No column exists today for a second, alternate
  URL. `SQLUpdateItem` (`sql_stmts.go:96-109`) is an upsert
  (`ON CONFLICT (link) DO UPDATE SET ...`) run by `saveItem` during
  harvest — a separate `UPDATE items SET archiveLink = ? WHERE link = ?`
  from the new archive step does not conflict with it, since harvest and
  archive run as distinct steps against the same row.
- **`WriteItem`'s link handling** (`html.go`, `resolveItemLink`,
  confirmed present from the item-formatting feature): resolves one
  `LinkResolution{Href, Label, AsPlainText, Omit}` per item and renders it
  in the `<footer>` as either `<a href="...">label</a>` or plain text. A
  second, independent link (the archive copy) is additive to this — it
  does not replace `linkRes`, it renders alongside it only when the item's
  `archiveLink` column is non-empty.
- **No archive.org integration exists anywhere in this codebase today** —
  confirmed by `grep -rn "archive.org\|web.archive\|ArchiveKey" *.go`
  returning nothing. This is new surface area, not a modification of
  existing behavior.

## 5. Proposed Design

### 5.1 `antenna.yaml`: per-collection opt-in

```yaml
collections:
  - title: Items of interest to the Pacific community and the Pacific RIM
    link: pacific.xml
    file: pacific.md
    generator: front_page.yaml
    filters:
      - UPDATE items SET status = 'review';
      - UPDATE items SET status = 'published' WHERE pubDate >= date('now', '-21 days');
    dbName: pacific.db
    archive: true
```

`archive` defaults to `false`/absent. Only collections that set it are ever
touched by the new `antenna archive` action.

### 5.2 New `items` column: `archiveLink`

```sql
ALTER TABLE items ADD COLUMN archiveLink TEXT DEFAULT '';
```

Empty string means "not yet archived" (or "archiving was attempted and
failed" — see §5.4 on whether to distinguish these). Non-empty means a
Wayback Machine snapshot URL, ready to render.

### 5.3 New CLI action: `antenna archive [COLLECTION_NAME]`

Mirrors `antenna generate [COLLECTION_NAME]` (`antenna-generate.7.md`):
optional collection name restricts to one collection; omitted, it processes
every collection — but, unlike `generate`, silently skips any collection
where `archive` is not `true`, rather than erroring. Suggested top-level
shape, directly modeled on `Harvest` (`harvest.go:36-56`):

```go
func (app AntennaApp) Archive(out io.Writer, eout io.Writer, cfgName string, args []string) error {
    cfg := &AppConfig{}
    if err := cfg.LoadConfig(cfgName); err != nil {
        return err
    }
    if len(args) == 0 {
        for _, col := range cfg.Collections {
            args = append(args, col.File)
        }
    }
    for _, cName := range args {
        col, err := cfg.GetCollection(cName)
        if err != nil {
            return err
        }
        if !col.Archive {
            continue
        }
        if err := col.ArchiveItems(out, eout); err != nil {
            fmt.Fprintf(eout, "warning %s: %s\n", col.File, err)
        }
    }
    return nil
}
```

`(*Collection).ArchiveItems` would:

1. Open `col.DbName`.
2. `SELECT link FROM items WHERE archiveLink = ''` (or `IS NULL`) — only
   items never successfully archived (§3, goal 3: idempotent).
3. For each link, call the Wayback Machine Save Page Now API (§5.4).
4. On success, `UPDATE items SET archiveLink = ? WHERE link = ?`.
5. On failure, log a warning to `eout` and move to the next item — one
   item's failure does not abort the collection (§2, goal 4).
6. Throttle between requests (a fixed delay, e.g. 1-2 seconds) — politeness
   toward archive.org's infrastructure and a simple way to stay under rate
   limits without building real backoff/retry logic for a first version.

### 5.4 archive.org integration: Save Page Now (SPN2)

Two viable endpoints:

- **Legacy `GET/POST https://web.archive.org/save/<url>`** — simplest:
  synchronous, no account required for light use, but unauthenticated
  requests are aggressively rate-limited and the response doesn't cleanly
  hand back the final snapshot URL — it's a redirect chain intended for a
  browser, not an API consumer.
- **SPN2 `POST https://web.archive.org/save` with API credentials** — the
  documented API. Requires an archive.org account and an access key/secret
  pair (Account Settings → S3 API Keys). Submitting a URL returns a job
  ID; polling `GET https://web.archive.org/save/status/<job_id>` until the
  job completes returns the captured timestamp, from which the canonical
  snapshot URL is constructed:
  `https://web.archive.org/web/<timestamp>/<original_url>`.

**Recommendation: SPN2 with polling**, given "select websites" implies low
volume (a handful of items per Pacific-collection harvest run, not
thousands) — the added complexity of polling a job to completion is small
and worth it for a real snapshot URL and clearer error reporting, versus
the legacy endpoint's rate limits and non-API-friendly response shape.

Credential handling (§2, goal 5): read the access key/secret from
environment variables (e.g. `ANTENNA_ARCHIVE_ACCESS_KEY` /
`ANTENNA_ARCHIVE_SECRET_KEY`) at `Archive` startup; error out early with a
clear message if any collection has `archive: true` but the environment
variables are unset, rather than silently skipping.

### 5.5 Rendering: a second, optional link in `WriteItem`

`WriteItem`'s existing footer construction (`html.go`, following the
item-formatting feature's `resolveItemLink`/`LinkResolution`) renders the
canonical link. Add a second, independent element when `archiveLink` is
non-empty:

```html
<footer>
  <a href="https://inkdroid.org/2026/07/19/bookmarks/">Continue reading</a>
  <a href="https://web.archive.org/web/20260719.../https://inkdroid.org/2026/07/19/bookmarks/">Archived copy</a>
  <p class="source">via Ed Summers blog, Inkdroid</p>
</footer>
```

This requires threading `archiveLink` through the same call chain that
already carries `link`, `sourceMarkdown`, etc. from the DB query into
`WriteItem` (`generator.go`'s `WriteHTML`/collection query, and
`WriteItem`'s parameter list) — additive, not a change to any existing
parameter's meaning.

Whether the label is a fixed string (`"Archived copy"`) or configurable
via `items:` (`page.yaml`) the way `items.link.label_fallback` is (per
`item_formatting_proposal.md` §5.2) is an open question — see §7.

## 6. Backward Compatibility

- Collections without `archive: true`: zero behavior change. No new
  column read, no new link rendered, `antenna archive` skips them outright.
- The new `archiveLink` column defaults to `''` for every existing row —
  `WriteItem` must treat `''` identically to a column that doesn't exist
  yet, so this ships safely against databases that predate the migration.
- `antenna generate` behavior for collections **with** `archive: true` but
  where `antenna archive` has never been run: identical to today — no
  items will have a non-empty `archiveLink` yet, so no second link
  renders. The feature is inert until `antenna archive` actually runs.

## 7. Open Questions

1. Should the archive link's label be configurable per collection (like
   `items.link.label_fallback`), or is a fixed `"Archived copy"` string
   sufficient for a first version?
2. Should `archiveLink` distinguish "never attempted" from "attempted and
   failed," so a failed submission can be retried on the next `antenna
   archive` run rather than silently never being retried again? A `''`-
   means-both design (§5.2) is simpler but means a permanently-failing URL
   (e.g. one the Wayback Machine's crawler is blocked from) gets retried
   forever, burning rate-limit budget on every run for a URL that will
   never succeed.
3. Should there be a per-item or per-collection cap on how old an item can
   be and still get archived retroactively (e.g. only archive items from
   the last N days), or should `antenna archive` always sweep every
   never-archived item in the collection's full history the first time
   it's run against an existing database?
4. Is a fixed request-to-request delay (§5.3, step 6) sufficient, or does
   this need real exponential backoff on a 429 response from archive.org?
   A fixed delay is simpler and likely sufficient at "select websites"
   volume, but worth deciding deliberately rather than defaulting to it
   for no stated reason.
5. Should failed archive attempts be visible anywhere in generated output
   (e.g. a warning comment, or surfaced via `antenna list`/TUI), or is
   `eout` logging during `antenna archive` sufficient?

## 8. Suggested Next Step

Not part of this proposal's scope — noted for whenever implementation
starts: given this project's TDD convention (see `html_test.go`,
`items_config_test.go`), a first implementation slice would likely be
`archiveLink` schema + `resolveArchiveLink`-style pure function (given a
stored `archiveLink`, decide what to render) with tests, before touching
the network-calling `ArchiveItems` step at all — mirroring how
`item_formatting_proposal.md`'s implementation separated pure rendering
logic (`resolveItemContent`, `resolveItemLink`) from I/O.
