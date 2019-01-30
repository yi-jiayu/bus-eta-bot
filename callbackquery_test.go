package busetabot

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/kr/pretty"
	"github.com/stretchr/testify/assert"
	"github.com/yi-jiayu/datamall/v2"
	"github.com/yi-jiayu/telegram-bot-api"

	"github.com/yi-jiayu/bus-eta-bot/v4/mocks"
	"github.com/yi-jiayu/bus-eta-bot/v4/telegram"
)

//go:generate mockgen -destination mocks/users.go -package mocks github.com/yi-jiayu/bus-eta-bot/v4 UserRepository

func newCallbackQueryFromMessage(data string) *tgbotapi.CallbackQuery {
	return &tgbotapi.CallbackQuery{
		ID: "1",
		From: &tgbotapi.User{
			ID:        1,
			FirstName: "Jiayu",
		},
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{
				ID: 1,
			},
			MessageID: 1,
		},
		Data: data,
	}
}

func newCallbackQueryFromInlineMessage(data string) *tgbotapi.CallbackQuery {
	return &tgbotapi.CallbackQuery{
		ID: "1",
		From: &tgbotapi.User{
			ID:        1,
			FirstName: "Jiayu",
		},
		InlineMessageID: "1",
		Data:            data,
	}
}

func TestRefreshCallbackHandler(t *testing.T) {
	now := func() (t time.Time) {
		return t
	}
	etaService := MockDatamall{}
	busStopRepository := MockBusStops{
		BusStop: &BusStop{
			BusStopCode: "96049",
			RoadName:    "Upp Changi Rd East",
			Description: "Opp Tropicana Condo",
		},
	}
	type testCase struct {
		Name          string
		CallbackQuery *tgbotapi.CallbackQuery
		ETAService    ETAService
		Expected      []Response
		ExpectError   bool
	}
	testCases := []testCase{
		{
			Name:          "for callback query from message",
			CallbackQuery: newCallbackQueryFromMessage(`{"t":"refresh","b":"96049"}`),
			Expected: []Response{
				ok(telegram.EditMessageTextRequest{
					ChatID:    1,
					MessageID: 1,
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
				}),
				ok(telegram.AnswerCallbackQueryRequest{CallbackQueryID: "1", Text: "ETAs updated!"}),
			},
		},
		{
			Name:          "for callback query from inline message",
			CallbackQuery: newCallbackQueryFromInlineMessage(`{"t":"refresh","b":"96049"}`),
			Expected: []Response{
				ok(telegram.EditMessageTextRequest{
					InlineMessageID: "1",
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
						},
					},
				}),
				ok(telegram.AnswerCallbackQueryRequest{CallbackQueryID: "1", Text: "ETAs updated!"}),
			},
		},
		{
			Name:          "when datamall returns an error",
			CallbackQuery: newCallbackQueryFromInlineMessage(`{"t":"refresh","b":"96049"}`),
			ETAService: MockETAService{
				Error: datamall.Error{StatusCode: 501},
			},
			Expected: []Response{
				ok(telegram.EditMessageTextRequest{
					InlineMessageID: "1",
					Text:            "*Opp Tropicana Condo (96049)*\nUpp Changi Rd East\n\nOh no! The LTA DataMall API that Bus Eta Bot relies on appears to be down at the moment (it returned HTTP status code 501).\n\n_Last updated at 01 Jan 01 08:00 SGT_",
					ParseMode:       "markdown",
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
				}),
				ok(telegram.AnswerCallbackQueryRequest{CallbackQueryID: "1", Text: "ETAs updated!"}),
			},
		},
		{
			Name:          "when ETAService returns some other error",
			CallbackQuery: newCallbackQueryFromInlineMessage(`{"t":"refresh","b":"96049"}`),
			ETAService: MockETAService{
				Error: errors.New("unexpected error"),
			},
			ExpectError: true,
		},
		{
			Name:          "when callback data is invalid",
			CallbackQuery: newCallbackQueryFromInlineMessage("invalid json"),
			ExpectError:   true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			bot := &BusEtaBot{
				Datamall: etaService,
				NowFunc:  now,
				BusStops: busStopRepository,
			}
			if tc.ETAService != nil {
				bot.Datamall = tc.ETAService
			}
			responses := make(chan Response, ResponseBufferSize)
			go RefreshCallbackHandler(context.TODO(), bot, tc.CallbackQuery, responses)
			actual, err := collectResponsesWithTimeout(responses, 5*time.Second)
			if err != nil {
				t.Fatal(err)
			}
			if tc.ExpectError {
				assert.Error(t, actual[0].Error)
				return
			}
			if !assert.Equal(t, tc.Expected, actual) {
				pretty.Println(actual)
			}
		})
	}
}

