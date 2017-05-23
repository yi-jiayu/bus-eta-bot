package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/yi-jiayu/datamall"
	"google.golang.org/appengine/aetest"
)

func newArrival(t time.Time) datamall.BusArrival {
	return datamall.BusArrival{
		BusStopID: "96049",
		Services: []datamall.Service{
			{
				ServiceNo: "2",
				NextBus: datamall.ArrivingBus{
					EstimatedArrival: t.Add(-100 * time.Second).Format(time.RFC3339),
				},
				SubsequentBus: datamall.ArrivingBus{
					EstimatedArrival: t.Add(600 * time.Second).Format(time.RFC3339),
				},
				SubsequentBus3: datamall.ArrivingBus{
					EstimatedArrival: t.Add(2200 * time.Second).Format(time.RFC3339),
				},
			},
			{
				ServiceNo: "24",
				NextBus: datamall.ArrivingBus{
					EstimatedArrival: t.Add(100 * time.Second).Format(time.RFC3339),
				},
				SubsequentBus: datamall.ArrivingBus{
					EstimatedArrival: t.Add(200 * time.Second).Format(time.RFC3339),
				},
				SubsequentBus3: datamall.ArrivingBus{
					EstimatedArrival: t.Add(400 * time.Second).Format(time.RFC3339),
				},
			},
		},
	}
}

func NewMockBusArrivalAPI(t time.Time) (*httptest.Server, error) {
	busArrival, err := json.Marshal(newArrival(t))
	if err != nil {
		return nil, err
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(busArrival)
	}))

	return ts, nil
}

func NewEmptyMockBusArrivalAPI(t time.Time) (*httptest.Server, error) {
	busArrival, err := json.Marshal(datamall.BusArrival{
		Services: []datamall.Service{},
	})
	if err != nil {
		return nil, err
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(busArrival)
	}))

	return ts, nil
}

func TestInferEtaQuery(t *testing.T) {
	t.Parallel()

	t.Run("Bus stop ID only", func(t *testing.T) {
		query := "96049"

		busStopID, serviceNos := InferEtaQuery(query)
		actual := struct {
			BusStopID  string
			ServiceNos []string
		}{
			busStopID,
			serviceNos,
		}
		expected := struct {
			BusStopID  string
			ServiceNos []string
		}{
			"96049",
			[]string{},
		}

		if !reflect.DeepEqual(actual, expected) {
			fmt.Printf("Expected:\n%#v\nActual:\n%#v\n", expected, actual)
			t.Fail()
		}
	})
	t.Run("Bus stop ID and services", func(t *testing.T) {
		query := "96049 2 24"

		busStopID, serviceNos := InferEtaQuery(query)
		actual := struct {
			BusStopID  string
			ServiceNos []string
		}{
			busStopID,
			serviceNos,
		}
		expected := struct {
			BusStopID  string
			ServiceNos []string
		}{
			"96049",
			[]string{"2", "24"},
		}

		if !reflect.DeepEqual(actual, expected) {
			fmt.Printf("Expected:\n%#v\nActual:\n%#v\n", expected, actual)
			t.Fail()
		}
	})
}

