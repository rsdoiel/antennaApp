# Enhanced Front Matter Processing Feature Request

## Summary

Two related improvements to front matter handling in Antenna App:

1. **Proper front matter stripping in `GeneratePosts`** — the current path renders YAML front matter as visible body text, which is a bug.
2. **Front matter → standard HTML metadata and PageFind filter attributes** — emit `<meta>` elements in `<head>` from front matter fields, and `data-pagefind-filter` meta elements to enable faceted search in PageFind 1.5+.

These changes make Antenna App a turnkey CMS that works with PageFind out of the box, without requiring any post-processing step.

---

## Background

### The `GeneratePosts` bug (generator.go:314)

`GeneratePosts` stores raw Markdown (including YAML front matter) in the database and later creates a `CommonMark` struct without parsing it:

```go
// current — wrong
doc := &CommonMark{Text: sourceMarkdown}
```

Because `doc.Parse()` is never called, `doc.FrontMatter` is empty and `doc.Text` contains the full source including the `---` delimiters. When goldmark converts that to HTML, the `---` delimiters become `<hr>` elements and the YAML becomes a visible paragraph in the page body.

The `antenna page` path (page.go:160) uses `LoadCommonMark`, which correctly calls `doc.Parse()`, so pages work as expected. Posts do not.

### Standard HTML metadata

Common front matter fields (`title`, `description`/`abstract`, `author`, `keywords`, `datePublished`) should be reflected as standard `<meta>` elements in the `<head>`. This improves SEO, social sharing, and accessibility tooling. Currently none of these are emitted from front matter.

### PageFind faceted search

PageFind 1.5+ supports faceted/filter search via `data-pagefind-filter` attributes in the page HTML, captured at index time. Without these attributes the `<pagefind-filter-pane>` and `<pagefind-filter-dropdown>` components remain empty. The fields most useful for filtering a blog are `keywords`, `series`, `author`, and `datePublished`.

PageFind filters work with `<meta>` elements in `<head>`, even outside `<body>`, making head meta the cleanest place to put per-page filter values.

---

## Proposed Changes

### 1. `cmarkdoc.go` — add a slice helper

Add a helper method next to the existing `GetAttributeString` and `GetAttributeBool`:

```go
// GetAttributeStringSlice returns a []string for a front matter key whose
// value is either a YAML sequence or a plain string. Returns nil if absent.
func (doc *CommonMark) GetAttributeStringSlice(key string) []string {
    val, ok := doc.FrontMatter[key]
    if !ok {
        return nil
    }
    switch v := val.(type) {
    case string:
        if v == "" {
            return nil
        }
        return []string{v}
    case []interface{}:
        out := make([]string, 0, len(v))
        for _, item := range v {
            if s, ok := item.(string); ok && s != "" {
                out = append(out, s)
            }
        }
        return out
    }
    return nil
}
```

### 2. `generator.go` — fix `GeneratePosts` (around line 314)

Replace the bare struct literal with a proper parse, preserving the existing fallback behaviour:

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

Also update the `WriteHtmlPage` call (line 334) to pass the front matter:

```go
// before
if err := gen.WriteHtmlPage(htmlName, link, postPath, pubDate, innerHTML); err != nil {

// after
if err := gen.WriteHtmlPage(htmlName, link, postPath, pubDate, innerHTML, doc.FrontMatter); err != nil {
```

### 3. `html.go` — extend `writeHeadElement`

Change the signature to accept front matter:

```go
// before
func (gen *Generator) writeHeadElement(out io.Writer, postPath string)

// after
func (gen *Generator) writeHeadElement(out io.Writer, postPath string, frontMatter map[string]interface{})
```

After the existing `<meta>` and `<link>` blocks, emit front matter–derived metadata. The `<title>` should prefer the front matter `title` over `gen.Title` when present:

