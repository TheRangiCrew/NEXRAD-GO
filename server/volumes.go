package main

import (
	"strconv"
	"time"
)

type Volume struct {
	ID                    string    `json:"id,omitempty"`
	InitTime              time.Time `json:"init_time"`
	FinishTime            time.Time `json:"finish_time,omitempty"`
	VCP                   int       `json:"vcp"`
	CurrentElevation      int       `json:"current_elevation_number"`
	CurrentElevationAngle int       `json:"current_elevation_angle"`
	//Scans                 *[]Scan
}

func GetVolumeID(t time.Time, icao string) string {
	year := strconv.Itoa(t.Year())
	month := PadZero(strconv.Itoa(int(t.Month())), 2)
	day := PadZero(strconv.Itoa(t.Day()), 2)

	hour := PadZero(strconv.Itoa(t.Hour()), 2)
	minute := PadZero(strconv.Itoa(t.Minute()), 2)
	second := PadZero(strconv.Itoa(t.Second()), 2)

	return year + month + day + hour + minute + second + icao
}
