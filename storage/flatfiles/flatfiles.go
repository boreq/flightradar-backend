package flatfiles

import (
	"bufio"
	"encoding/json"
	"errors"
	"github.com/boreq/flightradar-backend/logging"
	"github.com/boreq/flightradar-backend/storage"
	"os"
	"path"
	"regexp"
	"time"
)

var log = logging.GetLogger("storage/flatfiles")

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
	filepath, err := f.getFileForPlane(*data.ICAO)
	if err != nil {
		return err
	}

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

var icaoRegex = regexp.MustCompile("[^a-zA-Z0-9]+")

func (f *flatfiles) getFileForPlane(icao string) (string, error) {
	if icaoRegex.ReplaceAllString(icao, "") != icao {
		return "", errors.New("ICAO must be alphanumeric")
	}
	return path.Join(f.directory, "planes", icao), nil
}

func (f *flatfiles) Retrieve(icao string) (<-chan storage.StoredData, error) {
	filepath, err := f.getFileForPlane(icao)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	c := make(chan storage.StoredData)
	go func() {
		defer file.Close()
		defer close(c)

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			var storedData storage.StoredData
			print(scanner.Text())
			print("\n")
			err := json.Unmarshal(scanner.Bytes(), &storedData)
			if err != nil {
				log.Printf("Error for %s: %s", icao, err)
				continue
			}
			c <- storedData
		}

		if err := scanner.Err(); err != nil {
			log.Printf("Error for %s: %s", icao, err)
		}
	}()
	return c, nil
}
