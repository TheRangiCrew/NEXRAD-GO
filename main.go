package main

import (
	"os"

	"github.com/TheRangiCrew/nexrad-go/l2"
)

func main() {
	filename := "KVAX20240316_001243_V06"

	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	radar := l2.ParseL2Radar(file)

	radar.ToJSON()

}
