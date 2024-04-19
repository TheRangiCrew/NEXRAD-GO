package level2

import (
	"encoding/binary"
	"io"
	"time"
)

type VolumeHeader struct {
	Tape       [9]byte
	Extension  [3]byte
	JulianDate uint32
	Time       uint32
	ICAO       [4]byte
}

func (v *VolumeHeader) Date() time.Time {
	return JulianDateToTime(v.JulianDate, v.Time)
}

func GetVolumeHeader(file io.ReadSeeker) *VolumeHeader {
	file.Seek(0, io.SeekCurrent)
	volumeHeader := VolumeHeader{}

	err := binary.Read(file, binary.BigEndian, &volumeHeader)
	if err != nil {
		panic(err)
	}

	return &volumeHeader
}
