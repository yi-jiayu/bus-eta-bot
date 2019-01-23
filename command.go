package busetabot

import (
	"context"
	"fmt"
	"regexp"

	"github.com/yi-jiayu/telegram-bot-api"
	"google.golang.org/appengine/log"

	"github.com/yi-jiayu/bus-eta-bot/v4/telegram"
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

var commandHandlers = map[string]CommandHandler{
	"start":          StartHandler,
	"about":          AboutHandler,
	"version":        VersionHandler,
	"feedback":       FeedbackCmdHandler,
	"help":           HelpHandler,
	"privacy":        PrivacyHandler,
	"eta":            EtaHandler,
	"favourites":     ShowFavouritesCmdHandler,
	"favorites":      ShowFavouritesCmdHandler,
	"showfavourites": ShowFavouritesCmdHandler,
	"showfavorites":  ShowFavouritesCmdHandler,
	"hidefavourites": HideFavouritesCmdHandler,
	"hidefavorites":  HideFavouritesCmdHandler,
}

// CommandHandler is a handler for incoming commands.
type CommandHandler func(ctx context.Context, bot *BusEtaBot, message *tgbotapi.Message, responses chan<- Response)

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
func StartHandler(ctx context.Context, bot *BusEtaBot, message *tgbotapi.Message, responses chan<- Response) {
	go bot.LogEvent(ctx, message.From, CategoryCommand, ActionStartCommand, message.Chat.Type)

	text := "Hello " + message.From.FirstName + ",\n\nBus Eta Bot is a Telegram bot which can tell you how long you have to " +
		"wait for your bus to arrive.\n\nTo get started, try sending me a bus stop code such as `96049` to " +
		"get etas for.\n\nAlternatively, you can also search for bus stops by sending me an inline query. To " +
		"try this out, type @BusEtaBot followed by a bus stop code, description or road name in any chat." +
		"\n\nThanks for trying out Bus Eta Bot! If you find Bus Eta Bot useful, do help to spread the word or " +
		"send /feedback to leave some feedback about how to help make Bus Eta Bot even better!\n\n" +
		"If you're stuck, you can send /help to view help."
	query := "SUTD"
	request := telegram.SendMessageRequest{
		ChatID:    message.Chat.ID,
		Text:      text,
		ParseMode: "markdown",
		ReplyMarkup: &telegram.InlineKeyboardMarkup{
			InlineKeyboard: [][]telegram.InlineKeyboardButton{
				{
					{
						Text:         "Get etas for bus stop 96049",
						CallbackData: `{"t":"eta_demo"}`,
					},
					{
						Text:                         "Try an inline query",
						SwitchInlineQueryCurrentChat: &query,
					},
				},
			},
		},
	}
	responses <- ok(request)
	close(responses)
}

// VersionHandler handles the /version command
func VersionHandler(ctx context.Context, bot *BusEtaBot, message *tgbotapi.Message, responses chan<- Response) {
	go bot.LogEvent(ctx, message.From, CategoryCommand, ActionAboutCommand, message.Chat.Type)

	request := telegram.SendMessageRequest{
		ChatID: message.Chat.ID,
		Text:   "Bus Eta Bot " + Version + "\n" + RepoURL,
	}
	responses <- ok(request)
	close(responses)
}

// AboutHandler handles the /about command
func AboutHandler(ctx context.Context, bot *BusEtaBot, message *tgbotapi.Message, responses chan<- Response) {
	go bot.LogEvent(ctx, message.From, CategoryCommand, ActionAboutCommand, message.Chat.Type)

	request := telegram.SendMessageRequest{
		ChatID: message.Chat.ID,
		Text:   "Bus Eta Bot " + Version + "\n" + RepoURL,
	}
	responses <- ok(request)
	close(responses)
}

// FeedbackCmdHandler handles the /feedback command.
func FeedbackCmdHandler(ctx context.Context, bot *BusEtaBot, message *tgbotapi.Message, responses chan<- Response) {
	go bot.LogEvent(ctx, message.From, CategoryCommand, ActionFeedbackCommand, message.Chat.Type)

	request := telegram.SendMessageRequest{
		ChatID:    message.Chat.ID,
		Text:      fmt.Sprintf("Oops, the feedback command has not been implemented yet. In the meantime, you can raise issues or show your support for Bus Eta Bot at its GitHub repository [here](%s).", RepoURL),
		ParseMode: "markdown",
	}
	responses <- ok(request)
	close(responses)
}

// HelpHandler handles the /help command
func HelpHandler(ctx context.Context, bot *BusEtaBot, message *tgbotapi.Message, responses chan<- Response) {
	go bot.LogEvent(ctx, message.From, CategoryCommand, ActionHelpCommand, message.Chat.Type)

	request := telegram.SendMessageRequest{
		ChatID:    message.Chat.ID,
		Text:      fmt.Sprintf("You can find help on how to use Bus Eta Bot [here](%s).", HelpURL),
		ParseMode: "markdown",
	}
	responses <- ok(request)
	close(responses)
}

// PrivacyHandler handles the /privacy command.
func PrivacyHandler(ctx context.Context, bot *BusEtaBot, message *tgbotapi.Message, responses chan<- Response) {
	go bot.LogEvent(ctx, message.From, CategoryCommand, ActionPrivacyCommand, message.Chat.Type)

	request := telegram.SendMessageRequest{
		ChatID:    message.Chat.ID,
		Text:      fmt.Sprintf("You can find Bus Eta Bot's privacy policy [here](%s).", PrivacyPolicyURL),
		ParseMode: "markdown",
	}
	responses <- ok(request)
	close(responses)
}

// EtaHandler handles the /eta command.
func EtaHandler(ctx context.Context, bot *BusEtaBot, message *tgbotapi.Message, responses chan<- Response) {
	chatID := message.Chat.ID

	var reply tgbotapi.MessageConfig
	if args := message.CommandArguments(); args != "" {
		busStopID, serviceNos, err := InferEtaQuery(args)
		if err == errBusStopIDTooLong {
			reply = tgbotapi.NewMessage(chatID, "Oops, a bus stop code can only contain a maximum of 5 characters.")
		} else if err == errBusStopIDInvalid {
			reply = tgbotapi.NewMessage(chatID, "Oops, that did not seem to be a valid bus stop code.")
		} else if err != nil {
			log.Errorf(ctx, "%+v", err)
		} else {
			text, err := EtaMessageText(bot, busStopID, serviceNos)
			if err != nil {
				if err == errNotFound {
					reply = tgbotapi.NewMessage(chatID, text)
				} else {
					log.Errorf(ctx, "%+v", err)
				}
			} else {
				reply = tgbotapi.NewMessage(chatID, text)
				reply.ParseMode = "markdown"

				replyMarkup, err := EtaMessageReplyMarkup(busStopID, serviceNos, false)
				if err != nil {
					log.Errorf(ctx, "%+v", err)
				}
				reply.ReplyMarkup = replyMarkup
			}
		}

		if !message.Chat.IsPrivate() {
			reply.ReplyToMessageID = message.MessageID
		}

		go bot.LogEvent(ctx, message.From, CategoryCommand, ActionEtaCommandWithArgs, message.Chat.Type)

		_, err = bot.Telegram.Send(reply)
		if err != nil {
			log.Errorf(ctx, "%+v", err)
		}
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
	if err != nil {
		log.Errorf(ctx, "%+v", err)
	}
}

func showFavourites(bot *BusEtaBot, message *tgbotapi.Message, favourites []string) error {
	chatID := message.Chat.ID

	var reply tgbotapi.MessageConfig
	if len(favourites) == 0 {
		reply = tgbotapi.NewMessage(chatID, "Oops, you haven't saved any favourites yet.")
	} else {
		var keyboard [][]tgbotapi.KeyboardButton
		for _, fav := range favourites {
			keyboard = append(keyboard, []tgbotapi.KeyboardButton{
				{
					Text: fav,
				},
			})
		}

		reply = tgbotapi.NewMessage(chatID, "Favourites keyboard activated.")
		reply.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			Keyboard:       keyboard,
			ResizeKeyboard: true,
		}
	}

	if !message.Chat.IsPrivate() {
		reply.ReplyToMessageID = message.MessageID
	}

	_, err := bot.Telegram.Send(reply)
	return err
}

// ShowFavouritesCmdHandler will display a reply keyboard for quick access to the user's favourites.
func ShowFavouritesCmdHandler(ctx context.Context, bot *BusEtaBot, message *tgbotapi.Message, responses chan<- Response) {
	userID := message.From.ID

	favourites, err := GetUserFavourites(ctx, userID)
	if err != nil {
		log.Errorf(ctx, "%+v", err)
	}

	err = showFavourites(bot, message, favourites)
	if err != nil {
		responses <- notOk(err)
	}
	close(responses)
}

// HideFavouritesCmdHandler hides the favourites keyboard.
func HideFavouritesCmdHandler(ctx context.Context, bot *BusEtaBot, message *tgbotapi.Message, responses chan<- Response) {
	chatID := message.Chat.ID

	reply := tgbotapi.NewMessage(chatID, "Favourites keyboard hidden.")
	reply.ReplyMarkup = tgbotapi.ReplyKeyboardRemove{
		RemoveKeyboard: true,
	}

	_, err := bot.Telegram.Send(reply)
	if err != nil {
		responses <- notOk(err)
	}
	close(responses)
}

// StreetviewCmdHandler handlers the /streetview command.
// func StreetviewCmdHandler(ctx context.Context, bot *BusEtaBot, message *tgbotapi.Message) error {
// 	chatID := message.Chat.ID
//
// 	var reply tgbotapi.Chattable
// 	if args := message.CommandArguments(); args != "" {
// 		busStopID, _, err := InferEtaQuery(args)
// 		if err != nil {
// 			if err == errBusStopIDTooLong {
// 				reply = tgbotapi.NewMessage(chatID, "Oops, a bus stop code can only contain a maximum of 5 characters.")
// 			} else if err == errBusStopIDInvalid {
// 				reply = tgbotapi.NewMessage(chatID, "Oops, that did not seem to be a valid bus stop code.")
// 			} else {
// 				return errors.Wrap(err, "failed to infer eta query")
// 			}
// 		}
//
// 		busStop, err := GetBusStop(ctx, busStopID)
// 		if err != nil {
// 			return err
// 		}
//
// 		var imageURL string
// 		if bot.StreetView != nil {
// 			if lat, lon := busStop.Location.Lat, busStop.Location.Lng; lat != 0 && lon != 0 {
// 				URL, err := bot.StreetView.GetPhotoURLByLocation(lat, lon, 640, 480)
// 				if err != nil {
// 					log.Errorf(ctx, "%+v", err)
// 				} else {
// 					imageURL = URL
// 					reply = tgbotapi.NewPhotoShare()
// 				}
// 			} else {
// 				reply = tgbotapi.NewMessage(chatID, "Oops, couldn't find any imagery for that bus stop.")
// 			}
// 		} else {
// 			reply = tgbotapi.NewMessage(chatID, "Oops, couldn't find any imagery for that bus stop.")
// 		}
//
// 	}
//
// }
