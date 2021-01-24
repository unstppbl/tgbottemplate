package main

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"

	_ "github.com/heroku/x/hmetrics/onload"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}
	tgToken := os.Getenv("TOKEN")
	if tgToken == "" {
		log.Fatal("$TOKEN must be set")
	}
	addr := os.Getenv("URL")
	if addr == "" {
		log.Fatal("$URL must be set")
	}
	apiToken := os.Getenv("API_TOKEN")
	if apiToken == "" {
		log.Fatal("$API_TOKEN token must be set")
	}
	notificationPeriodStr := os.Getenv("APP_NOTIFICATION_PERIOD")
	period, err := strconv.Atoi(notificationPeriodStr)
	if err != nil {
		log.Fatalf("Couldn't convert notification period: %s", err)
	}
	log.Infof("Notification period is %d minutes\n", period)

	wakePeriodStr := os.Getenv("APP_WAKE_PERIOD")
	wakePeriod, err := strconv.Atoi(wakePeriodStr)
	if err != nil {
		log.Fatalf("Couldn't convert wake period: %s", err)
	}
	log.Infof("Wakeup period is %d minutes\n", wakePeriod)

	// gin router
	router := gin.New()
	router.Use(gin.Logger())

	// app
	app := appStruct{
		botToken: tgToken,
		baseURL:  addr,
		apiToken: apiToken,
	}
	// Periods
	app.notificationPeriod = time.Duration(period)
	app.wakePeriod = time.Duration(wakePeriod)
	// Redis
	redisURL := os.Getenv("REDIS_URL")
	c, err := redis.DialURL(strings.Replace(redisURL, "://h:", "://:", 1))
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	app.redis = c
	// Database
	if err := app.initDB(); err != nil {
		log.Fatal(err)
	}
	// telegram
	bot, err := tgbotapi.NewBotAPI(app.botToken)
	if err != nil {
		log.Fatal(err)
	}
	app.bot = bot

	// this perhaps should be conditional on GetWebhookInfo()
	// only set webhook if it is not set properly
	url := app.baseURL + app.botToken
	_, err = app.bot.SetWebhook(tgbotapi.NewWebhook(url))
	if err != nil {
		log.Fatal(err)
	}

	// register webhook handler
	router.POST("/"+app.bot.Token, app.webhookHandler)
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	// start notifications worker
	go app.notificationsWorker()
	// dyno wakeup goroutine
	go app.wakeMyDyno()
	// start server
	err = router.Run(":" + port)
	if err != nil {
		log.Fatal(err)
	}
}
