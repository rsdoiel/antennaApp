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
	"path/filepath"
	// 3rd Party packages
	"gopkg.in/yaml.v3"
)

// ApplyTheme takes a theme directory and an optional generator YAML filename
// and applies the theme to that generator YAML name saving the result.
func (app AntennaApp) ApplyTheme(cfgName string, args []string) error {
	themeName := ""
	generatorName := ""
	if len(args) == 0 {
		return fmt.Errorf("missing theme directory name")
	}
	if len(args) > 0 {
		themeName = args[0]
	}
	if len(args) > 1 {
		generatorName = args[1]
	}
	cfg := &AppConfig{}
	if err := cfg.LoadConfig(cfgName); err != nil {
		return err
	}
	gen, err := NewGenerator(themeName, cfg.BaseURL)
	if err != nil {
		return err
	}
	if generatorName == "" {
		generatorName = cfg.Generator

	}
	if generatorName == "" || themeName == "" {
		return fmt.Errorf("theme theme or generator name")
	}
	// Reading or create a generator using generatorName
	if _, err := os.Stat(generatorName); err == nil {
		if err := gen.LoadConfig(generatorName); err != nil {
			return err
		}
	} else {
		if err := InitPageGenerator(generatorName); err != nil {
			return err
		}
	}
	changed := false
	if ok, err := updateBodyElements(gen, themeName); err != nil {
		return err
	} else if ok {
		changed = ok
	}
	if ok, err := updateHeadElements(gen, themeName); err != nil {
		return err
	} else if ok {
		changed = ok
	}
	if changed {
		fmt.Printf("saving %s\n", generatorName)
		return saveGenerator(generatorName, gen)
	}
	fmt.Printf("%s left unchanged\n", generatorName)
	return nil
}

// Update a generator's head elements from a theme head.yaml file.
func updateHeadElements(gen *Generator, themeName string) (bool, error) {
	fName := filepath.Join(themeName, "head.yaml")
	changed := false
	if _, err := os.Stat(fName); err == nil {
		src, err := os.ReadFile(fName)
		if err != nil {
			return false, fmt.Errorf("failed to read %q, %s", fName, err)
		}
		head := &Generator{}
		if err := yaml.Unmarshal(src, &head); err != nil {
			return false, fmt.Errorf("failed to parse %q, %s", fName, err)
		}
		if head.Meta != nil && len(head.Meta) > 0 {
			// Clear the previous Meta attributes
			gen.Meta = []map[string]string{}
			changed = true
			// Populate the Meta attributes
			for _, obj := range head.Meta {
				gen.Meta = append(gen.Meta, obj)
			}

		}
		if head.Link != nil && len(head.Link) > 0 {
			// Clear the previous Link attributes
			gen.Link = []map[string]string{}
			changed = true
			// Populate the Link attributes
			for _, obj := range head.Link {
				gen.Link = append(gen.Link, obj)
			}
		}
		if head.Script != nil && len(head.Script) > 0 {
			// Clear the previous Script attributes
			gen.Script = []map[string]string{}
			changed = true
			// Populate the Script attributes
			for _, obj := range head.Script {
				gen.Script = append(gen.Script, obj)
			}
		}

	}
	// Handle an included CSS style description
	fName = filepath.Join(themeName, "style.css")
	if _, err := os.Stat(fName); err == nil {
		src, err := os.ReadFile(fName)
		if err != nil {
			return false, fmt.Errorf("failed to read %q, %s", fName, err)
		}
		if len(src) > 0 {
			gen.Style = fmt.Sprintf("%s\n\n", src)
			changed = true
		}
	}
	return changed, nil
}

// Update a generator's body elements from a theme directory.
func updateBodyElements(gen *Generator, themeName string) (bool, error) {
	// Walk the theme directory to build the HTML elements from the theme
	// if the element exists.
	bodyNames := map[string]string{
		"Header":        "header.md",
		"Nav":           "nav.md",
		"TopContent":    "top_content.md",
		"BottomContent": "bottom_content.md",
		"Footer":        "footer.md",
	}
	doc := &CommonMark{}
	changed := false
	for attr, mdName := range bodyNames {
		fName := filepath.Join(themeName, mdName)
		src, err := os.ReadFile(fName)
		if err == nil {
			// convert src from Markdown to HTML then assign to attribute.
			if err := doc.Parse(src); err != nil {
				return false, fmt.Errorf("failed to parse %q, %s\n", fName, err)
			}
			fmt.Printf("Setting attr from %q\n", fName)
			innerHTML, err := doc.ToHTML()
			if err != nil {
				return false, fmt.Errorf("failed to render %q, %s\n", fName, err)
			}
			switch attr {
			case "Header":
				gen.Header = innerHTML
				changed = true
			case "Nav":
				gen.Nav = innerHTML
				changed = true
			case "TopContent":
				gen.TopContent = innerHTML
				changed = true
			case "BottomContent":
				gen.BottomContent = innerHTML
				changed = true
			case "Footer":
				gen.Footer = innerHTML
				changed = true
			default:
				fmt.Printf("skipping %q, not a body element\n", fName)
			}
		}
	}
	return changed, nil
}

func saveGenerator(fName string, gen *Generator) error {
	src, err := yaml.Marshal(gen)
	if err != nil {
		return fmt.Errorf("failed to encode %q, %s", fName, err)
	}
	if err := os.WriteFile(fName, src, 0664); err != nil {
		return fmt.Errorf("failed to write %q, %s", fName, err)
	}
	return nil
}
