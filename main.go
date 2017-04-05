package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("Slackbot Initialize")
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load godotenv: %v", err)
	}

	slackAPIToken := os.Getenv("SLACK_API_TOKEN")
	slack := newSlackClient(slackAPIToken)
	slackMessage := newSlackMessage("platform-events", "test message")
	err := slack.sendMessage(slackMessage)
	if err != nil {
		log.Fatalf("Failed to send slack message: %v", err)
	}
}