```go
// Standard HTML metadata from front matter
if frontMatter != nil {
    // Override page title with post-specific title when present
    if t, ok := frontMatter["title"].(string); ok && t != "" {
        fmt.Fprintf(out, "  <title>%s</title>\n", t)
    }

    // description — prefer "abstract" then "description"
    desc := ""
    if v, ok := frontMatter["abstract"].(string); ok && v != "" {
        desc = v
    } else if v, ok := frontMatter["description"].(string); ok && v != "" {
        desc = v
    }
    if desc != "" {
        fmt.Fprintf(out, "  <meta name=\"description\" content=%q >\n", desc)
    }

    // author
    if v, ok := frontMatter["author"].(string); ok && v != "" {
        fmt.Fprintf(out, "  <meta name=\"author\" content=%q >\n", v)
    }

    // datePublished
    if v, ok := frontMatter["datePublished"].(string); ok && v != "" {
        fmt.Fprintf(out, "  <meta name=\"datePublished\" content=%q >\n", v)
    }

    // keywords — emit one combined <meta name="keywords"> and one
    // data-pagefind-filter meta per keyword for PageFind faceted search
    doc := &CommonMark{FrontMatter: frontMatter}
    keywords := doc.GetAttributeStringSlice("keywords")
    if len(keywords) > 0 {
        fmt.Fprintf(out, "  <meta name=\"keywords\" content=%q >\n",
            strings.Join(keywords, ", "))
        for _, kw := range keywords {
            fmt.Fprintf(out,
                "  <meta data-pagefind-filter=\"keywords[content]\" content=%q >\n", kw)
        }
    }

    // series — one data-pagefind-filter meta per entry
    series := doc.GetAttributeStringSlice("series")
    for _, s := range series {
        fmt.Fprintf(out,
            "  <meta data-pagefind-filter=\"series[content]\" content=%q >\n", s)
    }

    // author as a pagefind filter (single value)
    if v, ok := frontMatter["author"].(string); ok && v != "" {
        fmt.Fprintf(out,
            "  <meta data-pagefind-filter=\"author[content]\" content=%q >\n", v)
    }

    // datePublished as a pagefind filter
    if v, ok := frontMatter["datePublished"].(string); ok && v != "" {
        fmt.Fprintf(out,
            "  <meta data-pagefind-filter=\"datePublished[content]\" content=%q >\n", v)
    }
}
```

Update the two internal callers of `writeHeadElement`:

- `WriteHTML` (line 174): `gen.writeHeadElement(out, "", nil)` — collection aggregate pages have no single-post front matter.
- `WriteHtmlPage` (line 285): `gen.writeHeadElement(out, postPath, frontMatter)`.

### 4. `html.go` — extend `WriteHtmlPage`

Add `frontMatter map[string]interface{}` as the last parameter:

```go
// before
func (gen *Generator) WriteHtmlPage(htmlName string, link string, postPath, pubDate string, innerHTML string) error

// after
func (gen *Generator) WriteHtmlPage(htmlName string, link string, postPath, pubDate string, innerHTML string, frontMatter map[string]interface{}) error
```

Pass it through to `writeHeadElement` (change described above).

No change is needed to the `<article>` element itself: `datePublished` is already surfaced via `data-published` and will be captured by the `data-pagefind-filter` meta element in `<head>`.

### 5. `page.go` — pass front matter from `Page` (line 205)

`Page` already has `doc.FrontMatter` available after `LoadCommonMark`. Update the call:

```go
// before
if err := gen.WriteHtmlPage(htmlName, "", postPath, "", innerHTML); err != nil {

// after
if err := gen.WriteHtmlPage(htmlName, "", postPath, "", innerHTML, doc.FrontMatter); err != nil {
```

---

## Resulting HTML (example)

For a blog post with this front matter:

```yaml
title: Mostly Oberon
author: R. S. Doiel
datePublished: '2020-04-11'
keywords:
  - Oberon
  - programming
series:
  - mostly-oberon
abstract: An introduction to the Oberon programming language.
```

The generated `<head>` would include:

```html
<title>Mostly Oberon</title>
<meta name="description" content="An introduction to the Oberon programming language." >
<meta name="author" content="R. S. Doiel" >
<meta name="datePublished" content="2020-04-11" >
<meta name="keywords" content="Oberon, programming" >
<meta data-pagefind-filter="keywords[content]" content="Oberon" >
<meta data-pagefind-filter="keywords[content]" content="programming" >
<meta data-pagefind-filter="series[content]" content="mostly-oberon" >
<meta data-pagefind-filter="author[content]" content="R. S. Doiel" >
<meta data-pagefind-filter="datePublished[content]" content="2020-04-11" >
```

And the post body would contain only the post content — no rendered YAML.

---

## Search Page Integration

Once these changes are built, the `search.md` on the website can be updated to use PageFind's faceted search components:

```markdown
<pagefind-config faceted preload></pagefind-config>
<pagefind-input placeholder="Search…"></pagefind-input>
<pagefind-filter-pane></pagefind-filter-pane>
<pagefind-summary></pagefind-summary>
<pagefind-results></pagefind-results>
```

`<pagefind-filter-pane>` auto-discovers all indexed filter keys (`keywords`, `series`, `author`, `datePublished`) and renders collapsible groups for each. The `faceted` flag makes an empty search return all results so readers can browse by filter alone. The `preload` flag loads the index immediately on page load.

---

## Files Changed

| File | Change |
|------|--------|
| `cmarkdoc.go` | Add `GetAttributeStringSlice` helper |
| `generator.go` | Fix `GeneratePosts` to call `doc.Parse()`; update `WriteHtmlPage` call signature |
| `html.go` | Add `frontMatter` param to `writeHeadElement` and `WriteHtmlPage`; emit standard and pagefind meta elements |
| `page.go` | Pass `doc.FrontMatter` to `WriteHtmlPage` |

No database schema changes are required. No new dependencies are needed.
