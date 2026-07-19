# Antenna App — Decision Record

Decisions made during design exploration sessions, recorded before implementation begins.
Each entry states what was decided, why, and any alternatives that were considered and rejected.

---

## 2026-06-27 — Enhanced metadata processing for posts, pages, and feed items

### Context

The feature request `enhanced_front_matter_processing_feature_request.md` identified two bugs
and several improvements:

1. `GeneratePosts` renders YAML front matter as visible body text (a bug).
2. Front matter fields are not emitted as HTML `<meta>` elements.
3. PageFind 1.5+ faceted search attributes are not emitted.

During exploration the scope was widened to treat posts, pages, and harvested feed items
consistently, and to add feed item metadata that was being silently dropped on harvest.

---

### DEC-001 — Fix `GeneratePosts` to parse front matter before rendering

**Decision:** Replace `doc := &CommonMark{Text: sourceMarkdown}` in `GeneratePosts`
(`generator.go` ~line 314) with a proper `doc.Parse()` call. On parse error, fall back to
treating the entire source as body text.

**Why:** The bare struct literal leaves raw YAML front matter in `doc.Text`. Goldmark
converts the `---` delimiters to `<hr>` elements and the YAML content to a visible paragraph.
The `Page()` path already calls `LoadCommonMark` which runs `Parse()` correctly; `GeneratePosts`
must match that behaviour.

**Alternatives rejected:** Pre-stripping front matter with a regex before rendering — fragile
and duplicates logic already in `SplitFrontMatter`.

---

### DEC-002 — Add `categories JSON` column to the `items` table

**Decision:** Add a `categories JSON DEFAULT ''` column to the `items` table to store
`item.Categories []string` from harvested feeds (RSS `<category>`, Atom `<category term="">`).

**Why:** `gofeed.Item.Categories` is populated during harvest but currently discarded in
`saveItem`. Categories are the primary subject/topic vocabulary in RSS and Atom feeds and are
essential for faceted search. The `channels` table already has a `categories` column for
channel-level categories; the `items` table must match for consistent querying.

**Alternatives rejected:** Deriving categories from `dcExt.Subject` at render time — DC Subject
and RSS Category are separate vocabularies and must not be merged (see DEC-009).

---

### DEC-003 — Schema migration via sqlite3 CLI scripts, not auto-migration

**Decision:** Provide `migrate_categories.bash` and `migrate_categories.ps1` scripts that
operators run once against existing databases. Document the migration on the Antenna App
website. The application itself will not attempt auto-migration.

**Why:** Auto-migration on startup adds startup latency, requires the running process to have
DDL permissions, and can mask errors. Explicit operator-run scripts are auditable and
reversible. The existing `setupDatabase` function only creates tables for new databases;
extending it with migration logic would conflate two concerns.

**Migration statement:**
```sql
ALTER TABLE items ADD COLUMN categories JSON DEFAULT '';
```

---

### DEC-004 — Dublin Core extension fields are unmarshalled and rendered explicitly

**Decision:** When rendering feed items (in `WriteItem`), unmarshal the stored `dcExt` JSON
blob into `ext.DublinCoreExtension` and surface each non-empty field as a PageFind filter
attribute on the `<article>` element.

**Why:** `dcExt` is stored as an opaque JSON blob. It is a well-known, stable struct
(`DublinCoreExtension` in `gofeed/extensions`). Unmarshalling it is straightforward and gives
access to all 16 DC fields. The library and archival community relies on Dublin Core; rendering
it into searchable HTML metadata is a stated goal.

**DC fields available** (all `[]string`): Title, Creator, Author, Subject, Description,
Publisher, Contributor, Date, Type, Format, Identifier, Source, Language, Relation, Coverage,
Rights.

---

### DEC-005 — Dublin Core PageFind filter key naming uses `dc_` prefix with underscores

**Decision:** Dublin Core fields are exposed as PageFind filter keys using the pattern
`dc_fieldname` (lowercase, underscore separator): `dc_creator`, `dc_subject`, `dc_date`, etc.

