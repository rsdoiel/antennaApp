package antennaApp

import (
	"bytes"
	"database/sql"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// DR-0037: a relative href or src in a harvested item's body is resolved
// against the item's link when the item is rendered. These tests cover the
// resolver, then the real verbs (harvest, then generate) on a feed whose
// item bodies carry relative references.

const itemLink = "https://example.org/2026/a/"

func TestResolveRef(t *testing.T) {
	for name, c := range map[string]struct {
		ref, want string
	}{
		"root-relative":     {"/images/x.png", "https://example.org/images/x.png"},
		"root path":         {"/a/b", "https://example.org/a/b"},
		"page-relative":     {"c", "https://example.org/2026/a/c"},
		"dot-dot":           {"../c", "https://example.org/2026/c"},
		"protocol-relative": {"//cdn.example.net/x.js", "https://cdn.example.net/x.js"},
		"query only":        {"?page=2", "https://example.org/2026/a/?page=2"},
		"absolute":          {"https://other.example/", "https://other.example/"},
		"http absolute":     {"http://other.example/", "http://other.example/"},
		"mailto":            {"mailto:x@y.example", "mailto:x@y.example"},
		"tel":               {"tel:+15551234", "tel:+15551234"},
		"data":              {"data:image/png;base64,AAAA", "data:image/png;base64,AAAA"},
		"fragment only":     {"#top", "#top"},
		"empty":             {"", ""},
		"javascript":        {"javascript:void(0)", "javascript:void(0)"},
	} {
		t.Run(name, func(t *testing.T) {
			got, _ := resolveRef(itemLink, c.ref)
			if got != c.want {
				t.Errorf("resolveRef(%q, %q) = %q, want %q", itemLink, c.ref, got, c.want)
			}
		})
	}
}

func TestResolveRefNeedsAnAbsoluteHTTPBase(t *testing.T) {
	for _, base := range []string{"", "/relative/link", "mailto:x@y.example", "not a url", "ftp://example.org/a"} {
		got, ok := resolveRef(base, "/a/b")
		if ok || got != "/a/b" {
			t.Errorf("base %q: got (%q, %v), want the reference unchanged and not ok", base, got, ok)
		}
	}
}

func TestResolveHTMLRefs(t *testing.T) {
	in := `<p>See <a href="/a/b" class="x">this</a>, <a href="c">that</a>, <a href="https://o.example/">o</a>,
<a href="mailto:a@b.example">m</a> and <a href="#top">top</a>.</p>
<img src="/images/x.png" alt="x &amp; y"><img src="data:image/png;base64,AAAA">
<pre>&lt;a href="/not-a-link"&gt;</pre>`
	got, n := resolveHTMLRefs(in, itemLink)
	for _, want := range []string{
		`href="https://example.org/a/b" class="x"`,
		`href="https://example.org/2026/a/c"`,
		`href="https://o.example/"`,
		`href="mailto:a@b.example"`,
		`href="#top"`,
		`src="https://example.org/images/x.png" alt="x &amp; y"`,
		`src="data:image/png;base64,AAAA"`,
		`&lt;a href="/not-a-link"&gt;`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if n != 3 {
		t.Errorf("rewrote %d references, want 3", n)
	}
	if got2, n2 := resolveHTMLRefs(got, itemLink); got2 != got || n2 != 0 {
		t.Errorf("a second pass changed the text (%d)", n2)
	}
}

func TestResolveMarkdownRefs(t *testing.T) {
	in := "A [link](/a/b \"Title\"), an ![image](/images/x.png), a [rel](c),\n" +
		"an [angle](</a b/c>), [abs](https://o.example/) and [frag](#top).\n\n" +
		"A `[code](/not/a/link)` span and a [ref][1].\n\n" +
		"```\n[fenced](/not/a/link)\n```\n\n" +
		"[1]: /a/ref\n"
	got, n := resolveMarkdownRefs(in, itemLink)
	for _, want := range []string{
		"[link](https://example.org/a/b \"Title\")",
		"![image](https://example.org/images/x.png)",
		"[rel](https://example.org/2026/a/c)",
		"[angle](<https://example.org/a%20b/c>)",
		"[abs](https://o.example/)",
		"[frag](#top)",
		"`[code](/not/a/link)`",
		"[fenced](/not/a/link)",
		"[1]: https://example.org/a/ref",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if n != 5 {
		t.Errorf("rewrote %d references, want 5", n)
	}
}

// feedWithRelativeBodies is a feed whose items carry relative references,
// the case found on the antenna site, 2026-11-23.
const feedWithRelativeBodies = `<?xml version="1.0"?><rss version="2.0"><channel><title>Rel</title><link>https://example.org/</link>
<item><title>One</title><link>https://example.org/2026/a/</link><guid>one</guid><pubDate>Mon, 23 Nov 2026 10:00:00 GMT</pubDate>
<description><![CDATA[<p>See <a href="/a/b">root</a>, <a href="c">page</a>, <a href="https://other.example/">abs</a>,
<a href="mailto:x@y.example">mail</a> and <a href="#top">top</a>.</p><p><img src="/images/x.png" alt="pic"></p>]]></description></item>
</channel></rss>`

func harvestedRelativeSite(t *testing.T) (htdocsDir string) {
	t.Helper()
	srv, _ := fakeSite(map[string]route{
		"/feed.xml": {"application/rss+xml", feedWithRelativeBodies},
	})
	t.Cleanup(srv.Close)
	harvestSite(t, "[Rel]("+srv.URL+"/feed.xml)")
	if _, _, err := runCapture(t, "harvest", "news.md"); err != nil {
		t.Fatalf("harvest: %s", err)
	}
	return "."
}

func storedDescription(t *testing.T) string {
	t.Helper()
	db, err := sql.Open("sqlite", "news.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var d string
	if err := db.QueryRow(`SELECT description FROM items`).Scan(&d); err != nil {
		t.Fatal(err)
	}
	return d
}

func findOutput(t *testing.T, name string) string {
	t.Helper()
	for _, dir := range []string{".", "htdocs"} {
		if b, err := os.ReadFile(filepath.Join(dir, name)); err == nil {
			return string(b)
		}
	}
	t.Fatalf("%s was not written", name)
	return ""
}

// The reported bug, through the path it took: harvest, then generate. All
// three outputs carry the item's address, and the store does not change.
func TestGenerateResolvesRelativeLinksOfHarvestedItems(t *testing.T) {
	harvestedRelativeSite(t)
	before := storedDescription(t)
	if _, _, err := runCapture(t, "generate"); err != nil {
		t.Fatalf("generate: %s", err)
	}
	for _, c := range []struct {
		file string
		want []string
	}{
		{"news_items.html", []string{
			`href="https://example.org/a/b"`,
			`href="https://example.org/2026/a/c"`,
			`src="https://example.org/images/x.png"`,
			`href="https://other.example/"`, `href="mailto:x@y.example"`, `href="#top"`}},
		{"news_items.md", []string{
			"(https://example.org/a/b)",
			"(https://example.org/2026/a/c)",
			"(https://other.example/)", "(mailto:x@y.example)", "(#top)"}},
		{"news.xml", []string{
			`href="https://example.org/a/b"`,
			`src="https://example.org/images/x.png"`,
			`href="#top"`}},
	} {
		got := findOutput(t, c.file)
		for _, want := range c.want {
			if !strings.Contains(got, want) {
				t.Errorf("%s lacks %q", c.file, want)
			}
		}
		for _, bad := range []string{`href="/a/b"`, `src="/images/x.png"`, "](/a/b)", `href="c"`} {
			if strings.Contains(got, bad) {
				t.Errorf("%s still has %q", c.file, bad)
			}
		}
	}
	if after := storedDescription(t); after != before {
		t.Errorf("generate changed the stored item:\n%s\n--- to ---\n%s", before, after)
	}
	if !strings.Contains(before, `href="/a/b"`) {
		t.Errorf("harvest did not store the feed's text as written:\n%s", before)
	}
}

// A second generate with nothing changed leaves the files alone (DR-0011).
func TestGenerateResolvedLinksAreStable(t *testing.T) {
	harvestedRelativeSite(t)
	if _, _, err := runCapture(t, "generate"); err != nil {
		t.Fatal(err)
	}
	first := findOutput(t, "news_items.html")
	if _, _, err := runCapture(t, "generate"); err != nil {
		t.Fatal(err)
	}
	if findOutput(t, "news_items.html") != first {
		t.Errorf("a second generate changed the page")
	}
}

// Stored source:markdown is rewritten in the Markdown and in the HTML.
func TestItemsResolveRelativeLinksOfSourceMarkdown(t *testing.T) {
	md := "Read [the post](/2026/08/02/post/) and ![pic](/images/x.png) and [here](sibling).\n"
	_, htdocs := itemsMDSite(t, harvestedSource(itemLink, "One", "d", md, "2026-11-18 10:00:00", "published", "Feed A"))
	m, _ := os.ReadFile(filepath.Join(htdocs, "news_items.md"))
	h, _ := os.ReadFile(filepath.Join(htdocs, "news_items.html"))
	for _, want := range []string{"(https://example.org/2026/08/02/post/)", "(https://example.org/images/x.png)", "(https://example.org/2026/a/sibling)"} {
		if !strings.Contains(string(m), want) {
			t.Errorf("items.md lacks %q:\n%s", want, m)
		}
	}
	for _, want := range []string{`href="https://example.org/2026/08/02/post/"`, `src="https://example.org/images/x.png"`} {
		if !strings.Contains(string(h), want) {
			t.Errorf("items.html lacks %q", want)
		}
	}
}

// An item with no link has nothing to resolve against: its body is left as
// written and generate warns once for the collection, with the count.
func TestItemsWithoutAddressAreLeftAndReported(t *testing.T) {
	cfg := sitemapSite(t, "news.md")
	htdocs := filepath.Join(t.TempDir(), "htdocs")
	if err := os.MkdirAll(htdocs, 0775); err != nil {
		t.Fatal(err)
	}
	cfg.Htdocs = htdocs
	col := cfg.Collections[0]
	sitemapExec(t, "news.db",
		harvestedSource("no-link-one", "No link", "d", "See [x](/a/b).\n", "2026-11-18 10:00:00", "published", "Feed A"),
		harvestedSource("no-link-two", "Also none", "d", "See [y](/c/d).\n", "2026-11-19 10:00:00", "published", "Feed A"),
		harvestedSource(itemLink, "Fine", "d", "See [z](/e/f).\n", "2026-11-20 10:00:00", "published", "Feed A"))
	var eout bytes.Buffer
	if err := col.Generate(io.Discard, &eout, "antenna", cfg); err != nil {
		t.Fatalf("Generate: %s", err)
	}
	m, _ := os.ReadFile(filepath.Join(htdocs, "news_items.md"))
	if !strings.Contains(string(m), "(/a/b)") || !strings.Contains(string(m), "(/c/d)") {
		t.Errorf("a body with no address was changed:\n%s", m)
	}
	if !strings.Contains(string(m), "(https://example.org/e/f)") {
		t.Errorf("the item with an address was not resolved:\n%s", m)
	}
	warns := 0
	for _, line := range strings.Split(eout.String(), "\n") {
		if strings.Contains(line, "relative") && strings.Contains(line, "2 item") {
			warns++
		}
	}
	if warns != 1 {
		t.Errorf("want one warning naming 2 items, got %d:\n%s", warns, eout.String())
	}
}

// The site's own posts are not rewritten: their relative links are right.
func TestOwnPostsKeepRelativeLinksInTheFeed(t *testing.T) {
	var buf bytes.Buffer
	gen, err := NewGenerator("antenna", "https://example.com")
	if err != nil {
		t.Fatal(err)
	}
	err = gen.writeItemRSS(&buf, true, "https://example.com/p/", "Mine", `<p><a href="/about.html">about</a></p>`,
		nil, nil, "g", "2026-11-18", "", "", "published", "", "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `href="/about.html"`) {
		t.Errorf("an own post was rewritten:\n%s", buf.String())
	}
}
