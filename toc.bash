#!/bin/bash
#
# toc.bash writes the site's table of contents, /.well-known/toc.opml, for a
# BonBon browser or another Markdown reader, from the Markdown pages named on
# the command line.
#
#   toc.bash BASE_URL SITE_TITLE PAGE.md [PAGE.md ...] > .well-known/toc.opml
#
# BASE_URL is where the site is published, ending in a slash. The pages go in
# the order given in a "Pages" group, each with the text of its first "# "
# heading and the address of its Markdown file; README.md is listed as index.md,
# the page a BonBon browser asks for at the site's address. A generated manual
# page (name.1.md, name-topic.7.md) goes in a "Manual pages" group after them,
# titled by its command ("antenna generate"), because its first heading is
# always "NAME". The output is the same for the same input (the date is the last
# commit's, not today's), so a rebuild changes nothing in git.
#
# Adapted from cmarkbrowser's toc.bash.
#
set -eu

if [ "$#" -lt 3 ]; then
	echo "usage: toc.bash BASE_URL SITE_TITLE PAGE.md [PAGE.md ...]" >&2
	exit 2
fi
BASE_URL="$1"
SITE_TITLE="$2"
shift 2

case "${BASE_URL}" in
http://*/ | https://*/) ;;
*)
	echo "toc.bash: BASE_URL must be an http or https address ending in /: ${BASE_URL}" >&2
	exit 2
	;;
esac

# xml_escape quotes the characters that matter in an XML attribute or text. It
# uses sed: in bash 5.2 an "&" in the replacement of ${s//x/y} means "the
# matched text", which turned "<" into "<lt;".
xml_escape() {
	printf '%s' "$1" | sed -e 's/&/\&amp;/g' -e 's/</\&lt;/g' -e 's/>/\&gt;/g' -e 's/"/\&quot;/g'
}

# The date of the last commit that touched a page, or the newest file time
# when there is no git history; never the current time.
modified() {
	local d
	d=$(git log -1 --format=%cD -- "$@" 2>/dev/null || true)
	if [ -z "${d}" ]; then
		d=$(LC_ALL=C date -u -r "$(ls -t "$@" | head -1)" '+%a, %d %b %Y %H:%M:%S GMT')
	fi
	printf '%s' "${d}"
}

# is_manual succeeds for a generated manual page: name.1.md, name-topic.7.md.
is_manual() {
	case "$1" in
	*.[0-9].md) return 0 ;;
	esac
	return 1
}

# title prints the text of the page's entry: its first "# " heading, or for a
# manual page its command, from the file name (antenna-markdown-web.7.md is
# "antenna markdown-web").
title() {
	local page="$1" t base
	if is_manual "${page}"; then
		base="${page%.md}"
		base="${base%.[0-9]}"
		printf '%s' "${base/-/ }"
		return
	fi
	t=$(sed -n 's/^# \{1,\}\(.*[^ ]\) *$/\1/p' "${page}" | head -1)
	if [ -z "${t}" ]; then
		# the title in the front matter, quoted or not
		t=$(awk 'NR == 1 && /^---[ \t]*$/ { fm = 1; next } fm && /^---[ \t]*$/ { exit } fm && /^title:/ { sub(/^title:[ \t]*/, ""); print; exit }' "${page}" |
			sed -e 's/^"\(.*\)"$/\1/' -e "s/^'\(.*\)'\$/\1/")
	fi
	if [ -z "${t}" ]; then
		base="${page%.md}"
		t="${base//_/ }"
	fi
	printf '%s' "${t}"
}

# group prints one group of outlines: the pages that are (want=manual) or are
# not (want=page) manual pages, in the order given. Nothing when there are none.
group() {
	local want="$1" name="$2" page published count=0
	shift 2
	for page in "$@"; do
		if is_manual "${page}"; then
			[ "${want}" = "manual" ] || continue
		else
			[ "${want}" = "page" ] || continue
		fi
		if [ "${count}" -eq 0 ]; then
			printf '\t\t<outline text="%s">\n' "${name}"
		fi
		count=$((count + 1))
		published="${page}"
		if [ "${page}" = "README.md" ]; then
			published="index.md"
		fi
		printf '\t\t\t<outline text="%s" url="%s%s"></outline>\n' \
			"$(xml_escape "$(title "${page}")")" "${BASE_URL}" "${published}"
	done
	if [ "${count}" -gt 0 ]; then
		printf '\t\t</outline>\n'
	fi
}

for page in "$@"; do
	if [ ! -f "${page}" ]; then
		echo "toc.bash: ${page}: no such page" >&2
		exit 66
	fi
done

printf '<?xml version="1.0" encoding="UTF-8"?>\n'
printf '<opml version="2.0">\n'
printf '\t<head>\n\t\t<title>%s</title>\n\t\t<dateModified>%s</dateModified>\n\t</head>\n' \
	"$(xml_escape "${SITE_TITLE}")" "$(modified "$@")"
printf '\t<body>\n'
group page "Pages" "$@"
group manual "Manual pages" "$@"
printf '\t</body>\n</opml>\n'
