package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInMemoryBusStopRepository_Get(t *testing.T) {
	busStops := []BusStopJSON{
		{
			BusStopCode: "00481",
			RoadName:    "Woodlands Rd",
			Description: "BT PANJANG TEMP BUS PK",
			Latitude:    1.383764,
			Longitude:   103.7583,
		},
		{
			BusStopCode: "01012",
			RoadName:    "Victoria St",
			Description: "Hotel Grand Pacific",
			Latitude:    1.29684825487647,
			Longitude:   103.85253591654006,
		},
	}
	repo := NewInMemoryBusStopRepository(busStops)

	testCases := []struct {
		ID       string
		Expected *BusStopJSON
	}{
		{
			ID: "00481",
			Expected: &BusStopJSON{
				BusStopCode: "00481",
				RoadName:    "Woodlands Rd",
				Description: "BT PANJANG TEMP BUS PK",
				Latitude:    1.383764,
				Longitude:   103.7583,
			},
		},
		{
			ID: "01012",
			Expected: &BusStopJSON{
				BusStopCode: "01012",
				RoadName:    "Victoria St",
				Description: "Hotel Grand Pacific",
				Latitude:    1.29684825487647,
				Longitude:   103.85253591654006,
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

func TestInMemoryBusStopRepository_Nearby(t *testing.T) {
	busStops := []BusStopJSON{
		{
			BusStopCode: "01012",
			RoadName:    "Victoria St",
			Description: "Hotel Grand Pacific",
			Latitude:    1.29684825487647,
			Longitude:   103.85253591654006,
		},
		{
			BusStopCode: "00481",
			RoadName:    "Woodlands Rd",
			Description: "BT PANJANG TEMP BUS PK",
			Latitude:    1.383764,
			Longitude:   103.7583,
		},
	}
	repo := NewInMemoryBusStopRepository(busStops)

	testCases := []struct {
		Name     string
		Lat      float64
		Lon      float64
		Radius   float64
		Limit    int
		Expected []NearbyBusStop
	}{
		{
			Name:   "one nearby bus stop",
			Lat:    1.3837,
			Lon:    103.75,
			Radius: 10000,
			Limit:  0,
			Expected: []NearbyBusStop{
				{
					BusStopJSON: BusStopJSON{
						BusStopCode: "00481",
						RoadName:    "Woodlands Rd",
						Description: "BT PANJANG TEMP BUS PK",
						Latitude:    1.383764,
						Longitude:   103.7583,
					},
					Distance: 923.9831005649131},
			},
		},
		{
			Name:   "two nearby bus stops, sorted by distance",
			Lat:    1.3422,
			Lon:    103.8023,
			Radius: 10000,
			Limit:  0,
			Expected: []NearbyBusStop{
				{
					BusStopJSON: BusStopJSON{
						BusStopCode: "00481",
						RoadName:    "Woodlands Rd",
						Description: "BT PANJANG TEMP BUS PK",
						Latitude:    1.383764,
						Longitude:   103.7583,
					},
					Distance: 6716.655692096068,
				},
				{
					BusStopJSON: BusStopJSON{
						BusStopCode: "01012",
						RoadName:    "Victoria St",
						Description: "Hotel Grand Pacific",
						Latitude:    1.29684825487647,
						Longitude:   103.85253591654006},
					Distance: 7511.381516450915,
				},
			},
		},
		{
			Name:   "two nearby bus stops limited to one closest one",
			Lat:    1.3422,
			Lon:    103.8023,
			Radius: 10000,
			Limit:  1,
			Expected: []NearbyBusStop{
				{
					BusStopJSON: BusStopJSON{
						BusStopCode: "00481",
						RoadName:    "Woodlands Rd",
						Description: "BT PANJANG TEMP BUS PK",
						Latitude:    1.383764,
						Longitude:   103.7583,
					},
					Distance: 6716.655692096068,
				},
			},
		},
		{
			Name:   "limit greater than number of nearby bus stops",
			Lat:    1.3422,
			Lon:    103.8023,
			Radius: 10000,
			Limit:  3,
			Expected: []NearbyBusStop{
				{
					BusStopJSON: BusStopJSON{
						BusStopCode: "00481",
						RoadName:    "Woodlands Rd",
						Description: "BT PANJANG TEMP BUS PK",
						Latitude:    1.383764,
						Longitude:   103.7583,
					},
					Distance: 6716.655692096068,
				},
				{
					BusStopJSON: BusStopJSON{
						BusStopCode: "01012",
						RoadName:    "Victoria St",
						Description: "Hotel Grand Pacific",
						Latitude:    1.29684825487647,
						Longitude:   103.85253591654006},
					Distance: 7511.381516450915,
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			actual := repo.Nearby(tc.Lat, tc.Lon, tc.Radius, tc.Limit)
			assert.Equal(t, tc.Expected, actual)
		})
	}
}
