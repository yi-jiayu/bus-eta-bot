package busetabot

import (
	"context"
	"errors"
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yi-jiayu/datamall/v3"

	"github.com/yi-jiayu/bus-eta-bot/v5/telegram"
)

type mockBusStopRepository struct {
	BusStop        *BusStop
	NearbyBusStops []BusStop
}

func (b mockBusStopRepository) Search(ctx context.Context, query string, limit int) []BusStop {
	panic("implement me")
}

func (b mockBusStopRepository) Nearby(ctx context.Context, lat, lon, radius float64, limit int) (nearby []NearbyBusStop) {
	for _, bs := range b.NearbyBusStops {
		nearby = append(nearby, NearbyBusStop{
			BusStop:  bs,
			Distance: math.Sqrt(SquaredEuclideanDistanceAtEquator(lat, lon, bs.Latitude, bs.Longitude)),
		})
	}
	return
}

func (b mockBusStopRepository) Get(ID string) *BusStop {
	return b.BusStop
}

type mockETAService struct {
	BusArrival datamall.BusArrival
	Error      error
}

func (s mockETAService) GetBusArrival(string, string) (datamall.BusArrival, error) {
	return s.BusArrival, s.Error
}

type mockDatamall struct {
}

func (d mockDatamall) GetBusArrival(code string, serviceNo string) (datamall.BusArrival, error) {
	return newArrival(time.Time{}, code), nil
}

func newArrival(t time.Time, code string) datamall.BusArrival {
	return datamall.BusArrival{
		BusStopCode: code,
		Services: []datamall.Service{
			{
				ServiceNo: "2",
				NextBus: datamall.ArrivingBus{
					EstimatedArrival: t.Add(-100 * time.Second),
					Load:             "SDA",
					Type:             "DD",
				},
				NextBus2: datamall.ArrivingBus{
					EstimatedArrival: t.Add(600 * time.Second),
					Load:             "SDA",
					Type:             "DD",
				},
				NextBus3: datamall.ArrivingBus{
					EstimatedArrival: t.Add(2200 * time.Second),
					Load:             "LSD",
					Feature:          "WAB",
					Type:             "BD",
				},
			},
			{
				ServiceNo: "24",
				NextBus: datamall.ArrivingBus{
					EstimatedArrival: t.Add(100 * time.Second),
					Load:             "SEA",
					Type:             "SD",
				},
				NextBus2: datamall.ArrivingBus{
					EstimatedArrival: t.Add(200 * time.Second),
					Load:             "SDA",
					Type:             "DD",
					Feature:          "WAB",
				},
				NextBus3: datamall.ArrivingBus{
					EstimatedArrival: t.Add(400 * time.Second),
					Load:             "LSD",
					Type:             "BD",
				},
			},
		},
	}
}

func TestInferEtaQuery(t *testing.T) {
	testCases := []struct {
		Name     string
		Text     string
		Expected struct {
			BusStopID  string
			ServiceNos []string
			Err        error
		}
	}{
		{
			Name: "Bus stop ID only",
			Text: "96049",
			Expected: struct {
				BusStopID  string
				ServiceNos []string
				Err        error
			}{
				BusStopID:  "96049",
				ServiceNos: []string{},
			},
		},
		{
			Name: "Bus stop ID and services",
			Text: "96049 2 24",
			Expected: struct {
				BusStopID  string
				ServiceNos []string
				Err        error
			}{
				BusStopID:  "96049",
				ServiceNos: []string{"2", "24"},
			},
		},
		{
			Name: "Not comprised of 5 digits",
			Text: "hello",
			Expected: struct {
				BusStopID  string
				ServiceNos []string
				Err        error
			}{
				Err: errInvalidBusStopCode,
			},
		},
		{
			Name: "More than 5 digits",
			Text: "123456",
			Expected: struct {
				BusStopID  string
				ServiceNos []string
				Err        error
			}{
				Err: errInvalidBusStopCode,
			},
		},
		{
			Name: "doesn't start with bus stop code",
			Text: "hello 12345",
			Expected: struct {
				BusStopID  string
				ServiceNos []string
				Err        error
			}{
				Err: errInvalidBusStopCode,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			text := tc.Text
			busStopID, serviceNos, err := InferEtaQuery(text)
			if expected := tc.Expected.Err; expected != nil {
				assert.Equal(t, expected, err)
			} else {
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, tc.Expected.BusStopID, busStopID)
				assert.Equal(t, tc.Expected.ServiceNos, serviceNos)
			}
		})
	}
}

