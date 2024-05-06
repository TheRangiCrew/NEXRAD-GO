package server

import (
	"errors"
	"fmt"
	"strconv"
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
			PushScan(currentScan, volume.InitTime)
			fmt.Printf("Removing scan for %s\n", l2Radar.ICAO)
			_, err := RemoveScan(scanIndex, scans)
			if err != nil {
				return err
			}
		}

		fmt.Println(len(*scans))
	}

	return nil
}
