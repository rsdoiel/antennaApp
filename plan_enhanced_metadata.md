# Implementation Plan: Enhanced Metadata Processing

Implements the design in `design_enhanced_metadata.md`.
Decisions referenced throughout are in `decisions.md`.

Tests are written before implementation in every phase (TDD).
Each task is independent enough to commit separately.
Tasks within a phase may be done in parallel; phases must be done in order.

---

## Phase 1 — Foundation (no breaking changes)

These tasks touch only new code or additive struct fields. Nothing breaks until Phase 4
wires the new signatures into callers.

### Task 1.1 — Test `GetAttributeStringSlice` (new file)

Create `cmarkdoc_test.go`. Write `TestGetAttributeStringSlice` covering:

| Case | Input | Expected |
|---|---|---|
| Key absent | `FrontMatter{}` | `nil` |
| Empty string value | `{"k": ""}` | `nil` |
| Non-empty string | `{"k": "Oberon"}` | `["Oberon"]` |
| Sequence of strings | `{"k": ["a","b"]}` | `["a","b"]` |
| Sequence with empty entries | `{"k": ["a","","b"]}` | `["a","b"]` |
| Non-string/non-slice type | `{"k": 42}` | `nil` |

Verify tests fail before implementation (red).

### Task 1.2 — Implement `GetAttributeStringSlice` in `cmarkdoc.go`

Add after `GetAttributeBool`. Signature:

```go
func (doc *CommonMark) GetAttributeStringSlice(key string) []string
```

Run `go test -run TestGetAttributeStringSlice` — must pass (green).

### Task 1.3 — Add `AllowedMetaFields` to `Generator` struct in `generator.go`

Add field after existing struct fields:

```go
AllowedMetaFields []string `json:"allowed_meta_fields,omitempty" yaml:"allowed_meta_fields,omitempty"`
```

Run `go build ./...` — must compile cleanly.

---

## Phase 2 — SQL layer

Additive SQL changes. Existing databases are unaffected until the migration scripts are run.
New databases created after this phase will have the `categories` column from the start.

### Task 2.1 — Add `categories` to `SQLCreateTables` in `sql_stmts.go`

Append `categories JSON DEFAULT ''` as the last column in the `items` table definition.

### Task 2.2 — Update `SQLUpdateItem` in `sql_stmts.go`

Add `categories` as parameter `?15` in both the INSERT column list and the
ON CONFLICT UPDATE SET clause.

### Task 2.3 — Update `SQLDisplayItems` in `sql_stmts.go`

Add `ifnull(categories, '') as categories` to the SELECT column list.
This is the query used by `WriteHTML` to feed `WriteItem`.

Run `go build ./...` after all three SQL tasks — must compile cleanly (no callers updated yet;
string constant changes only).

---

## Phase 3 — Data layer

Updates `updateItem` and its callers to carry the new `categories` value through to the
database. Depends on Phase 2.

### Task 3.1 — Update `updateItem` signature in `schema.go`

Add `categories string` as the last parameter. Pass it as `?15` in the `db.Exec` call.

### Task 3.2 — Update `saveItem` in `harvest.go`

Marshal `item.Categories` to JSON and pass to `updateItem`:

```go
var categories []byte
if item.Categories != nil {
    categories, err = json.Marshal(item.Categories)
    if err != nil {
        return fmt.Errorf("failed to marshal item.Categories, %s", err)
    }
}
// ... existing code ...
// add string(categories) as final arg to updateItem call
```

### Task 3.3 — Update `Post()` in `schema.go` to extract front matter categories

After `doc.Parse()`, extract `categories` from front matter for posts that carry them:

```go
var categoriesSrc []byte
if cats := doc.GetAttributeStringSlice("categories"); len(cats) > 0 {
    categoriesSrc, _ = json.Marshal(cats)
}
```

Pass `string(categoriesSrc)` as the `categories` argument to `updateItem`.

Run `go build ./...` — must compile cleanly.

---

## Phase 4 — HTML rendering

