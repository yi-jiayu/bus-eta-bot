package busetabot

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/yi-jiayu/telegram-bot-api"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"github.com/yi-jiayu/bus-eta-bot/v5/telegram"
)

// MessageHandler is a handler for incoming messages
type MessageHandler func(context.Context, *BusEtaBot, *tgbotapi.Message) error

// TextHandler handles incoming text messages
func TextHandler(ctx context.Context, bot *BusEtaBot, message *tgbotapi.Message) error {
	if strings.Contains(message.Text, "Fetching etas...") {
		return nil
	}

	chatID := message.Chat.ID
	// a message is a continuation if it was a reply to a message asking for a bus stop code
	continuation := message.ReplyToMessage != nil && message.ReplyToMessage.Text == "Alright, send me a bus stop code to get etas for."

	busStopID, serviceNos, err := InferEtaQuery(message.Text)
	if err != nil {
		// if it wasn't a continuation, we ignore the message
		if !continuation {
			return nil
		}
		go bot.LogEvent(ctx, message.From, CategoryMessage, ActionContinuedTextMessage, message.Chat.Type)
		// else, we should inform the user if it was invalid
		req := telegram.SendMessageRequest{
			ChatID: chatID,
			Text:   "Oops, a bus stop code should be a 5-digit number.",
		}
		err = bot.TelegramService.Do(req)
		if err != nil {
			return errors.Wrap(err, "error sending message")
		}
		return nil
	}
	text, err := ETAMessageText(bot.BusStops, bot.Datamall, SummaryETAFormatter{}, bot.NowFunc(), busStopID, serviceNos)
	if err != nil {
		return err
	}
	markup := NewETAMessageReplyMarkup(busStopID, serviceNos, "", false)
	req := telegram.SendMessageRequest{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   "markdown",
		ReplyMarkup: markup,
	}

	if continuation {
		go bot.LogEvent(ctx, message.From, CategoryMessage, ActionContinuedTextMessage, message.Chat.Type)
	} else {
		go bot.LogEvent(ctx, message.From, CategoryMessage, ActionEtaTextMessage, message.Chat.Type)
	}

	err = bot.TelegramService.Do(req)
	if err != nil {
		return errors.Wrap(err, "error sending message")
	}
	return nil
}

// LocationHandler handles messages contain a location
func LocationHandler(ctx context.Context, bot *BusEtaBot, message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	location := message.Location

	nearby := bot.BusStops.Nearby(ctx, location.Latitude, location.Longitude, 500, 5)
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
		// sleep a while so that this message is received before the others
		time.Sleep(250 * time.Millisecond)

		for _, bs := range nearby {
			distance := bs.Distance

			callbackData := CallbackData{
				Type:      "new_eta",
				BusStopID: bs.BusStopCode,
			}

			callbackDataJSON, err := json.Marshal(callbackData)
			if err != nil {
				return err
			}
			callbackDataJSONStr := string(callbackDataJSON)

			reply := tgbotapi.NewVenue(chatID, fmt.Sprintf("%s (%s)", bs.Description, bs.BusStopCode), fmt.Sprintf("%.0f m away", distance), bs.Latitude, bs.Longitude)
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

			_, err = bot.Telegram.Send(reply)
			if err != nil {
				log.Errorf(ctx, "%+v", err)
			}
		}

		return nil
	}

	go bot.LogEvent(ctx, message.From, CategoryMessage, ActionLocationMessage, message.Chat.Type)

	reply := tgbotapi.NewMessage(chatID, "Oops, I couldn't find any bus stops within 500 m of your location.")
	_, err := bot.Telegram.Send(reply)
	return err
}

func messageErrorHandler(ctx context.Context, bot *BusEtaBot, message *tgbotapi.Message, err error) {
	log.Errorf(ctx, "%+v", err)

	text := fmt.Sprintf("Oh no! Something went wrong. \n\nRequest ID: `%s`", appengine.RequestID(ctx))
	reply := tgbotapi.NewMessage(message.Chat.ID, text)
	reply.ParseMode = "markdown"

	_, err = bot.Telegram.Send(reply)
	if err != nil {
		log.Errorf(ctx, "%+v", err)
	}
}
