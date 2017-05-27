package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/yi-jiayu/datamall"
	"golang.org/x/net/context"
)

var collapseWhitespaceRegex = regexp.MustCompile(`\s+`)
var illegalCharsRegex = regexp.MustCompile(`[^A-Z0-9 ]`)

var (
	errBusStopIDTooLong = errors.New("bus stop id too long")
	errBusStopIDInvalid = errors.New("bus stop id invalid")
)

// BusEtas represents the calculated time before buses arrive at a bus stop
type BusEtas struct {
	BusStopID   string
	UpdatedTime time.Time
	Services    [][4]string
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
	services := make([][4]string, 0)
	for _, service := range busArrival.Services {
		etas := [4]string{service.ServiceNo}

		var placeholder string
		if service.Status == "Not In Operation" {
			placeholder = "-"
		} else {
			placeholder = "?"
		}

		if next := service.NextBus.EstimatedArrival; next != "" {
			eta, err := time.Parse(time.RFC3339, next)
			if err != nil {
				return BusEtas{}, err
			}

			diff := (eta.Unix() - t.Unix()) / 60
			etas[1] = fmt.Sprintf("%d", diff)
		} else {
			etas[1] = placeholder
		}

		if next2 := service.SubsequentBus.EstimatedArrival; next2 != "" {
			eta, err := time.Parse(time.RFC3339, next2)
			if err != nil {
				return BusEtas{}, err
			}

			diff := (eta.Unix() - t.Unix()) / 60
			etas[2] = fmt.Sprintf("%d", diff)
		} else {
			etas[2] = placeholder
		}

		if next3 := service.SubsequentBus3.EstimatedArrival; next3 != "" {
			eta, err := time.Parse(time.RFC3339, next3)
			if err != nil {
				return BusEtas{}, err
			}

			diff := (eta.Unix() - t.Unix()) / 60
			etas[3] = fmt.Sprintf("%d", diff)
		} else {
			etas[3] = placeholder
		}

		services = append(services, etas)
	}

	return BusEtas{
		BusStopID:   busArrival.BusStopID,
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

// FormatEtas formats bus etas nicely to be sent in a message
func FormatEtas(busEtas BusEtas, busStop *BusStop, serviceNos []string) string {
	showing := 0
	services := make([][4]string, 0)
	for _, etas := range busEtas.Services {
		serviceNo := etas[0]
		if serviceNos != nil && len(serviceNos) != 0 && !contains(serviceNos, serviceNo) {
			continue
		}

		services = append(services, etas)
		showing++
	}

	sort.Sort(byServiceNo(services))

	table := EtaTable(services)

	var header string
	if busStop.Description != "" {
		header = fmt.Sprintf("*%s (%s)*\n", busStop.Description, busStop.BusStopID)
	} else {
		header = fmt.Sprintf("*%s*\n", busStop.BusStopID)
	}

	if busStop.Road != "" {
		header += fmt.Sprintf("%s\n", busStop.Road)
	}

	shown := fmt.Sprintf("Showing %d out of %d services for this bus stop.", showing, len(busEtas.Services))

	sgt := time.FixedZone("SGT", 8*3600)
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

// EtaMessage generates and returns the text for an eta message
func EtaMessage(ctx context.Context, bot *BusEtaBot, busStopID string, serviceNos []string) (string, error) {
	busArrival, err := bot.Datamall.GetBusArrival(busStopID, nil)
	if err != nil {
		return "", errors.Wrap(err, "error getting etas from datamall")
	}

	etas, err := CalculateEtas(bot.NowFunc(), busArrival)
	if err != nil {
		return "", err
	}

	busStop, err := GetBusStop(ctx, busStopID)
	if err != nil {
		if err == errNotFound {
			if len(etas.Services) == 0 {
				return fmt.Sprintf("Oh no! I couldn't find any information about bus stop %s.", busStopID), err
			}

			busStop.BusStopID = busStopID
		} else {
			return "", err
		}
	}

	msg := FormatEtas(etas, &busStop, serviceNos)
	return msg, nil
}
