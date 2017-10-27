package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/yi-jiayu/telegram-bot-api"
)

func TestFallbackCommandHandler(t *testing.T) {
	t.Parallel()

	tgAPI, reqChan, errChan := NewMockTelegramAPIWithPath()
	defer tgAPI.Close()

	tg := &tgbotapi.BotAPI{
		APIEndpoint: tgAPI.URL + "/bot%s/%s",
		Client:      http.DefaultClient,
	}

	bot := NewBusEtaBot(handlers, tg, nil, nil, nil)

	testCases := []struct {
		Name     string
		ChatType string
		Text     string
		Expected Request
	}{
		{
			Name:     "Bus stop code with slash in front, private chat",
			ChatType: ChatTypePrivate,
			Text:     "/96049",
			Expected: Request{
				Path: "/bot/sendMessage",
				Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&text=Oops%2C+that+was+not+a+valid+command%21+If+you+wanted+to+get+etas+for+bus+stop+96049%2C+just+send+the+bus+stop+code+without+the+leading+slash.",
			},
		},
		{
			Name: "Bus stop code with slash in front, group chat",
			Text: "/96049",
			Expected: Request{
				Path: "/bot/sendMessage",
				Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&reply_to_message_id=1&text=Oops%2C+that+was+not+a+valid+command%21+If+you+wanted+to+get+etas+for+bus+stop+96049%2C+just+send+the+bus+stop+code+without+the+leading+slash.",
			},
		},
		{
			Name: "Invalid command, private chat",
			Text: "/invalid",
			Expected: Request{
				Path: "/bot/sendMessage",
				Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&reply_to_message_id=1&text=Oops%2C+that+was+not+a+valid+command%21",
			},
		},
		{
			Name: "Invalid command, group chat",
			Text: "/invalid",
			Expected: Request{
				Path: "/bot/sendMessage",
				Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&reply_to_message_id=1&text=Oops%2C+that+was+not+a+valid+command%21",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			message := MockMessageWithType(tc.ChatType)
			message.Text = tc.Text

			err := FallbackCommandHandler(nil, &bot, &message)
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
				t.Fatal(err)
			}
		})
	}
}

