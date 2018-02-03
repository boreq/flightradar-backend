package commands

import (
	"github.com/boreq/flightradar/aggregator"
	"github.com/boreq/flightradar/config"
	"github.com/boreq/guinea"
)

var runCmd = guinea.Command{
	Run: runRun,
	Arguments: []guinea.Argument{
		{"config", false, "Config file"},
	},
	ShortDescription: "runs the program",
}

func runRun(c guinea.Context) error {
	configFilename := c.Arguments[0]

	if err := config.Load(configFilename); err != nil {
		return err
	}

	// Run the data collection
	aggr := aggregator.New()

	// Serve the collected data

	return nil
}
