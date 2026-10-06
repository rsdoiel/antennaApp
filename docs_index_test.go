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
	"os"
	"regexp"
	"strings"
	"testing"
)

// helpTopics returns the names under "Commands:" and "Reference:" in
// HelpTopicsText, the one list every manual index must agree with.
func helpTopics(t *testing.T) (commands, reference []string) {
	t.Helper()
	section := ""
	for _, line := range strings.Split(HelpTopicsText(), "\n") {
		switch {
		case strings.HasPrefix(line, "Commands:"):
			section = "commands"
		case strings.HasPrefix(line, "Reference:"):
			section = "reference"
		case strings.HasPrefix(line, "  ") && strings.TrimSpace(line) != "":
			name := strings.Fields(line)[0]
			if name == "topics" {
				continue
			}
			if section == "commands" {
				commands = append(commands, name)
			} else if section == "reference" {
				reference = append(reference, name)
			}
		}
	}
	if len(commands) == 0 || len(reference) == 0 {
		t.Fatalf("could not read the topic list: commands=%v reference=%v", commands, reference)
	}
	return commands, reference
}

func readDoc(t *testing.T, name string) string {
	t.Helper()
	src, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("%s: %s", name, err)
	}
	return string(src)
}

// antenna-topics.7.md is the printed topic list; it is a copy of
// HelpTopicsText and must not lag behind it.
func TestManualIndex_TopicsPageIsTheTopicList(t *testing.T) {
	got := strings.TrimSpace(readDoc(t, "antenna-topics.7.md"))
	want := strings.TrimSpace(HelpTopicsText())
	if got != want {
		t.Errorf("antenna-topics.7.md differs from HelpTopicsText(); regenerate it from `antenna help topics`.\nwant:\n%s\ngot:\n%s", want, got)
	}
}

// Every topic has a manual page, and the user manual links to it.
func TestManualIndex_UserManualLinksEveryTopic(t *testing.T) {
	commands, reference := helpTopics(t)
	manual := readDoc(t, "user_manual.md")
	for _, name := range append(append(commands, reference...), "topics") {
		page := "antenna-" + name + ".7.md"
		if _, err := os.Stat(page); err != nil {
			t.Errorf("topic %q has no manual page %s", name, page)
		}
		if !strings.Contains(manual, "["+name+"]("+page+")") {
			t.Errorf("user_manual.md does not link [%s](%s)", name, page)
		}
	}
}

// The ACTION section of the main manual describes every command.
func TestManualIndex_ActionListHasEveryCommand(t *testing.T) {
	commands, _ := helpTopics(t)
	man := readDoc(t, "antenna.1.md")
	a := strings.Index(man, "# ACTION")
	b := strings.Index(man, "# CONFIGURATION")
	if a < 0 || b < a {
		t.Fatalf("antenna.1.md has no ACTION section before CONFIGURATION")
	}
	action := man[a:b]
	for _, name := range commands {
		if !regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(name) + `( |$)`).MatchString(action) {
			t.Errorf("the ACTION section of antenna.1.md has no entry for %q", name)
		}
	}
}
