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

	// 3rd Party
	_ "github.com/glebarez/go-sqlite"
)

func (app *AntennaApp) Add(cfgName string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing collection name")
	}
	// create a cfg object
	cfg := &AppConfig{}
	// Load configuration
	if err := cfg.LoadConfig(cfgName); err != nil {
		return err
	}
	// Get the collection name
	cName := args[0]
	return cfg.AddCollection(cfgName, cName)
}


// Del removes one or more collections from your Antenna instance.
// NOTE: It does not remove generated files, e.g. SQLite3 database,
// YAML, HTML or RSS files.
func (app *AntennaApp) Del(cfgName string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing collection name")
	}

	// create a cfg object
	cfg := &AppConfig{}
	// Load configuration
	if err := cfg.LoadConfig(cfgName); err != nil {
		return err
	}
	for _, cName := range args {
	  if err := cfg.DelCollection(cfgName, cName); err != nil {
	  	return err
	  }
	}
	return nil
}
