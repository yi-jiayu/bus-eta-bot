package telegram

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type InputMessageContent interface {
	content() interface{}
}

type InputTextMessageContent struct {
	MessageText string
	ParseMode   string
}

func (c InputTextMessageContent) content() interface{} {
	return tgbotapi.InputTextMessageContent{
		Text:      c.MessageText,
		ParseMode: c.ParseMode,
	}
}

type InlineQueryResult interface {
	result() interface{}
}

type InlineQueryResultArticle struct {
	ID                  string
	Title               string
	Description         string
	ThumbURL            string
	InputMessageContent InputMessageContent
	ReplyMarkup         InlineKeyboardMarkup
}

func (a InlineQueryResultArticle) result() interface{} {
	result := tgbotapi.InlineQueryResultArticle{
		Type:                "article",
		ID:                  a.ID,
		Title:               a.Title,
		InputMessageContent: a.InputMessageContent.content(),
		Description:         a.Description,
		ThumbURL:            a.ThumbURL,
	}
	if len(a.ReplyMarkup.InlineKeyboard) > 0 {
		markup := a.ReplyMarkup.inlineKeyboardMarkup()
		result.ReplyMarkup = &markup
	}
	return result
}

type AnswerInlineQueryRequest struct {
	InlineQueryID string
	Results       []InlineQueryResult
	CacheTime     int
	IsPersonal    bool
}

func (r AnswerInlineQueryRequest) config() tgbotapi.InlineConfig {
	results := make([]interface{}, len(r.Results))
	for i := range r.Results {
		results[i] = r.Results[i].result()
	}
	return tgbotapi.InlineConfig{
		InlineQueryID: r.InlineQueryID,
		Results:       results,
		CacheTime:     r.CacheTime,
		IsPersonal:    r.IsPersonal,
	}
}

func (r AnswerInlineQueryRequest) doWith(c *client) (result interface{}, err error) {
	_, err = c.botAPI.AnswerInlineQuery(r.config())
	if err != nil {
		return nil, newError(err)
	}
	return
}
