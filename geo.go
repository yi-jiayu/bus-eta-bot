package busetabot

const (
	EquatorialLatitude  = 110574.0
	EquatorialLongitude = 111320.0
)

// EuclideanDistanceAtEquator returns the approximate squared distance between two points near the equator.
func SquaredEuclideanDistanceAtEquator(lat0, lon0, lat1, lon1 float64) float64 {
	dLat := EquatorialLatitude * (lat0 - lat1)
	dLon := EquatorialLongitude * (lon0 - lon1)
	return dLat*dLat + dLon*dLon
}
