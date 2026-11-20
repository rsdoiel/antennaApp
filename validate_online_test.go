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
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/
package antennaApp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// DR-0017 1, 5: validate makes no request unless -online is given, and with it
// asks each sub-site for its table of contents the way BonBon does.
func TestValidateOnlineChecksSubsites(t *testing.T) {
	var requests int32
	var accept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		accept = r.Header.Get("Accept")
		switch r.URL.Path {
		case "/up/.well-known/toc.opml":
			w.Header().Set("Content-Type", "text/x-opml")
			w.Write([]byte(`<opml version="2.0"><head/><body/></opml>`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	goodSite(t)
	editConfig(t, "base_url:", "subsites:\n    - url: "+srv.URL+"/up\n    - url: "+srv.URL+"/gone\nbase_url:")
	if err := runIn(t, "generate"); err != nil {
		t.Fatal(err)
	}

	// offline: no request, no subsite finding
	out, err := runValidate(t)
	if n := atomic.LoadInt32(&requests); n != 0 {
		t.Errorf("validate without -online made %d requests", n)
	}
	if strings.Contains(out, "subsite-unreachable") || err != nil {
		t.Errorf("offline run: %v\n%s", err, out)
	}

	// online: the answering sub-site is fine, the 404 is a warning
	out, err = runValidate(t, "-online")
	wantClass(t, "an unreachable sub-site", err, classNegative)
	line, ok := finding(out, "antenna.yaml", "warning", "subsite-unreachable")
	if !ok || !strings.Contains(line, "/gone") || !strings.Contains(line, "404") {
		t.Errorf("no subsite-unreachable for /gone: %q\n%s", line, out)
	}
	if strings.Count(out, "subsite-unreachable") != 1 {
		t.Errorf("the answering sub-site was reported:\n%s", out)
	}
	if !strings.Contains(accept, "text/x-opml") {
		t.Errorf("the request did not ask for OPML like BonBon: Accept %q", accept)
	}
}

// A server that cannot be reached at all is also a finding, not a crash.
func TestValidateOnlineConnectionRefusedIsAFinding(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	addr := srv.URL
	srv.Close() // nothing listens there now
	goodSite(t)
	editConfig(t, "base_url:", "subsites:\n    - url: "+addr+"/site\nbase_url:")
	if err := runIn(t, "generate"); err != nil {
		t.Fatal(err)
	}
	out, err := runValidate(t, "-online")
	wantClass(t, "connection refused", err, classNegative)
	if _, ok := finding(out, "antenna.yaml", "warning", "subsite-unreachable"); !ok {
		t.Errorf("no subsite-unreachable:\n%s", out)
	}
}
