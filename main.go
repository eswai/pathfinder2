package main

import (
	"log"

	"github.com/eswai/pathfinder2/ui"
)

func main() {
	app, err := ui.NewApp()
	if err != nil {
		log.Fatal(err)
	}
	app.Run()
}