The two rendering paths. Depends on Phase 1 (for `GetAttributeStringSlice` and
`AllowedMetaFields`) but is independent of Phases 2 and 3.

### Task 4.1 — Test `writeHeadElement` metadata output (new file)

Create `html_test.go`. Write `TestWriteHeadElement` using a `bytes.Buffer` to capture output.

Cover:

| Case | Setup | Must contain | Must not contain |
|---|---|---|---|
| nil frontMatter | `frontMatter = nil` | existing charset/viewport meta | any `data-pagefind-filter` |
| String field | `{"author": "R. S. Doiel"}` | `<meta name="author"`, `data-pagefind-filter="author[content]"` | — |
| Slice field | `{"keywords": ["a","b"]}` | two `name="keywords"` elements, two `data-pagefind-filter` elements | joined string |
| Title override | `{"title": "My Post"}`, `gen.Title = "Site"` | `<title>My Post</title>` | `<title>Site</title>` |
| Title fallback | `gen.Title = "Site"`, no frontMatter title | `<title>Site</title>` | — |
| AllowedMetaFields set | `{"author":…,"postPath":…}`, allowed=`["author"]` | `name="author"` | `name="postPath"` |
| Title key excluded from meta | `{"title": "X"}` | `<title>X</title>` | `<meta name="title"` |

Verify tests fail (red).

### Task 4.2 — Update `writeHeadElement` in `html.go`

Change signature:
```go
func (gen *Generator) writeHeadElement(out io.Writer, postPath string, frontMatter map[string]interface{})
```

**Title logic** — replace the existing unconditional `gen.Title` write:

```go
pageTitle := gen.Title
if frontMatter != nil {
    if t, ok := frontMatter["title"].(string); ok && t != "" {
        pageTitle = t
    }
}
if pageTitle != "" {
    fmt.Fprintf(out, "  <title>%s</title>\n", pageTitle)
}
```

**Front matter metadata block** — after all existing `<meta>`, `<link>`, `<script>`, `<style>`
elements, add:

```go
if frontMatter != nil {
    allowed := map[string]bool{}
    for _, k := range gen.AllowedMetaFields {
        allowed[k] = true
    }
    for key, val := range frontMatter {
        if key == "title" {
            continue  // already handled in <title>
        }
        if len(allowed) > 0 && !allowed[key] {
            continue
        }
        // Build a []string of values for this key
        doc := &CommonMark{FrontMatter: map[string]interface{}{key: val}}
        values := doc.GetAttributeStringSlice(key)
        if len(values) == 0 {
            if s, ok := val.(string); ok && s != "" {
                values = []string{s}
            }
        }
        for _, v := range values {
            fmt.Fprintf(out, "  <meta name=%q content=%q>\n", key, v)
            fmt.Fprintf(out, "  <meta data-pagefind-filter=%q content=%q>\n",
                key+"[content]", v)
        }
    }
}
```

Update the two existing internal callers immediately (required to compile):
- `WriteHTML`: `gen.writeHeadElement(out, "", nil)`
- `WriteHtmlPage`: `gen.writeHeadElement(out, postPath, frontMatter)` — also requires Task 4.3 first.

Run `TestWriteHeadElement` — must pass (green).

### Task 4.3 — Test `WriteItem` filter attribute output

Add `TestWriteItem` to `html_test.go`. Capture `<article>` output with a `bytes.Buffer`.

Cover:

| Case | Input | `data-pagefind-filter` must contain |
|---|---|---|
| No metadata | empty categories, nil dcExt, no authors | attribute absent |
| Single category | `categories=["Oberon"]` | `category:Oberon` |
| Multiple categories | `categories=["a","b"]` | `category:a` and `category:b` |
| DC subject | `dcExt` with `Subject: ["S"]` | `dc_subject:S` |
| DC creator | `dcExt` with `Creator: ["C"]` | `dc_creator:C` |
| Author | `authors=[{Name:"Alice"}]` | `author:Alice` |
| Label and channel | `label="Feed"`, `channel="https://…"` | `label:Feed`, `channel:https://…` |
| pubDate | `pubDate="2020-01-01"` | `datePublished:2020-01-01` |
| All combined | all of the above | all entries present, comma-separated |

