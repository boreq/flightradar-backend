// Package main contains the main function.
package main

import (
	"fmt"
	"github.com/boreq/guinea"
	"github.com/boreq/flightradar-backend/main/commands"
	"os"
)

func main() {
	e := guinea.Run(&commands.MainCmd)
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
