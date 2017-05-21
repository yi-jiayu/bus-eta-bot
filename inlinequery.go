package main

import (
	"encoding/json"
	"fmt"

	"github.com/yi-jiayu/telegram-bot-api"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

func inlineQueryHandler(ctx context.Context, bot *tgbotapi.BotAPI, ilq *tgbotapi.InlineQuery) error {
	query := ilq.Query

	busStops, err := SearchBusStops(ctx, query)
	if err != nil {
		return err
	}

	fmt.Printf("%v", busStops)

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

	config := tgbotapi.InlineConfig{
		InlineQueryID: ilq.ID,
		Results:       results,
	}

	resp, err := bot.AnswerInlineQuery(config)
	if err != nil {
		log.Errorf(ctx, "%v", resp)
		return err
	}

	return nil
}

func chosenInlineResultHandler(ctx context.Context, bot *tgbotapi.BotAPI, cir *tgbotapi.ChosenInlineResult) error {
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

	_, err = bot.Send(reply)
	if err != nil {
		return err
	}

	return nil
}
