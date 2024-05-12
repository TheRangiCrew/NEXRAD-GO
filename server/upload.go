package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func Upload(scanChan chan Scan) {

	for scan := range scanChan {
		go push(scan)
	}
}

func push(scan Scan) {
	jsonData, err := json.Marshal(scan)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	t := scan.InitTime

	year := strconv.Itoa(t.Year())
	month := PadZero(strconv.Itoa(int(t.Month())), 2)
	day := PadZero(strconv.Itoa(t.Day()), 2)

	hour := PadZero(strconv.Itoa(t.Hour()), 2)
	minute := PadZero(strconv.Itoa(t.Minute()), 2)
	second := PadZero(strconv.Itoa(t.Second()), 2)

	key := year + "/" + month + "/" + day + "/" + scan.ICAO + "/" + hour + "-" + minute + "-" + second + "-" + scan.ProductType + "-" + strconv.Itoa(scan.ElevationNumber)

	// Create an io.Reader from the JSON data
	reader := bytes.NewReader(jsonData)
	fmt.Println(key)
	uploader := manager.NewUploader(S3Client())
	_, err = uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String("witsnexrad"),
		Key:    aws.String(key + ".json"),
		Body:   reader,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("Uploaded " + key)
}
