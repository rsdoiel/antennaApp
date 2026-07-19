# Design: Item Layout Control in `page.yaml`

Status: Proposal / Draft
Author: (your name)
Related project: [antennaApp](https://github.com/rsdoiel/antennaApp)
Related files: `page.yaml`, `antenna.yaml`; implementation in `html.go`
(`WriteItem`), `harvest.go` (`saveItem`), `cmarkdoc.go` (`CommonMark`),
`generator.go`/`page.go`/`schema.go` (post/page rendering — unaffected).

## 1. Problem Statement

Antenna App currently gives fine-grained control over **local post** metadata
publication via `allowed_meta_fields` in `page.yaml` (documented in
`antenna-metadata.7.md`). Every YAML front-matter field on a post is emitted
as a `<meta>` element (and a `data-pagefind-filter` attribute) unless that
allowlist restricts it.

There is no equivalent control for **harvested feed items** — the entries
pulled from external RSS/Atom/JSON sources into a collection's aggregation
page (`COLLECTION.html`). Feed items differ from local posts in ways that
matter for rendering:

- Their metadata comes from an external feed, not YAML front matter you
  authored — you don't control which fields exist or their quality.
- Each item's body content can come from **two different sources with two
  different trust levels** (see §5.0) — this is new information confirmed by
  reading the code, and it is the central fact this revision is built
  around. Feed items are not simply "one field called `description`."
- Each item belongs to a **source feed/channel**, a concept with no
  analogue for local posts (a post has one author: you).
- The presence of a `link` back to the original article is not guaranteed
  and needs an explicit, predictable fallback behavior.

This document proposes extending `page.yaml` with a new `items:` block that
gives per-collection control over how feed items are rendered into the
aggregation HTML, following the same declarative, non-template-language
philosophy the project already uses (see `decisions.md`: "avoid template
languages, they get too involved to document and explain").

**Explicit goal for this revision:** give more control over the *displayed*
content of feed items, while leaving local-post rendering completely
untouched. When Markdown content is available for a feed item, it must be
preferred over the raw feed `description` — Markdown is the safer path,
because it already renders through a non-`unsafe` Markdown pipeline.
Concretely: **this is already how the code behaves today** (`html.go`
`WriteItem` prefers `sourceMarkdown` over `description` when present). That
behavior was previously an implicit, undocumented fallback; this proposal
turns it into an explicit, declared part of the design (§5.0) rather than
introducing new behavior.

## 2. Design Goals

1. **Consistency with existing config.** Reuse the `allowed_meta_fields`
   mental model (ordered allowlist of field names) rather than inventing a
   new mechanism.
2. **No template language.** No `{{ }}` placeholders, no conditional
   expressions embedded in strings. Configuration is a set of named,
   independently documentable options — each option does exactly one thing.
3. **Markdown-preferred, declared explicitly.** Feed item body content
   resolves from `sourceMarkdown` first, `description` only as a fallback
   (§5.0). This is existing behavior, now named and documented rather than
   left inside one `if` statement in `WriteItem`.
4. **Safe by default on the one path that has no established
   sanitization.** The `sourceMarkdown` path already renders through
   goldmark's safe mode (`CommonMark.ToHTML()`, no raw HTML passthrough) —
   that does not change. The raw `description` fallback path currently has
   *no* sanitization at all (`fmt.Fprintf` with `%s`, unescaped) — that is
   the one place this proposal changes default behavior, and it must be
   called out (§6).
5. **Explicit handling of missing data**, particularly missing `link`
   values, since the current code has no fallback behavior for this case at
   all.
6. **Backward compatible for local posts and pages — absolutely, no
   exceptions.** Nothing in this proposal touches `GeneratePosts`, `Page()`,
   `Post()`, or `CommonMark.ToUnsafeHTML()`. Posts and pages continue to
   render exactly as they do today, regardless of any `items:` configuration.
7. **Backward compatible for feed items, except where a called-out default
   is a deliberate improvement.** Collections without an `items:` block
   behave exactly as they do today, with two named exceptions — both called
   out individually rather than left implicit (§6): the raw-`description`-
   fallback safety default (§5.6), and the screen-reader-friendly anchor
   label default (§5.2, goal 9 below).
8. **Stay in `page.yaml`.** Item rendering is presentational — it belongs
   with the file that already owns "how to construct an HTML page," not in
   `antenna.yaml`, which owns collection registration and SQL `filter`
   statements.
9. **Screen-reader-friendly anchor text by default.** The current default —
   anchor text is the raw link URL — is read aloud character-by-character by
   screen readers, which is a poor listening experience. Consistent with this
   project's existing accessibility decisions (`decisions.md` DEC-016
   skip-navigation link, DEC-019 `<time>` element, DEC-020 heading warning),
   the default anchor label changes to a fixed, human-readable call-to-action
   string rather than the URL (§5.2). This is an intentional, called-out
   behavior change, not a silently changed default.

## 3. Non-Goals

- This document does not propose per-feed override configuration (e.g.
  "trust HTML from feed A but not feed B"). See §7 for a suggested home for
  that as a follow-up.
- **This document does not change local post or page rendering in any way.**
  `allowed_meta_fields`, front-matter `<meta>`/PageFind-filter emission, and
  `ToUnsafeHTML()` usage for posts/pages are entirely out of scope.
- **This document does not change which fields are emitted as PageFind
  `data-pagefind-filter` attributes on feed items** (categories, `dc_*`
  fields, `author`, `label`, `channel` — built in `WriteItem`'s `filters`
  slice). `decisions.md` DEC-010 already establishes that feed item metadata
  is always rendered in full for that purpose. The new `items.fields` in
  this proposal governs only the human-visible body fields (title, date,
  content, link, source) — a separate code path in `WriteItem` from the
  `filters` slice. This is an explicit carve-out so implementation does not
  conflate the two.
- This document does not address feed harvesting/fetching logic itself,
  except where §5.0 must describe the existing `harvest.go` behavior that
  produces `sourceMarkdown` in order to declare the rendering precedence
  correctly.

## 4. Current State (confirmed by reading the code)

Confirmed via `antenna-metadata.7.md` and `antenna-post.7.md`:

- `page.yaml` already supports `allowed_meta_fields`, an ordered list
  restricting which front-matter keys get published as `<meta>` tags for
  posts and pages only (`decisions.md` DEC-010).
- Local post Markdown is rendered via `CommonMark.ToUnsafeHTML()` — raw HTML
  passes through unchanged. This applies to files the *user* authors and
  trusts, and is unaffected by this proposal.

Confirmed via direct source reading (`html.go`, `harvest.go`, `cmarkdoc.go`,
`sql_stmts.go`):

- **Two rendering pipelines with two trust levels already exist:**
  `CommonMark.ToHTML()` (`cmarkdoc.go:326`) uses goldmark **without**
  `html.WithUnsafe()` — raw HTML embedded in Markdown text is dropped/
  escaped. `CommonMark.ToUnsafeHTML()` (`cmarkdoc.go:353`) enables
  `html.WithUnsafe()` — raw HTML passes through. Posts/pages
  (`generator.go:341`, `page.go:176`, `schema.go:597`) use the unsafe
  renderer; feed items (`html.go:153`) use the safe renderer. This trust
  boundary is correct as-is and this proposal preserves it.
- **`sourceMarkdown` is already preferred over `description` at render
  time.** `WriteItem` (`html.go:148-156`): `content := description`, then
  overwritten `if sourceMarkdown != ""` with the goldmark-safe-rendered
  Markdown. This existing precedence is what §5.0 formalizes.
- **`sourceMarkdown` is populated automatically at harvest time**
  (`harvest.go:291-315`, function `saveItem`): if the feed provides a
  `source:markdown` extension, that value is used (marked `FIXME` in the
  source — never verified against a real feed); otherwise the feed's HTML
  `description` is converted to Markdown via `html2md`
  (`github.com/JohannesKaufmann/html-to-markdown`). `sourceMarkdown` stays
  empty only when the item has no description at all, or when the `html2md`
  conversion fails (logged as a warning to stderr).
- **The raw-`description`-fallback path has no sanitization at all.** When
  `sourceMarkdown` is empty, `content := description` is interpolated
  directly via `fmt.Fprintf(out, ..., "%s", ..., content, ...)` — `%s` does
  not HTML-escape. This is the one path this proposal must close (§5.6).
- **No fallback exists today for a missing `link`.** An empty `link` value
  currently renders `<a href="">` with the link string (also empty) as the
  anchor text — confirms the concern in §1.
- **Anchor text is currently the link URL itself, not the item title.**
  `html.go:164-166`: `<a href=%q>%s</a>`, with `link` passed for both the
  `href` and the visible text. A `label_field` default of `title` would
  therefore be a real behavior change for existing sites (see §5.2, §6).
- **Dates are stored, not parsed, at render time.** `saveItem`
  (`harvest.go:249-259`) stores `pubDate`/`updated` as
  `"2006-01-02 15:04:05"` (Go reference layout) when `gofeed` successfully
  parses the feed's date, but falls back to the feed's raw, unparsed date
  string (any format) when it doesn't. `WriteItem` then only ever
  string-slices the first 10 characters — it never calls `time.Parse`. A
  configurable `date_format` needs its own explicit input-layout assumption
  and parse-failure fallback (§5.3).
- **Items table schema** (`sql_stmts.go:45-61`): `link`, `postPath`,
  `title`, `description`, `authors`, `enclosures`, `guid`, `pubDate`,
  `dcExt`, `channel`, `sourceMarkdown`, `status`, `label`, `updated`,
  `categories`. `source` (as used conceptually in §5.1) is not a literal
  column — it is derived from `label`/`channel`.

## 5. Proposed Schema

Add an `items:` block to `page.yaml`, alongside the existing
`allowed_meta_fields` key.

```yaml
allowed_meta_fields:
  - title
  - author
  - description
  - keywords
  - series
  - seriesNumber

items:
  fields:
    - title
    - source
    - pubDate
    - content

  link:
    label_field: static
    label_fallback: "Continue reading"
    required: false
    missing: unlinked

  date_format: "2006-01-02"
  content_max_length: 280
  show_source: true
  html: strip
```

### 5.0 Content source resolution (declared explicitly — not new behavior)

For every rendered item, the body content resolves in this order:

1. **`sourceMarkdown`, if non-empty.** Populated at harvest time from the
   feed's `source:markdown` extension when present, otherwise
   auto-converted from the feed's HTML `description` via `html2md`
   (§4). Rendered via `CommonMark.ToHTML()` — goldmark safe mode, no raw
   HTML passthrough.
2. **`description` (raw, feed-supplied HTML/text), only when
   `sourceMarkdown` is empty** — either the item genuinely has no
   description, or the harvest-time `html2md` conversion failed.

This precedence already exists in `WriteItem` today; this proposal does not
change it, only makes it an explicit, documented part of the configuration
model instead of an implicit fallback inside one `if` statement. Markdown is
preferred because it is already rendered through the same safe pipeline used
for feed items today; the raw `description` fallback is the one path with no
established sanitization, and is what §5.6 must close.

**This resolution is local to feed items only.** A post or page is a single
authored Markdown document with no `description`/`sourceMarkdown` duality,
and always renders via `CommonMark.ToUnsafeHTML()`, unchanged by this
proposal.

### 5.1 `items.fields` (list of string, default: all known item fields)

Ordered allowlist of item fields to render in the visible article body,
mirroring `allowed_meta_fields`. Known item fields: `title`, `link`,
`pubDate`, `content`, `source` (derived from the parent channel's `label`/
`channel` columns, not a literal item column).

`content` refers to the **resolved** body per §5.0 (Markdown-preferred) —
not literally the raw `description` table column. Selecting `content` says
nothing about which underlying source produced it; that is governed
entirely by §5.0's precedence and §5.6's trust handling.

Fields not in this list are not rendered for items in this collection.
Order in the list determines render order. This list has no effect on the
PageFind `data-pagefind-filter` attributes emitted in the `filters` slice
(categories, `dc_*`, author, label, channel) — see §3.

### 5.2 `items.link` (object)

Controls the anchor generated for each item.

| Key | Type | Default | Description |
|---|---|---|---|
| `label_field` | string | `static` | Item field supplying the anchor text, or the literal sentinel `static` (see below). **Default changed from current behavior** (§6, goal 9): renders a fixed, screen-reader-friendly label instead of the raw link URL. Set to `link` to restore the current URL-as-anchor-text behavior; set to `title` to use the item title instead. |
| `label_fallback` | string | `"Continue reading"` | When `label_field` names a field: literal text used when that field's value is empty or missing. When `label_field: static` (the default): the literal anchor text used for **every** item, unconditionally. |
| `required` | bool | `false` | If `true`, items lacking a `link` value fail collection generation (hard error) rather than being handled per `missing`. |
| `missing` | enum: `unlinked` \| `omit` \| `source_link` | `unlinked` | Per-item fallback when `link` is absent and `required` is `false`. |

**`label_field: static` is now the default**, not merely an available
option. **Why the default changed:** the current behavior — anchor text is
the raw link URL — is read aloud character-by-character by screen readers,
a poor listening experience for a link blog or aggregator meant to be
browsed by ear as much as by eye. This is a deliberate, explicit default
change (§6), consistent with this project's existing accessibility
decisions (`decisions.md` DEC-016, DEC-019, DEC-020), not a silent one.

A field-derived label can also never usefully fall back to `label_fallback`
when the chosen field is one that is realistically always present — `link`
is a case in point: `label_fallback` would never fire in practice, since a
rendered item's `link` is essentially never empty once it exists in the
`items` table. Sites that want a fixed call-to-action label regardless of
per-item data set `label_fallback` directly; `label_field: static` (the
default) means per-item data is ignored entirely and `label_fallback` is
always the rendered anchor text. Example — a site wanting "read me"
instead of the new default "Continue reading," leaving title/description
rendering untouched, needs only:

```yaml
items:
  link:
    label_fallback: "read me"
```

Restoring the pre-this-proposal behavior (anchor text = URL) requires an
explicit opt-out:

```yaml
items:
  link:
    label_field: link
```

`missing` semantics:

- `unlinked` — render `label_field`'s value (or `label_fallback`) as plain
  text, no `<a>` element.
- `omit` — exclude the item from the generated aggregation entirely.
- `source_link` — use the parent feed/channel's URL in place of the missing
  item URL. If the channel URL is *also* empty, fall back to `unlinked`
  rather than producing an empty `href` — this recursive-fallback case was
  previously undefined and must be pinned down before implementation.

Rationale: a boolean `show_link: true/false` cannot express these three
distinct, useful behaviors, so an enum is used instead of overloading a
boolean.

### 5.3 `items.date_format` (string, default: `"2006-01-02"`)

A Go `time` package reference-layout string, applied to `pubDate`/`updated`
when rendered.

**Input-layout assumption, made explicit:** `pubDate`/`updated` are stored
as `"2006-01-02 15:04:05"` when `gofeed` successfully parsed the feed's
date (§4); implementation must `time.Parse` against that layout before
reformatting. When the stored value does not parse against that layout
(the feed's original unparsed date string, in an unknown format — see §4),
implementation must fall back to the same truncation current code already
performs — the first 10 characters of the stored string (or the whole
string, if shorter) — rather than erroring the whole `generate` run **or**
displaying the full untruncated raw string. The latter was considered and
rejected: it would silently change output length for any item whose date
`gofeed` couldn't parse, which is exactly the kind of unflagged behavior
change goal 7 (§2) rules out. Matching the existing 10-character truncation
exactly means this fallback path produces byte-identical output to today's
code, with no exception to call out. There is no general-purpose parser for
arbitrary feed date formats already present in the codebase, so this
fallback is a design requirement, not an edge case to defer.

### 5.4 `items.content_max_length` (int, default: unset / no truncation)

Renamed from `description_max_length` for consistency with `content` (§5.0,
§5.1) — this truncates the **resolved, pre-render source text** (either
`sourceMarkdown` or raw `description`, whichever won precedence in §5.0),
on a word boundary, **before** Markdown/HTML conversion. Truncating
post-render HTML is explicitly rejected — it risks cutting mid-tag and
producing unbalanced markup.

### 5.5 `items.show_source` (bool, default: `true`)

Whether to render the originating feed/channel label alongside each item.
Meaningful only for feed items (an aggregation may draw from many
channels); has no equivalent for local posts, which is why it lives here
rather than in `allowed_meta_fields`.

### 5.6 `items.html` (enum: `strip` \| `escape` \| `unsafe`, default: `strip`)

Because the resolved body can come from either of two pipelines (§5.0),
this setting has two distinct effects depending on which one resolved for a
given item:

**When content resolved from `sourceMarkdown` (the common case):**

- `strip` / `escape` (default) — render via `CommonMark.ToHTML()`, the same
  goldmark-safe renderer already used for items today. No code path change;
  this only formalizes existing behavior as a declared setting.
- `unsafe` — render via `CommonMark.ToUnsafeHTML()` instead — the same
  renderer already used for local posts. Must be an explicit, deliberate
  per-collection opt-in; raw HTML embedded in the Markdown then passes
  through unchanged.

**When content resolved from raw `description` (fallback only, per §5.0 —
no `sourceMarkdown` available for that item):**

- `strip` (default) — strip HTML tags, render remaining text as plain text.
- `escape` — HTML-escape the raw description so tags display literally
  rather than render.
- `unsafe` — pass the raw description through unchanged. Not recommended:
  this is the one path with no established sanitization today (§4).

**This default (`strip` on the raw-description-fallback path) is a real
behavior change** from current code, which today interpolates raw
`description` completely unescaped whenever `sourceMarkdown` is empty (§4).
The Markdown-path default is **not** a behavior change — it already renders
via `ToHTML()` today. Only the fallback-path default needs to be called out
in release notes (§6); the common case is unaffected.

### 5.7 Theme integration: `items.yaml`

A theme directory (`antenna-themes.7.md`) may include an `items.yaml` file
alongside `header.md`, `nav.md`, `footer.md`, `top_content.md`,
`bottom_content.md`, and `head.yaml`. `antenna apply THEME_PATH` merges its
content under a single `items:` key in the resulting `page.yaml` — the same
one-file-to-one-wrapper-key pattern already used for `header.md` → `header:`,
`nav.md` → `nav:`, etc.

This is **not** the `head.yaml` pattern. `head.yaml`'s own top-level keys
(`title`, `meta`, `link`, `script`, `style`) are merged flat into
`page.yaml`'s top level, with no wrapper. `items.yaml` cannot follow that
pattern: its own sub-key `link` (§5.2, the anchor-fallback config) would
collide with `head.yaml`'s top-level `link:` (the page's `<link>` elements
list) the moment both landed in the same flat namespace. `items.yaml`'s
entire content is therefore nested under one `items:` key when applied,
exactly mirroring the Markdown-file theme components rather than
`head.yaml`.

Example `theme/items.yaml`:

```yaml
fields:
  - title
  - source
  - pubDate
  - content
link:
  label_fallback: "read me"
date_format: "Jan 2, 2006"
content_max_length: 320
show_source: true
html: strip
```

(`label_field` is omitted — `static` is now the default, §5.2 — so only
`label_fallback` needs to be set to override the label text.)

After `antenna apply theme/`, this appears in `page.yaml` as:

```yaml
items:
  fields:
    - title
    - source
    - pubDate
    - content
  link:
    label_fallback: "read me"
  date_format: "Jan 2, 2006"
  content_max_length: 320
  show_source: true
  html: strip
```

As with the other theme files, `antenna apply` should not overwrite an
existing `items:` block in `page.yaml` without the same conflict handling
already used for the other theme-sourced keys (`antenna-themes.7.md`).

## 6. Backward Compatibility

- **Local posts and pages: zero behavior change**, unconditionally. No part
  of `items:` configuration reaches `GeneratePosts`, `Page()`, `Post()`, or
  `ToUnsafeHTML()`.
- Collections without an `items:` block: feed item rendering is unchanged
  from current output, **except** for the two specific, deliberate defaults
  below, which apply even without an `items:` block:
  1. `html: strip` on the raw-`description`-fallback path (§5.6) — a safety
     fix. Applies only to items that lack `sourceMarkdown`. Must be called
     out in release notes; sites that depend on raw feed HTML passing
     through in that fallback case need an explicit opt-out
     (`html: unsafe`).
  2. `label_field: static` / `label_fallback: "Continue reading"` (§5.2,
     goal 9) — an accessibility fix. Anchor text for **every** item changes
     from the raw link URL to a fixed, human-readable label. This is a
     visible change to every existing collection's rendered output, not a
     narrow edge case like (1), and must be prominently called out in
     release notes. Sites that want the previous URL-as-anchor-text
     behavior back need an explicit opt-out (`label_field: link`).
  3. Nothing else changes by default — the `date_format` fallback for
     unparseable stored dates is defined to match current truncation
     behavior exactly (§5.3), specifically so it does *not* need to be
     added to this exception list.
- All keys are optional; the block itself is optional.
- Unknown keys under `items:` should be ignored (forward-compatible),
  consistent with tolerant YAML parsing elsewhere in the project.

## 7. Deferred: Per-Feed Overrides (Follow-up Proposal)

A likely follow-up need: trusting HTML from one feed but not another within
the same collection. This cannot cleanly live in the collection's link-list
Markdown file without complicating the existing
`- [FEED_LABEL](FEED_URL "OPTIONAL_FEED_DESCRIPTION")` syntax, which the
project intentionally keeps simple.

Recommended home: `antenna.yaml`, as a sibling to the existing per-
collection `filter` attribute (already a list of SQL statements scoped to
one collection). E.g.:

```yaml
collections:
  - name: travel.md
    filter:
      - "UPDATE items SET status = 'review'"
      - "UPDATE items SET status = 'published' WHERE pubDate >= date('now', '-21 days')"
    items_overrides:
      "https://example.com/feed.xml":
        html: unsafe
```

This keeps `antenna.yaml` as the place for collection-scoped, per-source
configuration, and `page.yaml` as the place for whole-collection
presentation structure. Not part of this proposal's implementation scope;
noted here so the schema in §5 doesn't need to be reworked later to
accommodate it.

## 8. Implementation Notes (for whoever picks this up)

- Suggested Go shape for the config struct, to sit alongside whatever
  struct currently backs `allowed_meta_fields`:

  ```go
  type ItemsConfig struct {
      Fields            []string   `yaml:"fields"`
      Link              LinkConfig `yaml:"link"`
      DateFormat        string     `yaml:"date_format"`
      ContentMaxLength  int        `yaml:"content_max_length"`
      ShowSource        bool       `yaml:"show_source"`
      HTML              string     `yaml:"html"` // "strip" | "escape" | "unsafe"
  }

  type LinkConfig struct {
      LabelField    string `yaml:"label_field"`    // default "static" (§5.2, goal 9); item field name, or "static"
      LabelFallback string `yaml:"label_fallback"` // default "Continue reading"; fallback text, or the fixed label when LabelField == "static"
      Required      bool   `yaml:"required"`
      Missing       string `yaml:"missing"` // "unlinked" | "omit" | "source_link"
  }
  ```

- **Content resolution should be one function**, implementing §5.0 and
  §5.6 together, e.g.:
  `resolveItemContent(description, sourceMarkdown string, cfg ItemsConfig) (html string, err error)`
  — decides `sourceMarkdown` vs `description` (§5.0), applies
  `content_max_length` truncation to the pre-render source text (§5.4), then
  renders via `ToHTML()` or `ToUnsafeHTML()` per `cfg.HTML` and which source
  won (§5.6). `WriteItem` calls this once instead of inlining the current
  `content := description; if sourceMarkdown != "" { ... }` block.
- **Date formatting** needs a small helper that attempts
  `time.Parse("2006-01-02 15:04:05", raw)` first, formats with
  `cfg.DateFormat` on success, and on parse failure falls back to the first
  10 characters of the raw stored string (matching current `WriteItem`
  truncation exactly, §5.3) — never the full untruncated string, and never a
  hard error.
- Rendering should remain plain Go string/HTML-builder composition — no
  introduction of `text/template`/`html/template`, per the project's stated
  design constraint. `strings.Builder`, `html.EscapeString`, and a small
  hand-rolled tag stripper (for the raw-`description` `strip` mode) are
  sufficient; goldmark already exists as a dependency and handles the
  Markdown-path cases.
- Validation: `items.html`, `items.link.missing` should be validated
  against their enum values at config-load time (wherever `page.yaml` is
  currently parsed), producing a clear error for typos rather than silently
  falling back.
- Testing: given the project's existing `_test.go` pattern (see
  `html_test.go`, `cmarkdoc_test.go`), an `items_test.go` addition covering:
  each `missing` mode (including the channel-URL-also-empty case for
  `source_link`, §5.2), each `html` mode **crossed with** which content
  source resolved (§5.0 × §5.6 — four meaningful combinations, not two),
  `content_max_length` boundary behavior (exact-length, word-boundary
  truncation), and `date_format` with both a parseable and an unparseable
  stored date string (§5.3).

## 9. Open Questions

1. ~~Should `items.fields` support any fields beyond the confirmed set?~~
   Resolved — confirmed against `sql_stmts.go`: `title`, `link`, `pubDate`,
   `content` (resolved per §5.0), `source` (derived, not a column).
2. Should `content_max_length` truncation add an ellipsis or other
   indicator, and should that be configurable or fixed?
3. Is per-collection `items:` sufficient, or is there a near-term need for
   per-theme defaults (i.e. a theme ships a suggested `items:` block that
   `page.yaml` can override)? Out of scope for this proposal but worth
   flagging before implementation locks in the override precedence order.
4. Should `html: unsafe` require a corresponding warning at `generate` time
   (similar in spirit to the `post` action's unsafe-mode warning in its man
   page), for both the Markdown-path and the raw-description-path meanings
   of `unsafe` (§5.6), so the risk is visible to whoever runs the command?
5. The `source:markdown` feed extension path in `harvest.go` is marked
   `FIXME` — never verified against a real feed. Should this proposal's
   implementation include finding or constructing a test feed that uses it,
   so §5.0's first precedence branch has real test coverage rather than
   only the `html2md`-converted branch?
