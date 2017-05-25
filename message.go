package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

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
	continuation := message.ReplyToMessage != nil && message.ReplyToMessage.Text == "Alright, send me a bus stop code to get etas for."

	var reply tgbotapi.MessageConfig
	busStopID, serviceNos, err := InferEtaQuery(message.Text)
	if err != nil {
		if !continuation {
			go bot.LogEvent(ctx, message.From, CategoryMessage, ActionIgnoredTextMessage, message.Chat.Type)
			return nil
		}

		if err == errBusStopIDInvalid {
			reply = tgbotapi.NewMessage(chatID, "Oops, that bus stop code was invalid.")
		} else {
			reply = tgbotapi.NewMessage(chatID, "Oops, a bus stop code can only contain a maximum of 5 characters.")
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

	_, err = bot.Telegram.Send(reply)
	return err
}

// LocationHandler handles messages contain a location
func LocationHandler(ctx context.Context, bot *BusEtaBot, message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	location := message.Location

	nearby, err := GetNearbyBusStops(ctx, location.Latitude, location.Longitude, 3)
	if err != nil {
		return err
	}

	if len(nearby) > 0 {
		go bot.LogEvent(ctx, message.From, CategoryMessage, ActionLocationMessage, message.Chat.Type)

		reply := tgbotapi.NewMessage(chatID, "Here are some bus stops near your location:")
		if !message.Chat.IsPrivate() {
			reply.ReplyToMessageID = message.MessageID
		}
		_, err := bot.Telegram.Send(reply)
		if err != nil {
			return err
		}

		var wg sync.WaitGroup
		wg.Add(len(nearby))

		for _, bs := range nearby {
			distance := bs.DistanceFrom(location.Latitude, location.Longitude)

			callbackData := EtaCallbackData{
				Type:      "new_eta",
				BusStopID: bs.BusStopID,
			}

			callbackDataJSON, err := json.Marshal(callbackData)
			if err != nil {
				return err
			}
			callbackDataJSONStr := string(callbackDataJSON)

			reply := tgbotapi.NewVenue(chatID, fmt.Sprintf("%s (%s)", bs.Description, bs.BusStopID), fmt.Sprintf("%.2f m away", distance), bs.Location.Lat, bs.Location.Lng)
			reply.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
				InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
					{
						tgbotapi.InlineKeyboardButton{
							Text:         "Get etas",
							CallbackData: &callbackDataJSONStr,
						},
					},
				},
			}

			time.Sleep(250 * time.Millisecond)

			go func() {
				defer wg.Done()

				_, err := bot.Telegram.Send(reply)
				if err != nil {
					log.Errorf(ctx, "%v", err)
				}
			}()
		}

		wg.Wait()
		return nil
	}

	go bot.LogEvent(ctx, message.From, CategoryMessage, ActionLocationMessage, message.Chat.Type)

	reply := tgbotapi.NewMessage(chatID, "Oops, I couldn't find any bus stops within 500 m of your location.")
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
