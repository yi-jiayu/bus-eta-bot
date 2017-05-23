package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"

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

// BusEtaBot contains all the handlers used by this bot
type BusEtaBot struct {
	CommandHandlers           map[string]MessageHandler
	TextHandler               MessageHandler
	LocationHandler           MessageHandler
	CallbackQueryHandlers     map[string]CallbackQueryHandler
	InlineQueryHandler        func(ctx context.Context, bot *tgbotapi.BotAPI, ilq *tgbotapi.InlineQuery) error
	ChosenInlineResultHandler func(ctx context.Context, bot *tgbotapi.BotAPI, cir *tgbotapi.ChosenInlineResult) error
	MessageErrorHandler       func(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message, err error)
	CallbackErrorHandler      func(ctx context.Context, bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, err error)
}

var busEtaBot = BusEtaBot{
	CommandHandlers:           commandHandlers,
	TextHandler:               TextHandler,
	LocationHandler:           LocationHandler,
	CallbackQueryHandlers:     callbackQueryHandlers,
	InlineQueryHandler:        InlineQueryHandler,
	ChosenInlineResultHandler: ChosenInlineResultHandler,
	MessageErrorHandler:       messageErrorHandler,
	CallbackErrorHandler:      callbackErrorHandler,
}

// HandleUpdate dispatches an incoming update to the corresponding handler depending on the update type
func (b BusEtaBot) HandleUpdate(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	if message := update.Message; message != nil {
		if command := message.Command(); command != "" {
			// handle command
			if handler, exists := b.CommandHandlers[command]; exists {
				err := handler(ctx, bot, message)
				if err != nil {
					messageErrorHandler(ctx, bot, message, err)
					return
				}
			}

			return
		}

		if text := message.Text; text != "" {
			err := b.TextHandler(ctx, bot, message)
			if err != nil {
				messageErrorHandler(ctx, bot, message, err)
				return
			}

			return
		}

		if location := message.Location; location != nil {
			err := b.LocationHandler(ctx, bot, message)
			if err != nil {
				messageErrorHandler(ctx, bot, message, err)
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
			callbackErrorHandler(ctx, bot, cbq, err)
			return
		}

		if cbqType, ok := data["t"].(string); ok {
			if handler, ok := b.CallbackQueryHandlers[cbqType]; ok {
				err := handler(ctx, bot, cbq)
				if err != nil {
					callbackErrorHandler(ctx, bot, cbq, err)
					return
				}
			}
		} else {
			callbackErrorHandler(ctx, bot, cbq, errors.New("unrecognised callback query"))
			return
		}

		return
	}

	if ilq := update.InlineQuery; ilq != nil {
		// handle inline query
		err := b.InlineQueryHandler(ctx, bot, ilq)
		if err != nil {
			log.Errorf(ctx, "%v", err)
			return
		}

		return
	}

	if cir := update.ChosenInlineResult; cir != nil {
		err := b.ChosenInlineResultHandler(ctx, bot, cir)
		if err != nil {
			log.Errorf(ctx, "%v", err)
			return
		}

		return
	}
}

func messageErrorHandler(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message, err error) {
	log.Errorf(ctx, "%v", err)
	go LogEvent(ctx, message.From.ID, "message", "error", fmt.Sprintf("%v", err))

	text := fmt.Sprintf("Oh no! Something went wrong. \n\nRequest ID: `%s`", appengine.RequestID(ctx))
	reply := tgbotapi.NewMessage(message.Chat.ID, text)
	reply.ParseMode = "markdown"

	_, err = bot.Send(reply)
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
