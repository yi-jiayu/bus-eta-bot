package main

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/pkg/errors"
	"github.com/yi-jiayu/telegram-bot-api"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

var callbackQueryHandlers = map[string]CallbackQueryHandler{
	"refresh":                       RefreshCallbackHandler,
	"resend":                        ResendCallbackHandler,
	"eta":                           EtaCallbackHandler,
	"eta_demo":                      EtaDemoCallbackHandler,
	"new_eta":                       NewEtaHandler,
	"no_show_redundant_eta_command": NoShowRedundantEtaCommandCallbackHandler,
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

func updateEtaMessage(ctx context.Context, bot *BusEtaBot, cbq *tgbotapi.CallbackQuery, busStopID string, serviceNos []string) error {
	text, err := EtaMessageText(ctx, bot, busStopID, serviceNos)
	if err != nil {
		return err
	}

	var reply tgbotapi.EditMessageTextConfig
	var inline bool
	if cbq.Message != nil {
		chatID := cbq.Message.Chat.ID
		messageID := cbq.Message.MessageID
		reply = tgbotapi.NewEditMessageText(chatID, messageID, text)

		inline = false
	} else {
		inlineMessageID := cbq.InlineMessageID
		reply = tgbotapi.EditMessageTextConfig{
			BaseEdit: tgbotapi.BaseEdit{
				InlineMessageID: inlineMessageID,
			},
			Text: text,
		}

		inline = true
	}

	reply.ParseMode = "markdown"

	replyMarkup, err := EtaMessageReplyMarkup(busStopID, serviceNos, inline)
	if err != nil {
		return err
	}
	reply.ReplyMarkup = replyMarkup

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
	var data CallbackData
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
	var data CallbackData
	err := json.Unmarshal([]byte(cbq.Data), &data)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error unmarshalling callback data: %s", cbq.Data))
	}

	var bsID string
	var sNos []string
	if data.Argstr != "" {
		bsID, sNos, _ = InferEtaQuery(data.Argstr)
	} else {
		bsID = data.BusStopID
		sNos = data.ServiceNos
	}

	callbackData := CallbackData{
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

	text, err := EtaMessageText(ctx, bot, "96049", nil)
	if err != nil {
		return err
	}

	reply := tgbotapi.NewMessage(chatID, text)
	reply.ParseMode = "markdown"

	replyMarkup, err := EtaMessageReplyMarkup("96049", nil, false)
	if err != nil {
		return err
	}
	reply.ReplyMarkup = replyMarkup

	if !cbq.Message.Chat.IsPrivate() {
		reply.ReplyToMessageID = cbq.Message.MessageID
	}

	answer := tgbotapi.NewCallback(cbq.ID, "Etas sent!")

	go bot.LogEvent(ctx, cbq.From, CategoryCallback, ActionEtaDemoCallback, cbq.Message.Chat.Type)

	err = answerCallbackQuery(bot, reply, answer)
	return err
}

// NewEtaHandler sends etas for a bus stop when a user taps "Get etas" on a bus stop location returned from a
// location query
func NewEtaHandler(ctx context.Context, bot *BusEtaBot, cbq *tgbotapi.CallbackQuery) error {
	chatID := cbq.Message.Chat.ID

	var data CallbackData
	err := json.Unmarshal([]byte(cbq.Data), &data)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error unmarshalling callback data: %s", cbq.Data))
	}

	bsID, sNos := data.BusStopID, data.ServiceNos

	text, err := EtaMessageText(ctx, bot, bsID, sNos)
	if err != nil {
		return err
	}

	reply := tgbotapi.NewMessage(chatID, text)
	reply.ParseMode = "markdown"

	replyMarkup, err := EtaMessageReplyMarkup(bsID, sNos, false)
	if err != nil {
		return err
	}
	reply.ReplyMarkup = replyMarkup

	if !cbq.Message.Chat.IsPrivate() {
		reply.ReplyToMessageID = cbq.Message.MessageID
	}

	answer := tgbotapi.NewCallback(cbq.ID, "Etas sent!")

	go bot.LogEvent(ctx, cbq.From, CategoryCallback, ActionEtaFromLocationCallback, cbq.Message.Chat.Type)

	err = answerCallbackQuery(bot, reply, answer)
	return err
}

// NoShowRedundantEtaCommandCallbackHandler handles the "Don't show again" callback button on the reminder sent to
// users that the /eta command isn't necessary in a private chat.
func NoShowRedundantEtaCommandCallbackHandler(ctx context.Context, bot *BusEtaBot, cbq *tgbotapi.CallbackQuery) error {
	chatID := cbq.Message.Chat.ID
	messageID := cbq.Message.MessageID
	userID := cbq.From.ID

	prefs, err := GetUserPreferences(ctx, userID)
	if err != nil {
		return err
	}

	prefs.NoRedundantEtaCommandReminder = true

	err = SetUserPreferences(ctx, userID, &prefs)
	if err != nil {
		return err
	}

	markup := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{},
	}
	edit := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, markup)

	answer := tgbotapi.NewCallback(cbq.ID, "Got it!")

	err = answerCallbackQuery(bot, edit, answer)
	return err
}

// ResendCallbackHandler handles the "Resend" button on eta results.
// todo: combine with NewEtaHandler
func ResendCallbackHandler(ctx context.Context, bot *BusEtaBot, cbq *tgbotapi.CallbackQuery) error {
	chatID := cbq.Message.Chat.ID

	var data CallbackData
	err := json.Unmarshal([]byte(cbq.Data), &data)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error unmarshalling callback data: %s", cbq.Data))
	}

	busStopID, serviceNos := data.BusStopID, data.ServiceNos

	text, err := EtaMessageText(ctx, bot, busStopID, serviceNos)
	if err != nil {
		return err
	}

	reply := tgbotapi.NewMessage(chatID, text)
	reply.ParseMode = "markdown"

	replyMarkup, err := EtaMessageReplyMarkup(busStopID, serviceNos, false)
	if err != nil {
		return err
	}
	reply.ReplyMarkup = replyMarkup

	if !cbq.Message.Chat.IsPrivate() {
		reply.ReplyToMessageID = cbq.Message.MessageID
	}

	answer := tgbotapi.NewCallback(cbq.ID, "Etas resent!")

	go bot.LogEvent(ctx, cbq.From, CategoryCallback, ActionResendCallback, cbq.Message.Chat.Type)

	err = answerCallbackQuery(bot, reply, answer)
	return errors.WithStack(err)
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
