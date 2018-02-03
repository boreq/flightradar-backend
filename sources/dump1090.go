package sources

import (
	"encoding/json"
	"fmt"
	"github.com/boreq/flightradar/repository"
	"net/http"
	"time"
)

type dump1090Data struct {
	ICAO              string  `json:"hex"`
	Squawk            string  `json:"squawk"`
	Flight            string  `json:"flight"`
	Latitude          float64 `json:"lat"`
	Longitude         float64 `json:"lon"`
	ValidPosition     bool    `json:"valid_position"`
	Altitude          float64 `json:"altitude"`
	VerticalRate      float64 `json:"vert_rate"`
	Heading           float64 `json:"track"`
	ValidHeading      float64 `json:"validtrack"`
	Speed             float64 `json:"speed"`
	NumeberOfMessages int     `json:"messages"`
	Seen              int     `json:"seen"`
}

func NewDump1090(address string, data chan<- repository.Data) error {
	ticker := time.NewTicker(time.Millisecond * 1)
	go func() {
		for t := range ticker.C {
			var data dump1090Data
			err := getDump1090Data(address, &data)
			if err == nil {
				if data.seen < 60 {
					var resultData respository.Data
					resultData.ICAO = data.Hex
					resultData.Flight = data.Flight
					resultData.Squawk = data.Squawk
					if data.ValidPosition {
						resultData.Latitude = data.Latitude
						resultData.Longitude = data.Longitude
					}
					resultData.Altitude = data.Altitude
					resultData.VerticalRate = data.Altitude
				}
			}
		}
	}()
	return nil
}

func getDump1090Data(address string, data interface{}) error {
	var client = &http.Client{Timeout: 10 * time.Second}
	r, err := client.Get(address)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(data)
}
