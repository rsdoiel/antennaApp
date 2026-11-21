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
	"strings"
	"testing"
)

// DR-0019: a published .md that still holds an include directive shows the
// directive as text to a Markdown reader such as BonBon.

func TestValidateIncludeDirectiveInPublishedMarkdown(t *testing.T) {
	bodySite(t)
	write(t, "docs/p.md", bodyDoc("Intro.\n\n@include-text-block author.md\n\n@include-code-block hello.go go\n"))
	out, err := runValidate(t, "docs/p.md")
	wantClass(t, "include-directive findings", err, classNegative)
	if n := strings.Count(out, ": include-directive: "); n != 2 {
		t.Errorf("want 2 include-directive findings, got %d:\n%s", n, out)
	}
	line, ok := finding(out, "docs/p.md", "warning", "include-directive")
	if !ok || !strings.HasPrefix(line, "docs/p.md:8: ") {
		t.Errorf("first directive is on line 8: %q\n%s", line, out)
	}
	if !strings.Contains(line, "antenna page") {
		t.Errorf("the message should name the fix: %q", line)
	}
}

// A page about the directives shows them in code; that is not a use.
func TestValidateIncludeDirectiveInCodeIsLeftAlone(t *testing.T) {
	bodySite(t)
	write(t, "docs/p.md", bodyDoc("Write `@include-text-block FILE` like this:\n\n```\n@include-text-block author.md\n```\n\n    @include-code-block x.go go\n\nOr inline @include-text-block mid-sentence.\n"))
	out, _ := runValidate(t, "docs/p.md")
	if strings.Contains(out, "include-directive") {
		t.Errorf("code and mid-sentence mentions are not directives:\n%s", out)
	}
}
