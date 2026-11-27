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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rsdoiel/frontmatter"
	mdweb "github.com/rsdoiel/markdownweb/page"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"
)

/** Severity is how serious a validate finding is (DR-0017 2). An error or a
 * warning makes `antenna validate` exit 1; info does not.
 */
type Severity string

// The three severities of a finding.
const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

/** Finding is one thing `antenna validate` found (DR-0017 2).
 *
 * Fields:
 *   Code     (string)   — the stable name of the check, such as "link-html"
 *   Severity (Severity) — error, warning or info
 *   File     (string)   — the file the finding is about, slash separated; a
 *                         site finding names the file it is about
 *   Line     (int)      — the line in File, or 0 when there is none
 *   Message  (string)   — what is wrong and what to do
 */
type Finding struct {
	Code     string   `json:"code"`
	Severity Severity `json:"severity"`
	File     string   `json:"file"`
	Line     int      `json:"line,omitempty"`
	Message  string   `json:"message"`
}

// findingsError is the negative answer of a validate run that found errors or
// warnings. It is not a tool failure, so JSON mode does not add an error object.
type findingsError struct{ errs, warns int }

func (e *findingsError) Error() string {
	return fmt.Sprintf("validate: %d errors and %d warnings found", e.errs, e.warns)
}

// validator holds the state of one validate run.
type validator struct {
	cfgName  string      // the configuration file, for findings about it
	cfg      *AppConfig  // nil when a file or directory is checked with no antenna.yaml
	targets  LinkTargets // what antenna renders, nil when cfg is nil
	online   bool
	color    bool // wrap severity and location in ANSI escapes (DR-0020 5)
	files    int
	findings []Finding

	anchorCache map[string]map[string]bool // disk path -> fragments its headings make; nil value = unreadable
}

// add records a finding.
func (v *validator) add(code string, sev Severity, file string, line int, format string, args ...any) {
	v.findings = append(v.findings, Finding{Code: code, Severity: sev, File: file, Line: line, Message: fmt.Sprintf(format, args...)})
}

// counts returns how many findings there are of each severity.
func (v *validator) counts() (errs, warns, infos int) {
	for _, f := range v.findings {
		switch f.Severity {
		case SeverityError:
			errs++
		case SeverityWarning:
			warns++
		default:
			infos++
		}
	}
	return
}

// sortFindings puts findings in a fixed order: file, line, code.
func (v *validator) sortFindings() {
	sort.SliceStable(v.findings, func(i, j int) bool {
		a, b := v.findings[i], v.findings[j]
		if a.File != b.File {
			return a.File < b.File
		}
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		return a.Code < b.Code
	})
}

/** Validate implements `antenna validate [-json] [-online] [PATH]` (DR-0017):
 * a read-only check of a Markdown file, the Markdown files under a directory,
 * or the whole site, for correctness and for what Markdown readers such as
 * BonBon need. Findings go to out; the result is nil when there is no error or
 * warning, a negative-class error when there is, a usage error for a bad
 * command line and a no-input error for a path that does not exist.
 *
 * Parameters:
 *   out     (io.Writer) — findings and the summary
 *   eout    (io.Writer) — unused except for tool warnings
 *   cfgName (string)    — the configuration file, usually antenna.yaml
 *   args    ([]string)  — flags and at most one PATH
 *
 * Returns:
 *   error — nil, or the class the exit status comes from
 *
 * Example:
 *   err := app.Validate(os.Stdout, os.Stderr, "antenna.yaml", []string{"blog"})
 */
func (app *AntennaApp) Validate(out io.Writer, eout io.Writer, cfgName string, args []string) error {
	jsonMode, online, colorMode, target, err := parseValidateArgs(args)
	if err != nil {
		return err
	}
	err = app.validate(out, cfgName, target, jsonMode, online, colorMode)
	var found *findingsError
	if err != nil && jsonMode && !errors.As(err, &found) {
		class := ExitCodeFor(err)
		b, _ := json.Marshal(map[string]any{"error": err.Error(), "class": class.Name, "code": class.Code})
		fmt.Fprintln(out, string(b))
	}
	return err
}

