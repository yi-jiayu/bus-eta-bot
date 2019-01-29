package busetabot

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
)

func TestDatastoreUserRepository_UpdateUserLastSeenTime(t *testing.T) {
	t.Parallel()

	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	users := new(DatastoreUserRepository)
	err = users.UpdateUserLastSeenTime(ctx, 1000, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	var u User
	ctx, err = appengine.Namespace(ctx, namespace)
	if err != nil {
		t.Fatal(err)
	}
	k := datastore.NewKey(ctx, KindUser, "", 1000, nil)
	err = datastore.Get(ctx, k, &u)
	if err != nil {
		t.Fatal(err)
	}
	expected := User{
		LastSeenTime: time.Time{},
	}
	assert.Equal(t, expected, u)
}

func TestDatastoreUserRepository_GetUserFavourites(t *testing.T) {
	t.Parallel()

	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	ctx, err = appengine.Namespace(ctx, namespace)
	if err != nil {
		t.Fatal(err)
	}
	const userID = 1
	userRepository := new(DatastoreUserRepository)
	t.Run("when user does not have any favourites", func(t *testing.T) {
		actual, err := userRepository.GetUserFavourites(ctx, userID)
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, actual, 0)
	})
	t.Run("when user has empty favourites", func(t *testing.T) {
		k := datastore.NewKey(ctx, KindUser, "", userID, nil)
		u := User{
			Favourites: []string{},
		}
		_, err = datastore.Put(ctx, k, &u)
		if err != nil {
			t.Fatal(err)
		}
		actual, err := userRepository.GetUserFavourites(ctx, userID)
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, actual, 0)
	})
	t.Run("when user has favourites", func(t *testing.T) {
		k := datastore.NewKey(ctx, KindUser, "", userID, nil)
		u := User{
			Favourites: []string{"96049", "81111"},
		}
		_, err = datastore.Put(ctx, k, &u)
		if err != nil {
			t.Fatal(err)
		}
		actual, err := userRepository.GetUserFavourites(ctx, userID)
		if err != nil {
			t.Fatal(err)
		}
		expected := u.Favourites
		assert.Equal(t, expected, actual)
	})
}

func TestDatastoreUserRepository_SetUserFavourites(t *testing.T) {
	t.Parallel()

	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	ctx, err = appengine.Namespace(ctx, namespace)
	if err != nil {
		t.Fatal(err)
	}
	const userID = 1
	favourites := []string{"96049", "81111"}
	userRepository := new(DatastoreUserRepository)
	err = userRepository.SetUserFavourites(ctx, userID, favourites)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	k := datastore.NewKey(ctx, KindUser, "", int64(userID), nil)
	var u User
	err = datastore.Get(ctx, k, &u)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, favourites, u.Favourites)
}
