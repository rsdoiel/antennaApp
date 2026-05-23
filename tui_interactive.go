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
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

// errCancelled is returned when the user aborts an action with "c" or ESC.
var errCancelled = fmt.Errorf("action cancelled")

// iHistoryEntry records one confirmed-and-run command within a session.
type iHistoryEntry struct {
	action string
	args   []string
}

// iSession holds the state for an interactive REPL session.
type iSession struct {
	app     *AntennaApp
	cfgName string
	scanner *bufio.Scanner
	history []iHistoryEntry
}

// ---------------------------------------------------------------------------
// Package-level helpers (no session state needed)
// ---------------------------------------------------------------------------

// iLine formats a single action entry for the help menu.
func iLine(name, desc string) string {
	return fmt.Sprintf("    %s%-18s%s %s\n", Yellow+Bold, name, Reset, desc)
}

// isCancelInput reports whether the raw line the user typed is a cancel signal.
// Recognised forms: the letter "c" (any case) or any ESC-prefixed sequence.
func isCancelInput(s string) bool {
	s = strings.TrimSpace(s)
	return strings.EqualFold(s, "c") || strings.HasPrefix(s, "\x1b")
}

// iStep prompts for a single parameter. defaultVal is shown in brackets; pressing
// Enter accepts it. Typing "c" or ESC returns errCancelled.
func iStep(scanner *bufio.Scanner, label, description, preVal, defaultVal string) (string, error) {
	fmt.Printf("\n%s%s%s\n", Bold+Cyan, label, Reset)
	fmt.Printf("  %s\n", description)
	if preVal != "" {
		fmt.Printf("  %sUsing:%s %s\n", Bold, Reset, preVal)
		return preVal, nil
	}
	if defaultVal != "" {
		fmt.Printf("  %s[%s]%s %s>%s ", Bold, defaultVal, Reset, Green+Bold, Reset)
	} else {
		fmt.Printf("  %s>%s ", Green+Bold, Reset)
	}
	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())
	if isCancelInput(input) {
		return "", errCancelled
	}
	if input == "" {
		return defaultVal, nil
	}
	return input, nil
}

// iOptionalStep is like iStep but makes clear the parameter is optional.
// Pressing Enter accepts the default (if any) or skips the parameter.
func iOptionalStep(scanner *bufio.Scanner, label, description, preVal, defaultVal string) (string, error) {
	fmt.Printf("\n%s%s%s %s(optional)%s\n", Bold+Cyan, label, Reset, Italic, Reset)
	fmt.Printf("  %s\n", description)
	if preVal != "" {
		fmt.Printf("  %sUsing:%s %s\n", Bold, Reset, preVal)
		return preVal, nil
	}
	if defaultVal != "" {
		fmt.Printf("  %s[%s]%s %s>%s ", Bold, defaultVal, Reset, Green+Bold, Reset)
	} else {
		fmt.Printf("  Press Enter to skip, or type a value.\n")
		fmt.Printf("  %s>%s ", Green+Bold, Reset)
	}
	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())
	if isCancelInput(input) {
		return "", errCancelled
	}
	if input == "" {
		return defaultVal, nil
	}
	return input, nil
}

// iShowCommand prints the command as it stands so the user can see what is
// being built.
func iShowCommand(appName, cfgName, action string, collectedArgs []string) {
	parts := []string{appName, "-config", cfgName, action}
	for _, a := range collectedArgs {
		if strings.ContainsAny(a, " \t") {
			parts = append(parts, fmt.Sprintf("%q", a))
		} else {
			parts = append(parts, a)
		}
	}
	fmt.Printf("\n  %sCommand:%s %s\n", Bold, Reset, strings.Join(parts, " "))
}

// ---------------------------------------------------------------------------
// Session methods
// ---------------------------------------------------------------------------

// confirmAndRun shows the completed command, asks for confirmation, records it
// to session history when confirmed, then runs it. Typing "c" or ESC returns
// errCancelled; "n" returns nil without running.
func (s *iSession) confirmAndRun(action string, args []string) error {
	iShowCommand(s.app.appName, s.cfgName, action, args)
	fmt.Printf("\n  %sRun this command?%s [Y/n] %s>%s ", Bold, Reset, Green+Bold, Reset)
	s.scanner.Scan()
	answer := strings.ToLower(strings.TrimSpace(s.scanner.Text()))
	if isCancelInput(answer) {
		fmt.Println("\n  Cancelled.")
		return errCancelled
	}
	if answer == "n" || answer == "no" {
		fmt.Println("\n  Cancelled.")
		return nil
	}
	fmt.Printf("\n%sRunning...%s\n\n", Bold, Reset)
	s.history = append(s.history, iHistoryEntry{action: action, args: args})
	return s.app.Run(os.Stdin, os.Stdout, os.Stderr, s.cfgName, action, args)
}

