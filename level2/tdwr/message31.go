package tdwr

import (
	"encoding/binary"
	"fmt"
	"io"

	geojson "github.com/paulmach/go.geojson"
)

type VolumeData struct {
	DataBlockType       [1]byte
	DataName            [3]byte
	LRTUP               uint16
	VersionMajor        uint8
	VersionMinor        uint8
	Lat                 float32
	Long                float32
	Height              uint16
	FeedhornHeight      uint16
	CalibrationConstant float32
	HorizTXPower        float32
	VertTXPower         float32
	DiffReflectivity    float32
	DiffPhase           float32
	VCPNumber           uint16
	ProcessingStatus    uint16
}

type ElevationData struct {
	DataBlockType       [1]byte
	DataName            [3]byte
	LRTUP               uint16
	ATMOS               uint16
	CalibrationConstant float32
}

type RadialData struct {
	DataBlockType [1]byte
	DataName      [3]byte
	LRTUP         uint16
	Range         uint16
}

type Message31Header struct {
	ICAO              [4]byte
	CollectionTime    uint32
	CollectionDate    uint16
	AzimuthNumber     uint16
	AzimuthAngle      float32
	Compression       uint8
	Spare             uint8
	RadialLength      uint16
	AzimuthResolution uint8 // 1 = 0.5, 2 = 1.0
	RadialStatus      uint8
	ElevationNumber   uint8
	CutSectorNumber   uint8
	ElevationAngle    float32
	RadialBlanking    uint8
	AzimuthIndexing   uint8
	DataBlockCount    uint16
	// Data block pointers
}

type GenericMoment struct {
	DataBlockType       [1]byte
	MomentName          [3]byte
	Reserved            uint32
	NumberGates         uint16
	Range               uint16
	RangeSampleInterval uint16
	TOVER               uint16
	SNRThreshold        uint16
	ControlFlags        uint8
	DataWordSize        uint8
	Scale               float32
	Offset              float32
}

type Moment struct {
	GenericMoment
	Data []float32
}

type Message31 struct {
	Header        Message31Header
	VolumeData    VolumeData
	ElevationData ElevationData
	RadialData    RadialData
	MomentData    map[string]Moment
}

func (m31 *Message31) ToGEOJson() *geojson.FeatureCollection {
	// polygons := []geojson.Geometry{}

	keys := make([]string, 0, len(m31.MomentData))
	for k := range m31.MomentData {
		keys = append(keys, k)
	}

	for _, k := range keys {
		fmt.Println(k)
	}

	return nil
}

func ParseMessage31(file io.ReadSeeker) Message31 {

	startPos, _ := file.Seek(0, io.SeekCurrent)

	header := Message31Header{}
	if err := binary.Read(file, binary.BigEndian, &header); err != nil {
		panic(err)
	}

	curr, _ := file.Seek(0, io.SeekCurrent)
	fmt.Printf("After header: %d\n", curr)

	message31 := Message31{
		Header:     header,
		MomentData: make(map[string]Moment),
	}

	blockPointers := make([]uint32, header.DataBlockCount)
	if err := binary.Read(file, binary.BigEndian, blockPointers); err != nil {
		panic(err.Error())
	}
	fmt.Println(blockPointers)

	curr, _ = file.Seek(0, io.SeekCurrent)
	fmt.Printf("After pointers: %d\n", curr)

	for _, pointer := range blockPointers {
		if pointer == 0 {
			continue
		}
		file.Seek(startPos+int64(pointer)+1, io.SeekStart)

		n := make([]byte, 3)
		if err := binary.Read(file, binary.BigEndian, &n); err != nil {
			panic(err.Error())
		}

		file.Seek(-4, io.SeekCurrent)

		name := string(n)
		curr, _ := file.Seek(0, io.SeekCurrent)
		fmt.Printf("Before %s %d\n", name, curr)
		switch name {
		case "VOL":
			binary.Read(file, binary.BigEndian, &message31.VolumeData)
			fmt.Println(message31.VolumeData.LRTUP)
		case "ELV":
			binary.Read(file, binary.BigEndian, &message31.ElevationData)
			fmt.Println(message31.ElevationData.LRTUP)

		case "RAD":
			binary.Read(file, binary.BigEndian, &message31.RadialData)
			// fmt.Println(string(message31.RadialData.:]))

		case "REF":
			fallthrough
		case "VEL":
			fallthrough
		case "SW ":

			m := GenericMoment{}
			binary.Read(file, binary.BigEndian, &m)

			ldm := (m.NumberGates * uint16(m.DataWordSize)) / 8
			fmt.Println(m.DataWordSize)

			fmt.Println(m.NumberGates)
			data := make([]byte, ldm)
			binary.Read(file, binary.BigEndian, data)

			converted := []float32{}
			if m.Scale != 0 {
				for _, n := range data {
					converted = append(converted, (float32(n)-float32(m.Offset))/float32(m.Scale))
				}
			}

			d := Moment{
				GenericMoment: m,
				Data:          converted,
			}

			message31.MomentData[name] = d
		}
		curr, _ = file.Seek(0, io.SeekCurrent)
		fmt.Printf("After %s %d\n", name, curr)
	}

	fmt.Printf("Range %d\n", message31.RadialData.LRTUP)

	return message31
}
