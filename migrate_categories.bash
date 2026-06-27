#!/usr/bin/env bash
# Adds the categories column to the items table in an Antenna App database.
# Safe to run more than once — guards against re-migration.
#
# Usage: migrate_categories.bash PATH_TO_DATABASE.db
set -euo pipefail

DB="${1:?usage: $0 PATH_TO_DATABASE.db}"

if [ ! -f "$DB" ]; then
    echo "Error: database not found: $DB" >&2
    exit 1
fi

EXISTS=$(sqlite3 "$DB" \
    "SELECT COUNT(*) FROM pragma_table_info('items') WHERE name='categories';")

if [ "$EXISTS" -eq "0" ]; then
    sqlite3 "$DB" "ALTER TABLE items ADD COLUMN categories JSON DEFAULT '';"
    echo "Migrated: $DB"
else
    echo "Already migrated: $DB"
fi
