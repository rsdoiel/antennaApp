# Design: Enhanced Metadata Processing

Covers the implementation of decisions DEC-001 through DEC-015 from `decisions.md`.
See also: `enhanced_front_matter_processing_feature_request.md` (original request).

---

## Scope

Three related improvements implemented together:

1. **Bug fix** — `GeneratePosts` renders YAML front matter as visible body text. Fix by calling `doc.Parse()` before converting to HTML (DEC-001).
2. **Schema extension** — add `categories` column to `items` to preserve feed item category data that is currently discarded on harvest (DEC-002).
3. **Metadata rendering** — emit front matter and feed item metadata as standard HTML `<meta>` elements and PageFind 1.5+ filter attributes (DEC-008 through DEC-015).

---

## Two rendering paths

The choice of rendering path is determined by the HTML output type (DEC-008):

```
Source document
  │
  ├─ Post or Page (one HTML file = one document)
  │     Front matter fields → <meta> elements in <head>
  │     data-pagefind-filter on each <meta> element
  │
  └─ Aggregate feed page (one HTML file = many items)
        Per-item metadata → data-pagefind-filter on each <article> element
        (head metadata applies to the whole page, not individual items)
```

---

## Schema changes

### `items` table — new `categories` column

Add to `SQLCreateTables` in `sql_stmts.go`:

```sql
CREATE TABLE IF NOT EXISTS items (
    link PRIMARY KEY,
    postPath TEXT DEFAULT '',
    title TEXT,
    description TEXT,
    authors JSON,
    enclosures JSON DEFAULT '',
    guid TEXT,
    pubDate DATETIME,
    dcExt JSON,
    channel TEXT,
    sourceMarkdown TEXT DEFAULT '',
    status TEXT DEFAULT '',
    label TEXT DEFAULT '',
    updated DATETIME,
    categories JSON DEFAULT ''        -- NEW
);
```

Update `SQLUpdateItem` to include `categories` in both INSERT and ON CONFLICT UPDATE.

Update `SQLDisplayItems` to SELECT `categories` so `WriteHTML` can pass it to `WriteItem`.

Note: the broader schema upgrade strategy (`sql_upgrade.go`, `antenna upgrade` command) is
tracked in `development_notes.md`. The migration for this column follows the same pattern
using `pragma_table_info` to guard the ALTER statement.

---

## New component: `GetAttributeStringSlice` in `cmarkdoc.go`

A helper alongside the existing `GetAttributeString` and `GetAttributeBool`:

```go
func (doc *CommonMark) GetAttributeStringSlice(key string) []string
```

Behaviour:
- Key absent → return nil
- Value is a `string` → return `[]string{value}` (or nil if empty)
- Value is `[]interface{}` → return each element cast to string, skipping empty values
- Any other type → return nil

Used by `writeHeadElement` to iterate multi-value front matter fields such as `keywords`,
`series`, and `categories`.

---

## Generator struct addition

In `generator.go`, add to the `Generator` struct (DEC-010):

```go
// AllowedMetaFields, when non-empty, limits which front matter keys are
// emitted as HTML metadata for posts and pages. When empty, all keys
// are emitted (default).
AllowedMetaFields []string `json:"allowed_meta_fields,omitempty" yaml:"allowed_meta_fields,omitempty"`
```

Example in a generator YAML file:

```yaml
allowed_meta_fields:
  - title
  - author
  - keywords
  - series
  - seriesNumber
  - datePublished
  - abstract
```

When this list is absent or empty every front matter key is emitted, including `postPath`,
`guid`, `dateModified`, and any custom fields (DEC-009).

---

## Modified: `writeHeadElement` in `html.go`

### Signature change

```go
// before
func (gen *Generator) writeHeadElement(out io.Writer, postPath string)

// after
func (gen *Generator) writeHeadElement(out io.Writer, postPath string, frontMatter map[string]interface{})
```

### Title logic (DEC-011)

The `<title>` element is written once. Priority:

1. `frontMatter["title"]` (string, non-empty) — use this.
2. `gen.Title` (non-empty) — use this.
3. Neither set — no `<title>` element emitted.

This replaces the current unconditional `gen.Title` write.

### Front matter metadata block

After all existing `<meta>`, `<link>`, and `<script>` elements, emit front matter metadata
when `frontMatter` is non-nil:

```
for each key in frontMatter:
    if AllowedMetaFields is non-empty and key not in AllowedMetaFields → skip
    if key == "title" → skip (already handled in <title> above)

    values = GetAttributeStringSlice(key)
    if values is nil:
        values = [string representation of value] if value is a non-empty string
    if values is empty → skip

    for each value in values:
        emit: <meta name="KEY" content="VALUE">
              <meta data-pagefind-filter="KEY[content]" content="VALUE">
```