Verify tests fail (red).

### Task 4.4 — Update `WriteItem` signature in `html.go`

Add `categories string` as the final parameter.

### Task 4.5 — Implement filter attribute construction in `WriteItem`

Build the filter string inside `WriteItem`:

```go
var filters []string

// Categories
var cats []string
if categories != "" {
    _ = json.Unmarshal([]byte(categories), &cats)
}
for _, c := range cats {
    if c != "" {
        filters = append(filters, "category:"+c)
    }
}

// Dublin Core
if dcExtSrc != "" {
    var dc ext.DublinCoreExtension
    if err := json.Unmarshal([]byte(dcExtSrc), &dc); err == nil {
        dcFields := []struct {
            key    string
            values []string
        }{
            {"dc_title", dc.Title},
            {"dc_creator", dc.Creator},
            {"dc_author", dc.Author},
            {"dc_subject", dc.Subject},
            {"dc_description", dc.Description},
            {"dc_publisher", dc.Publisher},
            {"dc_contributor", dc.Contributor},
            {"dc_date", dc.Date},
            {"dc_type", dc.Type},
            {"dc_format", dc.Format},
            {"dc_identifier", dc.Identifier},
            {"dc_source", dc.Source},
            {"dc_language", dc.Language},
            {"dc_relation", dc.Relation},
            {"dc_coverage", dc.Coverage},
            {"dc_rights", dc.Rights},
        }
        for _, f := range dcFields {
            for _, v := range f.values {
                if v != "" {
                    filters = append(filters, f.key+":"+v)
                }
            }
        }
    }
}

// Authors
for _, a := range authors {
    if a != nil && a.Name != "" {
        filters = append(filters, "author:"+a.Name)
    }
}

// Publication date
if pubDate != "" {
    filters = append(filters, "datePublished:"+pubDate)
}

// Feed provenance
if label != "" {
    filters = append(filters, "label:"+label)
}
if channel != "" {
    filters = append(filters, "channel:"+channel)
}
```

Emit on `<article>`:

```go
if len(filters) > 0 {
    fmt.Fprintf(out, `
    <article data-published=%q data-link=%q data-pagefind-filter=%q>
`, pubDate, link, strings.Join(filters, ", "))
} else {
    fmt.Fprintf(out, `
    <article data-published=%q data-link=%q>
`, pubDate, link)
}
```

Run `TestWriteItem` — must pass (green).

### Task 4.6 — Update `WriteHtmlPage` signature in `html.go`

Add `frontMatter map[string]interface{}` as the final parameter. Forward to
`writeHeadElement`:

```go
func (gen *Generator) WriteHtmlPage(htmlName, link, postPath, pubDate, innerHTML string,
    frontMatter map[string]interface{}) error
```

### Task 4.7 — Update `WriteHTML` scan in `html.go`

Add `categories` variable to the scan block to match the updated `SQLDisplayItems`.
Pass it as the final argument to `WriteItem`.

Run `go build ./...` — Phase 4 must compile cleanly before proceeding.

---

## Phase 5 — Call site updates

Wires Phases 1–4 into the three action paths. Depends on Phase 4 completing cleanly.

### Task 5.1 — Fix `GeneratePosts` in `generator.go`

Replace the bare struct literal with a proper parse (DEC-001):

```go
// before
doc := &CommonMark{Text: sourceMarkdown}

// after
doc := &CommonMark{}
if err := doc.Parse([]byte(sourceMarkdown)); err != nil {
    doc.Text = sourceMarkdown
}
```

Update the `WriteHtmlPage` call:

```go
// before
if err := gen.WriteHtmlPage(htmlName, link, postPath, pubDate, innerHTML); err != nil {

// after
if err := gen.WriteHtmlPage(htmlName, link, postPath, pubDate, innerHTML, doc.FrontMatter); err != nil {
```

