package server

import (
	"fmt"
	"github.com/boreq/flightradar-backend/aggregator"
	"github.com/boreq/flightradar-backend/config"
	"github.com/boreq/flightradar-backend/logging"
	"github.com/boreq/flightradar-backend/server/api"
	"github.com/boreq/flightradar-backend/storage"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

var log = logging.GetLogger("server")

type stats struct {
	DataPointsNumber int `json:"data_points_number"`
	PlanesNumber     int `json:"planes_number"`
	FlightsNumber    int `json:"flights_number"`
}

type handler struct {
	aggr       aggregator.Aggregator
	statsCache map[string]stats
}

func (h *handler) planes(r *http.Request, _ httprouter.Params) (interface{}, api.Error) {
	var response []storage.Data = make([]storage.Data, 0)

	for _, value := range h.aggr.Newest() {
		response = append(response, value)
	}

	return response, nil
}

func (h *handler) timeRange(r *http.Request, _ httprouter.Params) (interface{}, api.Error) {
	now := time.Now()

	fromText, ok := r.URL.Query()["from"]
	if !ok {
		return nil, api.BadRequest
	}
	toText, ok := r.URL.Query()["to"]
	if !ok {
		return nil, api.BadRequest
	}

	fromInt, err := strconv.ParseInt(fromText[0], 10, 64)
	if err != nil {
		return nil, api.BadRequest
	}

	toInt, err := strconv.ParseInt(toText[0], 10, 64)
	if err != nil {
		return nil, api.BadRequest
	}

	from := time.Unix(fromInt, 0)
	to := time.Unix(toInt, 0)

	response, err := h.aggr.RetrieveTimerange(from, to)
	if err != nil {
		return nil, api.InternalServerError
	}

	fmt.Printf("Time range total: %f\n", time.Since(now).Seconds())
	return response, nil
}

type polarResponse struct {
	Data     storage.StoredData `json:"data"`
	Distance float64            `json:"distance"`
}

func (h *handler) polar(r *http.Request, _ httprouter.Params) (interface{}, api.Error) {
	now := time.Now()

	fromText, ok := r.URL.Query()["from"]
	if !ok {
		return nil, api.BadRequest
	}
	toText, ok := r.URL.Query()["to"]
	if !ok {
		return nil, api.BadRequest
	}

	fromInt, err := strconv.ParseInt(fromText[0], 10, 64)
	if err != nil {
		return nil, api.BadRequest
	}

	toInt, err := strconv.ParseInt(toText[0], 10, 64)
	if err != nil {
		return nil, api.BadRequest
	}

	from := time.Unix(fromInt, 0)
	to := time.Unix(toInt, 0)

	data, err := h.aggr.RetrieveTimerange(from, to)
	if err != nil {
		return nil, api.InternalServerError
	}

	response := make(map[int]polarResponse)
	for _, d := range data {
		if d.Data.Longitude == nil || d.Data.Latitude == nil {
			continue
		}
		bearing := bearing(
			config.Config.StationLongitude,
			config.Config.StationLatitude,
			*d.Data.Longitude,
			*d.Data.Latitude)
		distance := distance(
			config.Config.StationLongitude,
			config.Config.StationLatitude,
			*d.Data.Longitude,
			*d.Data.Latitude)
		bearing = bearing + 180
		v, ok := response[int(bearing)]
		if !ok || v.Distance < distance {
			response[int(bearing)] = polarResponse{
				Distance: distance,
				Data:     d,
			}
		}
	}

	fmt.Printf("Polar seconds: %f\n", time.Since(now).Seconds())
	return response, nil
}

func (h *handler) plane(r *http.Request, ps httprouter.Params) (interface{}, api.Error) {
	icao := ps.ByName("icao")
	if strings.HasSuffix(icao, ".json") {
		icao = icao[:len(icao)-len(".json")]
	}

	response, err := h.aggr.Retrieve(icao)
	if err != nil {
		return nil, api.InternalServerError
	}

	return response, nil
}

type statsResponse struct {
	Date string `json:"date"`
	Data stats  `json:"data"`
}

func (h *handler) stats(r *http.Request, ps httprouter.Params) (interface{}, api.Error) {
	response := make([]statsResponse, 0)
	for k, v := range h.statsCache {
		response = append(response, statsResponse{k, v})
	}
	sort.Slice(response, func(i, j int) bool { return response[i].Date < response[j].Date })
	return response, nil
}

const statsCacheDateLayout = "2006-01-02"
const statsDataPoints = 30 // number of days stats are generated for
const statsUpdateEvery = 60 * time.Minute

func (h *handler) runStats() {
	h.updateStats()

	ticker := time.NewTicker(statsUpdateEvery)
	for range ticker.C {
		h.updateStats()
	}
}

func (h *handler) updateStats() {
	// Cleanup
	for k, _ := range h.statsCache {
		t, err := time.Parse(statsCacheDateLayout, k)
		if err != nil {
			log.Printf("updateStats cleanup error: %s", err)
			delete(h.statsCache, k)
		}
		if time.Since(t) > statsDataPoints*24*time.Hour {
			log.Debugf("updateStats cleanup: %s", k)
			delete(h.statsCache, k)
		}
	}

	// Load new
	for i := 0; i < statsDataPoints; i++ {
		start := time.Now().UTC().AddDate(0, 0, -i).Truncate(24 * time.Hour)
		end := start.AddDate(0, 0, 1)
		key := start.Format(statsCacheDateLayout)
		_, ok := h.statsCache[key]
		if !ok || i <= 1 {
			log.Debugf("updateStats loading: %s", key)
			stats, err := h.getStatsForRange(start, end)
			if err != nil {
				log.Printf("updateStats error: %s", err)
				continue
			}
			h.statsCache[key] = stats
		}
	}
}

func (h *handler) getStatsForRange(from, to time.Time) (stats, error) {
	rv := stats{}

	data, err := h.aggr.RetrieveTimerange(from, to)
	if err != nil {
		return rv, err
	}

	uniquePlanes := make(map[string]bool)
	uniqueFlights := make(map[string]bool)
	for _, storedData := range data {
		rv.DataPointsNumber++
		if storedData.Data.Icao != nil {
			uniquePlanes[*storedData.Data.Icao] = true
		}
		if storedData.Data.FlightNumber != nil {
			uniqueFlights[*storedData.Data.FlightNumber] = true
		}
	}
	rv.PlanesNumber = len(uniquePlanes)
	rv.FlightsNumber = len(uniqueFlights)

	return rv, nil
}

func Serve(aggr aggregator.Aggregator, address string) error {
	h := &handler{
		aggr:       aggr,
		statsCache: make(map[string]stats),
	}
	go h.runStats()

	router := httprouter.New()
	router.GET("/planes.json", api.Wrap(h.planes))
	router.GET("/plane/:icao", api.Wrap(h.plane))
	router.GET("/range.json", api.Wrap(h.timeRange))
	router.GET("/polar.json", api.Wrap(h.polar))
	router.GET("/stats.json", api.Wrap(h.stats))

	return http.ListenAndServe(address, router)
}
