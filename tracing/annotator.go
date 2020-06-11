package tracing

import (
	"context"
	"encoding/hex"
	"net/http"

	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
	"google.golang.org/grpc/metadata"
)

const (
	traceContextKey = "grpc-trace-bin"

	//ParentSpanIDKey is a header name for ParrentSpanId B3 header
	ParentSpanIDKey = "X-B3-ParentSpanId"

	//SpanIDKey is a header name for SpanId B3 header
	SpanIDKey = "X-B3-SpanId"

	//TraceIDKey is a header name for TraceId B3 header
	TraceIDKey = "X-B3-TraceId"

	//TraceSampledKey is a header name for Sampled B3 header
	TraceSampledKey = "X-B3-Sampled"

	//TODO: handle ParentSpanIDKey and TraceSampledKey as well
	//TODO: check headers names
)

//SpanContextAnnotator retrieve information about current span from context or HTTP headers
//and propogate in binary format to gRPC service
func SpanContextAnnotator(ctx context.Context, req *http.Request) metadata.MD {
	md := make(metadata.MD)

	span := trace.FromContext(ctx)
	if span != nil {
		traceContextBinary := propagation.Binary(span.SpanContext())
		md[traceContextKey] = []string{string(traceContextBinary)}

		return md
	}

	//Fallback to assembling trace.SpanContext from headers
	var sc trace.SpanContext

	spanID := req.Header.Get(SpanIDKey)
	if spanID == "" {
		return md
	}

	buf, err := hex.DecodeString(spanID)
	if err != nil {
		return md
	}

	copy(sc.SpanID[:], buf)

	traceID := req.Header.Get(TraceIDKey)
	if traceID == "" {
		return md
	}

	buf, err = hex.DecodeString(traceID)
	if err != nil {
		return md
	}

	copy(sc.TraceID[:], buf)

	//TODO: check how to use TraceSampledKey and logTraceParentSpanID

	traceContextBinary := propagation.Binary(sc)
	md[traceContextKey] = []string{string(traceContextBinary)}

	return md
}
