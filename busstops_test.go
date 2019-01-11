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
	repo := NewInMemoryBusStopRepository(busStops, nil)

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
	repo := NewInMemoryBusStopRepository(busStops, nil)

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

func Test_replaceSynonyms(t *testing.T) {
	synonyms := map[string]string{
		"bukit": "bk",
		"park":  "pk",
	}
	tokens := []string{"bukit", "park"}
	actual := replaceSynonyms(synonyms, tokens)
	expected := []string{"bk", "pk"}
	assert.Equal(t, expected, actual)
}

func TestInMemoryBusStopRepository_Search(t *testing.T) {
	busStops := []BusStopJSON{
		{
			BusStopCode: "01019",
			RoadName:    "Victoria St",
			Description: "Bras Basah Cplx",
			Latitude:    1.29698951191332,
			Longitude:   103.85302201172507,
		},
		{
			BusStopCode: "00481",
			RoadName:    "Woodlands Rd",
			Description: "BT PANJANG TEMP BUS PK",
			Latitude:    1.383764,
			Longitude:   103.7583,
		},
		{
			BusStopCode: "01029",
			RoadName:    "Nth Bridge Rd",
			Description: "Cosmic Insurance Bldg",
			Latitude:    1.2966729849642,
			Longitude:   103.85441422464267,
		},
		{
			BusStopCode: "04159",
			RoadName:    "Victoria St",
			Description: "Aft Chijmes",
			Latitude:    1.29489484954945,
			Longitude:   103.85108138663934,
		},
		{
			BusStopCode: "93031",
			RoadName:    "Marine Parade Rd",
			Description: "Opp Victoria Sch",
			Latitude:    1.30983637164414,
			Longitude:   103.92937539318251,
		},
	}
	synonyms := map[string]string{
		"bukit": "bk",
		"park":  "pk",
	}
	repo := NewInMemoryBusStopRepository(busStops, nil)
	repo.synonyms = synonyms

	testCases := []struct {
		Name     string
		Query    string
		Limit    int
		Expected []BusStopJSON
	}{
		{
			Name:     "an empty string matches all bus stops",
			Query:    "",
			Expected: busStops,
		},
		{
			Name:  "should respect limit when query is empty",
			Query: "",
			Limit: 1,
			Expected: []BusStopJSON{
				{
					BusStopCode: "01019",
					RoadName:    "Victoria St",
					Description: "Bras Basah Cplx",
					Latitude:    1.29698951191332,
					Longitude:   103.85302201172507,
				},
			},
		},
		{
			Name:  "search by bus stop code",
			Query: "01019",
			Expected: []BusStopJSON{
				{
					BusStopCode: "01019",
					RoadName:    "Victoria St",
					Description: "Bras Basah Cplx",
					Latitude:    1.29698951191332,
					Longitude:   103.85302201172507,
				},
			},
		},
		{
			Name:  "search for word",
			Query: "bridge",
			Expected: []BusStopJSON{
				{
					BusStopCode: "01029",
					RoadName:    "Nth Bridge Rd",
					Description: "Cosmic Insurance Bldg",
					Latitude:    1.2966729849642,
					Longitude:   103.85441422464267,
				},
			},
		},
		{
			Name:  "matches in description should rank higher than matches in road name",
			Query: "victoria",
			Expected: []BusStopJSON{
				{
					BusStopCode: "93031",
					RoadName:    "Marine Parade Rd",
					Description: "Opp Victoria Sch",
					Latitude:    1.30983637164414,
					Longitude:   103.92937539318251,
				},
				{
					BusStopCode: "01019",
					RoadName:    "Victoria St",
					Description: "Bras Basah Cplx",
					Latitude:    1.29698951191332,
					Longitude:   103.85302201172507,
				},
				{
					BusStopCode: "04159",
					RoadName:    "Victoria St",
					Description: "Aft Chijmes",
					Latitude:    1.29489484954945,
					Longitude:   103.85108138663934,
				},
			},
		},
		{
			Name:  "matches in description should rank higher than matches in road name and should respect limit",
			Query: "victoria",
			Limit: 1,
			Expected: []BusStopJSON{
				{
					BusStopCode: "93031",
					RoadName:    "Marine Parade Rd",
					Description: "Opp Victoria Sch",
					Latitude:    1.30983637164414,
					Longitude:   103.92937539318251,
				},
			},
		},
		{
			Name:  "searches for words should match their abbreviations",
			Query: "bukit park",
			Expected: []BusStopJSON{
				{
					BusStopCode: "00481",
					RoadName:    "Woodlands Rd",
					Description: "BT PANJANG TEMP BUS PK",
					Latitude:    1.383764,
					Longitude:   103.7583,
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			actual := repo.Search(tc.Query, tc.Limit)
			assert.Equal(t, tc.Expected, actual)
		})
	}
}
