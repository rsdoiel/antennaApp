
# TODO

Ideas, not quite a roadmap


## Release 0.0.28 blockers


## Bugs

- [x] The skip link class is `skip-link` (code, CSS, tests and DR-0002 all agree); no rename needed. (2026-10-05)
- [x] On aggregated pages the skip link jumped to the footer. `<main id="main-content">` is now `tabindex="-1"` so focus moves into it; checked on the live-site copy. Fixed in cc28293. (2026-10-05)

## Up Next


## Modernization (2026-11-16)

To match the workspace standard (knowledge and opml are the models; Go 1.27.3 across repositories). Do one repository at a time; run the tests and `make -n release` after each step.

- [ ] Go 1.27.3: change `go 1.26.4` in go.mod, the `Go >=` lines in codemeta.json, README.md and INSTALL.md; run `go mod tidy`, `go vet ./...` and the tests.
- [ ] Makefile: add `generate` (cmt for version.go, CITATION.cff, about.md, installers); make `build` stop running cmt. Model: opml/Makefile.
- [ ] Makefile: add `check-release` and make `release` run `generate` then `check-release`.
- [ ] Makefile: add `check-go-version` (go.mod, codemeta.json and the Go in use agree); `check-release` calls it.
- [ ] Drop the armv7l target and 32-bit Pi wording (dropped 2026-10-27).
- [ ] publish.bash/publish.ps1: take knowledge's checked versions (a failed pull no longer publishes nothing silently).
- [ ] release.bash: bring up to fountain's version.
- [ ] release.ps1: take knowledge's (existing-draft check, checksums).
- [ ] website.mak, page.tmpl, add-col-scope.lua, website.ps1: take knowledge's (check `NO_PUBLISH` and nav links).
