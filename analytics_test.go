package main

import (
	"encoding/json"
	"os"
	"testing"
)

var tid = os.Getenv("GA_TID")

func TestGAClient_Test(t *testing.T) {
	t.Parallel()

	if tid == "" {
		return
	}

	client := NewDefaultClient(tid)

	event := Event{
		Category: "command",
		Action:   "start",
	}

	user := User{
		UserID: "as8eknlll",
	}

	app := App{
		Name:    "My App",
		ID:      "com.platform.vending",
		Version: "1.2",
	}

	resp, err := client.Test(user, app, event)
	if err != nil {
		t.Fatal(err)
	}

	var response ValidationServerResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Fatal(err)
	}

	if !response.Results[0].Valid {
		t.Fatalf("%+v", response)
	}
}
