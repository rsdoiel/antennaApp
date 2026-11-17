#!/bin/bash
#
# merge-collections.bash - merge one antennaApp collection (SRC) into
# another (DEST): the SQLite3 channels/items tables and the Markdown feed list.
#
# Usage: merge-collections.bash [--apply] SRC.md DEST.md
#   e.g. cd /Sites/antenna && merge-collections.bash --apply planet.md pacific.md
#
# Default is a dry run that changes nothing. With --apply it makes
# timestamped backups, then merges. It does NOT remove SRC from antenna.yaml,
# harvest or generate; it prints those commands for you to run.
#
# Tested design target: antennaApp 0.0.28, sqlite3 3.40.1.

set -euo pipefail

APPLY=no
if [ "${1:-}" = "--apply" ]; then APPLY=yes; shift; fi
if [ $# -ne 2 ]; then
    echo "usage: $0 [--apply] SRC.md DEST.md" >&2
    exit 2
fi
SRC_MD="$1"; DEST_MD="$2"
SRC_DB="${SRC_MD%.md}.db"; DEST_DB="${DEST_MD%.md}.db"

for f in "$SRC_MD" "$DEST_MD" "$SRC_DB" "$DEST_DB"; do
    [ -f "$f" ] || { echo "missing: $f" >&2; exit 66; }
done
command -v sqlite3 >/dev/null || { echo "sqlite3 not found" >&2; exit 66; }

# Shared columns, in DEST order, so a schema that drifted between versions
# does not misalign the copy. Columns only in SRC are reported and dropped.
cols() { sqlite3 "$1" "SELECT name FROM pragma_table_info('$2') ORDER BY cid;"; }
shared() {
    local t="$1" c
    c=$(comm -12 <(cols "$DEST_DB" "$t" | sort) <(cols "$SRC_DB" "$t" | sort))
    # keep DEST order
    cols "$DEST_DB" "$t" | grep -Fx -f <(echo "$c") | paste -sd, -
}
for t in channels items; do
    only=$(comm -13 <(cols "$DEST_DB" "$t" | sort) <(cols "$SRC_DB" "$t" | sort) | paste -sd, -)
    [ -z "$only" ] || echo "note: $SRC_DB $t has columns not in $DEST_DB (dropped): $only"
done
CH_COLS=$(shared channels); IT_COLS=$(shared items)
[ -n "$CH_COLS" ] && [ -n "$IT_COLS" ] || { echo "no shared columns" >&2; exit 65; }

count() { sqlite3 "$1" "SELECT count(*) FROM $2;"; }
echo "before: $DEST_DB channels=$(count "$DEST_DB" channels) items=$(count "$DEST_DB" items)"
echo "source: $SRC_DB channels=$(count "$SRC_DB" channels) items=$(count "$SRC_DB" items)"

# Rows that would be new (links not already in DEST)
NEW_CH=$(sqlite3 "$DEST_DB" "ATTACH '$SRC_DB' AS s; SELECT count(*) FROM s.channels WHERE link NOT IN (SELECT link FROM main.channels);")
NEW_IT=$(sqlite3 "$DEST_DB" "ATTACH '$SRC_DB' AS s; SELECT count(*) FROM s.items WHERE link NOT IN (SELECT link FROM main.items);")
echo "would add: channels=$NEW_CH items=$NEW_IT (rest already in $DEST_DB)"

# Feed URLs appearing in both Markdown files
urls() { grep -Eo 'https?://[^ )>"]+' "$1" | sort -u; }
DUPS=$(comm -12 <(urls "$SRC_MD") <(urls "$DEST_MD") || true)
if [ -n "$DUPS" ]; then
    echo "feed URLs in both $SRC_MD and $DEST_MD (will be listed twice):"
    echo "$DUPS" | sed 's/^/  /'
fi

if [ "$APPLY" != yes ]; then
    echo "dry run only; re-run with --apply to merge"
    exit 0
fi

# Refuse if another process holds a write lock (e.g. a harvest running)
sqlite3 "$DEST_DB" "BEGIN IMMEDIATE; ROLLBACK;" || { echo "$DEST_DB is locked" >&2; exit 75; }

STAMP=$(date +%Y%m%d-%H%M%S)
for f in "$SRC_MD" "$DEST_MD" "$SRC_DB" "$DEST_DB"; do cp -p "$f" "$f.$STAMP.bak"; done
[ -f antenna.yaml ] && cp -p antenna.yaml "antenna.yaml.$STAMP.bak"
echo "backups: *.$STAMP.bak"

sqlite3 "$DEST_DB" <<SQL
ATTACH '$SRC_DB' AS s;
BEGIN;
INSERT OR IGNORE INTO main.channels ($CH_COLS) SELECT $CH_COLS FROM s.channels;
INSERT OR IGNORE INTO main.items ($IT_COLS) SELECT $IT_COLS FROM s.items;
COMMIT;
SQL

# Markdown: append SRC body, minus its YAML front matter, under a marker comment.
{
    printf '\n\n# Merged from %s on %s\n\n' "$SRC_MD" "$(date +%Y-%m-%d)"
    awk 'NR==1 && $0=="---" {fm=1; next} fm && $0=="---" {fm=0; next} !fm' "$SRC_MD"
} >> "$DEST_MD"

echo "after:  $DEST_DB channels=$(count "$DEST_DB" channels) items=$(count "$DEST_DB" items)"
sqlite3 "$DEST_DB" "PRAGMA integrity_check;"
cat <<EOF

Merged. Next, by hand:
  antenna del $SRC_MD          # drop the collection from antenna.yaml (keeps files)
  antenna harvest $DEST_MD
  antenna generate
  antenna preview              # check, then remove old $SRC_MD output/redirects in htdocs
To undo: copy the *.$STAMP.bak files back over the originals.
EOF
