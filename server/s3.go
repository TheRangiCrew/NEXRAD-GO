package main

import (
	"context"
	"io"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var s3Client *s3.Client

func S3Client() *s3.Client {
	return s3Client
}

func S3Init(sdkConfig aws.Config) {
	s3Client = s3.NewFromConfig(sdkConfig)
}

func GetObjectFromBucket(bucketName string, objectKey string) ([]byte, error) {
	result, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		log.Printf("Couldn't get object %v:%v. Here's why: %v\n", bucketName, objectKey, err)
		return nil, err
	}
	data, err := io.ReadAll(result.Body)
	if err != nil {
		log.Printf("Couldn't read object data")
		return nil, err
	}
	return data, nil
}
