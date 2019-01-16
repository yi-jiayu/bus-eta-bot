package busetabot

import (
	"context"
	"net/http"

	"google.golang.org/appengine"
)

type requestKey struct{}

func NewContext(r *http.Request) context.Context {
	ctx := appengine.NewContext(r)
	// Add the request onto the context too
	ctx = context.WithValue(ctx, requestKey{}, r)
	return ctx
}
