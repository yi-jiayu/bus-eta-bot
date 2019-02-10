package busetabot

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/yi-jiayu/datamall/v3"
	"google.golang.org/appengine"

	"github.com/yi-jiayu/bus-eta-bot/v4/telegram"
)

var (
	errInvalidBusStopCode = errors.New("bus stop code invalid")
)

var (
	sgt = time.FixedZone("SGT", 8*3600)
)

type ETA struct {
	BusStop  BusStop
	Now      time.Time
	Services []datamall.Service
	Error    string
}

type FormatterFactory interface {
	GetFormatter(ctx context.Context, userID int) Formatter
}

type BusStopGetter interface {
	Get(ID string) *BusStop
}

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

type ETAFormatter interface {
	Format(etas BusEtas, services []string) string
}

// SummaryETAFormatter displays ETAs for up to the next 3 buses at a bus stop.
type SummaryETAFormatter struct{}

type ETARequest struct {
	UserID   int
	Time     time.Time
	Code     string
	Services []string
}

type ETAMessageFactory struct {
	busStopGetter    BusStopGetter
	etaService       ETAService
	formatterFactory FormatterFactory
}

func (f ETAMessageFactory) Text(ctx context.Context, request ETARequest) (string, error) {
	var wg sync.WaitGroup
	var eta ETA
	var formatter Formatter
	wg.Add(2)
	go func() {
		eta = NewETA(ctx, f.busStopGetter, f.etaService, request)
		wg.Done()
	}()
	go func() {
		formatter = f.formatterFactory.GetFormatter(ctx, request.UserID)
		wg.Done()
	}()
	wg.Wait()
	return formatter.Format(eta)
}

func NewETA(ctx context.Context, busStopGetter BusStopGetter, etaService ETAService, request ETARequest) (eta ETA) {
	eta.Now = request.Time
	stop := busStopGetter.Get(request.Code)
	if stop != nil {
		eta.BusStop = *stop
	} else {
		eta.BusStop.BusStopCode = request.Code
	}
	arrival, err := etaService.GetBusArrival(request.Code, "")
	if err != nil {
		if err, ok := err.(*datamall.Error); ok {
			eta.Error = fmt.Sprintf("LTA DataMall could be down at the moment (status code %d)", err.StatusCode)
			// TODO: record metrics about DataMall disruptions
		} else {
			eta.Error = fmt.Sprintf("An error occurred while fetching ETAs (request ID: %s)", appengine.RequestID(ctx))
			logError(ctx, err)
		}
		return
	}
	if len(request.Services) == 0 {
		eta.Services = arrival.Services
	} else {
		for _, svc := range arrival.Services {
			if contains(request.Services, svc.ServiceNo) {
				eta.Services = append(eta.Services, svc)
			}
		}
	}
	return
}

func (SummaryETAFormatter) Format(etas BusEtas, serviceNos []string) string {
	showing := 0
	services := make([][4]string, 0)
	for _, service := range etas.Services {
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

	var servicesNoun string
	if len(etas.Services) > 1 {
		servicesNoun = "services"
	} else {
		servicesNoun = "service"
	}
	shown := fmt.Sprintf("Showing %d out of %d %s for this bus stop.", showing, len(etas.Services), servicesNoun)

	// the first line break immediately after the opening triple backticks is needed
	// otherwise the first pipe gets swallowed
	formatted := fmt.Sprintf("```\n%s```\n%s", table, shown)

	return formatted
}

func ETAMessageText(busStops BusStopRepository, etaService ETAService, formatter ETAFormatter, t time.Time, code string, services []string) (string, error) {
	stop := busStops.Get(code)
	header := "*" + code + "*"
	if stop != nil {
		header = fmt.Sprintf("*%s (%s)*\n%s", stop.Description, stop.BusStopCode, stop.RoadName)
	}
	var body string
	arrival, err := etaService.GetBusArrival(code, "")
	if err != nil {
		if err, ok := err.(*datamall.Error); ok {
			body = fmt.Sprintf("\nOh no! The LTA DataMall API that Bus Eta Bot relies on appears to be down at the moment (it returned HTTP status code %d).", err.StatusCode)
		} else {
			return "", errors.Wrap(err, "error getting etas from datamall")
		}
	} else {
		if stop == nil && len(arrival.Services) == 0 {
			body = "\nNo etas found for this bus stop."
		} else {
			etas := CalculateEtas(t, arrival)
			body = formatter.Format(etas, services)
		}
	}
	timestamp := fmt.Sprintf("Last updated at %s", t.In(sgt).Format(time.RFC822))
	return fmt.Sprintf("%s\n%s\n\n_%s_", header, body, timestamp), nil
}

// InferEtaQuery extracts a bus stop ID and service numbers from a text message.
func InferEtaQuery(text string) (string, []string, error) {
	m := busStopRegex.FindStringSubmatch(text)
	if m == nil {
		return "", nil, errInvalidBusStopCode
	}
	code := m[1]
	rest := strings.TrimPrefix(text, code)
	services := strings.Fields(rest)
	return code, services, nil
}

// CalculateEtas calculates the time before buses arrive from the LTA DataMall bus arrival response
func CalculateEtas(t time.Time, busArrival datamall.BusArrival) BusEtas {
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

			if estArrival := incomingBus.EstimatedArrival; estArrival != (time.Time{}) {
				diff := (estArrival.Unix() - t.Unix()) / 60
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
	}
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
