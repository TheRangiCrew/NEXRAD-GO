package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type RandomAccessFile struct {
	offset int
	file   []byte
	reader io.Reader
}

func CreateRAF(filename string) (*RandomAccessFile, error) {

	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return &RandomAccessFile{
		offset: 0,
		file:   file,
		reader: bytes.NewReader(file),
	}, nil
}

// Get the file's length in bytes
func (raf *RandomAccessFile) Len() int {
	return len(raf.file)
}

func (raf *RandomAccessFile) Position() int {
	return raf.offset
}

func (raf *RandomAccessFile) Seek(position int) error {
	if position > raf.Len() {
		return fmt.Errorf("seeking position %d is out of bounds on length %d", position, raf.Len())
	}
	raf.offset = position
	return nil
}

func (raf *RandomAccessFile) ReadCode() int8 {
	value := int8(raf.file[raf.offset])
	raf.offset++
	return value
}

func (raf *RandomAccessFile) Read2Code() int16 {
	start := raf.offset
	raf.offset += 2
	return int16(binary.BigEndian.Uint16(raf.file[start:raf.offset]))
}

func (raf *RandomAccessFile) ReadByte() uint8 {
	value := uint8(raf.file[raf.offset])
	raf.offset++
	return value
}

func (raf *RandomAccessFile) ReadShort() uint16 {
	start := raf.offset
	raf.offset += 2
	return binary.BigEndian.Uint16(raf.file[start:raf.offset])
}

func (raf *RandomAccessFile) ReadSignedShort() int16 {
	start := raf.offset
	raf.offset += 2
	return int16(binary.BigEndian.Uint16(raf.file[start:raf.offset]))
}

func (raf *RandomAccessFile) ReadInt() uint32 {
	start := raf.offset
	raf.offset += 4
	return binary.BigEndian.Uint32(raf.file[start:raf.offset])
}

func (raf *RandomAccessFile) ReadSignedInt() int32 {
	start := raf.offset
	raf.offset += 4
	return int32(binary.BigEndian.Uint32(raf.file[start:raf.offset]))
}

func (raf *RandomAccessFile) ReadFloat() float32 {
	start := raf.offset
	raf.offset += 4
	return float32(binary.BigEndian.Uint32(raf.file[start:raf.offset]))
}

func (raf *RandomAccessFile) ReadDouble() float64 {
	start := raf.offset
	raf.offset += 8
	return float64(binary.BigEndian.Uint64(raf.file[start:raf.offset]))
}

func (raf *RandomAccessFile) ReadString(len int) string {
	if len < 1 {
		panic(fmt.Errorf("cannot read string of length %d", len))
	}
	start := raf.offset
	raf.offset += len
	return string(raf.file[start:raf.offset])
}

func (raf *RandomAccessFile) Read(len int) ([]byte, error) {
	if len < 1 {
		return nil, fmt.Errorf("cannot read bytes of length %d", len)
	}
	start := raf.offset
	raf.offset += len
	return raf.file[start:raf.offset], nil
}

func (raf *RandomAccessFile) Skip(len int) {
	raf.offset += len
}
