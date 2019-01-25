package busetabot

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/yi-jiayu/datamall/v2"
	"github.com/yi-jiayu/telegram-bot-api"

	"github.com/yi-jiayu/bus-eta-bot/v4/telegram"
)

var collapseWhitespaceRegex = regexp.MustCompile(`\s+`)
var illegalCharsRegex = regexp.MustCompile(`[^A-Z0-9 ]`)

var (
	errBusStopIDTooLong = errors.New("bus stop id too long")
	errBusStopIDInvalid = errors.New("bus stop id invalid")
)

var (
	sgt = time.FixedZone("SGT", 8*3600)
)

// BusEtas represents the calculated time before buses arrive at a bus stop
type BusEtas struct {
	BusStopID   string
	UpdatedTime time.Time
	Services    []IncomingBuses
}

// IncomingBuses contains information about incoming buses for a single service at a bus stop.
type IncomingBuses struct {
	ServiceNo string
	Etas      [3]string
	Loads     [3]string
	Features  [3]string
	Types     [3]string
}

// CallbackData represents the data to be included with the Refresh inline keyboard button in eta messages.
type CallbackData struct {
	Type       string   `json:"t"`
	BusStopID  string   `json:"b,omitempty"`
	ServiceNos []string `json:"s,omitempty"`
	Argstr     string   `json:"a,omitempty"`
}

// InferEtaQuery extracts a bus stop ID and service numbers from a text message.
func InferEtaQuery(text string) (string, []string, error) {
	if len(text) > 30 {
		text = text[:30]
	}

	text = strings.ToUpper(text)
	tokens := collapseWhitespaceRegex.Split(text, -1)
	busStopID, serviceNos := tokens[0], tokens[1:]

	if len(busStopID) > 5 {
		return busStopID, serviceNos, errBusStopIDTooLong
	}

	if illegalCharsRegex.MatchString(busStopID) {
		return busStopID, serviceNos, errBusStopIDInvalid
	}

	return busStopID, serviceNos, nil
}

// CalculateEtas calculates the time before buses arrive from the LTA DataMall bus arrival response
func CalculateEtas(t time.Time, busArrival datamall.BusArrival) (BusEtas, error) {
	services := make([]IncomingBuses, 0)

	for _, service := range busArrival.Services {
		incomingBuses := IncomingBuses{
			ServiceNo: service.ServiceNo,
		}

		placeholder := "?"

		for i := 0; i < 3; i++ {
			var incomingBus datamall.ArrivingBus
			switch i {
			case 0:
				incomingBus = service.NextBus
			case 1:
				incomingBus = service.NextBus2
			case 2:
				incomingBus = service.NextBus3
			}

			if estArrival := incomingBus.EstimatedArrival; estArrival != "" {
				eta, err := time.Parse(time.RFC3339, estArrival)
				if err != nil {
					return BusEtas{}, errors.Wrap(err, "invalid timestamp in datmall response")
				}

				diff := (eta.Unix() - t.Unix()) / 60
				incomingBuses.Etas[i] = fmt.Sprintf("%d", diff)
			} else {
				incomingBuses.Etas[i] = placeholder
			}

			incomingBuses.Loads[i] = incomingBus.Load
			incomingBuses.Features[i] = incomingBus.Feature
			incomingBuses.Types[i] = incomingBus.Type
		}

		services = append(services, incomingBuses)
	}

	return BusEtas{
		BusStopID:   busArrival.BusStopCode,
		UpdatedTime: t,
		Services:    services,
	}, nil
}

func contains(serviceNos []string, serviceNo string) bool {
	for _, s := range serviceNos {
		if s == serviceNo {
			return true
		}
	}

	return false
}

type byServiceNo [][4]string

func (s byServiceNo) Len() int           { return len(s) }
func (s byServiceNo) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byServiceNo) Less(i, j int) bool { return s[i][0] < s[j][0] }

