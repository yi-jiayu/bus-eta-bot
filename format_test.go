package busetabot

import (
	"reflect"
	"testing"
	"text/template"
	"time"

	"github.com/kr/pretty"
	"github.com/stretchr/testify/assert"
	"github.com/yi-jiayu/datamall/v3"
)

var (
	baseTime = time.Unix(154454400, 0)

	busStop = BusStop{
		BusStopCode: "96049",
		RoadName:    "Upp Changi Rd East",
		Description: "Opp Tropicana Condo",
		Services:    []string{"2", "24"},
	}

	etaServicesPopulated = ETA{
		BusStop: busStop,
		Now:     baseTime,
		Services: []datamall.Service{
			{
				ServiceNo: "2",
				NextBus: datamall.ArrivingBus{
					EstimatedArrival: baseTime.Add(-100 * time.Second),
					Load:             "SDA",
					Type:             "DD",
				},
				NextBus2: datamall.ArrivingBus{
					EstimatedArrival: baseTime.Add(600 * time.Second),
					Load:             "SDA",
					Type:             "DD",
				},
				NextBus3: datamall.ArrivingBus{
					EstimatedArrival: baseTime.Add(2200 * time.Second),
					Load:             "LSD",
					Feature:          "WAB",
					Type:             "BD",
				},
			},
			{
				ServiceNo: "24",
				NextBus: datamall.ArrivingBus{
					EstimatedArrival: baseTime,
					Load:             "SEA",
					Type:             "SD",
				},
				NextBus2: datamall.ArrivingBus{
					EstimatedArrival: baseTime.Add(200 * time.Second),
					Load:             "SDA",
					Type:             "DD",
					Feature:          "WAB",
				},
				NextBus3: datamall.ArrivingBus{
					EstimatedArrival: baseTime.Add(400 * time.Second),
					Load:             "LSD",
					Type:             "BD",
				},
			},
		},
	}
	etaSomeServicesMissing = ETA{
		BusStop: busStop,
		Now:     baseTime,
		Services: []datamall.Service{
			{
				ServiceNo: "2",
				NextBus: datamall.ArrivingBus{
					EstimatedArrival: baseTime.Add(-100 * time.Second),
					Load:             "SDA",
					Type:             "DD",
				},
				NextBus2: datamall.ArrivingBus{
					EstimatedArrival: baseTime.Add(600 * time.Second),
					Load:             "SDA",
					Type:             "DD",
				},
				NextBus3: datamall.ArrivingBus{
					EstimatedArrival: baseTime.Add(2200 * time.Second),
					Load:             "LSD",
					Feature:          "WAB",
					Type:             "BD",
				},
			},
		},
	}
	etaSomeBusesEmpty = ETA{
		BusStop: busStop,
		Now:     baseTime,
		Services: []datamall.Service{
			{
				ServiceNo: "2",
				NextBus: datamall.ArrivingBus{
					EstimatedArrival: baseTime.Add(-100 * time.Second),
					Load:             "SDA",
					Type:             "DD",
				},
				NextBus2: datamall.ArrivingBus{
					EstimatedArrival: baseTime.Add(600 * time.Second),
					Load:             "SDA",
					Type:             "DD",
				},
			},
			{
				ServiceNo: "24",
				NextBus: datamall.ArrivingBus{
					EstimatedArrival: baseTime,
					Load:             "SEA",
					Type:             "SD",
				},
				NextBus2: datamall.ArrivingBus{
					EstimatedArrival: baseTime.Add(200 * time.Second),
					Load:             "SDA",
					Type:             "DD",
					Feature:          "WAB",
				},
			},
		},
	}
	etaServicesIsEmpty = ETA{
		BusStop: busStop,
		Now:     baseTime,
	}
	etaErrorNotNil = ETA{
		BusStop: busStop,
		Now:     baseTime,
		Error:   "Error fetching ETAs from LTA DataMall!",
	}
)

func TestSummaryFormatter(t *testing.T) {
	type testCase struct {
		Name     string
		ETA      ETA
		Expected string
	}
	testCases := []testCase{
		{
			Name:     "normal eta response",
			ETA:      etaServicesPopulated,
			Expected: "*Opp Tropicana Condo (96049)*\nUpp Changi Rd East\n```\n| Svc  | Nxt | 2nd | 3rd |\n|------|-----|-----|-----|\n| 2    |  -1 |  10 |  36 |\n| 24   |   0 |   3 |   6 |\n```\nShowing 2 out of 2 services for this bus stop.\n\n_Last updated on Sun, 24 Nov 74 00:00 SGT_",
		},
		{
			Name:     "services is empty",
			ETA:      etaServicesIsEmpty,
			Expected: "*Opp Tropicana Condo (96049)*\nUpp Changi Rd East\nNo ETAs available.\n\n_Last updated on Sun, 24 Nov 74 00:00 SGT_",
		},
		{
			Name:     "error is non-nil",
			ETA:      etaErrorNotNil,
			Expected: "*Opp Tropicana Condo (96049)*\nUpp Changi Rd East\nError fetching ETAs from LTA DataMall!\n\n_Last updated on Sun, 24 Nov 74 00:00 SGT_",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			actual, err := summaryFormatter.Format(tc.ETA)
			if err != nil {
				t.Fatal(err)
			}
			if !assert.Equal(t, tc.Expected, actual) {
				pretty.Println(actual)
			}
		})
	}
}

