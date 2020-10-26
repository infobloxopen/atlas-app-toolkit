package tracing

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
)

func TestSpanContextAnnotator_FromContext(t *testing.T) {
	ctx, span := trace.StartSpan(context.Background(), "")
	result := SpanContextAnnotator(ctx, nil)
	assert.Equal(t, []string{string(propagation.Binary(span.SpanContext()))}, result[traceContextKey])
}

func TestSpanContextAnnotator_FromRequest(t *testing.T) {
	_, span := trace.StartSpan(context.Background(), "")
	req, _ := http.NewRequest("", "", nil)
	defaultFormat.SpanContextToRequest(span.SpanContext(), req)
	sc, ok := defaultFormat.SpanContextFromRequest(req)
	assert.True(t, ok)

	result := SpanContextAnnotator(context.Background(), req)
	assert.Equal(t, []string{string(propagation.Binary(sc))}, result[traceContextKey])
}
