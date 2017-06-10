package main

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/yi-jiayu/telegram-bot-api"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

// Links to relevant documents
const (
	RepoURL          = "https://github.com/yi-jiayu/bus-eta-bot"
	PrivacyPolicyURL = "https://t.me/iv?url=https%3A%2F%2Fgithub.com%2Fyi-jiayu%2Fbus-eta-bot%2Fblob%2Fmaster%2FPRIVACY.md&rhash=a44cb5372834ee"
	HelpURL          = "http://telegra.ph/Bus-Eta-Bot-Help-02-23"
)

var (
	busStopRegex = regexp.MustCompile(`\d{5}`)
)

var commandHandlers = map[string]MessageHandler{
	"start":   StartHandler,
	"about":   AboutHandler,
	"version": VersionHandler,
	//"feedback": feedbackHandler,
	"help":    HelpHandler,
	"privacy": PrivacyHandler,
	"eta":     EtaHandler,
}

// FallbackCommandHandler catches commands which don't match any other handler.
func FallbackCommandHandler(ctx context.Context, bot *BusEtaBot, message *tgbotapi.Message) error {
	chatID := message.Chat.ID

	var reply tgbotapi.MessageConfig
	if busStopRegex.MatchString(message.Command()) {
		text := fmt.Sprintf("Oops, that was not a valid command! If you wanted to get etas for bus "+
			"stop %s, just send the bus stop code without the leading slash.", message.Command())
		reply = tgbotapi.NewMessage(chatID, text)
	} else {
		text := "Oops, that was not a valid command!"
		reply = tgbotapi.NewMessage(chatID, text)
	}

	if !message.Chat.IsPrivate() {
		messageID := message.MessageID
		reply.ReplyToMessageID = messageID
	}

	_, err := bot.Telegram.Send(reply)
	return err
}

// StartHandler handles a /start command.
func StartHandler(ctx context.Context, bot *BusEtaBot, message *tgbotapi.Message) error {
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

	s1, s2 := `{"t":"eta_demo"}`, "Tropicana"
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

	go bot.LogEvent(ctx, message.From, CategoryCommand, ActionStartCommand, message.Chat.Type)

	_, err := bot.Telegram.Send(reply)
	if err != nil {
		return err
	}

	return nil
}

func aboutMessage(message *tgbotapi.Message) tgbotapi.MessageConfig {
	chatID := message.Chat.ID

	text := "Bus Eta Bot " + Version + "\n" + RepoURL
	reply := tgbotapi.NewMessage(chatID, text)
	if !message.Chat.IsPrivate() {
		messageID := message.MessageID
		reply.ReplyToMessageID = messageID
	}

	return reply
}

// VersionHandler handles the /version command
func VersionHandler(ctx context.Context, bot *BusEtaBot, message *tgbotapi.Message) error {
	go bot.LogEvent(ctx, message.From, CategoryCommand, ActionVersionCommand, message.Chat.Type)

	reply := aboutMessage(message)
	_, err := bot.Telegram.Send(reply)
	return err
}

// AboutHandler handles the /about command
func AboutHandler(ctx context.Context, bot *BusEtaBot, message *tgbotapi.Message) error {
	go bot.LogEvent(ctx, message.From, CategoryCommand, ActionAboutCommand, message.Chat.Type)

	reply := aboutMessage(message)
	_, err := bot.Telegram.Send(reply)
	return err
}

func feedbackHandler(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	return nil
}

// HelpHandler handles the /help command
func HelpHandler(ctx context.Context, bot *BusEtaBot, message *tgbotapi.Message) error {
	chatID := message.Chat.ID

	text := fmt.Sprintf("You can find help on how to use Bus Eta Bot [here](%s).", HelpURL)
	reply := tgbotapi.NewMessage(chatID, text)
	reply.ParseMode = "markdown"
	if !message.Chat.IsPrivate() {
		messageID := message.MessageID
		reply.ReplyToMessageID = messageID
	}

	go bot.LogEvent(ctx, message.From, CategoryCommand, ActionHelpCommand, message.Chat.Type)

	_, err := bot.Telegram.Send(reply)
	return err
}

