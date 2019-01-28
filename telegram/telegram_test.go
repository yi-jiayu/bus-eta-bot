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

func TestSendMessageRequest_config(t *testing.T) {
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

func TestEditMessageTextRequest_config(t *testing.T) {
	type fields struct {
		ChatID          int64
		MessageID       int
		InlineMessageID string
		Text            string
		ParseMode       string
		ReplyMarkup     InlineKeyboardMarkup
	}
	tests := []struct {
		name       string
		fields     fields
		wantConfig tgbotapi.EditMessageTextConfig
	}{
		{
			name: "with ChatID and MessageID",
			fields: fields{
				ChatID:    1,
				MessageID: 1,
				Text:      "Edited message",
			},
			wantConfig: tgbotapi.NewEditMessageText(1, 1, "Edited message"),
		},
		{
			name: "with ChatID, MessageID and ParseMode",
			fields: fields{
				ChatID:    1,
				MessageID: 1,
				Text:      "Edited message",
				ParseMode: "markdown",
			},
			wantConfig: func() tgbotapi.EditMessageTextConfig {
				config := tgbotapi.NewEditMessageText(1, 1, "Edited message")
				config.ParseMode = "markdown"
				return config
			}(),
		},
		{
			name: "with ChatID, MessageID and InlineKeyboardMarkup",
			fields: fields{
				ChatID:    1,
				MessageID: 1,
				Text:      "Edited message",
				ReplyMarkup: InlineKeyboardMarkup{
					InlineKeyboard: [][]InlineKeyboardButton{
						{
							{
								Text:         "Button",
								CallbackData: "data",
							},
						},
					},
				},
			},
			wantConfig: func() tgbotapi.EditMessageTextConfig {
				config := tgbotapi.NewEditMessageText(1, 1, "Edited message")
				button := tgbotapi.NewInlineKeyboardButtonData("Button", "data")
				markup := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(button))
				config.ReplyMarkup = &markup
				return config
			}(),
		},
		{
			name: "with InlineMessageID",
			fields: fields{
				InlineMessageID: "1",
				Text:            "Edited message",
			},
			wantConfig: tgbotapi.EditMessageTextConfig{
				BaseEdit: tgbotapi.BaseEdit{
					InlineMessageID: "1",
				},
				Text: "Edited message",
			},
		},
		{
			name: "with InlineMessageID and ParseMode",
			fields: fields{
				InlineMessageID: "1",
				Text:            "Edited message",
				ParseMode:       "markdown",
			},
			wantConfig: tgbotapi.EditMessageTextConfig{
				BaseEdit: tgbotapi.BaseEdit{
					InlineMessageID: "1",
				},
				Text:      "Edited message",
				ParseMode: "markdown",
			},
		},
		{
			name: "with InlineMessageID and InlineKeyboardMarkup",
			fields: fields{
				InlineMessageID: "1",
				Text:            "Edited message",
				ReplyMarkup: InlineKeyboardMarkup{
					InlineKeyboard: [][]InlineKeyboardButton{
						{
							{
								Text:         "Button",
								CallbackData: "data",
							},
						},
					},
				},
			},
			wantConfig: func() tgbotapi.EditMessageTextConfig {
				config := tgbotapi.EditMessageTextConfig{
					BaseEdit: tgbotapi.BaseEdit{
						InlineMessageID: "1",
					},
					Text: "Edited message",
				}
				button := tgbotapi.NewInlineKeyboardButtonData("Button", "data")
				markup := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(button))
				config.ReplyMarkup = &markup
				return config
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := EditMessageTextRequest{
				ChatID:          tt.fields.ChatID,
				MessageID:       tt.fields.MessageID,
				InlineMessageID: tt.fields.InlineMessageID,
				Text:            tt.fields.Text,
				ParseMode:       tt.fields.ParseMode,
				ReplyMarkup:     tt.fields.ReplyMarkup,
			}
			gotConfig := r.config()
			assert.Equal(t, tt.wantConfig, gotConfig)
		})
	}
}
