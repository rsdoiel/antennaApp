# Implementation Plan: Feed Item Formatting Control

Implements the design in `design_item_formatting.md`.
Decisions referenced throughout are in `decisions.md` (DEC-022 through DEC-031).

Tests are written before implementation in every phase (TDD).
Each task is independent enough to commit separately.
Tasks within a phase may be done in parallel; phases must be done in order
unless noted otherwise.

---

## Phase 1 — Config types, defaults, validation

No callers changed yet. New types and pure functions only.

### Task 1.1 — Test `ItemsConfig`/`LinkConfig` YAML unmarshaling and defaults

Create `items_config_test.go`. Write `TestItemsConfigDefaults` covering:

| Case | Input YAML | Expected after `applyDefaults()` |
|---|---|---|
| Empty block | `items: {}` | `Fields=[title,source,pubDate,content]`, `Link.LabelField="static"`, `Link.LabelFallback="Continue reading"`, `Link.Missing="unlinked"`, `DateFormat="2006-01-02"`, `ShowSource=true`, `HTML="strip"` |
| Partial override | `items: {html: unsafe}` | only `HTML="unsafe"` differs from defaults |
| `show_source: false` explicit | `items: {show_source: false}` | `ShowSource=false` (not overwritten back to true) |
| No `items:` key at all | `page.yaml` omits `items:` | `Generator.Items` zero value; `applyDefaults()` still produces the same defaults as the empty-block case |

Verify tests fail before implementation (red).

### Task 1.2 — Test `validate()`

Add `TestItemsConfigValidate` covering:

| Case | Input | Expected |
|---|---|---|
| `html: bogus` | invalid enum | error mentioning `items.html` |
| `link.missing: bogus` | invalid enum | error mentioning `items.link.missing` |
| All valid enum values | `strip`/`escape`/`unsafe`, `unlinked`/`omit`/`source_link` | no error |

Verify tests fail (red).

### Task 1.3 — Implement `ItemsConfig`, `LinkConfig`, `applyDefaults()`, `validate()` in `generator.go`

Add the types and methods per `design_item_formatting.md`. Add `Items
ItemsConfig` to `Generator`.

Run `TestItemsConfigDefaults` and `TestItemsConfigValidate` — must pass
(green). Run `go build ./...` — must compile cleanly.

---

## Phase 2 — Content resolution and truncation

Depends on Phase 1 for `ItemsConfig`. Independent of Phases 3–4.

### Task 2.1 — Test `truncateWords`

Add `TestTruncateWords` to `html_test.go`:

| Case | Input | maxLen | Expected |
|---|---|---|---|
| Shorter than max | `"hello"` | 10 | `"hello"` (unchanged) |
| Exact length | `"hello"` | 5 | `"hello"` |
| Mid-word cut avoided | `"the quick brown fox"` | 12 | `"the quick"` (backs off to last space before/at 12, not `"the quick br"`) |
| No whitespace before limit | `"supercalifragilistic"` | 5 | falls back to a hard cut at `maxLen` (documented as the no-good-boundary case) |

Verify red.

### Task 2.2 — Test `stripTags`

Add `TestStripTags`:

| Case | Input | Expected |
|---|---|---|
| Plain text | `"hello"` | `"hello"` |
| Simple tags | `"<p>hello</p>"` | `"hello"` |
| Nested tags | `"<div><b>hi</b> there</div>"` | `"hi there"` |
| Unclosed/malformed tag | `"<p>hi"` | `"hi"` (document limitation: not a full parser) |

Verify red.

### Task 2.3 — Test `resolveItemContent`

Add `TestResolveItemContent`:

| Case | `description` | `sourceMarkdown` | `cfg.HTML` | Expected |
|---|---|---|---|---|
| Markdown present, default | `<p>raw</p>` | `**bold**` | `strip` | rendered via `ToHTML()`: `<strong>bold</strong>`-ish, `<p>raw</p>` never used |
| Markdown present, unsafe | any | `<script>x</script>` | `unsafe` | rendered via `ToUnsafeHTML()`, raw `<script>` passes through |
| No Markdown, default | `<p>raw & unsafe</p>` | `""` | `strip` | tags stripped: `"raw & unsafe"` (entity handling per `stripTags`) |
| No Markdown, escape | `<b>hi</b>` | `""` | `escape` | `"&lt;b&gt;hi&lt;/b&gt;"` |
| No Markdown, unsafe | `<b>hi</b>` | `""` | `unsafe` | `"<b>hi</b>"` unchanged |
| Truncation applied pre-render | long `sourceMarkdown` | — | any | truncated markdown source is what gets rendered, not truncated HTML |

