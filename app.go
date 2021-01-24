package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

type appStruct struct {
	bot        *tgbotapi.BotAPI
	botToken   string
	baseURL    string
	db         *sql.DB
	redis      redis.Conn
	router     *gin.Engine
	port       string
	wakePeriod time.Duration
}

type userStruct struct {
	ID           int    `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`     // optional
	UserName     string `json:"username"`      // optional
	LanguageCode string `json:"language_code"` // optional
	Tz           string `json:"tz"`
	Location     string `json:"location"`
	meta         string
}

type groupStruct struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Tz       string `json:"tz"`
	Location string `json:"location"`
	meta     string
}

const (
	telegramMsgLimit   = 4096
	entriesPerMsgLimit = 10
)

func (app *appStruct) init() error {
	port := os.Getenv("PORT")
	if port == "" {
		return fmt.Errorf("$PORT must be set")
	}
	tgToken := os.Getenv("TOKEN")
	if tgToken == "" {
		return fmt.Errorf("$TOKEN must be set")
	}
	addr := os.Getenv("URL")
	if addr == "" {
		return fmt.Errorf("$URL must be set")
	}
	notificationPeriodStr := os.Getenv("APP_NOTIFICATION_PERIOD")
	period, err := strconv.Atoi(notificationPeriodStr)
	if err != nil {
		return fmt.Errorf("Couldn't convert notification period: %s", err)
	}
	log.Infof("Notification period is %d minutes\n", period)

	wakePeriodStr := os.Getenv("APP_WAKE_PERIOD")
	wakePeriod, err := strconv.Atoi(wakePeriodStr)
	if err != nil {
		return fmt.Errorf("Couldn't convert wake period: %s", err)
	}
	log.Infof("Wakeup period is %d minutes\n", wakePeriod)

	// app
	app.botToken = tgToken
	app.baseURL = addr
	app.port = port

	// Periods
	app.wakePeriod = time.Duration(wakePeriod)
	// Database
	if err := app.initDB(); err != nil {
		return fmt.Errorf("Error in dbinit: %s", err)
	}
	// telegram
	bot, err := tgbotapi.NewBotAPI(app.botToken)
	if err != nil {
		return err
	}
	app.bot = bot

	// get webhook info, and set if not set properly
	webhookInfo, err := app.bot.GetWebhookInfo()
	if err != nil {
		return fmt.Errorf("Couldn't get webhook info: %s", err)
	}
	whURL := app.baseURL + app.botToken
	if !webhookInfo.IsSet() || webhookInfo.URL != whURL {
		_, err = app.bot.SetWebhook(tgbotapi.NewWebhook(whURL))
		if err != nil {
			return err
		}
	}

	// gin router
	router := gin.New()
	router.Use(gin.Logger())
	// register webhook handler
	router.POST("/"+app.bot.Token, app.webhookHandler)
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	app.router = router
	return nil
}

func (app *appStruct) run() (err error) {
	// dyno wakeup goroutine
	go app.wakeMyDyno()
	// start server
	return app.router.Run(":" + app.port)
}
