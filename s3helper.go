package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func getLastProcessedTimeFromS3(lookBack int, bucket, key string) (time.Time, error) {
	currentTime := time.Now().Add(-time.Duration(lookBack) * time.Minute)

	sess := session.Must(session.NewSession())

	downloader := s3manager.NewDownloader(sess)
	buf := aws.NewWriteAtBuffer([]byte{})

	_, err := downloader.Download(buf, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == s3.ErrCodeNoSuchKey {
			// Key doesn't exist
			lastProcessedTime := currentTime.Add(-time.Duration(maxLookBack) * time.Hour)
			return lastProcessedTime, nil
		}
		return time.Time{}, fmt.Errorf("error downloading file from S3: %v", err)
	}

	lastProcessedTime, err := time.Parse(time.RFC3339, string(buf.Bytes()))
	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing last processed time: %v", err)
	}

	if lastProcessedTime.Before(currentTime.Add(-maxLookBack * time.Hour)) {
		lastProcessedTime = currentTime.Add(-maxLookBack * time.Hour)
	}

	return lastProcessedTime, nil
}

// Store the last processed time to Amazon S3 in RFC3339 format
func storeLastProcessedTimeToS3(lastProcessedTime time.Time, bucket string, key string) error {
	// Create an AWS session
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return fmt.Errorf("error creating AWS session: %v", err)
	}

	// Create an S3 client
	svc := s3.New(sess)

	// Convert the last processed time to a string
	lastProcessedTimeStr := lastProcessedTime.Format(time.RFC3339)

	// Write the last processed time to a buffer
	buf := bytes.NewBufferString(lastProcessedTimeStr)

	// Upload the buffer to S3
	_, err = svc.PutObjectWithContext(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   io.ReadSeeker(bytes.NewReader(buf.Bytes())),
	})
	if err != nil {
		return fmt.Errorf("error uploading to S3: %v", err)
	}

	return nil
}
