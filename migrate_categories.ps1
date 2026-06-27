# Adds the categories column to the items table in an Antenna App database.
# Safe to run more than once — guards against re-migration.
#
# Usage: .\migrate_categories.ps1 -DB PATH_TO_DATABASE.db
param(
    [Parameter(Mandatory=$true)]
    [string]$DB
)

if (-not (Test-Path $DB)) {
    Write-Error "Database not found: $DB"
    exit 1
}

$exists = sqlite3 $DB `
    "SELECT COUNT(*) FROM pragma_table_info('items') WHERE name='categories';"

if ($exists -eq "0") {
    sqlite3 $DB "ALTER TABLE items ADD COLUMN categories JSON DEFAULT '';"
    Write-Host "Migrated: $DB"
} else {
    Write-Host "Already migrated: $DB"
}
