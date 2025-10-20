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
	    fmt.Printf("using %s\n", fName)
		src, err := os.ReadFile(fName)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(src, &cfg); err != nil {
			return err
		}
	} else {
	    fmt.Printf("creating %s\n", fName)
		cfg.Htdocs = ""
		cfg.Port = 8000
		cfg.BaseURL = "http://localhost:8000"
	}
	if cfg.Port == 0 {
		cfg.Port = 8000
	}
	if cfg.Host == "" && cfg.BaseURL == "" {
		cfg.Host = "localhost"
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = fmt.Sprintf("http://%s:%s", cfg.Host, cfg.Port)
	}

	if cfg.Htdocs != "" {
		if _, err := os.Stat(cfg.Htdocs); err != nil {
			return fmt.Errorf("problem with htdocs: %q in %s: %s", cfg.Htdocs, fName, err)
		}
	}
	if cfg.Generator == "" {
		cfg.Generator = "page.yaml"
	}
	if err := cfg.SaveConfig(cfgName); err != nil {
		return fmt.Errorf("failed to save %s, %s", cfgName, err)
	}

	if err := InitPageGenerator(cfg.Generator); err != nil {
		return fmt.Errorf("failed to generate default %s, %s", cfg.Generator, err)
	}
	// Add the default pages.md collection.
	cName := "pages.md"
	dbName := "pages.db"
	if _, err := os.Stat(cName); err != nil {
		fmt.Printf("Creating the %s\n", cName)
		if err := os.WriteFile(cName, []byte(DefaultPageCollectionMarkdown), 0664); err != nil {
			return fmt.Errorf("failed to created %s, %s", cName, err)
		}
	} else {
	    fmt.Printf("using existing %s\n", cName)
	}
	
	if _, err := os.Stat(dbName); err != nil {
		fmt.Printf("Adding %s\n", cName)
		if err := app.Add(cfgName, []string{cName}); err != nil {
			return fmt.Errorf("failed to create default collection, %s, %s", cName, err)
		}
	} else {
	    fmt.Printf("using existing %s\n", dbName)
	}
	return nil
}
