package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/yi-jiayu/telegram-bot-api"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

var illegalCharsRegex = regexp.MustCompile(`[^A-Z0-9 ]`)

// MessageHandler is a handler for incoming messages
type MessageHandler func(context.Context, *tgbotapi.BotAPI, *tgbotapi.Message) error

func updateHandler(ctx context.Context, bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	if message := update.Message; message != nil {
		if command := message.Command(); command != "" {
			// handle command
			if handler, exists := commandHandlers[command]; exists {
				err := handler(ctx, bot, message)
				if err != nil {
					messageErrorHandler(ctx, bot, message, err)
					return
				}
			}

			return
		}

		if text := message.Text; text != "" {
			err := TextHandler(ctx, bot, message)
			if err != nil {
				messageErrorHandler(ctx, bot, message, err)
				return
			}

			return
		}

		if location := message.Location; location != nil {
			err := locationHandler(ctx, bot, message)
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
			if handler, ok := callbackQueryHandlers[cbqType]; ok {
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
		err := inlineQueryHandler(ctx, bot, ilq)
		if err != nil {
			log.Errorf(ctx, "%v", err)
			return
		}

		return
	}

	if cir := update.ChosenInlineResult; cir != nil {
		err := chosenInlineResultHandler(ctx, bot, cir)
		if err != nil {
			log.Errorf(ctx, "%v", err)
			return
		}

		return
	}
}

func messageErrorHandler(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message, err error) {
	log.Errorf(ctx, "%v", err)

	text := fmt.Sprintf("Oh no! Something went wrong. \n\nRequest ID: `%s`", appengine.RequestID(ctx))
	reply := tgbotapi.NewMessage(message.Chat.ID, text)
	reply.ParseMode = "markdown"

	_, err = bot.Send(reply)
	if err != nil {
		log.Errorf(ctx, "%v", err)
	}
}

// TextHandler handles incoming text messages
func TextHandler(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	if strings.Contains(message.Text, "Fetching etas...") {
		return nil
	}

	chatID := message.Chat.ID
	busStopID, serviceNos := InferEtaQuery(message.Text)

	if busStopID == "" {
		return nil
	}

	text, err := EtaMessage(ctx, busStopID, serviceNos)
	if err != nil {
		return err
	}

	callbackData := EtaCallbackData{
		Type:       "refresh",
		BusStopID:  busStopID,
		ServiceNos: serviceNos,
	}

	callbackDataJSON, err := json.Marshal(callbackData)
	if err != nil {
		return err
	}
	callbackDataJSONStr := string(callbackDataJSON)

	reply := tgbotapi.NewMessage(chatID, text)
	reply.ParseMode = "markdown"
	reply.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.InlineKeyboardButton{
					Text:         "Refresh",
					CallbackData: &callbackDataJSONStr,
				},
			},
		},
	}

	if !message.Chat.IsPrivate() {
		reply.ReplyToMessageID = message.MessageID
	}

	_, err = bot.Send(reply)
	if err != nil {
		return err
	}

	return nil
}

func locationHandler(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	location := message.Location

	nearby, err := GetNearbyBusStops(ctx, location.Latitude, location.Longitude)
	if err != nil {
		return err
	}

	text := "Nearby bus stops: "
	for _, bs := range nearby {
		text += fmt.Sprintf("%s (%s), ", bs.Description, bs.BusStopID)
	}

	reply := tgbotapi.NewMessage(chatID, text)
	_, err = bot.Send(reply)
	if err != nil {
		return err
	}

	return nil
}

// InferEtaQuery extracts a bus stop ID and service numbers from a text message
func InferEtaQuery(text string) (string, []string) {
	if len(text) > 30 {
		text = text[:30]
	}

	text = strings.ToUpper(text)
	text = illegalCharsRegex.ReplaceAllString(text, "")
	tokens := strings.Split(text, " ")
	busStopID, serviceNos := tokens[0], tokens[1:]

	return busStopID, serviceNos
}
