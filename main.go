package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("Slackbot Initialize")
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load godotenv: %v", err)
	}

	cfApiUrl := os.Getenv("CF_API_URL")
	cfUsername := os.Getenv("CF_USERNAME")
	cfPassword := os.Getenv("CF_PASSWORD")

	c := &cfclient.Config{
		ApiAddress: cfApiUrl,
		Username:   cfUsername,
		Password:   cfPassword,
	}
	client, err := cfclient.NewClient(c)
	if err != nil {
		log.Fatalf("Failed to initialize cfclient: %v", err)
	}

	apps, err := client.ListApps()
	if err != nil {
		log.Fatalf("Failed to list apps: %v", err)
	}

	for _, app := range apps {
		fmt.Println("-----------------------")
		fmt.Println("App Name: ", app.Name)
		fmt.Println("Status: ", app.State)
		fmt.Println("Memory: ", app.Memory)
		fmt.Println("Disk: ", app.DiskQuota)
		appStats, err := client.GetAppStats(app.Guid)
		if err != nil {
			fmt.Printf("Failed to get app status for %s: %v\n", app.Guid, err)
			continue
		}
		for k, stats := range appStats {
			appMemoryUsageMb := (float64(stats.Stats.Usage.Mem) / 1000000)
			appMemoryUsagePercent := appMemoryUsageMb / float64(app.Memory) * 100

			diskUsageMb := (float64(stats.Stats.Usage.Disk) / 1000000)
			diskUsagePercent := diskUsageMb / float64(app.DiskQuota) * 100

			fmt.Println("App Index ", k)
			fmt.Println("State: ", stats.State)
			fmt.Println("Uptime: ", stats.Stats.Uptime)
			fmt.Println("CPU Usage: ", stats.Stats.Usage.CPU)
			fmt.Println("Disk Usage MB: ", diskUsageMb)
			fmt.Println("Disk Usage Percent: ", diskUsagePercent)
			fmt.Println("Memory Usage MB: ", appMemoryUsageMb)
			fmt.Println("Memory Usage Percent: ", appMemoryUsagePercent)
			fmt.Println("Time Usage: ", stats.Stats.Usage.Time)
		}
	}

	// events, err := getEvents(client, 1)
	//
	// fmt.Printf("Found %d events\n", len(events))
	//
	// for _, appEvent := range events {
	// 	meta := appEvent.MetaData.Request
	// 	fmt.Println("-------")
	// 	fmt.Printf("App Name: %s -- State: %s\n", meta.Name, meta.State)
	// 	fmt.Printf("Event Type: %s -- Timestamp: %v\n", appEvent.EventType, appEvent.Timestamp)
	// 	fmt.Printf("Actor: %s -- Name: %s -- Type: %s\n", appEvent.Actor, appEvent.ActorName, appEvent.ActorType)
	//
	// 	if len(meta.ExitStatus) > 0 || len(meta.ExitReason) > 0 || len(meta.ExitDescription) > 0 {
	// 		fmt.Printf("Exit Status: %s -- Reason: %s -- Description: %s\n", meta.ExitStatus, meta.ExitReason, meta.ExitDescription)
	// 	}
	// }

	// slackAPIToken := os.Getenv("SLACK_API_TOKEN")
	// slack := newSlackClient(slackAPIToken)
	// slackMessage := newSlackMessage("platform-events", "test message")
	// err = slack.sendMessage(slackMessage)
	// if err != nil {
	// 	log.Fatalf("Failed to send slack message: %v", err)
	// }
}

func getEvents(client *cfclient.Client, pageNum int) ([]cfclient.AppEventEntity, error) {
	var events []cfclient.AppEventEntity

	//TODO - add timestamp support q=timestamp>2017-04-03T00:00:00Z&
	query := fmt.Sprintf("/v2/events?order-direction=desc&page=%d", pageNum)

	var eventResponse cfclient.AppEventResponse

	r := client.NewRequest("GET", query)
	resp, err := client.DoRequest(r)
	if err != nil {
		return nil, fmt.Errorf("Error requesting appevents: %v", err)
	}
	resBody, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("Error reading appevents response body %v", err)
	}

	err = json.Unmarshal(resBody, &eventResponse)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling appevent %v", err)
	}

	for _, resource := range eventResponse.Resources {
		events = append(events, resource.Entity)
	}

	if pageNum < eventResponse.Pages {
		pageNum++
		nextPage, err := getEvents(client, pageNum)
		if err != nil {
			return events, err
		}

		events = append(events, nextPage...)
	}

	return events, nil
}
