package nexrad

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/TheRangiCrew/NEXRAD-GO/level2"
)

type ElevationMessages struct {
	M31 []*Message31
}

type Nexrad struct {
	IsArchive      bool
	ICAO           string
	VolumeHeader   level2.VolumeHeader
	VCP            *Message5
	ElevationScans map[int]*ElevationMessages
}

func ParseNexrad(file io.ReadSeeker) (*Nexrad, error) {
	// Start at the beginning of the file
	file.Seek(0, io.SeekStart)

	// Parse the Volume Header
	header, err := level2.GetVolumeHeader(file)
	if err != nil {
		return nil, err
	}

	radar := Nexrad{
		VolumeHeader:   *header,
		IsArchive:      string(header.Tape[:]) == "AR2V0006.",
		ElevationScans: make(map[int]*ElevationMessages),
	}

	// If the file is an archive file then the ICAO is provided for us already
	if radar.IsArchive {
		radar.ICAO = string(radar.VolumeHeader.ICAO[:])
	} else {
		file.Seek(0, io.SeekStart)
	}

	//fmt.Println(string(header.ICAO[:]))

	for {

		// Create the LDM Record
		ldmRecord := level2.LDM{}
		if err := binary.Read(file, binary.BigEndian, &ldmRecord.Size); err != nil {
			if err != io.EOF {
				return nil, fmt.Errorf("reached EOF when reading LDM record")
			}
			break
		}

		if ldmRecord.Size < 0 {
			ldmRecord.Size = -ldmRecord.Size
		}

		// Decompress the LDM Record
		if level2.IsCompressed(file) {
			ldmRecord.Data = level2.Decompress(file, int(ldmRecord.Size))
		} else {
			ldmRecord.Data = file
		}

		for {
			ldmRecord.Data.Seek(level2.CTMHeaderSize, io.SeekCurrent)

			messageHeader := level2.MessageHeader{}
			if err := binary.Read(ldmRecord.Data, binary.BigEndian, &messageHeader); err != nil {
				if err != io.EOF {
					return nil, err
				}
				break
			}

			switch messageHeader.MessageType {
			case 5:
				radar.VCP, err = ParseMessage5(ldmRecord.Data)
				if err != nil {
					return nil, err
				}
				ldmRecord.Data.Seek(int64(level2.DefaultMessageSize-(messageHeader.Size*2)-level2.CTMHeaderSize), io.SeekCurrent)
			case 31:
				m31, err := ParseMessage31(ldmRecord.Data)
				if err != nil {
					return nil, err
				}
				if radar.ICAO == "" {
					radar.ICAO = string(m31.Header.ICAO[:])
				}
				if radar.ElevationScans[int(m31.Header.ElevationNumber)] == nil {
					radar.ElevationScans[int(m31.Header.ElevationNumber)] = &ElevationMessages{
						M31: []*Message31{},
					}
				}
				radar.ElevationScans[int(m31.Header.ElevationNumber)].M31 = append(radar.ElevationScans[int(m31.Header.ElevationNumber)].M31, m31)
			default:
				ldmRecord.Data.Seek(int64(level2.MessageBodySize), io.SeekCurrent)
			}
		}

	}

	return &radar, nil
}