**Why:** PageFind's inline filter syntax uses `:` as the key–value separator
(`data-pagefind-filter="key:value"`). A key of `dc:subject` would be parsed as filter key `dc`
with value `subject:value` — incorrect. Underscores are unambiguous. The `dc_` prefix
distinguishes DC fields from native feed fields (`author`, `category`) without a crosswalk.

**Alternatives rejected:** Flat names without prefix (`subject`, `creator`) — ambiguous when
both DC and native fields carry the same concept, and the user explicitly wants separate facets
per vocabulary.

---

### DEC-006 — No crosswalk between metadata vocabularies

**Decision:** RSS/Atom `category`, post front matter `keywords`, `series`, and DC `dc_subject`
are rendered as separate PageFind filter keys. No automatic mapping between them.

**Why:** Merging heterogeneous vocabularies is a defined crosswalk process that belongs in a
separate programme or processing step, not in the renderer. Conflating `category` and `keywords`
would produce incorrect faceted search results. This is a well-understood problem in the library
and archival community.

---

### DEC-007 — Feed label and channel are included as PageFind filters on feed items

**Decision:** Each feed item rendered in an aggregate page (`WriteItem`) emits both
`data-pagefind-filter="label:VALUE"` (human-readable feed name) and
`data-pagefind-filter="channel:VALUE"` (feed URL) on its `<article>` element.

**Why:** Label and channel let readers filter an aggregate page by source feed — "show me only
items from source X." Both are already available as parameters to `WriteItem` at render time.
Label is the more user-friendly value; channel is the canonical identifier. Both are useful and
the cost of including them is negligible.

---

### DEC-008 — Two rendering paths: `<head>` meta for posts/pages, `<article>` attributes for feed items

**Decision:**

- **Posts and pages** (one HTML file = one document): front matter fields are emitted as
  `<meta name="key" content="value">` and `<meta data-pagefind-filter="key[content]" content="value">`
  elements in `<head>`.
- **Aggregate feed pages** (one HTML file = many items): each item's metadata is emitted as
  `data-pagefind-filter="key:value"` inline attribute values on its `<article>` element.

**Why:** PageFind indexes a page as one document. `<head>` metadata applies to the whole page,
making it correct for single-post pages. For aggregate pages, per-item faceting requires
attributes on each `<article>` element. Head meta also serves standard HTML purposes (SEO,
social sharing, browser tooling) on post pages.

---

### DEC-009 — Default: expose all front matter fields as HTML metadata

**Decision:** When no `allowed_meta_fields` list is configured, every key in a document's
front matter is emitted as HTML metadata — including internal fields like `postPath`, `guid`,
and `dateModified`.

**Why:** Front matter in published Markdown posts is already public. Defaulting to all fields
avoids silent omissions. Fields like `postPath` and `guid` are useful to downstream tooling.
Users who need to suppress specific fields can provide `allowed_meta_fields`.

---

### DEC-010 — `allowed_meta_fields` in Generator YAML constrains post/page metadata output

**Decision:** The `Generator` struct gains an `AllowedMetaFields []string` field
(`allowed_meta_fields` in YAML). When the list is non-empty, only the named keys are emitted
as HTML metadata for posts and pages.

**Why:** Some workflows require that only certain front matter fields are public-facing
(e.g., a collection that uses private `internal_review` flags in front matter). An explicit
allowlist respects that without changing the default-all behaviour.

**This constraint applies to posts and pages only.** Feed item metadata from harvested feeds is
always rendered in full — it is already public data.

---

### DEC-011 — Front matter `title` replaces `gen.Title` in the HTML `<title>` element

**Decision:** When a post or page has a `title` key in its front matter, that value is written
to the `<title>` element instead of `gen.Title`. `gen.Title` is used only when no front matter
title is present.

**Why:** `gen.Title` is the collection-level title (e.g., "My Blog"). A post titled
"Mostly Oberon" should produce `<title>Mostly Oberon</title>`, not `<title>My Blog</title>`.
The feature request states this intent; it also improves SEO and social-sharing previews.

**Implementation note:** The title decision must be made before writing `<head>` to avoid
emitting two `<title>` elements.

---