// shellLine renders one history entry as a shell command string.
func (s *iSession) shellLine(appName string, e iHistoryEntry) string {
	parts := []string{appName, "-config", s.cfgName, e.action}
	for _, a := range e.args {
		if strings.ContainsAny(a, " \t") {
			parts = append(parts, fmt.Sprintf("%q", a))
		} else {
			parts = append(parts, a)
		}
	}
	return strings.Join(parts, " ")
}

// showHistory prints the numbered list of commands run this session.
func (s *iSession) showHistory() {
	if len(s.history) == 0 {
		fmt.Printf("\n  No commands in history yet.\n")
		return
	}
	fmt.Printf("\n%sSession history:%s\n\n", Bold, Reset)
	for i, e := range s.history {
		parts := append([]string{e.action}, e.args...)
		fmt.Printf("  %2d.  %s%s%s\n", i+1, Cyan, strings.Join(parts, " "), Reset)
	}
}

// writeScript writes the session history as a shell script to w.
func (s *iSession) writeScript(w io.Writer, header []string, appName string) {
	for _, line := range header {
		fmt.Fprintln(w, line)
	}
	fmt.Fprintln(w)
	for _, e := range s.history {
		fmt.Fprintln(w, s.shellLine(appName, e))
	}
}

// exportToFile asks for a filename and writes the session history as a script.
func (s *iSession) exportToFile(shell string) error {
	if len(s.history) == 0 {
		fmt.Printf("\n  No commands in history to export.\n")
		return nil
	}
	defaultName := "antenna_session.sh"
	if shell == "powershell" {
		defaultName = "antenna_session.ps1"
	}
	fmt.Printf("\n  %sFilename%s [%s]: ", Bold+Cyan, Reset, defaultName)
	s.scanner.Scan()
	name := strings.TrimSpace(s.scanner.Text())
	if name == "" {
		name = defaultName
	}
	f, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("cannot create %q: %w", name, err)
	}
	defer f.Close()
	stamp := time.Now().Format("2006-01-02")
	if shell == "bash" {
		s.writeScript(f, []string{
			"#!/usr/bin/env bash",
			"# Generated by antenna interactive — " + stamp,
			"set -euo pipefail",
		}, s.app.appName)
		if err := os.Chmod(name, 0755); err != nil {
			return err
		}
	} else {
		s.writeScript(f, []string{
			"#!/usr/bin/env pwsh",
			"# Generated by antenna interactive — " + stamp,
			"$ErrorActionPreference = 'Stop'",
		}, s.app.appName+".exe")
	}
	fmt.Printf("\n  %sWrote %d command(s) to %s%s\n", Bold, len(s.history), name, Reset)
	return nil
}

// handleHistory processes the history sub-commands.
func (s *iSession) handleHistory(args []string) error {
	if len(args) == 0 {
		s.showHistory()
		return nil
	}
	switch strings.ToLower(args[0]) {
	case "bash", "sh":
		return s.exportToFile("bash")
	case "powershell", "ps1", "pwsh":
		return s.exportToFile("powershell")
	default:
		return fmt.Errorf("unknown format %q — use: history, history bash, history powershell", args[0])
	}
}

