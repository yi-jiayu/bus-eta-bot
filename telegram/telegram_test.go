package telegram

import (
	"testing"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/stretchr/testify/assert"
)

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
			Expected: tgbotapi.MessageConfig{
				BaseChat: tgbotapi.BaseChat{
					ChatID: 1,
				},
				Text: "Hello, World",
			},
		},
		{
			Name: "with parse mode",
			Request: SendMessageRequest{
				ChatID:    1,
				Text:      "**Hello, World**",
				ParseMode: "markdown",
			},
			Expected: tgbotapi.MessageConfig{
				BaseChat: tgbotapi.BaseChat{
					ChatID: 1,
				},
				Text:      "**Hello, World**",
				ParseMode: "markdown",
			},
		},
		{
			Name: "with reply markup",
			Request: SendMessageRequest{
				ChatID: 1,
				Text:   "Hello, World",
				ReplyMarkup: &InlineKeyboardMarkup{
					InlineKeyboard: [][]InlineKeyboardButton{
						{
							{
								Text:         "Refresh",
								CallbackData: `{"t":"refresh","b":"96049"}`,
							},
							{
								Text:                         "Resend",
								SwitchInlineQueryCurrentChat: NewSwitchInlineQueryCurrentChat("SUTD"),
							},
						},
					},
				},
			},
			Expected: tgbotapi.MessageConfig{
				BaseChat: tgbotapi.BaseChat{
					ChatID: 1,
					ReplyMarkup: tgbotapi.InlineKeyboardMarkup{
						InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
							{
								tgbotapi.NewInlineKeyboardButtonData("Refresh", `{"t":"refresh","b":"96049"}`),
								tgbotapi.InlineKeyboardButton{
									Text:                         "Resend",
									SwitchInlineQueryCurrentChat: NewSwitchInlineQueryCurrentChat("SUTD"),
								},
							},
						},
					},
				},
				Text: "Hello, World",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			actual := tc.Request.config()
			assert.Equal(t, tc.Expected, actual)
		})
	}
}
