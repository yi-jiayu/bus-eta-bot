package busetabot

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yi-jiayu/telegram-bot-api"

	"github.com/yi-jiayu/bus-eta-bot/v4/telegram"
)

// collectResponsesWithTimeout returns a slice of responses received from the provided channel. If nothing is received
// within timeout, it returns the responses received up until that point.
func collectResponsesWithTimeout(responses <-chan Response, timeout time.Duration) (collected []Response, err error) {
	for {
		deadline := time.NewTimer(timeout)
		select {
		case r, ok := <-responses:
			if !ok {
				return
			}
			collected = append(collected, r)
		case <-deadline.C:
			err = errors.New("timed out")
			return
		}
		deadline.Stop()
	}
}

func TestFallbackCommandHandler(t *testing.T) {
	t.Parallel()

	tgAPI, reqChan, errChan := NewMockTelegramAPIWithPath()
	defer tgAPI.Close()

	tg := &tgbotapi.BotAPI{
		APIEndpoint: tgAPI.URL + "/bot%s/%s",
		Client:      http.DefaultClient,
	}

	bot := NewBusEtaBot(handlers, tg, nil, nil, nil)

	testCases := []struct {
		Name     string
		ChatType string
		Text     string
		Expected Request
	}{
		{
			Name:     "Bus stop code with slash in front, private chat",
			ChatType: ChatTypePrivate,
			Text:     "/96049",
			Expected: Request{
				Path: "/bot/sendMessage",
				Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&text=Oops%2C+that+was+not+a+valid+command%21+If+you+wanted+to+get+etas+for+bus+stop+96049%2C+just+send+the+bus+stop+code+without+the+leading+slash.",
			},
		},
		{
			Name: "Bus stop code with slash in front, group chat",
			Text: "/96049",
			Expected: Request{
				Path: "/bot/sendMessage",
				Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&reply_to_message_id=1&text=Oops%2C+that+was+not+a+valid+command%21+If+you+wanted+to+get+etas+for+bus+stop+96049%2C+just+send+the+bus+stop+code+without+the+leading+slash.",
			},
		},
		{
			Name: "Invalid command, private chat",
			Text: "/invalid",
			Expected: Request{
				Path: "/bot/sendMessage",
				Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&reply_to_message_id=1&text=Oops%2C+that+was+not+a+valid+command%21",
			},
		},
		{
			Name: "Invalid command, group chat",
			Text: "/invalid",
			Expected: Request{
				Path: "/bot/sendMessage",
				Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&reply_to_message_id=1&text=Oops%2C+that+was+not+a+valid+command%21",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			message := MockMessageWithType(tc.ChatType)
			message.Text = tc.Text

			err := FallbackCommandHandler(nil, &bot, &message)
			if err != nil {
				t.Fatal(err)
			}

			select {
			case req := <-reqChan:
				actual := req
				expected := tc.Expected

				if actual != expected {
					fmt.Printf("Expected:\n%#v\nActual:\n%#v\n", expected, actual)
					t.Fail()
				}
			case err := <-errChan:
				t.Fatal(err)
			}
		})
	}
}

func TestEtaHandler(t *testing.T) {
	type testCase struct {
		Name     string
		ChatType string
		Text     string
		Expected []Response
	}
	testCases := []testCase{
		{
			Name:     "without arguments, in private chat",
			ChatType: "private",
			Text:     "/eta",
			Expected: []Response{
				ok(telegram.SendMessageRequest{
					ChatID:      1,
					Text:        "Alright, send me a bus stop code to get etas for.",
					ReplyMarkup: telegram.NewForceReply(true),
				}),
			},
		},
		{
			Name:     "without arguments, in group chat",
			ChatType: "group",
			Text:     "/eta",
			Expected: []Response{
				ok(telegram.SendMessageRequest{
					ChatID:           1,
					Text:             "Alright, send me a bus stop code to get etas for.",
					ReplyToMessageID: 1,
					ReplyMarkup:      telegram.NewForceReply(true),
				}),
			},
		},
	}
	bot := new(BusEtaBot)
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			message := MockMessageWithType(tc.ChatType)
			message.Text = tc.Text
			responses := make(chan Response, ResponseBufferSize)
			go EtaHandler(context.Background(), bot, &message, responses)
			actual, err := collectResponsesWithTimeout(responses, 5*time.Second)
			if err != nil {
				t.Fatal(err)
			}
			expected := tc.Expected
			assert.Equal(t, expected, actual)
		})
	}
}

