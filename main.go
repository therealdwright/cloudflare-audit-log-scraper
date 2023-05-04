package main

import (
	"log"
	"os"
	"time"

	"github.com/go-co-op/gocron"
)

func runCronJobs(apiKey, apiEmail, orgId string, lookBack int) {
	s := gocron.NewScheduler(time.UTC)

	s.Every(lookBack).Minutes().Do(func() {
		getAuditLogs(apiKey, apiEmail, orgId, lookBack)
	})

	s.StartBlocking()
}

func main() {
	apiEmail := os.Getenv("CLOUDFLARE_API_EMAIL")
	apiKey := os.Getenv("CLOUDFLARE_API_KEY")
	orgId := os.Getenv("CLOUDFLARE_ORGANIZATION_ID")
	interval := os.Getenv("CLOUDFLARE_LOOK_BACK_INTERVAL")

	if apiEmail == "" {
		log.Fatal("Must specify CLOUDFLARE_API_EMAIL")
	}
	if apiKey == "" {
		log.Fatal("Must specify CLOUDFLARE_API_KEY")
	}
	if orgId == "" {
		log.Fatal("Must specify CLOUDFLARE_ORGANIZATION_ID")
	}

	// Convert user supplied look back value to an integer otherwise set a default
	var lookBack int
	if interval != "" {
		envToInt, err := getEnvInt(interval)
		if err != nil {
			log.Fatal("If CLOUDFLARE_LOOK_BACK_INTERVAL it must be an integer")
		}
		lookBack = envToInt
	}
	lookBack = 5

	runCronJobs(apiKey, apiEmail, orgId, lookBack)
}
