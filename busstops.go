package main

import (
	"encoding/json"
	"os"
	"sort"

	"github.com/pkg/errors"
)

type NearbyBusStop struct {
	BusStopJSON
	Distance float64
}

type InMemoryBusStopRepository struct {
	busStops map[string]BusStopJSON
}

func (r *InMemoryBusStopRepository) Get(ID string) *BusStopJSON {
	busStop, ok := r.busStops[ID]
	if ok {
		return &busStop
	}
	return nil
}

// Nearby returns up to limit bus stops which are within a given radius from a point as well as their
// distance from that point.
func (r *InMemoryBusStopRepository) Nearby(lat, lon, radius float64, limit int) (nearby []NearbyBusStop) {
	for _, bs := range r.busStops {
		distance := EuclideanDistanceAtEquator(lat, lon, bs.Latitude, bs.Longitude)
		if distance <= radius {
			nearby = append(nearby, NearbyBusStop{
				BusStopJSON: bs,
				Distance:    distance,
			})
		}
	}
	sort.Slice(nearby, func(i, j int) bool {
		return nearby[i].Distance < nearby[j].Distance
	})
	if limit <= 0 || limit > len(nearby) {
		limit = len(nearby)
	}
	return nearby[:limit]

}

func NewInMemoryBusStopRepository(busStops []BusStopJSON) *InMemoryBusStopRepository {
	busStopsMap := make(map[string]BusStopJSON)
	for _, bs := range busStops {
		busStopsMap[bs.BusStopCode] = bs
	}
	return &InMemoryBusStopRepository{busStops: busStopsMap}
}

func NewInMemoryBusStopRepositoryFromFile(path string) (*InMemoryBusStopRepository, error) {
	busStopsJSONFile, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "error opening bus stops JSON file")
	}
	var busStopsJSON []BusStopJSON
	err = json.NewDecoder(busStopsJSONFile).Decode(&busStopsJSON)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding bus stops JSON file")
	}
	return NewInMemoryBusStopRepository(busStopsJSON), nil
}
