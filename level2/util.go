package level2

import (
	"bytes"
	"compress/bzip2"
	"encoding/binary"
	"io"
	"time"
)

func IsCompressed(file io.ReadSeeker) bool {
	b := make([]byte, 2)
	_, err := file.Read(b)
	if err != nil {
		panic(err)
	}
	file.Seek(-2, io.SeekCurrent)
	return string(b) == "BZ"
}

func Decompress(file io.ReadSeeker, size int) *bytes.Reader {
	compressedData := make([]byte, size)
	binary.Read(file, binary.BigEndian, &compressedData)
	bz2Reader := bzip2.NewReader(bytes.NewReader(compressedData))
	extractedData := bytes.NewBuffer([]byte{})
	io.Copy(extractedData, bz2Reader)
	return bytes.NewReader(extractedData.Bytes())
}

func JulianDateToTime(d uint32, t uint32) time.Time {
	return time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC).
		Add(time.Duration(d) * time.Hour * 24).
		Add(time.Duration(t) * time.Millisecond)
}

func IsNexradArchive(file io.ReadSeeker) (bool, error) {
	file.Seek(0, io.SeekStart)

	header, err := GetVolumeHeader(file)
	if err != nil {
		return false, err
	}

	file.Seek(0, io.SeekStart)

	return string(header.Tape[:]) == "AR2V0006.", nil
}

func IsTDWRArchive(file io.ReadSeeker) (bool, error) {
	file.Seek(0, io.SeekStart)

	header, err := GetVolumeHeader(file)
	if err != nil {
		return false, err
	}

	file.Seek(0, io.SeekStart)

	return string(header.Tape[:]) == "AR2V0008.", nil
}
