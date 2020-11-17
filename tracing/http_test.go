package tracing

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
)

var testHTTPOpts = &httpOptions{}
var expectedStr = "test"

func Test_truncatePayload(t *testing.T) {
	tests := []struct {
		in      []byte
		out     []byte
		outFlag bool

		limit int
	}{
		{
			in:      []byte("Hello World"),
			out:     []byte("Hello World"),
			outFlag: false,
			limit:   10000000,
		},
		{
			in:      []byte("Hello World"),
			out:     []byte("He..."),
			outFlag: true,
			limit:   5,
		},
		{
			in:      []byte("Hello"),
			out:     []byte("Hello"),
			outFlag: false,
			limit:   5,
		},
	}

	for _, tt := range tests {
		out, reduced := truncatePayload(tt.in, tt.limit)
		if tt.outFlag != reduced {
			t.Errorf("Unexpected result expected %t, got %t", tt.outFlag, reduced)
		}

		if string(out) != string(tt.out) {
			t.Errorf("Unexpected result\n\texpected %q\n\tgot %q", tt.out, out)
		}
	}
}

func Test_obfuscate(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{
			in:  "HelloWorld",
			out: "He...",
		},
		{
			in:  "",
			out: "...",
		},
		{
			in:  "H",
			out: "...",
		},
	}

	for _, tt := range tests {
		out := obfuscate(tt.in)
		if out != tt.out {
			t.Errorf("Unexpected result\n\texpected %q\n\tgot %q", tt.out, out)
		}
	}
}

func TestDefaultHTTPOptions(t *testing.T) {
	expected := &httpOptions{
		headerMatcher:  defaultHeaderMatcher,
		maxPayloadSize: 1048576,
	}

	result := defaultHTTPOptions()
	expectedHeader, expectedBool := expected.headerMatcher(expectedStr)
	resultHeader, resultBool := result.headerMatcher(expectedStr)
	assert.True(t, expectedBool)
	assert.Equal(t, expectedBool, resultBool)
	assert.Equal(t, expectedHeader, resultHeader)
	assert.Equal(t, expected.maxPayloadSize, result.maxPayloadSize)
}

func TestWithHeadersAnnotation(t *testing.T) {
	option := WithHeadersAnnotation(func(r *http.Request) bool {
		return true
	})
	option(testHTTPOpts)
	assert.True(t, testHTTPOpts.spanWithHeaders(nil))
}

func TestWithHeaderMatcher(t *testing.T) {
	option := WithHeaderMatcher(defaultHeaderMatcher)
	option(testHTTPOpts)
	resultStr, ok := testHTTPOpts.headerMatcher(expectedStr)
	assert.True(t, ok)
	assert.Equal(t, expectedStr, resultStr)
}

func TestWithPayloadAnnotation(t *testing.T) {
	option := WithPayloadAnnotation(func(r *http.Request) bool {
		return true
	})
	option(testHTTPOpts)
	assert.True(t, testHTTPOpts.spanWithPayload(nil))
}

func TestWithHTTPPayloadSize(t *testing.T) {
	option := WithHTTPPayloadSize(333)
	option(testHTTPOpts)
	assert.Equal(t, 333, testHTTPOpts.maxPayloadSize)
}

func TestNewMiddleware(t *testing.T) {
	handlerFunc := NewMiddleware(func(options *httpOptions) {
		options.spanWithHeaders = func(r *http.Request) bool {
			return true
		}
	})

	handler := handlerFunc(&httpHandlerMock{})
	ocHandler, ok := handler.(*ochttp.Handler)
	assert.True(t, ok)

	result, ok := (ocHandler.Handler).(*Handler)
	assert.True(t, ok)
	assert.True(t, result.options.spanWithHeaders(nil))
}

