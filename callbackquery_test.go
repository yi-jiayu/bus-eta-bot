package main

import (
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/yi-jiayu/datamall"
	"github.com/yi-jiayu/telegram-bot-api"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
)

func sliceCompare(actual, expected []Request) bool {
	sort.Slice(actual, func(i, j int) bool {
		if actual[i].Path == actual[j].Path {
			return actual[i].Body < actual[j].Body
		}

		return actual[i].Path < actual[j].Path
	})

	return reflect.DeepEqual(actual, expected)
}

func TestRefreshCallbackHandler(t *testing.T) {
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
	bot.NowFunc = func() time.Time {
		return now
	}

	message := MockMessage()

	testCases := []struct {
		Name            string
		Data            string
		Message         *tgbotapi.Message
		InlineMessageID string
		Expected1       Request
		Expected2       Request
	}{
		{
			Name:    "Refresh message, bus stop code only",
			Data:    `{"t":"refresh","b": "96049"}`,
			Message: &message,
			Expected1: Request{
				Path: "/bot/answerCallbackQuery",
				Body: "cache_time=0&callback_query_id=1&show_alert=false&text=Etas+updated%21",
			},
			Expected2: Request{
				Path: "/bot/editMessageText",
				Body: "chat_id=1&disable_web_page_preview=false&message_id=0&parse_mode=markdown&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A+%5C%2296049%5C%22%7D%22%7D%5D%5D%7D&text=%2AOpp+Tropicana+Condo+%2896049%29%2A%0AUpp+Changi+Rd+East%0A%60%60%60%0A%7C+Svc+%7C+Next+%7C++2nd+%7C++3rd+%7C%0A%7C-----%7C------%7C------%7C------%7C%0A%7C+2+++%7C+++-1+%7C+++10+%7C+++36+%7C%0A%7C+24++%7C++++1+%7C++++3+%7C++++6+%7C%60%60%60%0AShowing+2+out+of+2+services+for+this+bus+stop.%0A%0A_Last+updated+at+01+Jan+01+00%3A00+UTC_",
			},
		},
		{
			Name:    "Refresh message, bus stop code and services",
			Message: &message,
			Data:    `{"t":"refresh","b":"96049","s":["2","24"]}`,
			Expected1: Request{
				Path: "/bot/answerCallbackQuery",
				Body: "cache_time=0&callback_query_id=1&show_alert=false&text=Etas+updated%21",
			},
			Expected2: Request{
				Path: "/bot/editMessageText",
				Body: "chat_id=1&disable_web_page_preview=false&message_id=0&parse_mode=markdown&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A%5C%2296049%5C%22%2C%5C%22s%5C%22%3A%5B%5C%222%5C%22%2C%5C%2224%5C%22%5D%7D%22%7D%5D%5D%7D&text=%2AOpp+Tropicana+Condo+%2896049%29%2A%0AUpp+Changi+Rd+East%0A%60%60%60%0A%7C+Svc+%7C+Next+%7C++2nd+%7C++3rd+%7C%0A%7C-----%7C------%7C------%7C------%7C%0A%7C+2+++%7C+++-1+%7C+++10+%7C+++36+%7C%0A%7C+24++%7C++++1+%7C++++3+%7C++++6+%7C%60%60%60%0AShowing+2+out+of+2+services+for+this+bus+stop.%0A%0A_Last+updated+at+01+Jan+01+00%3A00+UTC_",
			},
		},
		{
			Name:            "Refresh inline message, bus stop code only",
			Data:            `{"t":"refresh","b": "96049"}`,
			InlineMessageID: "1",
			Expected1: Request{
				Path: "/bot/answerCallbackQuery",
				Body: "cache_time=0&callback_query_id=1&show_alert=false&text=Etas+updated%21",
			},
			Expected2: Request{
				Path: "/bot/editMessageText",
				Body: "disable_web_page_preview=false&inline_message_id=1&parse_mode=markdown&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A+%5C%2296049%5C%22%7D%22%7D%5D%5D%7D&text=%2AOpp+Tropicana+Condo+%2896049%29%2A%0AUpp+Changi+Rd+East%0A%60%60%60%0A%7C+Svc+%7C+Next+%7C++2nd+%7C++3rd+%7C%0A%7C-----%7C------%7C------%7C------%7C%0A%7C+2+++%7C+++-1+%7C+++10+%7C+++36+%7C%0A%7C+24++%7C++++1+%7C++++3+%7C++++6+%7C%60%60%60%0AShowing+2+out+of+2+services+for+this+bus+stop.%0A%0A_Last+updated+at+01+Jan+01+00%3A00+UTC_",
			},
		},
		{
			Name:            "Refresh inline message, bus stop code and services",
			Data:            `{"t":"refresh","b":"96049","s":["2","24"]}`,
			InlineMessageID: "1",
			Expected1: Request{
				Path: "/bot/answerCallbackQuery",
				Body: "cache_time=0&callback_query_id=1&show_alert=false&text=Etas+updated%21",
			},
			Expected2: Request{
				Path: "/bot/editMessageText",
				Body: "disable_web_page_preview=false&inline_message_id=1&parse_mode=markdown&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A%5C%2296049%5C%22%2C%5C%22s%5C%22%3A%5B%5C%222%5C%22%2C%5C%2224%5C%22%5D%7D%22%7D%5D%5D%7D&text=%2AOpp+Tropicana+Condo+%2896049%29%2A%0AUpp+Changi+Rd+East%0A%60%60%60%0A%7C+Svc+%7C+Next+%7C++2nd+%7C++3rd+%7C%0A%7C-----%7C------%7C------%7C------%7C%0A%7C+2+++%7C+++-1+%7C+++10+%7C+++36+%7C%0A%7C+24++%7C++++1+%7C++++3+%7C++++6+%7C%60%60%60%0AShowing+2+out+of+2+services+for+this+bus+stop.%0A%0A_Last+updated+at+01+Jan+01+00%3A00+UTC_",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			cbq := MockCallbackQuery()
			cbq.Data = tc.Data
			cbq.Message = tc.Message
			cbq.InlineMessageID = tc.InlineMessageID

			err := RefreshCallbackHandler(ctx, &bot, &cbq)
			if err != nil {
				t.Fatal(err)
			}

			select {
			case req := <-reqChan:
				actual1 := req
				actual2 := <-reqChan
				expected1 := tc.Expected1
				expected2 := tc.Expected2

				if (actual1 != expected1 || actual2 != expected2) && (actual1 != expected2 || actual2 != expected1) {
					fmt.Printf("Expected:\n%#v\n%#v\nActual:\n%#v\n%#v\n", expected1, expected2, actual1, actual2)
					t.Fail()
				}
			case err := <-errChan:
				t.Fatalf("%v", err)
			}
		})
	}
}