func TestCalculateEtas(t *testing.T) {
	now, _ := time.Parse(time.RFC3339, time.RFC3339)

	t.Run("In operation, all etas", func(t *testing.T) {
		etas := CalculateEtas(now, newArrival(now, ""))

		actual := etas.Services
		expected := []IncomingBuses{{ServiceNo: "2", Etas: [3]string{"-1", "10", "36"}, Loads: [3]string{"SDA", "SDA", "LSD"}, Features: [3]string{"", "", "WAB"}, Types: [3]string{"DD", "DD", "BD"}}, {ServiceNo: "24", Etas: [3]string{"1", "3", "6"}, Loads: [3]string{"SEA", "SDA", "LSD"}, Features: [3]string{"", "WAB", ""}, Types: [3]string{"SD", "DD", "BD"}}}

		if !reflect.DeepEqual(actual, expected) {
			fmt.Printf("Expected:\n%#v\nActual:\n%#v\n", expected, actual)
			t.Fail()
		}
	})
	t.Run("In operation, missing etas", func(t *testing.T) {
		arrival := newArrival(now, "")
		arrival.Services[0].NextBus2.EstimatedArrival = time.Time{}
		arrival.Services[0].NextBus3.EstimatedArrival = time.Time{}

		etas := CalculateEtas(now, arrival)

		actual := etas.Services
		expected := []IncomingBuses{{ServiceNo: "2", Etas: [3]string{"-1", "?", "?"}, Loads: [3]string{"SDA", "SDA", "LSD"}, Features: [3]string{"", "", "WAB"}, Types: [3]string{"DD", "DD", "BD"}}, {ServiceNo: "24", Etas: [3]string{"1", "3", "6"}, Loads: [3]string{"SEA", "SDA", "LSD"}, Features: [3]string{"", "WAB", ""}, Types: [3]string{"SD", "DD", "BD"}}}

		if !reflect.DeepEqual(actual, expected) {
			fmt.Printf("Expected:\n%#v\nActual:\n%#v\n", expected, actual)
			t.Fail()
		}
	})
	t.Run("In operation, no etas", func(t *testing.T) {
		arrival := newArrival(now, "")
		arrival.Services[0].NextBus.EstimatedArrival = time.Time{}
		arrival.Services[0].NextBus2.EstimatedArrival = time.Time{}
		arrival.Services[0].NextBus3.EstimatedArrival = time.Time{}

		etas := CalculateEtas(now, arrival)

		actual := etas.Services
		expected := []IncomingBuses{{ServiceNo: "2", Etas: [3]string{"?", "?", "?"}, Loads: [3]string{"SDA", "SDA", "LSD"}, Features: [3]string{"", "", "WAB"}, Types: [3]string{"DD", "DD", "BD"}}, {ServiceNo: "24", Etas: [3]string{"1", "3", "6"}, Loads: [3]string{"SEA", "SDA", "LSD"}, Features: [3]string{"", "WAB", ""}, Types: [3]string{"SD", "DD", "BD"}}}

		if !reflect.DeepEqual(actual, expected) {
			fmt.Printf("Expected:\n%#v\nActual:\n%#v\n", expected, actual)
			t.Fail()
		}
	})
}

func TestEtaTable(t *testing.T) {
	services := [][4]string{
		{"2", "-2", "12", "20"},
		{"24", "2", "9", "18"},
	}

	expected := "| Svc | Next |  2nd |  3rd |\n|-----|------|------|------|\n| 2   |   -2 |   12 |   20 |\n| 24  |    2 |    9 |   18 |"
	actual := EtaTable(services)

	if actual != expected {
		fmt.Printf("Expected:\n%q\nActual:\n%q\n", expected, actual)
		t.Fail()
	}
}

