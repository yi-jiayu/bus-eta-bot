package main

import (
	"errors"
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/search"
)

const (
	busStopKind = "BusStop"
)

var (
	errNotFound = errors.New("not found")
)

// BusStop is a bus stop as represented inside app engine datastore and search.
type BusStop struct {
	BusStopID   string
	Road        string
	Description string
	Location    appengine.GeoPoint
}

// BusStopJSON is a bus stop deserialised from JSON
type BusStopJSON struct {
	BusStopCode string
	RoadName    string
	Description string
	Latitude    float64
	Longitude   float64
}

// DistanceFrom returns the distance between a bus stop and a reference coordinate.
func (b BusStop) DistanceFrom(lat, lon float64) float64 {
	return Distance(b.Location.Lat, b.Location.Lng, lat, lon)
}

// GetBusStop looks up a bus stop by id from the datastore
func GetBusStop(ctx context.Context, id string) (BusStop, error) {
	var busStop BusStop
	key := datastore.NewKey(ctx, busStopKind, id, 0, nil)
	err := datastore.Get(ctx, key, &busStop)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return BusStop{}, errNotFound
		}

		return BusStop{}, err
	}

	return busStop, nil
}

// PutBusStopsDatastore inserts a list of bus stops into datastore
func PutBusStopsDatastore(ctx context.Context, busStops []BusStopJSON) (string, error) {
	keys := make([]*datastore.Key, 0)
	entities := make([]BusStop, 0)
	var last BusStopJSON
	for _, busStop := range busStops {
		keys = append(keys, datastore.NewKey(ctx, busStopKind, busStop.BusStopCode, 0, nil))
		entities = append(entities, BusStop{
			BusStopID:   busStop.BusStopCode,
			Road:        busStop.RoadName,
			Description: busStop.Description,
			Location: appengine.GeoPoint{
				Lat: busStop.Latitude,
				Lng: busStop.Longitude,
			},
		})

		last = busStop

		if len(entities) == 50 {
			_, err := datastore.PutMulti(ctx, keys, entities)
			if err != nil {
				return last.BusStopCode, err
			}

			keys = make([]*datastore.Key, 0)
			entities = make([]BusStop, 0)
		}
	}

	_, err := datastore.PutMulti(ctx, keys, entities)
	if err != nil {
		return last.BusStopCode, err
	}

	return last.BusStopCode, nil
}

// PutBusStopsSearch inserts a list of bus stops into the search index
func PutBusStopsSearch(ctx context.Context, busStops []BusStopJSON) (string, error) {
	index, err := search.Open("BusStops")
	if err != nil {
		return "", err
	}

	put := 0
	for _, busStop := range busStops {
		document := BusStop{
			BusStopID:   busStop.BusStopCode,
			Road:        busStop.RoadName,
			Description: busStop.Description,
			Location: appengine.GeoPoint{
				Lat: busStop.Latitude,
				Lng: busStop.Longitude,
			},
		}

		_, err := index.Put(ctx, document.BusStopID, &document)
		if err != nil {
			return document.BusStopID, err
		}

		put++
	}

	return fmt.Sprintf("%d", put), nil
}

// GetNearbyBusStops returns nearby bus stops to a specified location
func GetNearbyBusStops(ctx context.Context, lat, lng float64, radius, limit int) ([]BusStop, error) {
	index, err := search.Open("BusStops")
	if err != nil {
		return nil, err
	}

	var busStops []BusStop

	opts := &search.SearchOptions{
		Limit: limit,
		Sort: &search.SortOptions{
			Expressions: []search.SortExpression{
				{
					Expr:    fmt.Sprintf("distance(Location, geopoint(%f, %f))", lat, lng),
					Reverse: true,
				},
			},
		},
	}

	for t := index.Search(ctx, fmt.Sprintf("distance(Location, geopoint(%f, %f)) < %d", lat, lng, radius), opts); ; {
		var busStop BusStop
		_, err := t.Next(&busStop)
		if err != nil {
			if err == search.Done {
				return busStops, nil
			}

			return busStops, err
		}

		busStops = append(busStops, busStop)
	}
}

// SearchBusStops returns bus stops containing query
func SearchBusStops(ctx context.Context, query string, offset int) ([]BusStop, error) {
	index, err := search.Open("BusStops")
	if err != nil {
		return nil, err
	}

	var busStops []BusStop
	options := search.SearchOptions{
		Limit:  50,
		Offset: offset,
	}
	for t := index.Search(ctx, query, &options); ; {
		var bs BusStop
		_, err := t.Next(&bs)
		if err != nil {
			if err == search.Done {
				return busStops, nil
			}

			return busStops, err
		}

		busStops = append(busStops, bs)
	}
}
