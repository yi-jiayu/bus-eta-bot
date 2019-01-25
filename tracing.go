package busetabot

import (
	"context"
	"net/http"

	"go.opencensus.io/exporter/stackdriver/propagation"
	"go.opencensus.io/trace"
)

var HTTPFormat = propagation.HTTPFormat{}

func parentSpanFromContext(ctx context.Context) (spanContext trace.SpanContext, ok bool) {
	var r *http.Request
	r, ok = ctx.Value(requestKey{}).(*http.Request)
	if !ok {
		return
	}
	spanContext, ok = HTTPFormat.SpanContextFromRequest(r)
	return
}
