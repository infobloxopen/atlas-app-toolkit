package tracing

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"

	"github.com/infobloxopen/atlas-app-toolkit/auth"
)

const (
	//RequestHeaderAnnotationPrefix is a prefix which is added to each request header attribute
	RequestHeaderAnnotationPrefix = "request.header."

	//RequestTrailerAnnotationPrefix is a prefix which is added to each request trailer attribute
	RequestTrailerAnnotationPrefix = "request.trailer."

	//ResponseHeaderAnnotationPrefix is a prefix which is added to each response header attribute
	ResponseHeaderAnnotationPrefix = "response.header."

	//ResponseTrailerAnnotationPrefix is a prefix which is added to each response header attribute
	ResponseTrailerAnnotationPrefix = "response.trailer."

	//RequestPayloadAnnotationKey is a key under which request payload stored in span
	RequestPayloadAnnotationKey = "request.payload"

	//ResponsePayloadAnnotationKey is a key under which response payload stored in span
	ResponsePayloadAnnotationKey = "response.payload"

	//ResponseErrorKey is a key under which response error will be stored in span
	ResponseErrorKey = "response.error"

	//DefaultMaxPayloadSize represent max payload size which will be added to span
	DefaultMaxPayloadSize = 1024 * 1024

	//TruncatedMarkerKey is a key for annotation which will be presented in span in case payload was truncated
	TruncatedMarkerKey = "payload.truncated"

	//TruncatedMarkerValue is a value for annotation which will be presented in span in case payload was truncated
	TruncatedMarkerValue = "true"

	//ObfuscationFactor is a percent of value which will be omitted from obfuscated value
	ObfuscationFactor = 0.80
)

var sensitiveHeaders = map[string]struct{}{
	auth.AuthorizationHeader: struct{}{},
}

type headerMatcher func(string) (string, bool)

type httpOptions struct {
	spanWithHeaders func(*http.Request) bool
	headerMatcher   headerMatcher

	spanWithPayload func(*http.Request) bool
	maxPayloadSize  int
}

//HTTPOption allows extending handler with additional functionality
type HTTPOption func(*httpOptions)

func defaultHTTPOptions() *httpOptions {
	return &httpOptions{
		headerMatcher:  defaultHeaderMatcher,
		maxPayloadSize: DefaultMaxPayloadSize,

		//Keep spanWithHeaders and spanWithPayload equals to nil instead of dummy functions
		//to prevent path trough header for each request
	}
}

//WithHeadersAnnotation annotate span with http headers
func WithHeadersAnnotation(f func(*http.Request) bool) HTTPOption {
	return func(ops *httpOptions) {
		ops.spanWithHeaders = f
	}
}

//WithHeaderMatcher set header matcher to filterout or preprocess headers
func WithHeaderMatcher(f func(string) (string, bool)) HTTPOption {
	return func(ops *httpOptions) {
		ops.headerMatcher = f
	}
}

//WithPayloadAnnotation add request/response body as an attribute to span if f returns true
func WithPayloadAnnotation(f func(*http.Request) bool) HTTPOption {
	return func(ops *httpOptions) {
		ops.spanWithPayload = f
	}
}

//WithHTTPPayloadSize limit payload size propagated to span
//in case payload exceeds limit, payload truncated and
//annotation payload.truncated=true added into span
func WithHTTPPayloadSize(maxSize int) HTTPOption {
	return func(ops *httpOptions) {
		ops.maxPayloadSize = maxSize
	}
}

//NewMiddleware wrap handler
func NewMiddleware(ops ...HTTPOption) func(http.Handler) http.Handler {
	options := defaultHTTPOptions()
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

//Check that &Handler comply with http.Handler interface
var _ http.Handler = &Handler{}

//Handler is a opencensus http plugin wrapper which do some useful things to reach traces
type Handler struct {
	child http.Handler

	options *httpOptions
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	span := trace.FromContext(r.Context())

	withHeaders := h.options.spanWithHeaders != nil && h.options.spanWithHeaders(r)
	withPayload := h.options.spanWithPayload != nil && h.options.spanWithPayload(r)

	if withHeaders {
		//Annotate span with request headers
		attrs := headersToAttributes(r.Header, RequestHeaderAnnotationPrefix, h.options.headerMatcher)
		span.AddAttributes(attrs...)

		//Annotate span with response headers
		//calling in defer to get final headers state, after passing all handlers in chain
		defer func() {
			attrs := headersToAttributes(w.Header(), ResponseHeaderAnnotationPrefix, h.options.headerMatcher)
			span.AddAttributes(attrs...)
		}()
	}

	if withPayload {
		//Wraping of r.Body to allow child handlers read request body
		requestPayload, _ := ioutil.ReadAll(r.Body) //TODO: handle error
		_ = r.Body.Close()

		r.Body = ioutil.NopCloser(bytes.NewBuffer(requestPayload))

		requestPayload, truncated := truncatePayload(requestPayload, h.options.maxPayloadSize)
		if truncated {
			markSpanTruncated(span)
		}

		attrs := []trace.Attribute{trace.StringAttribute(RequestPayloadAnnotationKey, string(requestPayload))}
		span.Annotate(attrs, "Request payload")

		//Wrap for http.ResponseWriter to get response body after passing all handlers in chain
		wrapper := newResponseWrapper(w)
		w = wrapper

		defer func() {
			responsePayload, truncated := truncatePayload(wrapper.buffer.Bytes(), h.options.maxPayloadSize)
			if truncated {
				markSpanTruncated(span)
			}

			attrs := []trace.Attribute{trace.StringAttribute(ResponsePayloadAnnotationKey, string(responsePayload))}
			span.Annotate(attrs, "Response payload")
		}()
	}

	h.child.ServeHTTP(w, r)
}

func headersToAttributes(headers http.Header, prefix string, matcher headerMatcher) []trace.Attribute {
	attributes := make([]trace.Attribute, 0, len(headers))
	for k, vals := range headers {
		key, ok := matcher(k)
		if !ok {
			continue
		}

		key = fmt.Sprint(prefix, key)
		valsStr := strings.Join(vals, ", ")
		if _, ok := sensitiveHeaders[k]; ok {
			valsStr = obfuscate(valsStr)
		}

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
	//In case we receive an error from Writing into buffer we just skip it
	//because adding payload to span is not so critical as provide response
	_, _ = w.buffer.Write(b)
	return w.ResponseWriter.Write(b)
}

func markSpanTruncated(s *trace.Span) {
	s.AddAttributes(trace.StringAttribute(TruncatedMarkerKey, TruncatedMarkerValue))
}

func truncatePayload(payload []byte, payloadLimit int) ([]byte, bool) {
	if len(payload) <= payloadLimit {
		return payload, false
	}

	payload = payload[:payloadLimit]
	flag := []byte("...")
	payload = append(payload[:payloadLimit-3], flag...)

	return payload, true
}

func obfuscate(x string) string {
	countChars := int(float64(len(x)) * (1.0 - ObfuscationFactor))

	return x[:countChars] + "..."
}

//defaultHeaderMatcher is a header matcher which just accept all headers
func defaultHeaderMatcher(h string) (string, bool) {
	return h, true
}

//AlwaysHTTP for each request returns true
func AlwaysHTTP(_ *http.Request) bool {
	return true
}
