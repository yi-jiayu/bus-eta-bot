package main

import (
	"encoding/json"
	"os"

	"github.com/pkg/errors"
)

type InMemoryBusStopRepository struct {
	busStops map[string]BusStop
}

func (r *InMemoryBusStopRepository) Get(ID string) *BusStop {
	busStop, ok := r.busStops[ID]
	if ok {
		return &busStop
	}
	return nil
}

func NewInMemoryBusStopRepository(busStops []BusStopJSON) *InMemoryBusStopRepository {
	busStopsMap := make(map[string]BusStop)
	for _, bs := range busStops {
		busStopsMap[bs.BusStopCode] = BusStop{
			BusStopID:   bs.BusStopCode,
			Road:        bs.RoadName,
			Description: bs.Description,
		}
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
