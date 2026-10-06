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
	"fmt"
	"io"
	"os"
	"runtime"
	"path/filepath"
)

type AntennaApp struct {
	appName string
}

func NewAntennaApp(appName string) *AntennaApp {
	return &AntennaApp{
		appName: filepath.Base(os.Args[0]),
	}
}

// Run implements the command line functionality of the Antenna App.
func (app *AntennaApp) Run(in io.Reader, out io.Writer, eout io.Writer, cfgName string, action string, args []string) error {
	// A surplus argument is refused before anything runs (DR-0003): a verb that
	// takes none used to drop them silently, and `preview extra` started the
	// web server.
	if err := checkMaxArgs(app.appName, action, args); err != nil {
		return err
	}
	switch action {
	case "help":
		if len(args) == 0 {
			fmt.Fprintf(out, "%s\n", FmtHelp(HelpText, app.appName, Version, ReleaseDate, ReleaseHash))
			return nil
		}
		if !PrintHelpTopic(out, args[0], app.appName, Version, ReleaseDate, ReleaseHash) {
			return usageErrorf("unknown help topic %q — try 'antenna help topics'", args[0])
		}
		return nil
	case "init":
		return app.Init(cfgName, args)
	case "add":
		return app.Add(cfgName, args)
	case "themes":
		if len(args) > 0 && args[0] == "new" {
			return app.NewTheme(out, cfgName, args[1:])
		}
		return app.ListThemes(out, cfgName, args)
	case "apply":
		return app.ApplyTheme(cfgName, args)
	case "del":
		return app.Del(cfgName, args)
	case "post":
		return app.Post(cfgName, args)
	case "blogit":
		return app.BlogIt(cfgName, args)
	case "posts":
		return app.Posts(cfgName, args)
	case "unpost":
		return app.Unpost(cfgName, args)
	case "page":
		return app.Page(cfgName, args)
	case "pages":
		return app.Pages(cfgName, args)
	case "rss":
		return app.RssPosts(cfgName, args)
	case "unpage":
		return app.Unpage(cfgName, args)
	case "css":
		return app.GenerateCSS(out, cfgName, args)
	case "items":
		return app.Items(out, cfgName, args)
	case "list":
		return app.ListCollectionFiles(out, cfgName, args)
	case "harvest", "fetch":
		return app.Harvest(out, eout, cfgName, args)
	case "generate", "build":
		return app.Generate(out, eout, cfgName, args)
	case "sitemap":
		return app.Sitemap(cfgName, args)
	case "preview":
		return app.Preview(cfgName)
	case "quote", "reply": 
		return app.QuoteTextFragment(out, cfgName, args)
	case "interactive", "tui":
		return app.Interactive(cfgName, args)
	case "stylefrom":
		return app.ExtractStyles(out, args)
	case "completion":
		usage := usageErrorf("usage: %s completion bash|powershell [-install]", app.appName)
		install, shells := false, []string{}
		for _, a := range args {
			if a == "-install" || a == "--install" {
				install = true
			} else {
				shells = append(shells, a)
			}
		}
		if len(shells) != 1 {
			return usage
		}
		if !install {
			return WriteCompletion(out, app.appName, shells[0])
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		path, err := InstallCompletion(app.appName, shells[0], home, os.Getenv("XDG_DATA_HOME"), runtime.GOOS)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "installed %s\n", path)
		return nil
	default:
		return usageErrorf("%q not supported", action)
	}
}

// maxArgs is the most positional arguments each verb documents (see
// "antenna help VERB"). The least are checked where the verb reads them.
// Harvest, generate, del and interactive name several collections or free
// text and have no fixed maximum.
var maxArgs = map[string]int{
	"init": 0, "list": 0, "pages": 0, "preview": 0,
	"sitemap": 1,
	"css": 1, "items": 1, "unpage": 1, "quote": 1, "reply": 1,
	"apply": 2, "page": 2, "stylefrom": 2, "unpost": 2,
	"add": 3, "posts": 3, "post": 3, "blogit": 3,
	"rss": 4,
}

// checkMaxArgs returns a usage error when args has more positionals than
// action documents, and nil for any verb not in maxArgs.
func checkMaxArgs(appName, action string, args []string) error {
	max, ok := maxArgs[action]
	if !ok || len(args) <= max {
		return nil
	}
	return usageErrorf("unexpected argument %q; try '%s help %s'", args[max], appName, action)
}
