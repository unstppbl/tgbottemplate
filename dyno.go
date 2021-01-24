package main

import (
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func (app *appStruct) wakeMyDyno() {
	ticker := time.NewTicker(time.Minute * app.wakePeriod)
	for {
		select {
		case <-ticker.C:
			resp, err := http.Get(app.baseURL + "ping")
			if err != nil {
				log.Errorf("Couldn't ping my dyno: %s", err)
				continue
			}
			if resp.StatusCode != http.StatusOK {
				log.Errorf("Wrong status code received: %s", resp.Status)
				resp.Body.Close()
				continue
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Errorf("Couldn't read response body from my dyno: %s", err)
				resp.Body.Close()
				continue
			}
			log.Infof("Succesfully pinged dyno: %s", string(body))
			resp.Body.Close()
		}
	}
}
