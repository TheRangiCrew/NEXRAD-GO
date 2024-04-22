package tdwr

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/TheRangiCrew/NEXRAD-GO/level2"
)

type ElevationMessages struct {
	M5  []*Message5
	M31 []*Message31
}

type TDWR struct {
	IsArchive      bool
	ICAO           string
	VolumeHeader   level2.VolumeHeader
	VCP            Message5
	ElevationScans map[int]*ElevationMessages
}

func ParseTDWR(file io.ReadSeeker) *TDWR {
	// Start at the beginning of the file
	file.Seek(0, io.SeekStart)

	// Parse the Volume Header
	header := level2.GetVolumeHeader(file)

	tdwr := TDWR{
		VolumeHeader:   *header,
		IsArchive:      string(header.Tape[:]) == "AR2V0008.",
		ElevationScans: make(map[int]*ElevationMessages),
	}

	// If the file is an archive file then the ICAO is provided for us already
	if tdwr.IsArchive {
		tdwr.ICAO = string(tdwr.VolumeHeader.ICAO[:])
	} else {
		file.Seek(0, io.SeekStart)
	}

	fmt.Println(string(header.ICAO[:]))

	i := 0

	for {

		// Create the LDM Record
		ldmRecord := level2.LDM{}
		if err := binary.Read(file, binary.BigEndian, &ldmRecord.Size); err != nil {
			if err != io.EOF {
				panic(err)
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

		curr, _ := ldmRecord.Data.Seek(0, io.SeekStart)
		fmt.Printf("\nStarting position: %d\n", curr)

		for {
			curr, _ := ldmRecord.Data.Seek(level2.CTMHeaderSize, io.SeekCurrent)
			fmt.Printf("Current position: %d\n", curr)

			messageHeader := level2.MessageHeader{}
			if err := binary.Read(ldmRecord.Data, binary.BigEndian, &messageHeader); err != nil {
				if err != io.EOF {
					panic(err)
				}
				break
			}

			curr, _ = ldmRecord.Data.Seek(0, io.SeekCurrent)
			fmt.Printf("After header: %d\n", curr)

			fmt.Printf("Message: %d\n", messageHeader.MessageType)

			switch messageHeader.MessageType {
			case 5:
				tdwr.VCP = ParseMessage5(ldmRecord.Data)
				ldmRecord.Data.Seek(int64(level2.DefaultMessageSize-(messageHeader.Size*2)-level2.CTMHeaderSize), io.SeekCurrent)
			case 31:
				fmt.Printf("Message size: %d\n", messageHeader.Size*2)
				ParseMessage31(ldmRecord.Data)
				curr, _ = ldmRecord.Data.Seek(0, io.SeekCurrent)
				fmt.Printf("After message: %d\n", curr)
				// break
			default:
				ldmRecord.Data.Seek(int64(level2.MessageBodySize), io.SeekCurrent)
			}
		}

		i++

		// if i == 2 {
		// 	break
		// }

	}

	return &tdwr
}
