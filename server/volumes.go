package server

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/TheRangiCrew/NEXRAD-GO/level2/nexrad"
)

type Volume struct {
	ID                    string    `json:"id,omitempty"`
	InitTime              time.Time `json:"init_time"`
	FinishTime            time.Time `json:"finish_time,omitempty"`
	VCP                   int       `json:"vcp"`
	CurrentElevation      int       `json:"current_elevation_number"`
	CurrentElevationAngle int       `json:"current_elevation_angle"`
	//Scans                 *[]Scan
}

var lock = &sync.Mutex{}

var queue *[]Scan

func Scans() *[]Scan {
	lock.Lock()
	defer lock.Unlock()

	if queue == nil {
		queue = &[]Scan{}
	}

	return queue
}

func RemoveScan(slice *[]Scan, scanToRemove *Scan) *[]Scan {
	for i, scan := range *slice {
		if &scan == scanToRemove {
			// Swap the element to be removed with the last element
			(*slice)[i] = (*slice)[len(*slice)-1]
			// Truncate the slice by one element
			*slice = (*slice)[:len(*slice)-1]
			return slice
		}
	}
	// If the scan is not found, return the original slice
	return slice
}

func GetVolumeID(t time.Time, icao string) string {
	year := strconv.Itoa(t.Year())
	month := PadZero(strconv.Itoa(int(t.Month())), 2)
	day := PadZero(strconv.Itoa(t.Day()), 2)

	hour := PadZero(strconv.Itoa(t.Hour()), 2)
	minute := PadZero(strconv.Itoa(t.Minute()), 2)
	second := PadZero(strconv.Itoa(t.Second()), 2)

	return year + month + day + hour + minute + second + icao
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

func FindVolume(icao string, initTime time.Time) *Volume {
	//Surreal().Query()

	return nil
}

func AddToVolume(l2Radar *nexrad.Nexrad, chunkData ChunkFileData) error {
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

	volumeID := GetVolumeID(chunkData.InitTime, l2Radar.ICAO)

	volume, err := GetVolume(volumeID)
	if err != nil {
		return err
	}

	newScans := NexradToScans(l2Radar)

	scans := Scans()

	for _, newScan := range newScans {
		// Assign current scan if it is nil
		currentScan := FindScanElevationNumber(newScans[0], scans)
		if currentScan == nil {
			currentScan = &newScan
			*scans = append(*scans, *currentScan)
		} else {
			fmt.Println(len(currentScan.Gates))
			currentScan.Gates = append(currentScan.Gates, newScan.Gates...)
			fmt.Println(len(currentScan.Gates))
		}

		if newScan.EOE {
			currentScan.EOE = newScan.EOE
		}
		if newScan.EOV {
			currentScan.EOV = newScan.EOV
		}

		if (currentScan.EOE || currentScan.EOV) && volume.VCP != 0 {
			fmt.Printf("%s on elevation %d completed\n", currentScan.ProductType, currentScan.ElevationNumber)
			// AddSite(currentScan, *volume)
			fmt.Printf("Removing scan for %s\n", l2Radar.ICAO)
			RemoveScan(scans, currentScan)
		}
	}

	return nil
}

// func ToVolume(l2Radar *nexrad.Nexrad, chunkData ChunkFileData) {
// 	vcp := 0
// 	if l2Radar.VCP != nil {
// 		vcp = int(l2Radar.VCP.Header.PatternNumber)
// 	}

// 	scans := NexradToScans(l2Radar)

// 	volume := FindVolume(chunkData.Site, chunkData.InitTime)

// 	if volume == nil {
// 		if chunkData.ChunkType != "S" {
// 			return
// 		}
// 		// If none add it
// 		volume = &Volume{
// 			ICAO:     chunkData.Site,
// 			InitTime: chunkData.InitTime,
// 			VCP:      vcp,
// 			Scans:    &[]Scan{},
// 		}
// 		*volumes = append(*volumes, *volume)
// 		fmt.Printf("Adding volume for %s\n", volume.ICAO)

// 	}

// 	if l2Radar.ElevationScans[1] != nil {
// 		volume.Elevation = int(l2Radar.ElevationScans[1].M31[0].VolumeData.Height)
// 	}

// 	for _, newScan := range scans {
// 		// Find similar scan
// 		scan := FindScanElevationNumber(newScan, volume.Scans)
// 		if scan == nil {
// 			// If none add these
// 			scan = &newScan
// 			*volume.Scans = append(*volume.Scans, newScan)
// 		} else {
// 			scan.Gates = append(scan.Gates, newScan.Gates...)
// 			if newScan.EOE {
// 				scan.EOE = newScan.EOE
// 			}
// 		}
// 		if (scan.EOE || scan.EOV) && volume.VCP != 0 {
// 			fmt.Printf("%s on elevation %d completed\n", scan.ProductType, scan.ElevationNumber)
// 			AddSite(scan, *volume)
// 			if scan.EOV {
// 				fmt.Printf("Removing volume for %s\n", volume.ICAO)
// 				volume = nil
// 			}
// 		}
// 	}
// }
