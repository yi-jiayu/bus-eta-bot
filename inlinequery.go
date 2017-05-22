package main

import (
	"encoding/json"
	"fmt"

	"github.com/yi-jiayu/telegram-bot-api"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
	"strconv"
)

// InlineQueryHandler handles inline queries
func InlineQueryHandler(ctx context.Context, bot *tgbotapi.BotAPI, ilq *tgbotapi.InlineQuery) error {
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
	if len(busStops) > 0 {
		nextOffset = fmt.Sprintf("%d", offset+50)
	}

	config := tgbotapi.InlineConfig{
		InlineQueryID: ilq.ID,
		Results:       results,
		NextOffset:    nextOffset,
	}

	var action string
	if ilq.Offset == "" {
		action = "new"
	} else {
		action = "offset"
	}
	go LogEvent(ctx, ilq.From.ID, "inline_query", action, query)

	resp, err := bot.AnswerInlineQuery(config)
	if err != nil {
		log.Errorf(ctx, "%v", resp)
		return err
	}

	return nil
}

// ChosenInlineResultHandler handles a chosen inline result
func ChosenInlineResultHandler(ctx context.Context, bot *tgbotapi.BotAPI, cir *tgbotapi.ChosenInlineResult) error {
	busStopID := cir.ResultID

	text, err := EtaMessage(ctx, busStopID, nil)
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

	go LogEvent(ctx, cir.From.ID, "inline_query", "chosen_inline_result", fmt.Sprintf("%s %s", cir.ResultID, cir.Query))

	_, err = bot.Send(reply)
	return err
}
