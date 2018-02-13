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
	"strconv"
	"strings"
	"time"
)

var log = logging.GetLogger("server")

type handler struct {
	aggr aggregator.Aggregator
}

func (h *handler) planes(r *http.Request, _ httprouter.Params) (interface{}, api.Error) {
	var response []storage.Data

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

func Serve(aggr aggregator.Aggregator, address string) error {
	h := &handler{aggr}

	router := httprouter.New()
	router.GET("/planes.json", api.Wrap(h.planes))
	router.GET("/plane/:icao", api.Wrap(h.plane))
	router.GET("/range.json", api.Wrap(h.timeRange))
	router.GET("/polar.json", api.Wrap(h.polar))

	return http.ListenAndServe(address, router)
}
