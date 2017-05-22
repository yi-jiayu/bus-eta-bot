package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/yi-jiayu/telegram-bot-api"
	"golang.org/x/net/context"
)

type Request struct {
	Path string
	Body string
}

type Spy struct {
	Called  bool
	SpyFunc func()
}

func (s *Spy) MessageHandler(context.Context, *tgbotapi.BotAPI, *tgbotapi.Message) error {
	if s.SpyFunc != nil {
		s.SpyFunc()
	}

	s.Called = true
	return nil
}

func (s *Spy) CallbackQueryHandler(ctx context.Context, bot *tgbotapi.BotAPI, cbq *tgbotapi.CallbackQuery) error {
	if s.SpyFunc != nil {
		s.SpyFunc()
	}

	s.Called = true
	return nil
}

func (s *Spy) InlineQueryHandler(ctx context.Context, bot *tgbotapi.BotAPI, ilq *tgbotapi.InlineQuery) error {
	if s.SpyFunc != nil {
		s.SpyFunc()
	}

	s.Called = true
	return nil
}

func (s *Spy) ChosenInlineResultHandler(ctx context.Context, bot *tgbotapi.BotAPI, cir *tgbotapi.ChosenInlineResult) error {
	if s.SpyFunc != nil {
		s.SpyFunc()
	}

	s.Called = true
	return nil
}

func NewMockTelegramAPIWithPath() (*httptest.Server, chan Request, chan error) {
	reqChan := make(chan Request, 2)
	errChan := make(chan error, 2)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				errChan <- err
				return
			}

			reqChan <- Request{r.URL.Path, string(body)}
		}

		w.Write([]byte(`{"ok":true}`))
	}))

	return ts, reqChan, errChan
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

func MockMessage() tgbotapi.Message {
	return tgbotapi.Message{
		Chat: &tgbotapi.Chat{ID: 1},
		From: &tgbotapi.User{ID: 1, FirstName: "Jiayu"},
	}
}

func MockCallbackQuery() tgbotapi.CallbackQuery {
	return tgbotapi.CallbackQuery{
		ID: "1",
		From: &tgbotapi.User{
			ID:        1,
			FirstName: "Jiayu",
		},
	}
}

func MockInlineQuery() tgbotapi.InlineQuery {
	return tgbotapi.InlineQuery{
		ID: "1",
		From: &tgbotapi.User{
			ID:        1,
			FirstName: "Jiayu",
		},
	}
}

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

func TestBusEtaBot_HandleUpdate(t *testing.T) {
	t.Parallel()

	// commands
	startCmdSpy := Spy{}
	aboutCmdSpy := Spy{}
	versionCmdSpy := Spy{}
	helpCmdSpy := Spy{}
	privacyCmdSpy := Spy{}
	etaCmdSpy := Spy{}

	// text handler
	textHandlerSpy := Spy{}

	// location handler
	locationHandlerSpy := Spy{}

	// callback query handlers
	refreshCbqSpy := Spy{}

	// inline query handler
	ilqHandlerSpy := Spy{}

	// choseninlineresulthandler
	cirHandlerSpy := Spy{}

	busEtaBot := BusEtaBot{
		CommandHandlers: map[string]MessageHandler{
			"start":   startCmdSpy.MessageHandler,
			"about":   aboutCmdSpy.MessageHandler,
			"version": versionCmdSpy.MessageHandler,
			"help":    helpCmdSpy.MessageHandler,
			"privacy": privacyCmdSpy.MessageHandler,
			"eta":     etaCmdSpy.MessageHandler,
		},
		TextHandler:     textHandlerSpy.MessageHandler,
		LocationHandler: locationHandlerSpy.MessageHandler,
		CallbackQueryHandlers: map[string]CallbackQueryHandler{
			"refresh": refreshCbqSpy.CallbackQueryHandler,
		},
		InlineQueryHandler:        ilqHandlerSpy.InlineQueryHandler,
		ChosenInlineResultHandler: cirHandlerSpy.ChosenInlineResultHandler,
	}

	testCases := []struct {
		Name   string
		Spy    *Spy
		Update *tgbotapi.Update
	}{
		{
			Name: "start command",
			Spy:  &startCmdSpy,
			Update: &tgbotapi.Update{
				Message: &tgbotapi.Message{
					From: &tgbotapi.User{
						ID:        1,
						FirstName: "Jiayu",
					},
					Chat: &tgbotapi.Chat{
						ID:   1,
						Type: "private",
					},
					Text: "/start",
				},
			},
		},
		{
			Name: "Text message",
			Spy:  &textHandlerSpy,
			Update: &tgbotapi.Update{
				Message: &tgbotapi.Message{
					From: &tgbotapi.User{
						ID:        1,
						FirstName: "Jiayu",
					},
					Chat: &tgbotapi.Chat{
						ID:   1,
						Type: "private",
					},
					Text: "96049",
				},
			},
		},
		{
			Name: "Message with location",
			Spy:  &locationHandlerSpy,
			Update: &tgbotapi.Update{
				Message: &tgbotapi.Message{
					From: &tgbotapi.User{
						ID:        1,
						FirstName: "Jiayu",
					},
					Chat: &tgbotapi.Chat{
						ID:   1,
						Type: "private",
					},
					Location: &tgbotapi.Location{
						Latitude:  1.3406375,
						Longitude: 103.9613357,
					},
				},
			},
		},
		{
			Name: "Refresh callback query",
			Spy:  &refreshCbqSpy,
			Update: &tgbotapi.Update{
				CallbackQuery: &tgbotapi.CallbackQuery{
					Data: `{"t":"refresh"}`,
				},
			},
		},
		{
			Name: "Inline query",
			Spy:  &ilqHandlerSpy,
			Update: &tgbotapi.Update{
				InlineQuery: &tgbotapi.InlineQuery{},
			},
		},
		{
			Name: "Chosen inline result",
			Spy:  &cirHandlerSpy,
			Update: &tgbotapi.Update{
				ChosenInlineResult: &tgbotapi.ChosenInlineResult{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			busEtaBot.HandleUpdate(nil, nil, tc.Update)

			if !tc.Spy.Called {
				t.Fail()
			}
		})
	}
}
