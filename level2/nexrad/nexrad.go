package nexrad

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/TheRangiCrew/nexrad-go/level2"
)

type ElevationMessages struct {
	M5  []*Message5
	M31 []*level2.Message31
}

type Nexrad struct {
	IsNexradArchive bool
	ICAO            string
	VolumeHeader    level2.VolumeHeader
	ElevationScans  map[int]*ElevationMessages
}

func ParseNexrad(file io.ReadSeeker) *Nexrad {
	// Make sure we are starting at the beginning of the file
	file.Seek(0, io.SeekStart)

	header := level2.GetVolumeHeader(file)

	nexrad := Nexrad{
		VolumeHeader:    *header,
		IsNexradArchive: string(header.Tape[:]) == "AR2V0006.",
		ElevationScans:  make(map[int]*ElevationMessages),
	}

	if nexrad.IsNexradArchive {
		nexrad.ICAO = string(nexrad.VolumeHeader.ICAO[:])
	} else {
		file.Seek(0, io.SeekStart)
	}

	fmt.Println(string(header.Tape[:]))

	fmt.Println(string(header.ICAO[:]))

	for {

		ldm := level2.LDM{}
		if err := binary.Read(file, binary.BigEndian, &ldm.Size); err != nil {
			if err != io.EOF {
				panic(err)
			}
			break
		}

		if ldm.Size < 0 {
			ldm.Size = -ldm.Size
		}

		if level2.IsCompressed(file) {
			ldm.Data = level2.Decompress(file, int(ldm.Size))
		} else {
			ldm.Data = file
		}

		for {
			ldm.Data.Seek(level2.CTMHeaderSize, io.SeekCurrent)

			messageHeader := level2.MessageHeader{}
			if err := binary.Read(ldm.Data, binary.BigEndian, &messageHeader); err != nil {
				if err != io.EOF {
					panic(err)
				}
				break
			}

			fmt.Println(messageHeader.MessageType)

			// Double the size since the value of size is in halfwords
			messageHeader.Size *= 2

			switch messageHeader.MessageType {
			case 5:
				m5 := ParseMessage5(ldm.Data)
				fmt.Println(m5.Header.PatterNumber)
			case 31:
				m31 := level2.ParseMessage31(ldm.Data)
				if nexrad.ICAO == "" {
					nexrad.ICAO = string(m31.Header.ICAO[:])
				}
				if nexrad.ElevationScans[int(m31.Header.ElevationNumber)] == nil {
					nexrad.ElevationScans[int(m31.Header.ElevationNumber)] = &ElevationMessages{
						M31: []*level2.Message31{},
					}
				}
				nexrad.ElevationScans[int(m31.Header.ElevationNumber)].M31 = append(nexrad.ElevationScans[int(m31.Header.ElevationNumber)].M31, &m31)
			default:
				ldm.Data.Seek(level2.MessageBodySize, io.SeekCurrent)
			}
		}
	}

	return &nexrad
}
