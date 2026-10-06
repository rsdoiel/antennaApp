sitemap — generate sitemap XML files

SYNOPSIS
  antenna sitemap [-clean]

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

OPTIONS
  -clean   Remove the sitemap files an earlier run left behind. A site that
           shrank, or that stopped having anything to list, can have
           sitemap_N.xml files that the new sitemap_index.xml no longer names.
           With -clean, the regular files in the htdocs directory named
           exactly sitemap_NUMBER.xml that this run did not write are
           removed, and each removal is reported. When there is nothing to
           list, the old sitemap_index.xml is removed too, since nothing
           points at it. No other file is touched. If a collection could not
           be read nothing is removed, because the sitemap is then
           incomplete. Without -clean nothing is ever deleted.

EXAMPLE
  antenna sitemap
  antenna sitemap -clean

