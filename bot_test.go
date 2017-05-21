package main

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/yi-jiayu/telegram-bot-api"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
)

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

func TestTextHandler(t *testing.T) {
	t.Parallel()

	tg, bodyChan, errChan := NewMockTelegramAPI()
	defer tg.Close()

	now, _ := time.Parse(time.RFC3339, time.RFC3339)
	nowFunc = func() time.Time {
		return now
	}
	dm, err := NewMockBusArrivalAPI(now)
	defer dm.Close()
	datamallEndpoint = dm.URL

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

	bot := tgbotapi.BotAPI{
		APIEndpoint: tg.URL + "/bot%s/%s",
		Client:      http.DefaultClient,
	}

	testCases := []struct {
		Name     string
		Text     string
		Expected string
	}{
		{
			Name:     "Bus stop code only",
			Text:     "96049",
			Expected: "chat_id=1&disable_notification=false&disable_web_page_preview=false&parse_mode=markdown&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A%5C%2296049%5C%22%2C%5C%22s%5C%22%3A%5B%5D%7D%22%7D%5D%5D%7D&text=%2AOpp+Tropicana+Condo+%2896049%29%2A%0AUpp+Changi+Rd+East%0A%60%60%60%0A%7C+Svc+%7C+Next+%7C++2nd+%7C++3rd+%7C%0A%7C-----%7C------%7C------%7C------%7C%0A%7C+2+++%7C+++-1+%7C+++10+%7C+++36+%7C%0A%7C+24++%7C++++1+%7C++++3+%7C++++6+%7C%60%60%60%0AShowing+2+out+of+2+services+for+this+bus+stop.%0A%0A_Last+updated+at+01+Jan+01+00%3A00+UTC_",
		},
		{
			Name:     "Bus stop code and services",
			Text:     "96049 2",
			Expected: "chat_id=1&disable_notification=false&disable_web_page_preview=false&parse_mode=markdown&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A%5C%2296049%5C%22%2C%5C%22s%5C%22%3A%5B%5C%222%5C%22%5D%7D%22%7D%5D%5D%7D&text=%2AOpp+Tropicana+Condo+%2896049%29%2A%0AUpp+Changi+Rd+East%0A%60%60%60%0A%7C+Svc+%7C+Next+%7C++2nd+%7C++3rd+%7C%0A%7C-----%7C------%7C------%7C------%7C%0A%7C+2+++%7C+++-1+%7C+++10+%7C+++36+%7C%60%60%60%0AShowing+1+out+of+2+services+for+this+bus+stop.%0A%0A_Last+updated+at+01+Jan+01+00%3A00+UTC_",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			message := tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 1},
				Text: tc.Text,
			}

			err := TextHandler(ctx, &bot, &message)
			if err != nil {
				t.Fatalf("%v", err)
			}

			select {
			case body := <-bodyChan:
				actual := string(body)
				expected := tc.Expected

				if actual != expected {
					fmt.Printf("Expected:\n%q\nActual:\n%q\n", expected, actual)
					t.Fail()
				}
			case err := <-errChan:
				t.Fatalf("%v", err)
			}
		})
	}
}
