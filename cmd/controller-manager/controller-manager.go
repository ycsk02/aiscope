package main

import (
	"aiscope/cmd/controller-manager/app"
	"os"
)

func main() {
	command := app.NewControllerManagerCommand()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
