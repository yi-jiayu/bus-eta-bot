package busetabot

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

const KindUser = "User"

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
		return
	}
	k := datastore.NewKey(ctx, KindUser, "", int64(userID), nil)
	var u User
	err = datastore.Get(ctx, k, &u)
	if err != nil {
		if err != datastore.ErrNoSuchEntity {
			return
		}
		return nil, nil
	}
	favourites = u.Favourites
	return
}

func (r *DatastoreUserRepository) SetUserFavourites(ctx context.Context, userID int, favourites []string) error {
	k := datastore.NewKey(ctx, KindUser, "", int64(userID), nil)
	err := datastore.RunInTransaction(ctx, func(tc context.Context) (err error) {
		var u User
		err = datastore.Get(tc, k, &u)
		if err != nil && err != datastore.ErrNoSuchEntity {
			return errors.Wrap(err, "error getting user from datastore")
		}
		u.Favourites = favourites
		_, err = datastore.Put(tc, k, &u)
		if err != nil {
			return errors.Wrap(err, "error putting user into datastore")
		}
		return nil
	}, nil)
	if err != nil {
		return errors.Wrap(err, "error updating user in transaction")
	}
	return nil
}
