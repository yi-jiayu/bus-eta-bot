package main

import (
	"context"
	"os"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

const (
	busStopKind         = "BusStop"
	userPreferencesKind = "UserPreferences"
	favouritesKind      = "Favourites"
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

// Favourites contains a user's saved favourites.
type Favourites struct {
	Favourites []string
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

// GetUserFavourites retrieved a user's saved favourites.
func GetUserFavourites(ctx context.Context, userID int) ([]string, error) {
	// set namespace
	ctx, err := appengine.Namespace(ctx, namespace)
	if err != nil {
		return nil, err
	}

	var favourites Favourites
	key := datastore.NewKey(ctx, favouritesKind, "", int64(userID), nil)
	err = datastore.Get(ctx, key, &favourites)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return nil, nil
		}

		return nil, err
	}

	return favourites.Favourites, nil
}

// SetUserFavourites sets a user's saved favourites.
func SetUserFavourites(ctx context.Context, userID int, favourites []string) error {
	// set namespace
	ctx, err := appengine.Namespace(ctx, namespace)
	if err != nil {
		return err
	}

	favs := Favourites{
		Favourites: favourites,
	}

	key := datastore.NewKey(ctx, favouritesKind, "", int64(userID), nil)
	_, err = datastore.Put(ctx, key, &favs)
	return err
}