Both `<meta>` elements share the same `content` attribute value; the `data-pagefind-filter`
attribute instructs PageFind to read from `content` as the filter value.

### Callers after change

| Caller | Argument passed |
|---|---|
| `WriteHTML` — aggregate page | `nil` — no single-post front matter |
| `WriteHtmlPage` | `frontMatter` parameter (forwarded from caller) |

---

## Modified: `WriteHtmlPage` in `html.go`

### Signature change

```go
// before
func (gen *Generator) WriteHtmlPage(htmlName, link, postPath, pubDate, innerHTML string) error

// after
func (gen *Generator) WriteHtmlPage(htmlName, link, postPath, pubDate, innerHTML string,
    frontMatter map[string]interface{}) error
```

`frontMatter` is forwarded to `writeHeadElement`. No other change to the function body.

### Call sites

| File | Function | Change |
|---|---|---|
| `generator.go` | `GeneratePosts` | Fix `doc.Parse()` bug; pass `doc.FrontMatter` |
| `schema.go` | `Post()` | Pass `doc.FrontMatter` (already available) |
| `page.go` | `Page()` | Pass `doc.FrontMatter` (already available; no Parse bug) |

`GeneratePages` delegates to `Page()` — covered automatically.

---

## Modified: `GeneratePosts` in `generator.go`

Replace the bare struct literal (DEC-001):

```go
// before
doc := &CommonMark{Text: sourceMarkdown}

// after
doc := &CommonMark{}
if err := doc.Parse([]byte(sourceMarkdown)); err != nil {
    // malformed front matter — treat entire source as body text
    doc.Text = sourceMarkdown
}
```

Then update the `WriteHtmlPage` call to pass `doc.FrontMatter`:

```go
// before
if err := gen.WriteHtmlPage(htmlName, link, postPath, pubDate, innerHTML); err != nil {

// after
if err := gen.WriteHtmlPage(htmlName, link, postPath, pubDate, innerHTML, doc.FrontMatter); err != nil {
```

---

## Modified: `saveItem` in `harvest.go`

Marshal `item.Categories` and pass to the updated `updateItem`:

```go
var categories []byte
if item.Categories != nil {
    categories, err = json.Marshal(item.Categories)
    if err != nil {
        return fmt.Errorf("failed to marshal item.Categories, %s", err)
    }
}
```

Pass `string(categories)` as the new argument to `updateItem`.

---

## Modified: `updateItem` in `schema.go`

Add `categories string` parameter. Pass it through to `SQLUpdateItem`.

The `Post()` caller extracts categories from front matter:

```go
// In Post(), after doc.Parse():
var categoriesSrc []byte
if cats := doc.GetAttributeStringSlice("categories"); len(cats) > 0 {
    categoriesSrc, _ = json.Marshal(cats)
}
// pass string(categoriesSrc) to updateItem
```

---

## Modified: `WriteItem` in `html.go`

### Signature change

Add `categories string` parameter (JSON-encoded `[]string`):

```go
func (gen *Generator) WriteItem(out io.Writer, link, title, description string,
    authors []*gofeed.Person, sourceMarkdown string, enclosures []*Enclosure,
    guid, pubDate, dcExtSrc, channel, status, updated, label string,
    categories string) error   // NEW parameter
```

`dcExtSrc` is already passed but currently unused for filter rendering. Both `dcExtSrc` and
`categories` are now used to build the `data-pagefind-filter` attribute value.

### Filter attribute construction

Build a `[]string` of `"key:value"` entries, then join with `", "` as the attribute value:

```
filters = []

// Categories (RSS/Atom <category>)
unmarshal categories JSON → []string
for each cat: append "category:CAT"

// Dublin Core fields (DEC-004, DEC-005)
unmarshal dcExtSrc JSON → ext.DublinCoreExtension
for each non-empty DC field and each value in that field's []string:
    append "dc_FIELDNAME:VALUE"
    (field name lowercased: dc_creator, dc_subject, dc_date, etc.)

// Authors (native feed authors)
for each author in authors where author.Name != "":
    append "author:NAME"

// Publication date
if pubDate != "": append "datePublished:PUBDATE"

// Feed provenance (DEC-007)
if label != "": append "label:LABEL"
if channel != "": append "channel:CHANNEL"
```

Emit on `<article>` only when `filters` is non-empty:

```html
<!-- with filters -->
<article data-published="2020-04-11" data-link="https://…"
         data-pagefind-filter="category:Oberon, dc_subject:Programming Languages, author:R. S. Doiel, datePublished:2020-04-11, label:My Feed, channel:https://example.com/feed.xml">

<!-- without filters (no change from current output) -->
<article data-published="" data-link="https://…">
```

### Caller update in `WriteHTML`

`WriteHTML` scans `SQLDisplayItems` which now includes `categories`. The scan and `WriteItem`
call gain the `categories` column value.