// parseValidateArgs reads the flags and the path. Anything that does not parse
// is a usage error, and the mode is then plain text.
func parseValidateArgs(args []string) (jsonMode, online bool, colorMode, target string, err error) {
	paths := []string{}
	colorMode = "auto"
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-json" || a == "--json":
			jsonMode = true
		case a == "-online" || a == "--online":
			online = true
		case a == "-color" || a == "--color" || strings.HasPrefix(a, "-color=") || strings.HasPrefix(a, "--color="):
			var val string
			if eq := strings.Index(a, "="); eq >= 0 {
				val = a[eq+1:]
			} else if i+1 < len(args) {
				i++
				val = args[i]
			} else {
				return false, false, "", "", usageErrorf("%s needs auto, always or never; try 'antenna help validate'", a)
			}
			switch val {
			case "auto", "always", "never":
				colorMode = val
			default:
				return false, false, "", "", usageErrorf("-color %q is not auto, always or never; try 'antenna help validate'", val)
			}
		default:
			if strings.HasPrefix(a, "-") && len(a) > 1 {
				return false, false, "", "", usageErrorf("unknown option %q; try 'antenna help validate'", a)
			}
			paths = append(paths, a)
		}
	}
	if len(paths) > 1 {
		return false, false, "", "", usageErrorf("expected at most one PATH, got %d; try 'antenna help validate'", len(paths))
	}
	if len(paths) == 1 {
		target = paths[0]
	}
	return jsonMode, online, colorMode, target, nil
}

func (app *AntennaApp) validate(out io.Writer, cfgName, target string, jsonMode, online bool, colorMode string) error {
	v := &validator{online: online, cfgName: cfgName}
	cfg := &AppConfig{}
	cfgErr := cfg.LoadConfig(cfgName)
	var files []string
	if target == "" {
		if cfgErr != nil {
			return cfgErr
		}
		v.cfg = cfg
		var err error
		if v.targets, err = cfg.LinkTargets(); err != nil {
			return err
		}
		files = siteMarkdownFiles(v.targets)
		v.checkSite()
	} else {
		info, err := os.Stat(target)
		if err != nil {
			return noInputf("%s: %w", target, err)
		}
		if cfgErr == nil {
			v.cfg = cfg
			if v.targets, err = cfg.LinkTargets(); err != nil {
				return err
			}
		} else if !errors.Is(cfgErr, fs.ErrNotExist) {
			return cfgErr
		}
		if files, err = markdownUnder(target, info.IsDir()); err != nil {
			return err
		}
	}
	for _, f := range files {
		v.checkFile(f)
	}
	v.sortFindings()
	v.color = !jsonMode && useColor(colorMode, out)
	return v.report(out, jsonMode)
}

// report writes the findings and the summary, and returns the exit class.
func (v *validator) report(out io.Writer, jsonMode bool) error {
	errs, warns, infos := v.counts()
	if jsonMode {
		for _, f := range v.findings {
			b, _ := json.Marshal(f)
			fmt.Fprintln(out, string(b))
		}
		b, _ := json.Marshal(map[string]any{"summary": map[string]int{
			"files": v.files, "errors": errs, "warnings": warns, "info": infos}})
		fmt.Fprintln(out, string(b))
	} else {
		for _, f := range v.findings {
			where := f.File
			if f.Line > 0 {
				where = fmt.Sprintf("%s:%d", f.File, f.Line)
			}
			fmt.Fprintf(out, "%s: %s: %s: %s\n", v.paint("1", where), v.paint(severityColor(f.Severity), string(f.Severity)), f.Code, f.Message)
		}
		fmt.Fprintf(out, "checked %d %s: %s, %s, %s\n", v.files,
			map[bool]string{true: "file", false: "files"}[v.files == 1],
			v.paint(countColor(errs, "31"), fmt.Sprintf("%d errors", errs)),
			v.paint(countColor(warns, "33"), fmt.Sprintf("%d warnings", warns)),
			v.paint(countColor(infos, "2"), fmt.Sprintf("%d info", infos)))
	}
	if errs+warns > 0 {
		return classedAs(classNegative, &findingsError{errs: errs, warns: warns})
	}
	return nil
}