// showHelp prints the full action reference.
func (s *iSession) showHelp() {
	fmt.Printf("\n%sAvailable actions:%s\n\n", Bold, Reset)
	fmt.Printf("%sSetup%s\n", Bold, Reset)
	fmt.Print(iLine("init", "initialise a new antenna project"))
	fmt.Print(iLine("add", "register a feed collection"))
	fmt.Print(iLine("del", "remove a feed collection"))
	fmt.Print(iLine("list", "list registered collections"))
	fmt.Printf("\n%sContent%s\n", Bold, Reset)
	fmt.Print(iLine("post", "add a Markdown document as a post"))
	fmt.Print(iLine("blogit", "copy a file into a blog directory tree and post it"))
	fmt.Print(iLine("unpost", "remove a post from a collection"))
	fmt.Print(iLine("posts", "list posts in a collection"))
	fmt.Print(iLine("page", "add a standalone HTML page"))
	fmt.Print(iLine("unpage", "remove a page from a collection"))
	fmt.Print(iLine("pages", "list pages in a collection"))
	fmt.Printf("\n%sPublish%s\n", Bold, Reset)
	fmt.Print(iLine("harvest", "fetch content from feeds"))
	fmt.Print(iLine("generate", "build HTML and RSS output"))
	fmt.Print(iLine("rss", "generate an RSS feed file"))
	fmt.Print(iLine("sitemap", "generate sitemap XML files"))
	fmt.Print(iLine("preview", "serve the site on localhost for review"))
	fmt.Printf("\n%sThemes%s\n", Bold, Reset)
	fmt.Print(iLine("themes", "list available themes"))
	fmt.Print(iLine("apply", "apply a theme to a page generator"))
	fmt.Print(iLine("stylefrom", "extract CSS from a LibreOffice Writer file"))
	fmt.Printf("\n%sSession%s\n", Bold, Reset)
	fmt.Print(iLine("history", "show commands run this session"))
	fmt.Print(iLine("history bash", "export session to a Bash script"))
	fmt.Print(iLine("history powershell", "export session to a PowerShell script"))
	fmt.Print(iLine("quit", "exit the interactive session"))
	fmt.Printf("\n  Type %sc%s or %sESC%s at any prompt to cancel the current action.\n", Bold, Reset, Bold, Reset)
}

// guide dispatches to the action-specific guided workflow.
func (s *iSession) guide(action string, args []string) error {
	switch action {
	case "init":
		return s.guideInit(args)
	case "add":
		return s.guideAdd(args)
	case "del":
		return s.guideDel(args)
	case "list":
		return s.guideList(args)
	case "post":
		return s.guidePost(args)
	case "blogit":
		return s.guideBlogit(args)
	case "unpost":
		return s.guideUnpost(args)
	case "posts":
		return s.guidePosts(args)
	case "page":
		return s.guidePage(args)
	case "unpage":
		return s.guideUnpage(args)
	case "pages":
		return s.guidePages(args)
	case "harvest", "fetch":
		return s.guideHarvest(args)
	case "generate", "build":
		return s.guideGenerate(args)
	case "rss":
		return s.guideRss(args)
	case "sitemap":
		return s.guideSitemap(args)
	case "preview":
		return s.guidePreview(args)
	case "themes":
		return s.guideThemes(args)
	case "apply":
		return s.guideApply(args)
	case "stylefrom":
		return s.guideStylefrom(args)
	default:
		return fmt.Errorf("unknown action %q — type ? to see all actions", action)
	}
}

// run is the REPL loop. initialArgs may carry a first action to run before
// the loop begins.
func (s *iSession) run(initialArgs []string) error {
	// Offer to create the config file if it does not exist yet.
	if _, err := os.Stat(s.cfgName); os.IsNotExist(err) {
		fmt.Printf("\n  %s%s%s does not exist. Create it? [Y/n] %s>%s ",
			Bold, s.cfgName, Reset, Green+Bold, Reset)
		s.scanner.Scan()
		answer := strings.ToLower(strings.TrimSpace(s.scanner.Text()))
		if answer != "n" && answer != "no" {
			if err := s.app.Init(s.cfgName, []string{}); err != nil {
				return err
			}
		}
	}

	if len(initialArgs) > 0 {
		// An action was given on the command line — run it first.
		action := strings.ToLower(initialArgs[0])
		if err := s.guide(action, initialArgs[1:]); err != nil && err != errCancelled {
			fmt.Fprintf(os.Stderr, "%sError:%s %s\n", Red, Reset, err)
		}
	} else {
		// No initial action — welcome and show the help menu.
		fmt.Printf("\n%sWelcome to antenna's interactive mode.%s\n", Bold+Green, Reset)
		fmt.Printf("Each action is explained step by step and the exact command is\n")
		fmt.Printf("shown before it runs. Type %s?%s at any time to see all actions.\n", Bold, Reset)
		s.showHelp()
	}

	// REPL loop.
	for {
		fmt.Printf("\n%santenna%s %s>%s ", Cyan+Bold, Reset, Green+Bold, Reset)
		if !s.scanner.Scan() {
			break
		}
		input := strings.TrimSpace(s.scanner.Text())
		if input == "" {
			continue
		}

		parts := strings.Fields(input)
		verb := strings.ToLower(parts[0])
		rest := parts[1:]

		switch verb {
		case "q", "quit", "exit":
			fmt.Printf("\n%sGoodbye.%s\n", Bold, Reset)
			return nil
		case "?", "help":
			s.showHelp()
		case "history":
			if err := s.handleHistory(rest); err != nil {
				fmt.Fprintf(os.Stderr, "%sError:%s %s\n", Red, Reset, err)
			}
		default:
			if err := s.guide(verb, rest); err != nil {
				if err == errCancelled {
					s.showHelp()
				} else {
					fmt.Fprintf(os.Stderr, "%sError:%s %s\n", Red, Reset, err)
				}
			}
		}
	}
	return nil
}

