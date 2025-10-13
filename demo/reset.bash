#!/bin/bash

function removeFile() {
    if [ -f "$1" ]; then
        rm "$1"
    fi
}

for FNAME in antenna.yaml page.yaml; do
    removeFile "${FNAME}"
done

ls -1 *.html *.db 2>/dev/null | while read -r FNAME; do
    removeFile "${FNAME}"
done