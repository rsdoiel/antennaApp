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
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/
package antennaApp

import (
	"fmt"
	"time"
)

/** HarvestConfig holds the optional harvest settings in antenna.yaml
 * (DR-0030 8).
 *
 * Fields:
 *   Pause (string) — the wait between two requests to one host, a Go
 *                    duration such as "500ms" or "2s"; "0s" turns it off.
 *                    Absent means 500ms.
 *
 * Example:
 *   harvest:
 *     pause: 1s
 */
type HarvestConfig struct {
	Pause string `json:"pause,omitempty" yaml:"pause,omitempty"`
}

// harvestPause returns the pause between requests to a host: the configured
// one, else the default. validateHarvest has already checked it parses.
func (cfg *AppConfig) harvestPause() time.Duration {
	if cfg.Harvest == nil || cfg.Harvest.Pause == "" {
		return defaultHarvestPause
	}
	d, err := time.ParseDuration(cfg.Harvest.Pause)
	if err != nil {
		return defaultHarvestPause
	}
	return d
}

// validateHarvest checks the harvest mapping; a bad value is a configuration
// error found when the configuration loads.
func (cfg *AppConfig) validateHarvest() error {
	if cfg.Harvest == nil || cfg.Harvest.Pause == "" {
		return nil
	}
	d, err := time.ParseDuration(cfg.Harvest.Pause)
	if err != nil {
		return fmt.Errorf("harvest.pause %q is not a duration such as 500ms or 2s", cfg.Harvest.Pause)
	}
	if d < 0 {
		return fmt.Errorf("harvest.pause is %s; it must be 0 or more", cfg.Harvest.Pause)
	}
	return nil
}

// harvestFetcher returns the fetcher for one harvest run, with the
// configured User-Agent and pause.
func (cfg *AppConfig) harvestFetcher() *fetcher {
	return newFetcher(cfg.UserAgent, cfg.harvestPause())
}
