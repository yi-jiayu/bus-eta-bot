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

type mockStreetView string

func (s mockStreetView) GetPhotoURLByLocation(lat, lon float64, width, height int) (string, error) {
	return string(s), nil
}

type mockTelegramService struct {
	Requests []telegram.Request
}

func (s *mockTelegramService) Do(request telegram.Request) error {
	s.Requests = append(s.Requests, request)
	return nil
}

func TestInlineQueryHandler(t *testing.T) {
	busStops := []BusStop{
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
	}

	bot := BusEtaBot{
		StreetView: NewStreetViewAPI("API_KEY"),
		BusStops:   NewInMemoryBusStopRepository(busStops, nil),
	}

	testCases := []struct {
		Name             string
		Query            string
		Location         *tgbotapi.Location
		ExpectedRequests []telegram.Request
	}{
		{
			Name:  "Empty query",
			Query: "",
			ExpectedRequests: []telegram.Request{
				telegram.AnswerInlineQueryRequest{
					InlineQueryID: "1",
					Results: []telegram.InlineQueryResult{
						telegram.InlineQueryResultArticle{
							ID:                  "96041",
							Title:               "Bef Tropicana Condo (96041)",
							Description:         "Upp Changi Rd East",
							ThumbURL:            "https://maps.googleapis.com/maps/api/streetview?key=API_KEY&location=1.340415%2C103.961279&size=100x100",
							InputMessageContent: telegram.InputTextMessageContent{MessageText: "*Bef Tropicana Condo (96041)*\nUpp Changi Rd East\n`Fetching etas...`", ParseMode: "markdown"},
							ReplyMarkup: telegram.InlineKeyboardMarkup{
								InlineKeyboard: [][]telegram.InlineKeyboardButton{
									{
										{
											Text:         "Refresh",
											CallbackData: "{\"t\":\"refresh\",\"b\":\"96041\"}",
										},
									},
								},
							},
						},
						telegram.InlineQueryResultArticle{
							ID:                  "96049",
							Title:               "Opp Tropicana Condo (96049)",
							Description:         "Upp Changi Rd East",
							ThumbURL:            "https://maps.googleapis.com/maps/api/streetview?key=API_KEY&location=1.339954%2C103.960798&size=100x100",
							InputMessageContent: telegram.InputTextMessageContent{MessageText: "*Opp Tropicana Condo (96049)*\nUpp Changi Rd East\n`Fetching etas...`", ParseMode: "markdown"},
							ReplyMarkup: telegram.InlineKeyboardMarkup{
								InlineKeyboard: [][]telegram.InlineKeyboardButton{
									{
										{
											Text:         "Refresh",
											CallbackData: "{\"t\":\"refresh\",\"b\":\"96049\"}",
										},
									},
								},
							},
						},
					},
					CacheTime:  0,
					IsPersonal: false,
				},
			},
		},
		{
			Name:  "Empty query with location",
			Query: "",
			Location: &tgbotapi.Location{
				Latitude:  1.340,
				Longitude: 103.961,
			},
			ExpectedRequests: []telegram.Request{
				telegram.AnswerInlineQueryRequest{
					InlineQueryID: "1",
					Results: []telegram.InlineQueryResult{
						telegram.InlineQueryResultArticle{
							ID:                  "96049 geo",
							Title:               "Opp Tropicana Condo (96049)",
							Description:         "23 m away",
							ThumbURL:            "https://maps.googleapis.com/maps/api/streetview?key=API_KEY&location=1.339954%2C103.960798&size=100x100",
							InputMessageContent: telegram.InputTextMessageContent{MessageText: "*Opp Tropicana Condo (96049)*\nUpp Changi Rd East\n`Fetching etas...`", ParseMode: "markdown"},
							ReplyMarkup: telegram.InlineKeyboardMarkup{
								InlineKeyboard: [][]telegram.InlineKeyboardButton{
									{
										{
											Text:         "Refresh",
											CallbackData: "{\"t\":\"refresh\",\"b\":\"96049\"}",
										},
									},
								},
							},
						},
						telegram.InlineQueryResultArticle{
							ID:                  "96041 geo",
							Title:               "Bef Tropicana Condo (96041)",
							Description:         "55 m away",
							ThumbURL:            "https://maps.googleapis.com/maps/api/streetview?key=API_KEY&location=1.340415%2C103.961279&size=100x100",
							InputMessageContent: telegram.InputTextMessageContent{MessageText: "*Bef Tropicana Condo (96041)*\nUpp Changi Rd East\n`Fetching etas...`", ParseMode: "markdown"},
							ReplyMarkup: telegram.InlineKeyboardMarkup{
								InlineKeyboard: [][]telegram.InlineKeyboardButton{
									{
										{
											Text:         "Refresh",
											CallbackData: "{\"t\":\"refresh\",\"b\":\"96041\"}",
										},
									},
								},
							},
						},
					},
					CacheTime:  0,
					IsPersonal: false,
				},
			},
		},
		{
			Name:  "Matching query",
			Query: "tropicana",
			ExpectedRequests: []telegram.Request{
				telegram.AnswerInlineQueryRequest{
					InlineQueryID: "1",
					Results: []telegram.InlineQueryResult{
						telegram.InlineQueryResultArticle{
							ID:                  "96041",
							Title:               "Bef Tropicana Condo (96041)",
							Description:         "Upp Changi Rd East",
							ThumbURL:            "https://maps.googleapis.com/maps/api/streetview?key=API_KEY&location=1.340415%2C103.961279&size=100x100",
							InputMessageContent: telegram.InputTextMessageContent{MessageText: "*Bef Tropicana Condo (96041)*\nUpp Changi Rd East\n`Fetching etas...`", ParseMode: "markdown"},
							ReplyMarkup: telegram.InlineKeyboardMarkup{
								InlineKeyboard: [][]telegram.InlineKeyboardButton{
									{
										{
											Text:         "Refresh",
											CallbackData: "{\"t\":\"refresh\",\"b\":\"96041\"}",
										},
									},
								},
							},
						},
						telegram.InlineQueryResultArticle{
							ID:                  "96049",
							Title:               "Opp Tropicana Condo (96049)",
							Description:         "Upp Changi Rd East",
							ThumbURL:            "https://maps.googleapis.com/maps/api/streetview?key=API_KEY&location=1.339954%2C103.960798&size=100x100",
							InputMessageContent: telegram.InputTextMessageContent{MessageText: "*Opp Tropicana Condo (96049)*\nUpp Changi Rd East\n`Fetching etas...`", ParseMode: "markdown"},
							ReplyMarkup: telegram.InlineKeyboardMarkup{
								InlineKeyboard: [][]telegram.InlineKeyboardButton{
									{
										{
											Text:         "Refresh",
											CallbackData: "{\"t\":\"refresh\",\"b\":\"96049\"}",
										},
									},
								},
							},
						},
					},
					CacheTime:  86400,
					IsPersonal: false,
				},
			},
		},
		{
			Name:  "Non-matching query",
			Query: "anaciport",
			ExpectedRequests: []telegram.Request{
				telegram.AnswerInlineQueryRequest{
					InlineQueryID: "1",
					Results:       []telegram.InlineQueryResult{},
					CacheTime:     86400,
					IsPersonal:    false,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tg := &mockTelegramService{}
			bot.TelegramService = tg

			ilq := MockInlineQuery()
			ilq.Query = tc.Query
			ilq.Location = tc.Location

			err := InlineQueryHandler(context.Background(), &bot, &ilq)
			if err != nil {
				t.Fatal(err)
			}

			if !assert.ElementsMatch(t, tc.ExpectedRequests, tg.Requests) {
				pretty.Println(tg.Requests)
			}
		})
	}
}

func TestChosenInlineResultHandler(t *testing.T) {
	tgAPI, reqChan, errChan := NewMockTelegramAPIWithPath()
	defer tgAPI.Close()

	tg := &tgbotapi.BotAPI{
		APIEndpoint: tgAPI.URL + "/bot%s/%s",
		Client:      http.DefaultClient,
	}

	busStops := MockBusStops{
		BusStop: &BusStop{
			BusStopCode: "96049",
			RoadName:    "Upp Changi Rd East",
			Description: "Opp Tropicana Condo",
		},
	}

	bot := BusEtaBot{
		Telegram: tg,
		Datamall: MockDatamall{},
		BusStops: busStops,
		NowFunc: func() time.Time {
			return time.Time{}
		},
	}

	testCases := []struct {
		Name     string
		ResultID string
		Expected Request
	}{
		{
			Name:     "Normal chosen inline result",
			ResultID: "96049",
			Expected: Request{
				Path: "/bot/editMessageText",
				Body: "chat_id=0&disable_web_page_preview=false&message_id=0&parse_mode=markdown&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A%5C%2296049%5C%22%7D%22%7D%5D%5D%7D&text=%2AOpp+Tropicana+Condo+%2896049%29%2A%0AUpp+Changi+Rd+East%0A%60%60%60%0A%7C+Svc+%7C+Next+%7C++2nd+%7C++3rd+%7C%0A%7C-----%7C------%7C------%7C------%7C%0A%7C+2+++%7C+++-1+%7C+++10+%7C+++36+%7C%0A%7C+24++%7C++++1+%7C++++3+%7C++++6+%7C%60%60%60%0AShowing+2+out+of+2+services+for+this+bus+stop.%0A%0A_Last+updated+at+01+Jan+01+08%3A00+SGT_",
			},
		},
		{
			Name:     "Nearby chosen inline result",
			ResultID: "96049 geo",
			Expected: Request{
				Path: "/bot/editMessageText",
				Body: "chat_id=0&disable_web_page_preview=false&message_id=0&parse_mode=markdown&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A%5C%2296049%5C%22%7D%22%7D%5D%5D%7D&text=%2AOpp+Tropicana+Condo+%2896049%29%2A%0AUpp+Changi+Rd+East%0A%60%60%60%0A%7C+Svc+%7C+Next+%7C++2nd+%7C++3rd+%7C%0A%7C-----%7C------%7C------%7C------%7C%0A%7C+2+++%7C+++-1+%7C+++10+%7C+++36+%7C%0A%7C+24++%7C++++1+%7C++++3+%7C++++6+%7C%60%60%60%0AShowing+2+out+of+2+services+for+this+bus+stop.%0A%0A_Last+updated+at+01+Jan+01+08%3A00+SGT_",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			cir := tgbotapi.ChosenInlineResult{
				ResultID: tc.ResultID,
				From: &tgbotapi.User{
					ID:        1,
					FirstName: "Jiayu",
				},
			}

			err := ChosenInlineResultHandler(context.Background(), &bot, &cir)
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

func TestGetNearbyInlineQueryResults(t *testing.T) {
	streetView := mockStreetView("URL")
	busStops := NewInMemoryBusStopRepository([]BusStop{
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
	}, nil)
	expected := []telegram.InlineQueryResult{
		telegram.InlineQueryResultArticle{
			ID:                  "96049 geo",
			Title:               "Opp Tropicana Condo (96049)",
			Description:         "23 m away",
			ThumbURL:            "URL",
			InputMessageContent: telegram.InputTextMessageContent{MessageText: "*Opp Tropicana Condo (96049)*\nUpp Changi Rd East\n`Fetching etas...`", ParseMode: "markdown"},
			ReplyMarkup: telegram.InlineKeyboardMarkup{
				InlineKeyboard: [][]telegram.InlineKeyboardButton{
					{
						{
							Text:         "Refresh",
							CallbackData: "{\"t\":\"refresh\",\"b\":\"96049\"}",
						},
					},
				},
			},
		},
		telegram.InlineQueryResultArticle{
			ID:                  "96041 geo",
			Title:               "Bef Tropicana Condo (96041)",
			Description:         "55 m away",
			ThumbURL:            "URL",
			InputMessageContent: telegram.InputTextMessageContent{MessageText: "*Bef Tropicana Condo (96041)*\nUpp Changi Rd East\n`Fetching etas...`", ParseMode: "markdown"},
			ReplyMarkup: telegram.InlineKeyboardMarkup{
				InlineKeyboard: [][]telegram.InlineKeyboardButton{
					{
						{
							Text:         "Refresh",
							CallbackData: "{\"t\":\"refresh\",\"b\":\"96041\"}",
						},
					},
				},
			},
		},
	}
	actual, err := GetNearbyInlineQueryResults(context.Background(), streetView, busStops, 1.340, 103.961)
	if err != nil {
		t.Fatal(err)
	}
	if !assert.Equal(t, expected, actual) {
		pretty.Println(actual)
	}
}
