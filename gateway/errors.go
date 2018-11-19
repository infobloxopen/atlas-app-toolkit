package gateway

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/infobloxopen/atlas-app-toolkit/rpc/errdetails"
	"github.com/infobloxopen/atlas-app-toolkit/rpc/errfields"
)

// ProtoStreamErrorHandlerFunc handles the error as a gRPC error generated via status package and replies to the testRequest.
// Addition bool argument indicates whether method (http.ResponseWriter.WriteHeader) was called or not.
type ProtoStreamErrorHandlerFunc func(context.Context, bool, *runtime.ServeMux, runtime.Marshaler, http.ResponseWriter, *http.Request, error)

type RestResp struct {
	Error   []map[string]interface{} `json:"error,omitempty"`
	Success []map[string]interface{} `json:"success,omitempty"`
}

var (
	// ProtoMessageErrorHandler uses PrefixOutgoingHeaderMatcher.
	// To use ProtoErrorHandler with custom outgoing header matcher call NewProtoMessageErrorHandler.
	ProtoMessageErrorHandler = NewProtoMessageErrorHandler(PrefixOutgoingHeaderMatcher)
	// ProtoStreamErrorHandler uses PrefixOutgoingHeaderMatcher.
	// To use ProtoErrorHandler with custom outgoing header matcher call NewProtoStreamErrorHandler.
	ProtoStreamErrorHandler = NewProtoStreamErrorHandler(PrefixOutgoingHeaderMatcher)
)

// NewProtoMessageErrorHandler returns runtime.ProtoErrorHandlerFunc
func NewProtoMessageErrorHandler(out runtime.HeaderMatcherFunc) runtime.ProtoErrorHandlerFunc {
	h := &ProtoErrorHandler{out}
	return h.MessageHandler
}

// NewProtoStreamErrorHandler returns ProtoStreamErrorHandlerFunc
func NewProtoStreamErrorHandler(out runtime.HeaderMatcherFunc) ProtoStreamErrorHandlerFunc {
	h := &ProtoErrorHandler{out}
	return h.StreamHandler
}

// ProtoErrorHandler implements runtime.ProtoErrorHandlerFunc in method MessageHandler
// and ProtoStreamErrorHandlerFunc in method StreamHandler
// in accordance with REST API Syntax Specification.
// See RestError for the JSON format of an error
type ProtoErrorHandler struct {
	OutgoingHeaderMatcher runtime.HeaderMatcherFunc
}

// MessageHandler implements runtime.ProtoErrorHandlerFunc
// in accordance with REST API Syntax Specification.
// See RestError for the JSON format of an error
func (h *ProtoErrorHandler) MessageHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, rw http.ResponseWriter, req *http.Request, err error) {

	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		grpclog.Infof("error handler: failed to extract ServerMetadata from context")
	}

	handleForwardResponseServerMetadata(h.OutgoingHeaderMatcher, rw, md)
	handleForwardResponseTrailerHeader(rw, md)

	h.writeError(ctx, false, marshaler, rw, err)

	handleForwardResponseTrailer(rw, md)
}

// StreamHandler implements ProtoStreamErrorHandlerFunc
// in accordance with REST API Syntax Specification.
// See RestError for the JSON format of an error
func (h *ProtoErrorHandler) StreamHandler(ctx context.Context, headerWritten bool, mux *runtime.ServeMux, marshaler runtime.Marshaler, rw http.ResponseWriter, req *http.Request, err error) {
	h.writeError(ctx, headerWritten, marshaler, rw, err)
}