Verify red.

### Task 2.4 — Implement `truncateWords`, `stripTags`, `resolveItemContent` in `html.go`

Per `design_item_formatting.md`. Run all Phase 2 tests — must pass (green).

---

## Phase 3 — Link resolution

Depends on Phase 1. Independent of Phase 2.

### Task 3.1 — Test `resolveItemLink`

Add `TestResolveItemLink` to `html_test.go`:

| Case | `link` | `title` | `channelURL` | `cfg` | Expected |
|---|---|---|---|---|---|
| Normal, default label (DEC-026) | `"https://x"` | `"T"` | — | defaults (post-`applyDefaults()`: `label_field: static`, `label_fallback: "Continue reading"`) | `Href="https://x"`, `Label="Continue reading"` regardless of title/link |
| `label_field: link` (explicit opt-out) | `"https://x"` | `"T"` | — | `label_field: link` | `Label="https://x"` (restores pre-DEC-026 behavior) |
| `label_field: title` | `"https://x"` | `"T"` | — | `label_field: title` | `Label="T"` |
| `label_field: static`, custom fallback | `"https://x"` | `"T"` | — | `label_field: static, label_fallback: "read me"` | `Label="read me"` regardless of title/link |
| Empty field falls back | `"https://x"` | `""` | — | `label_field: title, label_fallback: "Read more"` | `Label="Read more"` |
| Missing link, default | `""` | `"T"` | — | defaults | `AsPlainText=true`, `Label="Continue reading"` |
| Missing link, omit | `""` | — | — | `missing: omit` | `Omit=true` |
| Missing link, source_link with channel | `""` | — | `"https://chan"` | `missing: source_link` | `Href="https://chan"` |
| Missing link, source_link, channel also empty | `""` | — | `""` | `missing: source_link` | falls back to `AsPlainText=true` (DEC-027) |
| Missing link, required | `""` | — | — | `required: true` | returns an error |

Verify red.

### Task 3.2 — Implement `resolveItemLink` and `LinkResolution` in `html.go`

Per `design_item_formatting.md`. Run `TestResolveItemLink` — must pass
(green).

---

## Phase 4 — Date formatting

Depends on Phase 1 only. Fully independent of Phases 2–3.

### Task 4.1 — Test `formatItemDate`

Add `TestFormatItemDate`:

| Case | `raw` | `layout` | Expected |
|---|---|---|---|
| Parseable, custom layout | `"2020-04-11 00:00:00"` | `"Jan 2, 2006"` | `"Apr 11, 2020"` |
| Parseable, default layout | `"2020-04-11 00:00:00"` | `"2006-01-02"` | `"2020-04-11"` |
| Unparseable, longer than 10 chars (raw feed string) | `"Sat, 11 Apr 2020"` | `"Jan 2, 2006"` | `"Sat, 11 Ap"` (first 10 characters — matches current `WriteItem` truncation exactly, DEC-028) |
| Unparseable, 10 chars or shorter | `"bad-date"` | any | `"bad-date"` unchanged (nothing to truncate) |
| Empty | `""` | any | `""` |

Verify red.

### Task 4.2 — Implement `formatItemDate` in `html.go`

Per `design_item_formatting.md`. Run `TestFormatItemDate` — must pass (green).

---

## Phase 5 — `WriteItem` / `WriteHTML` integration

Depends on Phases 1–4 all being complete. This is the phase that actually
changes existing, currently-tested behavior — run the full existing
`html_test.go` suite before starting, to have a known-green baseline.

### Task 5.1 — Extend existing `WriteItem` tests for the new `cfg` parameter

Update `html_test.go`'s existing `TestWriteItem_*` cases to pass a
default-valued `ItemsConfig` (post-`applyDefaults()`). Two of the two
existing exceptions from `decisions.md` §6/DEC-024/DEC-026 apply here, and
existing test expectations must be updated accordingly rather than asserted
byte-for-byte unchanged:

1. **Anchor text** in every existing `TestWriteItem_*` case that asserts an
   `<a href=...>` element must be updated: the visible anchor text changes
   from the item's link URL to `"Continue reading"` (DEC-026), even though
   `href` itself is unchanged. This is the larger, deliberately-flagged
   default change — confirm each pre-existing test case's expected string is
   updated, not just left passing by coincidence.
