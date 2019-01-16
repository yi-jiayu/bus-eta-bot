package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/yi-jiayu/telegram-bot-api"
	"google.golang.org/appengine/log"
)

// NearbyBusStopsRadius is the search range in metres for inline queries for nearby bus stops.
const NearbyBusStopsRadius = 1000.0

const InlineQueryResultsLimit = 50

type StreetViewProvider interface {
	GetPhotoURLByLocation(lat, lon float64, width, height int) (string, error)
}

func GetNearbyInlineQueryResults(ctx context.Context, streetView StreetViewProvider, busStops BusStopRepository, lat, lon float64) (results []interface{}, err error) {
	nearbyBusStops := busStops.Nearby(ctx, lat, lon, NearbyBusStopsRadius, InlineQueryResultsLimit)
	for _, nearby := range nearbyBusStops {
		var result tgbotapi.InlineQueryResultArticle
		result, err = buildInlineQueryResultGeo(streetView, nearby)
		if err != nil {
			return
		}
		results = append(results, result)
	}
	return
}

// InlineQueryHandler handles inline queries
func InlineQueryHandler(ctx context.Context, bot *BusEtaBot, ilq *tgbotapi.InlineQuery) error {
	query := ilq.Query
	var err error
	var showingNearby bool
	results := make([]interface{}, 0)
	if query == "" && ilq.Location != nil {
		showingNearby = true
		lat, lon := ilq.Location.Latitude, ilq.Location.Longitude
		results, err = GetNearbyInlineQueryResults(ctx, bot.StreetView, bot.BusStops, lat, lon)
		if err != nil {
			return err
		}
	} else {
		showingNearby = false
		busStops := bot.BusStops.Search(ctx, query, InlineQueryResultsLimit)
		for _, bs := range busStops {
			result, err := buildInlineQueryResult(bot.StreetView, bs)
			if err != nil {
				return err
			}
			results = append(results, result)
		}
	}
	var cacheTime int
	if ilq.Query == "" {
		cacheTime = 0
	} else {
		cacheTime = 24 * 3600
	}
	config := tgbotapi.InlineConfig{
		InlineQueryID: ilq.ID,
		Results:       results,
		CacheTime:     cacheTime,
	}
	if showingNearby {
		go bot.LogEvent(ctx, ilq.From, CategoryInlineQuery, ActionNewNearbyInlineQuery, "")
	} else {
		go bot.LogEvent(ctx, ilq.From, CategoryInlineQuery, ActionNewInlineQuery, "")
	}
	resp, err := bot.Telegram.AnswerInlineQuery(config)
	if err != nil {
		log.Errorf(ctx, "%v", resp)
		return err
	}
	return nil
}

func buildInlineQueryResult(streetView StreetViewProvider, bs BusStop) (result tgbotapi.InlineQueryResultArticle, err error) {
	text := fmt.Sprintf("*%s (%s)*\n%s\n`Fetching etas...`", bs.Description, bs.BusStopCode, bs.RoadName)
	var thumbnail string
	if streetView != nil {
		if lat, lon := bs.Latitude, bs.Longitude; lat != 0 && lon != 0 {
			thumbnail, err = streetView.GetPhotoURLByLocation(lat, lon, 100, 100)
			if err != nil {
				return
			}
		}
	}
	replyMarkup, err := EtaMessageReplyMarkup(bs.BusStopCode, nil, true)
	if err != nil {
		return
	}
	result = tgbotapi.InlineQueryResultArticle{
		Type:        "article",
		ID:          bs.BusStopCode,
		Title:       fmt.Sprintf("%s (%s)", bs.Description, bs.BusStopCode),
		Description: bs.RoadName,
		ThumbURL:    thumbnail,
		InputMessageContent: tgbotapi.InputTextMessageContent{
			Text:      text,
			ParseMode: "markdown",
		},
		ReplyMarkup: replyMarkup,
	}
	return
}

func buildInlineQueryResultGeo(streetView StreetViewProvider, stop NearbyBusStop) (result tgbotapi.InlineQueryResultArticle, err error) {
	result, err = buildInlineQueryResult(streetView, stop.BusStop)
	if err != nil {
		return
	}
	result.ID = stop.BusStopCode + " geo"
	result.Description = fmt.Sprintf("%.0f m away", stop.Distance)
	return
}

// ChosenInlineResultHandler handles a chosen inline result
func ChosenInlineResultHandler(ctx context.Context, bot *BusEtaBot, cir *tgbotapi.ChosenInlineResult) error {
	tokens := strings.Split(cir.ResultID, " ")
	busStopID := tokens[0]

	text, err := EtaMessageText(bot, busStopID, nil)
	if err != nil {
		return err
	}

	replyMarkup, err := EtaMessageReplyMarkup(busStopID, nil, true)
	if err != nil {
		return err
	}

	reply := tgbotapi.EditMessageTextConfig{
		BaseEdit: tgbotapi.BaseEdit{
			InlineMessageID: cir.InlineMessageID,
			ReplyMarkup:     replyMarkup,
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