/**
 * Interactive provides a guided REPL for the antenna command. Any action and
 * partial arguments already on the command line are used as pre-filled values
 * for the first guided step; the session then continues as a REPL until the
 * user quits. Each confirmed command is recorded so the session can be
 * exported as a Bash or PowerShell script with the history action.
 */
func (app *AntennaApp) Interactive(cfgName string, args []string) error {
	s := &iSession{
		app:     app,
		cfgName: cfgName,
		scanner: bufio.NewScanner(os.Stdin),
	}
	return s.run(args)
}

// startPreviewServer loads the config, binds the TCP port, and starts serving
// the htdocs directory in a goroutine. It returns the running *http.Server
// and the URL to open. The caller is responsible for calling Shutdown.
func startPreviewServer(cfgName string) (*http.Server, string, error) {
	cfg := &AppConfig{}
	if err := cfg.LoadConfig(cfgName); err != nil {
		return nil, "", err
	}
	if cfg.Host == "" {
		cfg.Host = "localhost"
	}
	port := cfg.Port
	if port == 0 {
		port = 8000
	}
	htdocs := cfg.Htdocs
	if htdocs == "" {
		htdocs = "."
	}

	mux := http.NewServeMux()
	mux.Handle("/", StaticRouter(http.FileServer(http.Dir(htdocs))))

	addr := fmt.Sprintf("%s:%d", cfg.Host, port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, "", fmt.Errorf("cannot start preview server on %s: %w", addr, err)
	}

	srv := &http.Server{Handler: mux}
	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "preview server: %s\n", err)
		}
	}()

	return srv, fmt.Sprintf("http://%s:%d", cfg.Host, port), nil
}

// ---------------------------------------------------------------------------
// Setup guides
// ---------------------------------------------------------------------------

func (s *iSession) guideInit(args []string) error {
	fmt.Printf(`
%sinit%s — Initialise a new antenna project

Creates the configuration file (%s%s%s) and a default 'pages.md'
collection with its SQLite3 database. Run this once in a new project
directory before using any other action.

The configuration filename is set by the %s-config%s flag and defaults
to %s%s%s. No further parameters are required.
`, Bold+Yellow, Reset, Bold, s.cfgName, Reset, Cyan, Reset, Bold, s.cfgName, Reset)

	return s.confirmAndRun("init", []string{})
}

func (s *iSession) guideAdd(args []string) error {
	fmt.Printf(`
%sadd%s — Register a feed collection

A collection is defined by a Markdown (.md) file whose body contains a list of
RSS/Atom feed URLs as Markdown links. Registering it adds the collection to your
antenna configuration so it can be harvested and its content generated into HTML.
`, Bold+Yellow, Reset)

	preFile, preName, preDesc := "", "", ""
	if len(args) > 0 { preFile = args[0] }
	if len(args) > 1 { preName = args[1] }
	if len(args) > 2 { preDesc = args[2] }

	collectionFile, err := iStep(s.scanner, "COLLECTION_FILE",
		"Path to the Markdown file that defines this collection of feeds (.md).",
		preFile, "my-feeds.md")
	if err != nil { return err }
	if collectionFile == "" {
		return fmt.Errorf("collection file is required")
	}

	name, err := iOptionalStep(s.scanner, "NAME",
		"A short display name shown in generated HTML for this collection.",
		preName, "")
	if err != nil { return err }

	description, err := iOptionalStep(s.scanner, "DESCRIPTION",
		"A sentence describing what this collection covers.",
		preDesc, "")
	if err != nil { return err }

	finalArgs := []string{collectionFile}
	if name != "" {
		finalArgs = append(finalArgs, name)
	}
	if description != "" {
		finalArgs = append(finalArgs, description)
	}

	return s.confirmAndRun("add", finalArgs)
}

