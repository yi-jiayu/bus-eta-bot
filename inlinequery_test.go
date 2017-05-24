package main

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/yi-jiayu/datamall"
	"github.com/yi-jiayu/telegram-bot-api"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/search"
)

func TestInlineQueryHandler(t *testing.T) {
	t.Parallel()

	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	busStops := []BusStop{
		{
			BusStopID:   "96041",
			Road:        "Upp Changi Rd East",
			Description: "Bef Tropicana Condo",
			Location: appengine.GeoPoint{
				Lat: 1.34041450268626,
				Lng: 103.96127892061004,
			},
		},
		{
			BusStopID:   "96049",
			Road:        "Upp Changi Rd East",
			Description: "Opp Tropicana Condo",
			Location: appengine.GeoPoint{
				Lat: 1.33995375346513,
				Lng: 103.96079768187379,
			},
		},
	}

	index, err := search.Open("BusStops")
	if err != nil {
		t.Fatal(err)
	}
	for _, bs := range busStops {
		_, err := index.Put(ctx, bs.BusStopID, &bs)
		if err != nil {
			t.Fatal(err)
		}
	}

	tgAPI, reqChan, errChan := NewMockTelegramAPIWithPath()
	defer tgAPI.Close()

	tg := &tgbotapi.BotAPI{
		APIEndpoint: tgAPI.URL + "/bot%s/%s",
		Client:      http.DefaultClient,
	}

	bot := NewBusEtaBot(handlers, tg, nil)
	bot.StreetView = &StreetViewAPI{
		Endpoint: StreetViewEndpoint,
		APIKey:   "API_KEY",
	}

	testCases := []struct {
		Name     string
		Query    string
		Offset   string
		Expected Request
	}{
		{
			Name:  "Empty query",
			Query: "",
			Expected: Request{
				Path: "/bot/answerInlineQuery",
				Body: "cache_time=0&inline_query_id=1&is_personal=false&next_offset=&results=%5B%7B%22type%22%3A%22article%22%2C%22id%22%3A%2296049%22%2C%22title%22%3A%22Opp+Tropicana+Condo+%2896049%29%22%2C%22input_message_content%22%3A%7B%22message_text%22%3A%22%2AOpp+Tropicana+Condo+%2896049%29%2A%5CnUpp+Changi+Rd+East%5Cn%60Fetching+etas...%60%22%2C%22parse_mode%22%3A%22markdown%22%2C%22disable_web_page_preview%22%3Afalse%7D%2C%22reply_markup%22%3A%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A%5C%2296049%5C%22%2C%5C%22s%5C%22%3Anull%7D%22%7D%5D%5D%7D%2C%22url%22%3A%22%22%2C%22hide_url%22%3Afalse%2C%22description%22%3A%22Upp+Changi+Rd+East%22%2C%22thumb_url%22%3A%22https%3A%2F%2Fmaps.googleapis.com%2Fmaps%2Fapi%2Fstreetview%3Fkey%3DAPI_KEY%5Cu0026location%3D1.339954%252C103.960798%5Cu0026size%3D100x100%22%2C%22thumb_width%22%3A0%2C%22thumb_height%22%3A0%7D%2C%7B%22type%22%3A%22article%22%2C%22id%22%3A%2296041%22%2C%22title%22%3A%22Bef+Tropicana+Condo+%2896041%29%22%2C%22input_message_content%22%3A%7B%22message_text%22%3A%22%2ABef+Tropicana+Condo+%2896041%29%2A%5CnUpp+Changi+Rd+East%5Cn%60Fetching+etas...%60%22%2C%22parse_mode%22%3A%22markdown%22%2C%22disable_web_page_preview%22%3Afalse%7D%2C%22reply_markup%22%3A%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A%5C%2296041%5C%22%2C%5C%22s%5C%22%3Anull%7D%22%7D%5D%5D%7D%2C%22url%22%3A%22%22%2C%22hide_url%22%3Afalse%2C%22description%22%3A%22Upp+Changi+Rd+East%22%2C%22thumb_url%22%3A%22https%3A%2F%2Fmaps.googleapis.com%2Fmaps%2Fapi%2Fstreetview%3Fkey%3DAPI_KEY%5Cu0026location%3D1.340415%252C103.961279%5Cu0026size%3D100x100%22%2C%22thumb_width%22%3A0%2C%22thumb_height%22%3A0%7D%5D&switch_pm_parameter=&switch_pm_text=",
			},
		},
		{
			Name:  "Matching query",
			Query: "tropicana",
			Expected: Request{
				Path: "/bot/answerInlineQuery",
				Body: "cache_time=0&inline_query_id=1&is_personal=false&next_offset=&results=%5B%7B%22type%22%3A%22article%22%2C%22id%22%3A%2296049%22%2C%22title%22%3A%22Opp+Tropicana+Condo+%2896049%29%22%2C%22input_message_content%22%3A%7B%22message_text%22%3A%22%2AOpp+Tropicana+Condo+%2896049%29%2A%5CnUpp+Changi+Rd+East%5Cn%60Fetching+etas...%60%22%2C%22parse_mode%22%3A%22markdown%22%2C%22disable_web_page_preview%22%3Afalse%7D%2C%22reply_markup%22%3A%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A%5C%2296049%5C%22%2C%5C%22s%5C%22%3Anull%7D%22%7D%5D%5D%7D%2C%22url%22%3A%22%22%2C%22hide_url%22%3Afalse%2C%22description%22%3A%22Upp+Changi+Rd+East%22%2C%22thumb_url%22%3A%22https%3A%2F%2Fmaps.googleapis.com%2Fmaps%2Fapi%2Fstreetview%3Fkey%3DAPI_KEY%5Cu0026location%3D1.339954%252C103.960798%5Cu0026size%3D100x100%22%2C%22thumb_width%22%3A0%2C%22thumb_height%22%3A0%7D%2C%7B%22type%22%3A%22article%22%2C%22id%22%3A%2296041%22%2C%22title%22%3A%22Bef+Tropicana+Condo+%2896041%29%22%2C%22input_message_content%22%3A%7B%22message_text%22%3A%22%2ABef+Tropicana+Condo+%2896041%29%2A%5CnUpp+Changi+Rd+East%5Cn%60Fetching+etas...%60%22%2C%22parse_mode%22%3A%22markdown%22%2C%22disable_web_page_preview%22%3Afalse%7D%2C%22reply_markup%22%3A%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A%5C%2296041%5C%22%2C%5C%22s%5C%22%3Anull%7D%22%7D%5D%5D%7D%2C%22url%22%3A%22%22%2C%22hide_url%22%3Afalse%2C%22description%22%3A%22Upp+Changi+Rd+East%22%2C%22thumb_url%22%3A%22https%3A%2F%2Fmaps.googleapis.com%2Fmaps%2Fapi%2Fstreetview%3Fkey%3DAPI_KEY%5Cu0026location%3D1.340415%252C103.961279%5Cu0026size%3D100x100%22%2C%22thumb_width%22%3A0%2C%22thumb_height%22%3A0%7D%5D&switch_pm_parameter=&switch_pm_text=",
			},
		},
		{
			Name:   "Offset query",
			Query:  "anaciport",
			Offset: "50",
			Expected: Request{
				Path: "/bot/answerInlineQuery",
				Body: "cache_time=0&inline_query_id=1&is_personal=false&next_offset=&results=%5B%5D&switch_pm_parameter=&switch_pm_text=",
			},
		},
		{
			Name:  "Non-matching query",
			Query: "anaciport",
			Expected: Request{
				Path: "/bot/answerInlineQuery",
				Body: "cache_time=0&inline_query_id=1&is_personal=false&next_offset=&results=%5B%5D&switch_pm_parameter=&switch_pm_text=",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ilq := MockInlineQuery()
			ilq.Query = tc.Query
			ilq.Offset = tc.Offset

			err := InlineQueryHandler(ctx, &bot, &ilq)
			if err != nil {
				t.Fatal(err)
			}

			select {
			case req := <-reqChan:
				actual := req
				expected := tc.Expected

				if actual != expected {
					fmt.Printf("Expected:\n%#v\nActual:\n%#v\n", expected, actual)
					t.Fail()
				}
			case err := <-errChan:
				t.Fatalf("%v", err)
			}
		})
	}
}

