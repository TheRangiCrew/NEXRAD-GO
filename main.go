package main

import (
	"os"

	"github.com/TheRangiCrew/NEXRAD-GO/server"
)

func main() {
	// filename := "./test/chunks/20240401-214657-002-I"
	// filename := "KLCH20240123_003612_V06"
	// filename := "TATL20240421_001129_V08"

	// file, err := os.Open(filename)
	// if err != nil {
	// 	panic(err)
	// }

	// tdwr.ParseTDWR(file)
	// l2Radar, err := nexrad.ParseNexrad(file)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(l2Radar.ElevationScans[1].M31[2].VolumeData.DataName[:]))

	// server.StartServer()

	files := []string{
		"20240401-214657-001-S",
		"20240401-214657-002-I",
		"20240401-214657-003-I",
		"20240401-214657-004-I",
		"20240401-214657-005-I",
		"20240401-214657-006-I",
		"20240401-214657-007-I",
		"20240401-214657-008-I",
		"20240401-214657-009-I",
		"20240401-214657-010-I",
	}
	for _, filename := range files {
		file, err := os.Open("./test/chunks/" + filename)
		if err != nil {
			panic(err)
		}

		chunkData, err := server.FilenameToChunkData(filename)
		if err != nil {
			panic(err)
		}
		server.ParseNewChunk(file, *chunkData)
	}
}
