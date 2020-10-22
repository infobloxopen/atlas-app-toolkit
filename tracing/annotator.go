package tracing

import (
	"context"
	"net/http"

	"go.opencensus.io/plugin/ochttp/propagation/b3"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
	"google.golang.org/grpc/metadata"
)

const (
	traceContextKey = "grpc-trace-bin"
)

var defaultFormat propagation.HTTPFormat = &b3.HTTPFormat{}

//SpanContextAnnotator retrieve information about current span from context or HTTP headers
//and propagate in binary format to gRPC service
func SpanContextAnnotator(ctx context.Context, req *http.Request) metadata.MD {
	md := make(metadata.MD)

	span := trace.FromContext(ctx)
	if span != nil {
		traceContextBinary := propagation.Binary(span.SpanContext())
		md[traceContextKey] = []string{string(traceContextBinary)}

		return md
	}

	sc, ok := defaultFormat.SpanContextFromRequest(req)
	if ok {
		traceContextBinary := propagation.Binary(sc)
		md[traceContextKey] = []string{string(traceContextBinary)}
	}

	return md
}