// markdownUnder returns the Markdown files to check for a path: the file
// itself, or every .md and .markdown file under a directory, in order, skipping
// folders whose name starts with a dot.
func markdownUnder(target string, isDir bool) ([]string, error) {
	if !isDir {
		return []string{filepath.ToSlash(filepath.Clean(target))}, nil
	}
	var files []string
	err := filepath.WalkDir(target, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if p != target && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if isMarkdownName(p) {
			files = append(files, filepath.ToSlash(filepath.Clean(p)))
		}
		return nil
	})
	if err != nil {
		return nil, ioErrorf("%s: %w", target, err)
	}
	sort.Strings(files)
	return files, nil
}

// isMarkdownName reports whether a file name is Markdown (.md or .markdown).
func isMarkdownName(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".md", ".markdown":
		return true
	}
	return false
}

// siteMarkdownFiles returns the Markdown files the site's own data names: the
// pages, the collection source files and the posts, that exist and are not
// generated items pages. Sorted.
func siteMarkdownFiles(targets LinkTargets) []string {
	var files []string
	for p := range targets {
		if !isMarkdownName(p) || strings.HasSuffix(strings.TrimSuffix(p, filepath.Ext(p)), "_items") {
			continue
		}
		if info, err := os.Stat(filepath.FromSlash(p)); err == nil && !info.IsDir() {
			files = append(files, p)
		}
	}
	sort.Strings(files)
	return files
}

// keyLine returns the 1-based line of the first front matter line that starts
// with key and a colon, or 1 when there is none.
func keyLine(raw []byte, key string) int {
	for i, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.HasPrefix(line, key+":") {
			return i + 1
		}
		if i > 0 && (line == "---" || line == "...") {
			break
		}
	}
	return 1
}

// bodyStartLine returns how many lines of raw come before the Markdown body,
// that is the front matter block with its two --- lines, or 0 when there is none.
func bodyStartLine(raw []byte) int {
	_, at, found, err := frontmatter.Block(raw)
	if err != nil || !found {
		return 0
	}
	return bytes.Count(raw[:at], []byte("\n"))
}

// checkFile runs the per-file checks on one Markdown file (DR-0017 4).
func (v *validator) checkFile(file string) {
	raw, err := os.ReadFile(filepath.FromSlash(file))
	if err != nil {
		v.files++
		v.add("front-matter", SeverityError, file, 0, "cannot be read: %s", err)
		return
	}
	fm, body, err := SplitFrontMatter(raw)
	if err != nil {
		v.files++
		v.add("front-matter", SeverityError, file, 1, "the front matter does not parse: %s", err)
		return
	}
	// Generated items pages (DR-0008) are not the author's to fix.
	if _, generated := fm["aggregateOf"]; generated {
		return
	}
	v.files++
	doc := &CommonMark{FrontMatter: fm, Text: body}
	v.checkFrontMatter(file, raw, doc)
	v.checkBody(file, raw, doc)
}

// checkFrontMatter runs the checks that read only the front matter and the
// heading.
func (v *validator) checkFrontMatter(file string, raw []byte, doc *CommonMark) {
	postPath := strings.TrimSpace(doc.GetAttributeString("postPath", ""))
	_, hasPub, _ := doc.GetAttributeDate("pubDate")
	_, hasDatePublished, _ := doc.GetAttributeDate("datePublished")
	if postPath != "" && !hasPub && !hasDatePublished {
		v.add("post-no-pubdate", SeverityWarning, file, keyLine(raw, "postPath"),
			"a post with no pubDate is a draft and is not published or listed; add pubDate, or use antenna blogit")
	}
	for _, key := range []string{"pubDate", "datePublished"} {
		stored, found, ok := doc.GetAttributeDate(key)
		if found && !ok {
			v.add("date-unreadable", SeverityError, file, keyLine(raw, key),
				"%s %q is not a date antenna can read (YYYY-MM-DD or an RSS date such as \"Thu, 19 Nov 2026 12:00:00 -0800\")", key, stored)
		}
	}
	if _, found, ok := doc.GetAttributeDate("pubDate"); found && ok && !isRSSDateValue(doc.FrontMatter["pubDate"]) {
		v.add("date-form", SeverityInfo, file, keyLine(raw, "pubDate"),
			"pubDate is not in the RSS 2.0 form BonBon expects (for example \"Thu, 19 Nov 2026 12:00:00 -0800\")")
	}
	if strings.TrimSpace(doc.GetAttributeString("title", "")) == "" {
		if heading := firstHeadingText(doc.Text); heading == "" {
			v.add("no-title", SeverityWarning, file, 1, "no title in the front matter and no level-1 heading")
		} else {
			v.add("title-from-heading", SeverityInfo, file, 1,
				"no title in the front matter; the heading %q would be used (add title, or run antenna page)", heading)
		}
	}
	v.checkSuggestions(file, raw, doc)
	if postPath != "" {
		if want := mdweb.MarkdownPath(postPath); want != "" && want != relativeToWorkDir(file) {
			v.add("postpath-mismatch", SeverityWarning, file, keyLine(raw, "postPath"),
				"postPath %q does not match where the file is", postPath)
		}
	}
	for _, key := range []string{"urlPrev", "urlNext", "rightMarginImage"} {
		if s, ok := doc.FrontMatter[key].(string); ok && strings.TrimSpace(s) != "" && !isSafePageAddress(s) {
			v.add("nav-address", SeverityError, file, keyLine(raw, key),
				"%s %q is not an http, https, mailto or relative address", key, s)
		}
	}
}

