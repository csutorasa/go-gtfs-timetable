package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/artonge/go-gtfs"
)

type handler struct {
	gs                      *gtfs.GTFS
	routeById               *LookupCache[string, gtfs.Route]
	tripById                *LookupCache[string, gtfs.Trip]
	stopById                *LookupCache[string, gtfs.Stop]
	stopTimesByStopId       *LookupGroupCache[string, gtfs.StopTime]
	calendarDateByServiceId *LookupCache[string, gtfs.CalendarDate]
}

var files = []string{
	"/index.html",
	"/search.html",
	"/api.js",
	"/storage.js",
	"/style.css",
}

func (h *handler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if request.URL.Path == "/status" {
		if request.Method == http.MethodGet {
			if h.gs == nil {
				response.WriteHeader(http.StatusServiceUnavailable)
				response.Write([]byte("Starting"))
				response.Write([]byte("Server is starting"))
				return
			}
			response.WriteHeader(http.StatusOK)
			response.Write([]byte("OK"))
			return
		}
		response.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if strings.Index(request.URL.Path, "/api") == 0 {
		if request.URL.Path == "/api/findstop" {
			h.findStop(response, request)
			return
		}
		if request.URL.Path == "/api/upcoming" {
			h.upcoming(response, request)
			return
		}
	}
	h.serveFiles(response, request)
}

func (h *handler) Loaded(gs *gtfs.GTFS) {
	h.routeById = NewLookupCache(gs.Routes, func(r *gtfs.Route) string { return r.ID })
	h.tripById = NewLookupCache(gs.Trips, func(t *gtfs.Trip) string { return t.ID })
	h.stopById = NewLookupCache(gs.Stops, func(s *gtfs.Stop) string { return s.ID })
	h.stopTimesByStopId = NewLookupGroupCache(gs.StopsTimes, func(s *gtfs.StopTime) string { return s.StopID })
	h.calendarDateByServiceId = NewLookupCache(gs.CalendarDates, func(c *gtfs.CalendarDate) string { return c.Date + c.ServiceID })
	h.gs = gs
}

func (h *handler) serveFiles(response http.ResponseWriter, request *http.Request) {
	var filePath string
	if request.URL.Path == "/" {
		filePath = "/index.html"
	} else {
		for _, file := range files {
			if file == request.URL.Path {
				filePath = file
				break
			}
		}
	}
	if filePath == "" {
		response.WriteHeader(http.StatusNotFound)
		response.Write([]byte("File not found"))
		return
	}
	if request.Method == http.MethodGet {
		http.ServeFile(response, request, "web"+filePath)
		return
	}
	response.WriteHeader(http.StatusMethodNotAllowed)
	response.Write([]byte("GET is accepted for getting files"))
}

func (h *handler) serveFile(response http.ResponseWriter, request *http.Request) {
	file, err := os.ReadFile("web/search.html")
	if err != nil {
		response.WriteHeader(http.StatusNotFound)
		return
	}
	response.Write(file)
}

func (h *handler) findStop(response http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet {
		stopNames, ok := request.URL.Query()["stopName"]
		if !ok {
			response.WriteHeader(http.StatusBadRequest)
			response.Write([]byte("stopName parameter is missing"))
			return
		}
		if len(stopNames) != 1 {
			response.WriteHeader(http.StatusBadRequest)
			response.Write([]byte("There must be a single stopName parameter"))
			return
		}
		stopName := stopNames[0]
		if request.Method == http.MethodGet {
			if h.gs == nil {
				response.WriteHeader(http.StatusServiceUnavailable)
				response.Write([]byte("Server is starting"))
				return
			}
			findStopResponses := h.doFindStop(stopName)
			data, err := json.Marshal(findStopResponses)
			if err != nil {
				response.WriteHeader(http.StatusInternalServerError)
				response.Write([]byte("Failed to marshal JSON"))
				return
			}
			response.WriteHeader(http.StatusOK)
			response.Write(data)
			return
		}
	}
	response.WriteHeader(http.StatusMethodNotAllowed)
}

func (h *handler) doFindStop(stopName string) []*FindStopResponse {
	stops := findStopsByName(h.gs, stopName)
	response := []*FindStopResponse{}
	for _, stop := range stops {
		routesSet := map[string]*Route{}
		stopTimes := h.stopTimesByStopId.Get(stop.ID)
		for _, stopTime := range stopTimes {
			trip, ok := h.tripById.Get(stopTime.TripID)
			if !ok {
				continue
			}
			route, ok := h.routeById.Get(trip.RouteID)
			if !ok {
				continue
			}
			routesSet[route.ShortName+trip.Headsign] = &Route{
				Name:      route.ShortName,
				HeadSign:  trip.Headsign,
				Color:     route.Color,
				TextColor: route.TextColor,
			}
		}
		routes := []*Route{}
		for _, v := range routesSet {
			routes = append(routes, v)
		}
		response = append(response, &FindStopResponse{
			ID:       stop.ID,
			StopName: stop.Name,
			Routes:   routes,
		})
	}
	return response
}

func (h *handler) upcoming(response http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet {
		query := request.URL.Query()
		stopIds, ok := query["stopId"]
		if !ok {
			response.WriteHeader(http.StatusBadRequest)
			response.Write([]byte("stopId parameter is missing"))
			return
		}
		untils, ok := query["until"]
		from := time.Now()
		var until time.Time
		if ok {
			if len(untils) != 1 {
				response.WriteHeader(http.StatusBadRequest)
				response.Write([]byte("until parameter should have a single value"))
				return
			}
			u, err := time.Parse("20060102 15:04:05", untils[0])
			if err != nil {
				response.WriteHeader(http.StatusBadRequest)
				response.Write([]byte("until parameter cannot be parsed to YYYYMMDD hh:mm:ss"))
				return
			}
			if u.Before(from) {
				response.WriteHeader(http.StatusBadRequest)
				response.Write([]byte("until parameter value is in the past"))
				return
			}
			until = u
		} else {
			until = from.Add(12 * time.Hour)
		}
		if h.gs == nil {
			response.WriteHeader(http.StatusServiceUnavailable)
			response.Write([]byte("Server is starting"))
			return
		}
		stops := []*gtfs.Stop{}
		for _, stopId := range stopIds {
			stop, ok := h.stopById.Get(stopId)
			if !ok {
				response.WriteHeader(http.StatusBadRequest)
				response.Write([]byte(fmt.Sprintf("invalid stopId parameter %s", stopId)))
				return
			}
			stops = append(stops, stop)
		}
		searchUpcomingResponse := h.searchUpcoming(stops, from, until)
		data, err := json.Marshal(searchUpcomingResponse)
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte("Failed to marshal JSON"))
			return
		}
		response.WriteHeader(http.StatusOK)
		response.Write(data)
		return
	}
	response.WriteHeader(http.StatusMethodNotAllowed)
}

