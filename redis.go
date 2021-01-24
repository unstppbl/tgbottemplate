package main

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"
)

func (app *appStruct) initRedis() {
	// Redis
	redisURL := os.Getenv("REDIS_URL")
	c, err := redis.DialURL(strings.Replace(redisURL, "://h:", "://:", 1))
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	app.redis = c
}

func (app *appStruct) reconnectToRedis() {
	redisURL := os.Getenv("REDIS_URL")
	c, err := redis.DialURL(strings.Replace(redisURL, "://h:", "://:", 1))
	if err != nil {
		log.Errorf("Couldn't connect to redis: %s", err)
	}
	app.redis = c
}

func (app *appStruct) getUserState(id int) (string, error) {
	str, err := redis.String(app.redis.Do("GET", fmt.Sprintf("user:%d:state", id)))
	if err != nil {
		if e, ok := err.(*net.OpError); ok {
			log.Errorf("Redis net error: %s", e)
			app.reconnectToRedis()
			return app.getUserState(id)
		}
		return "", err
	}
	return str, nil
}

func (app *appStruct) setUserState(id int, state string) (err error) {
	_, err = app.redis.Do("SET", fmt.Sprintf("user:%d:state", id), state)
	if err != nil {
		if e, ok := err.(*net.OpError); ok {
			log.Errorf("Redis net error: %s", e)
			app.reconnectToRedis()
			return app.setUserState(id, state)

		}
		return err
	}
	return nil
}

func (app *appStruct) deleteUserState(id int) (err error) {
	_, err = app.redis.Do("DEL", fmt.Sprintf("user:%d:state", id))
	if err != nil {
		if e, ok := err.(*net.OpError); ok {
			log.Errorf("Redis net error: %s", e)
			app.reconnectToRedis()
			return app.deleteUserState(id)

		}
		return err
	}
	return nil
}
