package main

import (
	"context"
	"os"

	"github.com/pkg/errors"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

const (
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