func TestEtaCallbackHandler(t *testing.T) {
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

	message := MockMessage()

	bot := NewBusEtaBot(handlers, tg, dm)
	bot.NowFunc = func() time.Time {
		return now
	}

	testCases := []struct {
		Name            string
		Data            string
		Message         *tgbotapi.Message
		InlineMessageID string
		Expected1       Request
		Expected2       Request
	}{
		{
			Name:    "Eta message, bus stop code only",
			Data:    `{"t":"eta","a": "96049"}`,
			Message: &message,
			Expected1: Request{
				Path: "/bot/answerCallbackQuery",
				Body: "cache_time=0&callback_query_id=1&show_alert=false&text=Etas+updated%21",
			},
			Expected2: Request{
				Path: "/bot/editMessageText",
				Body: "chat_id=1&disable_web_page_preview=false&message_id=0&parse_mode=markdown&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A%5C%2296049%5C%22%7D%22%7D%5D%5D%7D&text=%2AOpp+Tropicana+Condo+%2896049%29%2A%0AUpp+Changi+Rd+East%0A%60%60%60%0A%7C+Svc+%7C+Next+%7C++2nd+%7C++3rd+%7C%0A%7C-----%7C------%7C------%7C------%7C%0A%7C+2+++%7C+++-1+%7C+++10+%7C+++36+%7C%0A%7C+24++%7C++++1+%7C++++3+%7C++++6+%7C%60%60%60%0AShowing+2+out+of+2+services+for+this+bus+stop.%0A%0A_Last+updated+at+01+Jan+01+00%3A00+UTC_",
			},
		},
		{
			Name:    "Refresh message, bus stop code and services",
			Message: &message,
			Data:    `{"t":"eta","a":"96049 2 24"}`,
			Expected1: Request{
				Path: "/bot/answerCallbackQuery",
				Body: "cache_time=0&callback_query_id=1&show_alert=false&text=Etas+updated%21",
			},
			Expected2: Request{
				Path: "/bot/editMessageText",
				Body: "chat_id=1&disable_web_page_preview=false&message_id=0&parse_mode=markdown&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A%5C%2296049%5C%22%2C%5C%22s%5C%22%3A%5B%5C%222%5C%22%2C%5C%2224%5C%22%5D%7D%22%7D%5D%5D%7D&text=%2AOpp+Tropicana+Condo+%2896049%29%2A%0AUpp+Changi+Rd+East%0A%60%60%60%0A%7C+Svc+%7C+Next+%7C++2nd+%7C++3rd+%7C%0A%7C-----%7C------%7C------%7C------%7C%0A%7C+2+++%7C+++-1+%7C+++10+%7C+++36+%7C%0A%7C+24++%7C++++1+%7C++++3+%7C++++6+%7C%60%60%60%0AShowing+2+out+of+2+services+for+this+bus+stop.%0A%0A_Last+updated+at+01+Jan+01+00%3A00+UTC_",
			},
		},
		{
			Name:            "Refresh inline message, bus stop code only",
			Data:            `{"t":"eta","a": "96049"}`,
			InlineMessageID: "1",
			Expected1: Request{
				Path: "/bot/answerCallbackQuery",
				Body: "cache_time=0&callback_query_id=1&show_alert=false&text=Etas+updated%21",
			},
			Expected2: Request{
				Path: "/bot/editMessageText",
				Body: "disable_web_page_preview=false&inline_message_id=1&parse_mode=markdown&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A%5C%2296049%5C%22%7D%22%7D%5D%5D%7D&text=%2AOpp+Tropicana+Condo+%2896049%29%2A%0AUpp+Changi+Rd+East%0A%60%60%60%0A%7C+Svc+%7C+Next+%7C++2nd+%7C++3rd+%7C%0A%7C-----%7C------%7C------%7C------%7C%0A%7C+2+++%7C+++-1+%7C+++10+%7C+++36+%7C%0A%7C+24++%7C++++1+%7C++++3+%7C++++6+%7C%60%60%60%0AShowing+2+out+of+2+services+for+this+bus+stop.%0A%0A_Last+updated+at+01+Jan+01+00%3A00+UTC_",
			},
		},
		{
			Name:            "Refresh inline message, bus stop code and services",
			Data:            `{"t":"eta","a":"96049 2 24"}`,
			InlineMessageID: "1",
			Expected1: Request{
				Path: "/bot/answerCallbackQuery",
				Body: "cache_time=0&callback_query_id=1&show_alert=false&text=Etas+updated%21",
			},
			Expected2: Request{
				Path: "/bot/editMessageText",
				Body: "disable_web_page_preview=false&inline_message_id=1&parse_mode=markdown&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A%5C%2296049%5C%22%2C%5C%22s%5C%22%3A%5B%5C%222%5C%22%2C%5C%2224%5C%22%5D%7D%22%7D%5D%5D%7D&text=%2AOpp+Tropicana+Condo+%2896049%29%2A%0AUpp+Changi+Rd+East%0A%60%60%60%0A%7C+Svc+%7C+Next+%7C++2nd+%7C++3rd+%7C%0A%7C-----%7C------%7C------%7C------%7C%0A%7C+2+++%7C+++-1+%7C+++10+%7C+++36+%7C%0A%7C+24++%7C++++1+%7C++++3+%7C++++6+%7C%60%60%60%0AShowing+2+out+of+2+services+for+this+bus+stop.%0A%0A_Last+updated+at+01+Jan+01+00%3A00+UTC_",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			cbq := MockCallbackQuery()
			cbq.Data = tc.Data
			cbq.Message = tc.Message
			cbq.InlineMessageID = tc.InlineMessageID

			err := EtaCallbackHandler(ctx, &bot, &cbq)
			if err != nil {
				t.Fatal(err)
			}

			select {
			case req := <-reqChan:
				actual1 := req
				actual2 := <-reqChan
				expected1 := tc.Expected1
				expected2 := tc.Expected2

				if (actual1 != expected1 || actual2 != expected2) && (actual1 != expected2 || actual2 != expected1) {
					fmt.Printf("Expected:\n%#v\n%#v\nActual:\n%#v\n%#v\n", expected1, expected2, actual1, actual2)
					t.Fail()
				}
			case err := <-errChan:
				t.Fatalf("%v", err)
			}
		})
	}
}

