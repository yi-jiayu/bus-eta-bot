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

type SendMessageRequest struct {
	ChatID           int64
	Text             string
	ParseMode        string
	ReplyToMessageID int
	ReplyMarkup      ReplyMarkup
}

func (r SendMessageRequest) config() tgbotapi.MessageConfig {
	message := tgbotapi.NewMessage(r.ChatID, r.Text)
	if r.ParseMode != "" {
		message.ParseMode = r.ParseMode
	}
	if r.ReplyToMessageID != 0 {
		message.ReplyToMessageID = r.ReplyToMessageID
	}
	if r.ReplyMarkup != nil {
		message.ReplyMarkup = r.ReplyMarkup.markup()
	}
	return message
}

func (r SendMessageRequest) doWith(c *Client) (result interface{}, err error) {
	m, err := c.client.Send(r.config())
	if err != nil {
		return nil, Error{Description: err.(tgbotapi.Error).Message}
	}
	return m, nil
}

type EditMessageTextRequest struct {
	ChatID          int64
	MessageID       int
	InlineMessageID string
	Text            string
	ParseMode       string
	ReplyMarkup     InlineKeyboardMarkup
}

func (r EditMessageTextRequest) config() (config tgbotapi.EditMessageTextConfig) {
	if r.InlineMessageID != "" {
		config = tgbotapi.EditMessageTextConfig{
			BaseEdit: tgbotapi.BaseEdit{
				InlineMessageID: r.InlineMessageID,
			},
			Text: r.Text,
		}
	} else {
		config = tgbotapi.NewEditMessageText(r.ChatID, r.MessageID, r.Text)
	}
	if r.ParseMode != "" {
		config.ParseMode = r.ParseMode
	}
	if len(r.ReplyMarkup.InlineKeyboard) > 0 {
		markup := r.ReplyMarkup.inlineKeyboardMarkup()
		config.ReplyMarkup = &markup
	}
	return
}

func (r EditMessageTextRequest) doWith(c *Client) (result interface{}, err error) {
	m, err := c.client.Send(r.config())
	if err != nil {
		return nil, Error{Description: err.(tgbotapi.Error).Message}
	}
	return m, nil
}

type AnswerCallbackQueryRequest struct {
	CallbackQueryID string
	Text            string
}

func (r AnswerCallbackQueryRequest) doWith(c *Client) (result interface{}, err error) {
	config := tgbotapi.NewCallback(r.CallbackQueryID, r.Text)
	res, err := c.client.AnswerCallbackQuery(config)
	if err != nil {
		return nil, Error{Description: res.Description}
	}
	return
}

type AnswerInlineQueryRequest struct {
}

type Client struct {
	client *tgbotapi.BotAPI
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
		client: client,
	}, nil
}
