package commands

import (
	"github.com/boreq/flightradar-backend/config"
	"github.com/boreq/flightradar-backend/database"
	"github.com/boreq/guinea"
)

var createDbCmd = guinea.Command{
	Run: runCreateDb,
	Arguments: []guinea.Argument{
		{"config", false, "Config file"},
	},
	ShortDescription: "creates database tables",
}

func runCreateDb(c guinea.Context) error {
	configFilename := c.Arguments[0]
	if err := config.Load(configFilename); err != nil {
		return err
	}

	if err := database.Init(database.SQLite3, config.Config.DatabaseFile); err != nil {
		return err
	}

	return database.CreateTables()
}
