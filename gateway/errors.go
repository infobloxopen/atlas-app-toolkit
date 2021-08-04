package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/infobloxopen/atlas-app-toolkit/rpc/errdetails"
	"github.com/infobloxopen/atlas-app-toolkit/rpc/errfields"
)

// ProtoStreamErrorHandlerFunc handles the error as a gRPC error generated via status package and replies to the testRequest.
// Addition bool argument indicates whether method (http.ResponseWriter.WriteHeader) was called or not.
type ProtoStreamErrorHandlerFunc func(context.Context, bool, *runtime.ServeMux, runtime.Marshaler, http.ResponseWriter, *http.Request, error)

type RestErrs struct {
	Error []map[string]interface{} `json:"error,omitempty"`
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
func NewProtoMessageErrorHandler(out runtime.HeaderMatcherFunc) runtime.ErrorHandlerFunc {
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
	var fallback = `{"error":[{"message":"%s"}]}`
	if setStatusDetails {
		fallback = `{"error":[{"message":"%s", "code":500, "status": "INTERNAL"}]}`
	}

	st, ok := status.FromError(err)
	if !ok {
		st = status.New(codes.Unknown, err.Error())
	}
	statusCode, statusStr := HTTPStatus(ctx, st)

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

	restErr := map[string]interface{}{
		"message": st.Message(),
	}
	if len(details) > 0 {
		restErr["details"] = details
	}
	if fields != nil {
		restErr["fields"] = fields
	}
	if setStatusDetails {
		restErr["code"] = statusCode
		restErr["status"] = statusStr
	}

	errs, _, overrideErr := errorsAndSuccessFromContext(ctx)
	restResp := &RestErrs{
		Error: errs,
	}
	if !overrideErr {
		restResp.Error = append(restResp.Error, restErr)
	} else if setStatusDetails && len(restResp.Error) > 0 {
		restResp.Error[0]["code"] = statusCode
		restResp.Error[0]["status"] = statusStr
	}
	if !headerWritten {
		rw.Header().Del("Trailer")
		rw.Header().Set("Content-Type", marshaler.ContentType(nil))
		rw.WriteHeader(statusCode)
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
	GetFields() map[string]interface{}
	GetMessage() string
}
type messageWithFields struct {
	message string
	fields  map[string]interface{}
}

func (m *messageWithFields) Error() string {
	return m.message
}
func (m *messageWithFields) GetFields() map[string]interface{} {
	return m.fields
}
func (m *messageWithFields) GetMessage() string {
	return m.message
}

// NewWithFields returns a new MessageWithFields that requires a message string,
// and then treats the following arguments as alternating keys and values
// a non-string key will immediately return the result so far, ignoring later
// values. The values can be any type
func NewWithFields(message string, kvpairs ...interface{}) MessageWithFields {
	mwf := &messageWithFields{message: message, fields: make(map[string]interface{})}
	for i := 0; i+1 < len(kvpairs); i += 2 {
		k, ok := kvpairs[i].(string)
		if !ok {
			return mwf
		}
		mwf.fields[k] = kvpairs[i+1]
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

// WithError will save an error message into the grpc trailer metadata, if it
// is an error that implements MessageWithFields, it also saves the fields.
// This error will then be inserted into the return JSON if the ResponseForwarder
// is used
func WithError(ctx context.Context, err error) {
	i := atomic.AddUint32(counter, uint32(time.Now().Nanosecond()%100+1))
	md := metadata.Pairs(fmt.Sprintf("error-%d", i), fmt.Sprintf("message:%s", err.Error()))
	if mwf, ok := err.(MessageWithFields); ok {
		if f := mwf.GetFields(); f != nil {
			b, _ := json.Marshal(mwf.GetFields())
			md.Append(fmt.Sprintf("error-%d", i), fmt.Sprintf("fields:%q", b))
		}
	}
	grpc.SetTrailer(ctx, md)
}

// NewResponseError sets the error in the context with extra fields, to
// override the standard message-only error
func NewResponseError(ctx context.Context, msg string, kvpairs ...interface{}) error {
	md := metadata.Pairs("error", fmt.Sprintf("message:%s", msg))
	if len(kvpairs) > 0 {
		fields := make(map[string]interface{})
		for i := 0; i+1 < len(kvpairs); i += 2 {
			k, ok := kvpairs[i].(string)
			if !ok {
				grpclog.Infof("Key value for error details must be a string")
				continue
			}
			fields[k] = kvpairs[i+1]
		}
		b, _ := json.Marshal(fields)
		md.Append("error", fmt.Sprintf("fields:%q", b))
	}
	grpc.SetTrailer(ctx, md)
	return errors.New(msg) // Message should be overridden in response writer
}

// NewResponseErrorWithCode sets the return code and returns an error with extra
// fields in MD to be extracted in the gateway response writer
func NewResponseErrorWithCode(ctx context.Context, c codes.Code, msg string, kvpairs ...interface{}) error {
	SetStatus(ctx, status.New(c, msg))
	NewResponseError(ctx, msg, kvpairs...)
	return status.Error(c, msg)
}

// WithSuccess will save a MessageWithFields into the grpc trailer metadata.
// This success message will then be inserted into the return JSON if the
// ResponseForwarder is used
func WithSuccess(ctx context.Context, msg MessageWithFields) {
	i := atomic.AddUint32(counter, uint32(time.Now().Nanosecond()%100+1))
	md := metadata.Pairs(fmt.Sprintf("success-%d", i), fmt.Sprintf("message:%s", msg.Error()))
	if f := msg.GetFields(); f != nil {
		b, _ := json.Marshal(msg.GetFields())
		md.Append(fmt.Sprintf("success-%d", i), fmt.Sprintf("fields:%q", b))
	}
	grpc.SetTrailer(ctx, md)
}

// WithCodedSuccess wraps a SetStatus and WithSuccess call into one, just to make things a little more "elegant"
func WithCodedSuccess(ctx context.Context, c codes.Code, msg string, args ...interface{}) error {
	WithSuccess(ctx, NewWithFields(msg, args))
	return SetStatus(ctx, status.New(c, msg))
}

func errorsAndSuccessFromContext(ctx context.Context) (errors []map[string]interface{}, success map[string]interface{}, errorOverride bool) {
	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		return nil, nil, false
	}
	errors = make([]map[string]interface{}, 0)
	var primaryError map[string]interface{}
	latestSuccess := int64(-1)
	for k, vs := range md.TrailerMD {
		if k == "error" {
			err := make(map[string]interface{})
			for _, v := range vs {
				parts := strings.SplitN(v, ":", 2)
				if parts[0] == "fields" {
					uq, _ := strconv.Unquote(parts[1])
					json.Unmarshal([]byte(uq), &err)
				} else if parts[0] == "message" {
					err["message"] = parts[1]
				}
			}
			primaryError = err
			errorOverride = true
		}
		if strings.HasPrefix(k, "error-") {
			err := make(map[string]interface{})
			for _, v := range vs {
				parts := strings.SplitN(v, ":", 2)
				if parts[0] == "fields" {
					uq, _ := strconv.Unquote(parts[1])
					json.Unmarshal([]byte(uq), &err)
				} else if parts[0] == "message" {
					err["message"] = parts[1]
				}
			}
			errors = append(errors, err)
		}
		if num := strings.TrimPrefix(k, "success-"); num != k {
			// Let the later success messages override previous ones,
			// also account for the possiblity of wraparound with a generous check
			if i, err := strconv.ParseInt(num, 10, 32); err == nil {
				if i > latestSuccess || (i < 1<<12 && latestSuccess > 1<<28) {
					latestSuccess = i
				} else {
					continue
				}
			}
			success = make(map[string]interface{})
			for _, v := range vs {
				parts := strings.SplitN(v, ":", 2)
				if parts[0] == "fields" {
					uq, _ := strconv.Unquote(parts[1])
					json.Unmarshal([]byte(uq), &success)
				} else if parts[0] == "message" {
					success["message"] = parts[1]
				}
			}
		}
	}
	if errorOverride {
		errors = append([]map[string]interface{}{primaryError}, errors...)
	}
	return
}