func TestEtaCallbackHandler(t *testing.T) {
	now := func() (t time.Time) {
		return t
	}
	etaService := MockDatamall{}
	busStopRepository := MockBusStops{
		BusStop: &BusStop{
			BusStopCode: "96049",
			RoadName:    "Upp Changi Rd East",
			Description: "Opp Tropicana Condo",
		},
	}
	type testCase struct {
		Name          string
		CallbackQuery *tgbotapi.CallbackQuery
		ETAService    ETAService
		Expected      []Response
		ExpectError   bool
	}
	testCases := []testCase{
		{
			Name:          "for callback query containing argstr",
			CallbackQuery: newCallbackQueryFromMessage(`{"t":"eta","a":"96049"}`),
			Expected: []Response{
				ok(telegram.EditMessageTextRequest{
					ChatID:    1,
					MessageID: 1,
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
				}),
				ok(telegram.AnswerCallbackQueryRequest{CallbackQueryID: "1", Text: "ETAs updated!"}),
			},
		},
		{
			Name:          "for callback query containing bus stop and services",
			CallbackQuery: newCallbackQueryFromInlineMessage(`{"t":"eta","b":"96049"}`),
			Expected: []Response{
				ok(telegram.EditMessageTextRequest{
					InlineMessageID: "1",
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
						},
					},
				}),
				ok(telegram.AnswerCallbackQueryRequest{CallbackQueryID: "1", Text: "ETAs updated!"}),
			},
		},
		{
			Name:          "when callback data is invalid",
			CallbackQuery: newCallbackQueryFromInlineMessage("invalid json"),
			ExpectError:   true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			bot := &BusEtaBot{
				Datamall: etaService,
				NowFunc:  now,
				BusStops: busStopRepository,
			}
			if tc.ETAService != nil {
				bot.Datamall = tc.ETAService
			}
			responses := make(chan Response, ResponseBufferSize)
			go EtaCallbackHandler(context.TODO(), bot, tc.CallbackQuery, responses)
			actual, err := collectResponsesWithTimeout(responses, 5*time.Second)
			if err != nil {
				t.Fatal(err)
			}
			if tc.ExpectError {
				assert.Error(t, actual[0].Error)
				return
			}
			if !assert.Equal(t, tc.Expected, actual) {
				pretty.Println(actual)
			}
		})
	}
}

