package tracing

import (
	"context"
	"fmt"
	"time"

	"go.opencensus.io/trace"
	"go.opencensus.io/trace/tracestate"
)

type testExporter struct {
	spans []*trace.SpanData
}

func (t *testExporter) ExportSpan(s *trace.SpanData) {
	t.spans = append(t.spans, s)
}

var (
	tid               = trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 4, 8, 16, 32, 64, 128}
	sid               = trace.SpanID{1, 2, 4, 8, 16, 32, 64, 128}
	testTracestate, _ = tracestate.New(nil, tracestate.Entry{Key: "foo", Value: "bar"})
)

// startSpan returns a context with a new Span that is recording events and will be exported.
func startSpan(o trace.StartOptions) *trace.Span {
	_, span := trace.StartSpanWithRemoteParent(context.Background(), "span0",
		trace.SpanContext{
			TraceID:      tid,
			SpanID:       sid,
			TraceOptions: 1,
		},
		trace.WithSampler(o.Sampler),
		trace.WithSpanKind(o.SpanKind),
	)
	return span
}

// endSpan ends the Span in the context and returns the exported SpanData.
//
// It also does some tests on the Span, and tests and clears some fields in the SpanData.
func endSpan(span *trace.Span) (*trace.SpanData, error) {

	if !span.IsRecordingEvents() {
		return nil, fmt.Errorf("IsRecordingEvents: got false, want true")
	}
	if !span.SpanContext().IsSampled() {
		return nil, fmt.Errorf("IsSampled: got false, want true")
	}
	var te testExporter
	trace.RegisterExporter(&te)
	span.End()
	trace.UnregisterExporter(&te)
	fmt.Printf("%#v\n", te.spans)
	if len(te.spans) != 1 {
		return nil, fmt.Errorf("got exported spans %#v, want one span", te.spans)
	}
	got := te.spans[0]
	if got.SpanContext.SpanID == (trace.SpanID{}) {
		return nil, fmt.Errorf("exporting span: expected nonzero SpanID")
	}
	got.SpanContext.SpanID = trace.SpanID{}
	if !checkTime(&got.StartTime) {
		return nil, fmt.Errorf("exporting span: expected nonzero StartTime")
	}
	if !checkTime(&got.EndTime) {
		return nil, fmt.Errorf("exporting span: expected nonzero EndTime")
	}
	return got, nil
}

// checkTime checks that a nonzero time was set in x, then clears it.
func checkTime(x *time.Time) bool {
	if x.IsZero() {
		return false
	}
	*x = time.Time{}
	return true
}
