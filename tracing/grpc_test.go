package tracing

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opencensus.io/trace"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/stats"
)

var testGRPCOptions = &gRPCOptions{}

func TestDefaultGRPCOptions(t *testing.T) {
	expected := &gRPCOptions{
		metadataMatcher: defaultMetadataMatcher,
		maxPayloadSize:  1048576,
	}

	result := defaultGRPCOptions()
	expectedHeader, expectedBool := expected.metadataMatcher(expectedStr)
	resultHeader, resultBool := result.metadataMatcher(expectedStr)
	assert.True(t, expectedBool)
	assert.Equal(t, expectedBool, resultBool)
	assert.Equal(t, expectedHeader, resultHeader)
	assert.Equal(t, expected.maxPayloadSize, result.maxPayloadSize)
}

func TestWithMetadataAnnotation(t *testing.T) {
	option := WithMetadataAnnotation(func(ctx context.Context, stats stats.RPCStats) bool {
		return true
	})
	option(testGRPCOptions)
	assert.True(t, testGRPCOptions.spanWithMetadata(nil, nil))
}

func TestWithMetadataMatcher(t *testing.T) {
	option := WithMetadataMatcher(func(s string) (string, bool) {
		return s, true
	})
	option(testGRPCOptions)
	resultStr, ok := testGRPCOptions.metadataMatcher(expectedStr)
	assert.True(t, ok)
	assert.Equal(t, expectedStr, resultStr)
}

func TestWithGRPCPayloadAnnotation(t *testing.T) {
	option := WithGRPCPayloadAnnotation(func(ctx context.Context, rpcStats stats.RPCStats) bool {
		return true
	})
	option(testGRPCOptions)
	assert.True(t, testGRPCOptions.spanWithPayload(nil, nil))
}

func TestWithGRPCPayloadLimit(t *testing.T) {
	option := WithGRPCPayloadLimit(333)
	option(testGRPCOptions)
	assert.Equal(t, 333, testGRPCOptions.maxPayloadSize)
}

func TestNewServerHandler(t *testing.T) {
	result := NewServerHandler(func(options *gRPCOptions) {
		options.spanWithPayload = func(ctx context.Context, rpcStats stats.RPCStats) bool {
			return true
		}
	})

	matcherStr, ok := result.options.metadataMatcher(expectedStr)
	assert.True(t, ok)
	assert.Equal(t, expectedStr, matcherStr)
	assert.True(t, result.options.spanWithPayload(nil, nil))
	assert.Equal(t, DefaultMaxPayloadSize, result.options.maxPayloadSize)
}

func TestServerHandler_HandleRPC(t *testing.T) {
	handler := NewServerHandler(func(options *gRPCOptions) {
		options.spanWithPayload = func(ctx context.Context, rpcStats stats.RPCStats) bool {
			return true
		}

		options.spanWithMetadata = func(ctx context.Context, rpcStats stats.RPCStats) bool {
			return true
		}
	})

	expectedStats := []stats.RPCStats{
		&stats.End{
			Error: fmt.Errorf(""),
		},
		&stats.InHeader{
			Header: map[string][]string{
				"header1": {""},
			},
		},
		&stats.InTrailer{
			Trailer: map[string][]string{
				"trailer1": {""},
			},
		},
		&stats.OutHeader{
			Header: map[string][]string{
				"outHeader1": {""},
			},
		},
		&stats.OutTrailer{
			Trailer: map[string][]string{
				"outTrailer1": {""},
			},
		},
		&stats.InPayload{
			Payload: []byte(""),
		},
		&stats.OutPayload{
			Payload: []byte(""),
		},
	}

	ctx, _ := trace.StartSpan(context.Background(), "test span", trace.WithSampler(trace.AlwaysSample()))

	for _, v := range expectedStats {
		handler.HandleRPC(ctx, v)
	}

	expectedMap := map[string]string{
		"request.header.header1":       "true",
		"request.trailer.trailer1":     "true",
		"response.header.outHeader1":   "true",
		"response.trailer.outTrailer1": "true",
	}

	resultMap := make(map[string]string, 4)
	reflectAttrs := reflect.ValueOf(trace.FromContext(ctx)).Elem().Field(3).Elem().Field(0)
	reflectKeys := reflectAttrs.MapKeys()
	for _, k := range reflectKeys {
		key := k.Convert(reflectAttrs.Type().Key())
		val := reflectAttrs.MapIndex(key)
		resultMap[fmt.Sprint(key)] = fmt.Sprint(val)
	}

	assert.Equal(t, expectedMap, resultMap)

	expectedAnnotations := []string{
		"Response error", "Request payload", "Response payload",
	}

	resultAnnotations := make([]string, 0, 3)
	reflectedAnnotations := reflect.ValueOf(trace.FromContext(ctx)).Elem().Field(4).Elem().Field(0).Slice(0, 3)
	for i := 0; i < 3; i++ {
		resultAnnotations = append(resultAnnotations, fmt.Sprint(reflectedAnnotations.Index(i).Elem().Field(1)))
	}

	assert.Equal(t, expectedAnnotations, resultAnnotations)
}

func TestMetadataToAttributes(t *testing.T) {
	expected := []trace.Attribute{trace.StringAttribute(fmt.Sprint("prefix.", expectedStr), "test value")}
	result := metadataToAttributes(metadata.MD{expectedStr: {"test value"}}, "prefix.", defaultMetadataMatcher)
	assert.Equal(t, expected, result)
}

func TestPayloadToAttributes(t *testing.T) {
	expected := trace.StringAttribute(expectedStr, "\"test value\"")
	result, ok, err := payloadToAttributes(expectedStr, "test value", 12)
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.Equal(t, expected, result[0])
}

func TestDefaultMetadataMatcher(t *testing.T) {
	resultStr, ok := defaultMetadataMatcher(expectedStr)
	assert.True(t, ok)
	assert.Equal(t, expectedStr, resultStr)
}

func TestAlwaysGRPC(t *testing.T) {
	assert.True(t, AlwaysGRPC(nil, nil))
}
