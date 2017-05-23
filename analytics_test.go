package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
)

var tid = os.Getenv("GA_TID")

func TestGAClient_LogEvent(t *testing.T) {
	if tid == "" {
		return
	}

	gaClient := GAClient{
		Endpoint:   MeasurementProtocolValidationEndpoint,
		TrackingID: tid,
		Client:     http.DefaultClient,
	}

	resp, err := gaClient.LogEvent(1, "en-US", CategoryMessage, ActionCallbackError, "")
	if err != nil {
		t.Fatal(err)
	}

	var vsr ValidationServerResponse
	err = json.NewDecoder(resp.Body).Decode(&vsr)
	if err != nil {
		t.Fatal(err)
	}

	if !vsr.Results[0].Valid {
		fmt.Printf("%+v", vsr)
		t.Fail()
	}
}
