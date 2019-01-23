package telegram

import (
	"net/http"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type Request interface {
	doWith(c *Client) (result interface{}, err error)
}

type Error struct {
	Description string
}

func (err Error) Error() string {
	return err.Description
}

type InlineKeyboardButton struct {
	Text                         string
	CallbackData                 string
	SwitchInlineQueryCurrentChat *string
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton
}

type SendMessageRequest struct {
	ChatID      int64
	Text        string
	ParseMode   string
	ReplyMarkup *InlineKeyboardMarkup
}

func (r SendMessageRequest) config() tgbotapi.MessageConfig {
	message := tgbotapi.NewMessage(r.ChatID, r.Text)
	if r.ParseMode != "" {
		message.ParseMode = r.ParseMode
	}
	if r.ReplyMarkup != nil {
		var buttons [][]tgbotapi.InlineKeyboardButton
		for _, row := range r.ReplyMarkup.InlineKeyboard {
			var r []tgbotapi.InlineKeyboardButton
			for _, button := range row {
				b := tgbotapi.InlineKeyboardButton{
					Text: button.Text,
				}
				if d := button.CallbackData; d != "" {
					b.CallbackData = &d
				}
				if p := button.SwitchInlineQueryCurrentChat; p != nil {
					b.SwitchInlineQueryCurrentChat = p
				}
				r = append(r, b)
			}
			buttons = append(buttons, r)
		}
		message.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: buttons,
		}
	}
	return message
}

func (r SendMessageRequest) doWith(c *Client) (result interface{}, err error) {
	m, err := c.Client.Send(r.config())
	if err != nil {
		return nil, Error{Description: err.(tgbotapi.Error).Message}
	}
	return m, nil
}

type EditMessageRequest struct {
}

type AnswerCallbackQueryRequest struct {
}

type AnswerInlineQueryRequest struct {
}

type Client struct {
	Client *tgbotapi.BotAPI
}

func (c *Client) Do(request Request) error {
	_, err := request.doWith(c)
	return err
}

func NewClient(token string, httpClient *http.Client) (*Client, error) {
	client, err := tgbotapi.NewBotAPIWithClient(token, httpClient)
	if err != nil {
		return nil, err
	}
	return &Client{
		Client: client,
	}, nil
}
