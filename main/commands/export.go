package commands

import (
	"encoding/json"
	"github.com/boreq/guinea"
	"os"
)

var exportCmd = guinea.Command{
	Run: runExport,
	Arguments: []guinea.Argument{
		{"config", false, "Config file"},
		{"destination", false, "Destination file"},
	},
	ShortDescription: "exports data to a file",
}

func runExport(c guinea.Context) error {
	storage, err := initialize(c.Arguments[0])
	if err != nil {
		return err
	}

	file, err := os.OpenFile(c.Arguments[1], os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	data, err := storage.RetrieveAll()
	if err != nil {
		return err
	}

	for _, d := range data {
		j, err := json.Marshal(d)
		if err != nil {
			return err
		}
		j = append(j, []byte("\n")...)
		if _, err := file.Write(j); err != nil {
			return err
		}
	}

	if err := file.Close(); err != nil {
		return err
	}

	return nil
}
