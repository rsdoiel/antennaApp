/*
antennaApp is a package for creating and curating blog, link blogs and social websites
Copyright (C) 2025 R. S. Doiel

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/
package antennaApp

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// A server that promises more body than it sends and drops the connection
// leaves a feed that was only partly read. webget used to ignore the read error
// (`if err != err` is never true) and hand the truncated bytes to the parser.
// The failure must be reported, as an unreachable service, not parsed.
func TestWebget_TruncatedBodyIsAReadFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.Header().Set("Content-Length", "100000")
		w.Write([]byte(`<?xml version="1.0"?><rss version="2.0"><channel><title>Cut</title></channel></rss>`))
		// Returning here closes the connection with the promised bytes missing.
	}))
	defer srv.Close()
	feed, err := webget("", srv.URL)
	if err == nil {
		t.Fatalf("a truncated body was accepted as feed %q", feed.Title)
	}
	wantClass(t, "truncated feed body", err, classUnavailable)
}
