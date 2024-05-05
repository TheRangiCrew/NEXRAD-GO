package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	nexrad "github.com/TheRangiCrew/NEXRAD-GO/level2/nexrad"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/joho/godotenv"
)

type Message struct {
	Type      string `json:"Type"`
	MessageId string `json:"MessageId"`
	TopicArn  string `json:"TopicArn"`
	Message   string `json:"Message"`
	Timestamp string `json:"Timestamp"`
}

type Payload struct {
	S3Bucket  string `json:"S3Bucket"`
	Key       string `json:"Key"`
	SiteID    string `json:"SiteID"`
	DateTime  string `json:"DateTime"`
	VolumeID  int    `json:"VolumeID"`
	ChunkID   int    `json:"ChunkID"`
	ChunkType string `json:"ChunkType"`
	L2Version string `json:"L2Version"`
}

type ChunkFileData struct {
	Site      string
	InitTime  time.Time
	Number    int
	ChunkType string
}

func FilenameToChunkData(filename string) (*ChunkFileData, error) {
	segments := strings.Split(filename, "-")

	dateString := segments[0]
	timeString := segments[1]

	year, err := strconv.Atoi(dateString[0:4])
	month, err := strconv.Atoi(dateString[4:6])
	day, err := strconv.Atoi(dateString[6:])

	hour, err := strconv.Atoi(timeString[0:2])
	minute, err := strconv.Atoi(timeString[2:4])
	second, err := strconv.Atoi(timeString[4:])

	if err != nil {
		return nil, err
	}

	time := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)

	number, err := strconv.Atoi(segments[2])
	if err != nil {
		return nil, err
	}

	chunkType := segments[3]

	return &ChunkFileData{
		InitTime:  time,
		Number:    number,
		ChunkType: chunkType,
	}, nil
}

func PayloadToChunkData(payload Payload) (*ChunkFileData, error) {

	t, err := time.Parse("2006-01-02T15:04:05", payload.DateTime)
	if err != nil {
		return nil, err
	}

	return &ChunkFileData{
		Site:      payload.SiteID,
		InitTime:  t,
		Number:    payload.ChunkID,
		ChunkType: payload.ChunkType,
	}, nil
}

func ParseNewChunk(data io.ReadSeeker, chunkData ChunkFileData) error {

	fmt.Printf("Parsed new chunk %s %d %s %s\n", chunkData.Site, chunkData.Number, chunkData.ChunkType, chunkData.InitTime)

	l2Radar, err := nexrad.ParseNexrad(data)
	if err != nil {
		return err
	}

	scans := NexradToScans(l2Radar)

	AddToVolume(scans, chunkData)

	return nil
}

// GetMessages uses the ReceiveMessage action to get messages from an Amazon SQS queue.
func GetMessages(actor sqs.Client, maxMessages int32, waitTime int32) ([]types.Message, error) {
	var messages []types.Message
	result, err := actor.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(os.Getenv("NEXRAD_SQS_URL")),
		MaxNumberOfMessages: maxMessages,
		WaitTimeSeconds:     waitTime,
	})
	if err != nil {
		log.Printf("Couldn't get messages from queue %v. Here's why: %v\n", os.Getenv("NEXRAD_SQS_URL"), err)
	} else {
		messages = result.Messages
	}
	return messages, err
}

func StartServer() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}

	fmt.Println(os.Getenv("AWS_REGION"))

	sdkConfig, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		fmt.Println("Couldn't load default configuration. Have you set up your AWS account?")
		fmt.Println(err)
		return
	}

	s3Client := s3.NewFromConfig(sdkConfig)
	sqsClient := sqs.NewFromConfig(sdkConfig)

	for {
		if SurrealInit() != nil {
			log.Printf("Failed to connect to DB: %s\nTrying again in 30 seconds\n\n", err)
			time.Sleep(30 * time.Second)
			continue
		}
		messagesRaw, err := GetMessages(*sqsClient, 10, 20)
		if err != nil {
			log.Fatal(err)
		}
		if len(messagesRaw) != 0 {
			for _, mr := range messagesRaw {
				var message Message
				err := json.Unmarshal([]byte(*mr.Body), &message)
				if err != nil {
					log.Println(err)
				}
				var payload Payload
				err = json.Unmarshal([]byte(message.Message), &payload)
				if err != nil {
					log.Println(err)
				}
				go func() {
					data, err := GetObjectFromBucket(*s3Client, payload.S3Bucket, payload.Key)
					if err != nil {
						log.Println(err)
					} else {
						chunkData, err := PayloadToChunkData(payload)
						if err != nil {
							log.Println(err)
						} else {
							ParseNewChunk(bytes.NewReader(data), *chunkData)
						}
					}
				}()
			}
		} else {
			log.Println("No messages. Polling again...")
		}
	}
}
