package aggregator

import (
	"github.com/boreq/flightradar/storage"
)

type Aggregator interface {
	GetChannel() chan<- storage.Data
	Newest() map[string]storage.Data
}
