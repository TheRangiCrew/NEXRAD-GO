package level2

import (
	"io"
)

type LDM struct {
	Size int32
	Data io.ReadSeeker
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

// func (l2 *L2Radar) GetElevationNumbers() []int {
// 	keys := make([]int, len(l2.ElevationScans))

// 	i := 0
// 	for k := range l2.ElevationScans {
// 		keys[i] = k
// 		i++
// 	}

// 	return keys
// }

// func (l2 *L2Radar) PrintInfo() {
// 	fmt.Println((l2.GetElevationNumbers()))
// }
