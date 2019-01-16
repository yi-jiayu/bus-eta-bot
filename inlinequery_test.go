package busetabot

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yi-jiayu/telegram-bot-api"
)

type MockStreetView struct{}

func (s *MockStreetView) GetPhotoURLByLocation(lat, lon float64, width, height int) (string, error) {
	return "", nil
}

func TestInlineQueryHandler(t *testing.T) {
	busStops := []BusStop{
		{
			BusStopCode: "96041",
			RoadName:    "Upp Changi Rd East",
			Description: "Bef Tropicana Condo",
			Latitude:    1.34041450268626,
			Longitude:   103.96127892061004,
		},
		{
			BusStopCode: "96049",
			RoadName:    "Upp Changi Rd East",
			Description: "Opp Tropicana Condo",
			Latitude:    1.33995375346513,
			Longitude:   103.96079768187379,
		},
	}

	tgAPI, reqChan, errChan := NewMockTelegramAPIWithPath()
	defer tgAPI.Close()

	tg := &tgbotapi.BotAPI{
		APIEndpoint: tgAPI.URL + "/bot%s/%s",
		Client:      http.DefaultClient,
	}

	sv := NewStreetViewAPI("API_KEY")

	bot := NewBusEtaBot(handlers, tg, nil, &sv, nil)
	bot.BusStops = NewInMemoryBusStopRepository(busStops, nil)

	testCases := []struct {
		Name     string
		Query    string
		Location *tgbotapi.Location
		Expected Request
	}{
		{
			Name:  "Empty query",
			Query: "",
			Expected: Request{
				Path: "/bot/answerInlineQuery",
				Body: "cache_time=0&inline_query_id=1&is_personal=false&next_offset=&results=%5B%7B%22type%22%3A%22article%22%2C%22id%22%3A%2296041%22%2C%22title%22%3A%22Bef+Tropicana+Condo+%2896041%29%22%2C%22input_message_content%22%3A%7B%22message_text%22%3A%22%2ABef+Tropicana+Condo+%2896041%29%2A%5CnUpp+Changi+Rd+East%5Cn%60Fetching+etas...%60%22%2C%22parse_mode%22%3A%22markdown%22%2C%22disable_web_page_preview%22%3Afalse%7D%2C%22reply_markup%22%3A%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A%5C%2296041%5C%22%7D%22%7D%5D%5D%7D%2C%22url%22%3A%22%22%2C%22hide_url%22%3Afalse%2C%22description%22%3A%22Upp+Changi+Rd+East%22%2C%22thumb_url%22%3A%22https%3A%2F%2Fmaps.googleapis.com%2Fmaps%2Fapi%2Fstreetview%3Fkey%3DAPI_KEY%5Cu0026location%3D1.340415%252C103.961279%5Cu0026size%3D100x100%22%2C%22thumb_width%22%3A0%2C%22thumb_height%22%3A0%7D%2C%7B%22type%22%3A%22article%22%2C%22id%22%3A%2296049%22%2C%22title%22%3A%22Opp+Tropicana+Condo+%2896049%29%22%2C%22input_message_content%22%3A%7B%22message_text%22%3A%22%2AOpp+Tropicana+Condo+%2896049%29%2A%5CnUpp+Changi+Rd+East%5Cn%60Fetching+etas...%60%22%2C%22parse_mode%22%3A%22markdown%22%2C%22disable_web_page_preview%22%3Afalse%7D%2C%22reply_markup%22%3A%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A%5C%2296049%5C%22%7D%22%7D%5D%5D%7D%2C%22url%22%3A%22%22%2C%22hide_url%22%3Afalse%2C%22description%22%3A%22Upp+Changi+Rd+East%22%2C%22thumb_url%22%3A%22https%3A%2F%2Fmaps.googleapis.com%2Fmaps%2Fapi%2Fstreetview%3Fkey%3DAPI_KEY%5Cu0026location%3D1.339954%252C103.960798%5Cu0026size%3D100x100%22%2C%22thumb_width%22%3A0%2C%22thumb_height%22%3A0%7D%5D&switch_pm_parameter=&switch_pm_text=",
			},
		},
		{
			Name:  "Empty query with location",
			Query: "",
			Location: &tgbotapi.Location{
				Latitude:  1.340,
				Longitude: 103.961,
			},
			Expected: Request{
				Path: "/bot/answerInlineQuery",
				Body: "cache_time=0&inline_query_id=1&is_personal=false&next_offset=&results=%5B%7B%22type%22%3A%22article%22%2C%22id%22%3A%2296049+geo%22%2C%22title%22%3A%22Opp+Tropicana+Condo+%2896049%29%22%2C%22input_message_content%22%3A%7B%22message_text%22%3A%22%2AOpp+Tropicana+Condo+%2896049%29%2A%5CnUpp+Changi+Rd+East%5Cn%60Fetching+etas...%60%22%2C%22parse_mode%22%3A%22markdown%22%2C%22disable_web_page_preview%22%3Afalse%7D%2C%22reply_markup%22%3A%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A%5C%2296049%5C%22%7D%22%7D%5D%5D%7D%2C%22url%22%3A%22%22%2C%22hide_url%22%3Afalse%2C%22description%22%3A%2223+m+away%22%2C%22thumb_url%22%3A%22https%3A%2F%2Fmaps.googleapis.com%2Fmaps%2Fapi%2Fstreetview%3Fkey%3DAPI_KEY%5Cu0026location%3D1.339954%252C103.960798%5Cu0026size%3D100x100%22%2C%22thumb_width%22%3A0%2C%22thumb_height%22%3A0%7D%2C%7B%22type%22%3A%22article%22%2C%22id%22%3A%2296041+geo%22%2C%22title%22%3A%22Bef+Tropicana+Condo+%2896041%29%22%2C%22input_message_content%22%3A%7B%22message_text%22%3A%22%2ABef+Tropicana+Condo+%2896041%29%2A%5CnUpp+Changi+Rd+East%5Cn%60Fetching+etas...%60%22%2C%22parse_mode%22%3A%22markdown%22%2C%22disable_web_page_preview%22%3Afalse%7D%2C%22reply_markup%22%3A%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A%5C%2296041%5C%22%7D%22%7D%5D%5D%7D%2C%22url%22%3A%22%22%2C%22hide_url%22%3Afalse%2C%22description%22%3A%2255+m+away%22%2C%22thumb_url%22%3A%22https%3A%2F%2Fmaps.googleapis.com%2Fmaps%2Fapi%2Fstreetview%3Fkey%3DAPI_KEY%5Cu0026location%3D1.340415%252C103.961279%5Cu0026size%3D100x100%22%2C%22thumb_width%22%3A0%2C%22thumb_height%22%3A0%7D%5D&switch_pm_parameter=&switch_pm_text=",
			},
		},
		{
			Name:  "Matching query",
			Query: "tropicana",
			Expected: Request{
				Path: "/bot/answerInlineQuery",
				Body: "cache_time=86400&inline_query_id=1&is_personal=false&next_offset=&results=%5B%7B%22type%22%3A%22article%22%2C%22id%22%3A%2296041%22%2C%22title%22%3A%22Bef+Tropicana+Condo+%2896041%29%22%2C%22input_message_content%22%3A%7B%22message_text%22%3A%22%2ABef+Tropicana+Condo+%2896041%29%2A%5CnUpp+Changi+Rd+East%5Cn%60Fetching+etas...%60%22%2C%22parse_mode%22%3A%22markdown%22%2C%22disable_web_page_preview%22%3Afalse%7D%2C%22reply_markup%22%3A%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A%5C%2296041%5C%22%7D%22%7D%5D%5D%7D%2C%22url%22%3A%22%22%2C%22hide_url%22%3Afalse%2C%22description%22%3A%22Upp+Changi+Rd+East%22%2C%22thumb_url%22%3A%22https%3A%2F%2Fmaps.googleapis.com%2Fmaps%2Fapi%2Fstreetview%3Fkey%3DAPI_KEY%5Cu0026location%3D1.340415%252C103.961279%5Cu0026size%3D100x100%22%2C%22thumb_width%22%3A0%2C%22thumb_height%22%3A0%7D%2C%7B%22type%22%3A%22article%22%2C%22id%22%3A%2296049%22%2C%22title%22%3A%22Opp+Tropicana+Condo+%2896049%29%22%2C%22input_message_content%22%3A%7B%22message_text%22%3A%22%2AOpp+Tropicana+Condo+%2896049%29%2A%5CnUpp+Changi+Rd+East%5Cn%60Fetching+etas...%60%22%2C%22parse_mode%22%3A%22markdown%22%2C%22disable_web_page_preview%22%3Afalse%7D%2C%22reply_markup%22%3A%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A%5C%2296049%5C%22%7D%22%7D%5D%5D%7D%2C%22url%22%3A%22%22%2C%22hide_url%22%3Afalse%2C%22description%22%3A%22Upp+Changi+Rd+East%22%2C%22thumb_url%22%3A%22https%3A%2F%2Fmaps.googleapis.com%2Fmaps%2Fapi%2Fstreetview%3Fkey%3DAPI_KEY%5Cu0026location%3D1.339954%252C103.960798%5Cu0026size%3D100x100%22%2C%22thumb_width%22%3A0%2C%22thumb_height%22%3A0%7D%5D&switch_pm_parameter=&switch_pm_text=",
			},
		},
		{
			Name:  "Non-matching query",
			Query: "anaciport",
			Expected: Request{
				Path: "/bot/answerInlineQuery",
				Body: "cache_time=86400&inline_query_id=1&is_personal=false&next_offset=&results=%5B%5D&switch_pm_parameter=&switch_pm_text=",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ilq := MockInlineQuery()
			ilq.Query = tc.Query
			ilq.Location = tc.Location

			err := InlineQueryHandler(context.Background(), &bot, &ilq)
			if err != nil {
				t.Fatal(err)
			}

			select {
			case req := <-reqChan:
				actual := req
				expected := tc.Expected
				assert.Equal(t, expected, actual)
			case err := <-errChan:
				t.Fatalf("%v", err)
			}
		})
	}
}

