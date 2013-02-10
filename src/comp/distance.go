package main

import (
	. "math"
)

const EarthRadiusKm = 6371.0

func Dist(lat1, lon1, lat2, lon2 float64) float64 {
	lat1 = lat1 * Pi / 180.0
	lon1 = lon1 * Pi / 180.0
	lat2 = lat2 * Pi / 180.0
	lon2 = lon2 * Pi / 180.0

	x := (lon2 - lon1) * Cos((lat1+lat2)/2)
	y := (lat2 - lat1)

	return Sqrt(x*x+y*y) * EarthRadiusKm
}