func (h *handler) searchUpcoming(stops []*gtfs.Stop, from, until time.Time) []*SearchUpcomingResponse {
	response := []*SearchUpcomingResponse{}
	for _, stop := range stops {
		stopTimes := h.stopTimesByStopId.Get(stop.ID)
		departures := []*Departure{}
		for _, stopTime := range stopTimes {
			for _, date := range findDates(from, until) {
				departure, err := time.Parse("2006010215:04:05", date+stopTime.Departure)
				if err != nil {
					continue
				}
				if departure.Before(from) {
					continue
				}
				if departure.After(until) {
					continue
				}
				trip, ok := h.tripById.Get(stopTime.TripID)
				if !ok {
					continue
				}
				_, ok = h.calendarDateByServiceId.Get(date + trip.ServiceID)
				if !ok {
					continue
				}
				route, ok := h.routeById.Get(trip.RouteID)
				if !ok {
					continue
				}
				departures = append(departures, &Departure{
					DepartureTime: stopTime.Departure,
					Route: &Route{
						Name:      route.ShortName,
						HeadSign:  trip.Headsign,
						Color:     route.Color,
						TextColor: route.TextColor,
					},
				})
			}
		}
		response = append(response, &SearchUpcomingResponse{
			ID:         stop.ID,
			StopName:   stop.Name,
			Departures: departures,
		})
	}
	return response
}

func startWeb(h *handler) error {
	return http.ListenAndServe(":8080", h)
}
