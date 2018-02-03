package server

import (
	"github.com/boreq/flightradar-backend/aggregator"
	"github.com/boreq/flightradar-backend/server/api"
	"github.com/boreq/flightradar-backend/storage"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strings"
)

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

func (h *handler) plane(r *http.Request, ps httprouter.Params) (interface{}, api.Error) {
	var response []storage.StoredData

	icao := ps.ByName("icao")
	if strings.HasSuffix(icao, ".json") {
		icao = icao[:len(icao)-len(".json")]
	}

	c, err := h.aggr.Retrieve(icao)
	if err != nil {
		return nil, api.InternalServerError
	}
	for value := range c {
		response = append(response, value)
	}

	return response, nil
}

func Serve(aggr aggregator.Aggregator, address string) error {
	h := &handler{aggr}

	router := httprouter.New()
	router.GET("/planes.json", api.Wrap(h.planes))
	router.GET("/plane/:icao", api.Wrap(h.plane))

	return http.ListenAndServe(address, router)
}
