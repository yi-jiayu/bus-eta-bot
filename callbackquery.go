package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/yi-jiayu/telegram-bot-api"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

var callbackQueryHandlers = map[string]CallbackQueryHandler{
	"refresh":  RefreshCallbackHandler,
	"eta":      EtaCallbackHandler,
	"eta_demo": EtaDemoCallbackHandler,
}

// CallbackQueryHandler is a handler for callback queries
type CallbackQueryHandler func(ctx context.Context, bot *BusEtaBot, cbq *tgbotapi.CallbackQuery) error

func answerCallbackQuery(bot *BusEtaBot, c tgbotapi.Chattable, answer tgbotapi.CallbackConfig) error {
	errs := make([]error, 0)

	var wg sync.WaitGroup
	wg.Add(2)

	errChan := make(chan error, 2)
	go func() {
		for err := range errChan {
			errs = append(errs, err)
		}
	}()

	go func() {
		defer wg.Done()

		_, err := bot.Telegram.Send(c)
		if err != nil {
			errChan <- err
		}
	}()

	go func() {
		defer wg.Done()

		_, err := bot.Telegram.AnswerCallbackQuery(answer)
		if err != nil {
			errChan <- err
		}
	}()

	wg.Wait()
	close(errChan)

	switch len(errs) {
	case 1:
		return errs[0]
	case 2:
		return errors.New(errs[0].Error() + "\n" + errs[1].Error())
	default:
		return nil
	}
}

func updateEtaMessage(ctx context.Context, bot *BusEtaBot, cbq *tgbotapi.CallbackQuery, busStopID string, serviceNos []string) error {
	text, err := EtaMessage(ctx, bot, busStopID, serviceNos)
	if err != nil {
		return err
	}

	var reply tgbotapi.EditMessageTextConfig
	if cbq.Message != nil {
		chatID := cbq.Message.Chat.ID
		messageID := cbq.Message.MessageID
		reply = tgbotapi.NewEditMessageText(chatID, messageID, text)
	} else {
		inlineMessageID := cbq.InlineMessageID
		reply = tgbotapi.EditMessageTextConfig{
			BaseEdit: tgbotapi.BaseEdit{
				InlineMessageID: inlineMessageID,
			},
			Text: text,
		}
	}

	reply.ParseMode = "markdown"
	reply.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.InlineKeyboardButton{
					Text:         "Refresh",
					CallbackData: &cbq.Data,
				},
			},
		},
	}

	answer := tgbotapi.NewCallback(cbq.ID, "Etas updated!")

	if cbq.Message != nil {
		go bot.LogEvent(ctx, cbq.From, CategoryCallback, ActionRefreshCallback, cbq.Message.Chat.Type)
	} else {
		go bot.LogEvent(ctx, cbq.From, CategoryCallback, ActionRefreshCallback, LabelInlineMessage)
	}

	err = answerCallbackQuery(bot, reply, answer)
	return err
}

// RefreshCallbackHandler handles the callback for the Refresh button on an eta message.
func RefreshCallbackHandler(ctx context.Context, bot *BusEtaBot, cbq *tgbotapi.CallbackQuery) error {
	var data EtaCallbackData
	err := json.Unmarshal([]byte(cbq.Data), &data)
	if err != nil {
		return err
	}

	bsID, sNos := data.BusStopID, data.ServiceNos

	return updateEtaMessage(ctx, bot, cbq, bsID, sNos)
}

// EtaCallbackHandler handles callback queries from eta messages from old versions of the bot for
// backwards-compatibility.
func EtaCallbackHandler(ctx context.Context, bot *BusEtaBot, cbq *tgbotapi.CallbackQuery) error {
	var data EtaCallbackData
	err := json.Unmarshal([]byte(cbq.Data), &data)
	if err != nil {
		return err
	}

	var bsID string
	var sNos []string
	if data.Argstr != "" {
		bsID, sNos = InferEtaQuery(data.Argstr)
	} else {
		bsID = data.BusStopID
		sNos = data.ServiceNos
	}

	callbackData := EtaCallbackData{
		Type:       "refresh",
		BusStopID:  bsID,
		ServiceNos: sNos,
	}

	callbackDataJSON, err := json.Marshal(callbackData)
	if err != nil {
		return err
	}
	callbackDataJSONStr := string(callbackDataJSON)
	cbq.Data = callbackDataJSONStr

	return updateEtaMessage(ctx, bot, cbq, bsID, sNos)
}

// EtaDemoCallbackHandler handles an eta_demo callback from a start command.
func EtaDemoCallbackHandler(ctx context.Context, bot *BusEtaBot, cbq *tgbotapi.CallbackQuery) error {
	chatID := cbq.Message.Chat.ID

	text, err := EtaMessage(ctx, bot, "96049", nil)
	if err != nil {
		return err
	}

	callbackData := EtaCallbackData{
		Type:      "refresh",
		BusStopID: "96049",
	}

	callbackDataJSON, err := json.Marshal(callbackData)
	if err != nil {
		return err
	}
	callbackDataJSONStr := string(callbackDataJSON)

	reply := tgbotapi.NewMessage(chatID, text)
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

	if !cbq.Message.Chat.IsPrivate() {
		reply.ReplyToMessageID = cbq.Message.MessageID
	}

	answer := tgbotapi.NewCallback(cbq.ID, "")

	go bot.LogEvent(ctx, cbq.From, CategoryCallback, ActionEtaDemoCallback, cbq.Message.Chat.Type)

	err = answerCallbackQuery(bot, reply, answer)
	return err
}

// callbackErrorHandler is for informing the user about an error while processing a callback query.
func callbackErrorHandler(ctx context.Context, bot *BusEtaBot, cbq *tgbotapi.CallbackQuery, err error) {
	log.Errorf(ctx, "%v", err)

	text := fmt.Sprintf("Oh no! Something went wrong. \n\nRequest ID: `%s`", appengine.RequestID(ctx))
	answer := tgbotapi.NewCallbackWithAlert(cbq.ID, text)

	_, err = bot.Telegram.AnswerCallbackQuery(answer)
	if err != nil {
		log.Errorf(ctx, "%v", err)
	}
}
