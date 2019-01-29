package busetabot

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/pkg/errors"
	"github.com/yi-jiayu/telegram-bot-api"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"github.com/yi-jiayu/bus-eta-bot/v4/telegram"
)

var callbackQueryHandlers = map[string]CallbackQueryHandler{
	"refresh":  RefreshCallbackHandler,
	"resend":   NewEtaHandler,
	"eta":      EtaCallbackHandler,
	"eta_demo": EtaDemoCallbackHandler,
	"new_eta":  NewEtaHandler,
	"addf":     ToggleFavouritesHandler,
	"togf":     ToggleFavouritesHandler,
}

// CallbackQueryHandler is a handler for callback queries
type CallbackQueryHandler func(ctx context.Context, bot *BusEtaBot, cbq *tgbotapi.CallbackQuery, responses chan<- Response)

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
			err = errors.Wrap(err, fmt.Sprintf("%#v", c))
			errChan <- err
		}
	}()

	go func() {
		defer wg.Done()

		_, err := bot.Telegram.AnswerCallbackQuery(answer)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("%#v", answer))
			errChan <- err
		}
	}()

	wg.Wait()
	close(errChan)

	switch len(errs) {
	case 1:
		return errs[0]
	case 2:
		return fmt.Errorf("%v\n%v", errs[0], errs[1])
	default:
		return nil
	}
}

func updateETAMessage(ctx context.Context, bot *BusEtaBot, cbq *tgbotapi.CallbackQuery, code string, services []string, responses chan<- Response) {
	text, err := ETAMessageText(bot.BusStops, bot.Datamall, SummaryETAFormatter{}, bot.NowFunc(), code, services)
	if err != nil {
		responses <- notOk(err)
		return
	}
	editMessageTextRequest := telegram.EditMessageTextRequest{
		Text:      text,
		ParseMode: "markdown",
	}
	if cbq.InlineMessageID != "" {
		editMessageTextRequest.InlineMessageID = cbq.InlineMessageID
		editMessageTextRequest.ReplyMarkup = NewETAMessageReplyMarkup(code, services, true)
	} else {
		editMessageTextRequest.ChatID = cbq.Message.Chat.ID
		editMessageTextRequest.MessageID = cbq.Message.MessageID
		editMessageTextRequest.ReplyMarkup = NewETAMessageReplyMarkup(code, services, false)
	}
	responses <- ok(editMessageTextRequest)
	answerCallbackQueryRequest := telegram.AnswerCallbackQueryRequest{
		CallbackQueryID: cbq.ID,
		Text:            "ETAs updated!",
	}
	responses <- ok(answerCallbackQueryRequest)
}

func sendETAMessage(ctx context.Context, bot *BusEtaBot, cbq *tgbotapi.CallbackQuery, code string, services []string, responses chan<- Response) {
	text, err := ETAMessageText(bot.BusStops, bot.Datamall, SummaryETAFormatter{}, bot.NowFunc(), code, services)
	if err != nil {
		responses <- notOk(err)
		return
	}
	markup := NewETAMessageReplyMarkup(code, services, false)
	sendMessageRequest := telegram.SendMessageRequest{
		Text:        text,
		ChatID:      cbq.Message.Chat.ID,
		ParseMode:   "markdown",
		ReplyMarkup: markup,
	}
	responses <- ok(sendMessageRequest)
	answerCallbackQueryRequest := telegram.AnswerCallbackQueryRequest{
		CallbackQueryID: cbq.ID,
		Text:            "ETAs sent!",
	}
	responses <- ok(answerCallbackQueryRequest)
}

// RefreshCallbackHandler handles the callback for the Refresh button on an eta message.
func RefreshCallbackHandler(ctx context.Context, bot *BusEtaBot, cbq *tgbotapi.CallbackQuery, responses chan<- Response) {
	defer close(responses)

	var data CallbackData
	err := json.Unmarshal([]byte(cbq.Data), &data)
	if err != nil {
		responses <- notOk(errors.Wrap(err, "error unmarshalling callback data"))
		return
	}
	updateETAMessage(ctx, bot, cbq, data.BusStopID, data.ServiceNos, responses)
	return
}

// EtaCallbackHandler handles callback queries from eta messages from old versions of the bot for
// backwards-compatibility.
func EtaCallbackHandler(ctx context.Context, bot *BusEtaBot, cbq *tgbotapi.CallbackQuery, responses chan<- Response) {
	defer close(responses)

	var data CallbackData
	err := json.Unmarshal([]byte(cbq.Data), &data)
	if err != nil {
		responses <- notOk(errors.Wrap(err, "error unmarshalling callback data"))
		return
	}
	var code string
	var services []string
	if data.Argstr != "" {
		code, services, _ = InferEtaQuery(data.Argstr)
	} else {
		code = data.BusStopID
		services = data.ServiceNos
	}
	updateETAMessage(ctx, bot, cbq, code, services, responses)
	return
}

