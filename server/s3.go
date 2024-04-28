package server

import (
	"context"
	"io"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func GetObjectFromBucket(s3Client s3.Client, bucketName string, objectKey string) ([]byte, error) {
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