### DEC-012 — Multi-value front matter fields each emit a separate `<meta>` element

**Decision:** For front matter fields whose value is a YAML sequence (e.g., `keywords`,
`series`), each value produces its own `<meta>` element rather than a joined string.

**Why:** PageFind's attribute-extraction filter syntax
(`data-pagefind-filter="keywords[content]"`) reads a single `content` attribute value as one
filter value. To register multiple filter values for the same key, multiple `<meta>` elements
are needed. This also keeps individual values unambiguous (no comma-splitting needed).

**Implementation:** New helper `GetAttributeStringSlice(key string) []string` in `cmarkdoc.go`,
handling both `string` and `[]interface{}` YAML types.

---

### DEC-013 — `WriteHtmlPage` gains a `frontMatter` parameter; all call sites updated

**Decision:** `WriteHtmlPage` signature changes from:
```
func (gen *Generator) WriteHtmlPage(htmlName, link, postPath, pubDate, innerHTML string) error
```
to:
```
func (gen *Generator) WriteHtmlPage(htmlName, link, postPath, pubDate, innerHTML string, frontMatter map[string]interface{}) error
```

Call sites that must be updated:

| Location | Notes |
|---|---|
| `generator.go` — `GeneratePosts` | Also fix the `Parse()` bug (DEC-001) |
| `schema.go` — `Post()` | `doc.FrontMatter` already available after `LoadCommonMark` |
| `page.go` — `Page()` | `doc.FrontMatter` already available; no Parse bug here |

`GeneratePages` delegates to `Page()` and requires no separate change.

`WriteHTML` (aggregate collection page) passes `nil` as front matter — collection-level pages
have no single-post front matter.

---

### DEC-014 — Pages are treated identically to posts for metadata rendering

**Decision:** The `antenna page` command and the `GeneratePages` regeneration path expose front
matter fields the same way as `antenna post`: all fields by default, constrained by
`allowed_meta_fields` when set, with DC-prefixed keys and multi-value slice support.

**Why:** Pages are authored Markdown documents with front matter, just like posts. Inconsistent
metadata rendering between pages and posts would surprise authors and break site-wide faceted
search.

---

### DEC-015 — `series` and `seriesNumber` are valid public metadata fields

**Decision:** `series` and `seriesNumber` front matter fields are treated as public metadata,
rendered into HTML `<head>` as `<meta>` and `data-pagefind-filter` elements by default. They
represent periodical metadata (series, volume, number) analogous to `keywords` and `subject`.

**Why:** These fields are already in active use in the author's post front matter. They carry
the same public-facing intent as `keywords`. PageFind filter key names: `series` and
`seriesNumber`.

---

## 2026-06-27 — Improved HTML accessibility and ARIA support

### Context

An accessibility audit of generated HTML identified the following issues across the three page
types (`WriteHTML` aggregate, `WriteHtmlPage` posts/pages):

- No skip navigation link (WCAG 2.4.1 Level A)
- `<address>` misused as a wrapper for feed item source links
- `rel="altenate"` typo on the Markdown alternate link element
- `lang="en-US"` hardcoded, not configurable
- Dates inside `<h2>` heading text in `WriteItem`, no `<time>` element
- No warning when aggregate page has no visible `<h1>`

---

### DEC-016 — Emit a skip navigation link with styling in the default CSS

**Decision:** Both `WriteHTML` and `WriteHtmlPage` emit `<a href="#main-content" class="skip-link">Skip to main content</a>` as the first child of `<body>`, before any `<header>` or `<nav>`. The default CSS Antenna App provides includes a rule that hides the link off-screen until focused (`:focus` makes it visible), so it works without user configuration.

**Why:** WCAG 2.4.1 (Level A) — keyboard users must be able to bypass repeated navigation blocks. The `<main id="main-content">` target was already present; only the link was missing. Bundling the required CSS rule into the default Antenna App CSS guarantees out-of-the-box functionality; authors who use custom CSS must add the rule themselves (documented).

**Alternatives rejected:** Inline style on the `<a>` element — would apply even when the author overrides default CSS, causing visual conflicts.

---

