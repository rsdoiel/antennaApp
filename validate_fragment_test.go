package antennaApp

import (
	"strings"
	"testing"
)

// fragDocs writes the pages the fragment tests link between.
func fragDocs(t *testing.T) {
	t.Helper()
	bodySite(t)
	write(t, "docs/other.md", bodyDoc("# Other\n\n## Soil & Seeds\n\n## Notes\n\n## Notes\n\n## The *big* [idea](x.md)\n\n```\n# not a heading\n```\n\n<a id=\"hand-made\"></a>\n"))
}

func fragFindings(out string) []string {
	var got []string
	for _, l := range strings.Split(out, "\n") {
		if strings.Contains(l, ": link-fragment: ") {
			got = append(got, l)
		}
	}
	return got
}

func TestValidateLinkFragmentSameFile(t *testing.T) {
	fragDocs(t)
	write(t, "docs/p.md", bodyDoc("# Intro\n\n## Soil & Seeds\n\n[ok](#soil--seeds) [intro](#intro) [Case](#INTRO) [top](#top) [bad](#nope)\n"))
	out, _ := runValidate(t, "docs/p.md")
	got := fragFindings(out)
	if len(got) != 1 || !strings.Contains(got[0], "#nope") || !strings.Contains(got[0], "warning") {
		t.Fatalf("want one link-fragment warning for #nope, got %v\n%s", got, out)
	}
}

func TestValidateLinkFragmentOtherFile(t *testing.T) {
	fragDocs(t)
	write(t, "docs/p.md", bodyDoc("[a](other.md#soil--seeds) [b](other.md#notes-1) [c](other.md#the-big-idea) [d](other.md#hand-made) [e](other.md#gone) [f](other.md#not-a-heading) [g](../about.md#about)\n"))
	out, _ := runValidate(t, "docs/p.md")
	got := fragFindings(out)
	if len(got) != 2 {
		t.Fatalf("want 2 link-fragment findings (#gone, #not-a-heading), got %v\n%s", got, out)
	}
	for _, g := range got {
		if !strings.Contains(g, "#gone") && !strings.Contains(g, "#not-a-heading") {
			t.Errorf("unexpected finding %s", g)
		}
	}
}

func TestValidateLinkFragmentLeavesOtherCasesAlone(t *testing.T) {
	fragDocs(t)
	write(t, "docs/p.md", bodyDoc("[missing](nowhere.md#x) [html](../about.html#anything) [web](https://example.org/a.md#x) [mail](mailto:a@example.org) [dir](../series/#x)\n"))
	out, _ := runValidate(t, "docs/p.md")
	if got := fragFindings(out); len(got) != 0 {
		t.Errorf("a missing file, an HTML page or an outside address has no heading to check: %v", got)
	}
	if _, ok := finding(out, "docs/p.md", "warning", "link-missing"); !ok {
		t.Errorf("the missing file is still link-missing:\n%s", out)
	}
}
