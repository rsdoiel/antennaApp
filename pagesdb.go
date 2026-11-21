/*
antennaApp is a package for creating and curating blog, link blogs and social websites
Copyright (C) 2025 R. S. Doiel

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/
package antennaApp

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// pageRow is one row of the pages table (DR-0024). Output is the deprecated
// outputPath column. Outputs and Includes are the JSON columns decoded;
// Outputs is never empty for a row that has an Output (an old row reads as
// [Output]).
type pageRow struct {
	Input    string
	Output   string
	Updated  string
	Outputs  []string
	Includes []string
}

// firstWithExt returns the first path in list with the extension ext
// (".html" or ".md", compared without case), or "".
func firstWithExt(list []string, ext string) string {
	for _, p := range list {
		if strings.EqualFold(filepath.Ext(p), ext) {
			return p
		}
	}
	return ""
}

// pagesColumns returns the names of the columns of the pages table, or an
// empty set when there is no such table.
func pagesColumns(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query(`SELECT name FROM pragma_table_info('pages')`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols := map[string]bool{}
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		cols[n] = true
	}
	return cols, rows.Err()
}

/** ensurePagesColumns upgrades a pages table made before DR-0024 in place: it
 * adds the outputs and includes columns when they are missing, and copies each
 * row's outputPath into outputs as a one-element array. It is idempotent and
 * loses nothing. Verbs that write to the table call it; verbs that only read
 * do not (queryPageRows copes with an old table).
 *
 * Parameters:
 *   db (*sql.DB) — the pages database
 *
 * Returns:
 *   error — from the database
 *
 * Example:
 *   err := ensurePagesColumns(db)
 */
func ensurePagesColumns(db *sql.DB) error {
	cols, err := pagesColumns(db)
	if err != nil {
		return err
	}
	if len(cols) == 0 {
		return nil // no pages table
	}
	for _, name := range []string{"outputs", "includes"} {
		if !cols[name] {
			if _, err := db.Exec(`ALTER TABLE pages ADD COLUMN ` + name + ` TEXT DEFAULT ''`); err != nil {
				return fmt.Errorf("adding pages.%s: %w", name, err)
			}
		}
	}
	rows, err := db.Query(`SELECT inputPath, outputPath FROM pages WHERE (outputs IS NULL OR outputs = '') AND outputPath != ''`)
	if err != nil {
		return err
	}
	type fix struct{ in, out string }
	var fixes []fix
	for rows.Next() {
		var f fix
		if err := rows.Scan(&f.in, &f.out); err != nil {
			rows.Close()
			return err
		}
		fixes = append(fixes, f)
	}
	rows.Close()
	for _, f := range fixes {
		b, _ := json.Marshal([]string{f.out})
		if _, err := db.Exec(`UPDATE pages SET outputs = ?1 WHERE inputPath = ?2`, string(b), f.in); err != nil {
			return err
		}
	}
	return nil
}

// decodePathList decodes a JSON array of strings from a pages column; an
// empty value is an empty list, anything else that is not an array of strings
// is a data error naming the row.
func decodePathList(input, column, value string) ([]string, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	var list []string
	if err := json.Unmarshal([]byte(value), &list); err != nil {
		return nil, dataErrorf("pages row %q: %s is not a JSON array of strings: %w", input, column, err)
	}
	return list, nil
}

/** queryPageRows reads the rows of the pages table, ordered by the column
 * named (an unquoted column name, optionally followed by DESC). It works on a
 * table made before DR-0024: a missing outputs reads as [outputPath] and a
 * missing includes as empty.
 *
 * Parameters:
 *   db (*sql.DB) — the pages database
 *   orderBy (string) — for example "inputPath" or "updated DESC"
 *
 * Returns:
 *   []pageRow — the rows; none is an empty list, not an error
 *   error — from the database, or a data error for a bad JSON column
 *
 * Example:
 *   rows, err := queryPageRows(db, "inputPath")
 */
