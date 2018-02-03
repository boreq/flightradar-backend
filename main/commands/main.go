package commands

import (
	"fmt"
	"github.com/boreq/guinea"
)

var buildCommit string
var buildDate string

var MainCmd = guinea.Command{
	Options: []guinea.Option{
		guinea.Option{
			Name:        "version",
			Type:        guinea.Bool,
			Description: "Display version",
		},
	},
	Run: runMain,
	Subcommands: map[string]*guinea.Command{
		"run":            &runCmd,
		"default_config": &defaultConfigCmd,
	},
	ShortDescription: "SDR plane tracking software",
	Description:      "This software records plane tracking data collected by SDR radios.",
}

func runMain(c guinea.Context) error {
	if c.Options["version"].Bool() {
		fmt.Printf("BuildCommit %s\n", buildCommit)
		fmt.Printf("BuildDate %s\n", buildDate)
		return nil
	}
	return guinea.ErrInvalidParms
}
