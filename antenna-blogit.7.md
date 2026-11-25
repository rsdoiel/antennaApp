%antenna(7) user manual | version 0.2.1 c3213bd
% R. S. Doiel
% 2026-11-25

# NAME

blogit

# SYNOPSIS

antenna blogit [COLLECTION_NAME] FILEPATH [POST_DATE]

# DESCRIPTION

A variation of 'post' that places the source Markdown document in a
blog-style date directory tree (e.g. blog/2026/04/12/my-post.md) before
adding it to the collection. If POST_DATE is omitted, the current date is
used. The collection defaults to pages.md when COLLECTION_NAME is omitted.

Because blogit is for putting a post on the blog, it also makes the post
publishable. In the copy it makes (never in your original) it adds:

- pubDate, when the file has neither pubDate nor datePublished: the RSS 2.0 date
  form BonBon expects, with a numeric offset in your local time zone, for
  example "Thu, 19 Nov 2026 09:30:00 -0800". It is the current time, or, when
  POST_DATE is given, noon local on that day.
- title, when the file has none: the text of its first level-1 heading
  (# Heading). The body is not changed.

An existing pubDate, datePublished or title is never changed. If POST_DATE is
on a different day than the pubDate in the file, blogit warns and keeps the
file's date; the directory follows POST_DATE. It writes no author, description,
link, urlPrev or urlNext. A post with no pubDate stays a draft, so to keep one
as a draft use post, not blogit. See antenna help metadata.

The time and the day directory come from the time zone of the computer
antenna runs on, which the TZ environment variable overrides. If you work
over SSH on another machine or are travelling, say where you are:

    TZ=Asia/Tokyo antenna blogit my-post.md

A pubDate you wrote yourself, with its own offset, is kept as it is.

# PARAMETERS

COLLECTION_NAME
: (optional) path to the collection Markdown file

FILEPATH
: source Markdown document

POST_DATE
: (optional) date in YYYY-MM-DD format

# EXAMPLES

antenna blogit index.md notes/my-idea.md
antenna blogit index.md notes/my-idea.md 2026-04-01


