package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/yi-jiayu/telegram-bot-api"
)

type Request struct {
	Path string
	Body string
}

func NewMockTelegramAPIWithPath() (*httptest.Server, chan Request, chan error) {
	reqChan := make(chan Request, 2)
	errChan := make(chan error, 2)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				errChan <- err
				return
			}

			reqChan <- Request{r.URL.Path, string(body)}
		}

		w.Write([]byte(`{"ok":true}`))
	}))

	return ts, reqChan, errChan
}

func NewMockBusArrivalAPI(t time.Time) (*httptest.Server, error) {
	busArrival, err := json.Marshal(newArrival(t))
	if err != nil {
		return nil, err
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(busArrival)
	}))

	return ts, nil
}

func MockMessage() tgbotapi.Message {
	return tgbotapi.Message{
		Chat: &tgbotapi.Chat{ID: 1},
		From: &tgbotapi.User{ID: 1, FirstName: "Jiayu"},
	}
}

func MockCallbackQuery() tgbotapi.CallbackQuery {
	return tgbotapi.CallbackQuery{
		ID: "1",
		From: &tgbotapi.User{
			ID:        1,
			FirstName: "Jiayu",
		},
	}
}

func MockInlineQuery() tgbotapi.InlineQuery {
	return tgbotapi.InlineQuery{
		ID: "1",
		From: &tgbotapi.User{
			ID:        1,
			FirstName: "Jiayu",
		},
	}
}

func TestInferEtaQuery(t *testing.T) {
	t.Parallel()

	t.Run("Bus stop ID only", func(t *testing.T) {
		query := "96049"

		busStopID, serviceNos := InferEtaQuery(query)
		actual := struct {
			BusStopID  string
			ServiceNos []string
		}{
			busStopID,
			serviceNos,
		}
		expected := struct {
			BusStopID  string
			ServiceNos []string
		}{
			"96049",
			[]string{},
		}

		if !reflect.DeepEqual(actual, expected) {
			fmt.Printf("Expected:\n%#v\nActual:\n%#v\n", expected, actual)
			t.Fail()
		}
	})
	t.Run("Bus stop ID and services", func(t *testing.T) {
		query := "96049 2 24"

		busStopID, serviceNos := InferEtaQuery(query)
		actual := struct {
			BusStopID  string
			ServiceNos []string
		}{
			busStopID,
			serviceNos,
		}
		expected := struct {
			BusStopID  string
			ServiceNos []string
		}{
			"96049",
			[]string{"2", "24"},
		}

		if !reflect.DeepEqual(actual, expected) {
			fmt.Printf("Expected:\n%#v\nActual:\n%#v\n", expected, actual)
			t.Fail()
		}
	})
}
