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
*/
package antennaApp

import (
	"strings"
	"testing"
)

/* renderCases pin the HTML that CommonMark.ToHTML and ToUnsafeHTML produce
 * today (goldmark v1 plus the GFM, definition list, footnote, emoji, figure,
 * CJK and fences extensions). They exist so a change of Markdown engine can
 * be compared against known output.
 */
var renderCases = []struct {
	name string
	src  string
	want string // output of ToHTML with MathJax off
}{
	{"heading and emphasis", "# Title\n\nPara with *em* and **strong**.\n",
		"<h1>Title</h1>\n<p>Para with <em>em</em> and <strong>strong</strong>.</p>\n"},
	{"gfm table", "| a | b |\n|---|---|\n| 1 | 2 |\n",
		"<table>\n<thead>\n<tr>\n<th>a</th>\n<th>b</th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td>1</td>\n<td>2</td>\n</tr>\n</tbody>\n</table>\n"},
	{"gfm strikethrough", "~~gone~~\n", "<p><del>gone</del></p>\n"},
	{"gfm task list", "- [x] done\n- [ ] todo\n",
		"<ul>\n<li><input checked=\"\" disabled=\"\" type=\"checkbox\"> done</li>\n<li><input disabled=\"\" type=\"checkbox\"> todo</li>\n</ul>\n"},
	{"gfm autolink", "see https://example.org now\n",
		"<p>see <a href=\"https://example.org\">https://example.org</a> now</p>\n"},
	{"definition list", "Term\n: Definition\n", "<dl>\n<dt>Term</dt>\n<dd>Definition</dd>\n</dl>\n"},
	{"footnote", "Text[^1]\n\n[^1]: Note.\n",
		"<p>Text<sup id=\"fnref:1\"><a href=\"#fn:1\" class=\"footnote-ref\" role=\"doc-noteref\">1</a></sup></p>\n<div class=\"footnotes\" role=\"doc-endnotes\">\n<hr>\n<ol>\n<li id=\"fn:1\">\n<p>Note.&#160;<a href=\"#fnref:1\" class=\"footnote-backref\" role=\"doc-backlink\">&#x21a9;&#xfe0e;</a></p>\n</li>\n</ol>\n</div>\n"},
	{"emoji shortcode", "hello :smile:\n", "<p>hello &#x1f604;</p>\n"},
	{"cjk line break is joined", "日本語\nの文章\n", "<p>日本語の文章</p>\n"},
	{"fenced code", "```go\nfmt.Println(1)\n```\n",
		"<pre><code class=\"language-go\">fmt.Println(1)\n</code></pre>\n"},
	{"figure with caption line", "![Oscar](cat.png)\nAwesome **caption**.\n",
		"<figure>\n<a href=\"cat.png\">\n<img src=\"cat.png\" alt=\"Oscar\">\n</a>\n<figcaption><p>Awesome <strong>caption</strong>.</p></figcaption>\n</figure>\n"},
	{"figure with several images", "![A](a.png)\n![B](b.png)\nBoth.\n",
		"<figure>\n<a href=\"a.png\">\n<img src=\"a.png\" alt=\"A\">\n</a>\n<a href=\"b.png\">\n<img src=\"b.png\" alt=\"B\">\n</a>\n<figcaption><p>Both.</p></figcaption>\n</figure>\n"},
	{"image without caption stays an image", "![NoCap](n.png)\n", "<p><img src=\"n.png\" alt=\"NoCap\"></p>\n"},
	{"linked image is not made a figure", "[![Linked](l.png)](http://x.org)\nLink caption.\n",
		"<p><a href=\"http://x.org\"><img src=\"l.png\" alt=\"Linked\"></a>\nLink caption.</p>\n"},
	{"fences nest with class and id", ":::{.blue}\n## In\n\ntext\n\n:::{#id2 .red}\nnested\n:::\n:::\n",
		"<div data-fence=\"0\" class=\"blue\">\n<h2>In</h2>\n<p>text</p>\n<div data-fence=\"1\" id=\"id2\" class=\"red\">\n<p>nested</p>\n</div>\n</div>\n"},
	{"math is plain text when MathJax is off", "Inline $x^2$ and\n\n$$\ny=1\n$$\n",
		"<p>Inline $x^2$ and</p>\n<p>$$\ny=1\n$$</p>\n"},
}

