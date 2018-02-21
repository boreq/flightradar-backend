package aggregator

import (
	"github.com/boreq/flightradar-backend/storage"
	"testing"
	"time"
)

type st struct {
	counter int
}

func (s *st) Store(data storage.StoredData) error {
	s.counter++
	return nil
}

func (s *st) Retrieve(icao string) ([]storage.StoredData, error) {
	return nil, nil
}

func (s *st) RetrieveTimerange(from time.Time, to time.Time) ([]storage.StoredData, error) {
	return nil, nil
}

func (s *st) RetrieveAll() ([]storage.StoredData, error) {
	return nil, nil
}

func TestEnsureDataSavedOnceWhenTooOften(t *testing.T) {
	s := &st{}

	aggregator := New(s)

	data1 := storage.Data{
		Icao:      new(string),
		Latitude:  new(float64),
		Longitude: new(float64),
	}
	data2 := storage.Data{
		Icao:      new(string),
		Latitude:  new(float64),
		Longitude: new(float64),
	}
	*data1.Icao = "aaaaaaa"
	*data1.Latitude = 1
	*data1.Longitude = 1

	*data2.Icao = "aaaaaaa"
	*data2.Latitude = 2
	*data2.Longitude = 2

	aggregator.GetChannel() <- data1
	aggregator.GetChannel() <- data2

	<-time.After(1 * time.Second)
	if s.counter != 1 {
		t.Fatalf("Counter was %d", s.counter)
	}
}

func TestEnsureDataSavedOnceWhenIdentical(t *testing.T) {
	s := &st{}

	aggregator := New(s)

	data := storage.Data{
		Icao:      new(string),
		Latitude:  new(float64),
		Longitude: new(float64),
	}
	*data.Icao = "aaaaaaa"
	*data.Latitude = 1
	*data.Longitude = 1

	aggregator.GetChannel() <- data
	<-time.After(storeEveryTimeMin + 1*time.Second)
	aggregator.GetChannel() <- data

	<-time.After(1 * time.Second)
	if s.counter != 1 {
		t.Fatalf("Counter was %d", s.counter)
	}
}

func TestEnsureDataSavedTwiceWhenDifferent(t *testing.T) {
	s := &st{}

	aggregator := New(s)

	data1 := storage.Data{
		Icao:      new(string),
		Latitude:  new(float64),
		Longitude: new(float64),
	}
	data2 := storage.Data{
		Icao:      new(string),
		Latitude:  new(float64),
		Longitude: new(float64),
	}
	*data1.Icao = "aaaaaaa"
	*data1.Latitude = 1
	*data1.Longitude = 1

	*data2.Icao = "aaaaaaa"
	*data2.Latitude = 2
	*data2.Longitude = 2

	aggregator.GetChannel() <- data1
	<-time.After(storeEveryTimeMin + 1*time.Second)
	aggregator.GetChannel() <- data2

	<-time.After(1 * time.Second)
	if s.counter != 2 {
		t.Fatalf("Counter was %d", s.counter)
	}
}

func TestGetStoreEveryNilPointer(t *testing.T) {
	v := getStoreEvery(nil)
	if v != storeEveryTimeMin {
		t.Fatalf("Invalid d: %s != %s", v, storeEveryTimeMin)
	}

}

func TestStoreEveryLowerBoundary(t *testing.T) {
	altitude := storeEveryAltitudeMin
	d := getStoreEvery(&altitude)
	if d != storeEveryTimeMin {
		t.Fatalf("Invalid d: %s != %s", d, storeEveryTimeMin)
	}
}

func TestStoreEveryUpperBoundary(t *testing.T) {
	altitude := storeEveryAltitudeMax
	d := getStoreEvery(&altitude)
	if d != storeEveryTimeMax {
		t.Fatalf("Invalid d: %s != %s", d, storeEveryTimeMax)
	}
}