func TestAboutHandler(t *testing.T) {
	t.Parallel()

	tgAPI, reqChan, errChan := NewMockTelegramAPIWithPath()
	defer tgAPI.Close()

	tg := &tgbotapi.BotAPI{
		APIEndpoint: tgAPI.URL + "/bot%s/%s",
		Client:      http.DefaultClient,
	}

	bot := NewBusEtaBot(handlers, tg, nil, nil, nil)

	testCases := []struct {
		Name     string
		ChatType string
		Expected Request
	}{
		{
			Name:     "Private chat",
			ChatType: ChatTypePrivate,
			Expected: Request{
				Path: "/bot/sendMessage",
				Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&text=Bus+Eta+Bot+VERSION%0Ahttps%3A%2F%2Fgithub.com%2Fyi-jiayu%2Fbus-eta-bot",
			},
		},
		{
			Name: "Group, supergroup or channel",
			Expected: Request{
				Path: "/bot/sendMessage",
				Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&reply_to_message_id=1&text=Bus+Eta+Bot+VERSION%0Ahttps%3A%2F%2Fgithub.com%2Fyi-jiayu%2Fbus-eta-bot",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			message := MockMessageWithType(tc.ChatType)

			err := AboutHandler(nil, &bot, &message)
			if err != nil {
				t.Fatalf("%v", err)
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

func TestVersionHandler(t *testing.T) {
	t.Parallel()

	tgAPI, reqChan, errChan := NewMockTelegramAPIWithPath()
	defer tgAPI.Close()

	tg := &tgbotapi.BotAPI{
		APIEndpoint: tgAPI.URL + "/bot%s/%s",
		Client:      http.DefaultClient,
	}

	bot := NewBusEtaBot(handlers, tg, nil, nil, nil)

	testCases := []struct {
		Name     string
		ChatType string
		Expected Request
	}{
		{
			Name:     "Private chat",
			ChatType: ChatTypePrivate,
			Expected: Request{
				Path: "/bot/sendMessage",
				Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&text=Bus+Eta+Bot+VERSION%0Ahttps%3A%2F%2Fgithub.com%2Fyi-jiayu%2Fbus-eta-bot",
			},
		},
		{
			Name: "Group, supergroup or channel",
			Expected: Request{
				Path: "/bot/sendMessage",
				Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&reply_to_message_id=1&text=Bus+Eta+Bot+VERSION%0Ahttps%3A%2F%2Fgithub.com%2Fyi-jiayu%2Fbus-eta-bot",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			message := MockMessageWithType(tc.ChatType)

			err := VersionHandler(nil, &bot, &message)
			if err != nil {
				t.Fatalf("%v", err)
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

func TestStartHandler(t *testing.T) {
	t.Parallel()

	tgAPI, reqChan, errChan := NewMockTelegramAPIWithPath()
	defer tgAPI.Close()

	tg := &tgbotapi.BotAPI{
		APIEndpoint: tgAPI.URL + "/bot%s/%s",
		Client:      http.DefaultClient,
	}

	bot := NewBusEtaBot(handlers, tg, nil, nil, nil)
	message := MockMessage()

	err := StartHandler(nil, &bot, &message)
	if err != nil {
		t.Fatalf("%v", err)
	}

	select {
	case req := <-reqChan:
		actual := req
		expected := Request{
			Path: "/bot/sendMessage",
			Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&parse_mode=markdown&reply_markup=%7B%22inline_keyboard%22%3A%5B%5B%7B%22text%22%3A%22Get+etas+for+bus+stop+96049%22%2C%22callback_data%22%3A%22%7B%5C%22t%5C%22%3A%5C%22eta_demo%5C%22%7D%22%7D%5D%2C%5B%7B%22text%22%3A%22Try+an+inline+query%22%2C%22switch_inline_query_current_chat%22%3A%22Tropicana%22%7D%5D%5D%7D&text=Hello+Jiayu%2C%0A%0ABus+Eta+Bot+is+a+Telegram+bot+which+can+tell+you+how+long+you+have+to+wait+for+your+bus+to+arrive.%0A%0ATo+get+started%2C+try+sending+me+a+bus+stop+code+such+as+%6096049%60+to+get+etas+for.%0A%0AAlternatively%2C+you+can+also+search+for+bus+stops+by+sending+me+an+inline+query.+To+try+this+out%2C+type+%40BusEtaBot+followed+by+a+bus+stop+code%2C+description+or+road+name+in+any+chat.%0A%0AThanks+for+trying+out+Bus+Eta+Bot%21+If+you+find+Bus+Eta+Bot+useful%2C+do+help+to+spread+the+word+or+send+%2Ffeedback+to+leave+some+feedback+about+how+to+help+make+Bus+Eta+Bot+even+better%21%0A%0AIf+you%27re+stuck%2C+you+can+send+%2Fhelp+to+view+help.",
		}

		if actual != expected {
			fmt.Printf("Expected:\n%#v\nActual:\n%#v\n", expected, actual)
			t.Fail()
		}
	case err := <-errChan:
		t.Fatalf("%v", err)
	}
}

func TestHelpHandler(t *testing.T) {
	t.Parallel()

	tgAPI, reqChan, errChan := NewMockTelegramAPIWithPath()
	defer tgAPI.Close()

	tg := &tgbotapi.BotAPI{
		APIEndpoint: tgAPI.URL + "/bot%s/%s",
		Client:      http.DefaultClient,
	}

	bot := NewBusEtaBot(handlers, tg, nil, nil, nil)
	message := MockMessage()

	err := HelpHandler(nil, &bot, &message)
	if err != nil {
		t.Fatalf("%v", err)
	}

	select {
	case req := <-reqChan:
		actual := req
		expected := Request{
			Path: "/bot/sendMessage",
			Body: strings.Replace("chat_id=1&disable_notification=false&disable_web_page_preview=false&parse_mode=markdown&text=You+can+find+help+on+how+to+use+Bus+Eta+Bot+%5Bhere%5D%28$HELP_URL%29.", "$HELP_URL", url.QueryEscape(HelpURL), 1),
		}

		if actual != expected {
			fmt.Printf("Expected:\n%#v\nActual:\n%#v\n", expected, actual)
			t.Fail()
		}
	case err := <-errChan:
		t.Fatalf("%v", err)
	}
}

func TestPrivacyHandler(t *testing.T) {
	t.Parallel()

	tgAPI, reqChan, errChan := NewMockTelegramAPIWithPath()
	defer tgAPI.Close()

	tg := &tgbotapi.BotAPI{
		APIEndpoint: tgAPI.URL + "/bot%s/%s",
		Client:      http.DefaultClient,
	}

	sv := NewStreetViewAPI("API_KEY")

	bot := NewBusEtaBot(handlers, tg, nil, &sv, nil)
	message := MockMessage()

	err := PrivacyHandler(nil, &bot, &message)
	if err != nil {
		t.Fatalf("%v", err)
	}

	select {
	case req := <-reqChan:
		actual := req
		expected := Request{
			Path: "/bot/sendMessage",
			Body: strings.Replace("chat_id=1&disable_notification=false&disable_web_page_preview=false&parse_mode=markdown&text=You+can+find+Bus+Eta+Bot%27s+privacy+policy+%5Bhere%5D%28$PRIVACY_POLICY_URL%29.", "$PRIVACY_POLICY_URL", url.QueryEscape(PrivacyPolicyURL), 1),
		}

		if actual != expected {
			fmt.Printf("Expected:\n%#v\nActual:\n%#v\n", expected, actual)
			t.Fail()
		}
	case err := <-errChan:
		t.Fatalf("%v", err)
	}
}

func TestFeedbackCmdHandler(t *testing.T) {
	t.Parallel()

	tgAPI, reqChan, errChan := NewMockTelegramAPIWithPath()
	defer tgAPI.Close()

	tg := &tgbotapi.BotAPI{
		APIEndpoint: tgAPI.URL + "/bot%s/%s",
		Client:      http.DefaultClient,
	}

	bot := NewBusEtaBot(handlers, tg, nil, nil, nil)

	testCases := []struct {
		Name     string
		ChatType string
		Expected Request
	}{
		{
			Name:     "Private chat",
			ChatType: ChatTypePrivate,
			Expected: Request{
				Path: "/bot/sendMessage",
				Body: strings.Replace("chat_id=1&disable_notification=false&disable_web_page_preview=false&parse_mode=markdown&text=Oops%2C+the+feedback+command+has+not+been+implemented+yet.+In+the+meantime%2C+you+can+raise+issues+or+show+your+support+for+Bus+Eta+Bot+at+its+GitHub+repository+%5Bhere%5D%28$REPO_URL%29.", "$REPO_URL", url.QueryEscape(RepoURL), 1),
			},
		},
		{
			Name: "Group, supergroup or channel",
			Expected: Request{
				Path: "/bot/sendMessage",
				Body: strings.Replace("chat_id=1&disable_notification=false&disable_web_page_preview=false&parse_mode=markdown&reply_to_message_id=1&text=Oops%2C+the+feedback+command+has+not+been+implemented+yet.+In+the+meantime%2C+you+can+raise+issues+or+show+your+support+for+Bus+Eta+Bot+at+its+GitHub+repository+%5Bhere%5D%28$REPO_URL%29.", "$REPO_URL", url.QueryEscape(RepoURL), 1),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			message := MockMessageWithType(tc.ChatType)

			err := FeedbackCmdHandler(nil, &bot, &message)
			if err != nil {
				t.Fatalf("%v", err)
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

func TestShowFavouritesCmdHandler(t *testing.T) {
	t.Parallel()

	type TestCase struct {
		Name       string
		Favourites []string
		Expected   Request
	}

	cases := []TestCase{
		{
			Name:       "No favourites saved",
			Favourites: nil,
			Expected:   Request{Path: "/bot/sendMessage", Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&text=Oops%2C+you+haven%27t+saved+any+favourites+yet."},
		},
		{
			Name: "Show favourites keyboard",
			Favourites: []string{
				"96049 2 24",
				"83062 2 24",
			},
			Expected: Request{Path: "/bot/sendMessage", Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&reply_markup=%7B%22keyboard%22%3A%5B%5B%7B%22text%22%3A%2296049+2+24%22%2C%22request_contact%22%3Afalse%2C%22request_location%22%3Afalse%7D%5D%2C%5B%7B%22text%22%3A%2283062+2+24%22%2C%22request_contact%22%3Afalse%2C%22request_location%22%3Afalse%7D%5D%5D%2C%22resize_keyboard%22%3Atrue%2C%22one_time_keyboard%22%3Afalse%2C%22selective%22%3Afalse%7D&text=Favourites+keyboard+activated."},
		},
	}

	tgAPI, reqChan, errChan := NewMockTelegramAPIWithPath()
	defer tgAPI.Close()

	tg := &tgbotapi.BotAPI{
		APIEndpoint: tgAPI.URL + "/bot%s/%s",
		Client:      http.DefaultClient,
	}

	bot := NewBusEtaBot(handlers, tg, nil, nil, nil)

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			message := MockMessage()

			err := showFavourites(&bot, &message, tc.Favourites)
			if err != nil {
				t.Fatalf("%v", err)
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

func TestHideFavouritesCmdHandler(t *testing.T) {
	t.Parallel()

	type TestCase struct {
		Name       string
		Favourites []string
		Expected   Request
	}

	tgAPI, reqChan, errChan := NewMockTelegramAPIWithPath()
	defer tgAPI.Close()

	tg := &tgbotapi.BotAPI{
		APIEndpoint: tgAPI.URL + "/bot%s/%s",
		Client:      http.DefaultClient,
	}

	bot := NewBusEtaBot(handlers, tg, nil, nil, nil)

	message := MockMessage()

	err := HideFavouritesCmdHandler(nil, &bot, &message)
	if err != nil {
		t.Fatalf("%v", err)
	}

	select {
	case req := <-reqChan:
		actual := req
		expected := Request{Path: "/bot/sendMessage", Body: "chat_id=1&disable_notification=false&disable_web_page_preview=false&reply_markup=%7B%22remove_keyboard%22%3Atrue%2C%22selective%22%3Afalse%7D&text=Favourites+keyboard+hidden."}

		if actual != expected {
			fmt.Printf("Expected:\n%#v\nActual:\n%#v\n", expected, actual)
			t.Fail()
		}
	case err := <-errChan:
		t.Fatalf("%v", err)
	}
}
