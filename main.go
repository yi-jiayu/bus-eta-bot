package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/yi-jiayu/datamall"
	"github.com/yi-jiayu/telegram-bot-api"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
	"google.golang.org/appengine/urlfetch"
)

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

	bot.HandleUpdate(ctx, &update)
}

func initialiseDbAsync(w http.ResponseWriter, r *http.Request) {
	var env string
	if env = r.URL.Query().Get("environment"); env == "" {
		env = "dev"
	}

	ctx := appengine.NewContext(r)

	task := taskqueue.Task{
		Path:   "/initialise-db?environment=" + env,
		Method: http.MethodGet,
	}

	_, err := taskqueue.Add(ctx, &task, "")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func initialiseDb(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	accountKey := os.Getenv("DATAMALL_ACCOUNT_KEY")
	if accountKey == "" {
		log.Errorf(ctx, "DATAMALL_ACCOUNT_KEY not set")
		return
	}

	var env string
	if env = r.URL.Query().Get("environment"); env == "" {
		env = "dev"
	}

	log.Infof(ctx, "Populating bus stops...")

	err := PopulateBusStops(ctx, env, time.Now(), accountKey, datamall.DataMallEndpoint)
	if err != nil {
		log.Errorf(ctx, "error populating bus stops: %+v", err)
	}
}

func init() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/initialise-db", initialiseDb)
	http.HandleFunc("/initialise-db-async", initialiseDbAsync)

	if token := os.Getenv("TELEGRAM_BOT_TOKEN"); token != "" {
		http.HandleFunc("/"+token, webhookHandler)
	}
}
