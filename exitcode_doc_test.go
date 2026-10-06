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
	"fmt"
	"strings"
	"testing"
)

// The manual documents exactly the codes the table defines: an added, removed
// or renumbered class that the manual does not follow fails here.
func TestHelpText_ExitStatusDocumentsEveryClass(t *testing.T) {
	section := HelpText[strings.Index(HelpText, "# EXIT STATUS"):]
	if !strings.Contains(HelpText, "# EXIT STATUS") {
		t.Fatalf("HelpText has no EXIT STATUS section")
	}
	for _, c := range allExitClasses {
		row := fmt.Sprintf("| %d | %s |", c.Code, c.Name)
		if !strings.Contains(section, row) {
			t.Errorf("EXIT STATUS lacks a row for %s: want %q", c.Name, row)
		}
	}
}

func TestFailureTally(t *testing.T) {
	tally := &failureTally{}
	if tally.err("things") != nil {
		t.Errorf("an empty tally must report no error")
	}
	tally.add(nil)
	tally.add(dataErrorf("first bad one"))
	tally.add(nil)
	tally.add(unavailablef("second bad one"))
	err := tally.err("things")
	if err == nil {
		t.Fatal("two failures reported no error")
	}
	if !strings.HasPrefix(err.Error(), "2 of 4 things failed: ") {
		t.Errorf("expected the counts first, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "first bad one") {
		t.Errorf("expected the first failure in the message, got %q", err.Error())
	}
	// The class is the first failure's, not the last's.
	if got := ExitCodeFor(err); got != classData {
		t.Errorf("class %s, want data (the first failure)", got.Name)
	}
}
