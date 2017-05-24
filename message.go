package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/yi-jiayu/telegram-bot-api"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

// MessageHandler is a handler for incoming messages
type MessageHandler func(context.Context, *BusEtaBot, *tgbotapi.Message) error

// TextHandler handles incoming text messages
func TextHandler(ctx context.Context, bot *BusEtaBot, message *tgbotapi.Message) error {
	if strings.Contains(message.Text, "Fetching etas...") {
		return nil
	}

	chatID := message.Chat.ID
	busStopID, serviceNos := InferEtaQuery(message.Text)

	// ignore the message if the bus stop was all invalid characters
	if busStopID == "" {
		go bot.LogEvent(ctx, message.From, CategoryMessage, ActionIgnoredTextMessage, message.Chat.Type)
		return nil
	}

	var reply tgbotapi.MessageConfig

	if len(busStopID) > 5 {
		if message.ReplyToMessage != nil && message.ReplyToMessage.Text == "Alright, send me a bus stop code to get etas for." {
			reply = tgbotapi.NewMessage(chatID, "Oops, a bus stop code can only contain a maximum of 5 characters.")
		} else {
			go bot.LogEvent(ctx, message.From, CategoryMessage, ActionIgnoredTextMessage, message.Chat.Type)
			return nil
		}
	} else {
		text, err := EtaMessage(ctx, bot, busStopID, serviceNos)
		if err != nil {
			if err == errNotFound {
				reply = tgbotapi.NewMessage(chatID, text)
			} else {
				return err
			}
		} else {
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

			reply = tgbotapi.NewMessage(chatID, text)
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

		}
	}

	if !message.Chat.IsPrivate() {
		reply.ReplyToMessageID = message.MessageID
	}

	if message.ReplyToMessage != nil && message.ReplyToMessage.Text == "Alright, send me a bus stop code to get etas for." {
		go bot.LogEvent(ctx, message.From, CategoryMessage, ActionContinuedTextMessage, message.Chat.Type)
	} else {
		go bot.LogEvent(ctx, message.From, CategoryMessage, ActionEtaTextMessage, message.Chat.Type)
	}

	_, err := bot.Telegram.Send(reply)
	return err
}

// LocationHandler handles messages contain a location
func LocationHandler(ctx context.Context, bot *BusEtaBot, message *tgbotapi.Message) error {
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

	go bot.LogEvent(ctx, message.From, CategoryMessage, ActionLocationMessage, message.Chat.Type)

	reply := tgbotapi.NewMessage(chatID, text)
	_, err = bot.Telegram.Send(reply)
	return err
}

func messageErrorHandler(ctx context.Context, bot *BusEtaBot, message *tgbotapi.Message, err error) {
	log.Errorf(ctx, "%v", err)

	text := fmt.Sprintf("Oh no! Something went wrong. \n\nRequest ID: `%s`", appengine.RequestID(ctx))
	reply := tgbotapi.NewMessage(message.Chat.ID, text)
	reply.ParseMode = "markdown"

	_, err = bot.Telegram.Send(reply)
	if err != nil {
		log.Errorf(ctx, "%v", err)
	}
}
