package storage

import (
	"time"
)

type ReadStorage interface {
	Retrieve(icao string) (<-chan StoredData, error)
}

type WriteStorage interface {
	Store(data Data) error
}

type Storage interface {
	ReadStorage
	WriteStorage
}

type Data struct {
	ICAO            *string  `json:"icao"`
	FlightNumber    *string  `json:"flight_number"`
	TransponderCode *int     `json:"transponder_code"`
	Altitude        *int     `json:"altitude"`
	Speed           *int     `json:"speed"`
	Heading         *int     `json:"heading"`
	Latitude        *float64 `json:"latitude"`
	Longitude       *float64 `json:"longitude"`
}

type StoredData struct {
	Data Data      `json:"data"`
	Time time.Time `json:"time"`
}
