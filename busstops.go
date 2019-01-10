package main

import (
	"encoding/json"
	"os"

	"github.com/pkg/errors"
)

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
