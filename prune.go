package antennaApp

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

/** Prune implements `antenna prune [-w] [-all]` (DR-0028): it finds the
 * pages rows whose source file is gone and, with -w, deletes each row and
 * the output files it recorded. Without -w nothing changes; the rows and
 * outputs it would remove are listed on out and the answer is 1 (drift found).
 *
 * An output is kept, and named on eout, when a surviving row also records
 * it or when it resolves outside htdocs. When every row's source is missing
 * -w refuses, unless -all is given, because that is the shape of a wrong
 * working directory. The sitemap and toc.opml drop a pruned page on the next
 * generate.
 *
 * Parameters:
 *   out (io.Writer) — receives the list of rows and outputs
 *   eout (io.Writer) — receives notes about kept outputs and failures
 *   cfgName (string) — the antenna.yaml name
 *   args ([]string) — "-w" and "-all"
 *
 * Returns:
 *   error — nil when nothing needs pruning or -w pruned; a negative error for
 *   a dry run that found rows or a refusal; a usage error for a bad option;
 *   an I/O error when a delete failed (the rest is still processed)
 *
 * Example:
 *   err := app.Prune(os.Stdout, os.Stderr, "antenna.yaml", []string{"-w"})
 */
func (app *AntennaApp) Prune(out io.Writer, eout io.Writer, cfgName string, args []string) error {
	write, all := false, false
	for _, a := range args {
		switch a {
		case "-w", "--w":
			write = true
		case "-all", "--all":
			all = true
		default:
			return usageErrorf("unexpected argument %q; try 'antenna help prune'", a)
		}
	}
	cfg := &AppConfig{}
	if err := cfg.LoadConfig(cfgName); err != nil {
		return err
	}
	rows, err := cfg.GetPageRows()
	if err != nil {
		return err
	}
	var gone, alive []pageRow
	for _, r := range rows {
		if _, err := os.Stat(filepath.FromSlash(r.Input)); err != nil && os.IsNotExist(err) {
			gone = append(gone, r)
		} else {
			alive = append(alive, r)
		}
	}
	if len(gone) == 0 {
		return nil
	}
	if write && len(alive) == 0 && !all {
		fmt.Fprintf(eout, "antenna: prune: the source of every one of the %d pages is missing; is this the right directory? Use -all to prune them anyway\n", len(gone))
		return negativef("refusing to prune every page")
	}
	kept := map[string]string{} // output -> a surviving row that records it
	for _, r := range alive {
		for _, o := range rowOutputs(r) {
			kept[filepath.Clean(filepath.FromSlash(o))] = r.Input
		}
	}
	htdocs := cfg.htdocsOrDot()
	verb := "would delete"
	if write {
		verb = "delete"
	}
	var first error
	note := func(err error) {
		if first == nil {
			first = err
		}
	}
	done := map[string]bool{}
	for _, r := range gone {
		fmt.Fprintf(out, "%s\tsource missing\n", r.Input)
		for _, o := range rowOutputs(r) {
			rel := filepath.Clean(filepath.FromSlash(o))
			if filepath.IsAbs(rel) || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				fmt.Fprintf(eout, "antenna: prune: %s: output %s is outside htdocs, kept\n", r.Input, o)
				continue
			}
			if by, ok := kept[rel]; ok {
				fmt.Fprintf(eout, "antenna: prune: %s: output %s is also written by %s, kept\n", r.Input, o, by)
				continue
			}
			full := filepath.Join(htdocs, rel)
			info, err := os.Lstat(full)
			if err != nil || done[full] {
				continue
			}
			if !info.Mode().IsRegular() {
				fmt.Fprintf(eout, "antenna: prune: %s: output %s is not a regular file, kept\n", r.Input, o)
				continue
			}
			fmt.Fprintf(out, "  %s %s\n", verb, o)
			done[full] = true
			if write {
				if err := os.Remove(full); err != nil {
					fmt.Fprintf(eout, "antenna: prune: %s: %s\n", r.Input, err)
					note(ioErrorf("%s: %w", o, err))
				}
			}
		}
	}
	if !write {
		return negativef("%d pages have no source; run 'antenna prune -w' to remove them", len(gone))
	}
	db, dsn, err := cfg.openPagesDB()
	if err != nil {
		return err
	}
	defer db.Close()
	for _, r := range gone {
		if _, err := db.Exec(SQLDeletePageByInput, r.Input); err != nil {
			fmt.Fprintf(eout, "antenna: prune: %s: %s\n", r.Input, err)
			note(ioErrorf("%s, %w", dsn, err))
		}
	}
	fmt.Fprintf(out, "pruned %d pages; run 'antenna generate' to update the sitemap and toc.opml\n", len(gone))
	return first
}

// rowOutputs is every output a pages row records; an old row with only the
// deprecated outputPath reads as that one.
func rowOutputs(r pageRow) []string {
	if len(r.Outputs) > 0 {
		return r.Outputs
	}
	if r.Output != "" {
		return []string{r.Output}
	}
	return nil
}
