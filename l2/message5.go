package l2

import (
	"encoding/binary"
	"io"
)

type ElevationAngle struct {
	Angle                 uint16
	Channel               uint8
	Waveform              uint8
	SuperResControl       uint8
	SurveillanceNumber    uint8
	SurveillanceCount     uint16
	AzimuthRate           uint16
	ReflectivityThreshold uint16
	VelocityThreshold     uint16
	SWThreshold           uint16
	DiffRefThreshold      uint16
	DiffPhaseThreshold    uint16
	CoorCoefThreshold     uint16
	EdgeAngle             uint16
	DopplePRF             uint16
	DopplerPRFPulse       uint16
	SupplementalData      uint16
	_                     uint16
	_                     uint16
	_                     uint16
	EBCAngle              uint16
	_                     uint16
	_                     uint16
	_                     uint16
	_                     uint16
}

type Message5Header struct {
	Size                uint16
	PatternType         uint16
	PatterNumber        uint16
	NumberOfCuts        uint16
	VCPVersion          uint8
	ClutterMapGroup     uint8
	DopplerResolution   uint8
	PulseWidth          uint8
	_                   uint32
	Sequencing          uint16
	VCPSupplementalData uint16
	_                   uint16
}

type Message5 struct {
	Header          Message5Header
	ElevationAngles []ElevationAngle
}

func ParseMessage5(file io.ReadSeeker) Message5 {
	header := Message5Header{}
	binary.Read(file, binary.BigEndian, &header)

	m5 := Message5{
		Header:          header,
		ElevationAngles: []ElevationAngle{},
	}

	for i := 0; i < int(m5.Header.NumberOfCuts); i++ {
		elevationAngle := ElevationAngle{}
		if err := binary.Read(file, binary.BigEndian, &elevationAngle); err != nil {
			if err != io.EOF {
				panic(err)
			}
			break
		}

		m5.ElevationAngles = append(m5.ElevationAngles, elevationAngle)
	}

	return m5
}
