package main

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestInMemoryBusStopRepository(t *testing.T) {
	busStops := []BusStopJSON{
		{
			BusStopCode: "00481",
			RoadName:    "Woodlands Rd",
			Description: "BT PANJANG TEMP BUS PK",
		},
		{
			BusStopCode: "01012",
			RoadName:    "Victoria St",
			Description: "Hotel Grand Pacific",
		},
	}
	repo := NewInMemoryBusStopRepository(busStops)

	testCases := []struct {
		ID       string
		Expected *BusStop
	}{
		{
			ID: "00481",
			Expected: &BusStop{
				BusStopID:   "00481",
				Road:        "Woodlands Rd",
				Description: "BT PANJANG TEMP BUS PK"},
		},
		{
			ID: "01012",
			Expected: &BusStop{
				BusStopID:   "01012",
				Road:        "Victoria St",
				Description: "Hotel Grand Pacific",
			},
		},
		{
			ID:       "",
			Expected: nil,
		},
	}
	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			actual := repo.Get(tc.ID)
			assert.Equal(t, tc.Expected, actual)
		})
	}

}
