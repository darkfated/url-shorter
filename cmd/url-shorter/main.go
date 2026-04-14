package main

import (
	"log"

	"github.com/darkfated/url-shorter/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
