package main

import (
	"fmt"
	"time"
)

type slackBot struct {
	cfClient    *cfClient
	slackClient *slack
	lastRun     time.Time
}

func newSlackBot(cfClient *cfClient, slackClient *slack) *slackBot {
	return &slackBot{
		cfClient:    cfClient,
		slackClient: slackClient,
	}
}

func (bot *slackBot) execute() error {
	fmt.Println("Slackbot execute... last run: ", bot.lastRun)

	if err := bot.cfClient.checkApps(); err != nil {
		return fmt.Errorf("Failed to check apps: %v", err)
	}

	if err := bot.cfClient.checkEvents(bot.lastRun); err != nil {
		return fmt.Errorf("Failed to check events: %v", err)
	}

	slackMessage := newSlackMessage("platform-events", "test message")
	if err := bot.slackClient.sendMessage(slackMessage); err != nil {
		return fmt.Errorf("Failed to send slack message: %v", err)
	}

	return nil
}
