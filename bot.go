package busetabot

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/yi-jiayu/datamall/v3"
	"github.com/yi-jiayu/telegram-bot-api"
	aelog "google.golang.org/appengine/log"

	"github.com/yi-jiayu/bus-eta-bot/v4/telegram"
)

// ResponseBufferSize is the size of the channel used to queue responses to be sent via the Telegram Bot API.
const ResponseBufferSize = 10

const MaxMessageLength = 35

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
	BusStopGetter
	Nearby(ctx context.Context, lat, lon, radius float64, limit int) (nearby []NearbyBusStop)
	Search(ctx context.Context, query string, limit int) []BusStop
}

type UserRepository interface {
	UpdateUserLastSeenTime(ctx context.Context, userID int, t time.Time) error
	GetUserFavourites(ctx context.Context, userID int) (favourites []string, err error)
	SetUserFavourites(ctx context.Context, userID int, favourites []string) error
}

type ETAService interface {
	GetBusArrival(busStopCode string, serviceNo string) (datamall.BusArrival, error)
}

type TelegramService interface {
	Do(request telegram.Request) error
}

// BusEtaBot contains all the bot's dependencies
type BusEtaBot struct {
	Handlers            Handlers
	Telegram            *tgbotapi.BotAPI
	Datamall            ETAService
	StreetView          StreetViewProvider
	MeasurementProtocol *MeasurementProtocolClient
	NowFunc             func() time.Time
	BusStops            BusStopRepository
	Users               UserRepository
	TelegramService     TelegramService
}

// Handlers contains all the handlers used by the bot.
type Handlers struct {
	CommandHandlers           map[string]CommandHandler
	FallbackCommandHandler    MessageHandler
	TextHandler               MessageHandler
	LocationHandler           MessageHandler
	CallbackQueryHandlers     map[string]CallbackQueryHandler
	InlineQueryHandler        func(ctx context.Context, bot *BusEtaBot, ilq *tgbotapi.InlineQuery) error
	ChosenInlineResultHandler func(ctx context.Context, bot *BusEtaBot, cir *tgbotapi.ChosenInlineResult) error
	MessageErrorHandler       func(ctx context.Context, bot *BusEtaBot, message *tgbotapi.Message, err error)
	CallbackErrorHandler      func(ctx context.Context, bot *BusEtaBot, query *tgbotapi.CallbackQuery, err error)
}

type Response struct {
	Request telegram.Request
	Error   error
}

func ok(r telegram.Request) Response {
	return Response{
		Request: r,
	}
}

func notOk(err error) Response {
	return Response{
		Error: err,
	}
}

// DefaultHandlers returns a default set of handlers.
func DefaultHandlers() Handlers {
	return handlers
}

// NewBot creates a new Bus Eta Bot with the provided tgbotapi.BotAPI and datamall.APIClient.
func NewBot(handlers Handlers, tg *tgbotapi.BotAPI, dm ETAService, sv *StreetViewAPI, mp *MeasurementProtocolClient) BusEtaBot {
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

// Dispatch makes requests to the Telegram Bot API for each response in responses.
func (bot *BusEtaBot) Dispatch(ctx context.Context, responses <-chan Response) {
	var wg sync.WaitGroup
	for r := range responses {
		err := r.Error
		if err != nil {
			logError(ctx, err)
		} else {
			wg.Add(1)
			go func(request telegram.Request) {
				defer wg.Done()
				err := bot.TelegramService.Do(request)
				if err != nil {
					logError(ctx, err)
				}
			}(r.Request)
		}
	}
	wg.Wait()
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
					aelog.Warningf(ctx, "%+v", err)
				}
			}()
		}
		bot.handleMessage(ctx, message)
		return
	}

	if cbq := update.CallbackQuery; cbq != nil {
		if bot.Users != nil {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := bot.Users.UpdateUserLastSeenTime(ctx, cbq.From.ID, time.Now())
				if err != nil {
					aelog.Warningf(ctx, "%+v", err)
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
					aelog.Warningf(ctx, "%+v", err)
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

func (bot *BusEtaBot) handleMessage(ctx context.Context, message *tgbotapi.Message) {
	setUserContext(ctx, strconv.Itoa(message.From.ID))

	// ignore messages longer than a certain length
	if len(message.Text) > MaxMessageLength {
		go bot.LogEvent(ctx, message.From, CategoryMessage, ActionIgnoredTextMessage, message.Chat.Type)
		aelog.Infof(ctx, "ignoring long message")
		return
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
}

func (bot *BusEtaBot) handleCommand(ctx context.Context, command string, message *tgbotapi.Message) {
	if handler, exists := bot.Handlers.CommandHandlers[command]; exists {
		responses := make(chan Response, ResponseBufferSize)
		go handler(ctx, bot, message, responses)
		bot.Dispatch(ctx, responses)
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
	setUserContext(ctx, strconv.Itoa(cbq.From.ID))

	var data map[string]interface{}
	err := json.Unmarshal([]byte(cbq.Data), &data)
	if err != nil {
		callbackErrorHandler(ctx, bot, cbq, err)
		return
	}

	if cbqType, ok := data["t"].(string); ok {
		if handler, ok := bot.Handlers.CallbackQueryHandlers[cbqType]; ok {
			responses := make(chan Response, ResponseBufferSize)
			go handler(ctx, bot, cbq, responses)
			bot.Dispatch(ctx, responses)
		}
	} else {
		callbackErrorHandler(ctx, bot, cbq, errors.New("unrecognised callback query"))
	}
}

func (bot *BusEtaBot) handleInlineQuery(ctx context.Context, ilq *tgbotapi.InlineQuery) {
	err := bot.Handlers.InlineQueryHandler(ctx, bot, ilq)
	if err != nil {
		logError(ctx, err)
	}
}

func (bot *BusEtaBot) handleChosenInlineResult(ctx context.Context, cir *tgbotapi.ChosenInlineResult) {
	err := bot.Handlers.ChosenInlineResultHandler(ctx, bot, cir)
	if err != nil {
		aelog.Errorf(ctx, "%+v", err)
	}
}

// LogEvent logs an event to the Measurement Protocol if a MeasurementProtocolClient is set on the bot.
func (bot *BusEtaBot) LogEvent(ctx context.Context, user *tgbotapi.User, category, action, label string) {
	if bot.MeasurementProtocol != nil {
		_, err := bot.MeasurementProtocol.LogEvent(user.ID, user.LanguageCode, category, action, label)
		if err != nil {
			aelog.Errorf(ctx, "error while logging event: %v", err)
		}
	}
}