func TestFeaturesFormatter(t *testing.T) {
	type testCase struct {
		Name     string
		ETA      ETA
		Expected string
	}
	testCases := []testCase{
		{
			Name:     "normal eta response",
			ETA:      etaServicesPopulated,
			Expected: "*Opp Tropicana Condo (96049)*\nUpp Changi Rd East\n```\nSvc   Eta  Sea  Typ  Fea\n---   ---  ---  ---  ---\n2      -1  SDA   DD     \n24      0  SEA   SD     \n24      3  SDA   DD  WAB\n24      6  LSD   BD     \n2      10  SDA   DD     \n2      36  LSD   BD  WAB\n```\nShowing 2 out of 2 services for this bus stop.\n\n_Last updated on Sun, 24 Nov 74 00:00 SGT_",
		},
		{
			Name:     "services is empty",
			ETA:      etaServicesIsEmpty,
			Expected: "*Opp Tropicana Condo (96049)*\nUpp Changi Rd East\nNo ETAs available.\n\n_Last updated on Sun, 24 Nov 74 00:00 SGT_",
		},
		{
			Name:     "error is non-nil",
			ETA:      etaErrorNotNil,
			Expected: "*Opp Tropicana Condo (96049)*\nUpp Changi Rd East\nError fetching ETAs from LTA DataMall!\n\n_Last updated on Sun, 24 Nov 74 00:00 SGT_",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			actual, err := featuresFormatter.Format(tc.ETA)
			if err != nil {
				t.Fatal(err)
			}
			if !assert.Equal(t, tc.Expected, actual) {
				pretty.Println(actual)
			}
		})
	}
}

func TestMessageTemplate(t *testing.T) {
	type testCase struct {
		Name     string
		ETA      ETA
		Expected string
	}
	testCases := []testCase{
		{
			Name:     "normal eta response",
			ETA:      etaServicesPopulated,
			Expected: "Header\nBody\nFooter",
		},
		{
			Name:     "services is empty",
			ETA:      ETA{},
			Expected: "Header\nNo ETAs available.\nFooter",
		},
		{
			Name: "error is non-nil",
			ETA: ETA{
				Error: "Error fetching ETAs from LTA DataMall!",
			},
			Expected: "Header\nError fetching ETAs from LTA DataMall!\nFooter",
		},
	}
	formatter := TemplateFormatter{
		template: template.Must(template.New("message.tmpl").
			Funcs(funcMap).
			ParseFiles("templates/message.tmpl")),
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			actual, err := formatter.Format(tc.ETA)
			if err != nil {
				t.Fatal(err)
			}
			if !assert.Equal(t, tc.Expected, actual) {
				pretty.Println(actual)
			}
		})
	}
}

func TestSummaryTemplate(t *testing.T) {
	type testCase struct {
		Name     string
		ETA      ETA
		Expected string
	}
	testCases := []testCase{
		{
			Name:     "services are fully populated",
			ETA:      etaServicesPopulated,
			Expected: "```\n| Svc  | Nxt | 2nd | 3rd |\n|------|-----|-----|-----|\n| 2    |  -1 |  10 |  36 |\n| 24   |   0 |   3 |   6 |\n```\nShowing 2 out of 2 services for this bus stop.",
		},
		{
			Name:     "some buses have no data",
			ETA:      etaSomeBusesEmpty,
			Expected: "```\n| Svc  | Nxt | 2nd | 3rd |\n|------|-----|-----|-----|\n| 2    |  -1 |  10 |   ? |\n| 24   |   0 |   3 |   ? |\n```\nShowing 2 out of 2 services for this bus stop.",
		},
	}
	formatter := TemplateFormatter{
		template: template.Must(template.New("body_test.tmpl").
			Funcs(funcMap).
			ParseFiles("templates/body_test.tmpl", "templates/partials/summary.tmpl", "templates/partials/services_count.tmpl")),
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			actual, err := formatter.Format(tc.ETA)
			if err != nil {
				t.Fatal(err)
			}
			if !assert.Equal(t, tc.Expected, actual) {
				pretty.Println(actual)
			}
		})
	}
}

