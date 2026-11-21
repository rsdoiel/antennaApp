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
	"os"
	"strings"
	"testing"
)

// DR-0024 12: nested includes, the report of what was included, and the
// separation of an included block from the text around it.

func writeAll(t *testing.T, files map[string]string) {
	t.Helper()
	for name, body := range files {
		write(t, name, body)
	}
}

func TestIncludesNestExpandAndAreReportedInOrder(t *testing.T) {
	t.Chdir(t.TempDir())
	writeAll(t, map[string]string{
		"a.txt": "A start\n\n@include-text-block b.txt\n\nA end\n",
		"b.txt": "B start\n\n@include-text-block c.txt\n\nB end\n",
		"c.txt": "C only\n",
		"d.txt": "D only\n",
	})
	src := "Top.\n\n@include-text-block a.txt\n\nMiddle.\n\n@include-text-block d.txt\n\n@include-text-block c.txt\n"
	var eout bytes.Buffer
	out, files, err := expandIncludesReport(src, &eout)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"A start", "B start", "C only", "B end", "A end", "D only"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "@include-") {
		t.Errorf("a directive is left:\n%s", out)
	}
	if got := strings.Join(files, ","); got != "a.txt,b.txt,c.txt,d.txt" {
		t.Errorf("includes (each once, first-met order): %s", got)
	}
	if eout.Len() != 0 {
		t.Errorf("three levels or fewer draw no warning:\n%s", eout.String())
	}
}

func chainOf(n int) map[string]string {
	files := map[string]string{}
	for i := 1; i <= n; i++ {
		body := "level " + string(rune('0'+i)) + "\n"
		if i < n {
			body += "\n@include-text-block l" + string(rune('0'+i+1)) + ".txt\n"
		}
		files["l"+string(rune('0'+i))+".txt"] = body
	}
	return files
}

func TestIncludesWarnFromTheFourthLevelAndStopAtTheSixth(t *testing.T) {
	t.Chdir(t.TempDir())
	// The page is level 0; l1 is level 1 ... l5 is level 5.
	writeAll(t, chainOf(5))
	var eout bytes.Buffer
	out, files, err := expandIncludesReport("Top.\n\n@include-text-block l1.txt\n", &eout)
	if err != nil {
		t.Fatalf("five levels are allowed: %s", err)
	}
	if !strings.Contains(out, "level 5") || len(files) != 5 {
		t.Errorf("all five levels should be in:\n%s\n%v", out, files)
	}
	if n := strings.Count(eout.String(), "nested"); n != 2 {
		t.Errorf("one warning each for level 4 and level 5, got %d:\n%s", n, eout.String())
	}
	if !strings.Contains(eout.String(), "l4.txt") || !strings.Contains(eout.String(), "l5.txt") {
		t.Errorf("a warning names the file:\n%s", eout.String())
	}

	writeAll(t, chainOf(6))
	eout.Reset()
	_, _, err = expandIncludesReport("Top.\n\n@include-text-block l1.txt\n", &eout)
	wantClass(t, "six levels", err, classData)
	if err == nil || !strings.Contains(err.Error(), "l6.txt") || !strings.Contains(err.Error(), "l1.txt") {
		t.Errorf("the error names the chain: %v", err)
	}
}

func TestIncludeCycleIsADataError(t *testing.T) {
	t.Chdir(t.TempDir())
	writeAll(t, map[string]string{
		"a.txt":    "A\n\n@include-text-block b.txt\n",
		"b.txt":    "B\n\n@include-text-block a.txt\n",
		"self.txt": "S\n\n@include-text-block self.txt\n",
	})
	_, _, err := expandIncludesReport("Top.\n\n@include-text-block a.txt\n", &bytes.Buffer{})
	wantClass(t, "a cycle", err, classData)
	if err == nil || !strings.Contains(err.Error(), "a.txt") || !strings.Contains(err.Error(), "b.txt") {
		t.Errorf("the error names the chain: %v", err)
	}
	_, _, err = expandIncludesReport("Top.\n\n@include-text-block self.txt\n", &bytes.Buffer{})
	wantClass(t, "a file that includes itself", err, classData)
}

// A code sample may hold a directive as text; it is not expanded.
func TestIncludedCodeIsNeverExpanded(t *testing.T) {
	t.Chdir(t.TempDir())
	writeAll(t, map[string]string{
		"sample.txt": "write @include-text-block other.txt like this\n\n@include-text-block other.txt\n",
		"other.txt":  "OTHER\n",
	})
	out, files, err := expandIncludesReport("Top.\n\n@include-code-block sample.txt text\n", &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "OTHER") || !strings.Contains(out, "@include-text-block other.txt") {
		t.Errorf("the sample's directives stay text:\n%s", out)
	}
	if len(files) != 1 || files[0] != "sample.txt" {
		t.Errorf("only the code file is listed: %v", files)
	}
}

// Seen on 2026-11-20: "Intro line." then a blank line then a code directive
// came out as "Intro line.~~~go", which is not a fence.
func TestIncludedBlocksAreSeparatedFromTheTextAround(t *testing.T) {
	t.Chdir(t.TempDir())
	writeAll(t, map[string]string{"hello.go": "package main\n", "note.txt": "A note.\n"})
	src := "Intro line.\n\n@include-code-block hello.go go\n\nMiddle line.\n\n@include-text-block note.txt\n\nLast line.\n"
	out, _, err := expandIncludesReport(src, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "~~~") && !strings.HasPrefix(line, "~~~") {
			t.Errorf("a fence must start its line: %q\n%s", line, out)
		}
	}
	if !strings.Contains(out, "Intro line.\n\n~~~go\npackage main\n~~~\n") {
		t.Errorf("a blank line before the fence, no blank line inside it:\n%q", out)
	}
	if !strings.Contains(out, "Middle line.\n\nA note.\n") {
		t.Errorf("a blank line before an included text block:\n%q", out)
	}
	if !strings.Contains(out, "A note.\n\nLast line.") {
		t.Errorf("a blank line after it:\n%q", out)
	}
}

func TestIncludeMessagesGoToTheErrorStream(t *testing.T) {
	t.Chdir(t.TempDir())
	var eout bytes.Buffer
	out, files, err := expandIncludesReport("Top.\n\n@include-text-block gone.txt\n\n@include-text-block ../x.txt\n", &eout)
	if err != nil {
		t.Fatalf("a missing file is left in the text, not an error: %s", err)
	}
	if !strings.Contains(out, "@include-text-block gone.txt") {
		t.Errorf("a missing include is left as written:\n%s", out)
	}
	if len(files) != 1 || files[0] != "gone.txt" {
		t.Errorf("a missing file is listed, a refused path is not: %v", files)
	}
	if !strings.Contains(eout.String(), "gone.txt") || !strings.Contains(eout.String(), "../x.txt") {
		t.Errorf("both problems are reported on the error stream:\n%s", eout.String())
	}
}

func TestExpandIncludesKeepsDirectivesInsideFrontMatterOfIncludedFile(t *testing.T) {
	t.Chdir(t.TempDir())
	writeAll(t, map[string]string{"frag.txt": "---\ntitle: Fragment\n---\nBody of fragment.\n"})
	out, _, err := expandIncludesReport("Top.\n\n@include-text-block frag.txt\n", &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "title: Fragment") || !strings.Contains(out, "Body of fragment.") {
		t.Errorf("an included file's front matter is dropped, its body kept:\n%s", out)
	}
	_ = os.Stdout
}
