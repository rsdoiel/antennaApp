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
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	// Import the AntennaApp package
	"github.com/rsdoiel/AntennaApp"
)

var ()

func main() {
	appName := filepath.Base(os.Args[0])
	cfgName := strings.TrimSuffix(appName, ".exe") + ".yaml"
	helpText, fmtHelp := antennaApp.HelpText, antennaApp.FmtHelp
	version, releaseDate, releaseHash, licenseText := antennaApp.Version, antennaApp.ReleaseDate, antennaApp.ReleaseHash, antennaApp.LicenseText
	showHelp, showLicense, showVersion := false, false, false
	// Standard Options
	flag.BoolVar(&showHelp, "help", false, "display help")
	flag.BoolVar(&showLicense, "license", false, "display license")
	flag.BoolVar(&showVersion, "version", false, "display version")
	flag.StringVar(&cfgName, "config", cfgName, "set the configuration filename")
	flag.Parse()
	args := flag.Args()

	in := os.Stdin
	out := os.Stdout
	eout := os.Stderr

	if showHelp {
		if len(args) > 0 {
			if !antennaApp.PrintHelpTopic(out, args[0], appName, version, releaseDate, releaseHash) {
				fmt.Fprintf(eout, "unknown help topic %q — try 'antenna help topics'\n", args[0])
				os.Exit(1)
			}
		} else {
			fmt.Fprintf(out, "%s\n", fmtHelp(helpText, appName, version, releaseDate, releaseHash))
		}
		os.Exit(0)
	}
	if showVersion {
		fmt.Fprintf(out, "%s %s %s\n", appName, version, releaseHash)
		os.Exit(0)
	}
	if showLicense {
		fmt.Fprintf(out, "%s\n", licenseText)
		os.Exit(0)
	}
	// No action provided — enter interactive mode, which will offer to
	// create antenna.yaml if it does not exist.
	if len(args) == 0 {
		args = append(args, "interactive")
	}

	action, args := args[0], args[1:]
	antennaApp := antennaApp.NewAntennaApp(appName)
	if err := antennaApp.Run(in, out, eout, cfgName, action, args); err != nil {
		fmt.Fprintln(eout, err)
		os.Exit(1)
	}
	os.Exit(0)
}
