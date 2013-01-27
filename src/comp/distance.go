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

	return EarthRadiusKm * Acos(Sin(lat1)*Sin(lat2)+Cos(lat1)*Cos(lat2)*Cos(lon2-lon1))
}
