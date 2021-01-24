package main

import (
	log "github.com/sirupsen/logrus"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func (app *appStruct) sendMsg(msg tgbotapi.Chattable) {
	_, err := app.bot.Send(msg)
	if err != nil {
		log.Errorf("[!] Error sending message: %s", err)
	}
}