---

## HTML output examples

### Standalone post page — `<head>` section

Front matter:
```yaml
title: Mostly Oberon
author: R. S. Doiel
datePublished: '2020-04-11'
keywords:
  - Oberon
  - programming
series:
  - mostly-oberon
seriesNumber: 1
abstract: An introduction to the Oberon programming language.
postPath: posts/2020/04/11/mostly-oberon.md
```

Generated `<head>` (excerpt — existing charset/viewport/generator meta omitted):

```html
<title>Mostly Oberon</title>
<meta name="author" content="R. S. Doiel">
<meta data-pagefind-filter="author[content]" content="R. S. Doiel">
<meta name="datePublished" content="2020-04-11">
<meta data-pagefind-filter="datePublished[content]" content="2020-04-11">
<meta name="keywords" content="Oberon">
<meta data-pagefind-filter="keywords[content]" content="Oberon">
<meta name="keywords" content="programming">
<meta data-pagefind-filter="keywords[content]" content="programming">
<meta name="series" content="mostly-oberon">
<meta data-pagefind-filter="series[content]" content="mostly-oberon">
<meta name="seriesNumber" content="1">
<meta data-pagefind-filter="seriesNumber[content]" content="1">
<meta name="abstract" content="An introduction to the Oberon programming language.">
<meta data-pagefind-filter="abstract[content]" content="An introduction to the Oberon programming language.">
<meta name="postPath" content="posts/2020/04/11/mostly-oberon.md">
<meta data-pagefind-filter="postPath[content]" content="posts/2020/04/11/mostly-oberon.md">
```

### Aggregate page — `<article>` element

Feed item with categories, DC extension, and authors:

```html
<article data-published="2020-04-11" data-link="https://example.com/item"
         data-pagefind-filter="category:Oberon, category:programming, dc_creator:R. S. Doiel, dc_subject:Programming Languages, author:R. S. Doiel, datePublished:2020-04-11, label:My Feed, channel:https://example.com/feed.xml">
  <h2>Mostly Oberon</h2>
  …
</article>
```

---

## Migration scripts

Two scripts to add the `categories` column to existing databases.
Both guard against re-running on an already-migrated database.

### `migrate_categories.bash`

```bash
#!/usr/bin/env bash
# Adds the categories column to the items table in an Antenna App database.
# Usage: migrate_categories.bash PATH_TO_DATABASE.db
set -euo pipefail

DB="${1:?usage: $0 PATH_TO_DATABASE.db}"

EXISTS=$(sqlite3 "$DB" \
  "SELECT COUNT(*) FROM pragma_table_info('items') WHERE name='categories';")

if [ "$EXISTS" -eq "0" ]; then
    sqlite3 "$DB" "ALTER TABLE items ADD COLUMN categories JSON DEFAULT '';"
    echo "Migrated: $DB"
else
    echo "Already migrated: $DB"
fi
```

### `migrate_categories.ps1`

```powershell
# Adds the categories column to the items table in an Antenna App database.
# Usage: .\migrate_categories.ps1 -DB PATH_TO_DATABASE.db
param([Parameter(Mandatory)][string]$DB)

$exists = sqlite3 $DB `
    "SELECT COUNT(*) FROM pragma_table_info('items') WHERE name='categories';"

if ($exists -eq "0") {
    sqlite3 $DB "ALTER TABLE items ADD COLUMN categories JSON DEFAULT '';"
    Write-Host "Migrated: $DB"
} else {
    Write-Host "Already migrated: $DB"
}
```

Both scripts must be documented in the Antenna App website under a "Schema upgrades" section,
alongside the existing `sourceMarkdown` / `postPath` migration notes in `development_notes.md`.

---

## Files changed summary

| File | Nature of change |
|---|---|
| `sql_stmts.go` | Add `categories` to `SQLCreateTables`, `SQLUpdateItem`, `SQLDisplayItems` |
| `harvest.go` | Marshal `item.Categories`; pass to `saveItem` / `updateItem` |
| `schema.go` | Add `categories` param to `updateItem`; extract from front matter in `Post()`; update `WriteHtmlPage` call |
| `generator.go` | Add `AllowedMetaFields` to `Generator` struct; fix `GeneratePosts` Parse bug; update `WriteHtmlPage` call |
| `cmarkdoc.go` | Add `GetAttributeStringSlice` |
| `html.go` | Update `writeHeadElement` signature and body; update `WriteHtmlPage` signature; update `WriteItem` signature and body; update `WriteHTML` scan and call |
| `page.go` | Update `WriteHtmlPage` call to pass `doc.FrontMatter` |
| `migrate_categories.bash` | New — migration script for Linux/macOS |
| `migrate_categories.ps1` | New — migration script for Windows |
