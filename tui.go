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
	//"database/sql"
	"fmt"
	"os"
	"path/filepath"

	// 3rd Party
	//_ "github.com/glebarez/go-sqlite"

	//"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

/**
 * TUI
 */
 
 const (
 	actionInitialize = iota
 	actionCollections
 	actionCollection
 	actionItems
 	actionItem
 	actionPages
 	actionPage
 	actionPosts
 	actionPost
 )

var (
	baseStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

/**
 * Setup BubbleTea for managing the TUI
 */

/*
// Model, we need an interface to beable to pass specific model adjustments 
type Model interface {
    // Init is the first function that will be called. It returns an optional
    // initial command. To not perform an initial command return nil.
    Init() tea.Cmd

    // Update is called when a message is received. Use it to inspect messages
    // and, in response, update the model and/or send a command.
    Update(tea.Msg) (Model, tea.Cmd)

    // View renders the program's UI, which is just a string. The view is
    // rendered after every Update.
    View() string
}
*/

type errMsg struct{ err error }

// For messages that contain errors it's often handy to also implement the
// error interface on the message.
func (e errMsg) Error() string { return e.err.Error() }

type model struct {
	// terminal width
	width int
	// terminal hight
	height int
	
	// table for primary view navigation
	table table.Model

	// viewport holds views for help, individual items,
	// post and pages
	viewport viewport.Model

	// Choice (the current chosen view)
	Choice int
	err          error

	CfgName string
	Cfg *AppConfig
}

func (m model) Init() tea.Cmd {
	return nil
}


func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Handle resizing message
		m.height = msg.Height
		m.width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}
	// If we're not quitting then update using the right method
	switch m.Choice {
		case actionInitialize:
			return updateInit(msg, m)
		case actionCollections:
			return updateCollections(msg, m)
		/*
		case actionCollection:
			return updateCollection(msg, m)
		case actionItems:
			return updateItems(msg, m)
		case actionItem:
			return updateItem(m)
		case actionPages:
			return updatePages(m)
		case actionPage:
			return updatePage(m)
		case actionPosts:
			return updatePosts(m)
		case actionPost:
			return updatePost(m)
		*/
	}
	return m, nil
}

func (m model) View() string {
	// Handle the display for given model actions
	switch m.Choice {
		case actionInitialize:
			return viewInit(m)
		case actionCollections:
			return viewCollections(m)
		/*
		case actionCollection:
			return viewCollection(m)
		case actionItems:
			return viewItems(m)
		case actionItem:
			return viewItem(m)
		case actionPages:
			return viewPages(m)
		case actionPage:
			return viewPage(m)
		case actionPosts:
			return viewPosts(m)
		case actionPost:
			return viewPost(m)
		*/
	}
	return fmt.Sprintf("ERROR view %d not implememet", m.Choice)
}


/**
 * the following functions binds the model to specific view and update funcs
 */

func viewHelp(m model) string {
	const width = 78
	content := HelpText
	
	switch m.Choice {
	case actionCollections:
		content = collectionsTUIHelp
	}
	vp := viewport.New(width, 20)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		PaddingRight(2)

	// We need to adjust the width of the glamour render from our main width
	// to account for a few things:
	//
	//  * The viewport border width
	//  * The viewport padding
	//  * The viewport margins
	//  * The gutter glamour applies to the left side of the content
	//
	const glamourGutter = 2
	glamourRenderWidth := width - vp.Style.GetHorizontalFrameSize() - glamourGutter

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(glamourRenderWidth),
	)
	if err != nil {
		return fmt.Sprintf("%s\n", err)
	}

	str, err := renderer.Render(content)
	if err != nil {
		return fmt.Sprintf("%s\n", err)
	}

	vp.SetContent(str)
	return helpStyle.Render(m.viewport.View() + "\n[q]uit ↑/↓: Navigate\n")
}

func updateHelp(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		default:
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
	default:
		return m, nil
	}
}

func viewInit(m model) string {
	wDir, err := os.Getwd()
	if err != nil {
		wDir = "your current working directory"
	}
	appName := filepath.Base(os.Args[0])
	return fmt.Sprintf(`

%s is not a %s project directory.

Initialize (y/n)?

`, filepath.Base(wDir), appName)
}

func updateInit(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
			case "y", "Y", "enter":
				cfg, err := NewAppConfig(m.CfgName)
				if err != nil {
					m.err = err
					return m, nil
				} else {
					m.Cfg = cfg 
					// Now that we have a config the next view should
					// be the collections 
					m.Choice = actionCollections
					return m, nil
				}
			default:
				m.err = fmt.Errorf("aborting initialization")
				return m, tea.Quit
		}
	}
	return m, nil
}

func viewCollections(m model) string {
	columns := []table.Column{
		{Title: "#", Width: 2 },
		{Title: "Collection", Width: 16},
		{Title: "Title", Width: 48},
	}
	rows := []table.Row{}
	for i, collection := range m.Cfg.Collections {
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
	m.table = t
	if m.err != nil {
		return fmt.Sprintf("ERROR: %s\n", m.err) +
			baseStyle.Render(m.table.View()) +
		"\n[q]uit, [h]elp, [a]dd, [d]elete, [f]etch, [m]ake site, [c]onfigure\n"
	}
	return baseStyle.Render(m.table.View()) + "\n[q]uit, [h]elp, [a]dd, [d]elete, [f]etch, [m]ake site, [c]onfigure\n"
}

func updateCollections(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
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
			)
		case "d":
			// Display delete collection dialog
			return m, tea.Sequence(
				tea.Printf("Delete collection %s goes here", m.table.SelectedRow()[1]),
			)
		case "h":
			// Display Help screen
			return updateHelp(msg, m)
		case "enter":
			return m, tea.Batch(
				tea.Printf("Curate %s", m.table.SelectedRow()[1]),
			)
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// TUI provides a simple terminal interface to curating collections,
// pages, posts and feed items for publication in your Antenna site.
func (app *AntennaApp) TUI(cfgName string, args []string) error {
	// Figure out the initial view to use
	cfg := &AppConfig{}
	action := actionInitialize
	if _, err := os.Stat(cfgName); err == nil {
		action = actionCollections
		if err := cfg.LoadConfig(cfgName); err != nil {
			return err
		}
		if cfg.Collections == nil || len(cfg.Collections) == 0 {
			// NOTE: shouldn't see this unless you have a partially
			// initialized project
			return fmt.Errorf("no collections found in %s", cfgName)
		}
	}
	m := model {
		CfgName: cfgName,
		Cfg: cfg,
		Choice: action,
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
 }