func (s *iSession) guideDel(args []string) error {
	fmt.Printf(`
%sdel%s — Remove a feed collection

Unregisters the collection from your antenna configuration. The collection
file (.md) and its SQLite3 database are left on disk untouched.
`, Bold+Yellow, Reset)

	pre := ""
	if len(args) > 0 { pre = args[0] }

	collectionFile, err := iStep(s.scanner, "COLLECTION_FILE",
		"Path to the collection file (.md) you want to remove.",
		pre, "")
	if err != nil { return err }
	if collectionFile == "" {
		return fmt.Errorf("collection file is required")
	}

	return s.confirmAndRun("del", []string{collectionFile})
}

func (s *iSession) guideList(args []string) error {
	fmt.Printf(`
%slist%s — List registered collections

Prints the filename of every collection currently registered in your antenna
configuration. No additional parameters are needed.
`, Bold+Yellow, Reset)

	return s.confirmAndRun("list", []string{})
}

// ---------------------------------------------------------------------------
// Content guides
// ---------------------------------------------------------------------------

func (s *iSession) guidePost(args []string) error {
	fmt.Printf(`
%spost%s — Add a document as a post

Adds a Markdown (.md) file to a collection so it appears in RSS output.
The file's front matter should supply at minimum a 'title' or 'description'.
Include a 'postPath' in front matter to set the published path; pair it with
a 'link' pointing to the public URL.
`, Bold+Yellow, Reset)

	preCollection, preFile := "", ""
	if len(args) > 0 { preCollection = args[0] }
	if len(args) > 1 { preFile = args[1] }

	collection, err := iOptionalStep(s.scanner, "COLLECTION_NAME",
		"The collection to post into. Omit to post into the default 'pages.md' collection.",
		preCollection, "pages.md")
	if err != nil { return err }

	filePath, err := iStep(s.scanner, "FILEPATH",
		"Path to the Markdown document to add as a post (.md).",
		preFile, "")
	if err != nil { return err }
	if filePath == "" {
		return fmt.Errorf("filepath is required")
	}

	finalArgs := []string{}
	if collection != "" {
		finalArgs = append(finalArgs, collection)
	}
	finalArgs = append(finalArgs, filePath)

	return s.confirmAndRun("post", finalArgs)
}

func (s *iSession) guideBlogit(args []string) error {
	fmt.Printf(`
%sblogit%s — Copy a file into a blog directory tree and post it

Takes a Markdown (.md) file, copies it into a date-based path such as
blog/2026/03/01/my-post.md, then adds it to the collection as a post.
`, Bold+Yellow, Reset)

	preCollection, preFile, preDate := "", "", ""
	if len(args) > 0 { preCollection = args[0] }
	if len(args) > 1 { preFile = args[1] }
	if len(args) > 2 { preDate = args[2] }

	collection, err := iOptionalStep(s.scanner, "COLLECTION_NAME",
		"The collection to post into. Omit to use the default 'pages.md' collection.",
		preCollection, "pages.md")
	if err != nil { return err }

	filePath, err := iStep(s.scanner, "FILEPATH",
		"Path to the source Markdown document (.md).",
		preFile, "")
	if err != nil { return err }
	if filePath == "" {
		return fmt.Errorf("filepath is required")
	}

	postDate, err := iOptionalStep(s.scanner, "POST_DATE",
		"Date to use for the blog directory path (YYYY-MM-DD). Defaults to today.",
		preDate, time.Now().Format("2006-01-02"))
	if err != nil { return err }

	finalArgs := []string{}
	if collection != "" {
		finalArgs = append(finalArgs, collection)
	}
	finalArgs = append(finalArgs, filePath)
	if postDate != "" {
		finalArgs = append(finalArgs, postDate)
	}

	return s.confirmAndRun("blogit", finalArgs)
}

