package main

import (
	"encoding/json"

	"github.com/yi-jiayu/telegram-bot-api"
	"golang.org/x/net/context"
)

var commandHandlers = map[string]MessageHandler{
	"start":   StartHandler,
	"about":   AboutHandler,
	"version": VersionHandler,
	//"feedback": feedbackHandler,
	"help":    helpHandler,
	"privacy": PrivacyHandler,
	"eta":     EtaHandler,
}

// StartHandler handles a /start command.
func StartHandler(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	firstName := message.From.FirstName

	text := "Hello " + firstName + ",\n\nBus Eta Bot is a Telegram bot which can tell you how long you have to " +
		"wait for your bus to arrive.\n\nTo get started, try sending me a bus stop code such as `96049` to " +
		"get etas for.\n\nAlternatively, you can also search for bus stops by sending me an inline query. To " +
		"try this out, type @BusEtaBot followed by a bus stop code, description or road name in any chat." +
		"\n\nThanks for trying out Bus Eta Bot! If you find Bus Eta Bot useful, do help to spread the word or " +
		"send /feedback to leave some feedback about how to help make Bus Eta Bot even better!\n\n" +
		"If you're stuck, you can send /help to view help."

	reply := tgbotapi.NewMessage(chatID, text)
	reply.ParseMode = "markdown"

	s1, s2 := `{"t":"eta_demo"}`, "Changi"
	reply.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				{
					Text:         "Get etas for bus stop 96049",
					CallbackData: &s1,
				},
			},
			{
				{
					Text: "Try an inline query",
					SwitchInlineQueryCurrentChat: &s2,
				},
			},
		},
	}

	if !message.Chat.IsPrivate() {
		messageID := message.MessageID
		reply.ReplyToMessageID = messageID
	}

	go LogEvent(ctx, message.From.ID, "command", "start", "")

	_, err := bot.Send(reply)
	if err != nil {
		return err
	}

	return nil
}

// VersionHandler handles the /version command
func VersionHandler(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	chatID := message.Chat.ID

	text := "Bus Eta Bot v" + Version + "\nhttps://github.com/yi-jiayu/bus-eta-bot-3"
	reply := tgbotapi.NewMessage(chatID, text)
	if !message.Chat.IsPrivate() {
		messageID := message.MessageID
		reply.ReplyToMessageID = messageID
	}

	go LogEvent(ctx, message.From.ID, "command", "version", "")

	_, err := bot.Send(reply)
	return err
}

// AboutHandler handles the /about command
func AboutHandler(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	chatID := message.Chat.ID

	text := "Bus Eta Bot v" + Version + "\nhttps://github.com/yi-jiayu/bus-eta-bot-3"
	reply := tgbotapi.NewMessage(chatID, text)
	if !message.Chat.IsPrivate() {
		messageID := message.MessageID
		reply.ReplyToMessageID = messageID
	}

	go LogEvent(ctx, message.From.ID, "command", "about", "")

	_, err := bot.Send(reply)
	return err
}

func feedbackHandler(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	return nil
}

func helpHandler(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	text := "Here's some help on how to use Bus Eta Bot:\nhttp://telegra.ph/Bus-Eta-Bot-Help-02-23"
	reply := tgbotapi.NewMessage(chatID, text)

	go LogEvent(ctx, message.From.ID, "command", "help", "")

	_, err := bot.Send(reply)
	return err
}

// PrivacyHandler handles the /privacy command.
func PrivacyHandler(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	chatID := message.Chat.ID

	text := "You can find Bus Eta Bot's privacy policy [here](http://telegra.ph/Bus-Eta-Bot-Privacy-Policy-03-09)."
	reply := tgbotapi.NewMessage(chatID, text)
	reply.ParseMode = "markdown"
	if !message.Chat.IsPrivate() {
		messageID := message.MessageID
		reply.ReplyToMessageID = messageID
	}

	go LogEvent(ctx, message.From.ID, "command", "privacy", "")

	_, err := bot.Send(reply)
	return err
}

// EtaHandler handles the /eta command.
func EtaHandler(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	chatID := message.Chat.ID

	if args := message.CommandArguments(); args != "" {
		busStopID, serviceNos := InferEtaQuery(args)

		if busStopID == "" {
			return nil
		}

		text, err := EtaMessage(ctx, busStopID, serviceNos)
		if err != nil {
			return err
		}

		callbackData := EtaCallbackData{
			Type:       "refresh",
			BusStopID:  busStopID,
			ServiceNos: serviceNos,
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

		if !message.Chat.IsPrivate() {
			messageID := message.MessageID
			reply.ReplyToMessageID = messageID
		}

		go LogEvent(ctx, message.From.ID, "command", "eta", args)

		_, err = bot.Send(reply)
		if err != nil {
			return err
		}
	} else {
		text := "Alright, send me a bus stop code to get etas for."
		reply := tgbotapi.NewMessage(chatID, text)
		if !message.Chat.IsPrivate() {
			messageID := message.MessageID
			reply.ReplyToMessageID = messageID
			reply.ReplyMarkup = tgbotapi.ForceReply{
				ForceReply: true,
				Selective:  true,
			}
		}

		go LogEvent(ctx, message.From.ID, "command", "eta", "")

		_, err := bot.Send(reply)
		if err != nil {
			return err
		}
	}

	return nil
}