func TestAboutHandler(t *testing.T) {
	bot := new(BusEtaBot)
	message := MockMessage()
	responses := make(chan Response, ResponseBufferSize)
	go AboutHandler(context.Background(), bot, &message, responses)
	actual, err := collectResponsesWithTimeout(responses, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	expected := []Response{
		ok(telegram.SendMessageRequest{
			ChatID: 1,
			Text:   "Bus Eta Bot VERSION\nhttps://github.com/yi-jiayu/bus-eta-bot",
		}),
	}
	assert.Equal(t, expected, actual)
}

func TestVersionHandler(t *testing.T) {
	bot := new(BusEtaBot)
	message := MockMessage()
	responses := make(chan Response, ResponseBufferSize)
	go VersionHandler(context.Background(), bot, &message, responses)
	actual, err := collectResponsesWithTimeout(responses, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	expected := []Response{
		ok(telegram.SendMessageRequest{
			ChatID: 1,
			Text:   "Bus Eta Bot VERSION\nhttps://github.com/yi-jiayu/bus-eta-bot",
		}),
	}
	assert.Equal(t, expected, actual)
}

func TestStartHandler(t *testing.T) {
	bot := new(BusEtaBot)
	message := MockMessage()
	responses := make(chan Response, ResponseBufferSize)
	go StartHandler(context.Background(), bot, &message, responses)
	actual, err := collectResponsesWithTimeout(responses, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	s := "SUTD"
	expected := []Response{
		ok(telegram.SendMessageRequest{
			ChatID:    1,
			Text:      "Hello Jiayu,\n\nBus Eta Bot is a Telegram bot which can tell you how long you have to wait for your bus to arrive.\n\nTo get started, try sending me a bus stop code such as `96049` to get etas for.\n\nAlternatively, you can also search for bus stops by sending me an inline query. To try this out, type @BusEtaBot followed by a bus stop code, description or road name in any chat.\n\nThanks for trying out Bus Eta Bot! If you find Bus Eta Bot useful, do help to spread the word or send /feedback to leave some feedback about how to help make Bus Eta Bot even better!\n\nIf you're stuck, you can send /help to view help.",
			ParseMode: "markdown",
			ReplyMarkup: &telegram.InlineKeyboardMarkup{
				InlineKeyboard: [][]telegram.InlineKeyboardButton{
					{
						{
							Text:         "Get etas for bus stop 96049",
							CallbackData: "{\"t\":\"eta_demo\"}",
						},
						{
							Text:                         "Try an inline query",
							CallbackData:                 "",
							SwitchInlineQueryCurrentChat: &s,
						},
					},
				},
			},
		}),
	}
	assert.Equal(t, expected, actual)
}

func TestHelpHandler(t *testing.T) {
	bot := new(BusEtaBot)
	message := MockMessage()
	responses := make(chan Response, ResponseBufferSize)
	go HelpHandler(context.Background(), bot, &message, responses)
	actual, err := collectResponsesWithTimeout(responses, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	expected := []Response{
		ok(telegram.SendMessageRequest{
			ChatID:    1,
			Text:      "You can find help on how to use Bus Eta Bot [here](http://telegra.ph/Bus-Eta-Bot-Help-02-23).",
			ParseMode: "markdown",
		}),
	}
	assert.Equal(t, expected, actual)
}

func TestPrivacyHandler(t *testing.T) {
	bot := new(BusEtaBot)
	message := MockMessage()
	responses := make(chan Response, ResponseBufferSize)
	PrivacyHandler(context.Background(), bot, &message, responses)
	actual, err := collectResponsesWithTimeout(responses, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	expected := []Response{
		ok(telegram.SendMessageRequest{
			ChatID:    1,
			Text:      "You can find Bus Eta Bot's privacy policy [here](https://t.me/iv?url=https%3A%2F%2Fgithub.com%2Fyi-jiayu%2Fbus-eta-bot%2Fblob%2Fmaster%2FPRIVACY.md&rhash=a44cb5372834ee).",
			ParseMode: "markdown",
		}),
	}
	assert.Equal(t, expected, actual)
}

func TestFeedbackCmdHandler(t *testing.T) {
	bot := new(BusEtaBot)
	message := MockMessage()
	responses := make(chan Response, ResponseBufferSize)
	FeedbackCmdHandler(context.Background(), bot, &message, responses)
	actual, err := collectResponsesWithTimeout(responses, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	expected := []Response{
		ok(telegram.SendMessageRequest{
			ChatID:      1,
			Text:        "Oops, the feedback command has not been implemented yet. In the meantime, you can raise issues or show your support for Bus Eta Bot at its GitHub repository [here](https://github.com/yi-jiayu/bus-eta-bot).",
			ParseMode:   "markdown",
			ReplyMarkup: nil,
		}),
	}
	assert.Equal(t, expected, actual)
}

func TestShowFavouritesCmdHandler(t *testing.T) {
	t.Parallel()

	type TestCase struct {
		Name       string
		Favourites []string
		Expected   Request
	}

	cases := []TestCase{
		{
			Name:       "No favourites saved",
			Favourites: nil,
			Expected:   Request{Path: "/bot/sendMessage", Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&text=Oops%2C+you+haven%27t+saved+any+favourites+yet."},
		},
		{
			Name: "Show favourites keyboard",
			Favourites: []string{
				"96049 2 24",
				"83062 2 24",
			},
			Expected: Request{Path: "/bot/sendMessage", Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&reply_markup=%7B%22keyboard%22%3A%5B%5B%7B%22text%22%3A%2296049+2+24%22%2C%22request_contact%22%3Afalse%2C%22request_location%22%3Afalse%7D%5D%2C%5B%7B%22text%22%3A%2283062+2+24%22%2C%22request_contact%22%3Afalse%2C%22request_location%22%3Afalse%7D%5D%5D%2C%22resize_keyboard%22%3Atrue%2C%22one_time_keyboard%22%3Afalse%2C%22selective%22%3Afalse%7D&text=Favourites+keyboard+activated."},
		},
	}

	tgAPI, reqChan, errChan := NewMockTelegramAPIWithPath()
	defer tgAPI.Close()

	tg := &tgbotapi.BotAPI{
		APIEndpoint: tgAPI.URL + "/bot%s/%s",
		Client:      http.DefaultClient,
	}

	bot := NewBusEtaBot(handlers, tg, nil, nil, nil)

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			message := MockMessage()

			err := showFavourites(&bot, &message, tc.Favourites)
			if err != nil {
				t.Fatalf("%v", err)
			}

			select {
			case req := <-reqChan:
				actual := req
				expected := tc.Expected

				if actual != expected {
					fmt.Printf("Expected:\n%#v\nActual:\n%#v\n", expected, actual)
					t.Fail()
				}
			case err := <-errChan:
				t.Fatalf("%v", err)
			}
		})
	}
}

func TestHideFavouritesCmdHandler(t *testing.T) {
	t.Parallel()

	tgAPI, reqChan, errChan := NewMockTelegramAPIWithPath()
	defer tgAPI.Close()

	tg := &tgbotapi.BotAPI{
		APIEndpoint: tgAPI.URL + "/bot%s/%s",
		Client:      http.DefaultClient,
	}

	bot := NewBusEtaBot(handlers, tg, nil, nil, nil)

	message := MockMessage()

	responses := make(chan Response)
	go HideFavouritesCmdHandler(nil, &bot, &message, responses)

	select {
	case req := <-reqChan:
		actual := req
		expected := Request{Path: "/bot/sendMessage", Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&reply_markup=%7B%22remove_keyboard%22%3Atrue%2C%22selective%22%3Afalse%7D&text=Favourites+keyboard+hidden."}

		if actual != expected {
			fmt.Printf("Expected:\n%#v\nActual:\n%#v\n", expected, actual)
			t.Fail()
		}
	case err := <-errChan:
		t.Fatalf("%v", err)
	case r := <-responses:
		if err := r.Error; err != nil {
			t.Fatalf("%+v", err)
		}
	}
}
