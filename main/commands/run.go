package commands

import (
	"github.com/boreq/flightradar/aggregator"
	"github.com/boreq/flightradar/config"
	"github.com/boreq/flightradar/server"
	"github.com/boreq/flightradar/sources"
	"github.com/boreq/flightradar/storage/flatfiles"
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
	storage := flatfiles.New(config.Config.DataDirectory)
	aggr := aggregator.New(storage)
	if err := sources.NewDump1090(config.Config.Dump1090Address, aggr.GetChannel()); err != nil {
		return err
	}

	// Serve the collected data
	if err := server.Serve(aggr, config.Config.ServeAddress); err != nil {
		return err
	}

	return nil
}
