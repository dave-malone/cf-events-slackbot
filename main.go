package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("Slackbot Initialize")
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load godotenv: %v", err)
	}

	//TODO - get this from VCAP_SERVICES
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASS")
	redisDb, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		log.Fatalf("REDIS_DB must be an int value; was %s", os.Getenv("REDIS_DB"))
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDb,
	})

	pong, err := redisClient.Ping().Result()
	fmt.Println(pong, err)

	//TODO - get this from VCAP_SERVICES
	cfAPI := os.Getenv("CF_API_URL")
	cfUsername := os.Getenv("CF_USERNAME")
	cfPassword := os.Getenv("CF_PASSWORD")
	cfClient, err := newCloudFoundryClient(cfAPI, cfUsername, cfPassword)
	if err != nil {
		log.Fatalf("Failed to initialize cfclient: %v", err)
	}

	//TODO - get this from VCAP_SERVICES
	slackAPIToken := os.Getenv("SLACK_API_TOKEN")
	slackClient := newSlackClient(slackAPIToken)

	slackBot := newSlackBot(cfClient, slackClient)

	for {
		if lastRunStr, err := redisClient.Get("cf-slackbot-lastrun").Result(); err == nil {
			lastRun, err := time.Parse(eventsTimeFormat, lastRunStr)
			if err != nil {
				log.Fatalf("Failed to time.Parse %s with format %s", lastRunStr, eventsTimeFormat)
			}

			slackBot.lastRun = lastRun
		}

		if err := slackBot.execute(); err != nil {
			fmt.Println("Slackbot encountered an issue... ", err.Error())
		}

		if _, err := redisClient.Set("cf-slackbot-lastrun", time.Now().Format(eventsTimeFormat), 0).Result(); err != nil {
			fmt.Println("Failed to save cf-slackbot-lastrun in Redis: ", err)
		}

		time.Sleep(300 * time.Second)
	}

}
