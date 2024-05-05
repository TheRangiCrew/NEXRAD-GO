package tdwr

import (
	"encoding/binary"
	"io"
)

type ElevationCut struct {
	ElevationAngle        uint16
	ChannelConfig         uint8
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
	EdgeAngle2            uint16
	DopplePRF2            uint16
	DopplerPRFPulse2      uint16
	EBCAngle              uint16
	EdgeAngle3            uint16
	DopplePRF3            uint16
	DopplerPRFPulse3      uint16
	Reserved              uint16
}

type Message5Header struct {
	MessageSize         uint16
	PatternType         uint16
	PatterNumber        uint16
	NumberOfCuts        uint16
	VCPVersion          uint8
	ClutterMapGroup     uint8
	DopplerResolution   uint8
	PulseWidth          uint8
	_                   uint32
	VCPSequencing       uint16
	VCPSupplementalData uint16
	_                   uint16
}

type Message5 struct {
	Header          Message5Header
	ElevationAngles []ElevationCut
}

func ParseMessage5(file io.ReadSeeker) Message5 {
	header := Message5Header{}
	binary.Read(file, binary.BigEndian, &header)

	m5 := Message5{
		Header:          header,
		ElevationAngles: []ElevationCut{},
	}

	for i := 0; i < int(m5.Header.NumberOfCuts); i++ {
		elevationCut := ElevationCut{}
		if err := binary.Read(file, binary.BigEndian, &elevationCut); err != nil {
			if err != io.EOF {
				panic(err)
			}
			break
		}

		m5.ElevationAngles = append(m5.ElevationAngles, elevationCut)
	}

	return m5
}

type VCPSupplementalData struct {
}

func (message Message5) GetVCPBits() {
	// fmt.Println(message.Header.VCPSupplementalData&1 == 1)
	// fmt.Println(message.Header.VCPSupplementalData & 14)
	// fmt.Println(message.Header.VCPSupplementalData&16 != 0)
	// fmt.Println((uint8)(message.Header.VCPSupplementalData & 224))
	// fmt.Println("VCP")
	// fmt.Println(message.Header.VCPSupplementalData & 1792)
	// fmt.Println(message.Header.VCPSupplementalData & 2048)
	// fmt.Println(message.Header.VCPSupplementalData & 4096)
	// fmt.Println(message.Header.VCPSupplementalData & 57344)

}
