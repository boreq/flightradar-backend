package server

import (
	"fmt"
	"github.com/boreq/flightradar-backend/aggregator"
	"github.com/boreq/flightradar-backend/config"
	"github.com/boreq/flightradar-backend/logging"
	"github.com/boreq/flightradar-backend/server/api"
	"github.com/boreq/flightradar-backend/storage"
	"github.com/julienschmidt/httprouter"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

var log = logging.GetLogger("server")

type handler struct {
	aggr       aggregator.Aggregator
	statsCache map[string]stats
}

func (h *handler) Planes(r *http.Request, _ httprouter.Params) (interface{}, api.Error) {
	var response []storage.Data = make([]storage.Data, 0)

	for _, value := range h.aggr.Newest() {
		response = append(response, value)
	}

	return response, nil
}

func (h *handler) TimeRange(r *http.Request, _ httprouter.Params) (interface{}, api.Error) {
	from, err := timestampParamToTime(r, "from")
	if err != nil {
		return nil, api.BadRequest
	}

	to, err := timestampParamToTime(r, "to")
	if err != nil {
		return nil, api.BadRequest
	}

	response, err := h.aggr.RetrieveTimerange(from, to)
	if err != nil {
		return nil, api.InternalServerError
	}

	return response, nil
}

type polarResponse struct {
	Data     storage.StoredData `json:"data"`
	Distance float64            `json:"distance"`
}

func (h *handler) Polar(r *http.Request, _ httprouter.Params) (interface{}, api.Error) {
	from, err := timestampParamToTime(r, "from")
	if err != nil {
		return nil, api.BadRequest
	}

	to, err := timestampParamToTime(r, "to")
	if err != nil {
		return nil, api.BadRequest
	}

	data, err := h.aggr.RetrieveTimerange(from, to)
	if err != nil {
		return nil, api.InternalServerError
	}

	return toPolar(data), nil
}

func (h *handler) Plane(r *http.Request, ps httprouter.Params) (interface{}, api.Error) {
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

type stats struct {
	DataPointsNumber               int         `json:"data_points_number"`
	DataPointsAltitudeCrossSection map[int]int `json:"data_points_altitude_cross_section"`
	PlanesNumber                   int         `json:"planes_number"`
	FlightsNumber                  int         `json:"flights_number"`
	AverageDistance                float64     `json:"average_distance"`
	MedianDistance                 float64     `json:"median_distance"`
	MaxDistance                    float64     `json:"max_distance"`
}

type dailyStats struct {
	Date string `json:"date"`
	Data stats  `json:"data"`
}

type statsResponse struct {
	Stats                    []dailyStats `json:"stats"`
	AltitudeCrossSectionStep int          `json:"altitude_cross_section_step"`
}

func (h *handler) Stats(r *http.Request, ps httprouter.Params) (interface{}, api.Error) {
	response := statsResponse{
		AltitudeCrossSectionStep: statsAltitudeCrossSectionStep,
		Stats:                    make([]dailyStats, 0),
	}
	for k, v := range h.statsCache {
		response.Stats = append(response.Stats, dailyStats{k, v})
	}
	sort.Slice(response.Stats, func(i, j int) bool { return response.Stats[i].Date < response.Stats[j].Date })
	return response, nil
}

const statsCacheDateLayout = "2006-01-02"
const statsDataPoints = 30 // number of days stats are generated for
const statsUpdateEvery = 60 * time.Minute
const statsAltitudeCrossSectionStep = 5000

func (h *handler) runStats() {
	h.updateStats()

	ticker := time.NewTicker(statsUpdateEvery)
	for range ticker.C {
		h.updateStats()
	}
}

func (h *handler) updateStats() {
	h.cleanupStats()
	h.loadStats()
}

func (h *handler) cleanupStats() {
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
}

func (h *handler) loadStats() {
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

const distanceThreshold = 1000

func (h *handler) getStatsForRange(from, to time.Time) (stats, error) {
	rv := stats{}

	data, err := h.aggr.RetrieveTimerange(from, to)
	if err != nil {
		return rv, err
	}

	// Data points calculations
	uniquePlanes := make(map[string]bool)
	uniqueFlights := make(map[string]bool)
	altitudeCrossSection := make(map[int]int)
	for _, storedData := range data {
		rv.DataPointsNumber++

		if storedData.Data.Icao != nil {
			uniquePlanes[*storedData.Data.Icao] = true
		}

		if storedData.Data.FlightNumber != nil {
			uniqueFlights[*storedData.Data.FlightNumber] = true
		}

		var key int = -1
		if storedData.Data.Altitude != nil {
			key = *storedData.Data.Altitude / statsAltitudeCrossSectionStep
		}
		value, ok := altitudeCrossSection[key]
		if !ok {
			value = 0
		}
		altitudeCrossSection[key] = value + 1
	}
	rv.PlanesNumber = len(uniquePlanes)
	rv.FlightsNumber = len(uniqueFlights)
	rv.DataPointsAltitudeCrossSection = altitudeCrossSection

	// Range calculations
	var sum float64 = 0
	var max float64 = 0
	var distances []float64

	polar := toPolar(data)

	for _, v := range polar {
		if v.Distance > distanceThreshold {
			continue
		}
		sum += v.Distance
		if v.Distance > max {
			max = v.Distance
		}
		distances = append(distances, v.Distance)
	}

	sort.Slice(distances, func(i, j int) bool { return distances[i] < distances[j] })

	rv.MaxDistance = max
	if len(distances) > 0 {
		rv.MedianDistance = distances[len(distances)/2]
		rv.AverageDistance = sum / float64(len(polar))
	} else {
		rv.MedianDistance = 0
		rv.AverageDistance = 0
	}

	return rv, nil
}

func timestampParamToTime(r *http.Request, name string) (time.Time, error) {
	texts, ok := r.URL.Query()[name]
	if !ok {
		return time.Time{}, fmt.Errorf("Parameter %s missing", name)
	}
	timestamp, err := strconv.ParseInt(texts[0], 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(timestamp, 0), nil
}

const multiplier = (2.0 * math.Pi * 6371.0 / 360.0)

func fakeDistance(lon1, lat1, lon2, lat2 float64) float64 {
	a := lon2 - lon1
	b := lat2 - lat1
	return math.Sqrt(a*a+b*b) * multiplier
}

func fakeBearing(lon1, lat1, lon2, lat2 float64) float64 {
	return degrees(math.Atan2(lon2-lon1, lat2-lat1))
}

func toPolar(data []storage.StoredData) map[int]polarResponse {
	// Initial calculations with faster cartesian functions
	result := make(map[int]polarResponse)
	for i := 0; i < len(data); i++ {
		if data[i].Data.Longitude == nil || data[i].Data.Latitude == nil {
			continue
		}
		b := fakeBearing(
			config.Config.StationLongitude,
			config.Config.StationLatitude,
			*data[i].Data.Longitude,
			*data[i].Data.Latitude)
		d := fakeDistance(
			config.Config.StationLongitude,
			config.Config.StationLatitude,
			*data[i].Data.Longitude,
			*data[i].Data.Latitude)
		b = b + 180
		v, ok := result[int(b)]
		if !ok || v.Distance < d {
			result[int(b)] = polarResponse{
				Distance: d,
				Data:     data[i],
			}
		}
	}

	// Recalculate the selected points for increased accuracy
	rv := make(map[int]polarResponse)
	for k, v := range result {
		d := distance(
			config.Config.StationLongitude,
			config.Config.StationLatitude,
			*v.Data.Data.Longitude,
			*v.Data.Data.Latitude)
		v.Distance = d
		rv[k] = v
	}

	return result
}

func Serve(aggr aggregator.Aggregator, address string) error {
	h := &handler{
		aggr:       aggr,
		statsCache: make(map[string]stats),
	}
	go h.runStats()

	router := httprouter.New()
	router.GET("/planes.json", api.Wrap(h.Planes))
	router.GET("/plane/:icao", api.Wrap(h.Plane))
	router.GET("/range.json", api.Wrap(h.TimeRange))
	router.GET("/polar.json", api.Wrap(h.Polar))
	router.GET("/stats.json", api.Wrap(h.Stats))

	return http.ListenAndServe(address, router)
}
