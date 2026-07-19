# Design: Feed Item Formatting Control

Covers the implementation of decisions DEC-022 through DEC-031 from `decisions.md`.
See also: `item_formatting_proposal.md` (finalized proposal).

---

## Scope

One new configuration surface, `items:` in `page.yaml`, controlling how harvested
feed items render into aggregate collection pages. Three things this explicitly
does **not** touch:

1. Local post/page rendering (`GeneratePosts`, `Page()`, `Post()`,
   `CommonMark.ToUnsafeHTML()`) — DEC-031.
2. `allowed_meta_fields` — posts/pages metadata allowlisting is unrelated and
   unchanged.
3. PageFind `data-pagefind-filter` attribute emission on `<article>` elements
   (categories, `dc_*`, author, label, channel) — DEC-025, DEC-010 stays in
   force.

---

## Content resolution and rendering pipeline

```
Feed item body (per item)
  │
  ├─ sourceMarkdown non-empty (the common case — html2md-converted at
  │  harvest time from the feed's HTML description, or a native
  │  source:markdown extension value; see harvest.go saveItem)
  │     │
  │     ├─ items.html: strip | escape (default) → CommonMark.ToHTML()
  │     │  (goldmark safe mode — unchanged from current behavior)
  │     └─ items.html: unsafe → CommonMark.ToUnsafeHTML()
  │        (same renderer already used for local posts — explicit opt-in)
  │
  └─ sourceMarkdown empty (fallback — no description, or html2md failed)
        │
        ├─ items.html: strip (default) → tag-stripped plain text
        ├─ items.html: escape → html.EscapeString(description)
        └─ items.html: unsafe → description, unmodified (not recommended)
```

`items.content_max_length`, when set, truncates the pre-render source text
(whichever branch — `sourceMarkdown` or `description` — was selected) on a word
boundary, before this pipeline runs (DEC-029).

---

## Schema additions

### `Generator` struct (`generator.go`)

```go
// Items, when set, controls how harvested feed items render in aggregate
// collection pages. Has no effect on local post/page rendering (DEC-031).
Items ItemsConfig `json:"items,omitempty" yaml:"items,omitempty"`
```

### New types (`generator.go`, alongside `Generator`)

```go
type ItemsConfig struct {
	// Fields is an ordered allowlist of visible body fields to render.
	// Known values: "title", "link", "pubDate", "content", "source".
	// Empty means all fields (default).
	Fields []string `json:"fields,omitempty" yaml:"fields,omitempty"`

	Link LinkConfig `json:"link,omitempty" yaml:"link,omitempty"`

	// DateFormat is a Go reference-layout string applied to pubDate/updated.
	// Default: "2006-01-02".
	DateFormat string `json:"date_format,omitempty" yaml:"date_format,omitempty"`

	// ContentMaxLength truncates resolved pre-render source text on a word
	// boundary. Zero means no truncation (default).
	ContentMaxLength int `json:"content_max_length,omitempty" yaml:"content_max_length,omitempty"`

	// ShowSource controls whether the originating feed/channel label is
	// rendered. Default: true.
	ShowSource *bool `json:"show_source,omitempty" yaml:"show_source,omitempty"`

	// HTML is one of "strip" (default), "escape", "unsafe" — see DEC-024.
	HTML string `json:"html,omitempty" yaml:"html,omitempty"`
}

type LinkConfig struct {
	// LabelField names the item field supplying anchor text, or the
	// literal sentinel "static" (DEC-026). Default: "static" — a
	// deliberate, called-out change from the pre-existing behavior of
	// using the raw link URL as anchor text (screen readers read a URL
	// character-by-character, which is a poor listening experience). Set
	// to "link" to restore the previous behavior.
	LabelField string `json:"label_field,omitempty" yaml:"label_field,omitempty"`

	// LabelFallback is used when LabelField's value is empty/missing, or
	// unconditionally when LabelField == "static" (the default).
	// Default: "Continue reading".
	LabelFallback string `json:"label_fallback,omitempty" yaml:"label_fallback,omitempty"`

	// Required, if true, fails collection generation when an item's link
	// is empty, instead of applying Missing. Default: false.
	Required bool `json:"required,omitempty" yaml:"required,omitempty"`

	// Missing is one of "unlinked" (default), "omit", "source_link" — DEC-027.
	Missing string `json:"missing,omitempty" yaml:"missing,omitempty"`
}
```

`ShowSource` is `*bool` (not `bool`) so that an omitted `items:` block or an
omitted `show_source` key can be distinguished from an explicit `show_source:
false` — needed because the default is `true`, not Go's zero value.

### Defaulting

A single function, called once when `page.yaml` is loaded (alongside wherever
`AllowedMetaFields` currently gets no special defaulting, since it has no
default other than empty):