func (s *iSession) guideUnpost(args []string) error {
	fmt.Printf(`
%sunpost%s — Remove a post from a collection

Removes a post from the collection index using its URL or post path.
The Markdown file itself is not deleted from disk.
`, Bold+Yellow, Reset)

	preCollection, preURL := "", ""
	if len(args) > 0 { preCollection = args[0] }
	if len(args) > 1 { preURL = args[1] }

	collection, err := iStep(s.scanner, "COLLECTION_NAME",
		"The collection containing the post to remove.",
		preCollection, "pages.md")
	if err != nil { return err }
	if collection == "" {
		return fmt.Errorf("collection name is required")
	}

	urlOrPath, err := iStep(s.scanner, "URL or POST_PATH",
		"The URL or post path that identifies the post (as used in the collection).",
		preURL, "")
	if err != nil { return err }
	if urlOrPath == "" {
		return fmt.Errorf("url or post path is required")
	}

	return s.confirmAndRun("unpost", []string{collection, urlOrPath})
}

func (s *iSession) guidePosts(args []string) error {
	fmt.Printf(`
%sposts%s — List posts in a collection

Prints a Markdown list of posts ordered by publication date, newest first.
Optionally limit output with a count or a date range (not both).
`, Bold+Yellow, Reset)

	preCollection, preCount, preFrom, preTo := "", "", "", ""
	if len(args) > 0 { preCollection = args[0] }
	if len(args) > 1 { preCount = args[1] }
	if len(args) > 2 { preFrom = args[2] }
	if len(args) > 3 { preTo = args[3] }

	collection, err := iStep(s.scanner, "COLLECTION_NAME",
		"The collection whose posts you want to list.",
		preCollection, "pages.md")
	if err != nil { return err }
	if collection == "" {
		return fmt.Errorf("collection name is required")
	}

	count, err := iOptionalStep(s.scanner, "COUNT",
		"Maximum number of posts to return. Leave blank to filter by date range instead.",
		preCount, "")
	if err != nil { return err }

	finalArgs := []string{collection}
	if count != "" {
		finalArgs = append(finalArgs, count)
	} else {
		fromDate, err := iOptionalStep(s.scanner, "FROM_DATE",
			"Include posts published on or after this date (YYYY-MM-DD).",
			preFrom, "")
		if err != nil { return err }
		toDate, err := iOptionalStep(s.scanner, "TO_DATE",
			"Include posts published up to and including this date (YYYY-MM-DD).",
			preTo, "")
		if err != nil { return err }
		if fromDate != "" {
			finalArgs = append(finalArgs, fromDate)
		}
		if toDate != "" {
			finalArgs = append(finalArgs, toDate)
		}
	}

	return s.confirmAndRun("posts", finalArgs)
}

func (s *iSession) guidePage(args []string) error {
	fmt.Printf(`
%spage%s — Add a standalone HTML page

Converts a Markdown (.md) document to HTML and registers it in the pages
collection. Use this for an about page, home page, search page, and similar
non-post content.

%sWarning:%s unsafe HTML in Markdown files passes through unchanged.
Only use this action with files you trust completely.
`, Bold+Yellow, Reset, Red+Bold, Reset)

	preInput, preOutput := "", ""
	if len(args) > 0 { preInput = args[0] }
	if len(args) > 1 { preOutput = args[1] }

	inputPath, err := iStep(s.scanner, "INPUT_PATH",
		"Path to the Markdown document to convert to HTML (.md).",
		preInput, "")
	if err != nil { return err }
	if inputPath == "" {
		return fmt.Errorf("input path is required")
	}

	outputPath, err := iOptionalStep(s.scanner, "OUTPUT_PATH",
		"Where to write the generated HTML file. Defaults to the same base name with .html extension.",
		preOutput, "")
	if err != nil { return err }

	finalArgs := []string{inputPath}
	if outputPath != "" {
		finalArgs = append(finalArgs, outputPath)
	}

	return s.confirmAndRun("page", finalArgs)
}

func (s *iSession) guideUnpage(args []string) error {
	fmt.Printf(`
%sunpage%s — Remove a page from a collection

Removes the page from the collection index so it will no longer be rendered
when the generate action runs. The HTML and Markdown files stay on disk.
`, Bold+Yellow, Reset)

	pre := ""
	if len(args) > 0 { pre = args[0] }

	inputPath, err := iStep(s.scanner, "INPUT_PATH",
		"Path of the page to remove, given as the Markdown source path.",
		pre, "")
	if err != nil { return err }
	if inputPath == "" {
		return fmt.Errorf("input path is required")
	}

	return s.confirmAndRun("unpage", []string{inputPath})
}

func (s *iSession) guidePages(args []string) error {
	fmt.Printf(`
%spages%s — List pages in a collection

Prints all pages registered in the pages collection, ordered by most
recently updated. No additional parameters are needed.
`, Bold+Yellow, Reset)

	return s.confirmAndRun("pages", []string{})
}

