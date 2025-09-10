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
	"time"
)

// CheckWaitInterval checks to see if an interval of time has been met or exceeded.
// It returns the remaining time interval (possibly reset) and a boolean. The
// boolean is true when the time interval has been met or exceeded, false otherwise.
//
// ```
// tot := len(something) // calculate the total number of items to process
// t0 := time.Now()
// iTime := time.Now()
// reportProgress := false
//
//	for i, key := range records {
//	    // ... process stuff ...
//	    if iTime, reportProgress = checkWaitInterval(rptTime, (30 * time.Second)); reportProgress {
//	        log.Printf("%s", ProgressETA(t0, i, tot))
//	    }
//	}
//
// ```
func CheckWaitInterval(iTime time.Time, wait time.Duration) (time.Time, bool) {
	if time.Since(iTime) >= wait {
		iTime = time.Now()
		return iTime, true
	}
	return iTime, false
}

// ProgressETA returns a string with the percentage processed and estimated time remaining.
// It requires the a counter of records processed, the total count of records and a time zero value.
//
// ```
// tot := len(something) // calculate the total number of items to process
// t0 := time.Now()
// iTime := time.Now()
// reportProgress := false
//
//	for i, key := range records {
//	    // ... process stuff ...
//	    if iTime, reportProgress = CheckWaitInterval(rptTime, (30 * time.Second)); reportProgress {
//	        log.Printf("%s", ProgressETA(t0, i, tot))
//	    }
//	}
//
// ```
func ProgressETA(t0 time.Time, i int, tot int) string {
	if i == 0 {
		return fmt.Sprintf("%.2f%% ETA unknown", 0.0)
	}
	// percent completed
	percent := (float64(i) / float64(tot)) * 100.0
	// running time
	rt := time.Since(t0)
	// estimated time remaining
	eta := time.Duration((float64(rt) / float64(i) * float64(tot)) - float64(rt))
	return fmt.Sprintf("%.2f%% ETA %v", percent, eta.Round(time.Second))
}

// ProgressIPS returns a string with the elapsed time and increments per second.
// Takes a time zero, a counter and time unit. Returns a string with count, running time and
// increments per time unit.
// ```
// t0 := time.Now()
// iTime := time.Now()
// reportProgress := false
//
//	for i, key := range records {
//	    // ... process stuff ...
//	    if iTime, reportProgress = CheckWaitInterval(rptTime, (30 * time.Second)); reportProgress || i = 0 {
//	        log.Printf("%s", ProgressIPS(t0, i, time.Second))
//	    }
//	}
//
// ```
func ProgressIPS(t0 time.Time, i int, timeUnit time.Duration) string {
	if i == 0 {
		return fmt.Sprintf("(%d/%s) IPS unknown", i, time.Since(t0).Round(timeUnit))
	}
	ips := float64(i) / float64(time.Since(t0).Seconds())
	return fmt.Sprintf("(%d/%s) IPS %.2f i/sec.", i, time.Since(t0).Round(timeUnit), ips)
}