// PrivacyHandler handles the /privacy command.
func PrivacyHandler(ctx context.Context, bot *BusEtaBot, message *tgbotapi.Message) error {
	chatID := message.Chat.ID

	text := fmt.Sprintf("You can find Bus Eta Bot's privacy policy [here](%s).", PrivacyPolicyURL)
	reply := tgbotapi.NewMessage(chatID, text)
	reply.ParseMode = "markdown"
	if !message.Chat.IsPrivate() {
		messageID := message.MessageID
		reply.ReplyToMessageID = messageID
	}

	go bot.LogEvent(ctx, message.From, CategoryCommand, ActionPrivacyCommand, message.Chat.Type)

	_, err := bot.Telegram.Send(reply)
	return err
}

// EtaHandler handles the /eta command.
func EtaHandler(ctx context.Context, bot *BusEtaBot, message *tgbotapi.Message) error {
	chatID := message.Chat.ID

	if message.Chat.IsPrivate() {
		go func() {
			prefs, err := GetUserPreferences(ctx, message.From.ID)
			if err != nil {
				log.Errorf(ctx, "%v", err)
				return
			}

			if prefs.NoRedundantEtaCommandReminder {
				return
			}

			callbackData := CallbackData{
				Type: "no_show_redundant_eta_command",
			}
			callbackDataJSON, err := json.Marshal(callbackData)
			if err != nil {
				log.Errorf(ctx, "%v", err)
				return
			}
			callbackDataJSONStr := string(callbackDataJSON)

			text := "Did you know that in a private chat, you can just send a bus stop code directly, without using the /eta command?"
			info := tgbotapi.NewMessage(chatID, text)
			info.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
				InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
					{
						tgbotapi.InlineKeyboardButton{
							Text:         "Don't show again",
							CallbackData: &callbackDataJSONStr,
						},
					},
				},
			}

			_, err = bot.Telegram.Send(info)
			if err != nil {
				log.Errorf(ctx, "%v", err)
			}
		}()
	}

	var reply tgbotapi.MessageConfig
	if args := message.CommandArguments(); args != "" {
		busStopID, serviceNos, err := InferEtaQuery(args)
		if err == errBusStopIDTooLong {
			reply = tgbotapi.NewMessage(chatID, "Oops, a bus stop code can only contain a maximum of 5 characters.")
		} else if err == errBusStopIDInvalid {
			reply = tgbotapi.NewMessage(chatID, "Oops, that did not seem to be a valid bus stop code.")
		} else if err != nil {
			return err
		} else {
			text, err := EtaMessage(ctx, bot, busStopID, serviceNos)
			if err != nil {
				if err == errNotFound {
					reply = tgbotapi.NewMessage(chatID, text)
				} else {
					return err
				}
			} else {
				callbackData := CallbackData{
					Type:       "refresh",
					BusStopID:  busStopID,
					ServiceNos: serviceNos,
				}

				callbackDataJSON, err := json.Marshal(callbackData)
				if err != nil {
					return err
				}
				callbackDataJSONStr := string(callbackDataJSON)

				reply = tgbotapi.NewMessage(chatID, text)
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
			}
		}

		if !message.Chat.IsPrivate() {
			reply.ReplyToMessageID = message.MessageID
		}

		go bot.LogEvent(ctx, message.From, CategoryCommand, ActionEtaCommandWithArgs, message.Chat.Type)

		_, err = bot.Telegram.Send(reply)
		return err
	}

	text := "Alright, send me a bus stop code to get etas for."
	reply = tgbotapi.NewMessage(chatID, text)
	if !message.Chat.IsPrivate() {
		reply.ReplyToMessageID = message.MessageID
	}
	reply.ReplyMarkup = tgbotapi.ForceReply{
		ForceReply: true,
		Selective:  true,
	}

	go bot.LogEvent(ctx, message.From, CategoryCommand, ActionEtaCommandWithoutArgs, message.Chat.Type)

	_, err := bot.Telegram.Send(reply)
	return err
}