// ---------------------------------------------------------------------------
// Publish guides
// ---------------------------------------------------------------------------

func (s *iSession) guideHarvest(args []string) error {
	fmt.Printf(`
%sharvest%s — Fetch content from feeds

Retrieves the latest items from the RSS/Atom feeds listed in your
collections and stores them in the SQLite3 database. Run this regularly
to keep your content up to date before generating.
`, Bold+Yellow, Reset)

	pre := ""
	if len(args) > 0 { pre = args[0] }

	collection, err := iOptionalStep(s.scanner, "COLLECTION_NAME",
		"Harvest only this collection. Leave blank to harvest all registered collections.",
		pre, "")
	if err != nil { return err }

	finalArgs := []string{}
	if collection != "" {
		finalArgs = append(finalArgs, collection)
	}

	return s.confirmAndRun("harvest", finalArgs)
}

func (s *iSession) guideGenerate(args []string) error {
	fmt.Printf(`
%sgenerate%s — Build HTML and RSS output

Processes your collections and renders HTML pages and RSS 2.0 feeds into
the 'htdocs' directory set in your configuration. Run after harvesting to
publish the latest content.
`, Bold+Yellow, Reset)

	pre := ""
	if len(args) > 0 { pre = args[0] }

	collection, err := iOptionalStep(s.scanner, "COLLECTION_NAME",
		"Generate only this collection. Leave blank to generate all registered collections.",
		pre, "")
	if err != nil { return err }

	finalArgs := []string{}
	if collection != "" {
		finalArgs = append(finalArgs, collection)
	}

	return s.confirmAndRun("generate", finalArgs)
}

func (s *iSession) guideRss(args []string) error {
	fmt.Printf(`
%srss%s — Generate an RSS feed file

Produces an RSS 2.0 XML file from the posts in a collection. Optionally
limit the output with a count or a date range (not both).
`, Bold+Yellow, Reset)

	preCollection, preFile, preCount, preFrom, preTo := "", "", "", "", ""
	if len(args) > 0 { preCollection = args[0] }
	if len(args) > 1 { preFile = args[1] }
	if len(args) > 2 { preCount = args[2] }
	if len(args) > 3 { preFrom = args[3] }
	if len(args) > 4 { preTo = args[4] }

	collection, err := iStep(s.scanner, "COLLECTION_NAME",
		"The collection to generate the RSS feed from.",
		preCollection, "pages.md")
	if err != nil { return err }
	if collection == "" {
		return fmt.Errorf("collection name is required")
	}

	rssFile, err := iStep(s.scanner, "RSS_FILENAME",
		"Output filename for the RSS feed.",
		preFile, "feed.xml")
	if err != nil { return err }
	if rssFile == "" {
		return fmt.Errorf("RSS filename is required")
	}

	count, err := iOptionalStep(s.scanner, "COUNT",
		"Maximum number of items to include. Leave blank to filter by date range instead.",
		preCount, "")
	if err != nil { return err }

	finalArgs := []string{collection, rssFile}
	if count != "" {
		finalArgs = append(finalArgs, count)
	} else {
		fromDate, err := iOptionalStep(s.scanner, "FROM_DATE",
			"Include posts published on or after this date (YYYY-MM-DD).",
			preFrom, "")
		if err != nil { return err }
		toDate, err := iOptionalStep(s.scanner, "TO_DATE",
			"Include posts published up to and including this date (YYYY-MM-DD).",
			preTo, "")
		if err != nil { return err }
		if fromDate != "" {
			finalArgs = append(finalArgs, fromDate)
		}
		if toDate != "" {
			finalArgs = append(finalArgs, toDate)
		}
	}

	return s.confirmAndRun("rss", finalArgs)
}

func (s *iSession) guideSitemap(args []string) error {
	fmt.Printf(`
%ssitemap%s — Generate sitemap XML files

Creates sitemap_index.xml and the numbered sitemap files for all pages
and posts found through your antenna configuration. Submit the index URL
to search engines to improve content discoverability.

No additional parameters are needed.
`, Bold+Yellow, Reset)

	return s.confirmAndRun("sitemap", []string{})
}

/**
 * guidePreview starts the static file server in a background goroutine, shows
 * the URL to open, then blocks until the user presses Enter. The server is
 * shut down cleanly before returning to the REPL. The command is recorded in
 * session history so it appears in exported scripts.
 */
