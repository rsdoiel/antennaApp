#!/bin/bash

function removeFile() {
    if [ -f "$1" ]; then
        rm "$1"
    fi
}

for FNAME in antenna.yaml page.yaml pages.md; do
    removeFile "${FNAME}"
done

ls -1 *.html *.db *.opml 2>/dev/null | while read -r FNAME; do
    removeFile "${FNAME}"
done
