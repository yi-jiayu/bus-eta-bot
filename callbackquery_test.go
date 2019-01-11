package main

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yi-jiayu/telegram-bot-api"

	"google.golang.org/appengine/aetest"
)

func TestNoShowRedundantEtaCommandCallbackHandler(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	tgAPI, reqChan, errChan := NewMockTelegramAPIWithPath()
	defer tgAPI.Close()

	tg := &tgbotapi.BotAPI{
		APIEndpoint: tgAPI.URL + "/bot%s/%s",
		Client:      http.DefaultClient,
	}

	bot := NewBusEtaBot(handlers, tg, nil, nil, nil)

	message := MockMessageWithType(ChatTypePrivate)

	cbq := MockCallbackQuery()
	cbq.Message = &message
	cbq.Data = `{"t":"no_show_redundant_eta_command"}`

	err = NoShowRedundantEtaCommandCallbackHandler(ctx, &bot, &cbq)
	if err != nil {
		t.Fatal(err)
	}

	var reqs []Request
	timer := time.NewTimer(10 * time.Second)

	for len(reqs) < 2 {
		select {
		case req := <-reqChan:
			reqs = append(reqs, req)
		case err := <-errChan:
			t.Fatal(err)
		case <-timer.C:
			t.Fatal("timed out!")
		}
	}

	actual := reqs
	expected := []Request{
		{Path: "/bot/answerCallbackQuery", Body: "cache_time=0&callback_query_id=1&show_alert=false&text=Got+it%21"},
		{Path: "/bot/editMessageReplyMarkup", Body: "chat_id=1&message_id=1&reply_markup=%7B%22inline_keyboard%22%3A%5B%5D%7D"},
	}

	assert.ElementsMatch(t, expected, actual)

	prefs, err := GetUserPreferences(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}

	if !prefs.NoRedundantEtaCommandReminder {
		fmt.Println("Expected prefs.NoRedundantEtaCommandHandler to be true.")
		t.Fail()
	}
}
