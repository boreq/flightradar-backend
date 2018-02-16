package bolt

import (
	"bytes"
	"errors"
	"github.com/boltdb/bolt"
	"github.com/boreq/flightradar-backend/logging"
	"github.com/boreq/flightradar-backend/storage"
	"github.com/boreq/flightradar-backend/storage/bolt/messages"
	"github.com/golang/protobuf/proto"
	"io"
	"time"
)

var log = logging.GetLogger("storage/bolt")

// Key for the top level bucket which contains all data points.
var generalKey = []byte("all")

// Key got the top level bucket which contains plane specific buckets.
var planesKey = []byte("planes")

// The RFC3339 format provided in the standard library is not sortable due to
// the verying number of nanosecond digits.
const rfc3339NanoSortable = "2006-01-02T15:04:05.000000000Z07:00"

type Bolt interface {
	storage.Storage
	io.Closer
}

func New(filepath string) (Bolt, error) {
	// Open the database, create it if needed. Timeout ensures that the
	// function will not block idefinietly.
	db, err := bolt.Open(filepath, 0644, &bolt.Options{Timeout: 10 * time.Second})
	if err != nil {
		return nil, err
	}

	// Precreate the buckets - makes the future writes faster.
	err = db.Update(func(tx *bolt.Tx) error {
		// General bucket.
		if _, err := tx.CreateBucketIfNotExists(generalKey); err != nil {
			return err
		}
		// Plane specific bucket.
		if _, err := tx.CreateBucketIfNotExists(planesKey); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	rv := &blt{
		db: db,
	}
	return rv, nil
}

type blt struct {
	db *bolt.DB
}

func (b *blt) Store(data storage.StoredData) error {
	if data.Data.Icao == nil || *data.Data.Icao == "" {
		return errors.New("ICAO can't be empty!")
	}

	j, err := encode(data)
	if err != nil {
		return err
	}

	err = b.db.Batch(func(tx *bolt.Tx) error {
		// Store the data in the general bucket.
		generalB := tx.Bucket(generalKey)
		if generalB == nil {
			return errors.New("General bucket does not exist!")
		}
		if err := generalB.Put(timeAndIcaoToKey(data.Time, *data.Data.Icao), j); err != nil {
			return err
		}

		// Store the data in the plane specific bucket.
		planesB := tx.Bucket(planesKey)
		if planesB == nil {
			return errors.New("Planes bucket does not exist!")
		}
		planeB, err := planesB.CreateBucketIfNotExists([]byte(*data.Data.Icao))
		if err != nil {
			return err
		}
		if err := planeB.Put(timeToKey(data.Time), j); err != nil {
			return err
		}

		return nil
	})

	return err
}

func (b *blt) Retrieve(icao string) ([]storage.StoredData, error) {
	var rv []storage.StoredData

	t := time.Now()
	defer func() {
		log.Debugf("Retrieve: %f seconds", time.Since(t).Seconds())
	}()

	err := b.db.View(func(tx *bolt.Tx) error {
		planesB := tx.Bucket(planesKey)
		if planesB == nil {
			return errors.New("Planes bucket does not exist!")
		}

		planeB := planesB.Bucket([]byte(icao))
		if planeB != nil {
			planeB.ForEach(func(k, v []byte) error {
				storedData, err := decode(v)
				if err != nil {
					return err
				}
				rv = append(rv, storedData)
				return nil
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return rv, nil
}

func (b *blt) RetrieveTimerange(from time.Time, to time.Time) ([]storage.StoredData, error) {
	var rv []storage.StoredData

	t := time.Now()
	defer func() {
		log.Debugf("Retrieve timerange: %f seconds", time.Since(t).Seconds())
	}()

	err := b.db.View(func(tx *bolt.Tx) error {
		generalB := tx.Bucket(generalKey)
		if generalB == nil {
			return errors.New("General bucket does not exist!")
		}

		c := generalB.Cursor()
		min := timeToKey(from)
		max := timeToKey(to)

		for k, v := c.Seek(min); k != nil && bytes.Compare(k[0:30], max) <= 0; k, v = c.Next() {
			storedData, err := decode(v)
			if err != nil {
				return err
			}
			rv = append(rv, storedData)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return rv, nil
}

func (b *blt) RetrieveAll() ([]storage.StoredData, error) {
	var rv []storage.StoredData

	t := time.Now()
	defer func() {
		log.Debugf("Retrieve all: %f seconds", time.Since(t).Seconds())
	}()

	err := b.db.View(func(tx *bolt.Tx) error {
		generalB := tx.Bucket(generalKey)
		if generalB == nil {
			return errors.New("General bucket does not exist!")
		}

		generalB.ForEach(func(k, v []byte) error {
			storedData, err := decode(v)
			if err != nil {
				return err
			}
			rv = append(rv, storedData)
			return nil
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	return rv, nil
}

func (b *blt) Close() error {
	return b.db.Close()
}

func timeToKey(t time.Time) []byte {
	t = t.UTC()
	s := t.Format(rfc3339NanoSortable)
	return []byte(s)
}

func timeAndIcaoToKey(t time.Time, icao string) []byte {
	return append(timeToKey(t), []byte(icao)...)
}

func encode(storedData storage.StoredData) ([]byte, error) {
	protoStoredData := &messages.StoredData{
		Time: new(int64),
		Data: &messages.Data{
			Icao:         storedData.Data.Icao,
			FlightNumber: storedData.Data.FlightNumber,
			Latitude:     storedData.Data.Latitude,
			Longitude:    storedData.Data.Longitude,
		},
	}
	*protoStoredData.Time = storedData.Time.Unix()
	if storedData.Data.TransponderCode != nil {
		transponderCode := int32(*storedData.Data.TransponderCode)
		protoStoredData.Data.TransponderCode = &transponderCode
	}
	if storedData.Data.Altitude != nil {
		altitude := int32(*storedData.Data.Altitude)
		protoStoredData.Data.Altitude = &altitude
	}
	if storedData.Data.Speed != nil {
		speed := int32(*storedData.Data.Speed)
		protoStoredData.Data.Speed = &speed
	}
	if storedData.Data.Heading != nil {
		heading := int32(*storedData.Data.Heading)
		protoStoredData.Data.Heading = &heading
	}
	return proto.Marshal(protoStoredData)
}

func decode(data []byte) (storage.StoredData, error) {
	var protoStoredData messages.StoredData
	err := proto.Unmarshal(data, &protoStoredData)

	rv := storage.StoredData{
		Time: time.Unix(*protoStoredData.Time, 0),
		Data: storage.Data{
			Icao:         protoStoredData.Data.Icao,
			FlightNumber: protoStoredData.Data.FlightNumber,
			Latitude:     protoStoredData.Data.Latitude,
			Longitude:    protoStoredData.Data.Longitude,
		},
	}
	if protoStoredData.Data.TransponderCode != nil {
		transponderCode := int(*protoStoredData.Data.TransponderCode)
		rv.Data.TransponderCode = &transponderCode
	}
	if protoStoredData.Data.Altitude != nil {
		altitude := int(*protoStoredData.Data.Altitude)
		rv.Data.Altitude = &altitude
	}
	if protoStoredData.Data.Speed != nil {
		speed := int(*protoStoredData.Data.Speed)
		rv.Data.Speed = &speed
	}
	if protoStoredData.Data.Heading != nil {
		heading := int(*protoStoredData.Data.Heading)
		rv.Data.Heading = &heading
	}

	return rv, err
}