func TestChosenInlineResultHandler(t *testing.T) {
	t.Parallel()

	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	busStop := BusStop{
		BusStopID:   "96049",
		Road:        "Upp Changi Rd East",
		Description: "Opp Tropicana Condo",
	}

	key := datastore.NewKey(ctx, busStopKind, "96049", 0, nil)
	_, err = datastore.Put(ctx, key, &busStop)
	if err != nil {
		t.Fatal(err)
	}

	tgAPI, reqChan, errChan := NewMockTelegramAPIWithPath()
	defer tgAPI.Close()

	now, _ := time.Parse(time.RFC3339, time.RFC3339)
	nowFunc = func() time.Time {
		return now
	}
	dmAPI, err := NewMockBusArrivalAPI(now)
	if err != nil {
		t.Fatal(err)
	}
	defer dmAPI.Close()

	tg := &tgbotapi.BotAPI{
		APIEndpoint: tgAPI.URL + "/bot%s/%s",
		Client:      http.DefaultClient,
	}

	dm := &datamall.APIClient{
		Endpoint: dmAPI.URL,
		Client:   http.DefaultClient,
	}

	bot := NewBusEtaBot(handlers, tg, dm)

	cir := tgbotapi.ChosenInlineResult{
		ResultID: "96049",
		From: &tgbotapi.User{
			ID:        1,
			FirstName: "Jiayu",
		},
	}

	err = ChosenInlineResultHandler(ctx, &bot, &cir)
	if err != nil {
		t.Fatal(err)
	}

	select {
	case req := <-reqChan:
		actual := req
		expected := Request{
			Path: "/bot/editMessageText",
			Body: "chat_id=0&disable_web_page_preview=false&message_id=0&parse_mode=markdown&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A%5C%2296049%5C%22%2C%5C%22s%5C%22%3Anull%7D%22%7D%5D%5D%7D&text=%2AOpp+Tropicana+Condo+%2896049%29%2A%0AUpp+Changi+Rd+East%0A%60%60%60%0A%7C+Svc+%7C+Next+%7C++2nd+%7C++3rd+%7C%0A%7C-----%7C------%7C------%7C------%7C%0A%7C+2+++%7C+++-1+%7C+++10+%7C+++36+%7C%0A%7C+24++%7C++++1+%7C++++3+%7C++++6+%7C%60%60%60%0AShowing+2+out+of+2+services+for+this+bus+stop.%0A%0A_Last+updated+at+01+Jan+01+00%3A00+UTC_",
		}

		if actual != expected {
			fmt.Printf("Expected:\n%#v\nActual:\n%#v\n", expected, actual)
			t.Fail()
		}
	case err := <-errChan:
		t.Fatalf("%v", err)
	}
}
