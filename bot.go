package main

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/yi-jiayu/datamall"
	"github.com/yi-jiayu/telegram-bot-api"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

var handlers = Handlers{
	CommandHandlers:           commandHandlers,
	TextHandler:               TextHandler,
	LocationHandler:           LocationHandler,
	CallbackQueryHandlers:     callbackQueryHandlers,
	InlineQueryHandler:        InlineQueryHandler,
	ChosenInlineResultHandler: ChosenInlineResultHandler,
	MessageErrorHandler:       messageErrorHandler,
	CallbackErrorHandler:      callbackErrorHandler,
}

// BusEtaBot contains all the bot's dependencies
type BusEtaBot struct {
	Handlers   Handlers
	Telegram   *tgbotapi.BotAPI
	Datamall   *datamall.APIClient
	GA         *GAClient
	StreetView *StreetViewAPI
	NowFunc    func() time.Time
}

// Handlers contains all the handlers used by the bot.
type Handlers struct {
	CommandHandlers           map[string]MessageHandler
	TextHandler               MessageHandler
	LocationHandler           MessageHandler
	CallbackQueryHandlers     map[string]CallbackQueryHandler
	InlineQueryHandler        func(ctx context.Context, bot *BusEtaBot, ilq *tgbotapi.InlineQuery) error
	ChosenInlineResultHandler func(ctx context.Context, bot *BusEtaBot, cir *tgbotapi.ChosenInlineResult) error
	MessageErrorHandler       func(ctx context.Context, bot *BusEtaBot, message *tgbotapi.Message, err error)
	CallbackErrorHandler      func(ctx context.Context, bot *BusEtaBot, query *tgbotapi.CallbackQuery, err error)
}

// NewBusEtaBot creates a new Bus Eta Bot with the provided tgbotapi.BotAPI and datamall.APIClient.
func NewBusEtaBot(handlers Handlers, tg *tgbotapi.BotAPI, dm *datamall.APIClient) BusEtaBot {
	bot := BusEtaBot{
		Handlers: handlers,
		Telegram: tg,
		Datamall: dm,
	}

	bot.NowFunc = time.Now

	return bot
}

// HandleUpdate dispatches an incoming update to the corresponding handler depending on the update type
func (b *BusEtaBot) HandleUpdate(ctx context.Context, update *tgbotapi.Update) {
	if message := update.Message; message != nil {
		if command := message.Command(); command != "" {
			// handle command
			if handler, exists := b.Handlers.CommandHandlers[command]; exists {
				err := handler(ctx, b, message)
				if err != nil {
					messageErrorHandler(ctx, b, message, err)
					return
				}
			}

			return
		}

		if text := message.Text; text != "" {
			err := b.Handlers.TextHandler(ctx, b, message)
			if err != nil {
				messageErrorHandler(ctx, b, message, err)
				return
			}

			return
		}

		if location := message.Location; location != nil {
			err := b.Handlers.LocationHandler(ctx, b, message)
			if err != nil {
				messageErrorHandler(ctx, b, message, err)
				return
			}

			return
		}

		return
	}

	if cbq := update.CallbackQuery; cbq != nil {
		var data map[string]interface{}
		err := json.Unmarshal([]byte(cbq.Data), &data)
		if err != nil {
			callbackErrorHandler(ctx, b, cbq, err)
			return
		}

		if cbqType, ok := data["t"].(string); ok {
			if handler, ok := b.Handlers.CallbackQueryHandlers[cbqType]; ok {
				err := handler(ctx, b, cbq)
				if err != nil {
					callbackErrorHandler(ctx, b, cbq, err)
					return
				}
			}
		} else {
			callbackErrorHandler(ctx, b, cbq, errors.New("unrecognised callback query"))
			return
		}

		return
	}

	if ilq := update.InlineQuery; ilq != nil {
		// handle inline query
		err := b.Handlers.InlineQueryHandler(ctx, b, ilq)
		if err != nil {
			log.Errorf(ctx, "%v", err)
			return
		}

		return
	}

	if cir := update.ChosenInlineResult; cir != nil {
		err := b.Handlers.ChosenInlineResultHandler(ctx, b, cir)
		if err != nil {
			log.Errorf(ctx, "%v", err)
			return
		}

		return
	}
}

// LogEvent logs an event to the Measurement Protocol if a GAClient is set on the bot.
func (b *BusEtaBot) LogEvent(ctx context.Context, user *tgbotapi.User, category, action, label string) {
	if b.GA != nil {
		_, err := b.GA.LogEvent(user.ID, user.LanguageCode, category, action, label)
		if err != nil {
			log.Errorf(ctx, "error while logging event: %v", err)
		}
	}
}
