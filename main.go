package main

import (
	log "github.com/sirupsen/logrus"

	_ "github.com/heroku/x/hmetrics/onload"
)

func main() {
	app := appStruct{}
	if err := app.init(); err != nil {
		log.Fatal(err)
	}
	if err := app.run(); err != nil {
		log.Fatal(err)
	}
}
