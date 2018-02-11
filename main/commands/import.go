package commands

import (
	"bufio"
	"encoding/json"
	"github.com/boreq/flightradar-backend/storage"
	"github.com/boreq/guinea"
	"os"
	"sync"
)

var importCmd = guinea.Command{
	Run: runImport,
	Arguments: []guinea.Argument{
		{"config", false, "Config file"},
		{"source", false, "Source file"},
	},
	ShortDescription: "imports data from a file",
}

func runImport(c guinea.Context) error {
	st, err := initialize(c.Arguments[0])
	if err != nil {
		return err
	}

	file, err := os.Open(c.Arguments[1])
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var storedData storage.StoredData
		if err := json.Unmarshal(scanner.Bytes(), &storedData); err != nil {
			return err
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := st.Store(storedData); err != nil {
				panic(err)
			}
		}()
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	wg.Wait()

	return nil
}
