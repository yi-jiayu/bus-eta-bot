package main

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yi-jiayu/datamall"
	"github.com/yi-jiayu/telegram-bot-api"
)

type MockBusStops struct {
	BusStop        *BusStopJSON
	NearbyBusStops []BusStopJSON
}

func (b MockBusStops) Search(ctx context.Context, query string, limit int) []BusStopJSON {
	panic("implement me")
}

func (b MockBusStops) Nearby(ctx context.Context, lat, lon, radius float64, limit int) (nearby []NearbyBusStop) {
	for _, bs := range b.NearbyBusStops {
		nearby = append(nearby, NearbyBusStop{
			BusStopJSON: bs,
			Distance:    EuclideanDistanceAtEquator(lat, lon, bs.Latitude, bs.Longitude),
		})
	}
	return
}

func (b MockBusStops) Get(ID string) *BusStopJSON {
	return b.BusStop
}

type MockDatamall struct {
	BusArrival datamall.BusArrivalV2
	Error      error
}

func (d MockDatamall) GetBusArrivalV2(busStopCode string, serviceNo string) (datamall.BusArrivalV2, error) {
	return d.BusArrival, d.Error
}

func newArrival(t time.Time) datamall.BusArrivalV2 {
	return datamall.BusArrivalV2{
		BusStopID: "96049",
		Services: []datamall.ServiceV2{
			{
				ServiceNo: "2",
				NextBus: datamall.ArrivingBusV2{
					EstimatedArrival: t.Add(-100 * time.Second).Format(time.RFC3339),
				},
				NextBus2: datamall.ArrivingBusV2{
					EstimatedArrival: t.Add(600 * time.Second).Format(time.RFC3339),
					Load:             "SDA",
					Type:             "DD",
				},
				NextBus3: datamall.ArrivingBusV2{
					EstimatedArrival: t.Add(2200 * time.Second).Format(time.RFC3339),
					Load:             "LSD",
					Feature:          "WAB",
					Type:             "BD",
				},
			},
			{
				ServiceNo: "24",
				NextBus: datamall.ArrivingBusV2{
					EstimatedArrival: t.Add(100 * time.Second).Format(time.RFC3339),
					Load:             "SEA",
					Type:             "SD",
				},
				NextBus2: datamall.ArrivingBusV2{
					EstimatedArrival: t.Add(200 * time.Second).Format(time.RFC3339),
					Load:             "SDA",
					Type:             "DD",
					Feature:          "WAB",
				},
				NextBus3: datamall.ArrivingBusV2{
					EstimatedArrival: t.Add(400 * time.Second).Format(time.RFC3339),
					Load:             "LSD",
					Type:             "BD",
				},
			},
		},
	}
}

func newEtaMessageReplyMarkupInline(busStopCode string) *tgbotapi.InlineKeyboardMarkup {
	callbackData := fmt.Sprintf(`{"t":"refresh","b":"%s"}`, busStopCode)
	return &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				{
					Text:         "Refresh",
					CallbackData: &callbackData,
				},
			},
		},
	}
}

func TestInferEtaQuery(t *testing.T) {
	testCases := []struct {
		Name     string
		Text     string
		Expected struct {
			BusStopID  string
			ServiceNos []string
			Err        error
		}
	}{
		{
			Name: "Bus stop ID only",
			Text: "96049",
			Expected: struct {
				BusStopID  string
				ServiceNos []string
				Err        error
			}{
				BusStopID:  "96049",
				ServiceNos: []string{},
			},
		},
		{
			Name: "Bus stop ID and services",
			Text: "96049 2 24",
			Expected: struct {
				BusStopID  string
				ServiceNos []string
				Err        error
			}{
				BusStopID:  "96049",
				ServiceNos: []string{"2", "24"},
			},
		},
		{
			Name: "Bus stop ID too long",
			Text: "!@#$%!",
			Expected: struct {
				BusStopID  string
				ServiceNos []string
				Err        error
			}{
				Err: errBusStopIDTooLong,
			},
		},
		{
			Name: "Invalid bus stop ID",
			Text: "!@#$%",
			Expected: struct {
				BusStopID  string
				ServiceNos []string
				Err        error
			}{
				Err: errBusStopIDInvalid,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			text := tc.Text
			busStopID, serviceNos, err := InferEtaQuery(text)
			if err != nil {
				if expected := tc.Expected.Err; err != expected {
					fmt.Printf("Expected error: %v\nActual error:   %v\n", expected, err)
					t.Fail()
				}
			} else {
				actual := struct {
					BusStopID  string
					ServiceNos []string
					Err        error
				}{
					BusStopID:  busStopID,
					ServiceNos: serviceNos,
				}

				if !reflect.DeepEqual(actual, tc.Expected) {
					fmt.Printf("Expected: %v\nActual:   %v\n", tc.Expected, actual)
				}
			}
		})
	}
}

