package aggregator

import (
	"fmt"
	"github.com/boreq/flightradar-backend/storage"
	"os"
	"time"
)

func New(s storage.Storage) Aggregator {
	rv := &aggregator{
		storage:    s,
		data:       make(chan storage.Data),
		recent:     make(map[string]dataWithTime),
		storeTimes: make(map[string]time.Time),
	}
	go rv.run()
	return rv
}

const storeEvery = 30 * time.Second
const dataTimeoutThreshold = 30 * time.Second

type dataWithTime struct {
	storage.Data
	t time.Time
}

type aggregator struct {
	storage    storage.Storage
	data       chan storage.Data
	recent     map[string]dataWithTime
	storeTimes map[string]time.Time
}

func (a *aggregator) GetChannel() chan<- storage.Data {
	return a.data
}

func (a *aggregator) run() {
	cleanupTicker := time.NewTicker(60 * time.Second)
	defer cleanupTicker.Stop()

	for {
		select {
		case d := <-a.data:
			a.process(d)
		case <-cleanupTicker.C:
			a.cleanup()
		}
	}
}

func (a *aggregator) process(d storage.Data) {
	// It is impossible that this data is missing when using
	// ADS-B.
	if d.ICAO == nil {
		return
	}

	// Store the live data temporarily
	a.recent[*d.ICAO] = dataWithTime{d, time.Now()}

	// If the position is set record the data permanently every couple of
	// seconds
	if d.Latitude != nil && d.Longitude != nil {
		t, ok := a.storeTimes[*d.ICAO]
		if !ok || time.Since(t) > storeEvery {
			if err := a.storage.Store(d); err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
			}
			a.storeTimes[*d.ICAO] = time.Now()
		}
	}
}

func (a *aggregator) cleanup() {
	for key, value := range a.recent {
		if time.Since(value.t) > dataTimeoutThreshold {
			delete(a.recent, key)
		}
	}

	for key, value := range a.storeTimes {
		if time.Since(value) > storeEvery {
			delete(a.storeTimes, key)
		}
	}

}

func (a *aggregator) Newest() map[string]storage.Data {
	rv := make(map[string]storage.Data)
	for key, value := range a.recent {
		rv[key] = value.Data
	}
	return rv
}
