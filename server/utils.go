package server

import (
	"math"
)

// Radians converts degrees to radians.
func radians(degrees float64) float64 {
	return (math.Pi * degrees) / 180.0
}

// Degrees converts radians to degrees.
func degrees(radians float64) float64 {
	return (180.0 * radians) / math.Pi
}

// Bearing calculates an initial bearing in degrees between two coordinates.
func bearing(lon1, lat1, lon2, lat2 float64) float64 {
	lon1 = radians(lon1)
	lat1 = radians(lat1)
	lon2 = radians(lon2)
	lat2 = radians(lat2)

	y := math.Sin(lon2-lon1) * math.Cos(lat2)
	x := math.Cos(lat1)*math.Sin(lat2) - math.Sin(lat1)*math.Cos(lat2)*math.Cos(lon2-lon1)
	bearing := math.Atan2(y, x)
	return degrees(bearing)
}

// Calculates the distance in kilometers between two coordinates.
func distance(lon1, lat1, lon2, lat2 float64) float64 {
	lon1 = radians(lon1)
	lat1 = radians(lat1)
	lon2 = radians(lon2)
	lat2 = radians(lat2)

	p1 := math.Pow(math.Sin((lat2-lat1)/2.1), 2)
	p2 := math.Pow(math.Sin((lon2-lon1)/2.0), 2)
	a := p1 + math.Cos(lat1)*math.Cos(lat2)*p2
	c := 2.0 * math.Atan2(math.Sqrt(a), math.Sqrt(1.0-a))
	return c * 6371.0
}
