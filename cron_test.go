package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/yi-jiayu/datamall"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/search"
)

func TestUpdateBusStops(t *testing.T) {
	t.Parallel()

	ctx, done, err := NewStronglyConsistentContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	busStops := []datamall.BusStop{
		{
			BusStopCode: "01012",
			RoadName:    "Victoria St",
			Description: "Hotel Grand Pacific",
			Latitude:    1.29684825487647,
			Longitude:   103.85253591654006,
		},
		{
			BusStopCode: "01013",
			RoadName:    "Victoria St",
			Description: "St. Joseph's Ch",
			Latitude:    1.29770970610083,
			Longitude:   103.8532247463225,
		},
	}

	expected := []BusStop{
		{
			ID:          "01012",
			BusStopID:   "01012",
			Road:        "Victoria St",
			Description: "Hotel Grand Pacific",
			Location: appengine.GeoPoint{
				Lat: 1.29684825487647,
				Lng: 103.85253591654006,
			},
			UpdatedTime: time.Time{},
		},
		{
			ID:          "01013",
			BusStopID:   "01013",
			Road:        "Victoria St",
			Description: "St. Joseph's Ch",
			Location: appengine.GeoPoint{
				Lat: 1.29770970610083,
				Lng: 103.8532247463225,
			},
			UpdatedTime: time.Time{},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		offsetStr := r.URL.Query().Get("$skip")
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("failed to parse $skip parameter"))
			return
		}

		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		if offset == 0 {
			w.Write([]byte(`{"value":[{"BusStopCode":"01012","RoadName":"Victoria St","Description":"Hotel Grand Pacific","Latitude":1.29684825487647,"Longitude":103.85253591654006},{"BusStopCode":"01013","RoadName":"Victoria St","Description":"St. Joseph's Ch","Latitude":1.29770970610083,"Longitude":103.8532247463225}]}`))
		} else {
			w.Write([]byte(`{"value":[]}`))
		}
	}))

	index, err := search.Open("BusStops")
	if err != nil {
		t.Fatal(err)
	}

	// check that the bus stops do not exist yet
	for _, bs := range busStops {
		key := datastore.NewKey(ctx, busStopKind, bs.BusStopCode, 0, nil)
		var busStop BusStop

		err := datastore.Get(ctx, key, &busStop)
		if err != nil {
			if err != datastore.ErrNoSuchEntity {
				t.Fatal(err)
			}
		}

		err = index.Get(ctx, bs.BusStopCode, &busStop)
		if err != nil {
			if err != search.ErrNoSuchDocument {
				t.Fatal(err)
			}
		}
	}

	err = PopulateBusStops(ctx, getBotEnvironment(), time.Time{}, "", ts.URL)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	ctx, _ = appengine.Namespace(ctx, getBotEnvironment())

	// check that the bus stops exist now
	for _, bs := range expected {
		key := datastore.NewKey(ctx, busStopKind, bs.ID, 0, nil)
		var busStop BusStop

		err := datastore.Get(ctx, key, &busStop)
		if err != nil {
			t.Fatal(err)
		}

		if !bs.Equal(busStop) {
			fmt.Println("Datastore lookup:")
			fmt.Printf("Expected:\n%#v\nActual:\n%#v\n", bs, busStop)
			t.Fail()
		}

		err = index.Get(ctx, bs.ID, &busStop)
		if err != nil {
			t.Fatal(err)
		}

		if !bs.Equal(busStop) {
			fmt.Println("Search lookup:")
			fmt.Printf("Expected:\n%#v\nActual:\n%#v\n", bs, busStop)
			t.Fail()
		}
	}
}
