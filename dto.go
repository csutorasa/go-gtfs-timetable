package main

type FindStopResponse struct {
	ID       string   `json:"id"`
	StopName string   `json:"stopName"`
	Routes   []*Route `json:"routes"`
}

type SearchUpcomingResponse struct {
	ID         string       `json:"id"`
	StopName   string       `json:"stopName"`
	Departures []*Departure `json:"departures"`
}

type Departure struct {
	DepartureTime string `json:"departureTime"`
	Route         *Route `json:"route"`
}

type Route struct {
	Name      string `json:"name"`
	HeadSign  string `json:"headSign"`
	Color     string `json:"color"`
	TextColor string `json:"textColor"`
}
