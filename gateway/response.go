package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"google.golang.org/grpc/grpclog"
	"google.golang.org/protobuf/reflect/protoreflect"

	runtime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

type (
	// ForwardResponseMessageFunc forwards gRPC response to HTTP client inaccordance with REST API Syntax
	ForwardResponseMessageFunc func(context.Context, *runtime.ServeMux, runtime.Marshaler, http.ResponseWriter, *http.Request, protoreflect.ProtoMessage, ...func(context.Context, http.ResponseWriter, protoreflect.ProtoMessage) error)
	// ForwardResponseStreamFunc forwards gRPC stream response to HTTP client inaccordance with REST API Syntax
	ForwardResponseStreamFunc func(context.Context, *runtime.ServeMux, runtime.Marshaler, http.ResponseWriter, *http.Request, func() (protoreflect.ProtoMessage, error), ...func(context.Context, http.ResponseWriter, protoreflect.ProtoMessage) error)
)

// ResponseForwarder implements ForwardResponseMessageFunc in method ForwardMessage
// and ForwardResponseStreamFunc in method ForwardStream
// in accordance with REST API Syntax Specification.
// See: https://github.com/infobloxopen/atlas-app-toolkit#responses
// for format of JSON response.
type ResponseForwarder struct {
	OutgoingHeaderMatcher runtime.HeaderMatcherFunc
	MessageErrHandler     runtime.ErrorHandlerFunc
	StreamErrHandler      ProtoStreamErrorHandlerFunc
}

var (
	// ForwardResponseMessage is default implementation of ForwardResponseMessageFunc
	ForwardResponseMessage = NewForwardResponseMessage(PrefixOutgoingHeaderMatcher, ProtoMessageErrorHandler, ProtoStreamErrorHandler)
	// ForwardResponseStream is default implementation of ForwardResponseStreamFunc
	ForwardResponseStream = NewForwardResponseStream(PrefixOutgoingHeaderMatcher, ProtoMessageErrorHandler, ProtoStreamErrorHandler)

	setStatusDetails = false
)

// IncludeStatusDetails enables/disables output of status & code fields in all http json
// translated in the gateway with this package's ForwardResponseMessage
func IncludeStatusDetails(withDetails bool) {
	setStatusDetails = withDetails
}

// NewForwardResponseMessage returns ForwardResponseMessageFunc
func NewForwardResponseMessage(out runtime.HeaderMatcherFunc, meh runtime.ErrorHandlerFunc, seh ProtoStreamErrorHandlerFunc) ForwardResponseMessageFunc {
	fw := &ResponseForwarder{out, meh, seh}
	return fw.ForwardMessage
}

// NewForwardResponseStream returns ForwardResponseStreamFunc
func NewForwardResponseStream(out runtime.HeaderMatcherFunc, meh runtime.ErrorHandlerFunc, seh ProtoStreamErrorHandlerFunc) ForwardResponseStreamFunc {
	fw := &ResponseForwarder{out, meh, seh}
	return fw.ForwardStream
}

// ForwardMessage implements runtime.ForwardResponseMessageFunc
func (fw *ResponseForwarder) ForwardMessage(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, rw http.ResponseWriter, req *http.Request, resp protoreflect.ProtoMessage, opts ...func(context.Context, http.ResponseWriter, protoreflect.ProtoMessage) error) {
	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		grpclog.Infof("forward response message: failed to extract ServerMetadata from context")
		fw.MessageErrHandler(ctx, mux, marshaler, rw, req, fmt.Errorf("forward response message: internal error"))
	}

	handleForwardResponseServerMetadata(fw.OutgoingHeaderMatcher, rw, md)
	handleForwardResponseTrailerHeader(rw, md)

	rw.Header().Set("Content-Type", marshaler.ContentType(nil))

	if err := handleForwardResponseOptions(ctx, rw, resp, opts); err != nil {
		fw.MessageErrHandler(ctx, mux, marshaler, rw, req, err)
		return
	}

	// here we start doing a bit strange things
	// 1. marshal response into bytes
	// 2. unmarshal bytes into dynamic map[string]interface{}
	// 3. add our custom metadata into dynamic map
	// 4. marshal dynamic map into bytes again :\
	// all that steps are needed because of this requirements:
	// -- To allow compatibility with existing systems,
	// -- the results tag name can be changed to a service-defined tag.
	// -- In this way the success data becomes just a tag added to an existing structure.
	data, err := marshaler.Marshal(resp)
	if err != nil {
		grpclog.Infof("forward response: failed to marshal response: %v", err)
		fw.MessageErrHandler(ctx, mux, marshaler, rw, req, err)
	}

	var dynmap map[string]interface{}
	if err := json.Unmarshal(data, &dynmap); err != nil {
		grpclog.Infof("forward response: failed to unmarshal response: %v", err)
		fw.MessageErrHandler(ctx, mux, marshaler, rw, req, err)
	}

	httpStatus, statusStr := HTTPStatus(ctx, nil)

	retainFields(ctx, req, dynmap)
	errs, suc, _ := errorsAndSuccessFromContext(ctx)
	if _, ok := dynmap["error"]; len(errs) > 0 && !ok {
		dynmap["error"] = errs
	}
	// this is the edge case, if user sends response that has field 'success'
	// let him see his response object instead of our status
	if _, ok := dynmap["success"]; !ok {
		if setStatusDetails {
			if suc == nil {
				suc = map[string]interface{}{}
			}
			suc["code"] = httpStatus
			suc["status"] = statusStr
		}
		if suc != nil {
			dynmap["success"] = suc
		}
	}

	data, err = json.Marshal(dynmap)
	if err != nil {
		grpclog.Infof("forward response: failed to marshal response: %v", err)
		fw.MessageErrHandler(ctx, mux, marshaler, rw, req, err)
	}
	rw.WriteHeader(httpStatus)

	if _, err = rw.Write(data); err != nil {
		grpclog.Infof("forward response: failed to write response: %v", err)
	}

	handleForwardResponseTrailer(rw, md)
}

