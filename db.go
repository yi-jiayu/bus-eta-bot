package main

import (
	"context"
	"os"

	"github.com/pkg/errors"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

const (
	favouritesKind = "Favourites"
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