2. **Raw-description fallback** cases (items with no `sourceMarkdown`) change
   from unescaped raw HTML to tag-stripped plain text (DEC-024) — only
   relevant if any existing test exercises that branch; confirm whether one
   does, and update its expectation if so.

Everything else — content resolution, date rendering (post-DEC-028 fix),
field ordering, PageFind `filters` attribute construction — must remain
byte-for-byte unchanged. This narrower claim is the actual regression guard;
don't let it get confused with the two named exceptions above.

Add new cases:

| Case | `cfg` | Expected |
|---|---|---|
| `fields` restricts sections | `Fields: ["title", "content"]` | date and footer/link sections absent from output |
| `show_source: false` | — | source/label block absent |
| Omitted item | `link=""`, `missing: omit` | `WriteItem` writes nothing, returns `(true, nil)` or equivalent omit signal |
| Default label is accessible (DEC-026) | defaults, no `items:` override | anchor text is `"Continue reading"`, `href` still the item link |
| Opt back into URL-as-label | `label_field: link` | anchor text is the item's URL again, matching pre-DEC-026 output |
| Custom static label | `label_fallback: "read me"` (label_field left at default `static`) | anchor text is `"read me"` |

Verify new cases fail (red); existing-behavior regression cases should
already pass once Task 5.2's signature change compiles (confirm, don't
assume).

### Task 5.2 — Update `WriteItem` signature and body in `html.go`

Add `cfg ItemsConfig` parameter. Wire in `resolveItemLink` (checked first, for
`Omit`), `resolveItemContent`, `formatItemDate`, and `cfg.Fields`-based
section filtering, per `design_item_formatting.md`. Change the return
signature if needed to signal "item omitted" to the caller (e.g. `(bool,
error)` — omitted/not-omitted plus error).

Run all Phase 5 `TestWriteItem_*` tests — must pass (green).

### Task 5.3 — Update `WriteHTML` in `html.go`

Call `gen.Items.applyDefaults()` and `gen.Items.validate()` once at the start
of collection generation (return `validate()`'s error immediately — config
typos must abort `generate`, not silently misrender). Pass `gen.Items` to
every `WriteItem` call in the item-scan loop; skip counting/advancing past
items that come back omitted.

Run `go build ./...` and `go test ./...` — must both succeed.

---

## Phase 6 — Theme integration

Depends on Phase 1 (needs `ItemsConfig`) only. Independent of Phases 2–5,
except that Phase 7's smoke test exercises the full stack together.

### Task 6.1 — Test `isTargetFile` recognizes `items.yaml`

Add to `themes_test.go`: `TestIsTargetFile_ItemsYAML` — asserts
`isTargetFile("items.yaml") == true`.

Verify red.

### Task 6.2 — Add `"items.yaml"` to `isTargetFile` in `themes.go`

Run the new test — green.

### Task 6.3 — Test `updateItemsElement`

Add `TestUpdateItemsElement` to `themes_test.go` using a temp directory:

| Case | Setup | Expected |
|---|---|---|
| No `items.yaml` present | empty theme dir | `(false, nil)`, `gen.Items` unchanged |
| `items.yaml` present | valid YAML per the worked example in the design doc | `(true, nil)`, `gen.Items` matches the parsed content |
| Malformed `items.yaml` | invalid YAML | non-nil error |

Verify red.

### Task 6.4 — Implement `updateItemsElement` in `themes.go`; wire into `ApplyTheme`

Per `design_item_formatting.md`. Run `TestUpdateItemsElement` — green. Run
`go build ./...` and `go test ./...` — must both succeed.

---

## Phase 7 — Final verification

### Task 7.1 — Full test run

```bash
cd antennaApp && go test ./...
```

All tests must pass, including every pre-existing `TestWriteItem_*` case
(regression guard from Task 5.1).

### Task 7.2 — Build all programs

```bash
cd antennaApp && make build
```

All binaries must build cleanly.

### Task 7.3 — Smoke test: no `items:` block — verify the two named exceptions, byte-identical otherwise

1. Using an existing collection's `page.yaml` with no `items:` key, run
   `antenna harvest` then `antenna generate` against a feed with at least one
   item that has a `description` but no `source:markdown` extension.