func TestFeaturesTemplate(t *testing.T) {
	type testCase struct {
		Name     string
		ETA      ETA
		Expected string
	}
	testCases := []testCase{
		{
			Name:     "services are fully populated",
			ETA:      etaServicesPopulated,
			Expected: "```\nSvc   Eta  Sea  Typ  Fea\n---   ---  ---  ---  ---\n2      -1  SDA   DD     \n24      0  SEA   SD     \n24      3  SDA   DD  WAB\n24      6  LSD   BD     \n2      10  SDA   DD     \n2      36  LSD   BD  WAB\n```\nShowing 2 out of 2 services for this bus stop.",
		},
		{
			Name:     "some buses have no data",
			ETA:      etaSomeBusesEmpty,
			Expected: "```\nSvc   Eta  Sea  Typ  Fea\n---   ---  ---  ---  ---\n2      -1  SDA   DD     \n24      0  SEA   SD     \n24      3  SDA   DD  WAB\n2      10  SDA   DD     \n```\nShowing 2 out of 2 services for this bus stop.",
		},
	}
	formatter := TemplateFormatter{
		template: template.Must(template.New("body_test.tmpl").
			Funcs(funcMap).
			ParseFiles("templates/body_test.tmpl", "templates/partials/features.tmpl", "templates/partials/services_count.tmpl")),
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			actual, err := formatter.Format(tc.ETA)
			if err != nil {
				t.Fatal(err)
			}
			if !assert.Equal(t, tc.Expected, actual) {
				pretty.Println(actual)
			}
		})
	}
}

func TestHeaderTemplate(t *testing.T) {
	type testCase struct {
		Name     string
		ETA      ETA
		Expected string
	}
	testCases := []testCase{
		{
			Name: "description present",
			ETA: ETA{
				BusStop: busStop,
			},
			Expected: "*Opp Tropicana Condo (96049)*\nUpp Changi Rd East",
		},
		{
			Name: "description not present",
			ETA: ETA{
				BusStop: BusStop{
					BusStopCode: "96049",
				},
			},
			Expected: "*96049*",
		},
	}
	formatter := TemplateFormatter{
		template: template.Must(template.New("header_test.tmpl").
			Funcs(funcMap).
			ParseFiles("templates/header_test.tmpl", "templates/partials/header.tmpl")),
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			actual, err := formatter.Format(tc.ETA)
			if err != nil {
				t.Fatal(err)
			}
			if !assert.Equal(t, tc.Expected, actual) {
				pretty.Println(actual)
			}
		})
	}
}

func TestFooterTemplate(t *testing.T) {
	type testCase struct {
		Name     string
		ETA      ETA
		Expected string
	}
	testCases := []testCase{
		{
			Name:     "all services present",
			ETA:      etaServicesPopulated,
			Expected: "\n_Last updated on Sun, 24 Nov 74 00:00 SGT_",
		},
		{
			Name:     "not all services present",
			ETA:      etaSomeServicesMissing,
			Expected: "\n_Last updated on Sun, 24 Nov 74 00:00 SGT_",
		},
		{
			Name:     "services is empty",
			ETA:      etaServicesIsEmpty,
			Expected: "\n_Last updated on Sun, 24 Nov 74 00:00 SGT_",
		},
	}
	formatter := TemplateFormatter{
		template: template.Must(template.New("footer_test.tmpl").
			Funcs(funcMap).
			ParseFiles("templates/footer_test.tmpl", "templates/partials/footer.tmpl")),
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			actual, err := formatter.Format(tc.ETA)
			if err != nil {
				t.Fatal(err)
			}
			if !assert.Equal(t, tc.Expected, actual) {
				pretty.Println(actual)
			}
		})
	}
}

func TestServicesCountTemplate(t *testing.T) {
	type testCase struct {
		Name     string
		ETA      ETA
		Expected string
	}
	testCases := []testCase{
		{
			Name:     "all services present",
			ETA:      etaServicesPopulated,
			Expected: "Showing 2 out of 2 services for this bus stop.",
		},
		{
			Name:     "not all services present",
			ETA:      etaSomeServicesMissing,
			Expected: "Showing 1 out of 2 services for this bus stop.",
		},
		{
			Name:     "services is empty",
			ETA:      etaServicesIsEmpty,
			Expected: "",
		},
	}
	formatter := TemplateFormatter{
		template: template.Must(template.New("services_count_test.tmpl").
			Funcs(funcMap).
			ParseFiles("templates/services_count_test.tmpl", "templates/partials/services_count.tmpl")),
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			actual, err := formatter.Format(tc.ETA)
			if err != nil {
				t.Fatal(err)
			}
			if !assert.Equal(t, tc.Expected, actual) {
				pretty.Println(actual)
			}
		})
	}

}

func Test_otherServices(t *testing.T) {
	type args struct {
		stop     BusStop
		services []datamall.Service
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "all services at the bus stop are included",
			args: args{
				stop: BusStop{
					Services: []string{"2", "5", "24"},
				},
				services: []datamall.Service{
					{
						ServiceNo: "2",
					},
					{
						ServiceNo: "5",
					},
					{
						ServiceNo: "24",
					},
				},
			},
			want: nil,
		},
		{
			name: "some services at the bus stop are included",
			args: args{
				stop: BusStop{
					Services: []string{"2", "5", "24"},
				},
				services: []datamall.Service{
					{
						ServiceNo: "2",
					},
					{
						ServiceNo: "5",
					},
				},
			},
			want: []string{"24"},
		},
		{
			name: "no services at the bus stop are included",
			args: args{
				stop: BusStop{
					Services: []string{"2", "5", "24"},
				},
				services: nil,
			},
			want: []string{"2", "5", "24"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := otherServices(tt.args.stop, tt.args.services); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("otherServices() = %v, want %v", got, tt.want)
			}
		})
	}
}
