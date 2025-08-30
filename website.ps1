<#
# PowerShell script to replicate the Makefile for AntennaApp
#>
$PROJECT = "AntennaApp"
$PANDOC = (Get-Command pandoc -ErrorAction SilentlyContinue).Source
$MD_PAGES = Get-ChildItem -Path *.md
$HTML_PAGES = $MD_PAGES | ForEach-Object { $_.BaseName + ".html" }

function Invoke-Build {
    param(
        [string]$Target
    )

    switch ($Target) {
        "build" {
            # Build HTML pages
            foreach ($html in $HTML_PAGES) {
                $md = $html -replace '\.html$', '.md'
                if (Test-Path $PANDOC) {
                    & $PANDOC --metadata title=$html -s --to html5 $md -o $html `
                        --lua-filter=links-to-html.lua `
                        --template=page.tmpl
                }
                if ($html -eq "README.html") {
                    if (Test-Path "index.html") {
                        Remove-Item -Path "index.html" -Force
                    }
                    Rename-Item -Path "README.html" -NewName "index.html"
                }
            }

            # Run pagefind
            if (Get-Command pagefind -ErrorAction SilentlyContinue) {
                pagefind --verbose --glob="{*.html,docs/*.html}" --force-language en-US `
                    --exclude-selectors="nav,header,footer" --output-path ./pagefind --site .
                git add pagefind
            }
        }
        "clean" {
            # Clean up HTML files
            Remove-Item -Path *.html -Force -ErrorAction SilentlyContinue
        }
        default {
            Write-Host "Usage: .\build.ps1 [build|clean]"
        }
    }
}

# Parse command line arguments
$target = if ($args.Count -gt 0) { $args[0] } else { "build" }
Invoke-Build -Target $target
