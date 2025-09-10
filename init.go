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
	"os"

	// 3rd Party Modules
	"gopkg.in/yaml.v3"
)

func (app *AntennaApp) Init(cfgName string, args []string) error {
	fName := cfgName
	// Check if antenna.yaml exists, confirm I can read it.
	cfg := &AppConfig{}
	if _, err := os.Stat(fName); err == nil {
		src, err := os.ReadFile(fName)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(src, &cfg); err != nil {
			return err
		}
		if cfg.Port == 0 {
			return fmt.Errorf("invalid port number in %s", fName)
		}
		if cfg.Htdocs != "" {
			if _, err := os.Stat(cfg.Htdocs); err != nil {
				return fmt.Errorf("problem with htdocs: %q in %s: %s", cfg.Htdocs, fName, err)
			}			
		}
		if cfg.Collections == nil {
			return fmt.Errorf("no collections defined in %s, try adding one", fName)
		}
	}
	// If antenna.yaml does not exist, create it
	cfg.Port = 8000
	// By default the working directory is assumed to be the staging directory.
	cfg.Htdocs = ""
	src, err := yaml.Marshal(cfg)
	if err != nil {
		// This shouldn't happen ever, if it does it is a programming error
		return fmt.Errorf("unable to generate YAML, %s", err)
	}
	fp, err := os.Create(fName)
	if err != nil {
		return err
	}
	defer fp.Close()
	fmt.Fprintf(fp, "%s", src)
	return nil
}
