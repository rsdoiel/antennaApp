package antennaApp

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// generatorStamp matches the <meta name="generator" content="antenna/..."> tag
// antenna writes into every HTML page it renders, in either attribute order.
var generatorStamp = regexp.MustCompile(`<meta[^>]*name="generator"[^>]*content="antenna/|<meta[^>]*content="antenna/[^"]*"[^>]*name="generator"`)

/** retireOldItemsPage removes the items page antenna 0.1.2 wrote at STEM.html,
 * now that an aggregate collection's items are at STEM_items.html (DR-0034 4).
 * The file is removed only when it carries antenna's generator stamp and no
 * pages row records it, so a page the author rendered from STEM.md with
 * `antenna page` is kept silently; a STEM.html without the stamp is kept and
 * named in a warning. A removal writes one note to eout, once, since the file
 * is gone afterwards.
 *
 * Parameters:
 *   eout   (io.Writer) — where the note or warning goes
 *   htdocs (string)    — the output directory
 *   stem   (string)    — the collection source's name without extension
 *
 * Returns:
 *   error — from reading or removing the file; a missing file is not an error
 *
 * Example:
 *   err := cfg.retireOldItemsPage(os.Stderr, "htdocs", "news")
 */
func (cfg *AppConfig) retireOldItemsPage(eout io.Writer, htdocs, stem string) error {
	name := stem + ".html"
	p := filepath.Join(htdocs, name)
	src, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if !generatorStamp.Match(src) {
		fmt.Fprintf(eout, "warning: %s is not antenna's items page and was left alone; the items of %s.md are now at %s_items.html\n", name, stem, stem)
		return nil
	}
	if rows, err := cfg.GetPageRows(); err == nil {
		for _, r := range rows {
			for _, o := range rowOutputs(r) {
				if filepath.Clean(filepath.FromSlash(o)) == name {
					return nil
				}
			}
		}
	}
	if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if cfg.retiredItemsPages != nil {
		*cfg.retiredItemsPages = append(*cfg.retiredItemsPages, stem)
		return nil
	}
	fmt.Fprintln(eout, retiredItemsNote([]string{stem}))
	return nil
}

/** retiredItemsNote is the one line that says which old items pages were
 * removed and where the items are now (DR-0034 4).
 *
 * Parameters:
 *   stems ([]string) — the collection source names without extension, not empty
 *
 * Returns:
 *   string — the note, starting "note: "
 *
 * Example:
 *   retiredItemsNote([]string{"news", "food"})
 */
func retiredItemsNote(stems []string) string {
	if len(stems) == 1 {
		s := stems[0]
		return fmt.Sprintf("note: %s.html, the items page of %s.md in earlier releases, was removed; the items are now at %s_items.html, and links to %s.html now fail", s, s, s, s)
	}
	var old, now []string
	for _, s := range stems {
		old = append(old, s+".html")
		now = append(now, s+"_items.html")
	}
	return fmt.Sprintf("note: removed the items pages of %d collections from earlier releases (%s); the items are now at %s, and links to the old addresses now fail",
		len(stems), strings.Join(old, ", "), strings.Join(now, ", "))
}