// FormatEtasMultiple formats bus etas nicely to be sent in a message
func FormatEtasMultiple(busEtas BusEtas, busStop *BusStop, serviceNos []string) string {
	showing := 0
	services := make([][4]string, 0)
	for _, service := range busEtas.Services {
		serviceNo := service.ServiceNo
		if serviceNos != nil && len(serviceNos) != 0 && !contains(serviceNos, serviceNo) {
			continue
		}

		arrivals := [4]string{serviceNo, service.Etas[0], service.Etas[1], service.Etas[2]}
		services = append(services, arrivals)
		showing++
	}

	sort.Sort(byServiceNo(services))

	table := EtaTable(services)

	var header string
	if busStop.Description != "" {
		header = fmt.Sprintf("*%s (%s)*\n", busStop.Description, busStop.BusStopCode)
	} else {
		header = fmt.Sprintf("*%s*\n", busStop.BusStopCode)
	}

	if busStop.RoadName != "" {
		header += fmt.Sprintf("%s\n", busStop.RoadName)
	}

	shown := fmt.Sprintf("Showing %d out of %d services for this bus stop.", showing, len(busEtas.Services))

	updated := fmt.Sprintf("Last updated at %s", busEtas.UpdatedTime.In(sgt).Format(time.RFC822))

	formatted := fmt.Sprintf("%s```\n%s```\n%s\n\n_%s_", header, table, shown, updated)

	return formatted
}

func padLeft(s string, l int, c string) string {
	for {
		if len(s) >= l {
			return s
		}

		s = c + s
	}
}

func padRight(s string, l int, c string) string {
	for {
		if len(s) >= l {
			return s
		}

		s += c
	}
}

// EtaTable generates the table of bus etas used in a eta message
func EtaTable(etas [][4]string) string {
	table := [][4]string{
		{"Svc", "Next", " 2nd", " 3rd"},
	}

	for _, eta := range etas {
		table = append(table, eta)
	}

	maxWidths := make([]int, len(table[0]))
	for _, row := range table {
		for pos, col := range row {
			if len(col) > maxWidths[pos] {
				maxWidths[pos] = len(col)
			}
		}
	}

	output := ""
	for i, row := range table {
		if i == 1 {
			for pos := range row {
				switch pos {
				case 0:
					output += "|-" + padRight("", maxWidths[pos], "-")
				case 3:
					output += "-|-" + padRight("", maxWidths[pos], "-") + "-|"
				default:
					output += "-|-" + padRight("", maxWidths[pos], "-")
				}
			}

			output += "\n"
		}

		for pos, col := range row {
			switch pos {
			case 0:
				output += "| " + padRight(col, maxWidths[pos], " ")
			case 3:
				output += " | " + padLeft(col, maxWidths[pos], " ") + " |"
			default:
				output += " | " + padLeft(col, maxWidths[pos], " ")
			}
		}

		output += "\n"
	}

	// remove final newline
	return output[:len(output)-1]
}

// EtaMessageText generates and returns the text for an eta message
func EtaMessageText(bot *BusEtaBot, busStopCode string, serviceNos []string) (string, error) {
	busArrival, err := bot.Datamall.GetBusArrival(busStopCode, "")
	if err != nil {
		if err, ok := err.(datamall.Error); ok {
			return fmt.Sprintf("Oh no! The LTA DataMall API that Bus Eta Bot relies on appears to be down at the moment (it returned HTTP status code %d).", err.StatusCode), nil
		}
		return "", errors.Wrap(err, "error getting etas from datamall")
	}

	etas, err := CalculateEtas(bot.NowFunc(), busArrival)
	if err != nil {
		return "", err
	}

	busStop := bot.BusStops.Get(busStopCode)
	if busStop == nil {
		if len(etas.Services) == 0 {
			return fmt.Sprintf("Oh no! I couldn't find any information about bus stop %s.", busStopCode), nil
		}
		busStop = &BusStop{
			BusStopCode: busStopCode,
		}
	}

	msg := FormatEtasMultiple(etas, busStop, serviceNos)
	return msg, nil
}

