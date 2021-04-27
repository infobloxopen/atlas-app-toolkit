package tracing

import (
	"context"
	"errors"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"go.opencensus.io/trace"
	"google.golang.org/grpc/status"
)

var (
	//ErrSpanNotFound is an error which signal that there is no currently active spans
	ErrSpanNotFound = errors.New("there is no currently active spans")
)

//CurrentSpan returns current span
func CurrentSpan(ctx context.Context) (*trace.Span, error) {
	span := trace.FromContext(ctx)
	if span == nil {
		return nil, ErrSpanNotFound
	}

	return span, nil
}

//StartSpan starts span with name
func StartSpan(ctx context.Context, name string) (context.Context, *trace.Span) {
	return trace.StartSpan(ctx, name)
}

//TagSpan tags span
func TagSpan(span *trace.Span, attrs ...trace.Attribute) {
	span.AddAttributes(attrs...)
}

//TagCurrentSpan get current span from context and tag it
func TagCurrentSpan(ctx context.Context, attrs ...trace.Attribute) error {
	span, err := CurrentSpan(ctx)
	if err != nil {
		return err
	}

	TagSpan(span, attrs...)
	return nil
}

//AddMessageSpan adds message into span
func AddMessageSpan(span *trace.Span, message string, attrs ...trace.Attribute) {
	span.Annotate(attrs, message)
}

//AddMessageCurrentSpan get current span from context and adds message into it
func AddMessageCurrentSpan(ctx context.Context, message string, attrs ...trace.Attribute) error {
	span, err := CurrentSpan(ctx)
	if err != nil {
		return err
	}

	AddMessageSpan(span, message, attrs...)
	return nil
}

//AddErrorCurrentSpan get current span from context and adds error into it
func AddErrorCurrentSpan(ctx context.Context, err error) error {
	if err != nil {
		return nil
	}

	span, errSpan := CurrentSpan(ctx)
	if err != nil {
		ctxlogrus.Extract(ctx).Error("unable to get span from context, err: ", errSpan)
		return err
	}

	return AddErrorSpan(span, err)
}

//AddErrorSpan adds error into span
func AddErrorSpan(span *trace.Span, err error) error {
	var code int32 = trace.StatusCodeUnknown
	status, ok := status.FromError(err)
	if ok && status != nil {
		code = int32(status.Code())
	}

	span.SetStatus(trace.Status{
		Code:    code,
		Message: err.Error(),
	})

	return err
}
