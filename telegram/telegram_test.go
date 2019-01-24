package telegram

import (
	"testing"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/stretchr/testify/assert"
)

type mockReplyMarkup int

func (m mockReplyMarkup) markup() interface{} {
	return m
}

func Test_sendMessage(t *testing.T) {
	type testCase struct {
		Name     string
		Request  SendMessageRequest
		Expected tgbotapi.Chattable
	}
	testCases := []testCase{
		{
			Name: "with text",
			Request: SendMessageRequest{
				ChatID: 1,
				Text:   "Hello, World",
			},
			Expected: tgbotapi.NewMessage(1, "Hello, World"),
		},
		{
			Name: "with parse mode",
			Request: SendMessageRequest{
				ChatID:    1,
				Text:      "**Hello, World**",
				ParseMode: "markdown",
			},
			Expected: func() tgbotapi.MessageConfig {
				m := tgbotapi.NewMessage(1, "**Hello, World**")
				m.ParseMode = "markdown"
				return m
			}(),
		},
		{
			Name: "with reply to message ID",
			Request: SendMessageRequest{
				ChatID:           1,
				Text:             "Hello, World",
				ReplyToMessageID: 1,
			},
			Expected: func() tgbotapi.MessageConfig {
				m := tgbotapi.NewMessage(1, "Hello, World")
				m.ReplyToMessageID = 1
				return m
			}(),
		},
		{
			Name: "with reply markup",
			Request: SendMessageRequest{
				ChatID:      1,
				Text:        "Hello, World",
				ReplyMarkup: mockReplyMarkup(1),
			},
			Expected: func() tgbotapi.MessageConfig {
				m := tgbotapi.NewMessage(1, "Hello, World")
				m.ReplyMarkup = mockReplyMarkup(1)
				return m
			}(),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			actual := tc.Request.config()
			assert.Equal(t, tc.Expected, actual)
		})
	}
}
