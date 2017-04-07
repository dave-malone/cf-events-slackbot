package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"time"

	cfclient "github.com/dave-malone/go-cfclient"
)

const (
	eventsTimeFormat = "2006-01-02 15:04:05"
)

type cfClient struct {
	client *cfclient.Client
}

func newCloudFoundryClient(apiURL, username, password string) (*cfClient, error) {
	c := &cfclient.Config{
		ApiAddress: apiURL,
		Username:   username,
		Password:   password,
	}

	client, err := cfclient.NewClient(c)
	return &cfClient{client: client}, err
}

func (c *cfClient) checkApps() error {
	apps, err := c.client.ListApps()
	if err != nil {
		return fmt.Errorf("Failed to list apps: %v", err)
	}

	for _, app := range apps {
		fmt.Println("-----------------------")
		fmt.Println("App Name: ", app.Name)
		fmt.Println("Status: ", app.State)
		fmt.Println("Memory: ", app.Memory)
		fmt.Println("Disk: ", app.DiskQuota)
		appStats, err := c.client.GetAppStats(app.Guid)
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

	return nil
}

func (c *cfClient) checkEvents(eventsSince time.Time) error {
	events, err := c.getEvents(eventsSince, 1)
	if err != nil {
		return err
	}

	fmt.Printf("Found %d events\n", len(events))
	for _, appEvent := range events {
		meta := appEvent.MetaData.Request
		fmt.Println("-------")
		fmt.Println("App Name: ", meta.Name)
		fmt.Println("State: ", meta.State)
		fmt.Println("Event Type: ", appEvent.EventType)
		fmt.Println("Timestamp: ", appEvent.Timestamp)
		fmt.Println("Actor: ", appEvent.Actor)
		fmt.Println("Name: ", appEvent.ActorName)
		fmt.Println("Type: ", appEvent.ActorType)

		if len(meta.ExitStatus) > 0 || len(meta.ExitReason) > 0 || len(meta.ExitDescription) > 0 {
			fmt.Println("Exit Status: ", meta.ExitStatus)
			fmt.Println("Reason: ", meta.ExitReason)
			fmt.Println("Description: ", meta.ExitDescription)
		}
	}

	return nil
}

func (c *cfClient) getEvents(eventsSince time.Time, pageNum int) ([]cfclient.AppEventEntity, error) {
	var events []cfclient.AppEventEntity

	params := url.Values{}
	params.Add("order-direction", "desc")
	params.Add("page", fmt.Sprintf("%d", pageNum))
	if !eventsSince.IsZero() {
		timeFilter := fmt.Sprintf("timestamp>%s", eventsSince.Format(eventsTimeFormat))
		params.Add("q", timeFilter)
	}

	//TODO - this won't be necessary, and I'll be able to use the go-cfclient functions once this issue is fixed: https://github.com/cloudfoundry/cloud_controller_ng/issues/803
	query := fmt.Sprintf("/v2/events?%v", params.Encode())
	fmt.Println(query)

	var eventResponse cfclient.AppEventResponse

	r := c.client.NewRequest("GET", query)
	resp, err := c.client.DoRequest(r)
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
		nextPage, err := c.getEvents(eventsSince, pageNum)
		if err != nil {
			return events, err
		}

		events = append(events, nextPage...)
	}

	return events, nil
}
