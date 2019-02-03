package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/getsentry/raven-go"
	"github.com/pkg/errors"
	"github.com/yi-jiayu/datamall/v3"
	"github.com/yi-jiayu/telegram-bot-api"
	"go.opencensus.io/trace"
	"google.golang.org/appengine"
	aelog "google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"

	"github.com/yi-jiayu/bus-eta-bot/v4"
	"github.com/yi-jiayu/bus-eta-bot/v4/telegram"
)

var BotToken = os.Getenv("TELEGRAM_BOT_TOKEN")

var (
	busStopRepository busetabot.BusStopRepository
	userRepository    busetabot.UserRepository
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World"))
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	ctx := busetabot.NewContext(r)

	bs, err := ioutil.ReadAll(r.Body)
	if err != nil {
		aelog.Errorf(ctx, "%+v", err)

		// return a 200 status to all webhooks so that telegram does not redeliver them
		// w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// log update
	var pretty bytes.Buffer
	err = json.Indent(&pretty, bs, "", "  ")
	if err != nil {
		aelog.Errorf(ctx, "%+v", err)
	}
	aelog.Infof(ctx, "%s", pretty.String())

	var update tgbotapi.Update
	err = json.Unmarshal(bs, &update)
	if err != nil {
		aelog.Errorf(ctx, "%+v", err)

		// return a 200 status to all webhooks so that telegram does not redeliver them
		// w.WriteHeader(http.StatusInternalServerError)
		return
	}

	client := urlfetch.Client(ctx)

	tg := &tgbotapi.BotAPI{
		APIEndpoint: tgbotapi.APIEndpoint,
		Token:       BotToken,
		Client:      client,
	}

	dm := datamall.NewClient(os.Getenv("DATAMALL_ACCOUNT_KEY"), client)

	mp := busetabot.NewMeasurementProtocolClientWithClient(os.Getenv("GA_TID"), client)

	sv := busetabot.NewStreetViewAPI(os.Getenv("GOOGLE_API_KEY"))

	bot := busetabot.NewBot(busetabot.DefaultHandlers(), tg, dm, &sv, &mp)
	bot.BusStops = busStopRepository
	bot.Users = userRepository

	telegramService, err := telegram.NewClient(BotToken, client)
	if err != nil {
		err = errors.Wrap(err, "error creating telegram service")
		aelog.Errorf(ctx, "%+v", err)
		return
	}
	bot.TelegramService = telegramService

	bot.HandleUpdate(ctx, &update)
}

func init() {
	var err error
	busStopRepository, err = busetabot.NewInMemoryBusStopRepositoryFromFile("data/bus_stops.json", "")
	if err != nil {
		fmt.Printf("%+v\n", err)
		raven.CaptureErrorAndWait(err, nil)
		os.Exit(1)
	}

	userRepository = new(busetabot.DatastoreUserRepository)

	http.HandleFunc("/", rootHandler)

	if token := os.Getenv("TELEGRAM_BOT_TOKEN"); token != "" {
		http.HandleFunc("/"+token, webhookHandler)
	}

	if busetabot.GetBotEnvironment() != busetabot.EnvironmentDev {
		// Create and register a OpenCensus Stackdriver Trace exporter.
		exporter, err := stackdriver.NewExporter(stackdriver.Options{})
		if err != nil {
			log.Printf("error setting up opencensus stackdriver exporter: %+v\n", err)
			raven.CaptureError(err, nil)
		} else {
			trace.RegisterExporter(exporter)
			trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
		}
	}
}

func main() {
	appengine.Main()
}