### Task 5.2 — Update `Post()` call in `schema.go`

```go
// before
if err := gen.WriteHtmlPage(htmlName, link, postPath, pubDate, innerHTML); err != nil {

// after
if err := gen.WriteHtmlPage(htmlName, link, postPath, pubDate, innerHTML, doc.FrontMatter); err != nil {
```

`doc.FrontMatter` is already available from the `LoadCommonMark` call earlier in `Post()`.

### Task 5.3 — Update `Page()` call in `page.go`

```go
// before
if err := gen.WriteHtmlPage(htmlName, "", postPath, "", innerHTML); err != nil {

// after
if err := gen.WriteHtmlPage(htmlName, "", postPath, "", innerHTML, doc.FrontMatter); err != nil {
```

`doc.FrontMatter` is already available from the `LoadCommonMark` call earlier in `Page()`.

Run `go build ./...` — must compile cleanly with zero errors.
Run `go test ./...` — all tests must pass.

---

## Phase 6 — Migration scripts

Independent of all code phases. Can be written at any time.

### Task 6.1 — Write `migrate_categories.bash`

Create in `antennaApp/` root. Content per `design_enhanced_metadata.md` — guard with
`pragma_table_info` check before executing the `ALTER TABLE`.

Mark executable: `chmod +x migrate_categories.bash`.

### Task 6.2 — Write `migrate_categories.ps1`

Create in `antennaApp/` root. Same guard logic using sqlite3 CLI.

### Task 6.3 — Update `development_notes.md`

Add a new section under "Upgrading schema" documenting the categories migration, consistent
with the existing `sourceMarkdown` / `postPath` notes already there.

---

## Phase 7 — Final verification

### Task 7.1 — Full test run

```bash
cd antennaApp && go test ./...
```

All tests must pass. Address any failures before proceeding.

### Task 7.2 — Build all programs

```bash
cd antennaApp && make build
```

All binaries must build cleanly.

### Task 7.3 — Smoke test: post with front matter

Using a local Antenna App workspace:

1. Create a Markdown post with front matter including `title`, `author`, `keywords` (list),
   `series`, and `abstract`.
2. Run `antenna post`.
3. Inspect the generated HTML — verify:
   - No `---` delimiters or YAML visible in the body.
   - `<title>` matches the front matter `title`, not the collection title.
   - Each keyword produces its own `<meta name="keywords">` and `<meta data-pagefind-filter>`.
   - `series` and `abstract` appear as `<meta>` elements.

### Task 7.4 — Smoke test: harvest and generate

1. Run `antenna harvest` on a collection that uses feeds with `<category>` elements.
2. Run `antenna generate`.
3. Inspect the aggregate HTML page — verify:
   - Each `<article>` with categories has a `data-pagefind-filter` attribute.
   - Filter values include `category:VALUE` entries.
   - `label:` and `channel:` entries are present.
   - Items without categories have no `data-pagefind-filter` attribute.

### Task 7.5 — Smoke test: page with front matter

1. Create a Markdown page with front matter.
2. Run `antenna page`.
3. Inspect generated HTML — same metadata checks as Task 7.3.

### Task 7.6 — Smoke test: `allowed_meta_fields`

1. Add `allowed_meta_fields: [title, author, keywords]` to a generator YAML.
2. Regenerate a post with additional front matter fields (e.g., `postPath`, `guid`).
3. Verify that only `title`, `author`, and `keywords` appear as `<meta>` elements;
   `postPath` and `guid` are absent.

---

## Task dependency summary

```
Phase 1 (1.1–1.3)
    │
    ├─── Phase 2 (2.1–2.3) ──── Phase 3 (3.1–3.3)
    │                                   │
    └─── Phase 4 (4.1–4.7) ────────────┴──── Phase 5 (5.1–5.3) ──── Phase 7
                                                                          │
Phase 6 (6.1–6.3) ────────────────────────────────────────────────────────┘
```

Phase 6 (migration scripts) is independent and can be written any time before Phase 7.
