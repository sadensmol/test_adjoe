package main

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func main() {
	qURL := "http://sqs.eu-central-1.localhost.localstack.cloud:4566/000000000000/my-queue"

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region:      aws.String("eu-central-1"),
			Endpoint:    aws.String("http://localhost:4566"),
			Credentials: credentials.NewStaticCredentials("test", "test", ""),

			CredentialsChainVerboseErrors: aws.Bool(true),
		},
	}))
	sqc := sqs.New(sess)

	timeout := time.After(666 * time.Second) // Set a timeout for the reader
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			fmt.Println("Test timed out")
			return
		case <-ticker.C:
			result2, err := sqc.ReceiveMessage(&sqs.ReceiveMessageInput{
				MaxNumberOfMessages: aws.Int64(1),
				WaitTimeSeconds:     aws.Int64(3),
				QueueUrl:            aws.String(qURL),
			})
			if err != nil {
				log.Fatalf("error reading from sqs: %v", err)
			}
			for _, msg := range result2.Messages {
				// Process the received message
				fmt.Println(*msg.Body)
				// Delete the message
				_, err := sqc.DeleteMessage(&sqs.DeleteMessageInput{
					QueueUrl:      aws.String(qURL),
					ReceiptHandle: msg.ReceiptHandle,
				})
				if err != nil {
					log.Fatalf("error deleting message: %v", err)
				}
			}
		}
	}
}
