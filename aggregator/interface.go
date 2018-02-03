package aggregator

import (
	"github.com/boreq/flightradar-backend/storage"
)

type Aggregator interface {
	// GetChannel returns the channel which is used to receive the incoming
	// data. Send the received data on this channel and it will be
	// processed and stored by the aggregator.
	GetChannel() chan<- storage.Data

	// Newest returns the map which links the ADS-B/MODE-S aircraft id with
	// the latest data for that aircraft.
	Newest() map[string]storage.Data
}
