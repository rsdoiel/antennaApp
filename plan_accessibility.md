# Antenna App — Implementation Plan: HTML Accessibility and ARIA

Design: `design_accessibility.md`
Decisions: DEC-016–DEC-021 in `decisions.md`

TDD pattern: write failing tests first, confirm red, implement to green.

---

## Dependency order

```
Phase 1: Fix rel typo (isolated, no dependencies)
Phase 2: Lang field (Generator struct + LoadConfig + WriteHTML/WriteHtmlPage)
Phase 3: Skip navigation link (WriteHTML + WriteHtmlPage)
Phase 4: WriteItem restructure (date/time + address→footer)
Phase 5: No-header warning (WriteHTML)
Phase 6: Smoke test
```

---

## Phase 1 — Fix `rel="altenate"` typo (DEC-021)

**Files:** `html.go`, `html_test.go`

### 1.1 — Write test

In `html_test.go`, add a test in the `TestWriteHeadElement` group verifying that the
Markdown alternate link contains `rel="alternate"` and does NOT contain `rel="altenate"`.

Confirm: `go test -run TestWriteHeadElement` → red (or existing tests pass but new case catches typo).

### 1.2 — Fix typo

In `html.go` `writeHeadElement`, change `"altenate"` → `"alternate"`.

Confirm: `go test -run TestWriteHeadElement` → green.

---

## Phase 2 — Configurable `lang` attribute (DEC-018)

**Files:** `generator.go`, `html.go`, `html_test.go`

### 2.1 — Write tests

Add tests verifying:
- Default generator produces `<html lang="en-US">`.
- Generator with `Lang = "fr-FR"` produces `<html lang="fr-FR">`.

Tests use `WriteHTML` and `WriteHtmlPage` with a `bytes.Buffer`.

### 2.2 — Add `Lang` to `Generator` struct

In `generator.go`:
- Add `Lang string` field to `Generator`.
- In `NewGenerator`, set `gen.Lang = "en-US"`.
- In `LoadConfig`, read optional `lang:` key; if non-empty, set `gen.Lang`.

### 2.3 — Use `gen.Lang` in HTML output

In `html.go`:
- `WriteHTML`: change `<html lang="en-US">` → `fmt.Sprintf("<html lang=%q>", gen.Lang)`.
- `WriteHtmlPage`: same change.

Confirm: `go test -run TestLang` → green.

---

## Phase 3 — Skip navigation link (DEC-016)

**Files:** `html.go`, `html_test.go`

### 3.1 — Write tests

Add tests verifying that `WriteHTML` and `WriteHtmlPage` output contains:
- `<a href="#main-content" class="skip-link">Skip to main content</a>`
- The skip link appears before `<nav` in the output.

### 3.2 — Emit skip link

In `html.go`, immediately after `<body>` is written in both `WriteHTML` and `WriteHtmlPage`:

```go
fmt.Fprintln(out, `  <a href="#main-content" class="skip-link">Skip to main content</a>`)
```

Confirm: `go test -run TestSkipLink` → green.

---

## Phase 4 — `WriteItem` restructure (DEC-017, DEC-019)

**Files:** `html.go`, `html_test.go`

### 4.1 — Write tests

Update existing `TestWriteItem` tests to assert:
- No `<address>` in output.
- `<footer>` present inside `<article>`.
- `<time datetime="YYYY-MM-DD">` present for pubDate.
- When `updated` differs from pubDate, two `<time>` elements present.
- Date text not inside `<h2>`.
- `<h2>` contains only the title (not date).

Confirm: `go test -run TestWriteItem` → red.

### 4.2 — Rewrite `WriteItem` output block

In `html.go` `WriteItem`:

1. Build `titleStr` as just the title text (no date appended).
2. Build `dateStr` as `<time datetime="%s">%s</time>` for `pressTime` (10-char truncated pubDate).
   If `updated` differs, append `, updated: <time datetime="%s">%s</time>`.
3. Emit article structure:
   ```html
   <article …>
     <h2>TITLE</h2>
     <p><time datetime="DATE">DATE</time></p>
     <p>DESCRIPTION</p>
     <footer>
       <a href="LINK">LINK</a>
     </footer>
   </article>
   ```
4. The two template strings (with and without `data-pagefind-filter`) must both be updated.

Confirm: `go test -run TestWriteItem` → green.
Confirm: `go test ./...` → all green.

---

## Phase 5 — No-header warning (DEC-020)

**Files:** `html.go`, `html_test.go`

### 5.1 — Write test

Add a test for `WriteHTML` with empty `gen.Header` that captures stderr and verifies it
contains `"warning: aggregate page has no <h1>"`.

### 5.2 — Emit warning

In `html.go` `WriteHTML`, after the header conditional block:

```go
if gen.Header == "" {
    fmt.Fprintln(os.Stderr, "warning: aggregate page has no <h1>; set a 'header' value in the generator YAML")
}
```

Confirm: `go test -run TestNoHeaderWarning` → green.
Confirm: `go test ./...` → all green.

---

## Phase 6 — Smoke test

Using the blog at `~/Laboratory/rsdoiel.github.io`:

### 6.1 — Build

```bash
cd ~/Laboratory/antennaApp && go build -o bin/antenna cmd/antenna/*.go
```

### 6.2 — Regenerate posts

```bash
cd ~/Laboratory/rsdoiel.github.io && ~/Laboratory/antennaApp/bin/antenna generate pages.md
```

Verify in `blog/2026/04/15/fountain_and_recording_agents.html`:
- `<a href="#main-content" class="skip-link">Skip to main content</a>` present before `<nav`.
- `<html lang="en-US">` (unchanged default).
- `rel="alternate"` on the Markdown link (no typo).
- No regression in metadata `<meta>` elements.

### 6.3 — Verify `WriteItem` output

If a harvested feed DB is available, run `antenna generate` on a collection and inspect an
aggregate HTML page for `<footer>` and `<time>` elements, absence of `<address>`.

If no feed DB is available, rely on unit tests from Phase 4.

### 6.4 — Test custom `lang`

Temporarily add `lang: fr-FR` to `page.yaml`, regenerate one page, verify `<html lang="fr-FR">`,
revert.

---

## Summary table

| Phase | Decisions | Tests | Production code |
|-------|-----------|-------|-----------------|
| 1 | DEC-021 | `TestWriteHeadElement` update | `html.go` typo fix |
| 2 | DEC-018 | `TestLang` new | `generator.go` Lang field; `html.go` fmt.Sprintf |
| 3 | DEC-016 | `TestSkipLink` new | `html.go` skip link in both paths |
| 4 | DEC-017, DEC-019 | `TestWriteItem` updates | `html.go` WriteItem rewrite |
| 5 | DEC-020 | `TestNoHeaderWarning` new | `html.go` stderr warning |
| 6 | — | smoke test | — |
