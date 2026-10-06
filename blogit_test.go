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
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const blogitDoc = "---\ntitle: A post\n---\n\nbody\n"

func writeBlogitDoc(t *testing.T, name string) {
	t.Helper()
	if err := os.WriteFile(name, []byte(blogitDoc), 0o644); err != nil {
		t.Fatal(err)
	}
}

// blogit takes [COLLECTION_NAME] FILEPATH [POST_DATE]. Whether the second of
// two arguments is a date or a file is decided by whether it is a date, not by
// whether it contains a hyphen: hyphenated file names are everyday.
func TestBlogIt_ArgumentForms(t *testing.T) {
	today := time.Now().Format("2006/01/02")
	cases := []struct {
		label string
		args  []string
		dir   string // where the post must land under blog/
	}{
		{"file only", []string{"plain.md"}, today},
		{"hyphenated file only", []string{"my-post.md"}, today},
		{"collection and hyphenated file", []string{"pages.md", "my-post.md"}, today},
		{"file and date", []string{"my-post.md", "2026-01-02"}, "2026/01/02"},
		{"collection, file and date", []string{"pages.md", "my-post.md", "2026-01-02"}, "2026/01/02"},
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			newWorkspace(t, true)
			file := c.args[0]
			if len(c.args) >= 2 && strings.HasSuffix(c.args[1], ".md") {
				file = c.args[1]
			}
			writeBlogitDoc(t, file)
			if err := runIn(t, "blogit", c.args...); err != nil {
				t.Fatalf("blogit %v: %s", c.args, err)
			}
			want := filepath.Join("blog", filepath.FromSlash(c.dir), file)
			if _, err := os.Stat(want); err != nil {
				t.Errorf("blogit %v: expected %s: %s", c.args, want, err)
			}
		})
	}
}

// A date that is shaped like one but is not a real date is the user's typo,
// and the message names the argument that failed to parse, not another one.
func TestBlogIt_BadDateIsUsageAndNamesTheDate(t *testing.T) {
	cases := [][]string{
		{"my-post.md", "2026-13-45"},
		{"pages.md", "my-post.md", "2026-13-45"},
	}
	for _, args := range cases {
		newWorkspace(t, true)
		writeBlogitDoc(t, "my-post.md")
		err := runIn(t, "blogit", args...)
		wantClass(t, "blogit "+strings.Join(args, " "), err, classUsage)
		if err != nil && !strings.Contains(err.Error(), `failed to parse "2026-13-45" as post date`) {
			t.Errorf("blogit %v: the message should say it failed to parse the date %q, got: %v", args, "2026-13-45", err)
		}
	}
}
