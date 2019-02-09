package telegram

import (
	"testing"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/kr/pretty"
	"github.com/stretchr/testify/assert"
)

func TestInputTextMessageContent_content(t *testing.T) {
	actual := InputTextMessageContent{
		MessageText: "Hello",
		ParseMode:   "Markdown",
	}.content()
	expected := tgbotapi.InputTextMessageContent{
		Text:      "Hello",
		ParseMode: "Markdown",
	}
	assert.Equal(t, expected, actual)
}

func TestInlineQueryResultArticle_result(t *testing.T) {
	type fields struct {
		ID                  string
		Title               string
		Description         string
		ThumbURL            string
		InputMessageContent InputMessageContent
		ReplyMarkup         InlineKeyboardMarkup
	}
	tests := []struct {
		name   string
		fields fields
		want   interface{}
	}{
		{
			name: "without inline keyboard",
			fields: fields{
				ID:          "ID",
				Title:       "Title",
				Description: "Description",
				ThumbURL:    "ThumbURL",
				InputMessageContent: InputTextMessageContent{
					MessageText: "Hello",
					ParseMode:   "Markdown",
				},
			},
			want: tgbotapi.InlineQueryResultArticle{
				Type:  "article",
				ID:    "ID",
				Title: "Title",
				InputMessageContent: tgbotapi.InputTextMessageContent{
					Text:      "Hello",
					ParseMode: "Markdown",
				},
				Description: "Description",
				ThumbURL:    "ThumbURL",
			},
		},
		{
			name: "with inline keyboard",
			fields: fields{
				ID:          "ID",
				Title:       "Title",
				Description: "Description",
				ThumbURL:    "ThumbURL",
				InputMessageContent: InputTextMessageContent{
					MessageText: "Hello",
					ParseMode:   "Markdown",
				},
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
			want: tgbotapi.InlineQueryResultArticle{
				Type:  "article",
				ID:    "ID",
				Title: "Title",
				InputMessageContent: tgbotapi.InputTextMessageContent{
					Text:      "Hello",
					ParseMode: "Markdown",
				},
				ReplyMarkup: func() *tgbotapi.InlineKeyboardMarkup {
					button := tgbotapi.NewInlineKeyboardButtonData("Button", "data")
					markup := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(button))
					return &markup
				}(),
				Description: "Description",
				ThumbURL:    "ThumbURL",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := InlineQueryResultArticle{
				ID:                  tt.fields.ID,
				Title:               tt.fields.Title,
				Description:         tt.fields.Description,
				ThumbURL:            tt.fields.ThumbURL,
				InputMessageContent: tt.fields.InputMessageContent,
				ReplyMarkup:         tt.fields.ReplyMarkup,
			}
			if got := a.result(); !assert.Equal(t, tt.want, got) {
				pretty.Println(got)
			}
		})
	}
}

func TestAnswerInlineQueryRequest_config(t *testing.T) {
	type fields struct {
		InlineQueryID string
		Results       []InlineQueryResult
		CacheTime     int
		IsPersonal    bool
	}
	tests := []struct {
		name   string
		fields fields
		want   tgbotapi.InlineConfig
	}{
		{
			name: "no results",
			fields: fields{
				InlineQueryID: "ID",
				Results:       nil,
				CacheTime:     100,
				IsPersonal:    true,
			},
			want: tgbotapi.InlineConfig{
				InlineQueryID: "ID",
				Results:       make([]interface{}, 0),
				CacheTime:     100,
				IsPersonal:    true,
			},
		},
		{
			name: "with results",
			fields: fields{
				InlineQueryID: "InlineQueryID",
				Results: []InlineQueryResult{
					InlineQueryResultArticle{
						InputMessageContent: InputTextMessageContent{},
					},
					InlineQueryResultArticle{
						InputMessageContent: InputTextMessageContent{},
					},
				},
				CacheTime:  100,
				IsPersonal: true,
			},
			want: tgbotapi.InlineConfig{
				InlineQueryID: "InlineQueryID",
				Results: []interface{}{
					tgbotapi.InlineQueryResultArticle{
						Type:                "article",
						InputMessageContent: tgbotapi.InputTextMessageContent{},
					},
					tgbotapi.InlineQueryResultArticle{
						Type:                "article",
						InputMessageContent: tgbotapi.InputTextMessageContent{},
					},
				},
				CacheTime:  100,
				IsPersonal: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := AnswerInlineQueryRequest{
				InlineQueryID: tt.fields.InlineQueryID,
				Results:       tt.fields.Results,
				CacheTime:     tt.fields.CacheTime,
				IsPersonal:    tt.fields.IsPersonal,
			}
			if got := r.config(); !assert.Equal(t, tt.want, got) {
				pretty.Println(got)
			}
		})
	}
}
