package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/yi-jiayu/datamall"
	"github.com/yi-jiayu/telegram-bot-api"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
)

func TestAboutHandler(t *testing.T) {
	t.Parallel()

	tgAPI, reqChan, errChan := NewMockTelegramAPIWithPath()
	defer tgAPI.Close()

	tg := &tgbotapi.BotAPI{
		APIEndpoint: tgAPI.URL + "/bot%s/%s",
		Client:      http.DefaultClient,
	}

	bot := NewBusEtaBot(handlers, tg, nil, nil, nil)

	testCases := []struct {
		Name     string
		ChatType string
		Expected Request
	}{
		{
			Name:     "Private chat",
			ChatType: ChatTypePrivate,
			Expected: Request{
				Path: "/bot/sendMessage",
				Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&text=Bus+Eta+Bot+VERSION%0Ahttps%3A%2F%2Fgithub.com%2Fyi-jiayu%2Fbus-eta-bot",
			},
		},
		{
			Name: "Group, supergroup or channel",
			Expected: Request{
				Path: "/bot/sendMessage",
				Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&reply_to_message_id=1&text=Bus+Eta+Bot+VERSION%0Ahttps%3A%2F%2Fgithub.com%2Fyi-jiayu%2Fbus-eta-bot",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			message := MockMessageWithType(tc.ChatType)

			err := AboutHandler(nil, &bot, &message)
			if err != nil {
				t.Fatalf("%v", err)
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

func TestVersionHandler(t *testing.T) {
	t.Parallel()

	tgAPI, reqChan, errChan := NewMockTelegramAPIWithPath()
	defer tgAPI.Close()

	tg := &tgbotapi.BotAPI{
		APIEndpoint: tgAPI.URL + "/bot%s/%s",
		Client:      http.DefaultClient,
	}

	bot := NewBusEtaBot(handlers, tg, nil, nil, nil)

	testCases := []struct {
		Name     string
		ChatType string
		Expected Request
	}{
		{
			Name:     "Private chat",
			ChatType: ChatTypePrivate,
			Expected: Request{
				Path: "/bot/sendMessage",
				Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&text=Bus+Eta+Bot+VERSION%0Ahttps%3A%2F%2Fgithub.com%2Fyi-jiayu%2Fbus-eta-bot",
			},
		},
		{
			Name: "Group, supergroup or channel",
			Expected: Request{
				Path: "/bot/sendMessage",
				Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&reply_to_message_id=1&text=Bus+Eta+Bot+VERSION%0Ahttps%3A%2F%2Fgithub.com%2Fyi-jiayu%2Fbus-eta-bot",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			message := MockMessageWithType(tc.ChatType)

			err := VersionHandler(nil, &bot, &message)
			if err != nil {
				t.Fatalf("%v", err)
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

func TestStartHandler(t *testing.T) {
	t.Parallel()

	tgAPI, reqChan, errChan := NewMockTelegramAPIWithPath()
	defer tgAPI.Close()

	tg := &tgbotapi.BotAPI{
		APIEndpoint: tgAPI.URL + "/bot%s/%s",
		Client:      http.DefaultClient,
	}

	bot := NewBusEtaBot(handlers, tg, nil, nil, nil)
	message := MockMessage()

	err := StartHandler(nil, &bot, &message)
	if err != nil {
		t.Fatalf("%v", err)
	}

	select {
	case req := <-reqChan:
		actual := req
		expected := Request{
			Path: "/bot/sendMessage",
			Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&parse_mode=markdown&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Get+etas+for+bus+stop+96049%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22eta_demo%5C%22%7D%22%7D%5D%2C%5B%7B%22text%22%3A%22Try+an+inline+query%22%2C%22switch_inline_query_current_chat%22%3A%22Tropicana%22%7D%5D%5D%7D&text=Hello+Jiayu%2C%0A%0ABus+Eta+Bot+is+a+Telegram+bot+which+can+tell+you+how+long+you+have+to+wait+for+your+bus+to+arrive.%0A%0ATo+get+started%2C+try+sending+me+a+bus+stop+code+such+as+%6096049%60+to+get+etas+for.%0A%0AAlternatively%2C+you+can+also+search+for+bus+stops+by+sending+me+an+inline+query.+To+try+this+out%2C+type+%40BusEtaBot+followed+by+a+bus+stop+code%2C+description+or+road+name+in+any+chat.%0A%0AThanks+for+trying+out+Bus+Eta+Bot%21+If+you+find+Bus+Eta+Bot+useful%2C+do+help+to+spread+the+word+or+send+%2Ffeedback+to+leave+some+feedback+about+how+to+help+make+Bus+Eta+Bot+even+better%21%0A%0AIf+you%27re+stuck%2C+you+can+send+%2Fhelp+to+view+help.",
		}

		if actual != expected {
			fmt.Printf("Expected:\n%#v\nActual:\n%#v\n", expected, actual)
			t.Fail()
		}
	case err := <-errChan:
		t.Fatalf("%v", err)
	}
}

func TestHelpHandler(t *testing.T) {
	t.Parallel()

	tgAPI, reqChan, errChan := NewMockTelegramAPIWithPath()
	defer tgAPI.Close()

	tg := &tgbotapi.BotAPI{
		APIEndpoint: tgAPI.URL + "/bot%s/%s",
		Client:      http.DefaultClient,
	}

	bot := NewBusEtaBot(handlers, tg, nil, nil, nil)
	message := MockMessage()

	err := HelpHandler(nil, &bot, &message)
	if err != nil {
		t.Fatalf("%v", err)
	}

	select {
	case req := <-reqChan:
		actual := req
		expected := Request{
			Path: "/bot/sendMessage",
			Body: strings.Replace("chat_id=1&disable_notification=false&disable_web_page_preview=false&parse_mode=markdown&text=You+can+find+help+on+how+to+use+Bus+Eta+Bot+%5Bhere%5D%28$HELP_URL%29.", "$HELP_URL", url.QueryEscape(HelpURL), 1),
		}

		if actual != expected {
			fmt.Printf("Expected:\n%#v\nActual:\n%#v\n", expected, actual)
			t.Fail()
		}
	case err := <-errChan:
		t.Fatalf("%v", err)
	}
}

func TestPrivacyHandler(t *testing.T) {
	t.Parallel()

	tgAPI, reqChan, errChan := NewMockTelegramAPIWithPath()
	defer tgAPI.Close()

	tg := &tgbotapi.BotAPI{
		APIEndpoint: tgAPI.URL + "/bot%s/%s",
		Client:      http.DefaultClient,
	}

	sv := NewStreetViewAPI("API_KEY")

	bot := NewBusEtaBot(handlers, tg, nil, &sv, nil)
	message := MockMessage()

	err := PrivacyHandler(nil, &bot, &message)
	if err != nil {
		t.Fatalf("%v", err)
	}

	select {
	case req := <-reqChan:
		actual := req
		expected := Request{
			Path: "/bot/sendMessage",
			Body: strings.Replace("chat_id=1&disable_notification=false&disable_web_page_preview=false&parse_mode=markdown&text=You+can+find+Bus+Eta+Bot%27s+privacy+policy+%5Bhere%5D%28$PRIVACY_POLICY_URL%29.", "$PRIVACY_POLICY_URL", url.QueryEscape(PrivacyPolicyURL), 1),
		}

		if actual != expected {
			fmt.Printf("Expected:\n%#v\nActual:\n%#v\n", expected, actual)
			t.Fail()
		}
	case err := <-errChan:
		t.Fatalf("%v", err)
	}
}

func TestEtaHandler(t *testing.T) {
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
			Name: "No arguments",
			Text: "/eta",
			Expected: Request{
				Path: "/bot/sendMessage",
				Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&reply_markup=%7B%22force_reply%22%3Atrue%2C%22selective%22%3Atrue%7D&text=Alright%2C+send+me+a+bus+stop+code+to+get+etas+for.",
			},
		},
		{
			Name: "Bus stop code only",
			Text: "/eta 96049",
			Expected: Request{
				Path: "/bot/sendMessage",
				Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&parse_mode=markdown&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A%5C%2296049%5C%22%7D%22%7D%5D%5D%7D&text=%2AOpp+Tropicana+Condo+%2896049%29%2A%0AUpp+Changi+Rd+East%0A%60%60%60%0A%7C+Svc+%7C+Next+%7C++2nd+%7C++3rd+%7C%0A%7C-----%7C------%7C------%7C------%7C%0A%7C+2+++%7C+++-1+%7C+++10+%7C+++36+%7C%0A%7C+24++%7C++++1+%7C++++3+%7C++++6+%7C%60%60%60%0AShowing+2+out+of+2+services+for+this+bus+stop.%0A%0A_Last+updated+at+01+Jan+01+00%3A00+UTC_",
			},
		},
		{
			Name: "Bus stop code and services",
			Text: "/eta 96049 2",
			Expected: Request{
				Path: "/bot/sendMessage",
				Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&parse_mode=markdown&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A%5C%2296049%5C%22%2C%5C%22s%5C%22%3A%5B%5C%222%5C%22%5D%7D%22%7D%5D%5D%7D&text=%2AOpp+Tropicana+Condo+%2896049%29%2A%0AUpp+Changi+Rd+East%0A%60%60%60%0A%7C+Svc+%7C+Next+%7C++2nd+%7C++3rd+%7C%0A%7C-----%7C------%7C------%7C------%7C%0A%7C+2+++%7C+++-1+%7C+++10+%7C+++36+%7C%60%60%60%0AShowing+1+out+of+2+services+for+this+bus+stop.%0A%0A_Last+updated+at+01+Jan+01+00%3A00+UTC_",
			},
		},
		{
			Name: "Invalid bus stop code",
			Text: "/eta !#@$% 2",
			Expected: Request{
				Path: "/bot/sendMessage",
				Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&text=Oops%2C+that+did+not+seem+to+be+a+valid+bus+stop+code.",
			},
		},
		{
			Name: "Too long bus stop code",
			Text: "/eta 960499 2",
			Expected: Request{
				Path: "/bot/sendMessage",
				Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&text=Oops%2C+a+bus+stop+code+can+only+contain+a+maximum+of+5+characters.",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			message := MockMessage()
			message.Text = tc.Text

			err := EtaHandler(ctx, &bot, &message)
			if err != nil {
				t.Fatalf("%v", err)
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

func TestEtaHandlerPrivate(t *testing.T) {
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

	bot := NewBusEtaBot(handlers, tg, dm, nil, nil)
	bot.NowFunc = func() time.Time {
		return now
	}

	testCases := []struct {
		Name     string
		Text     string
		Expected []Request
	}{
		{
			Name: "No arguments",
			Text: "/eta",
			Expected: []Request{
				{Path: "/bot/sendMessage", Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&reply_markup=%7B%22force_reply%22%3Atrue%2C%22selective%22%3Atrue%7D&text=Alright%2C+send+me+a+bus+stop+code+to+get+etas+for."},
				{Path: "/bot/sendMessage", Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&text=Did+you+know+that+in+a+private+chat%2C+you+can+just+send+a+bus+stop+code+directly%2C+without+using+the+%2Feta+command%3F"},
			},
		},
		{
			Name: "Bus stop code only",
			Text: "/eta 96049",
			Expected: []Request{
				{Path: "/bot/sendMessage", Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&parse_mode=markdown&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A%5C%2296049%5C%22%7D%22%7D%5D%5D%7D&text=%2AOpp+Tropicana+Condo+%2896049%29%2A%0AUpp+Changi+Rd+East%0A%60%60%60%0A%7C+Svc+%7C+Next+%7C++2nd+%7C++3rd+%7C%0A%7C-----%7C------%7C------%7C------%7C%0A%7C+2+++%7C+++-1+%7C+++10+%7C+++36+%7C%0A%7C+24++%7C++++1+%7C++++3+%7C++++6+%7C%60%60%60%0AShowing+2+out+of+2+services+for+this+bus+stop.%0A%0A_Last+updated+at+01+Jan+01+00%3A00+UTC_"},
				{Path: "/bot/sendMessage", Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&text=Did+you+know+that+in+a+private+chat%2C+you+can+just+send+a+bus+stop+code+directly%2C+without+using+the+%2Feta+command%3F"},
			},
		},
		{
			Name: "Bus stop code and services",
			Text: "/eta 96049 2",
			Expected: []Request{
				{Path: "/bot/sendMessage", Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&parse_mode=markdown&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A%5C%2296049%5C%22%2C%5C%22s%5C%22%3A%5B%5C%222%5C%22%5D%7D%22%7D%5D%5D%7D&text=%2AOpp+Tropicana+Condo+%2896049%29%2A%0AUpp+Changi+Rd+East%0A%60%60%60%0A%7C+Svc+%7C+Next+%7C++2nd+%7C++3rd+%7C%0A%7C-----%7C------%7C------%7C------%7C%0A%7C+2+++%7C+++-1+%7C+++10+%7C+++36+%7C%60%60%60%0AShowing+1+out+of+2+services+for+this+bus+stop.%0A%0A_Last+updated+at+01+Jan+01+00%3A00+UTC_"},
				{Path: "/bot/sendMessage", Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&text=Did+you+know+that+in+a+private+chat%2C+you+can+just+send+a+bus+stop+code+directly%2C+without+using+the+%2Feta+command%3F"},
			},
		},
		{
			Name: "Invalid bus stop code",
			Text: "/eta !#@$% 2",
			Expected: []Request{
				{Path: "/bot/sendMessage", Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&text=Did+you+know+that+in+a+private+chat%2C+you+can+just+send+a+bus+stop+code+directly%2C+without+using+the+%2Feta+command%3F"},
				{Path: "/bot/sendMessage", Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&text=Oops%2C+that+did+not+seem+to+be+a+valid+bus+stop+code."},
			},
		},
		{
			Name: "Too long bus stop code",
			Text: "/eta 960499 2",
			Expected: []Request{
				{Path: "/bot/sendMessage", Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&text=Did+you+know+that+in+a+private+chat%2C+you+can+just+send+a+bus+stop+code+directly%2C+without+using+the+%2Feta+command%3F"},
				{Path: "/bot/sendMessage", Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&text=Oops%2C+a+bus+stop+code+can+only+contain+a+maximum+of+5+characters."},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			message := MockMessageWithType(ChatTypePrivate)
			message.Text = tc.Text

			err := EtaHandler(ctx, &bot, &message)
			if err != nil {
				t.Fatalf("%v", err)
			}

			var reqs []Request
			timer := time.NewTimer(5 * time.Second)
			for len(reqs) < 2 {
				select {
				case req := <-reqChan:
					reqs = append(reqs, req)
				case err := <-errChan:
					t.Fatalf("%v", err)
				case <-timer.C:
					t.Fatal("timed out")
				}
			}

			actual := reqs
			expected := tc.Expected

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
		})
	}
}
