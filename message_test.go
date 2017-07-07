package main

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/yi-jiayu/datamall"
	"github.com/yi-jiayu/telegram-bot-api"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/search"
)

func TestTextHandler(t *testing.T) {
	t.Parallel()

	ctx, done, err := NewDevContext()
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

	bot := NewBusEtaBot(handlers, tg, dm, nil, nil)
	bot.NowFunc = func() time.Time {
		return now
	}

	testCases := []struct {
		Name     string
		Text     string
		Expected Request
	}{
		{
			Name: "Bus stop code only",
			Text: "96049",
			Expected: Request{
				Path: "/bot/sendMessage",
				Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&parse_mode=markdown&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A%5C%2296049%5C%22%7D%22%7D%2C%7B%22text%22%3A%22Resend%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22resend%5C%22%2C%5C%22b%5C%22%3A%5C%2296049%5C%22%7D%22%7D%5D%5D%7D&text=%2AOpp+Tropicana+Condo+%2896049%29%2A%0AUpp+Changi+Rd+East%0A%60%60%60%0A%7C+Svc+%7C+Next+%7C++2nd+%7C++3rd+%7C%0A%7C-----%7C------%7C------%7C------%7C%0A%7C+2+++%7C+++-1+%7C+++10+%7C+++36+%7C%0A%7C+24++%7C++++1+%7C++++3+%7C++++6+%7C%60%60%60%0AShowing+2+out+of+2+services+for+this+bus+stop.%0A%0A_Last+updated+at+01+Jan+01+08%3A00+SGT_",
			},
		},
		{
			Name: "Bus stop code and services",
			Text: "96049 2",
			Expected: Request{
				Path: "/bot/sendMessage",
				Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&parse_mode=markdown&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A%5C%2296049%5C%22%2C%5C%22s%5C%22%3A%5B%5C%222%5C%22%5D%7D%22%7D%2C%7B%22text%22%3A%22Resend%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22resend%5C%22%2C%5C%22b%5C%22%3A%5C%2296049%5C%22%2C%5C%22s%5C%22%3A%5B%5C%222%5C%22%5D%7D%22%7D%5D%5D%7D&text=%2AOpp+Tropicana+Condo+%2896049%29%2A%0AUpp+Changi+Rd+East%0A%60%60%60%0A%7C+Svc+%7C+Next+%7C++2nd+%7C++3rd+%7C%0A%7C-----%7C------%7C------%7C------%7C%0A%7C+2+++%7C+++-1+%7C+++10+%7C+++36+%7C%60%60%60%0AShowing+1+out+of+2+services+for+this+bus+stop.%0A%0A_Last+updated+at+01+Jan+01+08%3A00+SGT_",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			message := MockMessage()
			message.Text = tc.Text

			err := TextHandler(ctx, &bot, &message)
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

func TestLocationHandler(t *testing.T) {
	t.Parallel()

	ctx, done, err := NewDevContext()
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

	bot := NewBusEtaBot(handlers, tg, nil, nil, nil)

	message := MockMessage()
	message.Location = &tgbotapi.Location{
		Latitude:  1.34041450268626,
		Longitude: 103.96127892061004,
	}

	err = LocationHandler(ctx, &bot, &message)
	if err != nil {
		t.Fatal(err)
	}

	var reqs []Request
	timer := time.NewTimer(10 * time.Second)

	for len(reqs) < 3 {
		select {
		case req := <-reqChan:
			reqs = append(reqs, req)
		case err := <-errChan:
			t.Fatal(err)
		case <-timer.C:
			t.Fatal("timed out!")
		}
	}

	actual := reqs
	expected := []Request{
		{Path: "/bot/sendMessage", Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&text=Here+are+some+bus+stops+near+your+location%3A"},
		{Path: "/bot/sendVenue", Body: "address=0.00+m+away&chat_id=1&disable_notification=false&latitude=1.340415&longitude=103.961279&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Get+etas%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22new_eta%5C%22%2C%5C%22b%5C%22%3A%5C%2296041%5C%22%7D%22%7D%5D%5D%7D&title=Bef+Tropicana+Condo+%2896041%29"},
		{Path: "/bot/sendVenue", Body: "address=74.15+m+away&chat_id=1&disable_notification=false&latitude=1.339954&longitude=103.960798&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Get+etas%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22new_eta%5C%22%2C%5C%22b%5C%22%3A%5C%2296049%5C%22%7D%22%7D%5D%5D%7D&title=Opp+Tropicana+Condo+%2896049%29"},
	}

	if !sliceCompare(actual, expected) {
		fmt.Println("Expected:")
		for _, e := range expected {
			fmt.Printf("%#v\n", e)
		}
		fmt.Println("Actual:")
		for _, a := range actual {
			fmt.Printf("%#v\n", a)
		}
		t.Fail()
	}
}

func TestLocationHandlerNothingNearby(t *testing.T) {
	t.Parallel()

	ctx, done, err := NewDevContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	tgAPI, reqChan, errChan := NewMockTelegramAPIWithPath()
	defer tgAPI.Close()

	tg := &tgbotapi.BotAPI{
		APIEndpoint: tgAPI.URL + "/bot%s/%s",
		Client:      http.DefaultClient,
	}

	bot := NewBusEtaBot(handlers, tg, nil, nil, nil)

	message := MockMessage()
	message.Location = &tgbotapi.Location{
		Latitude:  1.34041450268626,
		Longitude: 103.96127892061004,
	}

	err = LocationHandler(ctx, &bot, &message)
	if err != nil {
		t.Fatal(err)
	}

	select {
	case req := <-reqChan:
		actual := req
		expected := Request{
			Path: "/bot/sendMessage",
			Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&text=Oops%2C+I+couldn%27t+find+any+bus+stops+within+500+m+of+your+location.",
		}

		if actual != expected {
			fmt.Printf("Expected:\n%#v\nActual:\n%#v\n", expected, actual)
			t.Fail()
		}
	case err := <-errChan:
		t.Fatal(err)
	}
}
