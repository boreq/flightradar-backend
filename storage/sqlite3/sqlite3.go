package sqlite3

import (
	"fmt"
	"github.com/boreq/flightradar-backend/database"
	"github.com/boreq/flightradar-backend/logging"
	"github.com/boreq/flightradar-backend/storage"
	"time"
)

var log = logging.GetLogger("storage/flatfiles")

func New() storage.Storage {
	rv := &sqlite3{}
	return rv
}

type sqlite3 struct{}

type storedDataOut struct {
	Id              int
	Icao            *string
	FlightNumber    *string
	TransponderCode *int
	Altitude        *int
	Speed           *int
	Heading         *int
	Latitude        *float64
	Longitude       *float64
	Time            int64
}

func (s *sqlite3) Store(data storage.StoredData) error {
	now := time.Now()
	if _, err := database.DB.Exec("INSERT INTO planes (icao, latitude, longitude, flight_number, transponder_code, altitude, speed, heading, time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)", data.Data.Icao, data.Data.Latitude, data.Data.Longitude, data.Data.FlightNumber, data.Data.TransponderCode, data.Data.Altitude, data.Data.Speed, data.Data.Heading, data.Time.Unix()); err != nil {
		return err
	}
	fmt.Printf("%f\n", time.Since(now).Seconds())

	return nil
}

func (s *sqlite3) Retrieve(icao string) (<-chan storage.StoredData, error) {
	var data []storedDataOut
	if err := database.DB.Select(&data,
		`SELECT *
		FROM planes
		WHERE icao == $1`, icao); err != nil {
		return nil, err
	}
	return sliceAsChannel(data)
}

func (s *sqlite3) RetrieveTimerange(from time.Time, to time.Time) (<-chan storage.StoredData, error) {
	var data []storedDataOut
	if err := database.DB.Select(&data,
		`SELECT *
		FROM planes
		WHERE time >= $1 AND time <= $2
		`, from.Unix(), to.Unix()); err != nil {
		panic(err)
		return nil, err
	}
	print(len(data))
	return sliceAsChannel(data)
}

func sliceAsChannel(data []storedDataOut) (<-chan storage.StoredData, error) {
	c := make(chan storage.StoredData)
	go func() {
		defer close(c)
		for _, d := range data {
			c <- convert(d)
		}
	}()
	return c, nil
}

func convert(data storedDataOut) storage.StoredData {
	var rv storage.StoredData
	rv.Data.Icao = data.Icao
	rv.Data.FlightNumber = data.FlightNumber
	rv.Data.TransponderCode = data.TransponderCode
	rv.Data.Altitude = data.Altitude
	rv.Data.Speed = data.Speed
	rv.Data.Heading = data.Heading
	rv.Data.Latitude = data.Latitude
	rv.Data.Longitude = data.Longitude
	rv.Time = time.Unix(data.Time, 0)
	return rv
}
