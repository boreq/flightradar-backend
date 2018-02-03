package aggregator

import (
	"github.com/boreq/flightradar/storage"
)

func New() Aggregator {
	rv := &aggregator{
		data: make(chan storage.Data),
	}
	go rv.run()
	return rv
}

type aggregator struct {
	data chan storage.Data
}

func (a *aggregator) GetChannel() chan<- storage.Data {
	return a.data
}

func (a *aggregator) run() {
	for {
	}
}