```go
func (cfg *ItemsConfig) applyDefaults() {
	if len(cfg.Fields) == 0 {
		cfg.Fields = []string{"title", "source", "pubDate", "content"}
	}
	if cfg.Link.LabelField == "" {
		cfg.Link.LabelField = "static" // DEC-026 — accessibility default, not "link"
	}
	if cfg.Link.LabelFallback == "" {
		cfg.Link.LabelFallback = "Continue reading"
	}
	if cfg.Link.Missing == "" {
		cfg.Link.Missing = "unlinked"
	}
	if cfg.DateFormat == "" {
		cfg.DateFormat = "2006-01-02"
	}
	if cfg.ShowSource == nil {
		t := true
		cfg.ShowSource = &t
	}
	if cfg.HTML == "" {
		cfg.HTML = "strip"
	}
}
```

### Validation

At the same load point, validate enums and fail fast on typos (mirrors the
project's existing config-load error style):

```go
func (cfg *ItemsConfig) validate() error {
	switch cfg.HTML {
	case "strip", "escape", "unsafe":
	default:
		return fmt.Errorf("items.html: invalid value %q (want strip, escape, or unsafe)", cfg.HTML)
	}
	switch cfg.Link.Missing {
	case "unlinked", "omit", "source_link":
	default:
		return fmt.Errorf("items.link.missing: invalid value %q (want unlinked, omit, or source_link)", cfg.Link.Missing)
	}
	return nil
}
```

---

## New components (`html.go`)

### `resolveItemContent`

```go
func resolveItemContent(description, sourceMarkdown string, cfg ItemsConfig) (string, error)
```

1. `source, usedMarkdown := sourceMarkdown, true`; if `sourceMarkdown == ""`,
   `source, usedMarkdown = description, false`.
2. If `cfg.ContentMaxLength > 0`, `source = truncateWords(source,
   cfg.ContentMaxLength)`.
3. If `usedMarkdown`:
   - `doc := &CommonMark{Text: source}`
   - `cfg.HTML == "unsafe"` → `doc.ToUnsafeHTML()`
   - else → `doc.ToHTML()` (covers both `strip` and `escape` — goldmark safe
     mode already drops/escapes embedded raw HTML; there is no meaningful
     difference between the two on this branch, both just mean "don't use
     unsafe mode")
4. Else (raw description fallback):
   - `cfg.HTML == "strip"` (default) → `stripTags(source)`
   - `cfg.HTML == "escape"` → `html.EscapeString(source)`
   - `cfg.HTML == "unsafe"` → `source`, unchanged

`WriteItem` replaces its current inline block:
```go
content := description
if sourceMarkdown != "" {
    doc := &CommonMark{Text: sourceMarkdown}
    if src, err := doc.ToHTML(); err == nil {
        content = src
    }
}
```
with a single call to `resolveItemContent`.

### `truncateWords`

```go
func truncateWords(s string, maxLen int) string
```

If `len(s) <= maxLen`, return `s` unchanged. Otherwise slice to `maxLen`, then
back off to the last preceding whitespace boundary (never cut mid-word).
Operates on the source text (Markdown or raw description), before any
HTML/Markdown conversion — DEC-029.

### `stripTags`

```go
func stripTags(s string) string
```

A small hand-rolled tag stripper (regexp `<[^>]*>` removal), used only on the
raw-description-fallback path. This is a deliberate simplification, not a full
HTML parser — acceptable here because it only ever runs on the minority
fallback path (no `sourceMarkdown` available), and the alternative
(introducing a full HTML-sanitization dependency) is disproportionate to how
rarely this branch executes. Document this limitation in the function's
comment.

### `formatItemDate`

```go
func formatItemDate(raw string, layout string) string
```

```go
const storedLayout = "2006-01-02 15:04:05" // matches harvest.go saveItem

func formatItemDate(raw string, layout string) string {
	t, err := time.Parse(storedLayout, raw)
	if err != nil {
		// DEC-028 — unparseable (raw feed date string, format unknown).
		// Match today's WriteItem truncation exactly rather than returning
		// the full raw string, so this fallback stays byte-identical to
		// current output instead of becoming an unflagged exception.
		if len(raw) > 10 {
			return raw[:10]
		}
		return raw
	}
	return t.Format(layout)
}
```

Replaces the current string-slice-to-10-characters logic in `WriteItem` for
both `pubDate` and `updated`. Note the truncation logic isn't removed — it's
preserved as the parse-failure fallback, not superseded by `time.Parse`.

### `resolveItemLink`

```go
type LinkResolution struct {
	Href        string
	Label       string
	AsPlainText bool // true: render Label as text, no <a> (Missing == "unlinked")
	Omit        bool // true: exclude the item entirely (Missing == "omit")
}

func resolveItemLink(link, title, channelURL string, cfg LinkConfig) (LinkResolution, error)
```

Logic:
1. Compute `label`:
   - `cfg.Link.LabelField == "static"` → `label = cfg.LabelFallback`
   - else look up the named field's value (`link` or `title` are the only
     fields that make sense here; `source`/`pubDate`/`content` are rejected
     at `validate()` time as invalid `label_field` values) → if empty, `label
     = cfg.LabelFallback`, else the field's value.
2. If `link != ""` → `return LinkResolution{Href: link, Label: label}, nil`.
3. `link == ""`:
   - `cfg.Required` → return an error; caller aborts collection generation.
   - `cfg.Missing == "omit"` → `return LinkResolution{Omit: true}, nil`.
   - `cfg.Missing == "source_link"` → if `channelURL != ""`,
     `return LinkResolution{Href: channelURL, Label: label}, nil`; else fall
     through to `unlinked` (DEC-027).
   - `cfg.Missing == "unlinked"` (default, or the `source_link` fallthrough)
     → `return LinkResolution{Label: label, AsPlainText: true}, nil`.

`WriteItem` must check `Omit` **before writing any output** — the item is
skipped entirely, not partially rendered then discarded.

---

## `WriteItem` changes (`html.go`)

`WriteItem` gains a trailing `cfg ItemsConfig` parameter. Internally:

1. Call `resolveItemLink` first. If `Omit`, return immediately without
   writing anything.
2. Call `resolveItemContent` for the body.
3. Call `formatItemDate` for `pubDate`/`updated` using `cfg.DateFormat`.
4. Filter which of heading/date/content/footer-link/source-label sections are
   emitted based on `cfg.Fields` (ordered allowlist) — default (`cfg.Fields`
   unset before `applyDefaults()` runs, which never reaches `WriteItem`)
   emits all sections in their current order.
5. Render the footer anchor using `LinkResolution`: `<a href=%q>%s</a>` when
   `!AsPlainText`, otherwise the label as plain text with no `<a>`.
6. Gate the source/label block on `cfg.ShowSource`.

The `filters` slice (PageFind attributes) is entirely unaffected — DEC-025.

### `WriteHTML` caller changes

`WriteHTML` currently builds `gen` from `page.yaml` already. It now:

1. Runs `gen.Items.applyDefaults()` and `gen.Items.validate()` once, at
   collection-generation start (config-load time errors, not per-item).
2. Passes `gen.Items` to every `WriteItem` call in its item-scan loop.

---

## Theme integration (`themes.go`)

### `isTargetFile`

Add `"items.yaml"` to the target-file map so `findDirectoriesWithTargetFiles`
(used by `antenna themes list`) recognizes theme directories that define only
`items.yaml`.

### `updateItemsElement`

New function, following `updateHeadElements`'s existing pattern (read theme
file if present, unmarshal into a throwaway struct, copy over if non-empty —
not a deep merge):

