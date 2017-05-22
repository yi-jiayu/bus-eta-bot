package main

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/yi-jiayu/telegram-bot-api"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
)

func TestRefreshCallbackHandler(t *testing.T) {
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
		Name      string
		Data      string
		Message   bool
		Expected1 string
		Expected2 string
	}{
		{
			Name:      "Refresh message, bus stop code only",
			Data:      `{"t":"refresh","b": "96049"}`,
			Message:   true,
			Expected1: "cache_time=0&callback_query_id=1&show_alert=false&text=Etas+updated%21",
			Expected2: "chat_id=1&disable_web_page_preview=false&message_id=0&parse_mode=markdown&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A+%5C%2296049%5C%22%7D%22%7D%5D%5D%7D&text=%2AOpp+Tropicana+Condo+%2896049%29%2A%0AUpp+Changi+Rd+East%0A%60%60%60%0A%7C+Svc+%7C+Next+%7C++2nd+%7C++3rd+%7C%0A%7C-----%7C------%7C------%7C------%7C%0A%7C+2+++%7C+++-1+%7C+++10+%7C+++36+%7C%0A%7C+24++%7C++++1+%7C++++3+%7C++++6+%7C%60%60%60%0AShowing+2+out+of+2+services+for+this+bus+stop.%0A%0A_Last+updated+at+01+Jan+01+00%3A00+UTC_",
		},
		{
			Name:      "Refresh message, bus stop code and services",
			Message:   true,
			Data:      `{"t":"refresh","b":"96049","s":["2","24"]}`,
			Expected1: "cache_time=0&callback_query_id=1&show_alert=false&text=Etas+updated%21",
			Expected2: "chat_id=1&disable_web_page_preview=false&message_id=0&parse_mode=markdown&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A%5C%2296049%5C%22%2C%5C%22s%5C%22%3A%5B%5C%222%5C%22%2C%5C%2224%5C%22%5D%7D%22%7D%5D%5D%7D&text=%2AOpp+Tropicana+Condo+%2896049%29%2A%0AUpp+Changi+Rd+East%0A%60%60%60%0A%7C+Svc+%7C+Next+%7C++2nd+%7C++3rd+%7C%0A%7C-----%7C------%7C------%7C------%7C%0A%7C+2+++%7C+++-1+%7C+++10+%7C+++36+%7C%0A%7C+24++%7C++++1+%7C++++3+%7C++++6+%7C%60%60%60%0AShowing+2+out+of+2+services+for+this+bus+stop.%0A%0A_Last+updated+at+01+Jan+01+00%3A00+UTC_",
		},
		{
			Name:      "Refresh inline message, bus stop code only",
			Data:      `{"t":"refresh","b": "96049"}`,
			Expected1: "cache_time=0&callback_query_id=1&show_alert=false&text=Etas+updated%21",
			Expected2: "disable_web_page_preview=false&inline_message_id=1&parse_mode=markdown&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A+%5C%2296049%5C%22%7D%22%7D%5D%5D%7D&text=%2AOpp+Tropicana+Condo+%2896049%29%2A%0AUpp+Changi+Rd+East%0A%60%60%60%0A%7C+Svc+%7C+Next+%7C++2nd+%7C++3rd+%7C%0A%7C-----%7C------%7C------%7C------%7C%0A%7C+2+++%7C+++-1+%7C+++10+%7C+++36+%7C%0A%7C+24++%7C++++1+%7C++++3+%7C++++6+%7C%60%60%60%0AShowing+2+out+of+2+services+for+this+bus+stop.%0A%0A_Last+updated+at+01+Jan+01+00%3A00+UTC_",
		},
		{
			Name:      "Refresh inline message, bus stop code and services",
			Data:      `{"t":"refresh","b":"96049","s":["2","24"]}`,
			Expected1: "cache_time=0&callback_query_id=1&show_alert=false&text=Etas+updated%21",
			Expected2: "disable_web_page_preview=false&inline_message_id=1&parse_mode=markdown&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A%5C%2296049%5C%22%2C%5C%22s%5C%22%3A%5B%5C%222%5C%22%2C%5C%2224%5C%22%5D%7D%22%7D%5D%5D%7D&text=%2AOpp+Tropicana+Condo+%2896049%29%2A%0AUpp+Changi+Rd+East%0A%60%60%60%0A%7C+Svc+%7C+Next+%7C++2nd+%7C++3rd+%7C%0A%7C-----%7C------%7C------%7C------%7C%0A%7C+2+++%7C+++-1+%7C+++10+%7C+++36+%7C%0A%7C+24++%7C++++1+%7C++++3+%7C++++6+%7C%60%60%60%0AShowing+2+out+of+2+services+for+this+bus+stop.%0A%0A_Last+updated+at+01+Jan+01+00%3A00+UTC_",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			cbq := MockCallbackQuery()
			cbq.Data = tc.Data

			if tc.Message {
				message := MockMessage()
				cbq.Message = &message
			} else {
				cbq.InlineMessageID = "1"
			}

			err := RefreshCallbackHandler(ctx, &bot, &cbq)
			if err != nil {
				t.Fatal(err)
			}

			select {
			case body := <-bodyChan:
				actual1 := string(body)
				actual2 := string(<-bodyChan)
				expected1 := tc.Expected1
				expected2 := tc.Expected2

				if (actual1 != expected1 || actual2 != expected2) && (actual1 != expected2 || actual2 != expected1) {
					fmt.Printf("Expected:\n%q\n%q\nActual:\n%q\n%q\n", expected1, expected2, actual1, actual2)
					t.Fail()
				}
			case err := <-errChan:
				t.Fatalf("%v", err)
			}
		})
	}
}
