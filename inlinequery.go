package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/yi-jiayu/telegram-bot-api"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

// InlineQueryHandler handles inline queries
func InlineQueryHandler(ctx context.Context, bot *BusEtaBot, ilq *tgbotapi.InlineQuery) error {
	query := ilq.Query

	offset := 0
	if ilq.Offset != "" {
		var err error
		offset, err = strconv.Atoi(ilq.Offset)
		if err != nil {
			return err
		}
	}

	busStops, err := SearchBusStops(ctx, query, offset)
	if err != nil {
		return err
	}

	results := make([]interface{}, 0)
	for _, bs := range busStops {
		text := fmt.Sprintf("*%s (%s)*\n%s\n`Fetching etas...`", bs.Description, bs.BusStopID, bs.Road)

		var thumbnail string
		if bot.StreetView != nil {
			if lat, lon := bs.Location.Lat, bs.Location.Lng; lat != 0 && lon != 0 {
				tn, err := bot.StreetView.GetPhotoURLByLocation(lat, lon, 100, 100)
				if err != nil {
					log.Errorf(ctx, "%v", err)
				} else {
					thumbnail = tn
				}
			}
		}

		callbackData := EtaCallbackData{
			Type:      "refresh",
			BusStopID: bs.BusStopID,
		}

		callbackDataJSON, err := json.Marshal(callbackData)
		if err != nil {
			log.Errorf(ctx, "%v", err)
			continue
		}
		callbackDataJSONStr := string(callbackDataJSON)

		result := tgbotapi.InlineQueryResultArticle{
			Type:        "article",
			ID:          bs.BusStopID,
			Title:       fmt.Sprintf("%s (%s)", bs.Description, bs.BusStopID),
			Description: bs.Road,
			ThumbURL:    thumbnail,
			InputMessageContent: tgbotapi.InputTextMessageContent{
				Text:      text,
				ParseMode: "markdown",
			},
			ReplyMarkup: &tgbotapi.InlineKeyboardMarkup{
				InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
					{
						tgbotapi.InlineKeyboardButton{
							Text:         "Refresh",
							CallbackData: &callbackDataJSONStr,
						},
					},
				},
			},
		}
		results = append(results, result)
	}

	var nextOffset string
	if len(busStops) == 50 {
		nextOffset = fmt.Sprintf("%d", offset+50)
	}

	config := tgbotapi.InlineConfig{
		InlineQueryID: ilq.ID,
		Results:       results,
		NextOffset:    nextOffset,
	}

	if ilq.Offset == "" {
		go bot.LogEvent(ctx, ilq.From, CategoryInlineQuery, ActionNewInlineQuery, "")
	} else {
		go bot.LogEvent(ctx, ilq.From, CategoryInlineQuery, ActionOffsetInlineQuery, "")
	}

	resp, err := bot.Telegram.AnswerInlineQuery(config)
	if err != nil {
		log.Errorf(ctx, "%v", resp)
		return err
	}

	return nil
}

// ChosenInlineResultHandler handles a chosen inline result
func ChosenInlineResultHandler(ctx context.Context, bot *BusEtaBot, cir *tgbotapi.ChosenInlineResult) error {
	busStopID := cir.ResultID

	text, err := EtaMessage(ctx, bot, busStopID, nil)
	if err != nil {
		return err
	}

	callbackData := EtaCallbackData{
		Type:      "refresh",
		BusStopID: busStopID,
	}

	callbackDataJSON, err := json.Marshal(callbackData)
	if err != nil {
		return err
	}
	callbackDataJSONStr := string(callbackDataJSON)

	reply := tgbotapi.EditMessageTextConfig{
		BaseEdit: tgbotapi.BaseEdit{
			InlineMessageID: cir.InlineMessageID,
			ReplyMarkup: &tgbotapi.InlineKeyboardMarkup{
				InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
					{
						tgbotapi.InlineKeyboardButton{
							Text:         "Refresh",
							CallbackData: &callbackDataJSONStr,
						},
					},
				},
			},
		},
		Text:      text,
		ParseMode: "markdown",
	}

	go bot.LogEvent(ctx, cir.From, CategoryInlineQuery, ActionChosenInlineResult, "")

	_, err = bot.Telegram.Send(reply)
	return err
}