func TestChosenInlineResultHandler(t *testing.T) {
	t.Parallel()

	tgAPI, reqChan, errChan := NewMockTelegramAPIWithPath()
	defer tgAPI.Close()

	tg := &tgbotapi.BotAPI{
		APIEndpoint: tgAPI.URL + "/bot%s/%s",
		Client:      http.DefaultClient,
	}

	datamall := MockDatamall{
		BusArrival: newArrival(time.Time{}),
		Error:      nil,
	}

	busStops := MockBusStops{
		BusStop: &BusStop{
			BusStopCode: "96049",
			RoadName:    "Upp Changi Rd East",
			Description: "Opp Tropicana Condo",
		},
	}

	bot := BusEtaBot{
		Telegram: tg,
		Datamall: datamall,
		BusStops: busStops,
		NowFunc: func() time.Time {
			return time.Time{}
		},
	}

	testCases := []struct {
		Name     string
		ResultID string
		Expected Request
	}{
		{
			Name:     "Normal chosen inline result",
			ResultID: "96049",
			Expected: Request{
				Path: "/bot/editMessageText",
				Body: "chat_id=0&disable_web_page_preview=false&message_id=0&parse_mode=markdown&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A%5C%2296049%5C%22%7D%22%7D%5D%5D%7D&text=%2AOpp+Tropicana+Condo+%2896049%29%2A%0AUpp+Changi+Rd+East%0A%60%60%60%0A%7C+Svc+%7C+Next+%7C++2nd+%7C++3rd+%7C%0A%7C-----%7C------%7C------%7C------%7C%0A%7C+2+++%7C+++-1+%7C+++10+%7C+++36+%7C%0A%7C+24++%7C++++1+%7C++++3+%7C++++6+%7C%60%60%60%0AShowing+2+out+of+2+services+for+this+bus+stop.%0A%0A_Last+updated+at+01+Jan+01+08%3A00+SGT_",
			},
		},
		{
			Name:     "Nearby chosen inline result",
			ResultID: "96049 geo",
			Expected: Request{
				Path: "/bot/editMessageText",
				Body: "chat_id=0&disable_web_page_preview=false&message_id=0&parse_mode=markdown&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Refresh%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22refresh%5C%22%2C%5C%22b%5C%22%3A%5C%2296049%5C%22%7D%22%7D%5D%5D%7D&text=%2AOpp+Tropicana+Condo+%2896049%29%2A%0AUpp+Changi+Rd+East%0A%60%60%60%0A%7C+Svc+%7C+Next+%7C++2nd+%7C++3rd+%7C%0A%7C-----%7C------%7C------%7C------%7C%0A%7C+2+++%7C+++-1+%7C+++10+%7C+++36+%7C%0A%7C+24++%7C++++1+%7C++++3+%7C++++6+%7C%60%60%60%0AShowing+2+out+of+2+services+for+this+bus+stop.%0A%0A_Last+updated+at+01+Jan+01+08%3A00+SGT_",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			cir := tgbotapi.ChosenInlineResult{
				ResultID: tc.ResultID,
				From: &tgbotapi.User{
					ID:        1,
					FirstName: "Jiayu",
				},
			}

			err := ChosenInlineResultHandler(context.Background(), &bot, &cir)
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
				t.Fatalf("%v", err)
			}
		})
	}
}

