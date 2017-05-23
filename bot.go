package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/yi-jiayu/datamall"
	"github.com/yi-jiayu/telegram-bot-api"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

var illegalCharsRegex = regexp.MustCompile(`[^A-Z0-9 ]`)

var (
	gaTID = os.Getenv("GA_TID")
	app   = App{
		Name:    "Bus Eta Bot",
		ID:      "github.com/yi-jiayu/bus-eta-bot-3",
		Version: Version,
	}
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
	Handlers Handlers
	Telegram *tgbotapi.BotAPI
	Datamall *datamall.APIClient
}

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

func NewBusEtaBot(handlers Handlers, tg *tgbotapi.BotAPI, dm *datamall.APIClient) BusEtaBot {
	return BusEtaBot{
		Handlers: handlers,
		Telegram: tg,
		Datamall: dm,
	}
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

func messageErrorHandler(ctx context.Context, bot *BusEtaBot, message *tgbotapi.Message, err error) {
	log.Errorf(ctx, "%v", err)
	go LogEvent(ctx, message.From.ID, "message", "error", fmt.Sprintf("%v", err))

	text := fmt.Sprintf("Oh no! Something went wrong. \n\nRequest ID: `%s`", appengine.RequestID(ctx))
	reply := tgbotapi.NewMessage(message.Chat.ID, text)
	reply.ParseMode = "markdown"

	_, err = bot.Telegram.Send(reply)
	if err != nil {
		log.Errorf(ctx, "%v", err)
		go LogEvent(ctx, message.From.ID, "message", "error", fmt.Sprintf("%v", err))
	}
}

// LogEventWithValue logs an interaction with the bot with a value
func LogEventWithValue(ctx context.Context, userID int, category, action, label string, value int) {
	// don't record analytics data while testing
	if gaTID == "" || ctx == nil || userID == 1 {
		return
	}

	client := urlfetch.Client(ctx)
	gaClient := NewClient(gaTID, client)

	user := User{
		UserID: fmt.Sprintf("%d", userID),
	}

	event := Event{
		Category: category,
		Action:   action,
	}

	if label != "" {
		if len(label) > 500 {
			label = label[:500]
		}
		event.Label = &label
	}

	if value != 0 {
		event.Value = &value
	}

	_, err := gaClient.Send(user, app, event)
	if err != nil {
		log.Errorf(ctx, "%v", err)
	}
}

// LogEvent logs an interaction with the bot
func LogEvent(ctx context.Context, userID int, category, action, label string) {
	LogEventWithValue(ctx, userID, category, action, label, 0)
}
