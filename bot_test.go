package busetabot

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/yi-jiayu/telegram-bot-api"

	"github.com/yi-jiayu/bus-eta-bot/v4/mocks"
)

// Telegram chat types
const (
	ChatTypePrivate    = "private"
	ChatTypeGroup      = "group"
	ChatTypeSupergroup = "supergroup"
	ChatTypeChannel    = "channel"
)

type Request struct {
	Path string
	Body string
}

type Spy struct {
	Called  bool
	SpyFunc func()
}

func (s *Spy) MessageHandler(context.Context, *BusEtaBot, *tgbotapi.Message) error {
	if s.SpyFunc != nil {
		s.SpyFunc()
	}

	s.Called = true
	return nil
}

func (s *Spy) CommandHandler(ctx context.Context, bot *BusEtaBot, msg *tgbotapi.Message, responses chan<- Response) {
	defer close(responses)
	if s.SpyFunc != nil {
		s.SpyFunc()
	}
	s.Called = true
}

func (s *Spy) CallbackQueryHandler(ctx context.Context, bot *BusEtaBot, cbq *tgbotapi.CallbackQuery, responses chan<- Response) {
	defer close(responses)
	if s.SpyFunc != nil {
		s.SpyFunc()
	}
	s.Called = true
}

func (s *Spy) InlineQueryHandler(ctx context.Context, bot *BusEtaBot, ilq *tgbotapi.InlineQuery) error {
	if s.SpyFunc != nil {
		s.SpyFunc()
	}

	s.Called = true
	return nil
}

func (s *Spy) ChosenInlineResultHandler(ctx context.Context, bot *BusEtaBot, cir *tgbotapi.ChosenInlineResult) error {
	if s.SpyFunc != nil {
		s.SpyFunc()
	}

	s.Called = true
	return nil
}

func NewMockTelegramAPIWithPath() (*httptest.Server, chan Request, chan error) {
	reqChan := make(chan Request, 1)
	errChan := make(chan error, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				errChan <- err
				return
			}

			go func() {
				reqChan <- Request{Path: r.URL.Path, Body: string(body)}
			}()
		}

		w.Write([]byte(`{"ok":true}`))
	}))

	return ts, reqChan, errChan
}

func MockMessage() tgbotapi.Message {
	return tgbotapi.Message{
		Chat: &tgbotapi.Chat{ID: 1},
		From: &tgbotapi.User{ID: 1, FirstName: "Jiayu"},
	}
}

func MockMessageWithType(chatType string) tgbotapi.Message {
	return tgbotapi.Message{
		MessageID: 1,
		Chat:      &tgbotapi.Chat{ID: 1, Type: chatType},
		From:      &tgbotapi.User{ID: 1, FirstName: "Jiayu"},
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

func TestBusEtaBot_HandleUpdate(t *testing.T) {
	// commands
	startCmdSpy := Spy{}
	aboutCmdSpy := Spy{}
	versionCmdSpy := Spy{}
	helpCmdSpy := Spy{}
	privacyCmdSpy := Spy{}
	etaCmdSpy := Spy{}

	// command fallback
	fallbackCmdSpy := Spy{}

	// text handler
	textHandlerSpy := Spy{}

	// location handler
	locationHandlerSpy := Spy{}

	// callback query Handlers
	refreshCbqSpy := Spy{}

	// inline query handler
	ilqHandlerSpy := Spy{}

	// choseninlineresulthandler
	cirHandlerSpy := Spy{}

	busEtaBot := BusEtaBot{
		Handlers: Handlers{
			CommandHandlers: map[string]CommandHandler{
				"start":   startCmdSpy.CommandHandler,
				"about":   aboutCmdSpy.CommandHandler,
				"version": versionCmdSpy.CommandHandler,
				"help":    helpCmdSpy.CommandHandler,
				"privacy": privacyCmdSpy.CommandHandler,
				"eta":     etaCmdSpy.CommandHandler,
			},
			FallbackCommandHandler: fallbackCmdSpy.MessageHandler,
			TextHandler:            textHandlerSpy.MessageHandler,
			LocationHandler:        locationHandlerSpy.MessageHandler,
			CallbackQueryHandlers: map[string]CallbackQueryHandler{
				"refresh": refreshCbqSpy.CallbackQueryHandler,
			},
			InlineQueryHandler:        ilqHandlerSpy.InlineQueryHandler,
			ChosenInlineResultHandler: cirHandlerSpy.ChosenInlineResultHandler,
		},
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
			busEtaBot.HandleUpdate(nil, tc.Update)

			if !tc.Spy.Called {
				t.Fail()
			}
		})
	}
}

type MockUserRepository struct {
	Called bool
}

func (r *MockUserRepository) UpdateUserLastSeenTime(ctx context.Context, userID int, t time.Time) error {
	r.Called = true
	return nil
}

func TestBusEtaBot_HandleUpdate_UpdateUserLastSeenTime(t *testing.T) {
	const userID = 1
	testCases := []struct {
		Name   string
		Update *tgbotapi.Update
	}{
		{
			Name: "message",
			Update: &tgbotapi.Update{
				Message: &tgbotapi.Message{
					From: &tgbotapi.User{
						ID:        userID,
						FirstName: "Jiayu",
					},
				},
			},
		},
		{
			Name: "callback query",
			Update: &tgbotapi.Update{
				CallbackQuery: &tgbotapi.CallbackQuery{
					From: &tgbotapi.User{
						ID:        userID,
						FirstName: "Jiayu",
					},
				},
			},
		},
		{
			Name: "inline query",
			Update: &tgbotapi.Update{
				InlineQuery: &tgbotapi.InlineQuery{
					From: &tgbotapi.User{
						ID:        userID,
						FirstName: "Jiayu",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := mocks.NewMockUserRepository(ctrl)
			m.EXPECT().UpdateUserLastSeenTime(gomock.Any(), userID, gomock.Any()).Times(1)
			bot := &BusEtaBot{
				Users: m,
			}
			bot.HandleUpdate(context.Background(), tc.Update)
		})
	}
}
