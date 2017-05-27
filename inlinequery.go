package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

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

	var busStops []BusStop
	var err error
	var showingNearby bool
	if query == "" && ilq.Location != nil {
		showingNearby = true

		lat, lon := ilq.Location.Latitude, ilq.Location.Longitude
		busStops, err = GetNearbyBusStops(ctx, lat, lon, 1000, 50)
		if err != nil {
			return err
		}

		if len(busStops) == 0 {
			showingNearby = false
			busStops, err = SearchBusStops(ctx, query, offset)
			if err != nil {
				return err
			}
		}
	} else {
		showingNearby = false

		busStops, err = SearchBusStops(ctx, query, offset)
		if err != nil {
			return err
		}
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

		callbackData := CallbackData{
			Type:      "refresh",
			BusStopID: bs.BusStopID,
		}

		callbackDataJSON, err := json.Marshal(callbackData)
		if err != nil {
			log.Errorf(ctx, "%v", err)
			continue
		}
		callbackDataJSONStr := string(callbackDataJSON)

		var id, desc string
		if showingNearby {
			id = bs.BusStopID + " geo"

			lat, lon := ilq.Location.Latitude, ilq.Location.Longitude
			desc = fmt.Sprintf("%.2f m away", bs.DistanceFrom(lat, lon))
		} else {
			id = bs.BusStopID
			desc = bs.Road
		}

		result := tgbotapi.InlineQueryResultArticle{
			Type:        "article",
			ID:          id,
			Title:       fmt.Sprintf("%s (%s)", bs.Description, bs.BusStopID),
			Description: desc,
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

	var cacheTime int
	if ilq.Query == "" {
		cacheTime = 0
	} else {
		cacheTime = 24 * 3600
	}

	var nextOffset string
	if len(busStops) == 50 {
		nextOffset = fmt.Sprintf("%d", offset+50)
	}

	config := tgbotapi.InlineConfig{
		InlineQueryID: ilq.ID,
		Results:       results,
		CacheTime:     cacheTime,
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
	tokens := strings.Split(cir.ResultID, " ")
	busStopID := tokens[0]

	text, err := EtaMessage(ctx, bot, busStopID, nil)
	if err != nil {
		return err
	}

	callbackData := CallbackData{
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

	if len(tokens) > 1 && tokens[1] == "geo" {
		go bot.LogEvent(ctx, cir.From, CategoryChosenInlineResult, ActionChosenNearbyInlineResult, "")
	} else {
		go bot.LogEvent(ctx, cir.From, CategoryChosenInlineResult, ActionChosenInlineResult, "")
	}

	_, err = bot.Telegram.Send(reply)
	return err
}