### DEC-017 — Replace `<address>` with `<footer>` inside feed item `<article>` elements

**Decision:** In `WriteItem`, the source link block changes from `<address><a href="…">…</a></address>` to `<footer><a href="…">…</a></footer>`. The `<footer>` is a child of the enclosing `<article>`.

**Why:** HTML5 defines `<address>` as contact information for the author of the document or its nearest sectioning ancestor — not for arbitrary resource links. Screen readers announce `<address>` as author contact information, which is misleading when the content is a feed item URL. `<footer>` inside `<article>` is semantically correct for supplementary information about the article (source, attribution).

**Impact on CSS:** Any CSS rule targeting `article address` in user stylesheets must be updated to `article footer`. This must be noted in the website documentation for the aggregation list feature.

**Alternatives rejected:** `<p class="item-source">` — avoids semantic conflict but provides no landmark semantics. `<cite>` — correct for citing the title of a work but not for a clickable URL block.

---

### DEC-018 — Add configurable `lang` field to the Generator struct

**Decision:** Add `Lang string` to the `Generator` struct, with a default of `"en-US"`. The generator YAML supports a `lang:` key. Both `WriteHTML` and `WriteHtmlPage` write `<html lang="%s">` using `gen.Lang`.

**Why:** `lang="en-US"` is hardcoded in the current generator, which is wrong for multilingual content and for sites written in languages other than US English. Screen readers and translation services use this attribute. The default preserves existing behaviour for sites that do not set the key.

---

### DEC-019 — Move dates out of `<h2>` and use `<time>` element in `WriteItem`

**Decision:** In `WriteItem`, the date is removed from the `<h2>` title string and emitted as a separate `<time datetime="YYYY-MM-DD">YYYY-MM-DD</time>` element (in a `<p>`) immediately below the heading, inside the `<article>` but outside the heading.

**Why:** Including the date in the heading's text makes the accessible name of the article noisy for screen reader users (e.g. "Post Title date 2020-04-11" is announced as the heading). The `<time>` element with a `datetime` attribute is machine-readable and meaningful to search engines and AT. The date is supplementary metadata, not part of the article's title.

**Impact on CSS:** Any CSS rule targeting `article h2` that also styles the date as part of the heading must be updated. This must be noted in documentation. The default CSS will account for the new structure.

---

### DEC-020 — Warn to stderr when aggregate page has no visible `<h1>`

**Decision:** When `WriteHTML` is called and `gen.Header` is empty, emit a warning to stderr: `warning: aggregate page has no <h1>; set a 'header' value in the generator YAML`. No automatic `<h1>` is injected.

**Why:** Injecting an `<h1>` from `gen.Title` would work for well-configured generators but would be surprising for authors who intentionally omit a visible heading. A warning preserves author control while prompting them to explicitly address the issue. WCAG 2.4.6 (Level AA) requires headings to describe content, so the absence deserves visibility.

**Alternatives rejected:** Auto-inject `<h1>gen.Title</h1>` — silently modifies page structure; could duplicate a heading that already appears in `gen.Header` content.

---

### DEC-021 — Fix `rel="altenate"` typo

**Decision:** Fix the typo on the Markdown source alternate link: `"altenate"` → `"alternate"` (html.go, `writeHeadElement`).

**Why:** This is a bug. The `rel` attribute value is not a valid link relationship. Browsers and feed discovery tools silently ignore unknown `rel` values, so the Markdown alternate link is never discoverable.

---

## 2026-07-19 — Feed item formatting control (`items:` block)

### Context

`item_formatting_proposal.md` proposed a new `items:` block in `page.yaml` giving
per-collection control over how harvested feed items render into aggregate pages,
distinct from `allowed_meta_fields` (posts/pages only, DEC-010). Reading `html.go`,
`harvest.go`, and `cmarkdoc.go` during design review surfaced that some of the
proposal's stated goals — preferring Markdown content over raw feed HTML, and a
safe/unsafe rendering split — are **already implemented** in `WriteItem` and
`CommonMark`, just undeclared. This session's decisions formalize that existing
behavior and add the genuinely new configuration surface around it.

---