func TestHandler_ServeHTTP(t *testing.T) {
	handlerFunc := NewMiddleware(func(options *httpOptions) {
		options.spanWithHeaders = func(r *http.Request) bool {
			return true
		}

		options.spanWithPayload = func(r *http.Request) bool {
			return true
		}
	})

	ctx, _ := trace.StartSpan(context.Background(), "test span", trace.WithSampler(trace.AlwaysSample()))

	r, _ := http.NewRequest("", "", bytes.NewBuffer([]byte("test body")))
	r.Header = map[string][]string{
		"test1": {"test11"},
		"test2": {"test22"},
	}
	r = r.WithContext(ctx)

	w := httptest.NewRecorder()
	w.Header().Add("test3", "")

	result := &httpHandlerMock{}
	handler := handlerFunc(result)
	handler.ServeHTTP(w, r)
	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test that Span attributes were populated with Headers
	reflectAttrMap := reflect.ValueOf(trace.FromContext(result.request.Context())).Elem().Field(3).Elem().Field(0)
	reflectKeys := reflectAttrMap.MapKeys()
	assert.Len(t, reflectKeys, 8)

	resultHeadersMap := make(map[string]string)
	for _, k := range reflectKeys {
		key := k.Convert(reflectAttrMap.Type().Key())
		val := reflectAttrMap.MapIndex(key)
		resultHeadersMap[fmt.Sprint(key)] = fmt.Sprint(val)
	}
	assert.Equal(t, "true", resultHeadersMap[fmt.Sprint(RequestHeaderAnnotationPrefix, "test1")])
	assert.Equal(t, "true", resultHeadersMap[fmt.Sprint(RequestHeaderAnnotationPrefix, "test2")])
	assert.Equal(t, "true", resultHeadersMap[fmt.Sprint(ResponseHeaderAnnotationPrefix, "Test3")])

	// Test that Span annotations were populated with payload attributes and annotation messages
	reflectAnnotations := reflect.ValueOf(trace.FromContext(result.request.Context())).Elem().Field(4).Elem().Field(0).Slice(0, 2)
	resultRequestPayloadMsg := fmt.Sprint(reflectAnnotations.Index(0).Elem().Field(1))
	resultResponsePayloadMsg := fmt.Sprint(reflectAnnotations.Index(1).Elem().Field(1))
	assert.Equal(t, "Request payload", resultRequestPayloadMsg)
	assert.Equal(t, "Response payload", resultResponsePayloadMsg)

	// Test that Span contains given payload
	reflectAttrMap = reflectAnnotations.Index(0).Elem().Field(2)
	reflectKeys = reflectAttrMap.MapKeys()
	resultPayload := fmt.Sprint(reflectAttrMap.MapIndex(reflectKeys[0]))
	assert.Equal(t, "test body", resultPayload)
}

func TestHeadersToAttributes(t *testing.T) {
	expected := append(make([]trace.Attribute, 0, 2), trace.StringAttribute("prefix-test1", "test11"), trace.StringAttribute("prefix-test2", "test22"))
	testHeaders := map[string][]string{
		"test1": {"test11"},
		"test2": {"test22"},
	}

	result := headersToAttributes(testHeaders, "prefix-", defaultHeaderMatcher)

	assert.Len(t, result, 2)
	for _, attribute := range result {
		found := false
		for _, expAttribute := range expected {
			if reflect.DeepEqual(expAttribute, attribute) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Attribute %+v not found in result", attribute)
		}
	}
}

func TestMarkSpanTruncated(t *testing.T) {
	_, span := trace.StartSpan(context.Background(), "test span", trace.WithSampler(trace.AlwaysSample()))
	markSpanTruncated(span)

	reflectAttrMap := reflect.ValueOf(span).Elem().Field(3).Elem().Field(0)
	reflectKeys := reflectAttrMap.MapKeys()
	assert.Len(t, reflectKeys, 1)

	reflectKey := reflectKeys[0].Convert(reflectAttrMap.Type().Key())
	resultKey := fmt.Sprint(reflectKey)
	resultValue := fmt.Sprint(reflectAttrMap.MapIndex(reflectKey))
	assert.Equal(t, TruncatedMarkerKey, resultKey)
	assert.Equal(t, resultValue, TruncatedMarkerValue)
}

func TestNewResponseWrapper(t *testing.T) {
	expected := &bytes.Buffer{}
	result := newResponseWrapper(&httpResponseWriterMock{})
	assert.NotNil(t, result.ResponseWriter)
	assert.Equal(t, expected, result.buffer)
}

func TestResponseBodyWrapper_Write(t *testing.T) {
	wrapper := newResponseWrapper(&httpResponseWriterMock{})
	result, err := wrapper.Write([]byte("3"))
	assert.NoError(t, err)
	assert.Equal(t, 0, result)
}

func TestDefaultHeaderMatcher(t *testing.T) {
	result, ok := defaultHeaderMatcher(expectedStr)
	assert.Equal(t, expectedStr, result)
	assert.True(t, ok)
}

func TestAlwaysHTTP(t *testing.T) {
	test, _ := http.NewRequest("", "", nil)
	result := AlwaysHTTP(test)
	assert.True(t, result)
}

type httpResponseWriterMock struct {
	http.ResponseWriter
}

func (fake *httpResponseWriterMock) Write(_ []byte) (int, error) {
	return 0, nil
}

type httpHandlerMock struct {
	http.Handler
	writer  http.ResponseWriter
	request *http.Request
}

func (fake *httpHandlerMock) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fake.writer = w
	fake.request = r
}
