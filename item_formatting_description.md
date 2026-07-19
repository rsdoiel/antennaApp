# Design: Item Layout Control in `page.yaml`

Status: Proposal / Draft
Author: (your name)
Related project: [antennaApp](https://github.com/rsdoiel/antennaApp)
Related files: `page.yaml`, `antenna.yaml`, likely implementation in `generate.go` / `items.go` (methods `Generate`, `Items` dispatched from `antenna.go`)

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
- Their content (typically an item `description`) is third-party HTML/text
  and has no established trust boundary, unlike a post's Markdown, which
  the `antenna help post` docs explicitly describe as passing through
  "unsafe mode" because you trust files you author.
- Each item belongs to a **source feed/channel**, a concept with no
  analogue for local posts (a post has one author: you).
- The presence of a `link` back to the original article is not guaranteed
  and needs an explicit, predictable fallback behavior.

This document proposes extending `page.yaml` with a new `items:` block that
gives per-collection control over how feed items are rendered into the
aggregation HTML, following the same declarative, non-template-language
philosophy the project already uses (see `decisions.md`: "avoid template
languages, they get too involved to document and explain").

## 2. Design Goals

1. **Consistency with existing config.** Reuse the `allowed_meta_fields`
   mental model (ordered allowlist of field names) rather than inventing a
   new mechanism.
2. **No template language.** No `{{ }}` placeholders, no conditional
   expressions embedded in strings. Configuration is a set of named,
   independently documentable options — each option does exactly one thing.
3. **Safe by default.** Feed-sourced HTML must not pass through unescaped
   by default. Only local posts get "unsafe mode," and that is documented
   and intentional; feed items should not silently inherit that behavior.
4. **Explicit handling of missing data**, particularly missing `link`
   values, since the current docs and code (as far as accessible) do not
   define this behavior.
5. **Backward compatible.** Collections without an `items:` block behave
   exactly as they do today. All new keys have sane defaults.
6. **Stay in `page.yaml`.** Item rendering is presentational — it belongs
   with the file that already owns "how to construct an HTML page," not in
   `antenna.yaml`, which owns collection registration and SQL `filter`
   statements.

## 3. Non-Goals

- This document does not propose per-feed override configuration (e.g.
  "trust HTML from feed A but not feed B"). See §7 for a suggested home for
  that as a follow-up.
- This document does not change how local posts are rendered or how
  `allowed_meta_fields` behaves for posts.
- This document does not address feed harvesting/fetching logic — only the
  rendering of already-harvested items stored in the `items` SQLite3 table.

## 4. Current State (established facts vs. inference)

Confirmed via `antenna-metadata.7.md` and `antenna-post.7.md`:

- `page.yaml` already supports `allowed_meta_fields`, an ordered list
  restricting which front-matter keys get published as `<meta>` tags.
- Local post Markdown is rendered in "unsafe mode" — raw HTML passes
  through unchanged. This is documented as applying to files the *user*
  authors and trusts.

Confirmed via `antenna.go` (source read directly):

- The CLI dispatches distinct actions/methods for `Post`, `Harvest`,
  `Generate`, `Items` — i.e., feed items and local posts are already
  handled by separate code paths at the method level, which is favorable
  for adding item-specific config without touching post rendering.

Not confirmed (implementation files not accessible at time of writing):

- Whether feed item `description`/content fields currently receive *any*
  HTML sanitization before being written into aggregation HTML. This
  design assumes **no sanitization currently exists** and treats adding a
  safe default as part of the proposal, not merely documentation of
  existing behavior. This assumption should be verified against
  `generate.go`/`items.go` (or equivalent) before implementation.

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
    - description

  link:
    label_field: title
    label_fallback: "Read more"
    required: false
    missing: unlinked

  date_format: "2006-01-02"
  description_max_length: 280
  show_source: true
  html: strip
```

### 5.1 `items.fields` (list of string, default: all known item fields)

Ordered allowlist of item fields to render, mirroring
`allowed_meta_fields`. Known item fields (from the `items` SQLite3 table,
pending confirmation against schema): `title`, `link`, `pubDate`,
`description`, `source` (derived from the parent `channels` row, not a
literal column).

Fields not in this list are not rendered for items in this collection.
Order in the list determines render order.

### 5.2 `items.link` (object)

Controls the anchor generated for each item.

| Key | Type | Default | Description |
|---|---|---|---|
| `label_field` | string | `title` | Item field supplying the anchor text. Must name a field, not a template expression. |
| `label_fallback` | string | `"Read more"` | Literal text used when `label_field`'s value is empty or missing. |
| `required` | bool | `false` | If `true`, items lacking a `link` value fail collection generation (hard error) rather than being handled per `missing`. |
| `missing` | enum: `unlinked` \| `omit` \| `source_link` | `unlinked` | Per-item fallback when `link` is absent and `required` is `false`. |

`missing` semantics:

- `unlinked` — render `label_field`'s value (or `label_fallback`) as plain
  text, no `<a>` element.
- `omit` — exclude the item from the generated aggregation entirely.
- `source_link` — use the parent feed/channel's URL in place of the
  missing item URL.

Rationale: a boolean `show_link: true/false` cannot express these three
distinct, useful behaviors, so an enum is used instead of overloading a
boolean.

### 5.3 `items.date_format` (string, default: `"2006-01-02"`)

A Go `time` package reference-layout string, applied to `pubDate` when
rendered. This is a format specifier, not template logic — consistent with
`pubDate`'s existing documented format (`YYYY-MM-DD recommended`) in
`antenna-post.7.md` — and does not conflict with the project's
"avoid template languages" principle.

### 5.4 `items.description_max_length` (int, default: unset / no truncation)

If set, truncates the rendered `description` field to N characters
(implementation should truncate on a word boundary, not mid-word, and
should apply truncation *before* HTML handling per §5.6 to avoid producing
unbalanced tags).

### 5.5 `items.show_source` (bool, default: `true`)

Whether to render the originating feed/channel label alongside each item.
Meaningful only for feed items (an aggregation may draw from many
channels); has no equivalent for local posts, which is why it lives here
rather than in `allowed_meta_fields`.

### 5.6 `items.html` (enum: `strip` \| `escape` \| `unsafe`, default: `strip`)

Controls handling of HTML found in feed-sourced fields (primarily
`description`).

- `strip` (default) — remove HTML tags, render as plain text.
- `escape` — HTML-escape the content so tags display literally rather than
  render.
- `unsafe` — pass through unchanged, mirroring local-post "unsafe mode."
  Must be an explicit, deliberate opt-in per collection; **must not** be
  the default for any collection, given feed content has no trust
  boundary by default.

## 6. Backward Compatibility

- Collections without an `items:` block: behavior is unchanged from
  current output (pending §4's verification caveat — if no sanitization
  currently exists, introducing the `strip` default is a **behavior
  change**, not a pure default-preserving addition; this must be called
  out in release notes and ideally gated behind a version flag or an
  explicit opt-out (`html: unsafe`) for existing sites that depend on
  current raw-passthrough behavior, if any).
- All keys are optional; the block itself is optional.
- Unknown keys under `items:` should be ignored (forward-compatible),
  consistent with tolerant YAML parsing elsewhere in the project.

## 7. Deferred: Per-Feed Overrides (Follow-up Proposal)

A likely follow-up need: trusting HTML from one feed but not another
within the same collection. This cannot cleanly live in the collection's
link-list Markdown file without complicating the existing
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

- Verify against `generate.go`/`items.go` (exact filenames unconfirmed at
  time of writing — grep for `func (app *AntennaApp) Items` and
  `func (app *AntennaApp) Generate` if filenames differ) whether any
  sanitization currently exists on the `description` field before HTML
  output. This determines whether §5.6's default is additive or a breaking
  change (see §6).
- Suggested Go shape for the config struct, to sit alongside whatever
  struct currently backs `allowed_meta_fields`:

  ```go
  type ItemsConfig struct {
      Fields                 []string   `yaml:"fields"`
      Link                   LinkConfig `yaml:"link"`
      DateFormat             string     `yaml:"date_format"`
      DescriptionMaxLength   int        `yaml:"description_max_length"`
      ShowSource             bool       `yaml:"show_source"`
      HTML                   string     `yaml:"html"` // "strip" | "escape" | "unsafe"
  }

  type LinkConfig struct {
      LabelField    string `yaml:"label_field"`
      LabelFallback string `yaml:"label_fallback"`
      Required      bool   `yaml:"required"`
      Missing       string `yaml:"missing"` // "unlinked" | "omit" | "source_link"
  }
  ```

- Rendering should remain plain Go string/HTML-builder composition (e.g.
  `strings.Builder`, `html.EscapeString`, a small hand-rolled tag stripper
  or an existing sanitization library if one is already a dependency) —
  no introduction of `text/template`/`html/template`, per the project's
  stated design constraint.
- A single `renderItem(item Item, cfg ItemsConfig, channel Channel) string`
  (or equivalent) function is a reasonable implementation seam: takes one
  harvested item, the collection's `ItemsConfig`, and its parent channel
  (needed for `show_source` and `missing: source_link`), returns the HTML
  fragment for that item.
- Validation: `items.html`, `items.link.missing` should be validated
  against their enum values at config-load time (e.g. in whatever function
  currently parses `page.yaml`), producing a clear error for typos rather
  than silently falling back.
- Testing: given the project's existing `_test.go` pattern (see
  `cmarkdoc_test.go`, `css_test.go`), an `items_test.go` covering each
  `missing` mode, each `html` mode, and `description_max_length` boundary
  behavior (exact-length, truncation mid-word) would fit the established
  convention.

## 9. Open Questions

1. Should `items.fields` support any fields beyond the confirmed set
   (`title`, `link`, `pubDate`, `description`, `source`)? Needs schema
   check against the `items`/`channels` SQLite3 tables.
2. Should `description_max_length` truncation add an ellipsis or other
   indicator, and should that be configurable or fixed?
3. Is per-collection `items:` sufficient, or is there a near-term need for
   per-theme defaults (i.e. a theme ships a suggested `items:` block that
   `page.yaml` can override)? Out of scope for this proposal but worth
   flagging before implementation locks in the override precedence order.
4. Should `html: unsafe` require a corresponding warning at `generate`
   time (similar in spirit to the `post` action's unsafe-mode warning in
   its man page), so the risk is visible to whoever runs the command, not
   just to whoever wrote the YAML?
