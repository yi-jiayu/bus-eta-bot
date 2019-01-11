package main

import "math"

const (
	EquatorialLatitude  = 110574.0
	EquatorialLongitude = 111320.0
)

// hsin and Distance implementation from https://gist.github.com/cdipaolo/d3f8db3848278b49db68

// haversin(θ) function
func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}

// Distance function returns the distance (in meters) between two points of
//     a given longitude and latitude relatively accurately (using a spherical
//     approximation of the Earth) through the Haversin Distance Formula for
//     great arc distance on a sphere with accuracy for small distances
//
// point coordinates are supplied in degrees and converted into rad. in the func
//
// distance returned is METERS!!!!!!
// http://en.wikipedia.org/wiki/Haversine_formula
func Distance(lat1, lon1, lat2, lon2 float64) float64 {
	// convert to radians
	// must cast radius as float to multiply later
	var la1, lo1, la2, lo2, r float64
	la1 = lat1 * math.Pi / 180
	lo1 = lon1 * math.Pi / 180
	la2 = lat2 * math.Pi / 180
	lo2 = lon2 * math.Pi / 180

	r = 6378100 // Earth radius in METERS

	// calculate
	h := hsin(la2-la1) + math.Cos(la1)*math.Cos(la2)*hsin(lo2-lo1)

	return 2 * r * math.Asin(math.Sqrt(h))
}

// EuclideanDistanceAtEquator returns the approximate distance between two points near the equator.
func EuclideanDistanceAtEquator(lat0, lon0, lat1, lon1 float64) float64 {
	dLat := EquatorialLatitude * (lat0 - lat1)
	dLon := EquatorialLongitude * (lon0 - lon1)
	return math.Sqrt(dLat*dLat + dLon*dLon)
}
