package sources

import (
	"encoding/json"
	"fmt"
	"github.com/boreq/flightradar-backend/storage"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type dump1090Data struct {
	ICAO              string  `json:"hex"`
	TransponderCode   string  `json:"squawk"`
	FlightNumber      string  `json:"flight"`
	Latitude          float64 `json:"lat"`
	Longitude         float64 `json:"lon"`
	ValidPosition     int     `json:"validposition"`
	Altitude          int     `json:"altitude"`
	VerticalRate      int     `json:"vert_rate"`
	Heading           int     `json:"track"`
	ValidHeading      int     `json:"validtrack"`
	Speed             int     `json:"speed"`
	NumeberOfMessages int     `json:"messages"`
	Seen              int     `json:"seen"`
}

const DataAgeThreshold = 10 // DataAgeThreshold specifies after how many seconds the data is considered obsolete and will be rejected.

func NewDump1090(address string, dataChan chan<- storage.Data) error {
	ticker := time.NewTicker(time.Second * 1)
	go func() {
		for range ticker.C {
			var datas []dump1090Data

			err := getDump1090Data(address, &datas)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting Dump1090 data: %s\n", err)
				continue
			}

			for _, dataTmp := range datas {
				var data = dataTmp
				if data.Seen < DataAgeThreshold {
					var resultData storage.Data
					if data.ICAO != "" {
						resultData.ICAO = &data.ICAO
					}
					if data.FlightNumber != "" {
						trim := strings.TrimSpace(data.FlightNumber)
						resultData.FlightNumber = &trim
					}
					if data.TransponderCode != "" {
						if v, err := strconv.Atoi(data.TransponderCode); err == nil {
							resultData.TransponderCode = &v
						}
					}
					if data.ValidPosition != 0 {
						resultData.Latitude = &data.Latitude
						resultData.Longitude = &data.Longitude
					}
					resultData.Altitude = &data.Altitude
					resultData.Speed = &data.Speed
					if data.ValidHeading != 0 {
						resultData.Heading = &data.Heading
					}

					dataChan <- resultData
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
