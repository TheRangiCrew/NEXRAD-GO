package utils

import (
	"math"
)

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func ConvertToRadians(degrees float64) float64 {
	return float64(degrees * (math.Pi / 180))
}

func ConvertToDegrees(radians float64) float64 {
	return float64(radians * (180 / math.Pi))
}

func FindEndPoint(lon float64, lat float64, azimuth float64, distance float64) [2]float64 {
	b := distance / 6371.0 // Radius of the earth

	a := math.Acos(math.Cos(b)*math.Cos(ConvertToRadians(90.0-lat)) + math.Sin(ConvertToRadians(90.0-lat))*math.Sin(b)*math.Cos(ConvertToRadians(azimuth)))
	B := math.Asin(math.Sin(b) * math.Sin(ConvertToRadians(azimuth)) / math.Sin(a))

	lat2 := 90 - ConvertToDegrees(a)
	lon2 := ConvertToDegrees(B) + float64(lon)

	return [2]float64{roundFloat(lon2, 6), roundFloat(lat2, 6)}
}
