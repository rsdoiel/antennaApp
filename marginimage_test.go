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
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var marginImgRE = regexp.MustCompile(`<img class="right-margin-image"[^>]*>`)

func TestRightMarginImageWithAltText(t *testing.T) {
	_, eout := navSite(t, map[string]string{"one": "rightMarginImage: img/cat.jpg\nrightMarginImageAlt: A cat on a desk\n"})
	got := readNavPage(t, "one.html")
	img := marginImgRE.FindString(got)
	if img == "" {
		t.Fatalf("no right-margin-image in:\n%s", got)
	}
	for _, want := range []string{`src="img/cat.jpg"`, `alt="A cat on a desk"`} {
		if !strings.Contains(img, want) {
			t.Errorf("%s lacks %s", img, want)
		}
	}
	// inside <main>, ahead of the page's own content, so it floats beside the top of the text
	i, m, h, e := strings.Index(got, "<img class=\"right-margin-image\""), strings.Index(got, "<main"), strings.Index(got, "<h1>one"), strings.Index(got, "</main>")
	if !(m < i && i < h && h < e) {
		t.Errorf("image not inside <main> before the content (main %d, img %d, h1 %d, /main %d)", m, i, h, e)
	}
	if strings.Contains(eout, "warning") {
		t.Errorf("warnings: %s", eout)
	}
	head := got[:strings.Index(got, "<body>")]
	if strings.Contains(head, "rightMarginImage") {
		t.Errorf("keys also written as <meta>:\n%s", head)
	}
}

// BonBon's key carries no alt text: without one the image is decorative, an
// empty alt, and nothing is invented and nothing is reported (DR-0009 3).
func TestRightMarginImageWithoutAltIsDecorative(t *testing.T) {
	_, eout := navSite(t, map[string]string{"one": "rightMarginImage: img/cat.jpg\n"})
	img := marginImgRE.FindString(readNavPage(t, "one.html"))
	if !strings.Contains(img, `alt=""`) {
		t.Errorf("want an empty alt, got %s", img)
	}
	if strings.Contains(eout, "warning") {
		t.Errorf("a missing alt must not warn: %s", eout)
	}
}

func TestRightMarginImageAltAloneShowsNothing(t *testing.T) {
	navSite(t, map[string]string{"one": "rightMarginImageAlt: orphan\n"})
	if got := readNavPage(t, "one.html"); strings.Contains(got, "right-margin-image") || strings.Contains(got, "orphan") {
		t.Errorf("alt without an image rendered something:\n%s", got)
	}
}

func TestRightMarginImageEscapesAndRefusesBadSources(t *testing.T) {
	_, eout := navSite(t, map[string]string{
		"one":   "rightMarginImage: \"https://x.example/a.jpg?w=1&h=2\"\nrightMarginImageAlt: 'He said \"hi\" <b>'\n",
		"two":   "rightMarginImage: \"javascript:alert(1)\"\n",
		"three": "rightMarginImage: \"data:image/svg+xml,<svg onload=alert(1)>\"\n",
	})
	one := marginImgRE.FindString(readNavPage(t, "one.html"))
	for _, want := range []string{`src="https://x.example/a.jpg?w=1&amp;h=2"`, `alt="He said &#34;hi&#34; &lt;b&gt;"`} {
		if !strings.Contains(one, want) {
			t.Errorf("%s lacks %s", one, want)
		}
	}
	for _, name := range []string{"two.html", "three.html"} {
		if p := readNavPage(t, name); strings.Contains(p, "right-margin-image") || strings.Contains(p, "javascript:") || strings.Contains(p, "onload") {
			t.Errorf("%s used a script or data source:\n%s", name, p)
		}
	}
	if strings.Count(eout, "rightMarginImage") < 2 {
		t.Errorf("want a warning for each refused source: %q", eout)
	}
}

// Posts are rendered by Post and by GeneratePosts: both show the image.
func TestPostsGetRightMarginImage(t *testing.T) {
	cfg := sitemapSite(t, "news.md")
	cfg.Htdocs = "htdocs"
	cfg.Generator = "page.yaml"
	cfg.Collections[0].Generator = "page.yaml"
	if err := os.WriteFile("page.yaml", []byte("header: |\n  <h1>Test</h1>\n"), 0664); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll("htdocs/p", 0775); err != nil {
		t.Fatal(err)
	}
	src := "---\ntitle: Two\npostPath: p/two.md\npubDate: \"2026-11-19\"\nrightMarginImage: ../img/a.jpg\nrightMarginImageAlt: An apple\n---\nBody.\n"
	if err := os.WriteFile("two.md", []byte(src), 0664); err != nil {
		t.Fatal(err)
	}
	if err := cfg.Post("news.md", "two.md"); err != nil {
		t.Fatalf("Post: %s", err)
	}
	for _, step := range []string{"Post", "GeneratePosts"} {
		if step == "GeneratePosts" {
			os.Remove("htdocs/p/two.html")
			var eout bytes.Buffer
			if err := cfg.Collections[0].GeneratePosts(&eout, "antenna", cfg); err != nil {
				t.Fatalf("GeneratePosts: %s\n%s", err, eout.String())
			}
		}
		b, err := os.ReadFile(filepath.Join("htdocs", "p", "two.html"))
		if err != nil {
			t.Fatalf("%s: %s", step, err)
		}
		img := marginImgRE.FindString(string(b))
		if !strings.Contains(img, `src="../img/a.jpg"`) || !strings.Contains(img, `alt="An apple"`) {
			t.Errorf("%s: img = %q", step, img)
		}
	}
}

func TestGenerateCSS_FloatsTheRightMarginImage(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &AppConfig{Htdocs: tmpDir}
	var buf bytes.Buffer
	if err := cfg.GenerateCSS(&buf, "css/site.css"); err != nil {
		t.Fatalf("GenerateCSS: %s", err)
	}
	content, _ := os.ReadFile(filepath.Join(tmpDir, "css", "site.css"))
	css := string(content)
	i := strings.Index(css, ".right-margin-image")
	if i < 0 || !strings.Contains(css[i:], "float: right") {
		t.Errorf("stylesheet does not float .right-margin-image")
	}
	// on a narrow screen the float is dropped so text keeps its width
	if !strings.Contains(css[i:], "@media (max-width") {
		t.Errorf("no narrow-screen rule for .right-margin-image")
	}
}
