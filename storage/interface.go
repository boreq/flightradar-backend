package storage

import (
	"time"
)

type ReadStorage interface {
	Retrieve(icao string) ([]StoredData, error)
	RetrieveTimerange(from time.Time, to time.Time) ([]StoredData, error)
	RetrieveAll() ([]StoredData, error)
}

type WriteStorage interface {
	Store(data StoredData) error
}

type Storage interface {
	ReadStorage
	WriteStorage
}

type Data struct {
	Icao            *string  `json:"icao,omitempty"`
	FlightNumber    *string  `json:"flight_number,omitempty"`
	TransponderCode *int     `json:"transponder_code,omitempty"`
	Altitude        *int     `json:"altitude,omitempty"`
	Speed           *int     `json:"speed,omitempty"`
	Heading         *int     `json:"heading,omitempty"`
	Latitude        *float64 `json:"latitude,omitempty"`
	Longitude       *float64 `json:"longitude,omitempty"`
}

type StoredData struct {
	Data Data      `json:"data"`
	Time time.Time `json:"time"`
}
