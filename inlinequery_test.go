package busetabot

import (
	"context"
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
									{
										{
											Text:         "Show incoming bus details",
											CallbackData: "{\"t\":\"refresh\",\"b\":\"96041\",\"f\":\"f\"}",
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
									{
										{
											Text:         "Show incoming bus details",
											CallbackData: "{\"t\":\"refresh\",\"b\":\"96049\",\"f\":\"f\"}",
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
									{
										{
											Text:         "Show incoming bus details",
											CallbackData: "{\"t\":\"refresh\",\"b\":\"96049\",\"f\":\"f\"}",
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
									{
										{
											Text:         "Show incoming bus details",
											CallbackData: "{\"t\":\"refresh\",\"b\":\"96041\",\"f\":\"f\"}",
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
									{
										{
											Text:         "Show incoming bus details",
											CallbackData: "{\"t\":\"refresh\",\"b\":\"96041\",\"f\":\"f\"}",
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
									{
										{
											Text:         "Show incoming bus details",
											CallbackData: "{\"t\":\"refresh\",\"b\":\"96049\",\"f\":\"f\"}",
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
	busStops := mockBusStopRepository{
		BusStop: &BusStop{
			BusStopCode: "96049",
			RoadName:    "Upp Changi Rd East",
			Description: "Opp Tropicana Condo",
		},
	}
	testCases := []struct {
		Name     string
		ResultID string
		Expected []telegram.Request
	}{
		{
			Name:     "Normal chosen inline result",
			ResultID: "96049",
			Expected: []telegram.Request{
				telegram.EditMessageTextRequest{
					InlineMessageID: "ID",
					Text:            "*Opp Tropicana Condo (96049)*\nUpp Changi Rd East\n```\n| Svc | Next |  2nd |  3rd |\n|-----|------|------|------|\n| 2   |   -1 |   10 |   36 |\n| 24  |    1 |    3 |    6 |```\nShowing 2 out of 2 services for this bus stop.\n\n_Last updated at 01 Jan 01 08:00 SGT_",
					ParseMode:       "markdown",
					ReplyMarkup: telegram.InlineKeyboardMarkup{
						InlineKeyboard: [][]telegram.InlineKeyboardButton{
							{
								{
									Text:         "Refresh",
									CallbackData: "{\"t\":\"refresh\",\"b\":\"96049\"}",
								},
							},
							{
								{
									Text:         "Show incoming bus details",
									CallbackData: "{\"t\":\"refresh\",\"b\":\"96049\",\"f\":\"f\"}",
								},
							},
						},
					},
				},
			},
		},
		{
			Name:     "Nearby chosen inline result",
			ResultID: "96049 geo",
			Expected: []telegram.Request{
				telegram.EditMessageTextRequest{
					InlineMessageID: "ID",
					Text:            "*Opp Tropicana Condo (96049)*\nUpp Changi Rd East\n```\n| Svc | Next |  2nd |  3rd |\n|-----|------|------|------|\n| 2   |   -1 |   10 |   36 |\n| 24  |    1 |    3 |    6 |```\nShowing 2 out of 2 services for this bus stop.\n\n_Last updated at 01 Jan 01 08:00 SGT_",
					ParseMode:       "markdown",
					ReplyMarkup: telegram.InlineKeyboardMarkup{
						InlineKeyboard: [][]telegram.InlineKeyboardButton{
							{
								{
									Text:         "Refresh",
									CallbackData: "{\"t\":\"refresh\",\"b\":\"96049\"}",
								},
							},
							{
								{
									Text:         "Show incoming bus details",
									CallbackData: "{\"t\":\"refresh\",\"b\":\"96049\",\"f\":\"f\"}",
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			cir := tgbotapi.ChosenInlineResult{
				InlineMessageID: "ID",
				ResultID:        tc.ResultID,
				From: &tgbotapi.User{
					ID:        1,
					FirstName: "Jiayu",
				},
			}
			tg := new(mockTelegramService)
			bot := &BusEtaBot{
				Datamall: mockDatamall{},
				BusStops: busStops,
				NowFunc: func() time.Time {
					return time.Time{}
				},
				TelegramService: tg,
			}
			err := ChosenInlineResultHandler(context.Background(), bot, &cir)
			if err != nil {
				t.Fatal(err)
			}
			if !assert.Equal(t, tc.Expected, tg.Requests) {
				pretty.Println(tg.Requests)
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
					{
						{
							Text:         "Show incoming bus details",
							CallbackData: "{\"t\":\"refresh\",\"b\":\"96049\",\"f\":\"f\"}",
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
					{
						{
							Text:         "Show incoming bus details",
							CallbackData: "{\"t\":\"refresh\",\"b\":\"96041\",\"f\":\"f\"}",
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
