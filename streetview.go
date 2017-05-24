package main

import (
	"fmt"
	"net/http"
	"net/url"
)

// StreetViewEndpoint is the endpoint for the Street View Image API.
const StreetViewEndpoint = "https://maps.googleapis.com/maps/api/streetview"

// StreetViewAPI contains information for making a request to the Street View Image API.
type StreetViewAPI struct {
	Endpoint string
	APIKey   string
	Client   *http.Client
}

func (s StreetViewAPI) getPhotoURL(params url.Values) (string, error) {
	u, err := url.Parse(s.Endpoint)
	if err != nil {
		return "", err
	}
	u.RawQuery = params.Encode()

	return fmt.Sprint(u), nil
}

// GetPhotoURLByLocation returns a street view image url based on a latitude and longitude.
func (s StreetViewAPI) GetPhotoURLByLocation(lat, lon float64, width, height int) (string, error) {
	params := url.Values{}
	params.Set("key", s.APIKey)
	params.Set("location", fmt.Sprintf("%f,%f", lat, lon))
	params.Set("size", fmt.Sprintf("%dx%d", width, height))

	return s.getPhotoURL(params)
}

// GetPhotoURLByAddress returns a street view image url based on an address string.
func (s StreetViewAPI) GetPhotoURLByAddress(addr string, width, height int) (string, error) {
	params := url.Values{}
	params.Set("key", s.APIKey)
	params.Set("location", addr)
	params.Set("size", fmt.Sprintf("%dx%d", width, height))

	return s.getPhotoURL(params)
}

// NewStreetViewAPIWithClient returns a new StreetViewAPI with the provided api key and http.Client.
func NewStreetViewAPIWithClient(key string, client *http.Client) StreetViewAPI {
	return StreetViewAPI{
		Endpoint: StreetViewEndpoint,
		APIKey:   key,
		Client:   client,
	}
}
