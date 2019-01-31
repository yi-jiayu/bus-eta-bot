package busetabot

import (
	"context"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
)

type requestKey struct{}

func NewContext(r *http.Request) context.Context {
	ctx := appengine.NewContext(r)
	// Add the request onto the context too
	ctx = context.WithValue(ctx, requestKey{}, r)
	return ctx
}

// NewContextWithOptions starts an instance of the development API server
// with the given options, and returns a context that will route all
// API calls to that server, as well as a closure that must be called
// when the Context is no longer required.
// If opts is nil the default values are used.
func NewContextWithOptions(opts *aetest.Options) (context.Context, func(), error) {
	inst, err := aetest.NewInstance(opts)
	if err != nil {
		return nil, nil, err
	}
	req, err := inst.NewRequest("GET", "/", nil)
	if err != nil {
		inst.Close()
		return nil, nil, err
	}
	ctx := appengine.NewContext(req)
	return ctx, func() {
		inst.Close()
	}, nil
}
