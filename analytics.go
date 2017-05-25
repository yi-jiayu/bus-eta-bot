package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Measurement Protocol constants
const (
	MeasurementProtocolVersion            = "1"
	MeasurementProtocolEndpoint           = "https://www.google-analytics.com/collect"
	MeasurementProtocolValidationEndpoint = "https://www.google-analytics.com/debug/collect"
)

// Event categories
const (
	CategoryCommand     = "command"
	CategoryMessage     = "message"
	CategoryInlineQuery = "inline_query"
	CategoryCallback    = "callback_query"
	CategoryError       = "error"
)

// Event actions
const (
	ActionEtaCommandWithArgs    = "eta_command_with_args"
	ActionEtaCommandWithoutArgs = "eta_command_without_args"
	ActionStartCommand          = "start_command"
	ActionAboutCommand          = "about_command"
	ActionVersionCommand        = "version_command"
	ActionHelpCommand           = "help_command"
	ActionPrivacyCommand        = "privacy_command"
	ActionFeedbackCommand       = "feedback_command"

	ActionEtaTextMessage       = "eta_text_message"
	ActionContinuedTextMessage = "continued_text_message"
	ActionIgnoredTextMessage   = "ignored_text_message"
	ActionLocationMessage      = "location_message"

	ActionNewInlineQuery     = "new_inline_query"
	ActionOffsetInlineQuery  = "offset_inline_query"
	ActionChosenInlineResult = "chosen_inline_result"

	ActionRefreshCallback         = "refresh_callback"
	ActionResendCallback          = "resend_callback"
	ActionEtaCallback             = "eta_callback"
	ActionEtaDemoCallback         = "eta_demo_callback"
	ActionEtaFromLocationCallback = "eta_from_location_callback"

	ActionCommandError     = "command_error"
	ActionMessageError     = "message_error"
	ActionInlineQueryError = "inline_query_error"
	ActionCallbackError    = "callback_error"
)

// Event labels
const (
	LabelInlineMessage = "inline_message"
)

var errStatusCode = errors.New("status code error")

// Application details
var (
	ApplicationName    = "Bus Eta Bot"
	ApplicationID      = "github.com/yi-jiayu/bus-eta-bot-3"
	ApplicationVersion = Version
)

// GAClient contains the endpoint, tracking id and http client to use to send hits to the Measurement Protocol
type GAClient struct {
	Endpoint   string
	TrackingID string
	Client     *http.Client
}

// ValidationServerResponse is a response from the Measurement Protocol Validation Server
type ValidationServerResponse struct {
	Results []HitParsingResult `json:"hitParsingResult"`
}

// HitParsingResult contains whether the a hit is valid and additional information fom the validation server
type HitParsingResult struct {
	Valid          bool            `json:"valid"`
	Hit            string          `json:"hit"`
	ParserMessages []ParserMessage `json:"parserMessage"`
}

// ParserMessage  is a message from the validation server parser.
type ParserMessage struct {
	Type        string `json:"messageType"`
	Description string `json:"description"`
	Parameter   string `json:"parameter"`
}

// NewDefaultClient returns a GAClient which uses http.DefaultClient
func NewDefaultClient(tid string) GAClient {
	return GAClient{
		Endpoint:   MeasurementProtocolEndpoint,
		TrackingID: tid,
		Client:     http.DefaultClient,
	}
}

// NewClient constructs a GAClient with the provided tid and http.Client
func NewClient(tid string, client *http.Client) GAClient {
	return GAClient{
		Endpoint:   MeasurementProtocolEndpoint,
		TrackingID: tid,
		Client:     client,
	}
}

// LogEvent logs an event to the Measurement Protocol.
func (c GAClient) LogEvent(userID int, languageCode, category, action, label string) (*http.Response, error) {
	values := url.Values{}

	// protocol version
	values.Set("v", MeasurementProtocolVersion)

	// tid
	values.Set("tid", c.TrackingID)

	// user ID and language code
	values.Set("uid", fmt.Sprintf("%d", userID))
	if languageCode != "" {
		values.Set("ul", languageCode)
	}

	// hit type
	values.Set("t", "event")

	// application details
	values.Set("an", ApplicationName)
	values.Set("aid", ApplicationID)
	values.Set("av", ApplicationVersion)

	// event details
	values.Set("ec", category)
	values.Set("ea", action)
	values.Set("el", label)

	req, err := http.NewRequest("POST", c.Endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp, errStatusCode
	}

	return resp, nil
}
