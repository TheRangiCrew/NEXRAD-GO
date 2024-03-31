package l2

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"sort"

	"github.com/TheRangiCrew/nexrad-go/utils"
	geojson "github.com/paulmach/go.geojson"
)

type LDM struct {
	size int32
	data io.ReadSeeker
}

type MessageHeader struct {
	Size           uint16
	Channel        uint8
	MessageType    uint8
	SequenceNumber uint16
	JulianDate     uint16
	DayMS          uint32
	Segments       uint16
	SegmentNumber  uint16
}

type L2Radar struct {
	VolumeHeader   VolumeHeader
	ElevationScans map[int][]*Message31
}

func ParseL2Radar(file io.ReadSeeker) *L2Radar {
	// Make sure we are starting at the beginning of the file
	file.Seek(0, io.SeekStart)

	header := GetVolumeHeader(file)

	l2Radar := L2Radar{
		VolumeHeader:   *header,
		ElevationScans: make(map[int][]*Message31),
	}

	fmt.Printf("Tape: %s %s\nICAO: %s\n", l2Radar.VolumeHeader.Tape, l2Radar.VolumeHeader.Extension, l2Radar.VolumeHeader.ICAO)

	for {

		ldm := LDM{}
		if err := binary.Read(file, binary.BigEndian, &ldm.size); err != nil {
			if err != io.EOF {
				panic(err)
			}
			break
		}

		if ldm.size < 0 {
			ldm.size = -ldm.size
		}

		if IsCompressed(file) {
			ldm.data = Decompress(file, int(ldm.size))
		} else {
			ldm.data = file
		}
		for {
			ldm.data.Seek(CTMHeaderSize, io.SeekCurrent)

			messageHeader := MessageHeader{}
			if err := binary.Read(ldm.data, binary.BigEndian, &messageHeader); err != nil {
				if err != io.EOF {
					panic(err)
				}
				break
			}

			// Double the size since the value of size is in halfwords
			messageHeader.Size *= 2

			switch messageHeader.MessageType {
			// case 5:
			// 	ParseMessage5(ldm.data)
			case 31:
				m31 := ParseMessage31(ldm.data)

				l2Radar.ElevationScans[int(m31.Header.ElevationNumber)] = append(l2Radar.ElevationScans[int(m31.Header.ElevationNumber)], &m31)
			default:
				ldm.data.Seek(MessageBodySize, io.SeekCurrent)
			}
		}
	}

	return &l2Radar
}

func (l2 *L2Radar) ToJSON() {
	type ScanOutput struct {
		ProductType       string      `json:"productType"`
		ElevationAngle    float32     `json:"elevationAngle"`
		ElevationNumber   int         `json:"elevationNumber"`
		StartAzimuth      float32     `json:"startAngle"`
		AzimuthResolution float32     `json:"azimuthResolution"`
		StartRange        float32     `json:"startRange"`
		GateInterval      float32     `json:"gateInterval"`
		Lat               float32     `json:"lat"`
		Lon               float32     `json:"lon"`
		Gates             [][]float32 `json:"gates"`
	}

	type MomentBlocks struct {
		AzimuthAngle float32
		Gates        []float32
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

	elevations := map[int]*Elevation{}

	for k, e := range l2.ElevationScans {
		elevations[k] = &Elevation{
			Number:  k,
			Moments: map[string]*Moment{},
		}

		elevation := elevations[k]

		var angleSum float32 = 0.0
		n := 0

		for _, scan := range e {
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
								AzimuthAngle: scan.Header.AzimuthAngle,
								Gates:        m.Data,
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

	scans := []ScanOutput{}

	for _, e := range elevations {
		moment := e.Moments["VEL"]

		if moment == nil {
			continue
		}

		gates := [][]float32{}

		for _, m := range moment.Blocks {
			gates = append(gates, m.Gates)
		}

		scans = append(scans, ScanOutput{
			ProductType:       "VEL",
			ElevationNumber:   e.Number,
			ElevationAngle:    e.Angle,
			StartAzimuth:      moment.Blocks[0].AzimuthAngle,
			AzimuthResolution: e.AzimuthResolution,
			StartRange:        moment.StartRange,
			GateInterval:      moment.GateInterval,
			Lat:               e.Lat,
			Lon:               e.Lon,
			Gates:             gates,
		})

	}

	sort.Slice(scans, func(i, j int) bool {
		return scans[i].ElevationNumber < scans[j].ElevationNumber
	})

	s, _ := json.Marshal(scans[0])

	err := os.WriteFile("out.json", s, os.ModePerm.Perm())
	if err != nil {
		panic(err)
	}

}

func (l2 *L2Radar) ToTippecanoe() {

	type MomentBlocks struct {
		AzimuthAngle float32
		Gates        []float32
	}

	type Moment struct {
		StartRange   float32
		GateInterval float32
		Name         string
		Blocks       []MomentBlocks
	}

	type Elevation struct {
		Angle             float32
		AzimuthResolution uint8
		Lat               float32
		Lon               float32
		Number            int
		Moments           map[string]*Moment
	}

	elevations := map[int]*Elevation{}

	for k, e := range l2.ElevationScans {
		elevations[k] = &Elevation{
			Number:  k,
			Moments: map[string]*Moment{},
		}

		elevation := elevations[k]

		var angleSum float32 = 0.0
		n := 0

		for _, scan := range e {
			if elevation.Lat == 0 && elevation.Lon == 0 {
				elevation.Lat = scan.VolumeData.Lat
				elevation.Lon = scan.VolumeData.Long
			}
			if elevation.AzimuthResolution == 0 {
				elevation.AzimuthResolution = scan.Header.AzimuthResolution
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
								AzimuthAngle: scan.Header.AzimuthAngle,
								Gates:        m.Data,
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

	e := elevations[1]
	azimuthResolution := float64(e.AzimuthResolution) / 2.0
	m := e.Moments["REF"]
	startAngle := m.Blocks[0].AzimuthAngle
	// block := m.Blocks[0]

	polygons := geojson.NewFeatureCollection()
	// for _, m := range e.Moments {
	for j, block := range m.Blocks {

		angle := math.Mod(float64(startAngle)+(float64(j)*azimuthResolution), 360.0)

		for i, g := range block.Gates {

			if g >= 2 {
				distance := float64((m.StartRange + (m.GateInterval * float32(i))))
				startP1 := utils.FindEndPoint(float64(e.Lon), float64(e.Lat), angle-(azimuthResolution)/2, distance)
				startP2 := utils.FindEndPoint(float64(e.Lon), float64(e.Lat), angle+(azimuthResolution)/2, distance)

				distance = float64((m.StartRange + (m.GateInterval * float32(i+1))))
				endP1 := utils.FindEndPoint(float64(e.Lon), float64(e.Lat), angle-(azimuthResolution)/2, distance)
				endP2 := utils.FindEndPoint(float64(e.Lon), float64(e.Lat), angle+(azimuthResolution)/2, distance)

				gj := geojson.NewPolygonFeature([][][]float64{{startP1[:], startP2[:], endP2[:], endP1[:], startP1[:]}})

				gj.SetProperty("value", g)

				polygons.AddFeature(gj)
			}

		}

	}
	s, _ := polygons.MarshalJSON()

	err := os.WriteFile("out.json", s, os.ModePerm.Perm())
	if err != nil {
		panic(err)
	}
	// }

}
