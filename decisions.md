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
