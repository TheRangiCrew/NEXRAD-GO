package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/joho/godotenv"
)

func main() {
	Init(true)

	scanChan := make(chan Scan)

	go Ingest(scanChan)

	go Upload(scanChan)

	select {}
}

func Init(env bool) {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("error loading .env file: " + err.Error())
	}

	sdkConfig, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		fmt.Println("Couldn't load default configuration. Have you set up your AWS account?")
		fmt.Println(err)
		return
	}

	S3Init(sdkConfig)
	SQSInit(sdkConfig)

	for {

		if SurrealInit() != nil {
			log.Printf("Failed to connect to DB: %s\nTrying again in 30 seconds\n\n", err)
			time.Sleep(30 * time.Second)
		} else {
			break
		}
	}
}
