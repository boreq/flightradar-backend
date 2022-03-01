package server

import (
	"fmt"
	"github.com/boreq/flightradar-backend/storage"
	"testing"
	"time"
)

func BenchmarkPolar(b *testing.B) {
	fakeStoredData := createFakeStoredData()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		toPolar(fakeStoredData)
	}
}

func TestPolar(t *testing.T) {
	var hashes []int64

	for i := 0; i < 100; i++ {
		fakeStoredData := createDeteministicStoredData()
		polar := toPolar(fakeStoredData)

		for k, v := range polar {
			t.Log(k, *v.Data.Data.Icao)
		}

		hash := calculateHash(polar)
		hashes = append(hashes, hash)

		if len(hashes) > 1 {
			if hashes[i] != hashes[i-1] {
				t.Fatal("execution is nondeterministic")
			}
		}
	}
}

func calculateHash(polar map[int]polarResponse) int64 {
	var result int64
	for k, v := range polar {
		result += int64(k)
		result += int64(v.Distance * 1000)
		result += v.Data.Time.Unix()
	}
	return result
}

func createDeteministicStoredData() []storage.StoredData {
	var rv []storage.StoredData
	for lat := 49.0; lat < 50.0; lat += 0.01 {
		for lon := 19.0; lon < 21.0; lon += 0.01 {
			data := storage.Data{
				Icao:         new(string),
				FlightNumber: new(string),
				Altitude:     new(int),
				Latitude:     new(float64),
				Longitude:    new(float64),
			}
			*data.Icao = fmt.Sprintf("icao-%d", len(rv))
			*data.FlightNumber = "flight"
			*data.Altitude = 2000
			*data.Latitude = lat
			*data.Longitude = lon
			storedData := storage.StoredData{
				Time: time.Date(1999, time.April, 11, 10, 11, 1, 1, time.UTC),
				Data: data,
			}
			rv = append(rv, storedData)
		}
	}
	return rv
}

func createFakeStoredData() []storage.StoredData {
	rv := make([]storage.StoredData, 0, 1000000)
	for lat := 49.0; lat < 50.0; lat += 0.01 {
		for lon := 19.0; lon < 21.0; lon += 0.01 {
			for _, icao := range []string{"aaaaaa", "bbbbbb", "cccccc"} {
				for _, flightNumber := range []string{"aaaaaa", "bbbbbb", "cccccc"} {
					for _, altitude := range []int{2000, 8000, 30000} {
						data := storage.Data{
							Icao:         new(string),
							FlightNumber: new(string),
							Altitude:     new(int),
							Latitude:     new(float64),
							Longitude:    new(float64),
						}
						*data.Icao = icao
						*data.FlightNumber = flightNumber
						*data.Altitude = altitude
						*data.Latitude = lat
						*data.Longitude = lon
						storedData := storage.StoredData{
							Time: time.Now(),
							Data: data,
						}
						rv = append(rv, storedData)
					}
				}
			}
		}
	}
	return rv
}
