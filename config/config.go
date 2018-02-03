// Package config holds the global config.
package config

import (
	"encoding/json"
	"io/ioutil"
)

type ConfigStruct struct {
	Debug           bool
	ServeAddress    string
	Dump1090Address string
	DataDirectory   string
}

// Config points to the current config struct used by the other parts of the
// program.
var Config *ConfigStruct = Default()

// Default returns the default config.
func Default() *ConfigStruct {
	conf := &ConfigStruct{
		Debug:           false,
		ServeAddress:    "127.0.0.1:8118",
		Dump1090Address: "127.0.0.1:8080",
		DataDirectory:   "data",
	}
	return conf
}

// Load loads the config from the specified json file. If certain keys are not
// present in the loaded config file the current values of the config struct
// are preserved.
func Load(filename string) error {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(content, Config)
}
