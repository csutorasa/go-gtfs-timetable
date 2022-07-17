package main

import (
	"strings"
	"time"

	"github.com/artonge/go-gtfs"
)

func findStopsByName(gs *gtfs.GTFS, stopName string) []*gtfs.Stop {
	stops := []*gtfs.Stop{}
	comp := strings.ToLower(stopName)
	for i, stop := range gs.Stops {
		if strings.Contains(strings.ToLower(stop.Name), comp) {
			stops = append(stops, &gs.Stops[i])
		}
	}
	return stops
}

func findDates(from, to time.Time) []string {
	dates := []string{}
	toDate := to.Format("20060102")
	date := from
	for {
		d := date.Format("20060102")
		dates = append(dates, d)
		if d == toDate {
			break
		}
		date = date.Add(24 * time.Hour)
	}
	return dates
}