func TestEtaDemoCallbackHandler(t *testing.T) {
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

	message := MockMessage()

	bot := NewBusEtaBot(handlers, tg, dm)
	bot.NowFunc = func() time.Time {
		return now
	}

	cbq := MockCallbackQuery()
	cbq.Data = `{"t":"eta_demo"}`
	cbq.Message = &message

	err = EtaDemoCallbackHandler(ctx, &bot, &cbq)
	if err != nil {
		t.Fatal(err)
	}

	select {
	case req := <-reqChan:
		actual1 := req
		actual2 := <-reqChan
		expected1 := Request{
			Path: "/bot/answerCallbackQuery",
			Body: "cache_time=0&callback_query_id=1&show_alert=false",
		}
		expected2 := Request{
			Path: "/bot/sendMessage",
			Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&parse_mode=markdown&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A%5C%2296049%5C%22%7D%22%7D%5D%5D%7D&text=%2AOpp+Tropicana+Condo+%2896049%29%2A%0AUpp+Changi+Rd+East%0A%60%60%60%0A%7C+Svc+%7C+Next+%7C++2nd+%7C++3rd+%7C%0A%7C-----%7C------%7C------%7C------%7C%0A%7C+2+++%7C+++-1+%7C+++10+%7C+++36+%7C%0A%7C+24++%7C++++1+%7C++++3+%7C++++6+%7C%60%60%60%0AShowing+2+out+of+2+services+for+this+bus+stop.%0A%0A_Last+updated+at+01+Jan+01+00%3A00+UTC_",
		}

		if (actual1 != expected1 || actual2 != expected2) && (actual1 != expected2 || actual2 != expected1) {
			fmt.Printf("Expected:\n%#v\n%#v\nActual:\n%#v\n%#v\n", expected1, expected2, actual1, actual2)
			t.Fail()
		}
	case err := <-errChan:
		t.Fatalf("%v", err)
	}
}

