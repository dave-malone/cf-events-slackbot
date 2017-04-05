package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	postMessageURL = "https://slack.com/api/chat.postMessage"
)

type slack struct {
	apiToken string
}

type slackMessage struct {
	message string
	channel string
}

func newSlackClient(apiToken string) *slack {
	return &slack{apiToken: apiToken}
}

func newSlackMessage(channel, message string) *slackMessage {
	return &slackMessage{
		message: message,
		channel: channel,
	}
}

func (s *slack) sendMessage(message *slackMessage) error {
	formValues := url.Values{"token": {s.apiToken}, "channel": {message.channel}, "text": {message.message}}

	resp, err := http.PostForm(postMessageURL, formValues)
	if err != nil {
		return fmt.Errorf("Failed to http post: %v", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Failed to read response body: %v", err)
	}

	fmt.Println("response body: ", string(body))

	return nil
}
