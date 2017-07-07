package main

import (
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/search"
)

const (
	busStopKind         = "BusStop"
	userPreferencesKind = "UserPreferences"
)

const (
	devEnvironment        = "dev"
	stagingEnvironment    = "staging"
	productionEnvironment = "prod"
)

var (
	errNotFound = errors.New("not found")
)

var namespace = getBotEnvironment()

// BusStop is a bus stop as represented inside app engine datastore and search.
type BusStop struct {
	BusStopID   string
	ID          string
	Road        string
	Description string
	Location    appengine.GeoPoint
	UpdatedTime time.Time
}

// BusStopJSON is a bus stop deserialised from JSON.
type BusStopJSON struct {
	BusStopCode string
	RoadName    string
	Description string
	Latitude    float64
	Longitude   float64
}

// UserPreferences represents a bus eta bot user's preferences.
type UserPreferences struct {
	NoRedundantEtaCommandReminder bool
}

func getBotEnvironment() string {
	switch os.Getenv("BOT_ENVIRONMENT") {
	case stagingEnvironment:
		return stagingEnvironment
	case productionEnvironment:
		return productionEnvironment
	default:
		return devEnvironment
	}
}

// DistanceFrom returns the distance between a bus stop and a reference coordinate.
func (b BusStop) DistanceFrom(lat, lon float64) float64 {
	return Distance(b.Location.Lat, b.Location.Lng, lat, lon)
}

// Equal checks if a bus stop's information is the same as another one.
func (b BusStop) Equal(bs BusStop) bool {
	return b.ID == bs.ID &&
		b.Description == bs.Description &&
		b.Road == bs.Road &&
		b.Location == bs.Location
}

// GetBusStop looks up a bus stop by id from the datastore
func GetBusStop(ctx context.Context, id string) (BusStop, error) {
	// set namespace
	ctx, err := appengine.Namespace(ctx, namespace)
	if err != nil {
		return BusStop{}, errors.Wrap(err, "error setting namespace on context")
	}

	var busStop BusStop
	key := datastore.NewKey(ctx, busStopKind, id, 0, nil)
	err = datastore.Get(ctx, key, &busStop)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return BusStop{}, errNotFound
		}

		return BusStop{}, errors.Wrap(err, "error getting bus stop information from datastore")
	}

	return busStop, nil
}

// GetNearbyBusStops returns nearby bus stops to a specified location.
func GetNearbyBusStops(ctx context.Context, lat, lng float64, radius, limit int) ([]BusStop, error) {
	// set namespace
	ctx, err := appengine.Namespace(ctx, namespace)
	if err != nil {
		return nil, err
	}

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

// SearchBusStops returns bus stops containing query.
func SearchBusStops(ctx context.Context, query string, offset int) ([]BusStop, error) {
	// set namespace
	ctx, err := appengine.Namespace(ctx, namespace)
	if err != nil {
		return nil, err
	}

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

// GetUserPreferences retrieves a user's preferences.
func GetUserPreferences(ctx context.Context, userID int) (UserPreferences, error) {
	// set namespace
	ctx, err := appengine.Namespace(ctx, namespace)
	if err != nil {
		return UserPreferences{}, err
	}

	var prefs UserPreferences
	key := datastore.NewKey(ctx, userPreferencesKind, "", int64(userID), nil)

	err = datastore.Get(ctx, key, &prefs)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return UserPreferences{}, nil
		}

		return UserPreferences{}, err
	}

	return prefs, nil
}

// SetUserPreferences sets a user's preferences.
func SetUserPreferences(ctx context.Context, userID int, prefs *UserPreferences) error {
	// set namespace
	ctx, err := appengine.Namespace(ctx, namespace)
	if err != nil {
		return err
	}

	key := datastore.NewKey(ctx, userPreferencesKind, "", int64(userID), nil)
	_, err = datastore.Put(ctx, key, prefs)
	return err
}