func queryPageRows(db *sql.DB, orderBy string) ([]pageRow, error) {
	cols, err := pagesColumns(db)
	if err != nil {
		return nil, err
	}
	if len(cols) == 0 {
		return nil, nil
	}
	outputsCol, includesCol := "''", "''"
	if cols["outputs"] {
		outputsCol = "COALESCE(outputs, '')"
	}
	if cols["includes"] {
		includesCol = "COALESCE(includes, '')"
	}
	rows, err := db.Query(`SELECT inputPath, COALESCE(outputPath, ''), COALESCE(updated, ''), ` +
		outputsCol + `, ` + includesCol + ` FROM pages ORDER BY ` + orderBy)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []pageRow
	for rows.Next() {
		var r pageRow
		var outs, incs string
		if err := rows.Scan(&r.Input, &r.Output, &r.Updated, &outs, &incs); err != nil {
			return nil, dataErrorf("failed to read a pages row: %w", err)
		}
		if r.Outputs, err = decodePathList(r.Input, "outputs", outs); err != nil {
			return nil, err
		}
		if r.Includes, err = decodePathList(r.Input, "includes", incs); err != nil {
			return nil, err
		}
		if len(r.Outputs) == 0 && r.Output != "" {
			r.Outputs = []string{r.Output}
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// openPagesDB opens the pages collection's database.
func (cfg *AppConfig) openPagesDB() (*sql.DB, string, error) {
	collection, err := cfg.GetCollection("pages.md")
	if err != nil {
		return nil, "", err
	}
	dsn := collection.DbName
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, dsn, fmt.Errorf("failed to open %s, %w", dsn, err)
	}
	return db, dsn, nil
}

/** recordPage writes the pages row for a source: the outputs and includes of
 * this run replace the row's (the last run wins), and outputPath, deprecated,
 * is the first .html output. The table is upgraded first if it is an old one.
 *
 * Parameters:
 *   source (string) — the source file, the row's key
 *   outputs ([]string) — every output path written, in command-line order
 *   includes ([]string) — every file the source includes, nested ones too
 *
 * Returns:
 *   error — from the database
 *
 * Example:
 *   err := cfg.recordPage("index.txt", []string{"index.md", "index.html"}, []string{"author.md"})
 */
func (cfg *AppConfig) recordPage(source string, outputs, includes []string) error {
	db, dsn, err := cfg.openPagesDB()
	if err != nil {
		return err
	}
	defer db.Close()
	if err := ensurePagesColumns(db); err != nil {
		return fmt.Errorf("%s, %w", dsn, err)
	}
	outJSON, _ := json.Marshal(append([]string{}, outputs...))
	incJSON := []byte("")
	if len(includes) > 0 {
		incJSON, _ = json.Marshal(includes)
	}
	timestamp := time.Now().Format(time.RFC3339)
	if _, err := db.Exec(SQLUpdatePage, source, firstWithExt(outputs, ".html"), timestamp, string(outJSON), string(incJSON)); err != nil {
		return fmt.Errorf("%s, %w", dsn, err)
	}
	return nil
}

/** pageStale says whether a row's outputs are out of date (DR-0024 13): ""
 * for a current page, "missing" when an output file is absent, "stale" when the
 * source or a file it includes was modified after the row's updated time, the
 * time antenna last built the page. A row with no readable updated time is
 * compared with its oldest output instead. Changes to the generator YAML, the
 * theme or antenna.yaml are not seen. A source that is gone is not stale here;
 * validate reports that row.
 *
 * Parameters:
 *   r (pageRow) — the row
 *   htdocs (string) — the directory the outputs are under
 *
 * Returns:
 *   string — "", "missing" or "stale"
 *
 * Example:
 *   state := pageStale(row, cfg.Htdocs)
 */
func pageStale(r pageRow, htdocs string) string {
	var newest time.Time
	for _, dep := range append([]string{r.Input}, r.Includes...) {
		if info, err := os.Stat(filepath.FromSlash(dep)); err == nil && info.ModTime().After(newest) {
			newest = info.ModTime()
		}
	}
	builtAt, err := time.Parse(time.RFC3339, r.Updated)
	oldestOutput := time.Time{}
	for _, out := range r.Outputs {
		info, err := os.Stat(filepath.Join(htdocs, filepath.FromSlash(out)))
		if err != nil {
			return "missing"
		}
		if oldestOutput.IsZero() || info.ModTime().Before(oldestOutput) {
			oldestOutput = info.ModTime()
		}
	}
	if err != nil {
		builtAt = oldestOutput
	}
	// updated has a granularity of a second.
	if !newest.IsZero() && newest.Truncate(time.Second).After(builtAt) {
		return "stale"
	}
	return ""
}

/** pageMarkdown returns the Markdown a page is listed and navigated by: the
 * Markdown copy its row records, else its source when that is Markdown and is
 * published, else the Markdown named like its HTML (DR-0023, DR-0024). The
 * table of contents and the derived previous and next links both use it, so
 * they agree on which entry is the page.
 *
 * Parameters:
 *   input (string) — the page's source file
 *   recorded (string) — the first .md output the row records, or ""
 *   html (string) — the page's HTML output, site-relative
 *
 * Returns:
 *   string — a site-relative .md path, or "" when there is none
 *
 * Example:
 *   md := cfg.pageMarkdown("bio.md", "", "about.html") // "bio.md"
 */
func (cfg *AppConfig) pageMarkdown(input, recorded, html string) string {
	if recorded != "" {
		return markdownPathFor(recorded)
	}
	if ext := strings.ToLower(filepath.Ext(input)); ext == ".md" || ext == ".markdown" {
		if src := markdownPathFor(input); src != "" && published(cfg.htdocsOrDot(), src) {
			return src
		}
	}
	return markdownPathFor(html)
}

// htdocsOrDot is the htdocs directory, or the working directory when none is set.
func (cfg *AppConfig) htdocsOrDot() string {
	if cfg.Htdocs == "" {
		return "."
	}
	return cfg.Htdocs
}