func TestSummaryETAFormatter_Format(t *testing.T) {
	svc2 := IncomingBuses{
		ServiceNo: "2",
		Etas:      [3]string{"-1", "10", "36"},
		Loads:     [3]string{"", "SDA", "LSD"},
		Features:  [3]string{"", "", "WAB"},
		Types:     [3]string{"", "DD", "BD"},
	}
	svc24 := IncomingBuses{
		ServiceNo: "24",
		Etas:      [3]string{"1", "3", "6"},
		Loads:     [3]string{"SEA", "SDA", "LSD"},
		Features:  [3]string{"", "WAB", ""},
		Types:     [3]string{"SD", "DD", "BD"},
	}
	type testCase struct {
		Name       string
		ETAs       BusEtas
		ServiceNos []string
		Expected   string
	}
	testCases := []testCase{
		{
			Name: "single eta",
			ETAs: BusEtas{
				Services: []IncomingBuses{
					svc2,
				},
			},
			Expected: "```\n| Svc | Next |  2nd |  3rd |\n|-----|------|------|------|\n| 2   |   -1 |   10 |   36 |```\nShowing 1 out of 1 service for this bus stop.",
		},
		{
			Name: "multiple etas",
			ETAs: BusEtas{
				Services: []IncomingBuses{
					svc2,
					svc24,
				},
			},
			Expected: "```\n| Svc | Next |  2nd |  3rd |\n|-----|------|------|------|\n| 2   |   -1 |   10 |   36 |\n| 24  |    1 |    3 |    6 |```\nShowing 2 out of 2 services for this bus stop.",
		},
		{
			Name: "filtered etas",
			ETAs: BusEtas{
				Services: []IncomingBuses{
					svc2,
					svc24,
				},
			},
			ServiceNos: []string{"2"},
			Expected:   "```\n| Svc | Next |  2nd |  3rd |\n|-----|------|------|------|\n| 2   |   -1 |   10 |   36 |```\nShowing 1 out of 2 services for this bus stop.",
		},
	}
	for _, tc := range testCases {
		formatter := SummaryETAFormatter{}
		t.Run(tc.Name, func(t *testing.T) {
			actual := formatter.Format(tc.ETAs, tc.ServiceNos)
			assert.Equal(t, tc.Expected, actual)
		})
	}

}

func TestNewRefreshButton(t *testing.T) {
	type testCase struct {
		Name        string
		BusStopCode string
		ServiceNos  []string
		Expected    telegram.InlineKeyboardButton
	}
	testCases := []testCase{
		{
			Name:        "with bus stop code only",
			BusStopCode: "96049",
			ServiceNos:  nil,
			Expected: telegram.InlineKeyboardButton{
				Text:         "Refresh",
				CallbackData: `{"t":"refresh","b":"96049"}`,
			},
		},
		{
			Name:        "with bus stop code and service nos",
			BusStopCode: "96049",
			ServiceNos:  []string{"2", "24"},
			Expected: telegram.InlineKeyboardButton{
				Text:                         "Refresh",
				CallbackData:                 `{"t":"refresh","b":"96049","s":["2","24"]}`,
				SwitchInlineQueryCurrentChat: (*string)(nil),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			actual := NewRefreshButton(tc.BusStopCode, tc.ServiceNos, "")
			assert.Equal(t, tc.Expected, actual)
		})
	}
}

func TestNewResendButton(t *testing.T) {
	type testCase struct {
		Name        string
		BusStopCode string
		ServiceNos  []string
		Expected    telegram.InlineKeyboardButton
	}
	testCases := []testCase{
		{
			Name:        "with bus stop code only",
			BusStopCode: "96049",
			ServiceNos:  nil,
			Expected: telegram.InlineKeyboardButton{
				Text:         "Resend",
				CallbackData: `{"t":"resend","b":"96049"}`,
			},
		},
		{
			Name:        "with bus stop code and service nos",
			BusStopCode: "96049",
			ServiceNos:  []string{"2", "24"},
			Expected: telegram.InlineKeyboardButton{
				Text:                         "Resend",
				CallbackData:                 `{"t":"resend","b":"96049","s":["2","24"]}`,
				SwitchInlineQueryCurrentChat: (*string)(nil),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			actual := NewResendButton(tc.BusStopCode, tc.ServiceNos, "")
			assert.Equal(t, tc.Expected, actual)
		})
	}
}