func TestCalculateEtas(t *testing.T) {
	now, _ := time.Parse(time.RFC3339, time.RFC3339)

	t.Run("In operation, all etas", func(t *testing.T) {
		etas, err := CalculateEtas(now, newArrival(now))
		if err != nil {
			t.Fatalf("%v", err)
		}

		actual := etas.Services
		expected := []IncomingBuses{{ServiceNo: "2", Etas: [3]string{"-1", "10", "36"}, Loads: [3]string{"", "SDA", "LSD"}, Features: [3]string{"", "", "WAB"}, Types: [3]string{"", "DD", "BD"}}, {ServiceNo: "24", Etas: [3]string{"1", "3", "6"}, Loads: [3]string{"SEA", "SDA", "LSD"}, Features: [3]string{"", "WAB", ""}, Types: [3]string{"SD", "DD", "BD"}}}

		if !reflect.DeepEqual(actual, expected) {
			fmt.Printf("Expected:\n%#v\nActual:\n%#v\n", expected, actual)
			t.Fail()
		}
	})
	t.Run("In operation, missing etas", func(t *testing.T) {
		arrival := newArrival(now)
		arrival.Services[0].NextBus2.EstimatedArrival = ""
		arrival.Services[0].NextBus3.EstimatedArrival = ""

		etas, err := CalculateEtas(now, arrival)
		if err != nil {
			t.Fatalf("%v", err)
		}

		actual := etas.Services
		expected := []IncomingBuses{{ServiceNo: "2", Etas: [3]string{"-1", "?", "?"}, Loads: [3]string{"", "SDA", "LSD"}, Features: [3]string{"", "", "WAB"}, Types: [3]string{"", "DD", "BD"}}, {ServiceNo: "24", Etas: [3]string{"1", "3", "6"}, Loads: [3]string{"SEA", "SDA", "LSD"}, Features: [3]string{"", "WAB", ""}, Types: [3]string{"SD", "DD", "BD"}}}

		if !reflect.DeepEqual(actual, expected) {
			fmt.Printf("Expected:\n%#v\nActual:\n%#v\n", expected, actual)
			t.Fail()
		}
	})
	t.Run("In operation, no etas", func(t *testing.T) {
		arrival := newArrival(now)
		arrival.Services[0].NextBus.EstimatedArrival = ""
		arrival.Services[0].NextBus2.EstimatedArrival = ""
		arrival.Services[0].NextBus3.EstimatedArrival = ""

		etas, err := CalculateEtas(now, arrival)
		if err != nil {
			t.Fatalf("%v", err)
		}

		actual := etas.Services
		expected := []IncomingBuses{{ServiceNo: "2", Etas: [3]string{"?", "?", "?"}, Loads: [3]string{"", "SDA", "LSD"}, Features: [3]string{"", "", "WAB"}, Types: [3]string{"", "DD", "BD"}}, {ServiceNo: "24", Etas: [3]string{"1", "3", "6"}, Loads: [3]string{"SEA", "SDA", "LSD"}, Features: [3]string{"", "WAB", ""}, Types: [3]string{"SD", "DD", "BD"}}}

		if !reflect.DeepEqual(actual, expected) {
			fmt.Printf("Expected:\n%#v\nActual:\n%#v\n", expected, actual)
			t.Fail()
		}
	})
}

func TestEtaTable(t *testing.T) {
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
	now, _ := time.Parse(time.RFC3339, time.RFC3339)
	busArrival := newArrival(now)

	etas, err := CalculateEtas(now, busArrival)
	if err != nil {
		t.Fatalf("%v", err)
	}

	busStop := BusStopJSON{BusStopCode: "96049", RoadName: "Upp Changi Rd East", Description: "Opp Tropicana Condo"}

	t.Run("Showing all bus stops", func(t *testing.T) {
		actual := FormatEtasMultiple(etas, &busStop, nil)
		expected := "*Opp Tropicana Condo (96049)*\nUpp Changi Rd East\n```\n| Svc | Next |  2nd |  3rd |\n|-----|------|------|------|\n| 2   |   -1 |   10 |   36 |\n| 24  |    1 |    3 |    6 |```\nShowing 2 out of 2 services for this bus stop.\n\n_Last updated at 01 Jan 01 08:00 SGT_"

		if actual != expected {
			fmt.Printf("Expected:\n%q\nActual:\n%q\n", expected, actual)
			t.Fail()
		}
	})
	t.Run("Showing filtered bus stops", func(t *testing.T) {
		actual := FormatEtasMultiple(etas, &busStop, []string{"2"})
		expected := "*Opp Tropicana Condo (96049)*\nUpp Changi Rd East\n```\n| Svc | Next |  2nd |  3rd |\n|-----|------|------|------|\n| 2   |   -1 |   10 |   36 |```\nShowing 1 out of 2 services for this bus stop.\n\n_Last updated at 01 Jan 01 08:00 SGT_"

		if actual != expected {
			fmt.Printf("Expected:\n%q\nActual:\n%q\n", expected, actual)
			t.Fail()
		}
	})
}

func TestEtaMessage(t *testing.T) {
	bot := &BusEtaBot{
		Datamall: MockDatamall{
			BusArrival: datamall.BusArrivalV2{},
			Error:      nil,
		},
		NowFunc: func() time.Time {
			return time.Time{}
		},
		BusStops: MockBusStops{
			BusStop: nil,
		},
	}

	actual, err := EtaMessageText(bot, "invalid", nil)
	if err != nil && err != errNotFound {
		t.Fatal(err)
	}

	expected := "Oh no! I couldn't find any information about bus stop invalid."
	if actual != expected {
		t.Fail()
	}
}

func TestEtaMessageReplyMarkup(t *testing.T) {
	t.Run("for inline message", func(t *testing.T) {
		expected := newEtaMessageReplyMarkupInline("96049")
		actual, err := EtaMessageReplyMarkup("96049", nil, true)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, expected, actual)
	})
}
