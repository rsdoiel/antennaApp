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
*/
package antennaApp

import (
	"bufio"
	//"database/sql"
	"errors"
	"fmt"
	"os"
	"path"
	//"strconv"
	"strings"
	"time"
	//"unicode"


	// 3rd Party
	//_ "github.com/glebarez/go-sqlite"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

/**
 * TUI
 */

/**
 * file handler
 */
type modelFile struct {
	filepicker   filepicker.Model
	selectedFile string
	quitting     bool
	err          error
}

type clearErrorMsg struct{}

func clearErrorAfter(t time.Duration) tea.Cmd {
	return tea.Tick(t, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

func (m modelFile) Init() tea.Cmd {
	return m.filepicker.Init()
}

func (m modelFile) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		}
	case clearErrorMsg:
		m.err = nil
	}

	var cmd tea.Cmd
	m.filepicker, cmd = m.filepicker.Update(msg)

	// Did the user select a file?
	if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
		// Get the path of the selected file.
		m.selectedFile = path
	}

	// Did the user select a disabled file?
	// This is only necessary to display an error to the user.
	if didSelect, path := m.filepicker.DidSelectDisabledFile(msg); didSelect {
		// Let's clear the selectedFile and display an error.
		m.err = errors.New(path + " is not valid.")
		m.selectedFile = ""
		return m, tea.Batch(cmd, clearErrorAfter(2*time.Second))
	}

	return m, cmd
}

func (m modelFile) View() string {
	if m.quitting {
		return ""
	}
	var s strings.Builder
	s.WriteString("\n  ")
	if m.err != nil {
		s.WriteString(m.filepicker.Styles.DisabledFile.Render(m.err.Error()))
	} else if m.selectedFile == "" {
		s.WriteString("Pick a file:")
	} else {
		s.WriteString("Selected file: " + m.filepicker.Styles.Selected.Render(m.selectedFile))
	}
	s.WriteString("\n\n" + m.filepicker.View() + "\n")
	return s.String()
}


/**
 * Curate Post
 */

/**
 * Curate Item
 */

/**
 * Curate Page
 */

/**
 * Curate Collection
 */

/**
 * Curate Collections
 */
var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type modelCollections struct {
	cfgName string
	cfg *AppConfig
	table table.Model
}

// addCollection prompts to find or create a collection Markdown file.
func (m modelCollections) addCollection() tea.Cmd {
	//FIXME: Need to add the file dialog here, or create a collection from scratch
	return tea.Batch(
		tea.Printf("FIXME: prompt for collection to add\n"),
	)
}

// delCollection prompts to delete collection Markdown file.
func (m modelCollections) delCollection() tea.Cmd {
	//FIXME: Need to delete the file dialog here
	return tea.Batch(
		tea.Printf("FIXME: prompt for collection to delete %s\n", m.table.SelectedRow()[1]),
	)
}

func (m modelCollections) Init() tea.Cmd { return nil }

func (m modelCollections) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "a":
			// Display add a collection dialog
			return m, tea.Sequence(
				tea.Printf("Add a collection goes here"),
				m.addCollection(),
			)
		case "d":
			// Display delete collection dialog
			return m, tea.Sequence(
				tea.Printf("Delete collection %s goes here", m.table.SelectedRow()[1]),
				m.delCollection(),
			)
		case "h":
			// Display Help screen
			return m, tea.Batch(
				tea.Printf("Display help for currating collections"),
			)
		case "enter":
			return m, tea.Batch(
				tea.Printf("Curate %s", m.table.SelectedRow()[1]),
			)
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m modelCollections) View() string {
	return baseStyle.Render(m.table.View()) + "\n [q]uit, [h]elp, [a]dd or [d]elete collections\n"
}

func curateCollections(cfgName string, cfg *AppConfig) error {
	columns := []table.Column{
		{Title: "#", Width: 2 },
		{Title: "Collection", Width: 16},
		{Title: "Title", Width: 48},
	}
	rows := []table.Row{}
	for i, collection := range cfg.Collections {
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", i+1),
			collection.File,
			collection.Title,
		})
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	m := modelCollections{
		cfgName: cfgName,
		cfg: cfg,
		table: t,
	}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		return fmt.Errorf("Error running program:", err)
	}
	return nil
}

/**
 * init a project
 */

// Interactive provides a simple terminal interface to curating collections,
// pages, posts and feed items for publication in your Antenna site.
func (app *AntennaApp) Interactive(cfgName string, args []string) error {
	if _, err := os.Stat(cfgName); os.IsNotExist(err) {
		wDir, err := os.Getwd()
		if err != nil {
			return err
		}
		fmt.Printf(`

This directory, %s
is not setup as an %s project. Initialize?
(type yes and press press enter or Ctrl-C to abort)

`, wDir, path.Base(os.Args[0]))
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if strings.HasPrefix(answer, "y") {
			if err := app.Init(cfgName, []string{}); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("aborting initialization")
		}
	}
	cfg := &AppConfig{}
	if err := cfg.LoadConfig(cfgName); err != nil {
		return err
	}
	if cfg.Collections == nil || len(cfg.Collections) == 0 {
		// NOTE: shouldn't see this unless you have a partially
		// initialized project
		return fmt.Errorf("no collections found in %s", cfgName)
	}
	if err := curateCollections(cfgName, cfg); err != nil {
		return err
	}
	return nil
 }