func TestGetNearbyInlineQueryResults(t *testing.T) {
	streetView := &MockStreetView{}
	busStops := NewInMemoryBusStopRepository([]BusStop{
		{
			BusStopCode: "96041",
			RoadName:    "Upp Changi Rd East",
			Description: "Bef Tropicana Condo",
			Latitude:    1.34041450268626,
			Longitude:   103.96127892061004,
		},
		{
			BusStopCode: "96049",
			RoadName:    "Upp Changi Rd East",
			Description: "Opp Tropicana Condo",
			Latitude:    1.33995375346513,
			Longitude:   103.96079768187379,
		},
	}, nil)
	expected := []interface{}{
		tgbotapi.InlineQueryResultArticle{
			Type:  "article",
			ID:    "96049 geo",
			Title: "Opp Tropicana Condo (96049)",
			InputMessageContent: tgbotapi.InputTextMessageContent{
				Text:                  "*Opp Tropicana Condo (96049)*\nUpp Changi Rd East\n`Fetching etas...`",
				ParseMode:             "markdown",
				DisableWebPagePreview: false,
			},
			ReplyMarkup: newEtaMessageReplyMarkupInline("96049"),
			URL:         "",
			HideURL:     false,
			Description: "23 m away",
			ThumbURL:    "",
			ThumbWidth:  0,
			ThumbHeight: 0,
		},
		tgbotapi.InlineQueryResultArticle{
			Type:  "article",
			ID:    "96041 geo",
			Title: "Bef Tropicana Condo (96041)",
			InputMessageContent: tgbotapi.InputTextMessageContent{
				Text:                  "*Bef Tropicana Condo (96041)*\nUpp Changi Rd East\n`Fetching etas...`",
				ParseMode:             "markdown",
				DisableWebPagePreview: false,
			},
			ReplyMarkup: newEtaMessageReplyMarkupInline("96041"),
			URL:         "",
			HideURL:     false,
			Description: "55 m away",
			ThumbURL:    "",
			ThumbWidth:  0,
			ThumbHeight: 0,
		},
	}
	actual, err := GetNearbyInlineQueryResults(context.Background(), streetView, busStops, 1.340, 103.961)
	if err != nil {
		t.Fatal(err)
	}
	for _, result := range actual {
		article := result.(tgbotapi.InlineQueryResultArticle)
		t.Logf("%#v", *article.ReplyMarkup)
	}
	assert.Equal(t, expected, actual)
}
