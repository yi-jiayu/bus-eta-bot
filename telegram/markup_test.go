package telegram

import (
	"testing"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/stretchr/testify/assert"
)

func TestInlineKeyboardMarkup_markup(t *testing.T) {
	markup := InlineKeyboardMarkup{
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
	}
	actual := markup.markup()
	expected := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.NewInlineKeyboardButtonData("Refresh", `{"t":"refresh","b":"96049"}`),
				tgbotapi.InlineKeyboardButton{
					Text:                         "Resend",
					SwitchInlineQueryCurrentChat: NewSwitchInlineQueryCurrentChat("SUTD"),
				},
			},
		},
	}
	assert.Equal(t, expected, actual)
}

func TestForceReply_markup(t *testing.T) {
	t.Run("not selective", func(t *testing.T) {
		forceReply := ForceReply{
			Selective: false,
		}
		actual := forceReply.markup()
		expected := tgbotapi.ForceReply{
			ForceReply: true,
			Selective:  false,
		}
		assert.Equal(t, expected, actual)
	})
	t.Run("selective", func(t *testing.T) {
		forceReply := ForceReply{
			Selective: true,
		}
		actual := forceReply.markup()
		expected := tgbotapi.ForceReply{
			ForceReply: true,
			Selective:  true,
		}
		assert.Equal(t, expected, actual)
	})
}

func TestNewSwitchInlineQueryCurrentChat(t *testing.T) {
	actual := NewSwitchInlineQueryCurrentChat("SUTD")
	q := "SUTD"
	expected := &q
	assert.Equal(t, expected, actual)
}
