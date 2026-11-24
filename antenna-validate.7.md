%antenna(7) user manual | version 0.1.1 930fbba
% R. S. Doiel
% 2026-11-20

# NAME

validate

# SYNOPSIS

antenna validate [-json] [-online] [-color WHEN] [PATH]

# DESCRIPTION

Checks Markdown for mistakes and for what Markdown readers such as BonBon
need, and reports each finding as one line:

    PATH:LINE: SEVERITY: CODE: message

validate only reads. It changes no file, and makes no network request unless
-online is given. With no PATH it checks the whole site: the pages, the
collection source files and the posts the site's own data names, plus the site
itself (the configuration, the table of contents, bonbon.json). A directory is
walked for .md and .markdown files, skipping folders whose name begins with a
dot and the items pages antenna generates. A single file is checked alone.
Checks that need to know what antenna renders are skipped when there is no
configuration file.

A finding has a severity: error, warning or info. The last line counts them.
Only errors and warnings affect the exit status.

# OPTIONS

-json
: Write one JSON object per finding (code, severity, file, line, message) and
a last object {"summary": {...}}. A failure of the command itself is also
written as an object with error, class and code.

-online
: Also ask each sub-site named in the subsites setting for its table of
contents, the way BonBon does, and warn about one that does not answer.

-color WHEN
: Color the report: the severity word (error red, warning yellow, info dim) and
the PATH:LINE. WHEN is auto (the default: color when standard output is a
terminal and NO_COLOR is not set), always or never. The text is the same with
or without color. -json is never colored.

# CHECKS

Each file:

front-matter
: The front matter cannot be read (error).

post-no-pubdate
: A file with a postPath and no pubDate is a draft; readers do not see it
(warning). antenna blogit adds the date.

date-unreadable
: A pubDate or datePublished antenna cannot read (error).

date-form
: pubDate is not the RSS 2.0 form BonBon expects (info).

no-title
: No title in the front matter and no level-1 heading (warning).

title-from-heading
: No title in the front matter, but a level-1 heading the title can come from
(info). antenna blogit and antenna page use that heading.

no-description
: No description, or an empty one; the table of contents and search results
use it (info).

description-long
: A description longer than descriptionMax characters (info). descriptionMax
is set in antenna.yaml and is 300 by default; 0 turns the check off.

no-author
: No author in the front matter and no author set in antenna.yaml (info).

postpath-mismatch
: postPath names a different file than the one it is in (warning).

nav-address
: urlPrev, urlNext or rightMarginImage is not an http, https, mailto or
relative address (error).

link-html
: A relative link to a .html page that antenna renders from Markdown. A
Markdown reader needs the .md link. The message says which (warning). A link to
a collection's page is named as its items Markdown when that file exists.

link-unrendered
: A relative .md link to a file that exists but that antenna does not render
(warning).

link-missing
: A relative link to a file that does not exist (warning).

link-absolute-self
: An absolute link to a page or file of this site, where a relative link works
for every reader (info).

raw-html
: Raw HTML, script or SVG, which Markdown readers strip (warning).

include-directive
: An @include-text-block or @include-code-block line in a Markdown file, which
a Markdown reader shows as text. Keep the source as a .txt file and write the
Markdown with antenna page (warning).

image-type
: An image that is not PNG, JPEG, GIF or WebP (warning).

macro-missing
: A {{page.NAME}} that names a key the page does not have (warning).

Code spans and code blocks are never looked into.

The whole site:

no-base-url
: There is no base_url (warning).

htdocs-missing
: htdocs is not a directory (error).

row-source-missing
: The pages table has a row for a file that does not exist (error). Remove it
with antenna unpage, or all of them with antenna prune.

markdown-only-generator
: A collection with html: false names a generator. No HTML is written for the
collection, so the theme is not used (warning).

alternate-static
: The generator YAML's link list holds a Markdown alternate (text/markdown or
application/markdown). The list is copied into every page, so the entry names
the same Markdown for all of them; antenna writes the right one for each page
(warning).

sequence-path-missing
: A path in the sequences setting of antenna.yaml is not there (warning).

sequence-empty
: A sequence has no members (info).

sequence-overlap
: A document is in more than one sequence and takes the neighbours of the
first (info).

row-output-missing
: The pages table records an output that is not there (warning). Run
antenna generate.

row-duplicate-output
: Two rows of the pages table write the same output, as after a rename
(warning). Remove the stale one with antenna unpage, or all of them with antenna prune.

include-missing
: A page's source includes a file that does not exist (warning). The directive
is left in the page as written.

output-stale
: An output is older than its source or a file it includes (info). Run
antenna generate.

not-published
: A page's or post's Markdown is not in htdocs, so the table of contents leaves
it out (warning).

toc-stale
: .well-known/toc.opml is missing or is not what generate would write (warning).

bonbon-json
: .well-known/bonbon.json differs from the bonbon setting, is not a JSON
object, or holds a known key BonBon would drop or break (warning).

nojekyll
: A github.io site served from the repository root has no .nojekyll file, so
GitHub Pages hides .well-known (warning).

subsite-unreachable
: With -online, a sub-site's table of contents does not answer 200 (warning).

# EXIT STATUS

0
: No errors or warnings (info does not count).

1
: At least one error or warning was found.

2
: The command line is wrong (an unknown option, more than one PATH).

66
: PATH does not exist, or there is no configuration file for a whole-site check.

78
: The configuration file is present but wrong.

# EXAMPLES

antenna validate
antenna validate blog
antenna validate blog/2026/11/19/my-post.md
antenna validate -online
antenna validate -json | jq 'select(.severity == "error")'

# SEE ALSO

antenna help markdown-web
antenna help metadata
antenna help blogit