func TestCalculateEtas(t *testing.T) {
	t.Parallel()

	now, _ := time.Parse(time.RFC3339, time.RFC3339)

	t.Run("In operation, all etas", func(t *testing.T) {
		etas, err := CalculateEtas(now, newArrival(now))
		if err != nil {
			t.Fatalf("%v", err)
		}

		actual := etas.Services
		expected := [][4]string{{"2", "-1", "10", "36"}, {"24", "1", "3", "6"}}

		if !reflect.DeepEqual(actual, expected) {
			fmt.Printf("Expected:\n%#v\nActual:\n%#v\n", expected, actual)
			t.Fail()
		}
	})
	t.Run("In operation, missing etas", func(t *testing.T) {
		arrival := newArrival(now)
		arrival.Services[0].SubsequentBus.EstimatedArrival = ""
		arrival.Services[0].SubsequentBus3.EstimatedArrival = ""

		etas, err := CalculateEtas(now, arrival)
		if err != nil {
			t.Fatalf("%v", err)
		}

		actual := etas.Services
		expected := [][4]string{{"2", "-1", "?", "?"}, {"24", "1", "3", "6"}}

		if !reflect.DeepEqual(actual, expected) {
			fmt.Printf("Expected:\n%#v\nActual:\n%#v\n", expected, actual)
			t.Fail()
		}
	})
	t.Run("In operation, no etas", func(t *testing.T) {
		arrival := newArrival(now)
		arrival.Services[0].NextBus.EstimatedArrival = ""
		arrival.Services[0].SubsequentBus.EstimatedArrival = ""
		arrival.Services[0].SubsequentBus3.EstimatedArrival = ""

		etas, err := CalculateEtas(now, arrival)
		if err != nil {
			t.Fatalf("%v", err)
		}

		actual := etas.Services
		expected := [][4]string{{"2", "?", "?", "?"}, {"24", "1", "3", "6"}}

		if !reflect.DeepEqual(actual, expected) {
			fmt.Printf("Expected:\n%#v\nActual:\n%#v\n", expected, actual)
			t.Fail()
		}
	})
	t.Run("Not in operation, all etas", func(t *testing.T) {
		arrival := newArrival(now)
		arrival.Services[0].Status = "Not In Operation"

		etas, err := CalculateEtas(now, arrival)
		if err != nil {
			t.Fatalf("%v", err)
		}

		actual := etas.Services
		expected := [][4]string{
			{"2", "-1", "10", "36"},
			{"24", "1", "3", "6"},
		}

		if !reflect.DeepEqual(actual, expected) {
			fmt.Printf("Expected:\n%#v\nActual:\n%#v\n", expected, actual)
			t.Fail()
		}
	})
	t.Run("Not in operation, missing etas", func(t *testing.T) {
		arrival := newArrival(now)
		arrival.Services[0].Status = "Not In Operation"
		arrival.Services[0].SubsequentBus.EstimatedArrival = ""
		arrival.Services[0].SubsequentBus3.EstimatedArrival = ""

		etas, err := CalculateEtas(now, arrival)
		if err != nil {
			t.Fatalf("%v", err)
		}

		actual := etas.Services
		expected := [][4]string{{"2", "-1", "-", "-"}, {"24", "1", "3", "6"}}

		if !reflect.DeepEqual(actual, expected) {
			fmt.Printf("Expected:\n%#v\nActual:\n%#v\n", expected, actual)
			t.Fail()
		}
	})
	t.Run("Not in operation, no etas", func(t *testing.T) {
		arrival := newArrival(now)
		arrival.Services[0].Status = "Not In Operation"
		arrival.Services[0].NextBus.EstimatedArrival = ""
		arrival.Services[0].SubsequentBus.EstimatedArrival = ""
		arrival.Services[0].SubsequentBus3.EstimatedArrival = ""

		etas, err := CalculateEtas(now, arrival)
		if err != nil {
			t.Fatalf("%v", err)
		}

		actual := etas.Services
		expected := [][4]string{{"2", "-", "-", "-"}, {"24", "1", "3", "6"}}

		if !reflect.DeepEqual(actual, expected) {
			fmt.Printf("Expected:\n%#v\nActual:\n%#v\n", expected, actual)
			t.Fail()
		}
	})
}

func TestEtaTable(t *testing.T) {
	t.Parallel()

	services := [][4]string{
		{"2", "-2", "12", "20"},
		{"24", "2", "9", "18"},
	}

	expected := "| Svc | Next |  2nd |  3rd |\n|-----|------|------|------|\n| 2   |   -2 |   12 |   20 |\n| 24  |    2 |    9 |   18 |"
	actual := EtaTable(services)

	if actual != expected {
		fmt.Printf("Expected:\n%q\nActual:\n%q\n", expected, actual)
		t.Fail()
	}
}

func TestFormatEtas(t *testing.T) {
	t.Parallel()

	now, _ := time.Parse(time.RFC3339, time.RFC3339)
	busArrival := newArrival(now)

	etas, err := CalculateEtas(now, busArrival)
	if err != nil {
		t.Fatalf("%v", err)
	}

	busStop := BusStop{BusStopID: "96049", Road: "Upp Changi Rd East", Description: "Opp Tropicana Condo"}

	t.Run("Showing all bus stops", func(t *testing.T) {
		actual := FormatEtas(etas, &busStop, nil)
		expected := "*Opp Tropicana Condo (96049)*\nUpp Changi Rd East\n```\n| Svc | Next |  2nd |  3rd |\n|-----|------|------|------|\n| 2   |   -1 |   10 |   36 |\n| 24  |    1 |    3 |    6 |```\nShowing 2 out of 2 services for this bus stop.\n\n_Last updated at 01 Jan 01 00:00 UTC_"

		if actual != expected {
			fmt.Printf("Expected:\n%q\nActual:\n%q\n", expected, actual)
			t.Fail()
		}
	})
	t.Run("Showing filtered bus stops", func(t *testing.T) {
		actual := FormatEtas(etas, &busStop, []string{"2"})
		expected := "*Opp Tropicana Condo (96049)*\nUpp Changi Rd East\n```\n| Svc | Next |  2nd |  3rd |\n|-----|------|------|------|\n| 2   |   -1 |   10 |   36 |```\nShowing 1 out of 2 services for this bus stop.\n\n_Last updated at 01 Jan 01 00:00 UTC_"

		if actual != expected {
			fmt.Printf("Expected:\n%q\nActual:\n%q\n", expected, actual)
			t.Fail()
		}
	})
}

func TestEtaMessage(t *testing.T) {
	t.Parallel()

	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	now, _ := time.Parse(time.RFC3339, time.RFC3339)
	nowFunc = func() time.Time {
		return now
	}
	dmAPI, err := NewEmptyMockBusArrivalAPI(now)
	if err != nil {
		t.Fatal(err)
	}
	defer dmAPI.Close()

	bot := &BusEtaBot{
		Datamall: &datamall.APIClient{
			Endpoint: dmAPI.URL,
			Client:   http.DefaultClient,
		},
	}

	actual, err := EtaMessage(ctx, bot, "invalid", nil)
	if err != nil && err != errNotFound {
		t.Fatal(err)
	}

	expected := "Oh no! I couldn't find any information about bus stop invalid."
	if actual != expected {
		t.Fail()
	}
}