func TestEtaDemoCallbackHandler(t *testing.T) {
	now := func() (t time.Time) {
		return t
	}
	etaService := MockDatamall{}
	busStopRepository := MockBusStops{
		BusStop: &BusStop{
			BusStopCode: "96049",
			RoadName:    "Upp Changi Rd East",
			Description: "Opp Tropicana Condo",
		},
	}
	bot := &BusEtaBot{
		Datamall: etaService,
		NowFunc:  now,
		BusStops: busStopRepository,
	}
	cbq := newCallbackQueryFromMessage("")
	responses := make(chan Response, ResponseBufferSize)
	go EtaDemoCallbackHandler(context.TODO(), bot, cbq, responses)
	actual, err := collectResponsesWithTimeout(responses, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	expected := []Response{
		ok(telegram.SendMessageRequest{
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
		}),
		ok(telegram.AnswerCallbackQueryRequest{CallbackQueryID: "1", Text: "ETAs sent!"}),
	}
	if !assert.Equal(t, expected, actual) {
		pretty.Println(actual)
	}
}

func TestNewEtaHandler(t *testing.T) {
	now := func() (t time.Time) {
		return t
	}
	etaService := MockDatamall{}
	busStopRepository := MockBusStops{
		BusStop: &BusStop{
			BusStopCode: "96049",
			RoadName:    "Upp Changi Rd East",
			Description: "Opp Tropicana Condo",
		},
	}
	bot := &BusEtaBot{
		Datamall: etaService,
		NowFunc:  now,
		BusStops: busStopRepository,
	}
	cbq := newCallbackQueryFromMessage(`{"t":"new_eta","b": "96049"}`)
	responses := make(chan Response, ResponseBufferSize)
	go NewEtaHandler(context.TODO(), bot, cbq, responses)
	actual, err := collectResponsesWithTimeout(responses, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	expected := []Response{
		ok(telegram.SendMessageRequest{
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
		}),
		ok(telegram.AnswerCallbackQueryRequest{CallbackQueryID: "1", Text: "ETAs sent!"}),
	}
	if !assert.Equal(t, expected, actual) {
		pretty.Println(actual)
	}
}

func TestToggleFavouritesHandler(t *testing.T) {
	const userID = 1
	type testCase struct {
		Name               string
		Favourites         []string
		BusStopCode        string
		ExpectedFavourites []string
		ExpectedResponses  []Response
	}
	testCases := []testCase{
		{
			Name:               "when toggling a new favourite",
			Favourites:         nil,
			BusStopCode:        "96049",
			ExpectedFavourites: []string{"96049"},
			ExpectedResponses: []Response{
				{
					Request: telegram.SendMessageRequest{
						ChatID:    1,
						Text:      "ETA query `96049` added to favourites!",
						ParseMode: "markdown",
						ReplyMarkup: telegram.ReplyKeyboardMarkup{
							Keyboard: [][]telegram.KeyboardButton{
								{
									{Text: "96049"},
								},
							},
							ResizeKeyboard: true,
						},
					},
				},
				{
					Request: telegram.AnswerCallbackQueryRequest{CallbackQueryID: "1"},
				},
			},
		},
		{
			Name:               "when toggling an existing favourite",
			Favourites:         []string{"96049", "81111"},
			BusStopCode:        "96049",
			ExpectedFavourites: []string{"81111"},
			ExpectedResponses: []Response{
				{
					Request: telegram.SendMessageRequest{
						ChatID:    1,
						Text:      "ETA query `96049` removed from favourites!",
						ParseMode: "markdown",
						ReplyMarkup: telegram.ReplyKeyboardMarkup{
							Keyboard: [][]telegram.KeyboardButton{
								{
									{Text: "81111"},
								},
							},
							ResizeKeyboard: true,
						},
					},
				},
				{
					Request: telegram.AnswerCallbackQueryRequest{CallbackQueryID: "1"},
				},
			},
		},
		{
			Name:               "when toggling the only favourite",
			Favourites:         []string{"96049"},
			BusStopCode:        "96049",
			ExpectedFavourites: []string{},
			ExpectedResponses: []Response{
				{
					Request: telegram.SendMessageRequest{
						ChatID:      1,
						Text:        "ETA query `96049` removed from favourites!",
						ParseMode:   "markdown",
						ReplyMarkup: telegram.ReplyKeyboardRemove{},
					},
				},
				{
					Request: telegram.AnswerCallbackQueryRequest{CallbackQueryID: "1"},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := mocks.NewMockUserRepository(ctrl)
			m.EXPECT().GetUserFavourites(gomock.Any(), userID).Return(tc.Favourites, nil)
			m.EXPECT().SetUserFavourites(gomock.Any(), userID, tc.ExpectedFavourites).Return(nil)
			bot := &BusEtaBot{
				Users: m,
				NowFunc: func() time.Time {
					return time.Time{}
				},
			}
			cbq := newCallbackQueryFromMessage(fmt.Sprintf(`{"t":"togf","a": "%s"}`, tc.BusStopCode))
			responses := make(chan Response, ResponseBufferSize)
			go ToggleFavouritesHandler(context.TODO(), bot, cbq, responses)
			actual, err := collectResponsesWithTimeout(responses, 5*time.Second)
			if err != nil {
				t.Fatal(err)
			}
			if !assert.Equal(t, tc.ExpectedResponses, actual) {
				pretty.Println(actual)
			}
		})
	}
}
