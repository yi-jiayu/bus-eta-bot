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
