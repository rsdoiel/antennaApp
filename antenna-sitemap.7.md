sitemap — generate sitemap XML files

SYNOPSIS
  antenna sitemap

DESCRIPTION
  Generates a set of sitemap files (sitemap_index.xml, sitemap_1.xml, …)
  in the htdocs directory so search engines can discover your content.

  The sitemap lists the pages of your site, which come from the pages
  collection (pages.md), and your local posts, which are the items with a
  postPath in any collection. Items harvested from feeds are not pages of
  your site and are never listed, so a site that only aggregates feeds has
  no sitemap. When there is nothing to list, antenna says so on standard
  error, writes no files and exits 0. A collection whose database file is
  missing is reported and the others are still mapped, then the command
  exits 66.

EXAMPLE
  antenna sitemap