func TestToHTML_Rendering(t *testing.T) {
	for _, c := range renderCases {
		t.Run(c.name, func(t *testing.T) {
			doc := &CommonMark{Text: c.src}
			got, err := doc.ToHTML()
			if err != nil {
				t.Fatalf("ToHTML() error: %v", err)
			}
			if got != c.want {
				t.Errorf("ToHTML()\n got: %q\nwant: %q", got, c.want)
			}
		})
	}
}

/* Without html.WithUnsafe, ToHTML drops raw HTML; ToUnsafeHTML keeps it. */
func TestToHTML_RawHTMLOmitted(t *testing.T) {
	src := "<div class=\"x\">hi</div>\n\n<script>alert(1)</script>\n"
	doc := &CommonMark{Text: src}
	safe, err := doc.ToHTML()
	if err != nil {
		t.Fatal(err)
	}
	if want := "<!-- raw HTML omitted -->\n<!-- raw HTML omitted -->\n"; safe != want {
		t.Errorf("ToHTML()\n got: %q\nwant: %q", safe, want)
	}
	unsafe, err := doc.ToUnsafeHTML()
	if err != nil {
		t.Fatal(err)
	}
	if want := "<div class=\"x\">hi</div>\n<script>alert(1)</script>\n"; unsafe != want {
		t.Errorf("ToUnsafeHTML()\n got: %q\nwant: %q", unsafe, want)
	}
}

/* Everything except raw HTML renders the same through both entry points. */
func TestToUnsafeHTML_MatchesToHTMLWithoutRawHTML(t *testing.T) {
	for _, c := range renderCases {
		t.Run(c.name, func(t *testing.T) {
			doc := &CommonMark{Text: c.src}
			got, err := doc.ToUnsafeHTML()
			if err != nil {
				t.Fatalf("ToUnsafeHTML() error: %v", err)
			}
			if got != c.want {
				t.Errorf("ToUnsafeHTML()\n got: %q\nwant: %q", got, c.want)
			}
		})
	}
}

func TestToHTML_MathJax(t *testing.T) {
	src := "Inline $x^2$ and\n\n$$\ny=1\n$$\n"
	want := "<p>Inline <span class=\"math inline\">\\(x^2\\)</span> and</p>\n<p><span class=\"math display\">\\[y=1\n\\]</span></p>\n"
	doc := &CommonMark{Text: src}
	doc.UseMathJax(true)
	for name, render := range map[string]func() (string, error){"ToHTML": doc.ToHTML, "ToUnsafeHTML": doc.ToUnsafeHTML} {
		got, err := render()
		if err != nil {
			t.Fatalf("%s() error: %v", name, err)
		}
		if got != want {
			t.Errorf("%s()\n got: %q\nwant: %q", name, got, want)
		}
	}
	doc.UseMathJax(false)
	got, _ := doc.ToHTML()
	if strings.Contains(got, "math inline") {
		t.Errorf("MathJax markup present after UseMathJax(false): %q", got)
	}
}

/* Markdown with front matter: Parse splits it off and only the body renders. */
func TestToHTML_BodyAfterFrontMatter(t *testing.T) {
	doc := &CommonMark{}
	if err := doc.Parse([]byte("---\ntitle: T\n---\n\n# Hello\n")); err != nil {
		t.Fatal(err)
	}
	got, err := doc.ToHTML()
	if err != nil {
		t.Fatal(err)
	}
	if want := "<h1>Hello</h1>\n"; got != want {
		t.Errorf("ToHTML()\n got: %q\nwant: %q", got, want)
	}
}

func TestToHTML_Empty(t *testing.T) {
	got, err := (&CommonMark{}).ToHTML()
	if err != nil || got != "" {
		t.Errorf("ToHTML() on empty text = %q, %v; want empty, nil", got, err)
	}
}
