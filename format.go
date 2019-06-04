package busetabot

import (
	"bytes"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/yi-jiayu/datamall/v3"
)

var (
	funcMap = map[string]interface{}{
		"join":          strings.Join,
		"until":         minutesUntil,
		"arrivingBuses": arrivingBuses,
		"sortByArrival": sortByArrival,
		"sortByService": sortByService,
		"inSGT":         inSGT,
		"otherServices": otherServices,
		"take":          take,
	}
)

var (
	FeaturesFormatter = TemplateFormatter{
		template: template.Must(template.New("message.tmpl").
			Funcs(funcMap).
			ParseFiles("templates/message.tmpl",
				"templates/partials/header.tmpl",
				"templates/partials/features.tmpl",
				"templates/partials/services_count.tmpl",
				"templates/partials/footer.tmpl")),
	}
	SummaryFormatter = TemplateFormatter{
		template: template.Must(template.New("message.tmpl").
			Funcs(funcMap).
			ParseFiles("templates/message.tmpl",
				"templates/partials/header.tmpl",
				"templates/partials/summary.tmpl",
				"templates/partials/footer.tmpl",
				"templates/partials/services_count.tmpl")),
	}
)

type ArrivingBus struct {
	ServiceNo string
	datamall.ArrivingBus
}

type Formatter interface {
	Format(eta ETA) (string, error)
}

type TemplateFormatter struct {
	template *template.Template
}

func (f TemplateFormatter) Format(eta ETA) (string, error) {
	b := new(bytes.Buffer)
	err := f.template.Execute(b, eta)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

// take returns up to the first n elements of coll.
func take(n int, coll []ArrivingBus) []ArrivingBus {
	if n <= 0 || n > len(coll) {
		n = len(coll)
	}
	return coll[:n]
}

// minutesUntil returns the number of minutes from now until then.
func minutesUntil(now time.Time, then time.Time) string {
	if then.IsZero() {
		return "?"
	}
	return strconv.Itoa(int(then.Sub(now).Minutes()))
}

func arrivingBuses(services []datamall.Service) []ArrivingBus {
	var buses []ArrivingBus
	for _, service := range services {
		if !service.NextBus.EstimatedArrival.IsZero() {
			buses = append(buses, ArrivingBus{
				ServiceNo:   service.ServiceNo,
				ArrivingBus: service.NextBus,
			})
		}
		if !service.NextBus2.EstimatedArrival.IsZero() {
			buses = append(buses, ArrivingBus{
				ServiceNo:   service.ServiceNo,
				ArrivingBus: service.NextBus2,
			})
		}
		if !service.NextBus3.EstimatedArrival.IsZero() {
			buses = append(buses, ArrivingBus{
				ServiceNo:   service.ServiceNo,
				ArrivingBus: service.NextBus3,
			})
		}
	}
	return buses
}

func sortByArrival(buses []ArrivingBus) []ArrivingBus {
	sort.Slice(buses, func(i, j int) bool {
		return buses[i].EstimatedArrival.Before(buses[j].EstimatedArrival)
	})
	return buses
}

func sortByService(services []datamall.Service) []datamall.Service {
	sort.Slice(services, func(i, j int) bool {
		return services[i].ServiceNo < services[j].ServiceNo
	})
	return services
}

func inSGT(t time.Time) string {
	return t.In(sgt).Format("Mon, 02 Jan 06 15:04 MST")
}

func otherServices(stop BusStop, services []datamall.Service) []string {
	var others []string
	contains := func(serviceNo string, services []datamall.Service) bool {
		for _, service := range services {
			if serviceNo == service.ServiceNo {
				return true
			}
		}
		return false
	}
	for _, serviceNo := range stop.Services {
		if !contains(serviceNo, services) {
			others = append(others, serviceNo)
		}
	}
	return others
}