func TestNewEtaHandler(t *testing.T) {
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
	bot.NowFunc = func() time.Time {
		return now
	}

	message := MockMessage()
	cbq := MockCallbackQuery()
	cbq.Data = `{"t":"new_eta","b":"96049"}`
	cbq.Message = &message

	err = NewEtaHandler(ctx, &bot, &cbq)
	if err != nil {
		t.Fatal(err)
	}

	var reqs []Request
	timer := time.NewTimer(10 * time.Second)

	for len(reqs) < 2 {
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
		{Path: "/bot/answerCallbackQuery", Body: "cache_time=0&callback_query_id=1&show_alert=false"},
		{Path: "/bot/sendMessage", Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&parse_mode=markdown&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22new_eta%5C%22%2C%5C%22b%5C%22%3A%5C%2296049%5C%22%7D%22%7D%5D%5D%7D&text=%2AOpp+Tropicana+Condo+%2896049%29%2A%0AUpp+Changi+Rd+East%0A%60%60%60%0A%7C+Svc+%7C+Next+%7C++2nd+%7C++3rd+%7C%0A%7C-----%7C------%7C------%7C------%7C%0A%7C+2+++%7C+++-1+%7C+++10+%7C+++36+%7C%0A%7C+24++%7C++++1+%7C++++3+%7C++++6+%7C%60%60%60%0AShowing+2+out+of+2+services+for+this+bus+stop.%0A%0A_Last+updated+at+01+Jan+01+00%3A00+UTC_"},
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