type delimited interface {
	// Delimiter returns the record seperator for the stream.
	Delimiter() []byte
}

// ForwardStream implements runtime.ForwardResponseStreamFunc.
// RestStatus comes first in the chuncked result.
func (fw *ResponseForwarder) ForwardStream(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, rw http.ResponseWriter, req *http.Request, recv func() (protoreflect.ProtoMessage, error), opts ...func(context.Context, http.ResponseWriter, protoreflect.ProtoMessage) error) {
	flusher, ok := rw.(http.Flusher)
	if !ok {
		grpclog.Infof("forward response stream: flush not supported in %T", rw)
		fw.StreamErrHandler(ctx, false, mux, marshaler, rw, req, fmt.Errorf("forward response message: internal error"))
		return
	}

	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		grpclog.Infof("forward response stream: failed to extract ServerMetadata from context")
		fw.StreamErrHandler(ctx, false, mux, marshaler, rw, req, fmt.Errorf("forward response message: internal error"))
		return
	}
	handleForwardResponseServerMetadata(fw.OutgoingHeaderMatcher, rw, md)

	rw.Header().Set("Transfer-Encoding", "chunked")
	rw.Header().Set("Content-Type", marshaler.ContentType(nil))

	if err := handleForwardResponseOptions(ctx, rw, nil, opts); err != nil {
		fw.StreamErrHandler(ctx, false, mux, marshaler, rw, req, err)
		return
	}

	httpStatus, _ := HTTPStatus(ctx, nil)
	// if user did not set status explicitly
	if httpStatus == http.StatusOK {
		httpStatus = HTTPStatusFromCode(PartialContent)
	}

	rw.WriteHeader(httpStatus)

	var delimiter []byte
	if d, ok := marshaler.(delimited); ok {
		delimiter = d.Delimiter()
	} else {
		delimiter = []byte("\n")
	}

	for {
		resp, err := recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			fw.StreamErrHandler(ctx, true, mux, marshaler, rw, req, err)
			return
		}
		if err := handleForwardResponseOptions(ctx, rw, resp, opts); err != nil {
			fw.StreamErrHandler(ctx, true, mux, marshaler, rw, req, err)
			return
		}

		data, err := marshaler.Marshal(resp)
		if err != nil {
			fw.StreamErrHandler(ctx, true, mux, marshaler, rw, req, err)
			return
		}

		if _, err := rw.Write(data); err != nil {
			grpclog.Infof("forward response stream: failed to write response object: %s", err)
			return
		}

		if _, err = rw.Write(delimiter); err != nil {
			grpclog.Infof("forward response stream: failed to send delimiter chunk: %v", err)
			return
		}
		flusher.Flush()
	}
}

func handleForwardResponseOptions(ctx context.Context, rw http.ResponseWriter, resp protoreflect.ProtoMessage, opts []func(context.Context, http.ResponseWriter, protoreflect.ProtoMessage) error) error {
	if len(opts) == 0 {
		return nil
	}
	for _, opt := range opts {
		if err := opt(ctx, rw, resp); err != nil {
			grpclog.Infof("error handling ForwardResponseOptions: %v", err)
			return err
		}
	}
	return nil
}
