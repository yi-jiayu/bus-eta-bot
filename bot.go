package busetabot

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/yi-jiayu/datamall/v2"
	"github.com/yi-jiayu/telegram-bot-api"
	"google.golang.org/appengine/log"
)

var handlers = Handlers{
	CommandHandlers:           commandHandlers,
	FallbackCommandHandler:    FallbackCommandHandler,
	TextHandler:               TextHandler,
	LocationHandler:           LocationHandler,
	CallbackQueryHandlers:     callbackQueryHandlers,
	InlineQueryHandler:        InlineQueryHandler,
	ChosenInlineResultHandler: ChosenInlineResultHandler,
	MessageErrorHandler:       messageErrorHandler,
	CallbackErrorHandler:      callbackErrorHandler,
}

// BusStopRepository provides bus stop information.
type BusStopRepository interface {
	Get(ID string) *BusStop
	Nearby(ctx context.Context, lat, lon, radius float64, limit int) (nearby []NearbyBusStop)
	Search(ctx context.Context, query string, limit int) []BusStop
}

type UserRepository interface {
	UpdateUserLastSeenTime(ctx context.Context, userID int, t time.Time) error
}

type BusETAs interface {
	GetBusArrival(busStopCode string, serviceNo string) (datamall.BusArrival, error)
}

// BusEtaBot contains all the bot's dependencies
type BusEtaBot struct {
	Handlers            Handlers
	Telegram            *tgbotapi.BotAPI
	Datamall            BusETAs
	StreetView          *StreetViewAPI
	MeasurementProtocol *MeasurementProtocolClient
	NowFunc             func() time.Time
	BusStops            BusStopRepository
	Users               UserRepository
}

// Handlers contains all the handlers used by the bot.
type Handlers struct {
	CommandHandlers           map[string]MessageHandler
	FallbackCommandHandler    MessageHandler
	TextHandler               MessageHandler
	LocationHandler           MessageHandler
	CallbackQueryHandlers     map[string]CallbackQueryHandler
	InlineQueryHandler        func(ctx context.Context, bot *BusEtaBot, ilq *tgbotapi.InlineQuery) error
	ChosenInlineResultHandler func(ctx context.Context, bot *BusEtaBot, cir *tgbotapi.ChosenInlineResult) error
	MessageErrorHandler       func(ctx context.Context, bot *BusEtaBot, message *tgbotapi.Message, err error)
	CallbackErrorHandler      func(ctx context.Context, bot *BusEtaBot, query *tgbotapi.CallbackQuery, err error)
}

// DefaultHandlers returns a default set of handlers.
func DefaultHandlers() Handlers {
	return handlers
}

// NewBusEtaBot creates a new Bus Eta Bot with the provided tgbotapi.BotAPI and datamall.APIClient.
func NewBusEtaBot(handlers Handlers, tg *tgbotapi.BotAPI, dm BusETAs, sv *StreetViewAPI, mp *MeasurementProtocolClient) BusEtaBot {
	bot := BusEtaBot{
		Handlers:            handlers,
		Telegram:            tg,
		Datamall:            dm,
		StreetView:          sv,
		MeasurementProtocol: mp,
	}

	bot.NowFunc = time.Now

	return bot
}