### DEC-022 — Add `items:` block to `page.yaml`, scoped to feed-item body content only

**Decision:** `page.yaml` gains an `items:` key (`Generator.Items ItemsConfig`)
controlling the human-visible body of harvested feed items (title, date, content,
link, source) in aggregate collection pages.

**Why:** No equivalent control exists today; `allowed_meta_fields` only governs
posts/pages. This proposal's scope is deliberately narrow: it does not touch
`allowed_meta_fields`, does not touch local post/page rendering (DEC-031), and does
not touch the PageFind `data-pagefind-filter` attributes already governed by
DEC-007 through DEC-010 (DEC-025).

---

### DEC-023 — `sourceMarkdown` is preferred over raw `description`, declared explicitly

**Decision:** The existing precedence in `WriteItem` — use `sourceMarkdown` when
non-empty, else raw `description` — is formalized as `items:`'s foundational,
documented content-resolution rule rather than left as an undeclared fallback
inside one `if` statement.

**Why:** Reading `harvest.go` (`saveItem`) confirmed `sourceMarkdown` is already
auto-populated for nearly every harvested item, by converting the feed's HTML
`description` via `html2md`, unless a `source:markdown` feed extension is present
(untested — marked `FIXME` in source) or the conversion fails. This is not new
behavior; making it explicit lets `items.html` (DEC-024) be specified correctly
against two distinct pipelines instead of one assumed pipeline.

---

### DEC-024 — `items.html` has dual semantics depending on which content pipeline resolved

**Decision:** `items.html` (`strip` \| `escape` \| `unsafe`, default `strip`) means:

- When `sourceMarkdown` resolved (the common case): `strip`/`escape` keep the
  existing `CommonMark.ToHTML()` goldmark-safe renderer (no raw HTML passthrough);
  `unsafe` switches that item's rendering to `CommonMark.ToUnsafeHTML()` — the same
  renderer already used for local posts — as an explicit, deliberate per-collection
  opt-in.
- When raw `description` resolved (fallback only — no `sourceMarkdown` for that
  item): `strip` (default) strips HTML tags to plain text; `escape` HTML-escapes
  the raw text; `unsafe` passes it through unchanged.

**Why:** `WriteItem` currently interpolates raw `description` via `fmt.Fprintf`
with `%s` — no escaping at all — whenever `sourceMarkdown` is empty. That is a
real, unsanitized-HTML gap. The Markdown-path default is **not** a behavior
change (goldmark safe mode is already what runs today); only the raw-description-
fallback default changes existing behavior, and only for the minority of items
that lack `sourceMarkdown`.

**Alternatives rejected:** A single flat `html` setting applied uniformly
regardless of which pipeline resolved — rejected because the two pipelines have
different current trust levels (one already goldmark-safe, one entirely
unsanitized) and conflating them would either weaken the safe path or fail to fix
the actually-unsafe path.

---

### DEC-025 — `items.fields` governs body content only, not PageFind filter attributes

**Decision:** `items.fields` (ordered allowlist: `title`, `link`, `pubDate`,
`content`, `source`) controls only the visible article body built in `WriteItem`.
It has no effect on the `data-pagefind-filter` attribute values (categories,
`dc_*` fields, author, label, channel) built separately in `WriteItem`'s `filters`
slice.

