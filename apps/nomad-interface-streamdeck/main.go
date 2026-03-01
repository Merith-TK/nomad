package main

import (
	"log"
)

func main() {
	app := NewApp()

	if err := app.Init(); err != nil {
		log.Fatal(err)
	}
	defer app.Shutdown()

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
