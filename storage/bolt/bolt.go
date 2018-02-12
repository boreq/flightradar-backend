package bolt

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/boltdb/bolt"
	"github.com/boreq/flightradar-backend/logging"
	"github.com/boreq/flightradar-backend/storage"
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

func New(filepath string) (storage.Storage, error) {
	// Open the database, create it if needed. Timeout ensures that the
	// function will not block idefinietly.
	db, err := bolt.Open(filepath, 0600, &bolt.Options{Timeout: 10 * time.Second})
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

	j, err := json.Marshal(data)
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

	err := b.db.View(func(tx *bolt.Tx) error {
		planesB := tx.Bucket(planesKey)
		if planesB == nil {
			return errors.New("Planes bucket does not exist!")
		}

		planeB := planesB.Bucket([]byte(icao))
		if planeB != nil {
			planeB.ForEach(func(k, v []byte) error {
				var storedData storage.StoredData
				err := json.Unmarshal(v, &storedData)
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

	err := b.db.View(func(tx *bolt.Tx) error {
		generalB := tx.Bucket(generalKey)
		if generalB == nil {
			return errors.New("General bucket does not exist!")
		}

		c := generalB.Cursor()
		min := timeToKey(from)
		max := timeToKey(to)

		for k, v := c.Seek(min); k != nil && bytes.Compare(k[0:30], max) <= 0; k, v = c.Next() {
			var storedData storage.StoredData
			err := json.Unmarshal(v, &storedData)
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

	err := b.db.View(func(tx *bolt.Tx) error {
		generalB := tx.Bucket(generalKey)
		if generalB == nil {
			return errors.New("General bucket does not exist!")
		}

		generalB.ForEach(func(k, v []byte) error {
			var storedData storage.StoredData
			err := json.Unmarshal(v, &storedData)
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

func timeToKey(t time.Time) []byte {
	t = t.UTC()
	s := t.Format(rfc3339NanoSortable)
	return []byte(s)
}

func timeAndIcaoToKey(t time.Time, icao string) []byte {
	return append(timeToKey(t), []byte(icao)...)
}
