package main

import (
	"encoding/json"
	"io/ioutil"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

const (
	stateSubscribe   = "sub"
	stateUnsubscribe = "unsub"
)

func (app *appStruct) webhookHandler(c *gin.Context) {
	defer c.Request.Body.Close()

	bytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Println(err)
		return
	}

	var update tgbotapi.Update
	err = json.Unmarshal(bytes, &update)
	if err != nil {
		log.Println(err)
		return
	}
	// Do nothing if message is nil
	if update.Message == nil {
		return
	}

	// to monitor changes run: heroku logs --tail
	log.Printf("From: %+v Text: %+v", update.Message.From, update.Message.Text)
	app.handleMessage(&update)
}

func (app *appStruct) handleMessage(update *tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	msg.Text = "Hello, mate!"
	app.sendMsg(msg)
}