// isRSSDateValue reports whether a front matter value is a string in the RSS
// 2.0 date form.
func isRSSDateValue(v interface{}) bool {
	s, ok := v.(string)
	if !ok {
		return false
	}
	_, ok = parseRSSDate(s)
	return ok
}

// relativeToWorkDir returns a slash separated path relative to the working
// directory when p is under it, else p.
func relativeToWorkDir(p string) string {
	if !filepath.IsAbs(p) {
		return filepath.ToSlash(filepath.Clean(p))
	}
	if wd, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(wd, p); err == nil && !strings.HasPrefix(rel, "..") {
			return filepath.ToSlash(rel)
		}
	}
	return filepath.ToSlash(p)
}

// outIsTerminal reports whether w is a terminal. A variable so a test can say
// yes without a pty.
var outIsTerminal = func(w any) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

// useColor decides whether the report is colored (DR-0020 5): always and never
// are as said; auto is on for a terminal when NO_COLOR is unset or empty.
func useColor(mode string, out io.Writer) bool {
	switch mode {
	case "always":
		return true
	case "never":
		return false
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return outIsTerminal(out)
}

// paint wraps s in the ANSI escape for code when the report is colored.
func (v *validator) paint(code, s string) string {
	if !v.color || code == "" {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

// severityColor is the ANSI code for a severity: error red, warning yellow,
// info dim.
func severityColor(sev Severity) string {
	switch sev {
	case SeverityError:
		return "31"
	case SeverityWarning:
		return "33"
	}
	return "2"
}

// countColor colors a summary count only when it is not zero.
func countColor(n int, code string) string {
	if n == 0 {
		return ""
	}
	return code
}

// checkSuggestions runs the info checks for front matter a page or post could
// carry (DR-0020 1): description, author, a description that is too long.
func (v *validator) checkSuggestions(file string, raw []byte, doc *CommonMark) {
	desc := strings.TrimSpace(doc.GetAttributeString("description", ""))
	if desc == "" {
		v.add("no-description", SeverityInfo, file, 1, "no description; the table of contents and search results use it")
	} else if max := v.descriptionMax(); max > 0 && utf8.RuneCountInString(desc) > max {
		v.add("description-long", SeverityInfo, file, keyLine(raw, "description"),
			"description is %d characters, over the %d a feed or table of contents line suits", utf8.RuneCountInString(desc), max)
	}
	if !hasAuthor(doc.FrontMatter["author"]) && (v.cfg == nil || strings.TrimSpace(v.cfg.Author) == "") {
		v.add("no-author", SeverityInfo, file, 1, "no author; add author to the front matter, or set author in antenna.yaml for the whole site")
	}
}

// descriptionMax is the longest description before description-long: the
// descriptionMax of antenna.yaml, else 300. 0 means never.
func (v *validator) descriptionMax() int {
	if v.cfg != nil && v.cfg.DescriptionMax != nil {
		return *v.cfg.DescriptionMax
	}
	return 300
}

// hasAuthor reports whether a front matter author value names someone.
func hasAuthor(a any) bool {
	switch x := a.(type) {
	case nil:
		return false
	case string:
		return strings.TrimSpace(x) != ""
	case []any:
		return len(x) > 0
	case map[string]any:
		return len(x) > 0
	}
	return true
}