2. Diff the generated aggregate HTML against a copy generated before this
   change. Two differences are expected and correct, not regressions:
   - Every item's footer anchor **text** now reads `"Continue reading"`
     instead of the item's URL (DEC-026) — `href` itself is unchanged.
   - The raw-`description`-fallback item (no `sourceMarkdown`) is now
     tag-stripped instead of raw HTML passthrough (DEC-024) — only relevant
     if the test feed actually has such an item; verify whether it does and
     note the result either way.
3. Everything else (content resolution/rendering, date display, field
   ordering, PageFind `filters` attributes) must be byte-identical to the
   pre-change output. Any other difference is a regression.

### Task 7.4 — Smoke test: default anchor label is screen-reader-friendly (DEC-026)

1. With no `items:` block at all, regenerate a collection and inspect the
   aggregate HTML. Verify every item's footer anchor reads
   `<a href="ITEM_LINK">Continue reading</a>` — not the URL spelled out as
   the link text.
2. Add `label_field: link` explicitly and regenerate — verify the anchor
   text reverts to the raw URL, confirming the opt-out path works.

### Task 7.5 — Smoke test: custom static label (the original motivating use case)

1. Add to a collection's `page.yaml`:
   ```yaml
   items:
     link:
       label_fallback: "read me"
   ```
   (`label_field` intentionally omitted — relying on the `static` default.)
2. Regenerate. Verify every item's footer anchor text reads "read me"
   regardless of title, and `href` is still each item's own link.

### Task 7.6 — Smoke test: `missing: omit` and `missing: source_link`

1. Manually clear the `link` column for one test item in the collection's
   SQLite3 database (`UPDATE items SET link = '' WHERE ...`).
2. With `missing: omit` — regenerate, verify that item is absent from the
   aggregate page entirely.
3. With `missing: source_link` — regenerate, verify that item's anchor now
   points at its channel's feed URL.

### Task 7.7 — Smoke test: `html: unsafe` opt-in

1. Harvest an item whose `sourceMarkdown` contains embedded raw HTML (e.g. an
   `<iframe>` or `<script>` a feed included inline).
2. With default config — verify the raw HTML does not appear in rendered
   output (goldmark safe mode).
3. With `items: {html: unsafe}` — verify the raw HTML now passes through
   unchanged.

### Task 7.8 — Smoke test: theme `items.yaml`

1. Create a theme directory with only `header.md` and `items.yaml` (using the
   worked example from `design_item_formatting.md`).
2. Run `antenna apply THEME_PATH page.yaml`.
3. Inspect `page.yaml` — verify the `items:` block matches `items.yaml`'s
   content, nested under one `items:` key (not flat-merged).
4. Regenerate the collection and confirm rendering matches Task 7.5's
   expectations.

### Task 7.9 — Smoke test: `date_format` and `content_max_length`

1. Set `date_format: "Jan 2, 2006"` and `content_max_length: 100` on a test
   collection.
2. Regenerate. Verify dates render in the new format and long item
   descriptions truncate at a word boundary at or before 100 characters of
   source text.

### Task 7.10 — Smoke test: unparseable stored date matches legacy truncation (DEC-028)

1. Manually set one test item's `pubDate` in the collection's SQLite3
   database to a string that doesn't match `"2006-01-02 15:04:05"` (e.g. a
   raw RFC822-style string like `"Sat, 11 Apr 2020 00:00:00 GMT"`).
2. Regenerate with a custom `date_format` set. Verify the rendered date for
   that item is the first 10 characters of the stored string, unchanged
   from what today's (pre-this-feature) `WriteItem` would have shown —
   *not* the full raw string, and *not* an error aborting `generate`.

---

## Task dependency summary

```
Phase 1 (1.1–1.3)
    │
    ├─── Phase 2 (2.1–2.4) ──┐
    ├─── Phase 3 (3.1–3.2) ──┼──── Phase 5 (5.1–5.3) ──── Phase 7
    ├─── Phase 4 (4.1–4.2) ──┘                                │
    └─── Phase 6 (6.1–6.4) ─────────────────────────────────────┘
```

Phases 2, 3, 4, and 6 all depend only on Phase 1 and may proceed in parallel.
Phase 5 depends on Phases 2, 3, and 4 (it wires all three into `WriteItem`).
Phase 7 depends on Phase 5 and Phase 6 both being complete.
