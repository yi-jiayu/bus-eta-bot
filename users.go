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