**Why:** DEC-010 already established that feed item metadata is always rendered
in full for PageFind faceting — it is already public data. `items.fields` is a
new, separate axis (what's visible in the rendered article) and must not
silently collapse into or override that earlier decision.

---

### DEC-026 — `items.link.label_field` defaults to `static` with fallback `"Continue reading"` — an accessibility-motivated default change

**Decision:** `items.link.label_field` defaults to `static` (not `link`), and
`items.link.label_fallback` defaults to `"Continue reading"` (not `"Read
more"`). A new literal value for `label_field`, `static`, makes
`label_fallback` the anchor text unconditionally for every item, ignoring
per-item data entirely. Setting `label_field` to `link` restores the
pre-existing behavior (anchor text is the item's URL); setting it to `title`
uses the item title instead. Both are explicit, called-out opt-outs from the
new default, not the default itself.

**Why:** The current, unconfigured behavior — anchor text is the raw item
URL (`html.go:164-166`) — is read aloud character-by-character by screen
readers, a poor listening experience for content meant to be browsed by ear
as much as by eye. Defaulting to a fixed, human-readable call-to-action label
instead is consistent with this project's existing accessibility decisions
(DEC-016 skip-navigation link, DEC-019 `<time>` element instead of a
heading-embedded date, DEC-020 missing-`<h1>` warning). This is a deliberate,
visible behavior change to every existing collection's rendered output —
unlike the narrower `html: strip` fallback-path default (DEC-024), which only
affects items lacking `sourceMarkdown` — and must be prominently called out
in release notes, not treated as a minor detail.

A separate, narrower need motivated adding the `static` mechanism in the
first place: `label_fallback`'s original design ("used when the named field
is empty") can never fire in practice against the field `link`, since a
rendered item's link is essentially always present. Sites wanting a fixed
call-to-action label regardless of item data — the motivating real-world
case, `rsdoiel.github.io/antenna`, wanting "read me" instead of a URL — need
a direct way to say that, not a fallback path that never triggers. Making
`static` the *default* (rather than leaving `link` as default and `static` as
an opt-in) folds the accessibility fix and this mechanism together: every
site gets the accessible default for free, and any site wanting a different
fixed label only needs to set `label_fallback`.

**Alternatives rejected:** Keeping `label_field: link` as the default and
treating the accessible label as opt-in — rejected because it would leave
the accessibility problem present by default for every collection that
doesn't know to configure it, the same objection already settled against for
DEC-016/DEC-019/DEC-020's opt-out-not-opt-in defaults.

---

### DEC-027 — `items.link.missing` recursively falls back to `unlinked`

**Decision:** `items.link.missing: source_link` uses the parent channel's URL
when an item's own `link` is empty. If the channel URL is *also* empty, behavior
falls back to `unlinked` (plain text, no `<a>`) rather than emitting an empty
`href`.

**Why:** `WriteItem` has no missing-link handling today at all — an empty `link`
currently renders `<a href="">`. This proposal introduces the first defined
behavior for that case; the recursive fallback closes an otherwise-undefined
edge case (source_link's own source being unavailable) rather than leaving it to
produce broken markup.

---

### DEC-028 — `items.date_format` parses against the known stored layout, falls back to the existing 10-character truncation on failure

**Decision:** `items.date_format` (Go reference layout, default `"2006-01-02"`)
is applied by parsing the stored `pubDate`/`updated` value against
`"2006-01-02 15:04:05"` — the layout `saveItem` (`harvest.go`) writes when
`gofeed` successfully parses the feed's date — then reformatting. When the
stored value does not match that layout (the feed's raw, unparsed date
string, format unknown, preserved when `gofeed` failed to parse it), the
result is truncated to the first 10 characters (or returned as-is if
shorter) — i.e. exactly what current `WriteItem` code already does
unconditionally — rather than erroring `generate` or returning the full,
untruncated raw string.

**Why:** Current code only string-slices the first 10 characters; it never
parses. A real `date_format` requires `time.Parse`, and there is no
general-purpose parser for arbitrary feed date formats already in the
codebase, so a fallback is required for the parse-failure case. An earlier
version of this decision proposed returning the full raw string unchanged in
that case; on review this was rejected — it would silently change output
length for any item whose date `gofeed` couldn't parse, which is exactly the
kind of unflagged behavior change this design otherwise goes out of its way
to avoid (contrast DEC-024, DEC-026, which change behavior deliberately and
say so). Matching the existing truncation exactly means this fallback path
needs no release-note callout at all — collections without an `items:` block
get byte-identical date rendering to today, full stop.

---

### DEC-029 — `items.content_max_length` truncates pre-render source text, never rendered HTML

**Decision:** `items.content_max_length` truncates whichever pre-render source
text resolved (`sourceMarkdown` or raw `description`, per DEC-023), on a word
boundary, before Markdown/HTML conversion.

**Why:** Truncating already-rendered HTML risks cutting mid-tag and producing
unbalanced markup. Truncating the source text first, then rendering, avoids that
class of bug entirely.

---

### DEC-030 — Theme directories may include `items.yaml`, applied like `head.yaml`'s field-copy pattern but wrapped under one `items:` key

**Decision:** A theme directory may include `items.yaml`. `ApplyTheme` gains a
new `updateItemsElement(gen *Generator, themeName string) (bool, error)`,
following `updateHeadElements`'s existing read-and-overwrite-if-present pattern
(unmarshal into a throwaway struct, copy over if present — no deep merge). Unlike
`head.yaml`, whose own top-level keys (`meta`, `link`, `script`, `style`) copy
flat onto `Generator`'s top-level fields, `items.yaml`'s entire content is
unmarshaled into one `ItemsConfig` and assigned to `Generator.Items` as a single
nested block. `isTargetFile` (`themes.go`) gains `"items.yaml"` so theme
directories defining only this file are still discovered by
`findDirectoriesWithTargetFiles`.

**Why:** `items.yaml`'s own `link:` sub-key would collide with `head.yaml`'s
existing top-level `link:` (the page's `<link>` elements) if flat-merged the same
way — nesting under one wrapper key avoids that collision, matching the
`header.md`/`nav.md`/`footer.md` one-file-to-one-key convention instead.

---

### DEC-031 — `items:` configuration has zero effect on local post/page rendering

**Decision:** No part of `items:` configuration reaches `GeneratePosts`,
`Page()`, `Post()`, or their use of `CommonMark.ToUnsafeHTML()`. Local posts and
pages render exactly as they do today, unconditionally.

**Why:** Explicitly stated and verified against current call sites
(`generator.go`, `page.go`, `schema.go`) so implementation does not accidentally
thread `ItemsConfig` into the unrelated, unsafe-mode post/page rendering path.

---

## 2026-07-19 — Phase 7 smoke-test findings (implementation-time corrections)

Two real defects surfaced only by running the built `antenna` binary end-to-end
against a live SQLite database, not by unit tests alone. Both are recorded here
because they correct or extend DEC-028/the general implementation, not because
they change any documented config surface.

### DEC-032 — `formatItemDate` must also parse RFC3339, not only the space-separated layout

**Decision:** `formatItemDate` (DEC-028) tries `time.RFC3339` before falling
back to `"2006-01-02 15:04:05"`, then to the legacy 10-character truncation.

**Why:** The production driver used everywhere in `antenna generate`/`harvest`
(`sql.Open("sqlite", ...)`, backed by `github.com/glebarez/go-sqlite`, a
pure-Go SQLite driver) auto-converts `DATETIME`-affinity columns to RFC3339
(`2026-07-19T00:00:00Z`) when scanning into a Go `string` — regardless of the
space-separated layout `saveItem` (`harvest.go`) originally wrote as a bound
parameter. DEC-028's original assumption only covered the write-side format
and was never checked against what the driver actually returns on read. This
was caught by an end-to-end smoke test (`antenna init` → seed a real
`pages.db` → `antenna generate` → inspect `pages.html`): `date_format` had
*zero effect* before this fix, silently. The project's own test suite did not
catch this because `html_test.go`'s in-memory test databases use a different
driver (`mattn/go-sqlite3`, registered as `"sqlite3"`), which returns
whatever string was stored verbatim rather than reinterpreting it — the two
drivers disagree on this, and only the production one matters here.

### DEC-033 — Excluded `items.fields` sections omit their wrapper element entirely

**Decision:** When `items.fields` excludes `pubDate` or `content`, `WriteItem`
omits that section's `<p>...</p>` element from the rendered article body
entirely, rather than emitting an empty `<p></p>`.

**Why:** The original implementation kept a fixed two-slot `<p>%s</p>`
template and simply left the inner value blank when a field was excluded,
which produced stray empty paragraph elements — found via the same live
smoke test (`items: {fields: [title, content]}` against a real generated
page showed a bare `<p></p>` where the excluded `pubDate` section had been).
Building the article body from a list of only the present sections, joined
together, avoids emitting markup for a section the configuration says
shouldn't appear at all.

---
