package tracing

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
)

const (
	//RequestHeaderAnnotationPrefix is a prefix which is added to each request header attribute
	RequestHeaderAnnotationPrefix = "request.header."

	//ResponseHeaderAnnotationPrefix is a prefix which is added to each response header attribute
	ResponseHeaderAnnotationPrefix = "response.header."

	//RequestPayloadAnnotationKey is a key under which request payload stored in span
	RequestPayloadAnnotationKey = "request.payload"

	//ResponsePayloadAnnotationKey is a key under which response payload stored in span
	ResponsePayloadAnnotationKey = "response.payload"

	//DefaultMaxPayloadSize represent max payload size which will be added to span
	DefaultMaxPayloadSize = 1024 * 1024

	//ReducedMarkerKey is a key for annotation which will be presented in span in case payload was reduced
	ReducedMarkerKey = "reduced_payload"

	//ReducedMarkerValue is a value for annotation which will be presented in span in case payload was reduced
	ReducedMarkerValue = "true"
)

type headerMatcher func(string) (string, bool)

type options struct {
	spanWithHeaders func(*http.Request) bool
	headerMatcher   headerMatcher

	spanWithPayload func(*http.Request) bool
	maxPayloadSize  int
}

//Option is a configuration unit for handler
type Option func(*options)

func newDefaultOptions() *options {
	return &options{
		headerMatcher:  defaultHeaderMatcher,
		maxPayloadSize: DefaultMaxPayloadSize,

		//Keep spanWithHeaders and spanWithPayload equals to nil instead of dummy functions
		//to prevent path trough header for each request
	}
}

//WithHeadersAnnotation annotate span with
func WithHeadersAnnotation(f func(*http.Request) bool) Option {
	return func(ops *options) {
		ops.spanWithHeaders = f
	}
}

//WithHeaderMatcher add request h
func WithHeaderMatcher(matcher func(string) (string, bool)) Option {
	return func(ops *options) {
		ops.headerMatcher = matcher
	}
}

//WithPayloadAnnotation add request/response body as an attribute to span if f returns true
func WithPayloadAnnotation(f func(*http.Request) bool) Option {
	return func(ops *options) {
		ops.spanWithPayload = f
	}
}

//WithPayloadSize ...
func WithPayloadSize(maxSize int) Option {
	return func(ops *options) {
		ops.maxPayloadSize = maxSize
	}
}

//NewTracingMiddleware wrap handler
func NewTracingMiddleware(ops ...Option) func(http.Handler) http.Handler {
	options := newDefaultOptions()
	for _, op := range ops {
		op(options)
	}

	return func(h http.Handler) http.Handler {
		och := &ochttp.Handler{
			Handler: &Handler{
				child:   h,
				options: options,
			},
		}

		return och
	}
}

//Handler is a opencensus http plugin wrapper which do some usefull things to reach traces
type Handler struct {
	child http.Handler

	options *options
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	span := trace.FromContext(r.Context())

	withHeaders := h.options.spanWithHeaders != nil && h.options.spanWithHeaders(r)
	withPayload := h.options.spanWithPayload != nil && h.options.spanWithPayload(r)

	if withHeaders {
		//Annotate span with request headers
		attrs := headerToAttributes(r.Header, RequestHeaderAnnotationPrefix, h.options.headerMatcher)
		span.AddAttributes(attrs...)

		//Annotate span with response headers
		//calling in defer to get final headers state, after passing all handlers in chain
		defer func() {
			attrs := headerToAttributes(w.Header(), ResponseHeaderAnnotationPrefix, h.options.headerMatcher)
			span.AddAttributes(attrs...)
		}()
	}

	if withPayload {
		//Wraping of r.Body to allow child handlers read request body
		requestPayload, _ := ioutil.ReadAll(r.Body) //TODO: handle error
		_ = r.Body.Close()

		r.Body = ioutil.NopCloser(bytes.NewBuffer(requestPayload))

		requestPayload, reduced := shrinkPayloadToLimit(requestPayload, h.options.maxPayloadSize)
		if reduced {
			markSpanReduced(span)
		}

		attrs := []trace.Attribute{trace.StringAttribute(RequestPayloadAnnotationKey, string(requestPayload))}
		span.Annotate(attrs, "Request payload")

		//Wrap for http.ResponseWriter to get response body after passing all handlers in chain
		wrapper := newResponseWrapper(w)
		w = wrapper

		defer func() {
			responsePayload, reduced := shrinkPayloadToLimit(wrapper.buffer.Bytes(), h.options.maxPayloadSize)
			if reduced {
				markSpanReduced(span)
			}

			attrs := []trace.Attribute{trace.StringAttribute(ResponsePayloadAnnotationKey, string(responsePayload))}
			span.Annotate(attrs, "Response payload")
		}()
	}

	h.child.ServeHTTP(w, r)
}

func headerToAttributes(headers http.Header, prefix string, matcher headerMatcher) []trace.Attribute {
	attributes := make([]trace.Attribute, 0, len(headers))
	for k, vals := range headers {
		k, ok := matcher(k)
		if !ok {
			continue
		}

		key := fmt.Sprint(prefix, k)
		valsStr := strings.Join(vals, ", ")

		attributes = append(attributes, trace.StringAttribute(key, valsStr))
	}

	return attributes
}

func newResponseWrapper(w http.ResponseWriter) *responseBodyWrapper {
	return &responseBodyWrapper{
		ResponseWriter: w,
		buffer:         &bytes.Buffer{},
	}
}

//responseBodyWrapper duplicate all bytes written to it to buffer
type responseBodyWrapper struct {
	http.ResponseWriter

	buffer *bytes.Buffer
}

func (w *responseBodyWrapper) Write(b []byte) (int, error) {
	//In case we recieve an error from Writing into buffer we just skip it
	//because adding payload to span is not so critical as provide response
	_, _ = w.buffer.Write(b)
	return w.ResponseWriter.Write(b)
}

func markSpanReduced(s *trace.Span) {
	s.AddAttributes(trace.StringAttribute(ReducedMarkerKey, ReducedMarkerValue))
}

func shrinkPayloadToLimit(payload []byte, payloadLimit int) ([]byte, bool) {
	if len(payload) <= payloadLimit {
		return payload, false
	}

	payload = payload[:payloadLimit]
	flag := []byte("...")
	payload = append(payload[:payloadLimit-3], flag...)

	return payload, true
}

//defaultHeaderMatcher is a header matcher which just accept all headers
func defaultHeaderMatcher(h string) (string, bool) {
	return h, true
}

//Always for each request returns true
func Always(_ *http.Request) bool {
	return true
}