```go
// Update a generator's items configuration from a theme items.yaml file.
func updateItemsElement(gen *Generator, themeName string) (bool, error) {
	fName := filepath.Join(themeName, "items.yaml")
	if _, err := os.Stat(fName); err != nil {
		return false, nil
	}
	src, err := os.ReadFile(fName)
	if err != nil {
		return false, fmt.Errorf("failed to read %q, %s", fName, err)
	}
	var items ItemsConfig
	if err := yaml.Unmarshal(src, &items); err != nil {
		return false, fmt.Errorf("failed to parse %q, %s", fName, err)
	}
	gen.Items = items
	return true, nil
}
```

Wired into `ApplyTheme` alongside the existing calls:

```go
if ok, err := updateItemsElement(gen, themeName); err != nil {
	return err
} else if ok {
	changed = ok
}
```

Note: unlike `updateHeadElements`, which copies individual sub-fields
(`Meta`/`Link`/`Script`/`Style`) conditionally, `updateItemsElement` assigns
the whole `ItemsConfig` wholesale when `items.yaml` is present — there is only
one field to copy (`gen.Items`), so there's no need for the
per-sub-field-presence checks `updateHeadElements` uses. `applyDefaults()` and
`validate()` still run later, at `WriteHTML` time, not here.

---

## Worked example

`theme/items.yaml`:

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

`label_field` is omitted — `static` is now the default (DEC-026), so only
`label_fallback` needs overriding to get a site-specific label instead of the
default "Continue reading".

After `antenna apply theme/ page.yaml`, `page.yaml` contains:

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

Rendered `<article>` output changes in the footer anchor — from today's
`<a href="https://example.com/post">https://example.com/post</a>` to
`<a href="https://example.com/post">read me</a>` — and also picks up the
`date_format`/`content_max_length` effects; title and content-selection
otherwise unaffected by this particular config. Note that even *without* this
`items.yaml` at all, the anchor text would already have changed to the
default "Continue reading" (DEC-026) — that default applies workspace-wide,
not just to collections using this theme.

---

## Files changed summary

| File | Nature of change |
|---|---|
| `generator.go` | Add `Items ItemsConfig` to `Generator`; define `ItemsConfig`, `LinkConfig`, `applyDefaults()`, `validate()` |
| `html.go` | Add `resolveItemContent`, `truncateWords`, `stripTags`, `formatItemDate`, `resolveItemLink`, `LinkResolution`; update `WriteItem` signature and body; update `WriteHTML` to call `applyDefaults()`/`validate()` and pass `gen.Items` through |
| `themes.go` | Add `"items.yaml"` to `isTargetFile`; add `updateItemsElement`; wire into `ApplyTheme` |
| `items_test.go` / `html_test.go` | New tests per `plan_item_formatting.md` |
| `item_formatting_proposal.md` | Finalized reference document (no further change expected) |
| `decisions.md` | DEC-022 through DEC-031 (already recorded) |
