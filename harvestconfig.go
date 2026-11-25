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

	// FailingDays, FailingAttempts and LogDays are DR-0038's settings: a feed
	// is failing when it has FailingAttempts or more attempts in the last
	// FailingDays days and every one failed; harvest_log rows older than
	// LogDays days are deleted. Absent means 7, 3 and 90. A value below 1 is a
	// configuration error.
	FailingDays     *int `json:"failing_days,omitempty" yaml:"failing_days,omitempty"`
	FailingAttempts *int `json:"failing_attempts,omitempty" yaml:"failing_attempts,omitempty"`
	LogDays         *int `json:"log_days,omitempty" yaml:"log_days,omitempty"`
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
	if cfg.Harvest == nil {
		return nil
	}
	for name, v := range map[string]*int{
		"harvest.failing_days":     cfg.Harvest.FailingDays,
		"harvest.failing_attempts": cfg.Harvest.FailingAttempts,
		"harvest.log_days":         cfg.Harvest.LogDays,
	} {
		if v != nil && *v < 1 {
			return fmt.Errorf("%s is %d; it must be a whole number of 1 or more", name, *v)
		}
	}
	if cfg.Harvest.Pause == "" {
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

// harvestPolicy returns the failing-feed settings: the configured ones, else
// the defaults. validateHarvest has already checked them.
func (cfg *AppConfig) harvestPolicy() harvestPolicy {
	p := defaultHarvestPolicy
	if cfg.Harvest == nil {
		return p
	}
	if v := cfg.Harvest.FailingDays; v != nil {
		p.failingDays = *v
	}
	if v := cfg.Harvest.FailingAttempts; v != nil {
		p.failingAttempts = *v
	}
	if v := cfg.Harvest.LogDays; v != nil {
		p.logDays = *v
	}
	return p
}

// harvestFetcher returns the fetcher for one harvest run, with the
// configured User-Agent and pause.
func (cfg *AppConfig) harvestFetcher() *fetcher {
	f := newFetcher(cfg.UserAgent, cfg.harvestPause())
	f.policy = cfg.harvestPolicy()
	return f
}
