package telegram

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type ReplyMarkup interface {
	markup() interface{}
}

type InlineKeyboardButton struct {
	Text                         string
	CallbackData                 string
	SwitchInlineQueryCurrentChat *string
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton
}

func (m InlineKeyboardMarkup) markup() interface{} {
	return m.inlineKeyboardMarkup()
}

func (m InlineKeyboardMarkup) inlineKeyboardMarkup() tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, r := range m.InlineKeyboard {
		var row []tgbotapi.InlineKeyboardButton
		for _, b := range r {
			button := tgbotapi.InlineKeyboardButton{
				Text: b.Text,
			}
			if d := b.CallbackData; d != "" {
				button.CallbackData = &d
			}
			if p := b.SwitchInlineQueryCurrentChat; p != nil {
				button.SwitchInlineQueryCurrentChat = p
			}
			row = append(row, button)
		}
		rows = append(rows, row)
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

type ForceReply struct {
	Selective bool
}

func (r ForceReply) markup() interface{} {
	return tgbotapi.ForceReply{
		ForceReply: true,
		Selective:  r.Selective,
	}
}

// NewSwitchInlineQueryCurrentChat returns a pointer to a string which can be passed to InlineKeyboardButton.
func NewSwitchInlineQueryCurrentChat(q string) *string {
	return &q
}

// NewForceReply is a convenience function for creating a ForceReply.
func NewForceReply(selective bool) ForceReply {
	return ForceReply{
		Selective: selective,
	}
}
