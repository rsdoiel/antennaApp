//go:build cgo

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
	"errors"

	sqlite3 "github.com/mattn/go-sqlite3"
)

// mattnSQLiteCode returns the SQLite result code carried by an error from the
// mattn/go-sqlite3 driver. That driver is cgo-only: the sqlite3.Error type
// does not exist in a build with cgo off (every cross-compile), so naming it
// is confined to this file and exitcode_nocgo.go answers "not one" instead.
func mattnSQLiteCode(err error) (int, bool) {
	var e sqlite3.Error
	if errors.As(err, &e) {
		return int(e.Code), true
	}
	return 0, false
}