// HandleUpdate dispatches an incoming update to the corresponding handler depending on the update type
func (bot *BusEtaBot) HandleUpdate(ctx context.Context, update *tgbotapi.Update) {
	var wg sync.WaitGroup
	defer wg.Wait()

	if message := update.Message; message != nil {
		if bot.Users != nil {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := bot.Users.UpdateUserLastSeenTime(ctx, message.From.ID, time.Now())
				if err != nil {
					log.Warningf(ctx, "%+v", err)
				}
			}()
		}

		if command := message.Command(); command != "" {
			bot.handleCommand(ctx, command, message)
			return
		}

		if text := message.Text; text != "" {
			bot.handleText(ctx, message)
			return
		}

		if location := message.Location; location != nil {
			bot.handleLocation(ctx, message)
			return
		}

		return
	}

	if cbq := update.CallbackQuery; cbq != nil {
		if bot.Users != nil {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := bot.Users.UpdateUserLastSeenTime(ctx, cbq.From.ID, time.Now())
				if err != nil {
					log.Warningf(ctx, "%+v", err)
				}
			}()
		}

		if bot.Handlers.CallbackQueryHandlers != nil {
			bot.handleCallbackQuery(ctx, cbq)
		}
		return
	}

	if ilq := update.InlineQuery; ilq != nil {
		if bot.Users != nil {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := bot.Users.UpdateUserLastSeenTime(ctx, ilq.From.ID, time.Now())
				if err != nil {
					log.Warningf(ctx, "%+v", err)
				}
			}()
		}

		if bot.Handlers.InlineQueryHandler != nil {
			bot.handleInlineQuery(ctx, ilq)
		}
		return
	}

	if cir := update.ChosenInlineResult; cir != nil {
		bot.handleChosenInlineResult(ctx, cir)
		return
	}
}

func (bot *BusEtaBot) handleCommand(ctx context.Context, command string, message *tgbotapi.Message) {
	if handler, exists := bot.Handlers.CommandHandlers[command]; exists {
		err := handler(ctx, bot, message)
		if err != nil {
			messageErrorHandler(ctx, bot, message, err)
		}
	} else {
		err := bot.Handlers.FallbackCommandHandler(ctx, bot, message)
		if err != nil {
			messageErrorHandler(ctx, bot, message, err)
		}
	}
}

func (bot *BusEtaBot) handleText(ctx context.Context, message *tgbotapi.Message) {
	err := bot.Handlers.TextHandler(ctx, bot, message)
	if err != nil {
		messageErrorHandler(ctx, bot, message, err)
	}
}

func (bot *BusEtaBot) handleLocation(ctx context.Context, message *tgbotapi.Message) {
	err := bot.Handlers.LocationHandler(ctx, bot, message)
	if err != nil {
		messageErrorHandler(ctx, bot, message, err)
	}
}

func (bot *BusEtaBot) handleCallbackQuery(ctx context.Context, cbq *tgbotapi.CallbackQuery) {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(cbq.Data), &data)
	if err != nil {
		callbackErrorHandler(ctx, bot, cbq, err)
		return
	}

	if cbqType, ok := data["t"].(string); ok {
		if handler, ok := bot.Handlers.CallbackQueryHandlers[cbqType]; ok {
			err := handler(ctx, bot, cbq)
			if err != nil {
				callbackErrorHandler(ctx, bot, cbq, err)
			}
		}
	} else {
		callbackErrorHandler(ctx, bot, cbq, errors.New("unrecognised callback query"))
	}
}

func (bot *BusEtaBot) handleInlineQuery(ctx context.Context, ilq *tgbotapi.InlineQuery) {
	err := bot.Handlers.InlineQueryHandler(ctx, bot, ilq)
	if err != nil {
		log.Errorf(ctx, "%v", err)
	}
}

func (bot *BusEtaBot) handleChosenInlineResult(ctx context.Context, cir *tgbotapi.ChosenInlineResult) {
	err := bot.Handlers.ChosenInlineResultHandler(ctx, bot, cir)
	if err != nil {
		log.Errorf(ctx, "%v", err)
	}
}

// LogEvent logs an event to the Measurement Protocol if a MeasurementProtocolClient is set on the bot.
func (bot *BusEtaBot) LogEvent(ctx context.Context, user *tgbotapi.User, category, action, label string) {
	if bot.MeasurementProtocol != nil {
		_, err := bot.MeasurementProtocol.LogEvent(user.ID, user.LanguageCode, category, action, label)
		if err != nil {
			log.Errorf(ctx, "error while logging event: %v", err)
		}
	}
}