// EtaMessageReplyMarkup generates the reply markup for an eta message, including a resend callback button only when
// the message is not an inline message.
func EtaMessageReplyMarkup(busStopID string, serviceNos []string, inline bool) (*tgbotapi.InlineKeyboardMarkup, error) {
	refreshCallbackData := CallbackData{
		Type:       "refresh",
		BusStopID:  busStopID,
		ServiceNos: serviceNos,
	}

	refreshCallbackDataJSON, err := json.Marshal(refreshCallbackData)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("error marshalling callback data: %#v", refreshCallbackData))
	}
	refreshCallbackDataJSONStr := string(refreshCallbackDataJSON)

	replyMarkup := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.InlineKeyboardButton{
					Text:         "Refresh",
					CallbackData: &refreshCallbackDataJSONStr,
				},
			},
		},
	}

	if !inline {
		resendCallbackData := CallbackData{
			Type:       "resend",
			BusStopID:  busStopID,
			ServiceNos: serviceNos,
		}

		resendCallbackDataJSON, err := json.Marshal(resendCallbackData)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("error marshalling callback data: %#v", refreshCallbackData))
		}
		resendCallbackDataJSONStr := string(resendCallbackDataJSON)

		replyMarkup.InlineKeyboard[0] = append(replyMarkup.InlineKeyboard[0], tgbotapi.InlineKeyboardButton{
			Text:         "Resend",
			CallbackData: &resendCallbackDataJSONStr,
		})

		argstr := busStopID
		if len(serviceNos) > 0 {
			argstr += " " + strings.Join(serviceNos, " ")
		}
		toggleFavouriteCallbackData := CallbackData{
			Type:   "togf",
			Argstr: argstr,
		}
		addFavouriteCallbackDataJSON, err := json.Marshal(toggleFavouriteCallbackData)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("error marshalling callback data: %#v", refreshCallbackData))
		}
		addFavouriteCallbackDataJSONStr := string(addFavouriteCallbackDataJSON)

		replyMarkup.InlineKeyboard[0] = append(replyMarkup.InlineKeyboard[0], tgbotapi.InlineKeyboardButton{
			Text:         "⭐",
			CallbackData: &addFavouriteCallbackDataJSONStr,
		})
	}

	return &replyMarkup, nil
}

func NewRefreshButton(busStopCode string, serviceNos []string) telegram.InlineKeyboardButton {
	data := CallbackData{
		Type:       "refresh",
		BusStopID:  busStopCode,
		ServiceNos: serviceNos,
	}
	JSON, _ := json.Marshal(data)
	return telegram.InlineKeyboardButton{
		Text:         "Refresh",
		CallbackData: string(JSON),
	}
}

func NewResendButton(busStopCode string, serviceNos []string) telegram.InlineKeyboardButton {
	data := CallbackData{
		Type:       "resend",
		BusStopID:  busStopCode,
		ServiceNos: serviceNos,
	}
	JSON, _ := json.Marshal(data)
	return telegram.InlineKeyboardButton{
		Text:         "Resend",
		CallbackData: string(JSON),
	}
}

func NewToggleFavouriteButton(busStopCode string, serviceNos []string) telegram.InlineKeyboardButton {
	argstr := busStopCode
	if len(serviceNos) > 0 {
		argstr += " " + strings.Join(serviceNos, " ")
	}
	data := CallbackData{
		Type:   "togf",
		Argstr: argstr,
	}
	JSON, _ := json.Marshal(data)
	return telegram.InlineKeyboardButton{
		Text:         "⭐",
		CallbackData: string(JSON),
	}
}

func NewETAMessageReplyMarkup(busStopCode string, serviceNos []string, inline bool) telegram.InlineKeyboardMarkup {
	row := []telegram.InlineKeyboardButton{
		NewRefreshButton(busStopCode, serviceNos),
	}
	if !inline {
		row = append(row,
			NewResendButton(busStopCode, serviceNos),
			NewToggleFavouriteButton(busStopCode, serviceNos))
	}
	return telegram.InlineKeyboardMarkup{
		InlineKeyboard: [][]telegram.InlineKeyboardButton{
			row,
		},
	}
}
