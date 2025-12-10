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
)

// Curate provides a web interface for feed items by collection. It does this
// creating a localhost webservice that your web browser may interact with to curate
// your collections.
func (app *AntennaApp) Curate(cfgName string, args []string) error {
	cfg := &AppConfig{}
	if err := cfg.LoadConfig(cfgName); err != nil {
		return err
	}

	// Setup handlers for
	//
	// - A List collections
	// - List feeds and pages in a collection
	// - Curate a list of feed items
	// - Curate a list of pages in a collection
	// - Curate Themes
	// - Curate filter methods for rendering feed items

	// Run localhost webservice with curation services
	return fmt.Errorf("curate not implemented")
}