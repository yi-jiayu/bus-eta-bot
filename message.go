package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/yi-jiayu/telegram-bot-api"
	"golang.org/x/net/context"
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
		go LogEvent(ctx, message.From.ID, "message", "text", "ignored")
		return nil
	}

	var reply tgbotapi.MessageConfig
	var label string

	if len(busStopID) > 5 {
		if message.ReplyToMessage != nil && message.ReplyToMessage.Text == "Alright, send me a bus stop code to get etas for." {
			reply = tgbotapi.NewMessage(chatID, "Oops, a bus stop code can only contain a maximum of 5 characters.")
			label = "continued_invalid"
		} else {
			return nil
		}
	} else {
		text, err := EtaMessage(ctx, bot, busStopID, serviceNos)
		if err != nil {
			if err == errNotFound {
				reply = tgbotapi.NewMessage(chatID, text)
				label = "continued_not_found"
			} else {
				go LogEvent(ctx, message.From.ID, "message", "text", "ignored")
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

			if message.ReplyToMessage != nil && message.ReplyToMessage.Text == "Alright, send me a bus stop code to get etas for." {
				label = "continued_ok"
			} else {
				label = "ok"
			}
		}
	}

	if !message.Chat.IsPrivate() {
		reply.ReplyToMessageID = message.MessageID
	}

	go LogEvent(ctx, message.From.ID, "message", "text_eta", label)

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

	go LogEvent(ctx, message.From.ID, "message", "location", "")

	reply := tgbotapi.NewMessage(chatID, text)
	_, err = bot.Telegram.Send(reply)
	return err
}
