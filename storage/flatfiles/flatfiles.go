package flatfiles

import (
	"encoding/json"
	"errors"
	"github.com/boreq/flightradar-backend/storage"
	"os"
	"path"
	"time"
)

func New(directory string) storage.Storage {
	rv := &flatfiles{directory: directory}
	return rv
}

type flatfiles struct {
	directory string
}

func (f *flatfiles) Store(data storage.Data) error {
	// Sanity
	if data.ICAO == nil || *data.ICAO == "" {
		return errors.New("ICAO can't be empty")
	}

	if data.Latitude == nil || data.Longitude == nil {
		return errors.New("There is no point in saving the data without the position")
	}

	// Marshal
	j, err := json.Marshal(storage.StoredData{data, time.Now().UTC()})
	if err != nil {
		return err
	}
	j = append(j, []byte("\n")...)

	// Create directories
	filepath := f.getFileForPlane(*data.ICAO)
	if err := os.MkdirAll(path.Dir(filepath), 0755); err != nil {
		return err
	}

	// Append to the file
	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	if _, err := file.Write(j); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}

	return nil
}

func (f *flatfiles) getFileForPlane(icao string) string {
	return path.Join(f.directory, "planes", icao)
}
