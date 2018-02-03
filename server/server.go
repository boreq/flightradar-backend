package server

import (
	"github.com/boreq/flightradar-backend/aggregator"
	"github.com/boreq/flightradar-backend/server/api"
	"github.com/boreq/flightradar-backend/storage"
	"github.com/julienschmidt/httprouter"
	"net/http"
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

func Serve(aggr aggregator.Aggregator, address string) error {
	h := &handler{aggr}

	router := httprouter.New()
	router.GET("/planes.json", api.Wrap(h.planes))

	return http.ListenAndServe(address, router)
}
