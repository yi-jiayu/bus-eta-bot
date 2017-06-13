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

// PopulateBusStops gets information about all bus stops from the LTA DataMall API and updates the bus stop information
// in datastore and search.
func PopulateBusStops(ctx context.Context, env string, updateTime time.Time, accountKey, datamallEndpoint string) error {
	// wrap context in namespace
	ctx, err := appengine.Namespace(ctx, env)
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
	count := 0
	for {
		busStops, err := datamallAPI.GetBusStops(offset)
		if err != nil {
			return errors.Wrapf(err, "error fetching bus stops at offset %d", offset)
		}

		if getBotEnvironment() == "dev" && offset > 100 {
			break
		}

		if len(busStops.Value) == 0 {
			break
		}
		log.Infof(ctx, "Got %d bus stops from offset %d", len(busStops.Value), offset)

		keys := make([]*datastore.Key, 0)
		entities := make([]BusStop, 0)

		// compare against existing bus stops
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

			keys = append(keys, datastore.NewKey(ctx, busStopKind, bs.BusStopCode, 0, nil))
			entities = append(entities, busStop)
		}

		existing := make([]BusStop, len(busStops.Value))
		toPutKeys := make([]*datastore.Key, 0)
		toPut := make([]BusStop, 0)
		err = datastore.GetMulti(ctx, keys, existing)
		if err != nil {
			if multiErr, ok := err.(appengine.MultiError); ok {
				for i := range existing {
					if multiErr[i] != nil {
						toPutKeys = append(toPutKeys, keys[i])
						toPut = append(toPut, entities[i])
					} else {
						if !existing[i].Equal(entities[i]) {
							toPutKeys = append(toPutKeys, keys[i])
							toPut = append(toPut, entities[i])
						}
					}
				}
			} else {
				return errors.Wrap(err, "error while getting bus stops from datastore")
			}
		} else {
			for i := range existing {
				if !existing[i].Equal(entities[i]) {
					toPutKeys = append(toPutKeys, keys[i])
					toPut = append(toPut, entities[i])
				}
			}
		}

		log.Infof(ctx, "Updating %d changed bus stops", len(toPut))

		if len(toPut) > 0 {
			_, err = datastore.PutMulti(ctx, toPutKeys, toPut)
			if err != nil {
				return errors.Wrapf(err, "error storing bus stops into datastore at offset %d", offset)
			}

			for _, bs := range toPut {
				_, err = index.Put(ctx, bs.ID, &bs)
				if err != nil {
					return errors.Wrapf(err, "error storing bus stop %v into search at offset %d", bs, offset)
				}
			}
		}

		offset += 50
		count += len(busStops.Value)
	}

	log.Infof(ctx, "Successfully populated database with %d bus stops", count)

	return nil
}
