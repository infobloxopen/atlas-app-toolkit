package tracing

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/trace"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/stats"
)

var sensitiveMetadata = map[string]struct{}{
	"grpcgateway-authorization": struct{}{},
	"authorization":             struct{}{},
}

//GRPCOption allows extending handler with additional functionality
type GRPCOption func(*gRPCOptions)

type metadataMatcher func(string) (string, bool)

type gRPCOptions struct {
	spanWithMetadata func(context.Context, stats.RPCStats) bool
	metadataMatcher  metadataMatcher

	spanWithPayload func(context.Context, stats.RPCStats) bool
	maxPayloadSize  int
}

func defaultGRPCOptions() *gRPCOptions {
	return &gRPCOptions{
		metadataMatcher: defaultMetadataMatcher,
		maxPayloadSize:  DefaultMaxPayloadSize,

		//Keep spanWithMetadata and spanWithPayload equals to nil instead of dummy functions
		//to prevent path trough header for each request
	}
}

//WithMetadataAnnotation annotate span with request metadata
func WithMetadataAnnotation(f func(context.Context, stats.RPCStats) bool) GRPCOption {
	return func(ops *gRPCOptions) {
		ops.spanWithMetadata = f
	}
}

//WithMetadataMatcher set metadata matcher to filterout or preprocess metadata
func WithMetadataMatcher(f func(string) (string, bool)) GRPCOption {
	return func(ops *gRPCOptions) {
		ops.metadataMatcher = f
	}
}

//WithGRPCPayloadAnnotation add Inbound/Outbound payload as an attribute to span if f returns true
func WithGRPCPayloadAnnotation(f func(context.Context, stats.RPCStats) bool) GRPCOption {
	return func(ops *gRPCOptions) {
		ops.spanWithPayload = f
	}
}

//WithGRPCPayloadLimit limit payload size propogated to span
//in case payload exceeds limit, payload truncated and
//annotation payload.truncated=true added into span
func WithGRPCPayloadLimit(limit int) GRPCOption {
	return func(ops *gRPCOptions) {
		ops.maxPayloadSize = limit
	}
}

//Check that &ServerHandler{} comply stats.Handler interface
var _ stats.Handler = &ServerHandler{}

//ServerHandler is a wrapper over ocgrpc.ServerHandler
//wrapper extends metadata added into the  span
type ServerHandler struct {
	ocgrpc.ServerHandler

	options *gRPCOptions
}

//NewServerHandler returns wrapper over ocgrpc.ServerHandler
func NewServerHandler(ops ...GRPCOption) *ServerHandler {
	options := defaultGRPCOptions()
	for _, op := range ops {
		op(options)
	}

	return &ServerHandler{options: options}
}

// HandleRPC implements per-RPC tracing and stats instrumentation.
func (s *ServerHandler) HandleRPC(ctx context.Context, rs stats.RPCStats) {
	withHeaders := s.options.spanWithMetadata != nil && s.options.spanWithMetadata(ctx, rs)
	withPayload := s.options.spanWithPayload != nil && s.options.spanWithPayload(ctx, rs)

	span := trace.FromContext(ctx)

	if withPayload {
		switch rs := rs.(type) {
		case *stats.End:
			if rs.Error != nil {
				attrs := []trace.Attribute{trace.StringAttribute(ResponseErrorKey, rs.Error.Error())}
				span.Annotate(attrs, "Response error")
			}
		}
	}

	s.ServerHandler.HandleRPC(ctx, rs)

	span = trace.FromContext(ctx)

	if withHeaders {
		switch rs := rs.(type) {
		case *stats.InHeader:
			attrs := metadataToAttributes(rs.Header, RequestHeaderAnnotationPrefix, s.options.metadataMatcher)
			span.AddAttributes(attrs...)
		case *stats.InTrailer:
			attrs := metadataToAttributes(rs.Trailer, RequestTrailerAnnotationPrefix, s.options.metadataMatcher)
			span.AddAttributes(attrs...)
		case *stats.OutHeader:
			attrs := metadataToAttributes(rs.Header, ResponseHeaderAnnotationPrefix, s.options.metadataMatcher)
			span.AddAttributes(attrs...)
		case *stats.OutTrailer:
			attrs := metadataToAttributes(rs.Trailer, ResponseTrailerAnnotationPrefix, s.options.metadataMatcher)
			span.AddAttributes(attrs...)
		}
	}

	if withPayload {
		switch rs := rs.(type) {
		case *stats.InPayload:
			attrs, truncated, err := payloadToAttributes(RequestPayloadAnnotationKey, rs.Payload, s.options.maxPayloadSize)
			if err != nil {
				ctxlogrus.Extract(ctx).Errorln("unable to marshal response, err - ", err)
				return
			}

			if truncated {
				markSpanTruncated(span)
			}

			span.Annotate(attrs, "Request payload")
		case *stats.OutPayload:
			attrs, truncated, err := payloadToAttributes(ResponsePayloadAnnotationKey, rs.Payload, s.options.maxPayloadSize)
			if err != nil {
				ctxlogrus.Extract(ctx).Errorln("unable to marshal response, err - ", err)
				return
			}

			if truncated {
				markSpanTruncated(span)
			}

			span.Annotate(attrs, "Response payload")
		}
	}

}

func metadataToAttributes(md metadata.MD, prefix string, matcher metadataMatcher) []trace.Attribute {
	attrs := make([]trace.Attribute, 0, len(md))
	for k, vals := range md {
		key, ok := matcher(k)
		if !ok {
			continue
		}

		key, valsStr := fmt.Sprint(prefix, key), strings.Join(vals, ", ")
		if _, ok := sensitiveMetadata[k]; ok {
			valsStr = obfuscate(valsStr)
		}

		//Check that key and value is a valid utf-8 encoded strings, in case they not span will be omitted from span
		if !utf8.ValidString(key) || !utf8.ValidString(valsStr) {
			continue
		}

		attrs = append(attrs, trace.StringAttribute(key, valsStr))
	}

	return attrs
}

func payloadToAttributes(key string, value interface{}, limit int) ([]trace.Attribute, bool, error) {
	bytes, err := json.Marshal(value)
	if err != nil {
		return nil, false, err
	}

	bytes, truncated := truncatePayload(bytes, limit)
	attrs := []trace.Attribute{trace.StringAttribute(key, string(bytes))}
	return attrs, truncated, nil
}

//defaultHeaderMatcher is a header matcher which just accept all headers
func defaultMetadataMatcher(h string) (string, bool) {
	return h, true
}

//AlwaysGRPC for each call returns true
func AlwaysGRPC(_ context.Context, _ stats.RPCStats) bool {
	return true
}
