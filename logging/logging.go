package logging

import (
	"github.com/boreq/flightradar/config"
	"log"
	"os"
)

// Logger defines methods used for logging in normal mode and debug mode. Debug
// mode log messages are displayed only if a proper field is set in the config.
type Logger interface {
	Print(...interface{})
	Printf(string, ...interface{})
	Debug(...interface{})
	Debugf(string, ...interface{})
}

var debug *bool
var loggers map[string]Logger

func init() {
	loggers = make(map[string]Logger)
	debug = &config.Config.Debug
}

// GetLogger creates a new logger or returns an already existing logger created
// with the given name using this method.
func GetLogger(name string) Logger {
	if _, ok := loggers[name]; !ok {
		loggers[name] = &logger{log.New(os.Stdout, name+": ", 0)}
	}
	return loggers[name]
}
