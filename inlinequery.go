package busetabot

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/yi-jiayu/telegram-bot-api"

	"github.com/yi-jiayu/bus-eta-bot/v4/telegram"
)

// NearbyBusStopsRadius is the search range in metres for inline queries for nearby bus stops.
const NearbyBusStopsRadius = 1000.0

const InlineQueryResultsLimit = 50

type StreetViewProvider interface {
	GetPhotoURLByLocation(lat, lon float64, width, height int) (string, error)
}

func GetNearbyInlineQueryResults(ctx context.Context, streetView StreetViewProvider, busStops BusStopRepository, lat, lon float64) ([]telegram.InlineQueryResult, error) {
	nearbyBusStops := busStops.Nearby(ctx, lat, lon, NearbyBusStopsRadius, InlineQueryResultsLimit)
	var results []telegram.InlineQueryResult
	for _, nearby := range nearbyBusStops {
		var result telegram.InlineQueryResultArticle
		result, err := buildInlineQueryResultGeo(streetView, nearby)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

// InlineQueryHandler handles inline queries
func InlineQueryHandler(ctx context.Context, bot *BusEtaBot, ilq *tgbotapi.InlineQuery) error {
	query := ilq.Query
	var err error
	var showingNearby bool
	results := make([]telegram.InlineQueryResult, 0)
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
	answer := telegram.AnswerInlineQueryRequest{
		InlineQueryID: ilq.ID,
		Results:       results,
		CacheTime:     cacheTime,
	}
	if showingNearby {
		go bot.LogEvent(ctx, ilq.From, CategoryInlineQuery, ActionNewNearbyInlineQuery, "")
	} else {
		go bot.LogEvent(ctx, ilq.From, CategoryInlineQuery, ActionNewInlineQuery, "")
	}
	err = bot.TelegramService.Do(answer)
	if err != nil {
		return errors.Wrap(err, "error answering inline query")
	}
	return nil
}

func buildInlineQueryResult(streetView StreetViewProvider, bs BusStop) (telegram.InlineQueryResultArticle, error) {
	text := fmt.Sprintf("*%s (%s)*\n%s\n`Fetching etas...`", bs.Description, bs.BusStopCode, bs.RoadName)
	var thumbnail string
	if streetView != nil {
		if lat, lon := bs.Latitude, bs.Longitude; lat != 0 && lon != 0 {
			var err error
			thumbnail, err = streetView.GetPhotoURLByLocation(lat, lon, 100, 100)
			if err != nil {
				return telegram.InlineQueryResultArticle{}, err
			}
		}
	}
	markup := NewETAMessageReplyMarkup(bs.BusStopCode, nil, true)
	result := telegram.InlineQueryResultArticle{
		ID:          bs.BusStopCode,
		Title:       fmt.Sprintf("%s (%s)", bs.Description, bs.BusStopCode),
		Description: bs.RoadName,
		ThumbURL:    thumbnail,
		InputMessageContent: telegram.InputTextMessageContent{
			MessageText: text,
			ParseMode:   "markdown",
		},
		ReplyMarkup: markup,
	}
	return result, nil
}

func buildInlineQueryResultGeo(streetView StreetViewProvider, stop NearbyBusStop) (telegram.InlineQueryResultArticle, error) {
	result, err := buildInlineQueryResult(streetView, stop.BusStop)
	if err != nil {
		return telegram.InlineQueryResultArticle{}, err
	}
	result.ID = stop.BusStopCode + " geo"
	result.Description = fmt.Sprintf("%.0f m away", stop.Distance)
	return result, nil
}

// ChosenInlineResultHandler handles a chosen inline result
func ChosenInlineResultHandler(ctx context.Context, bot *BusEtaBot, cir *tgbotapi.ChosenInlineResult) error {
	tokens := strings.Split(cir.ResultID, " ")
	busStopID := tokens[0]

	text, err := ETAMessageText(bot.BusStops, bot.Datamall, SummaryETAFormatter{}, bot.NowFunc(), busStopID, nil)
	if err != nil {
		return err
	}
	markup := NewETAMessageReplyMarkup(busStopID, nil, true)
	reply := telegram.EditMessageTextRequest{
		InlineMessageID: cir.InlineMessageID,
		Text:            text,
		ParseMode:       "markdown",
		ReplyMarkup:     markup,
	}

	if len(tokens) > 1 && tokens[1] == "geo" {
		go bot.LogEvent(ctx, cir.From, CategoryChosenInlineResult, ActionChosenNearbyInlineResult, "")
	} else {
		go bot.LogEvent(ctx, cir.From, CategoryChosenInlineResult, ActionChosenInlineResult, "")
	}

	err = bot.TelegramService.Do(reply)
	if err != nil {
		return errors.Wrap(err, "error updating inline query after chosen inline result")
	}
	return nil
}
