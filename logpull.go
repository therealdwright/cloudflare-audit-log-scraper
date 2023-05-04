package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/cloudflare/cloudflare-go"
)

const fileName = "lastProcessed.txt"
const maxLookBack = 360 // 6 hours in minutes

// we need a small amount of state to check the last time logs were checked as in
// the event our program crashes or is rescheduled, we may need to look back
// further than the look back interval. We will cap the maximum look back at 6
// hours to ensure we don't abuse the cloudflare API.
func getLastProcessedTime() time.Time {
	currentTime := time.Now()
	var lastProcessedTime time.Time
	fileContents, err := ioutil.ReadFile(fileName)

	// if the file has contents, we need to process it
	if err == nil && len(fileContents) > 0 {
		lastProcessedTime, err = time.Parse(time.RFC3339, string(fileContents))
		if err != nil {
			log.Fatalf("Error parsing last processed time: %v", err)
		}
		if lastProcessedTime.Before(currentTime.Add(-maxLookBack * time.Hour)) {
			lastProcessedTime = currentTime.Add(-maxLookBack * time.Hour)
		}
	} else {
		// this must be the first run ever, we'll scrape 6 hours and then start
		lastProcessedTime = currentTime.Add(time.Duration(-maxLookBack) * time.Minute)
	}
	return lastProcessedTime
}

// Store the last processed time to file in RFC3339 format
func storeLastProcessedTimeToDisk(lastProcessedTime time.Time) error {
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("Error creating file: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(lastProcessedTime.Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("Error writing to file: %v", err)
	}
	return nil
}

// Get audit logs and process them until no more records are returned
func getAuditLogs(apiKey, apiEmail, orgId string, lookBack int) error {
	api, err := cloudflare.New(apiKey, apiEmail)
	if err != nil {
		return fmt.Errorf("Error creating Cloudflare API client: %v", err)
	}

	// Get current time minus look back and store in RFC3339
	searchUntil := getLastProcessedTime()

	// audit logs are returned in pages, we must continue to process until we run out of results
	pageNumber := 1
	for {
		filterOpts := cloudflare.AuditLogFilter{Since: searchUntil.Format(time.RFC3339), Page: pageNumber}
		results, err := api.GetOrganizationAuditLogs(context.Background(), orgId, filterOpts)
		if err != nil {
			return fmt.Errorf("Error getting audit logs: %v", err)
		}

		if len(results.Result) == 0 {
			break
		}

		for _, record := range results.Result {
			b, _ := json.Marshal(record)
			fmt.Println(string(b))
		}
		pageNumber++
	}
	if err := storeLastProcessedTimeToDisk(time.Now()); err != nil {
		return fmt.Errorf("Error storing last processed time to disk: %v", err)
	}
	return nil
}
