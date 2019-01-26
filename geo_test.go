package busetabot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSquaredEuclideanDistanceAtEquator(t *testing.T) {
	testCases := []struct {
		Lat0, Lon0, Lat1, Lon1 float64
		Expected               float64
	}{
		{
			Lat0:     1.383764,
			Lon0:     103.7583,
			Lat1:     1.29684825487647,
			Lon1:     103.85253591654006,
			Expected: 2.024e8,
		},
	}
	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			actual := SquaredEuclideanDistanceAtEquator(tc.Lat0, tc.Lon0, tc.Lat1, tc.Lon1)
			assert.InEpsilon(t, tc.Expected, actual, 0.01)
		})
	}
}
