package main

import (
	"aiscope/cmd/apiserver/app"
	"log"
)

func main() {

	cmd := app.NewAPIServerCommand()

	if err := cmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
