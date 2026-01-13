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
)

type AntennaApp struct {
	appName string
}

func NewAntennaApp(appName string) *AntennaApp {
	return &AntennaApp{
		appName: appName,
	}
}

// Init initialize a new Antenna App project
func (app *AntennaApp) Init(cfgName string, args []string) error {
	_, err := NewAppConfig(cfgName)
	return err
}

// Run implements the command line functionality of the Antenna App.
func (app *AntennaApp) Run(in io.Reader, out io.Writer, eout io.Writer, cfgName string, action string, args []string) error {
	switch action {
	case "help":
		fmt.Fprintf(out, "%s\n", FmtHelp(HelpText, app.appName, Version, ReleaseDate, ReleaseHash))
		return nil
	case "init":
		return app.Init(cfgName, args)
	case "add":
		return app.Add(cfgName, args)
	case "apply":
		return app.ApplyTheme(cfgName, args)
	case "del":
		return app.Del(cfgName, args)
	case "post":
		return app.Post(cfgName, args)
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
	case "harvest", "fetch":
		return app.Harvest(out, eout, cfgName, args)
	case "generate":
		return app.Generate(out, eout, cfgName, args)
	case "sitemap":
		return app.Sitemap(cfgName, args)
	case "preview":
		return app.Preview(cfgName)
	case "quote", "reply":
		return app.QuoteTextFragment(out, cfgName, args)
	case "interactive", "tui":
		return app.TUI(cfgName, args)
	default:
		return fmt.Errorf("%q not supported", action)
	}
}
