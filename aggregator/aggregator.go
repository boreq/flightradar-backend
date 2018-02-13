package aggregator

import (
	"fmt"
	"github.com/boreq/flightradar-backend/logging"
	"github.com/boreq/flightradar-backend/storage"
	"os"
	"time"
)

var log = logging.GetLogger("aggregator")

func New(s storage.Storage) Aggregator {
	rv := &aggregator{
		storage: s,
		data:    make(chan storage.Data),
		recent:  make(map[string]storage.StoredData),
		stored:  make(map[string]storage.StoredData),
	}
	go rv.run()
	return rv
}

// storeEvery specifies how often the data is permanently stored.
const storeEvery = 5 * time.Second

// dataTimeoutThreshold specifies at which point the cached newest data is
// considered outdated - how long plane's data can be saved or retrieved
// after it disappears.
const dataTimeoutThreshold = 15 * time.Second

// storedDataTimeoutThreshold specifies how long the stored data points are
// held here for reference - this makes sure that two data points with the
// same position don't get saved for the given aircraft twice in the row.
const storedDataTimeoutThreshold = 5 * time.Minute

type aggregator struct {
	storage storage.Storage
	data    chan storage.Data
	recent  map[string]storage.StoredData
	stored  map[string]storage.StoredData
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
	// I think it is impossible that this data is missing when using ADS-B,
	// but check just to be sure.
	if d.Icao == nil {
		return
	}

	storedData := storage.StoredData{Data: d, Time: time.Now()}
	a.recent[*d.Icao] = storedData

	// If the position is set record the data permanently every couple of
	// seconds but only if the position doesn't duplicate the already stored
	// data.
	if d.Latitude != nil && d.Longitude != nil {
		lastStoredData, ok := a.stored[*d.Icao]
		if !ok || time.Since(lastStoredData.Time) > storeEvery {
			if !ok || (*storedData.Data.Latitude != *lastStoredData.Data.Latitude &&
				*storedData.Data.Longitude != *lastStoredData.Data.Longitude) {
				if err := a.storage.Store(storedData); err != nil {
					fmt.Fprintf(os.Stderr, "%s\n", err)
				}
				a.stored[*d.Icao] = storedData
			}
		}
	}
}

func (a *aggregator) cleanup() {
	for key, value := range a.recent {
		if time.Since(value.Time) > dataTimeoutThreshold {
			delete(a.recent, key)
		}
	}

	for key, value := range a.stored {
		if time.Since(value.Time) > storedDataTimeoutThreshold {
			delete(a.stored, key)
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

func (a *aggregator) Retrieve(icao string) ([]storage.StoredData, error) {
	return a.storage.Retrieve(icao)
}

func (a *aggregator) RetrieveTimerange(from time.Time, to time.Time) ([]storage.StoredData, error) {
	return a.storage.RetrieveTimerange(from, to)
}

func (a *aggregator) RetrieveAll() ([]storage.StoredData, error) {
	return a.storage.RetrieveAll()
}
