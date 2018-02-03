package storage

type Storage interface {
}

type Data struct {
	ICAO            *string
	FlightNumber    *string
	TransponderCode *int
	Altitude        *int
	Speed           *int
	Heading         *int
	Latitude        *float64
	Longitude       *float64
}

type StoredData struct {
	Data
}
