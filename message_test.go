package busetabot

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/kr/pretty"
	"github.com/stretchr/testify/assert"
	"github.com/yi-jiayu/telegram-bot-api"

	"github.com/yi-jiayu/bus-eta-bot/v4/telegram"
)

func TestLocationHandler(t *testing.T) {
	tgAPI, reqChan, errChan := NewMockTelegramAPIWithPath()
	defer tgAPI.Close()

	tg := &tgbotapi.BotAPI{
		APIEndpoint: tgAPI.URL + "/bot%s/%s",
		Client:      http.DefaultClient,
	}

	bot := NewBot(handlers, tg, nil, nil, nil)
	bot.BusStops = &MockBusStops{
		NearbyBusStops: []BusStop{
			{
				BusStopCode: "96041",
				RoadName:    "Upp Changi Rd East",
				Description: "Bef Tropicana Condo",
				Latitude:    1.34041450268626,
				Longitude:   103.96127892061004,
			},
			{
				BusStopCode: "96049",
				RoadName:    "Upp Changi Rd East",
				Description: "Opp Tropicana Condo",
				Latitude:    1.33995375346513,
				Longitude:   103.96079768187379,
			},
		},
	}

	message := MockMessage()
	message.Location = &tgbotapi.Location{
		Latitude:  1.34041450268626,
		Longitude: 103.96127892061004,
	}

	err := LocationHandler(context.Background(), &bot, &message)
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
		{Path: "/bot/sendVenue", Body: "address=0+m+away&chat_id=1&disable_notification=false&latitude=1.340415&longitude=103.961279&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Get+etas%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22new_eta%5C%22%2C%5C%22b%5C%22%3A%5C%2296041%5C%22%7D%22%7D%5D%5D%7D&title=Bef+Tropicana+Condo+%2896041%29"},
		{Path: "/bot/sendVenue", Body: "address=74+m+away&chat_id=1&disable_notification=false&latitude=1.339954&longitude=103.960798&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Get+etas%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22new_eta%5C%22%2C%5C%22b%5C%22%3A%5C%2296049%5C%22%7D%22%7D%5D%5D%7D&title=Opp+Tropicana+Condo+%2896049%29"},
	}

	assert.Equal(t, expected, actual)
}

func TestLocationHandlerNothingNearby(t *testing.T) {
	tgAPI, reqChan, errChan := NewMockTelegramAPIWithPath()
	defer tgAPI.Close()

	tg := &tgbotapi.BotAPI{
		APIEndpoint: tgAPI.URL + "/bot%s/%s",
		Client:      http.DefaultClient,
	}

	bot := NewBot(handlers, tg, nil, nil, nil)
	bot.BusStops = &MockBusStops{
		NearbyBusStops: make([]BusStop, 0),
	}

	message := MockMessage()
	message.Location = &tgbotapi.Location{
		Latitude:  1.34041450268626,
		Longitude: 103.96127892061004,
	}

	err := LocationHandler(context.Background(), &bot, &message)
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

func TestTextHandler(t *testing.T) {
	busStops := MockBusStops{
		BusStop: &BusStop{
			BusStopCode: "96049",
			RoadName:    "Upp Changi Rd East",
			Description: "Opp Tropicana Condo",
		},
	}
	type testCase struct {
		Name     string
		Message  *tgbotapi.Message
		Expected []telegram.Request
	}
	testCases := []testCase{
		{
			Name:     "should ignore inline queries from itself",
			Message:  MockMessageWithText("Fetching etas..."),
			Expected: nil,
		},
		{
			Name:    "should respond with eta message when message contains a bus stop code",
			Message: MockMessageWithText("96049"),
			Expected: []telegram.Request{
				telegram.SendMessageRequest{
					ChatID:    1,
					Text:      "*Opp Tropicana Condo (96049)*\nUpp Changi Rd East\n```\n| Svc | Next |  2nd |  3rd |\n|-----|------|------|------|\n| 2   |   -1 |   10 |   36 |\n| 24  |    1 |    3 |    6 |```\nShowing 2 out of 2 services for this bus stop.\n\n_Last updated at 01 Jan 01 08:00 SGT_",
					ParseMode: "markdown",
					ReplyMarkup: telegram.InlineKeyboardMarkup{
						InlineKeyboard: [][]telegram.InlineKeyboardButton{
							{
								{
									Text:         "Refresh",
									CallbackData: "{\"t\":\"refresh\",\"b\":\"96049\"}",
								},
								{
									Text:         "Resend",
									CallbackData: "{\"t\":\"resend\",\"b\":\"96049\"}",
								},
								{
									Text:         "⭐",
									CallbackData: "{\"t\":\"togf\",\"a\":\"96049\"}",
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "should validate eta message when message is a reply to a message asking for a bus stop code",
			Message: func() *tgbotapi.Message {
				reply := MockMessageWithText("Alright, send me a bus stop code to get etas for.")
				message := MockMessageWithText("jiayu")
				message.ReplyToMessage = reply
				return message
			}(),
			Expected: []telegram.Request{
				telegram.SendMessageRequest{
					ChatID:           1,
					Text:             "Oops, a bus stop code should be a 5-digit number.",
					ParseMode:        "",
					ReplyToMessageID: 0,
					ReplyMarkup:      nil,
				},
			},
		},
		{
			Name:     "should ignore messages that are not replies which do not contain a bus stop code",
			Message:  MockMessageWithText("jiayu"),
			Expected: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tg := &mockTelegramService{}
			bot := &BusEtaBot{
				Datamall: MockDatamall{},
				BusStops: busStops,
				NowFunc: func() time.Time {
					return time.Time{}
				},
				TelegramService: tg,
			}
			err := TextHandler(context.Background(), bot, tc.Message)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			if actual := tg.Requests; !assert.Equal(t, tc.Expected, actual) {
				pretty.Println(actual)
			}
		})
	}
}
