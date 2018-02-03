package logging

import (
	"log"
)

type logger struct {
	logger *log.Logger
}

func (l *logger) Print(v ...interface{}) {
	l.logger.Print(v...)
}

func (l *logger) Printf(format string, v ...interface{}) {
	l.logger.Printf(format, v...)
}

func (l *logger) Debug(v ...interface{}) {
	if *debug {
		l.logger.Print(v...)
	}
}

func (l *logger) Debugf(format string, v ...interface{}) {
	if *debug {
		l.logger.Printf(format, v...)
	}
}
