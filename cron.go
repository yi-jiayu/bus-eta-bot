package main

import (
	"time"

	"github.com/pkg/errors"
	"github.com/yi-jiayu/datamall"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/search"
	"google.golang.org/appengine/urlfetch"
)

func PopulateBusStops(ctx context.Context, updateTime time.Time, accountKey, datamallEndpoint string) error {
	// wrap context in namespace
	ctx, err := appengine.Namespace(ctx, namespace)
	if err != nil {
		return errors.Wrap(err, "error wrapping context in namespace")
	}

	log.Infof(ctx, "Updating bus stops at %v.", updateTime)

	client := urlfetch.Client(ctx)
	datamallAPI := datamall.APIClient{
		Endpoint:   datamallEndpoint,
		AccountKey: accountKey,
		Client:     client,
	}

	index, err := search.Open("BusStops")
	if err != nil {
		return errors.Wrap(err, "error opening search index")
	}

	offset := 0
	for {
		busStops, err := datamallAPI.GetBusStops(offset)
		if err != nil {
			return errors.Wrapf(err, "error fetching bus stops at offset %d", offset)
		}

		if len(busStops.Value) == 0 {
			break
		}

		for _, bs := range busStops.Value {
			busStop := BusStop{
				ID:          bs.BusStopCode,
				BusStopID:   bs.BusStopCode,
				Description: bs.Description,
				Road:        bs.RoadName,
				Location: appengine.GeoPoint{
					Lat: bs.Latitude,
					Lng: bs.Longitude,
				},
				UpdatedTime: updateTime,
			}

			key := datastore.NewKey(ctx, busStopKind, bs.BusStopCode, 0, nil)
			_, err := datastore.Put(ctx, key, &busStop)
			if err != nil {
				return errors.Wrapf(err, "error storing bus stop %v at offset %d into datastore", busStop, offset)
			}

			_, err = index.Put(ctx, bs.BusStopCode, &busStop)
			if err != nil {
				return errors.Wrapf(err, "error storing bus stop %v at offset %d into search", busStop, offset)
			}
		}

		offset += 50
	}

	return nil
}
