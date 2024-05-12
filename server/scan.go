package main

import (
	"errors"
	"sort"
	"sync"
	"time"

	nexrad "github.com/TheRangiCrew/NEXRAD-GO/level2/nexrad"
)

type Scan struct {
	ICAO               string       `json:"icao"`
	ProductType        string       `json:"productType"`
	ElevationAngle     float32      `json:"elevationAngle"`
	ElevationNumber    int          `json:"elevationNumber"`
	StartAzimuth       float32      `json:"startAngle"`
	StartAzimuthNumber int          `json:"-"`
	AzimuthResolution  float32      `json:"azimuthResolution"`
	StartRange         float32      `json:"startRange"`
	GateInterval       float32      `json:"gateInterval"`
	Lat                float32      `json:"lat"`
	Lon                float32      `json:"lon"`
	Gates              *[][]float32 `json:"gates"`
	InitTime           time.Time    `json:"init_time"`
	EOE                bool         `json:"-"` // End of elevation
	EOV                bool         `json:"-"` // EOV
}

type MomentBlocks struct {
	AzimuthAngle  float32
	AzimuthNumber int
	Gates         []float32
}

type Moment struct {
	StartRange   float32
	GateInterval float32
	Name         string
	Blocks       []MomentBlocks
}

type Elevation struct {
	Angle             float32
	AzimuthResolution float32
	Lat               float32
	Lon               float32
	Number            int
	Moments           map[string]*Moment
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

func RemoveScan(index int, scans *[]Scan) (Scan, error) {
	if index >= len(*scans) {
		return Scan{}, errors.New("index out of bounds")
	}

	removedScan := (*scans)[index]

	*scans = append((*scans)[:index], (*scans)[index+1:]...)

	return removedScan, nil
}

func NexradToScans(l2Radar *nexrad.Nexrad) []Scan {

	elevations := map[int]*Elevation{}

	eoe := false
	eov := false

	for k, e := range l2Radar.ElevationScans {
		elevations[k] = &Elevation{
			Number:  k,
			Moments: map[string]*Moment{},
		}

		elevation := elevations[k]

		var angleSum float32 = 0.0
		n := 0

		for _, scan := range e.M31 {
			if elevation.Lat == 0 && elevation.Lon == 0 {
				elevation.Lat = scan.VolumeData.Lat
				elevation.Lon = scan.VolumeData.Long
			}
			if elevation.AzimuthResolution == 0 {
				elevation.AzimuthResolution = float32(scan.Header.AzimuthResolution) / 2.0
			}

			angleSum += scan.Header.ElevationAngle
			n++
			for key, m := range scan.MomentData {
				moment := elevation.Moments[string(key)]

				if moment == nil {
					elevation.Moments[string(key)] = &Moment{
						StartRange:   float32(m.Range) / 1000.0,
						GateInterval: float32(m.RangeSampleInterval) / 1000.0,
						Name:         string(key),
						Blocks: []MomentBlocks{
							{
								AzimuthAngle:  scan.Header.AzimuthAngle,
								AzimuthNumber: int(scan.Header.AzimuthNumber),
								Gates:         m.Data,
							},
						},
					}
				} else {
					moment.Blocks = append(moment.Blocks, MomentBlocks{
						AzimuthAngle: scan.Header.AzimuthAngle,
						Gates:        m.Data,
					})
				}
			}

			if scan.Header.RadialStatus == 2 {
				eoe = true
			}
			if scan.Header.RadialStatus == 4 {
				eoe = true
				eov = true
			}
		}

		elevation.Angle = angleSum / float32(n)
	}

	scans := []Scan{}

	for _, e := range elevations {
		for k, moment := range e.Moments {

			gates := [][]float32{}

			for _, m := range moment.Blocks {
				tempGates := []float32{}

				mask := 0
				for _, g := range m.Gates {
					if len(tempGates) > 0 && g == tempGates[len(tempGates)-1] {
						mask++
					} else {
						if mask > 0 {
							tempGates = append(tempGates, float32(-1000-mask))
							mask = 0
						} else {
							tempGates = append(tempGates, g)
						}
					}
				}

				gates = append(gates, tempGates)
			}

			scans = append(scans, Scan{
				ICAO:               l2Radar.ICAO,
				ProductType:        k,
				ElevationNumber:    e.Number,
				ElevationAngle:     e.Angle,
				StartAzimuth:       moment.Blocks[0].AzimuthAngle,
				StartAzimuthNumber: moment.Blocks[0].AzimuthNumber,
				AzimuthResolution:  e.AzimuthResolution,
				StartRange:         moment.StartRange,
				GateInterval:       moment.GateInterval,
				Lat:                e.Lat,
				Lon:                e.Lon,
				Gates:              &gates,
				InitTime:           time.Now(),
				EOE:                eoe,
				EOV:                eov,
			})
		}

	}

	sort.Slice(scans, func(i, j int) bool {
		return scans[i].ElevationNumber < scans[j].ElevationNumber
	})

	return scans
}

/*
Finds the given scan in the slice of the scans. Returns the index. If the scan cannot be found, index is -1
*/
func FindScanIndex(scan Scan, scans *[]Scan) int {
	for i, s := range *scans {
		if s.ElevationNumber == scan.ElevationNumber && s.ProductType == scan.ProductType && s.ICAO == scan.ICAO {
			return i
		}
	}

	return -1
}

func FindScanElevationAngle(scan *Scan, scans []*Scan) *Scan {
	min := scan.ElevationAngle - 0.1
	max := scan.ElevationAngle + 0.1
	for _, s := range scans {
		if s.ElevationAngle >= min && s.ElevationAngle <= max {
			return s
		}
	}

	return nil
}
