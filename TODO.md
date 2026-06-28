
# TODO

Ideas, not quite a roadmap

## Release 0.0.25 blockers

- [X] rss.go: wire `categories` through to `WriteItemRSS` and emit `<category>` elements in RSS output — currently scanned from DB but discarded
- [X] post.go:229: replace `fmt.Sprintf(SQLRssDateRangePosts, fromDate, toDate)` with parameterized query to eliminate SQL injection surface; `fromDate`/`toDate` come from CLI args
- [X] webserver.go:508: replace `bytes.Compare(secret.Key, u.Key)` with `subtle.ConstantTimeCompare` to prevent timing attacks on password hash comparison
- [X] webserver.go: replace all `ioutil.ReadFile` / `ioutil.WriteFile` calls with `os.ReadFile` / `os.WriteFile` (ioutil is deprecated since Go 1.16)

## Bugs

- [X] helptext.go should hold all the help constants needed for topic guides an man pages, currently many topics are hard coded into help_dispatch.go, these need to be migrated to helptext.go
- [X] help topics are help guides, they need to start with the Markdown manpage header used for chapter seven man pages, example ```%{app_name}(7) user manual | version {version} {release_hash}
% R. S. Doiel
% {release_date}```
- [ ] Man page text that is themes chapter 5 should be merged into the themes for chapter 7

## Up Next