func (h *ProtoErrorHandler) writeError(ctx context.Context, headerWritten bool, marshaler runtime.Marshaler, rw http.ResponseWriter, err error) {
	const fallback = `{"code":"INTERNAL","status":500,"message":"%s"}`

	st, ok := status.FromError(err)
	if !ok {
		st = status.New(codes.Unknown, err.Error())
	}

	details := []interface{}{}
	var fields interface{}

	for _, d := range st.Details() {
		switch d.(type) {
		case *errdetails.TargetInfo:
			details = append(details, d)
		case *errfields.FieldInfo:
			fields = d
		default:
			grpclog.Infof("error handler: failed to recognize error message")
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	restErr := Status(ctx, st).ToMap()
	if len(details) > 0 {
		restErr["details"] = details
	}
	if fields != nil {
		restErr["fields"] = fields
	}

	errs, sucs := errorsAndSuccessesFromContext(ctx)
	restResp := &RestResp{
		Error:   errs,
		Success: sucs,
	}
	restResp.Error = append(restResp.Error, restErr)
	if !headerWritten {
		rw.Header().Del("Trailer")
		rw.Header().Set("Content-Type", marshaler.ContentType())
		rw.WriteHeader(Status(ctx, st).HTTPStatus)
	}

	buf, merr := marshaler.Marshal(restResp)
	if merr != nil {
		grpclog.Infof("error handler: failed to marshal error message %q: %v", restErr, merr)
		rw.WriteHeader(http.StatusInternalServerError)

		if _, err := io.WriteString(rw, fmt.Sprintf(fallback, merr)); err != nil {
			grpclog.Infof("error handler: failed to write response: %v", err)
		}
		return
	}

	if _, err := rw.Write(buf); err != nil {
		grpclog.Infof("error handler: failed to write response: %v", err)
	}
}

// For small performance bump, switch map[string]string to a tuple-type (string, string)

type MessageWithFields interface {
	error
	GetFields() map[string]string
	GetMessage() string
}
type messageWithFields struct {
	message string
	fields  map[string]string
}

func (m *messageWithFields) Error() string {
	return m.message
}
func (m *messageWithFields) GetFields() map[string]string {
	return m.fields
}
func (m *messageWithFields) GetMessage() string {
	return m.message
}

// NewWithFields stub comment TODO
func NewWithFields(message string, kvpairs ...string) MessageWithFields {
	mwf := &messageWithFields{message: message, fields: make(map[string]string)}
	for i := 0; i+1 < len(kvpairs); i += 2 {
		mwf.fields[kvpairs[i]] = kvpairs[i+1]
	}
	return mwf
}

// For giving each error a unique metadata key, but not leaking the exact count
// of errors or something like that
var counter *uint32

func init() {
	counter = new(uint32)
	*counter = uint32(time.Now().Nanosecond() % math.MaxUint32)
}

func WithError(ctx context.Context, err error) {
	i := atomic.AddUint32(counter, uint32(time.Now().Nanosecond()%100+1))
	md := metadata.Pairs(fmt.Sprintf("error-%d", i), fmt.Sprintf("message:%s", err.Error()))
	if mwf, ok := err.(MessageWithFields); ok {
		for k, v := range mwf.GetFields() {
			md.Append(fmt.Sprintf("error-%d", i), fmt.Sprintf("%s:%s", k, v))
		}
	}
	grpc.SetTrailer(ctx, md)
}

func WithSuccess(ctx context.Context, msg MessageWithFields) {
	i := atomic.AddUint32(counter, uint32(time.Now().Nanosecond()%100+1))
	md := metadata.Pairs(fmt.Sprintf("success-%d", i), fmt.Sprintf("message:%s", msg.Error()))
	for k, v := range msg.GetFields() {
		md.Append(fmt.Sprintf("success-%d", i), fmt.Sprintf("%s:%s", k, v))

	}
	grpc.SetTrailer(ctx, md)

}

func errorsAndSuccessesFromContext(ctx context.Context) (errors []map[string]interface{}, successes []map[string]interface{}) {
	md, _ := runtime.ServerMetadataFromContext(ctx)
	errors = make([]map[string]interface{}, 0)
	successes = make([]map[string]interface{}, 0)
	for k, vs := range md.TrailerMD {
		if strings.HasPrefix(k, "error-") {
			err := make(map[string]interface{})
			for _, v := range vs {
				parts := strings.SplitN(v, ":", 2)
				err[parts[0]] = parts[1]
			}
			errors = append(errors, err)
		}
		if strings.HasPrefix(k, "success-") {
			success := make(map[string]interface{})
			for _, v := range vs {
				parts := strings.SplitN(v, ":", 2)
				success[parts[0]] = parts[1]
			}
			successes = append(successes, success)
		}
	}
	return
}