// EtaDemoCallbackHandler handles an eta_demo callback from a start command.
func EtaDemoCallbackHandler(ctx context.Context, bot *BusEtaBot, cbq *tgbotapi.CallbackQuery, responses chan<- Response) {
	go bot.LogEvent(ctx, cbq.From, CategoryCallback, ActionEtaDemoCallback, cbq.Message.Chat.Type)

	sendETAMessage(ctx, bot, cbq, "96049", nil, responses)
	close(responses)
}

// NewEtaHandler sends etas for a bus stop when a user taps "Get etas" on a bus stop location returned from a
// location query
func NewEtaHandler(ctx context.Context, bot *BusEtaBot, cbq *tgbotapi.CallbackQuery, responses chan<- Response) {
	var data CallbackData
	err := json.Unmarshal([]byte(cbq.Data), &data)
	if err != nil {
		responses <- notOk(err)
		return
	}
	code, services := data.BusStopID, data.ServiceNos
	sendETAMessage(ctx, bot, cbq, code, services, responses)
	close(responses)
}

func stringInSlice(a string, list []string) (bool, int) {
	for i, b := range list {
		if b == a {
			return true, i
		}
	}
	return false, 0
}

// ToggleFavouritesHandler handles the toggle favourite callback button on etas
func ToggleFavouritesHandler(ctx context.Context, bot *BusEtaBot, cbq *tgbotapi.CallbackQuery, responses chan<- Response) {
	userID := cbq.From.ID

	var data CallbackData
	err := json.Unmarshal([]byte(cbq.Data), &data)
	if err != nil {
		responses <- notOk(errors.Wrap(err, fmt.Sprintf("error unmarshalling callback data: %s", cbq.Data)))
		return
	}

	favourites, err := GetUserFavourites(ctx, userID)
	if err != nil {
		responses <- notOk(errors.Wrap(err, "could not retrieve user favourites"))
		return
	}

	var action string

	// if the entry is already in the favourites, we remove it
	if exists, pos := stringInSlice(data.Argstr, favourites); exists {
		// remove item from slice
		copy(favourites[pos:], favourites[pos+1:])
		favourites[len(favourites)-1] = ""
		favourites = favourites[:len(favourites)-1]

		action = "removed from"
	} else {
		favourites = append(favourites, data.Argstr)
		action = "added to"
	}

	err = SetUserFavourites(ctx, userID, favourites)
	if err != nil {
		responses <- notOk(errors.Wrap(err, "error updating user favourites"))
	}

	text := fmt.Sprintf("Eta query `%s` %s favourites!", data.Argstr, action)
	msg := tgbotapi.NewMessage(int64(userID), text)
	msg.ParseMode = "markdown"

	if len(favourites) > 0 {
		var keyboard [][]tgbotapi.KeyboardButton
		for _, fav := range favourites {
			keyboard = append(keyboard, []tgbotapi.KeyboardButton{
				{
					Text: fav,
				},
			})
		}
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			Keyboard:       keyboard,
			ResizeKeyboard: true,
		}
	} else {
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
	}

	answer := tgbotapi.NewCallback(cbq.ID, "")

	if action == "removed_from" {
		go bot.LogEvent(ctx, cbq.From, CategoryCallback, ActionRemoveFavouriteCalback, cbq.Message.Chat.Type)
	} else {
		go bot.LogEvent(ctx, cbq.From, CategoryCallback, ActionAddFavouriteCalback, cbq.Message.Chat.Type)
	}

	err = answerCallbackQuery(bot, msg, answer)
	if err != nil {
		responses <- notOk(err)
		return
	}
}

// callbackErrorHandler is for informing the user about an error while processing a callback query.
func callbackErrorHandler(ctx context.Context, bot *BusEtaBot, cbq *tgbotapi.CallbackQuery, err error) {
	log.Errorf(ctx, "%+v", err)

	text := fmt.Sprintf("Oh no! Something went wrong. \n\nRequest ID: `%s`", appengine.RequestID(ctx))
	answer := tgbotapi.NewCallbackWithAlert(cbq.ID, text)

	_, err = bot.Telegram.AnswerCallbackQuery(answer)
	if err != nil {
		err := errors.Wrap(err, fmt.Sprintf("%#v", answer))
		log.Errorf(ctx, "%+v", err)
	}
}
