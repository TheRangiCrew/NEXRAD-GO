package server

import (
	"fmt"
	"sync"
	"time"
)

type Volume struct {
	ICAO     string
	InitTime time.Time
	Scans    []*Scan
}

var lock = &sync.Mutex{}

var queue *[]Volume

func Volumes() *[]Volume {
	lock.Lock()
	defer lock.Unlock()

	if queue == nil {
		queue = &[]Volume{}
	}

	return queue
}

func FindVolume(icao string, initTime time.Time) *Volume {
	scans := Volumes()

	for _, scan := range *scans {
		if scan.ICAO == icao && scan.InitTime == initTime {
			return &scan
		}
	}

	return nil
}

func AddToVolume(scans []*Scan, chunkData ChunkFileData) {
	volumes := Volumes()

	for _, s := range scans {

		volume := FindVolume(s.ICAO, chunkData.InitTime)

		if volume == nil {
			*volumes = append(*volumes, Volume{
				ICAO:     s.ICAO,
				InitTime: chunkData.InitTime,
				Scans:    []*Scan{s},
			})
			s.
		} else {
			other := FindScanElevationNumber(s, volume.Scans)
			if other == nil {
				volume.Scans = append(volume.Scans, s)
			} else {
				fmt.Println(len(other.Gates))
				other.Gates = append(other.Gates, s.Gates...)
				fmt.Println(len(other.Gates))
				fmt.Println(other.AzimuthResolution)
				fmt.Println(s.StartAzimuthNumber)

			}
		}
	}
}
