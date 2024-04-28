package server

import (
	"fmt"
	"sort"

	nexrad "github.com/TheRangiCrew/NEXRAD-GO/level2/nexrad"
)

type Scan struct {
	ICAO               string  `json:"icao"`
	ProductType        string  `json:"productType"`
	ElevationAngle     float32 `json:"elevationAngle"`
	ElevationNumber    int     `json:"elevationNumber"`
	StartAzimuth       float32 `json:"startAngle"`
	StartAzimuthNumber int
	AzimuthResolution  float32     `json:"azimuthResolution"`
	StartRange         float32     `json:"startRange"`
	GateInterval       float32     `json:"gateInterval"`
	Lat                float32     `json:"lat"`
	Lon                float32     `json:"lon"`
	Gates              [][]float32 `json:"gates"`
	Complete           bool
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

func NexradToScans(l2Radar *nexrad.Nexrad) []*Scan {

	elevations := map[int]*Elevation{}

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

			fmt.Println(scan.Header.RadialStatus)

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
		}

		elevation.Angle = angleSum / float32(n)
	}

	scans := []*Scan{}

	for _, e := range elevations {
		for k, moment := range e.Moments {

			gates := [][]float32{}

			for _, m := range moment.Blocks {
				gates = append(gates, m.Gates)
			}

			scans = append(scans, &Scan{
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
				Gates:              gates,
			})
		}

	}

	sort.Slice(scans, func(i, j int) bool {
		return scans[i].ElevationNumber < scans[j].ElevationNumber
	})

	return scans

	// for _, s := range scans {
	// 	j, _ := json.Marshal(s)

	// 	path := fmt.Sprintf("%s/%s", s.ICAO, strconv.Itoa(s.ElevationNumber))
	// 	filename := fmt.Sprintf("/%s.json", s.ProductType)

	// 	_, err := os.Stat(path + filename)
	// 	if os.IsNotExist(err) {
	// 		os.MkdirAll(path, os.ModePerm.Perm()) // Create your file
	// 	}

	// 	err = os.WriteFile(path+filename, j, os.ModePerm.Perm())
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }

}

func FindScanElevationNumber(scan *Scan, scans []*Scan) *Scan {
	for _, s := range scans {
		if s.ElevationNumber == scan.ElevationNumber && s.ProductType == scan.ProductType {
			return s
		}
	}

	return nil
}

func FindScanElevationAngle(scan *Scan, scans []*Scan) {}
