package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

var sqsClient *sqs.Client

func SQSClient() *sqs.Client {
	return sqsClient
}

func SQSInit(sdkConfig aws.Config) {
	sqsClient = sqs.NewFromConfig(sdkConfig)
}

// GetMessages uses the ReceiveMessage action to get messages from an Amazon SQS queue.
func GetMessages(maxMessages int32, waitTime int32) ([]types.Message, error) {
	var messages []types.Message
	result, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(os.Getenv("NEXRAD_SQS_URL")),
		MaxNumberOfMessages: maxMessages,
		WaitTimeSeconds:     waitTime,
	})
	if err != nil {
		log.Printf("Couldn't get messages from queue %v. Here's why: %v\n", os.Getenv("NEXRAD_SQS_URL"), err)
	} else {
		messages = result.Messages
		fmt.Printf("Received %d messages\n", len(messages))
	}
	return messages, err
}

func DeleteMessages(messages []types.Message) {
	entries := make([]types.DeleteMessageBatchRequestEntry, len(messages))
	for msgIndex := range messages {
		entries[msgIndex].Id = aws.String(fmt.Sprintf("%v", msgIndex))
		entries[msgIndex].ReceiptHandle = messages[msgIndex].ReceiptHandle
	}
	_, err := sqsClient.DeleteMessageBatch(context.TODO(), &sqs.DeleteMessageBatchInput{
		Entries:  entries,
		QueueUrl: aws.String(os.Getenv("NEXRAD_SQS_URL")),
	})
	if err != nil {
		log.Printf("Couldn't delete messages from queue. Here's why: %v\n", err)
	}
}
