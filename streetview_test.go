package main

import (
	"fmt"
	"testing"
)

func TestStreetViewAPI_GetPhotoUrlByLocation(t *testing.T) {
	sv := NewStreetViewAPI("API_KEY")

	lat, lon := 1.34041450268626, 103.96127892061004
	height, width := 100, 100

	actual, err := sv.GetPhotoURLByLocation(lat, lon, height, width)
	if err != nil {
		t.Fatal(err)
	}

	expected := "https://maps.googleapis.com/maps/api/streetview?key=API_KEY&location=1.340415%2C103.961279&size=100x100"

	if actual != expected {
		fmt.Printf("Expected: %s\nActual:   %s\n", expected, actual)
		t.Fail()
	}
}