func (s *iSession) guidePreview(args []string) error {
	fmt.Printf(`
%spreview%s — Serve the site on localhost for review

Starts a local web server serving your htdocs directory. Open the URL shown
below in your browser to review your site. Press Enter to stop the server
and return to the antenna prompt.
`, Bold+Yellow, Reset)

	iShowCommand(s.app.appName, s.cfgName, "preview", []string{})
	fmt.Printf("\n  %sStart preview server?%s [Y/n] %s>%s ", Bold, Reset, Green+Bold, Reset)
	s.scanner.Scan()
	answer := strings.ToLower(strings.TrimSpace(s.scanner.Text()))
	if isCancelInput(answer) {
		fmt.Println("\n  Cancelled.")
		return errCancelled
	}
	if answer == "n" || answer == "no" {
		fmt.Println("\n  Cancelled.")
		return nil
	}

	srv, url, err := startPreviewServer(s.cfgName)
	if err != nil {
		return err
	}
	s.history = append(s.history, iHistoryEntry{action: "preview", args: []string{}})

	fmt.Printf("\n  %sPreview running at %s%s%s\n", Bold+Green, Cyan, url, Reset)
	fmt.Printf("  Open that URL in your browser to review your site.\n")
	fmt.Printf("  %sPress Enter to stop the server.%s\n", Bold, Reset)
	s.scanner.Scan()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx) //nolint:errcheck

	fmt.Printf("\n  Preview server stopped.\n")
	return nil
}

// ---------------------------------------------------------------------------
// Theme guides
// ---------------------------------------------------------------------------

func (s *iSession) guideThemes(args []string) error {
	fmt.Printf(`
%sthemes%s — List available themes

Scans your project directory for theme directories. A directory is
recognised as a theme when it contains files such as header.md,
footer.md, nav.md, or style.css. No additional parameters are needed.
`, Bold+Yellow, Reset)

	return s.confirmAndRun("themes", []string{})
}

func (s *iSession) guideApply(args []string) error {
	fmt.Printf(`
%sapply%s — Apply a theme to a page generator

Reads a theme directory and updates a page generator YAML file with the
theme's header, nav, footer, and CSS. The generator YAML controls how
HTML pages are structured when the generate action runs.
`, Bold+Yellow, Reset)

	prePath, preYAML := "", ""
	if len(args) > 0 { prePath = args[0] }
	if len(args) > 1 { preYAML = args[1] }

	themePath, err := iStep(s.scanner, "THEME_PATH",
		"Path to the theme directory containing the Markdown and CSS theme files.",
		prePath, "theme")
	if err != nil { return err }
	if themePath == "" {
		return fmt.Errorf("theme path is required")
	}

	yamlFile, err := iOptionalStep(s.scanner, "YAML_FILE_PATH",
		"Page generator YAML file to update. Defaults to the generator named in your configuration.",
		preYAML, "page.yaml")
	if err != nil { return err }

	finalArgs := []string{themePath}
	if yamlFile != "" {
		finalArgs = append(finalArgs, yamlFile)
	}

	return s.confirmAndRun("apply", finalArgs)
}

func (s *iSession) guideStylefrom(args []string) error {
	fmt.Printf(`
%sstylefrom%s — Extract CSS from a LibreOffice Writer file

Reads a LibreOffice Writer HTML export (.html, .htm) and extracts its
paragraph and text styles as a CSS file. The result is ready to drop
into a theme directory as style.css.
`, Bold+Yellow, Reset)

	preInput, preOutput := "", ""
	if len(args) > 0 { preInput = args[0] }
	if len(args) > 1 { preOutput = args[1] }

	inputFile, err := iStep(s.scanner, "INPUT_FILE",
		"Path to the LibreOffice Writer HTML export (.html or .htm).",
		preInput, "")
	if err != nil { return err }
	if inputFile == "" {
		return fmt.Errorf("input file is required")
	}

	outputPath, err := iOptionalStep(s.scanner, "OUTPUT_PATH",
		"Where to write the extracted CSS. Defaults to theme/style.css.",
		preOutput, "theme/style.css")
	if err != nil { return err }

	finalArgs := []string{inputFile}
	if outputPath != "" {
		finalArgs = append(finalArgs, outputPath)
	}

	return s.confirmAndRun("stylefrom", finalArgs)
}
