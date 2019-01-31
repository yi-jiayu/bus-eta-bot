package busetabot

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

const (
	KindFavourites = "Favourites"
	KindUser       = "User"
)

// Favourites contains a user's saved favourites.
type Favourites struct {
	Favourites []string
}

type User struct {
	LastSeenTime time.Time
	Favourites   []string
}

type DatastoreUserRepository struct {
}

func (r *DatastoreUserRepository) UpdateUserLastSeenTime(ctx context.Context, userID int, t time.Time) error {
	ctx, err := appengine.Namespace(ctx, namespace)
	if err != nil {
		return err
	}
	k := datastore.NewKey(ctx, KindUser, "", int64(userID), nil)
	e := &User{
		LastSeenTime: t,
	}
	_, err = datastore.Put(ctx, k, e)
	if err != nil {
		return errors.Wrap(err, "error updating user last seen time")
	}
	return nil
}

func (r *DatastoreUserRepository) GetUserFavourites(ctx context.Context, userID int) (favourites []string, err error) {
	ctx, err = appengine.Namespace(ctx, namespace)
	if err != nil {
		err = errors.Wrap(err, "error setting namespace")
		return
	}
	k := datastore.NewKey(ctx, KindFavourites, "", int64(userID), nil)
	var f Favourites
	err = datastore.Get(ctx, k, &f)
	if err != nil {
		if err != datastore.ErrNoSuchEntity {
			err = errors.Wrap(err, "error getting user favourites")
			return
		}
		return nil, nil
	}
	favourites = f.Favourites
	return
}

func (r *DatastoreUserRepository) SetUserFavourites(ctx context.Context, userID int, favourites []string) error {
	ctx, err := appengine.Namespace(ctx, namespace)
	if err != nil {
		return errors.Wrap(err, "error setting namespace")
	}
	k := datastore.NewKey(ctx, KindFavourites, "", int64(userID), nil)
	err = datastore.RunInTransaction(ctx, func(tc context.Context) (err error) {
		var f Favourites
		err = datastore.Get(tc, k, &f)
		if err != nil && err != datastore.ErrNoSuchEntity {
			return errors.Wrap(err, "error getting user favourites from datastore")
		}
		f.Favourites = favourites
		_, err = datastore.Put(tc, k, &f)
		if err != nil {
			return errors.Wrap(err, "error putting user favourites into datastore")
		}
		return nil
	}, nil)
	if err != nil {
		return errors.Wrap(err, "error updating user favourites in transaction")
	}
	return nil
}
