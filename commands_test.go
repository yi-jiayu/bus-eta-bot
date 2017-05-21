package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yi-jiayu/telegram-bot-api"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
)

func NewMockTelegramAPI() (*httptest.Server, chan []byte, chan error) {
	bodyChan := make(chan []byte, 1)
	errChan := make(chan error, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				errChan <- err
				return
			}

			bodyChan <- body
		}

		w.Write([]byte(`{"ok":true}`))
	}))

	return ts, bodyChan, errChan
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

func TestAboutHandler(t *testing.T) {
	t.Parallel()

	ts, bodyChan, errChan := NewMockTelegramAPI()
	defer ts.Close()

	bot := tgbotapi.BotAPI{
		APIEndpoint: ts.URL + "/bot%s/%s",
		Client:      http.DefaultClient,
	}

	message := tgbotapi.Message{
		Chat: &tgbotapi.Chat{ID: 1},
	}

	err := AboutHandler(nil, &bot, &message)
	if err != nil {
		t.Fatalf("%v", err)
	}

	select {
	case body := <-bodyChan:
		actual := string(body)
		expected := fmt.Sprintf("chat_id=1&disable_notification=false&disable_web_page_preview=false&text=Bus+Eta+Bot+v%s%%0Ahttps%%3A%%2F%%2Fgithub.com%%2Fyi-jiayu%%2Fbus-eta-bot-3", Version)

		if actual != expected {
			fmt.Printf("Expected:\n%s\nActual:\n%s\n", expected, actual)
			t.Fail()
		}
	case err := <-errChan:
		t.Fatalf("%v", err)
	}
}

func TestStartHandler(t *testing.T) {
	t.Parallel()

	ts, bodyChan, errChan := NewMockTelegramAPI()
	defer ts.Close()

	bot := tgbotapi.BotAPI{
		APIEndpoint: ts.URL + "/bot%s/%s",
		Client:      http.DefaultClient,
	}

	message := tgbotapi.Message{
		Chat: &tgbotapi.Chat{ID: 1},
		From: &tgbotapi.User{FirstName: "Jiayu"},
	}

	err := StartHandler(nil, &bot, &message)
	if err != nil {
		t.Fatalf("%v", err)
	}

	select {
	case body := <-bodyChan:
		actual := string(body)
		expected := "chat_id=1&disable_notification=false&disable_web_page_preview=false&parse_mode=markdown&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Get+etas+for+bus+stop+96049%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22eta_demo%5C%22%7D%22%7D%5D%2C%5B%7B%22text%22%3A%22Try+an+inline+query%22%2C%22switch_inline_query_current_chat%22%3A%22Changi%22%7D%5D%5D%7D&text=Hello+Jiayu%2C%0A%0ABus+Eta+Bot+is+a+Telegram+bot+which+can+tell+you+how+long+you+have+to+wait+for+your+bus+to+arrive.%0A%0ATo+get+started%2C+try+sending+me+a+bus+stop+code+such+as+%6096049%60+to+get+etas+for.%0A%0AAlternatively%2C+you+can+also+search+for+bus+stops+by+sending+me+an+inline+query.+To+try+this+out%2C+type+%40BusEtaBot+followed+by+a+bus+stop+code%2C+description+or+road+name+in+any+chat.%0A%0AThanks+for+trying+out+Bus+Eta+Bot%21+If+you+find+Bus+Eta+Bot+useful%2C+do+help+to+spread+the+word+or+send+%2Ffeedback+to+leave+some+feedback+about+how+to+help+make+Bus+Eta+Bot+even+better%21%0A%0AIf+you%27re+stuck%2C+you+can+send+%2Fhelp+to+view+help."

		if actual != expected {
			fmt.Printf("Expected:\n%q\nActual:\n%q\n", expected, actual)
			t.Fail()
		}
	case err := <-errChan:
		t.Fatalf("%v", err)
	}
}

func TestPrivacyHandler(t *testing.T) {
	t.Parallel()

	ts, bodyChan, errChan := NewMockTelegramAPI()
	defer ts.Close()

	bot := tgbotapi.BotAPI{
		APIEndpoint: ts.URL + "/bot%s/%s",
		Client:      http.DefaultClient,
	}

	message := tgbotapi.Message{
		Chat: &tgbotapi.Chat{ID: 1},
	}

	err := PrivacyHandler(nil, &bot, &message)
	if err != nil {
		t.Fatalf("%v", err)
	}

	select {
	case body := <-bodyChan:
		actual := string(body)
		expected := "chat_id=1&disable_notification=false&disable_web_page_preview=false&parse_mode=markdown&text=You+can+find+Bus+Eta+Bot%27s+privacy+policy+%5Bhere%5D%28http%3A%2F%2Ftelegra.ph%2FBus-Eta-Bot-Privacy-Policy-03-09%29."

		if actual != expected {
			fmt.Printf("Expected:\n%q\nActual:\n%q\n", expected, actual)
			t.Fail()
		}
	case err := <-errChan:
		t.Fatalf("%v", err)
	}
}

func TestEtaHandler(t *testing.T) {
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
			Name:     "No arguments",
			Text:     "/eta",
			Expected: "chat_id=1&disable_notification=false&disable_web_page_preview=false&reply_markup=%7B%22force_reply%22%3Atrue%2C%22selective%22%3Atrue%7D&text=Alright%2C+send+me+a+bus+stop+code+to+get+etas+for.",
		},
		{
			Name:     "Bus stop code only",
			Text:     "/eta 96049",
			Expected: "chat_id=1&disable_notification=false&disable_web_page_preview=false&parse_mode=markdown&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A%5C%2296049%5C%22%2C%5C%22s%5C%22%3A%5B%5D%7D%22%7D%5D%5D%7D&text=%2AOpp+Tropicana+Condo+%2896049%29%2A%0AUpp+Changi+Rd+East%0A%60%60%60%0A%7C+Svc+%7C+Next+%7C++2nd+%7C++3rd+%7C%0A%7C-----%7C------%7C------%7C------%7C%0A%7C+2+++%7C+++-1+%7C+++10+%7C+++36+%7C%0A%7C+24++%7C++++1+%7C++++3+%7C++++6+%7C%60%60%60%0AShowing+2+out+of+2+services+for+this+bus+stop.%0A%0A_Last+updated+at+01+Jan+01+00%3A00+UTC_",
		},
		{
			Name:     "Bus stop code and services",
			Text:     "/eta 96049 2",
			Expected: "chat_id=1&disable_notification=false&disable_web_page_preview=false&parse_mode=markdown&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A%5C%2296049%5C%22%2C%5C%22s%5C%22%3A%5B%5C%222%5C%22%5D%7D%22%7D%5D%5D%7D&text=%2AOpp+Tropicana+Condo+%2896049%29%2A%0AUpp+Changi+Rd+East%0A%60%60%60%0A%7C+Svc+%7C+Next+%7C++2nd+%7C++3rd+%7C%0A%7C-----%7C------%7C------%7C------%7C%0A%7C+2+++%7C+++-1+%7C+++10+%7C+++36+%7C%60%60%60%0AShowing+1+out+of+2+services+for+this+bus+stop.%0A%0A_Last+updated+at+01+Jan+01+00%3A00+UTC_",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			message := tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: 1},
				Text: tc.Text,
			}

			err := EtaHandler(ctx, &bot, &message)
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