func TestNewToggleFavouriteButton(t *testing.T) {
	type testCase struct {
		Name        string
		BusStopCode string
		ServiceNos  []string
		Expected    telegram.InlineKeyboardButton
	}
	testCases := []testCase{
		{
			Name:        "with bus stop code only",
			BusStopCode: "96049",
			ServiceNos:  nil,
			Expected: telegram.InlineKeyboardButton{
				Text:         "⭐",
				CallbackData: `{"t":"togf","a":"96049"}`,
			},
		},
		{
			Name:        "with bus stop code and service nos",
			BusStopCode: "96049",
			ServiceNos:  []string{"2", "24"},
			Expected: telegram.InlineKeyboardButton{
				Text:                         "⭐",
				CallbackData:                 `{"t":"togf","a":"96049 2 24"}`,
				SwitchInlineQueryCurrentChat: (*string)(nil),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			actual := NewToggleFavouriteButton(tc.BusStopCode, tc.ServiceNos)
			assert.Equal(t, tc.Expected, actual)
		})
	}
}

func TestNewETAMessageReplyMarkup(t *testing.T) {
	type args struct {
		busStopCode string
		serviceNos  []string
		formatter   string
		inline      bool
	}
	tests := []struct {
		name string
		args args
		want telegram.InlineKeyboardMarkup
	}{
		{
			name: "when not inline",
			args: args{
				busStopCode: "96049",
				inline:      false,
			},
			want: telegram.InlineKeyboardMarkup{
				InlineKeyboard: [][]telegram.InlineKeyboardButton{
					{
						{
							Text:         "Refresh",
							CallbackData: `{"t":"refresh","b":"96049"}`,
						},
						{
							Text:         "Resend",
							CallbackData: `{"t":"resend","b":"96049"}`,
						},
						{
							Text:         "⭐",
							CallbackData: `{"t":"togf","a":"96049"}`,
						},
					},
					{
						{
							Text:         "Show incoming bus details",
							CallbackData: `{"t":"refresh","b":"96049","f":"f"}`,
						},
					},
				},
			},
		},
		{
			name: "when not inline, with formatter",
			args: args{
				busStopCode: "96049",
				formatter:   FormatterSummary,
				inline:      false,
			},
			want: telegram.InlineKeyboardMarkup{
				InlineKeyboard: [][]telegram.InlineKeyboardButton{
					{
						{
							Text:         "Refresh",
							CallbackData: `{"t":"refresh","b":"96049","f":"s"}`,
						},
						{
							Text:         "Resend",
							CallbackData: `{"t":"resend","b":"96049","f":"s"}`,
						},
						{
							Text:         "⭐",
							CallbackData: `{"t":"togf","a":"96049"}`,
						},
					},
					{
						{
							Text:         "Show incoming bus details",
							CallbackData: `{"t":"refresh","b":"96049","f":"f"}`,
						},
					},
				},
			},
		},
		{
			name: "when not inline, with features formatter",
			args: args{
				busStopCode: "96049",
				formatter:   FormatterFeatures,
				inline:      false,
			},
			want: telegram.InlineKeyboardMarkup{
				InlineKeyboard: [][]telegram.InlineKeyboardButton{
					{
						{
							Text:         "Refresh",
							CallbackData: `{"t":"refresh","b":"96049","f":"f"}`,
						},
						{
							Text:         "Resend",
							CallbackData: `{"t":"resend","b":"96049","f":"f"}`,
						},
						{
							Text:         "⭐",
							CallbackData: `{"t":"togf","a":"96049"}`,
						},
					},
					{
						{
							Text:         "Show incoming bus summary",
							CallbackData: `{"t":"refresh","b":"96049","f":"s"}`,
						},
					},
				},
			},
		},
		{
			name: "when inline",
			args: args{
				busStopCode: "96049",
				inline:      true,
			},
			want: telegram.InlineKeyboardMarkup{
				InlineKeyboard: [][]telegram.InlineKeyboardButton{
					{
						{
							Text:         "Refresh",
							CallbackData: `{"t":"refresh","b":"96049"}`,
						},
					},
					{
						{
							Text:         "Show incoming bus details",
							CallbackData: `{"t":"refresh","b":"96049","f":"f"}`,
						},
					},
				},
			},
		},
		{
			name: "when inline, with formatter",
			args: args{
				busStopCode: "96049",
				formatter:   FormatterFeatures,
				inline:      true,
			},
			want: telegram.InlineKeyboardMarkup{
				InlineKeyboard: [][]telegram.InlineKeyboardButton{
					{
						{
							Text:         "Refresh",
							CallbackData: `{"t":"refresh","b":"96049","f":"f"}`,
						},
					},
					{
						{
							Text:         "Show incoming bus summary",
							CallbackData: `{"t":"refresh","b":"96049","f":"s"}`,
						},
					},
				},
			},
		},
		{
			name: "when inline, with features formatter",
			args: args{
				busStopCode: "96049",
				formatter:   FormatterSummary,
				inline:      true,
			},
			want: telegram.InlineKeyboardMarkup{
				InlineKeyboard: [][]telegram.InlineKeyboardButton{
					{
						{
							Text:         "Refresh",
							CallbackData: `{"t":"refresh","b":"96049","f":"s"}`,
						},
					},
					{
						{
							Text:         "Show incoming bus details",
							CallbackData: `{"t":"refresh","b":"96049","f":"f"}`,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := NewETAMessageReplyMarkup(tt.args.busStopCode, tt.args.serviceNos, tt.args.formatter, tt.args.inline)
			assert.Equal(t, tt.want, actual)
		})
	}
}

type MockETAFormatter string

func (s MockETAFormatter) Format(etas BusEtas, services []string) string {
	return string(s)
}

func TestETAMessageText(t *testing.T) {
	stop := BusStop{
		BusStopCode: "96049",
		RoadName:    "Upp Changi Rd East",
		Description: "Opp Tropicana Condo",
	}
	type args struct {
		busStopRepository BusStopRepository
		etaService        ETAService
		formatter         ETAFormatter
		t                 time.Time
		code              string
		services          []string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "when bus stop is found and has arriving buses",
			args: args{
				busStopRepository: mockBusStopRepository{
					BusStop: &stop,
				},
				etaService: mockDatamall{},
				formatter:  MockETAFormatter("formatter output"),
				t:          time.Time{},
				code:       "",
				services:   nil,
			},
			want: "*Opp Tropicana Condo (96049)*\nUpp Changi Rd East\n" +
				"formatter output\n" +
				"\n_Last updated at 01 Jan 01 08:00 SGT_",
			wantErr: false,
		},
		{
			name: "when bus stop is found but has no arriving buses",
			args: args{
				busStopRepository: mockBusStopRepository{
					BusStop: &stop,
				},
				etaService: mockDatamall{},
				formatter:  MockETAFormatter("formatter output"),
				t:          time.Time{},
				code:       "",
			},
			want: "*Opp Tropicana Condo (96049)*\nUpp Changi Rd East\n" +
				"formatter output\n" +
				"\n_Last updated at 01 Jan 01 08:00 SGT_",
			wantErr: false,
		},
		{
			name: "when bus stop is found but datamall is down",
			args: args{
				busStopRepository: mockBusStopRepository{
					BusStop: &stop,
				},
				etaService: mockETAService{
					Error: &datamall.Error{StatusCode: 501},
				},
				formatter: MockETAFormatter("body"),
				t:         time.Time{},
				code:      "",
				services:  nil,
			},
			want: "*Opp Tropicana Condo (96049)*\nUpp Changi Rd East\n" +
				"\nOh no! The LTA DataMall API that Bus Eta Bot relies on appears to be down at the moment (it returned HTTP status code 501).\n" +
				"\n_Last updated at 01 Jan 01 08:00 SGT_",
			wantErr: false,
		},
		{
			name: "when bus stop is not found but has arriving buses",
			args: args{
				busStopRepository: mockBusStopRepository{},
				etaService:        mockDatamall{},
				formatter:         MockETAFormatter("formatter output"),
				t:                 time.Time{},
				code:              "96049",
				services:          nil,
			},
			want: "*96049*\n" +
				"formatter output\n" +
				"\n_Last updated at 01 Jan 01 08:00 SGT_",
			wantErr: false,
		},
		{
			name: "when bus stop is not found and has no arriving buses",
			args: args{
				busStopRepository: mockBusStopRepository{},
				etaService: mockETAService{
					BusArrival: datamall.BusArrival{},
				},
				formatter: MockETAFormatter("formatter output"),
				t:         time.Time{},
				code:      "96049",
				services:  nil,
			},
			want: "*96049*\n" +
				"\nNo etas found for this bus stop.\n" +
				"\n_Last updated at 01 Jan 01 08:00 SGT_",
		},
		{
			name: "when bus stop is not found and datamall is down",
			args: args{
				busStopRepository: mockBusStopRepository{},
				etaService: mockETAService{
					Error: &datamall.Error{StatusCode: 501},
				},
				formatter: MockETAFormatter("body"),
				t:         time.Time{},
				code:      "96049",
				services:  nil,
			},
			want: "*96049*\n" +
				"\nOh no! The LTA DataMall API that Bus Eta Bot relies on appears to be down at the moment (it returned HTTP status code 501).\n" +
				"\n_Last updated at 01 Jan 01 08:00 SGT_",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ETAMessageText(tt.args.busStopRepository, tt.args.etaService, tt.args.formatter, tt.args.t, tt.args.code, tt.args.services)
			if (err != nil) != tt.wantErr {
				t.Errorf("ETAMessageText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestNewETA(t *testing.T) {
	busStop := BusStop{
		BusStopCode: "96049",
		RoadName:    "Upp Changi Rd East",
		Description: "Opp Tropicana Condo",
	}
	t.Run("populates bus stop when bus stop is found", func(t *testing.T) {
		getter := mockBusStopRepository{
			BusStop: &busStop,
		}
		actual := NewETA(context.Background(), getter, mockDatamall{}, ETARequest{})
		assert.Equal(t, busStop, actual.BusStop)
	})
	t.Run("populates bus stop code when bus stop is not found", func(t *testing.T) {
		getter := mockBusStopRepository{}
		code := "96049"
		actual := NewETA(context.Background(), getter, mockDatamall{}, ETARequest{
			Code: code,
		})
		expected := BusStop{
			BusStopCode: code,
		}
		assert.Equal(t, expected, actual.BusStop)
	})
	t.Run("contains appropriate error when datamall is down", func(t *testing.T) {
		getter := mockBusStopRepository{}
		etaService := mockETAService{
			Error: &datamall.Error{
				StatusCode: 503,
			},
		}
		actual := NewETA(context.Background(), getter, etaService, ETARequest{})
		expected := "LTA DataMall could be down at the moment (status code 503)"
		assert.Equal(t, expected, actual.Error)
	})
	t.Run("contains appropriate error when something else went wrong", func(t *testing.T) {
		getter := mockBusStopRepository{}
		etaService := mockETAService{
			Error: errors.New("some other error"),
		}
		actual := NewETA(context.Background(), getter, etaService, ETARequest{})
		expected := "An error occurred while fetching ETAs (request ID: )"
		assert.Equal(t, expected, actual.Error)
	})
	t.Run("doesn't filter services when services is empty", func(t *testing.T) {
		getter := mockBusStopRepository{}
		arrival := newArrival(time.Time{}, "")
		etaService := mockETAService{
			BusArrival: arrival,
		}
		actual := NewETA(context.Background(), getter, etaService, ETARequest{})
		assert.Equal(t, arrival.Services, actual.Services)
	})
	t.Run("doesn't filter services when services is empty", func(t *testing.T) {
		getter := mockBusStopRepository{}
		arrival := newArrival(time.Time{}, "")
		etaService := mockETAService{
			BusArrival: arrival,
		}
		request := ETARequest{
			Services: []string{"2"},
		}
		actual := NewETA(context.Background(), getter, etaService, request)
		expected := []datamall.Service{
			arrival.Services[0],
		}
		assert.Equal(t, expected, actual.Services)
	})
}

type mockFormatter struct{}

func (f mockFormatter) GetFormatter(ctx context.Context, userID int) Formatter {
	return f
}

func (f mockFormatter) Format(eta ETA) (string, error) {
	var output string
	for _, svc := range eta.Services {
		output += " " + svc.ServiceNo
	}
	return "services:" + output, nil
}

func TestETAMessageFactory_Text(t *testing.T) {
	f := ETAMessageFactory{
		busStopGetter:    mockBusStopRepository{},
		etaService:       mockDatamall{},
		formatterFactory: mockFormatter{},
	}
	actual, err := f.Text(context.Background(), ETARequest{})
	if err != nil {
		t.Fatal(err)
	}
	expected := "services: 2 24"
	assert.Equal(t, expected, actual)
}
