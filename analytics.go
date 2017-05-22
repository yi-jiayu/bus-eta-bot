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
	ProtocolVersion = "1"
	Endpoint        = "https://www.google-analytics.com/collect"
	DebugEndpoint   = "https://www.google-analytics.com/debug/collect"
)

// GAClient contains the endpoint, tracking id and http client to use to send hits to the Measurement Protocol
type GAClient struct {
	Endpoint   string
	TrackingID string
	Client     *http.Client
}

// User contains a client id or a user id
type User struct {
	ClientID string
	UserID   string
}

// Event represents an event hit type
type Event struct {
	Category string  `json:"ec"`
	Action   string  `json:"ea"`
	Label    *string `json:"el"`
	Value    *int    `json:"ev"`
}

// App contains information for app tracking
type App struct {
	Name        string
	ID          string
	Version     string
	InstallerID string
}

// Hit represents any object which can be serialised into the Measurement Protocol
type Hit interface {
	Values() map[string]string
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

// Values serialises an Event's properties
func (e Event) Values() map[string]string {
	values := map[string]string{
		"t":  "event",
		"ec": e.Category,
		"ea": e.Action,
	}

	if e.Label != nil {
		values["el"] = *e.Label
	}

	if e.Value != nil {
		values["ev"] = fmt.Sprintf("%d", *e.Value)
	}

	return values
}

// Values serialises an App's properties
func (a App) Values() map[string]string {
	values := map[string]string{
		"an": a.Name,
	}

	if a.ID != "" {
		values["aid"] = a.ID
	}

	if a.Version != "" {
		values["av"] = a.Version
	}

	if a.InstallerID != "" {
		values["aiid"] = a.InstallerID
	}

	return values
}

// NewDefaultClient returns a GAClient which uses http.DefaultClient
func NewDefaultClient(tid string) GAClient {
	return GAClient{
		Endpoint:   Endpoint,
		TrackingID: tid,
		Client:     http.DefaultClient,
	}
}

// NewClient constructs a GAClient with the provided tid and http.Client
func NewClient(tid string, client *http.Client) GAClient {
	return GAClient{
		Endpoint:   Endpoint,
		TrackingID: tid,
		Client:     client,
	}
}

func (c GAClient) do(endpoint string, user User, hits ...Hit) (*http.Response, error) {
	values := url.Values{}
	values.Set("v", ProtocolVersion)
	values.Set("tid", c.TrackingID)

	if user.ClientID != "" {
		values.Set("cid", user.ClientID)
	} else {
		values.Set("uid", user.UserID)
	}

	for _, hit := range hits {
		for k, v := range hit.Values() {
			values.Set(k, v)
		}
	}

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode != http.StatusOK {
		return resp, errors.New("status code error")
	}

	return resp, nil
}

// Send sends a hit to the measurement protocol
func (c GAClient) Send(user User, hits ...Hit) (*http.Response, error) {
	return c.do(c.Endpoint, user, hits...)
}

// Test sends a hit to the measurement protocol validation server
func (c GAClient) Test(user User, hits ...Hit) (*http.Response, error) {
	return c.do(DebugEndpoint, user, hits...)
}
