package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

// we need a small amount of state to check the last time logs were checked as in
// the event our program crashes or is rescheduled, we may need to look back
// further than the look back interval. We will cap the maximum look back at 6
// hours to ensure we don't abuse the cloudflare API.
func getLastProcessedTime(lookBack int, fileName string) (time.Time, error) {
	currentTime := time.Now().Add(time.Duration(-lookBack))
	var lastProcessedTime time.Time
	fileContents, err := ioutil.ReadFile(fileName)

	if err != nil {
		if os.IsNotExist(err) {
			// This must be the first run ever, we'll scrape 6 hours and then start
			lastProcessedTime = currentTime.Add(time.Duration(-maxLookBack) * time.Minute)
			return lastProcessedTime, nil
		}
		return lastProcessedTime, fmt.Errorf("error reading file: %v", err)
	}

	if len(fileContents) == 0 {
		// This must be the first run ever, we'll scrape 6 hours and then start
		lastProcessedTime = currentTime.Add(time.Duration(-maxLookBack) * time.Minute)
		return lastProcessedTime, nil
	}

	lastProcessedTime, err = time.Parse(time.RFC3339, string(fileContents))
	if err != nil {
		return lastProcessedTime, fmt.Errorf("error parsing last processed time: %v", err)
	}

	if lastProcessedTime.Before(currentTime.Add(-maxLookBack * time.Hour)) {
		lastProcessedTime = currentTime.Add(-maxLookBack * time.Hour)
	}

	return lastProcessedTime, nil
}

// Store the last processed time to file in RFC3339 format
func storeLastProcessedTimeToDisk(lastProcessedTime time.Time, fileName string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(lastProcessedTime.Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}
	return nil
}
