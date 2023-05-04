package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const fileName = "lastProcessed.txt"
const maxLookBack = 360 // 6 hours in minutes

var (
	logsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "cloudflare_audit_logs_processed_total",
		Help: "The total number of processed events",
	})
)

// Get audit logs and process them until no more records are returned
func getAuditLogs(apiKey, apiEmail, orgId, s3Bucket string, lookBack int) error {
	api, err := cloudflare.New(apiKey, apiEmail)
	if err != nil {
		return fmt.Errorf("error creating Cloudflare API client: %v", err)
	}
	ctx := context.Background()
	userDetails, err := api.UserDetails(ctx)
	if err != nil {
		log.Fatal(err)
	}

	if len(userDetails.Email) > 0 {
		// Get current time minus look back and store in RFC3339
		var searchUntil time.Time
		var fileError error
		if len(s3Bucket) > 0 {
			searchUntil, fileError = getLastProcessedTimeFromS3(lookBack, s3Bucket, fileName)
		} else {
			searchUntil, fileError = getLastProcessedTime(lookBack, fileName)
		}
		if fileError != nil {
			log.Fatal(fileError)
		}

		// audit logs are returned in pages, we must continue to process until we run out of results
		pageNumber := 1
		for {
			filterOpts := cloudflare.AuditLogFilter{Since: searchUntil.Format(time.RFC3339), Page: pageNumber}
			results, err := api.GetOrganizationAuditLogs(context.Background(), orgId, filterOpts)
			if err != nil {
				return fmt.Errorf("error getting audit logs: %v", err)
			}

			if len(results.Result) == 0 {
				break
			}

			for _, record := range results.Result {
				b, _ := json.Marshal(record)
				logsProcessed.Inc()
				fmt.Println(string(b))
			}
			pageNumber++
		}
	}
	if len(s3Bucket) > 0 {
		if err := storeLastProcessedTimeToS3(time.Now(), s3Bucket, fileName); err != nil {
			return fmt.Errorf("error storing last processed time to S3: %v", err)
		}
	} else {
		if err := storeLastProcessedTimeToDisk(time.Now(), fileName); err != nil {
			return fmt.Errorf("error storing last processed time to disk: %v", err)
		}
	}
	return nil
}
