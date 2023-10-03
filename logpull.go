package main

import (
	"context"
	"encoding/json"
	"fmt"
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
		return err
	}

	if len(userDetails.Email) == 0 {
		return nil
	}

	var searchUntil time.Time
	if len(s3Bucket) > 0 {
		searchUntil, err = getLastProcessedTimeFromS3(lookBack, s3Bucket, fileName)
	} else {
		searchUntil, err = getLastProcessedTime(lookBack, fileName)
	}
	if err != nil {
		return err
	}

	if err := processAuditLogs(ctx, api, orgId, searchUntil); err != nil {
		return err
	}

	if len(s3Bucket) > 0 {
		return storeLastProcessedTimeToS3(time.Now(), s3Bucket, fileName)
	}
	return storeLastProcessedTimeToDisk(time.Now(), fileName)
}

func processAuditLogs(ctx context.Context, api *cloudflare.API, orgId string, searchUntil time.Time) error {
	pageNumber := 1
	for {
		filterOpts := cloudflare.AuditLogFilter{Since: searchUntil.Format(time.RFC3339), Page: pageNumber}
		results, err := api.GetOrganizationAuditLogs(ctx, orgId, filterOpts)
		if err != nil {
			return fmt.Errorf("error getting audit logs: %v", err)
		}

		if len(results.Result) == 0 {
			break
		}

		for _, record := range results.Result {
			b, err := json.Marshal(record)
			if err != nil {
				return err
			}
			logsProcessed.Inc()
			fmt.Println(string(b))
		}
		pageNumber++
	}
	return nil
}
