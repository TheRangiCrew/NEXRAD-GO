package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/TheRangiCrew/NEXRAD-GO/level2"
	nexrad "github.com/TheRangiCrew/NEXRAD-GO/level2/nexrad"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
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

var scanChan chan Scan

func Ingest(scanCh chan Scan) {
	scanChan = scanCh
	for {
		messagesRaw, err := GetMessages(10, 20)
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
				go func(payload Payload, mr types.Message) {

					HandlePayload(payload)
					DeleteMessages([]types.Message{mr})
				}(payload, mr)
			}
		} else {
			log.Println("No messages. Polling again...")
		}

	}
}

func HandlePayload(payload Payload) {
	object, err := GetObjectFromBucket(payload.S3Bucket, payload.Key)
	if err != nil {
		log.Println(err)
		return
	}
	chunkData, err := PayloadToChunkData(payload)
	if err != nil {
		log.Println(err)
		return
	}

	data := bytes.NewReader(object)

	l2Radar, err := nexrad.ParseNexrad(data)
	if err != nil {
		log.Println(err)
		return
	}

	if chunkData.ChunkType == "S" {
		tdwr, _ := level2.IsTDWRArchive(data)
		if tdwr {
			return
		}
		err := NewVolume(l2Radar, *chunkData)
		if err != nil {
			log.Println(err)
			return
		}
	} else {
		AddToVolume(l2Radar, *chunkData)
	}
}

func HandleFile(data io.ReadSeeker, chunkData ChunkFileData) {

	l2Radar, err := nexrad.ParseNexrad(data)
	if err != nil {
		log.Println(err)
		return
	}

	if chunkData.ChunkType == "S" {
		tdwr, _ := level2.IsTDWRArchive(data)
		if tdwr {
			return
		}
		err := NewVolume(l2Radar, chunkData)
		if err != nil {
			log.Println(err)
			return
		}
	} else {
		AddToVolume(l2Radar, chunkData)
	}
}

func NewVolume(l2Radar *nexrad.Nexrad, chunkData ChunkFileData) error {

	if l2Radar.ICAO == "" {
		return errors.New("radar data did not contain a valid ICAO")
	}

	site, err := GetSite(l2Radar.ICAO)
	if err != nil {
		return err
	}

	if site == nil {
		site, err = AddSite(l2Radar.ICAO)
		if err != nil {
			return err
		}
	}

	vcp := site.VCP
	if l2Radar.VCP != nil {
		vcp = int(l2Radar.VCP.Header.PatternNumber)
	}

	volume := Volume{
		ID:                    GetVolumeID(chunkData.InitTime, l2Radar.ICAO),
		InitTime:              chunkData.InitTime,
		VCP:                   vcp,
		CurrentElevation:      0,
		CurrentElevationAngle: 0.0,
	}

	_, err = Surreal().Create("radar_volume", volume)
	if err != nil {
		return err
	}

	_, err = Surreal().Query(fmt.Sprintf("RELATE radar_site:%s->radar_site_volume->radar_volume:%s", l2Radar.ICAO, volume.ID), map[string]string{})
	if err != nil {
		return err
	}

	scans := NexradToScans(l2Radar)
	if len(scans) > 0 {
		fmt.Println("This volume already has scans")
	}

	return nil
}

func AddToVolume(l2Radar *nexrad.Nexrad, chunkData ChunkFileData) (*Scan, error) {
	if l2Radar.ICAO == "" {
		return nil, errors.New("radar data did not contain a valid ICAO")
	}

	site, err := GetSite(l2Radar.ICAO)
	if err != nil {
		return nil, err
	}

	if site == nil {
		site, err = AddSite(l2Radar.ICAO)
		if err != nil {
			return nil, err
		}
	}

	volumeID := GetVolumeID(chunkData.InitTime, l2Radar.ICAO)

	volume, err := GetVolume(volumeID)
	if err != nil {
		return nil, err
	}
	if volume == nil {
		return nil, fmt.Errorf("no volume found")
	}

	newScans := NexradToScans(l2Radar)

	scans := Scans()

	for _, newScan := range newScans {
		// Assign current scan if it is nil
		var currentScan *Scan
		scanIndex := FindScanIndex(newScan, scans)
		if scanIndex == -1 {
			currentScan = &newScan
			*scans = append(*scans, *currentScan)
			scanIndex = len(*scans) - 1
		} else {
			currentScan = &(*scans)[scanIndex]
			*currentScan.Gates = append(*currentScan.Gates, *newScan.Gates...)
		}

		if newScan.EOE {
			currentScan.EOE = newScan.EOE
		}
		if newScan.EOV {
			currentScan.EOV = newScan.EOV
		}

		if (currentScan.EOE || currentScan.EOV) && volume.VCP != 0 {
			fmt.Printf("%s on elevation %d completed\n", currentScan.ProductType, currentScan.ElevationNumber)
			scanChan <- *currentScan
			fmt.Printf("Removing scan for %s\n", l2Radar.ICAO)
			_, err := RemoveScan(scanIndex, scans)
			if err != nil {
				return nil, err
			}
		}
	}

	return nil, nil
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
