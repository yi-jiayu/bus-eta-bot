package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/yi-jiayu/datamall"
	"github.com/yi-jiayu/telegram-bot-api"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World"))
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf(ctx, "%v", err)

		// return a 200 status to all webhooks so that telegram does not redeliver them
		// w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// log update
	log.Infof(ctx, string(bytes))

	var update tgbotapi.Update
	err = json.Unmarshal(bytes, &update)
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

	bot.HandleUpdate(ctx, &update)
}

func addBusStopsToDatastoreHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		if r.Header.Get("Telegram-Bot-Token") != os.Getenv("TELEGRAM_BOT_TOKEN") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var busStops []BusStopJSON
		err := json.NewDecoder(r.Body).Decode(&busStops)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("%v", err)))
		}

		ctx := appengine.NewContext(r)

		n, err := PutBusStopsDatastore(ctx, busStops)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("%v", err)))
		}

		w.Write([]byte(n))
	case http.MethodPut:
	}
}

func addBusStopsToSearchHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		if r.Header.Get("Telegram-Bot-Token") != os.Getenv("TELEGRAM_BOT_TOKEN") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var busStops []BusStopJSON
		err := json.NewDecoder(r.Body).Decode(&busStops)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("%v", err)))
		}

		ctx := appengine.NewContext(r)

		n, err := PutBusStopsSearch(ctx, busStops)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("%s %v", n, err)))
		}

		w.Write([]byte(n))
	case http.MethodPut:
	}

}

func init() {
	http.HandleFunc("/", rootHandler)

	if token := os.Getenv("TELEGRAM_BOT_TOKEN"); token != "" {
		http.HandleFunc("/"+token, webhookHandler)
		http.HandleFunc("/datastore/bus-stops", addBusStopsToDatastoreHandler)
		http.HandleFunc("/search/bus-stops", addBusStopsToSearchHandler)
	}
}
