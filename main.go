package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/yi-jiayu/datamall"
	"github.com/yi-jiayu/telegram-bot-api"
	"go.opencensus.io/trace"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

var busStopRepository BusStopRepository

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World"))
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	bs, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf(ctx, "%v", err)

		// return a 200 status to all webhooks so that telegram does not redeliver them
		// w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// log update
	var pretty bytes.Buffer
	err = json.Indent(&pretty, bs, "", "  ")
	if err != nil {
		log.Errorf(ctx, "%v", err)
	}
	log.Infof(ctx, "%s", pretty.String())

	var update tgbotapi.Update
	err = json.Unmarshal(bs, &update)
	if err != nil {
		log.Errorf(ctx, "%v", err)

		// return a 200 status to all webhooks so that telegram does not redeliver them
		// w.WriteHeader(http.StatusInternalServerError)
		return
	}

	client := urlfetch.Client(ctx)

	tg := &tgbotapi.BotAPI{
		APIEndpoint: tgbotapi.APIEndpoint,
		Token:       os.Getenv("TELEGRAM_BOT_TOKEN"),
		Client:      client,
	}

	dm := &datamall.APIClient{
		Endpoint:   datamall.DataMallEndpoint,
		AccountKey: os.Getenv("DATAMALL_ACCOUNT_KEY"),
		Client:     client,
	}

	mp := NewMeasurementProtocolClientWithClient(os.Getenv("GA_TID"), client)

	sv := NewStreetViewAPI(os.Getenv("GOOGLE_API_KEY"))

	bot := NewBusEtaBot(handlers, tg, dm, &sv, &mp)
	bot.BusStops = busStopRepository

	bot.HandleUpdate(ctx, &update)
}

func init() {
	var err error
	busStopRepository, err = NewInMemoryBusStopRepositoryFromFile("data/bus_stops.json", "")
	if err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}

	http.HandleFunc("/", rootHandler)

	if token := os.Getenv("TELEGRAM_BOT_TOKEN"); token != "" {
		http.HandleFunc("/"+token, webhookHandler)
	}

	if getBotEnvironment() != devEnvironment {
		// Create and register a OpenCensus Stackdriver Trace exporter.
		exporter, err := stackdriver.NewExporter(stackdriver.Options{})
		if err != nil {
			fmt.Printf("%+v\n", err)
			os.Exit(1)
		}
		trace.RegisterExporter(exporter)
	}
}

func main() {
	appengine.Main()
}
