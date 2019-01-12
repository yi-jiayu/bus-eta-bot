package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/yi-jiayu/datamall"
	"github.com/yi-jiayu/telegram-bot-api"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

func newWebhookHandler() http.HandlerFunc {
	busStopRepository, err := NewInMemoryBusStopRepositoryFromFile("data/bus_stops.json", "")
	if err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).DialContext,
		},
		Timeout: 60 * time.Second,
	}
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

	bot := BusEtaBot{
		Handlers:            handlers,
		Telegram:            tg,
		Datamall:            dm,
		StreetView:          &sv,
		MeasurementProtocol: &mp,
		BusStops:            busStopRepository,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		bs, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Errorf(ctx, "%v", err)
			return
		}

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
			return
		}
		bot.HandleUpdate(ctx, &update)
	}
}

func init() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if len(token) > 0 {
		http.HandleFunc("/"+token, newWebhookHandler())
	}
}

func main() {
	appengine.Main()
}
