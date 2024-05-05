package main

import (
	"github.com/TheRangiCrew/NEXRAD-GO/server"
)

func main() {
	// filename := "./test/chunks/20240401-214657-008-I"
	// filename := "KLCH20240123_003612_V06"
	// filename := "TATL20240421_001129_V08"

	// file, err := os.Open(filename)
	// if err != nil {
	// 	panic(err)
	// }

	// tdwr.ParseTDWR(file)
	// nexrad.ParseNexrad(file)

	// chunkData, err := server.FilenameToChunkData(filename)
	// if err != nil {
	// 	panic(err)
	// }
	// server.ParseNewChunk(file, *chunkData)
	server.StartServer()
}
