package busetabot

import (
	"context"
	"net/http"

	"github.com/getsentry/raven-go"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

type requestKey struct{}
type sentryKey struct{}

func NewContext(r *http.Request) (ctx context.Context) {
	// create an appengine context
	ctx = appengine.NewContext(r)

	// add the request onto the context too
	ctx = context.WithValue(ctx, requestKey{}, r)

	// include a raven client on the context
	sentry, err := raven.New("")
	if err != nil {
		log.Warningf(ctx, "error creating raven client: %+v", err)
		return
	}
	ctx = context.WithValue(ctx, sentryKey{}, sentry)
	return ctx
}

func setUserContext(ctx context.Context, ID string) {
	if sentry, ok := ctx.Value(sentryKey{}).(*raven.Client); ok {
		sentry.SetUserContext(&raven.User{
			ID: ID,
		})
	}
}

func logError(ctx context.Context, err error) {
	if appengine.IsAppEngine() || appengine.IsDevAppServer() {
		log.Errorf(ctx, "%+v", err)
	}
	if sentry, ok := ctx.Value(sentryKey{}).(*raven.Client); ok {
		sentry.CaptureError(err, nil)
	}
}
