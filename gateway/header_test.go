package gateway

import (
	"context"
	"testing"

	"google.golang.org/grpc/metadata"
)

func TestHeader(t *testing.T) {
	imd := metadata.Pairs("key1", "val1")
	omd := metadata.Pairs("key2", "val2", "grpcgateway-key2", "val2")

	ictx := metadata.NewIncomingContext(context.Background(), imd)
	ctx := metadata.NewOutgoingContext(ictx, omd)

	if v, ok := Header(ctx, "key1"); !ok {
		t.Error("failed to get 'key1'")
	} else if v != "val1" {
		t.Errorf("invalid value of 'key1': %s", v)
	}

	if v, ok := Header(ctx, "key2"); !ok {
		t.Error("failed to get 'key2'")
	} else if v != "val2" {
		t.Errorf("invalid value of 'key2': %s", v)
	}
}

func TestHeaderN(t *testing.T) {
	imd := metadata.Pairs("key1", "val1")
	omd := metadata.Pairs("key2", "val2", "grpcgateway-key2", "val2")

	ictx := metadata.NewIncomingContext(context.Background(), imd)
	ctx := metadata.NewOutgoingContext(ictx, omd)

	if v, ok := HeaderN(ctx, "key1", -1); !ok {
		t.Error("failed to get 'key1'")
	} else if len(v) != 1 || v[0] != "val1" {
		t.Errorf("invalid value of 'key1': %s", v)
	}

	if v, ok := HeaderN(ctx, "key2", 2); !ok {
		t.Error("failed to get 'key2'")
	} else if len(v) != 2 || v[0] != "val2" || v[1] != "val2" {
		t.Errorf("invalid value of 'key2': %s", v)
	}

	if v, ok := HeaderN(ctx, "key1", 0); ok || v != nil {
		t.Errorf("invalid result with n==0: %s, %v", v, ok)
	}

	if v, ok := HeaderN(ctx, "key1", 10); ok || v != nil {
		t.Errorf("invalid result with n>len(md): %s, %v", v, ok)
	}
}

func TestPrefixOutgoingHeaderMatcher(t *testing.T) {
	key := "Content-Type"
	v, ok := PrefixOutgoingHeaderMatcher(key)
	if ok {
		t.Errorf("header %s hasn't been discarded: %s", key, v)
	}
}

func TestExtendedDefaultHeaderMatcher(t *testing.T) {
	var customMatcherTests = []struct {
		name          string
		customHeaders []string
		in            string
		isValid       bool
	}{
		{
			name:          "Custom headers | success",
			customHeaders: []string{"Request-ID", "ophid"},
			in:            "Request-Id",
			isValid:       true,
		},
		{
			name:          "Custom headers | failure",
			customHeaders: []string{"Request-ID", "ophid"},
			in:            "RequestId",
			isValid:       false,
		},
		{
			name:          "Default headers | success",
			customHeaders: []string{"Request-ID", "ophid"},
			in:            "grpc-metadata-Request-Id",
			isValid:       true,
		},
		{
			name:          "Default headers | without custom headers",
			customHeaders: []string{},
			in:            "grpc-metadata-Request-Id",
			isValid:       true,
		},
		{
			name:          "custom headers in | without custom headers | failure",
			customHeaders: []string{},
			in:            "CustomHeader",
			isValid:       false,
		},
	}
	for _, tt := range customMatcherTests {
		t.Run(tt.name, func(t *testing.T) {
			f := ExtendedDefaultHeaderMatcher(tt.customHeaders...)
			_, ok := f(tt.in)
			if ok != tt.isValid {
				t.Errorf("got %v, want %v", ok, tt.isValid)
			}
		})
	}
}

func TestAtlasDefaultHeaderMatcher(t *testing.T) {
	var customMatcherTests = []struct {
		name    string
		in      string
		isValid bool
	}{
		{
			name:    "X-Geo-Org | success",
			in:      "X-Geo-Org",
			isValid: true,
		},
		{
			name:    "X-Geo-Country-Code | success",
			in:      "X-Geo-Country-Code",
			isValid: true,
		},
		{
			name:    "X-Geo-Country-Name | success",
			in:      "X-Geo-Country-Name",
			isValid: true,
		},
		{
			name:    "X-Geo-Region-Code | success",
			in:      "X-Geo-Region-Code",
			isValid: true,
		},
		{
			name:    "X-Geo-Region-Name | success",
			in:      "X-Geo-Region-Name",
			isValid: true,
		},
		{
			name:    "X-Geo-City-Name | success",
			in:      "X-Geo-City-Name",
			isValid: true,
		},
		{
			name:    "X-Geo-Postal-Code  | success",
			in:      "X-Geo-Postal-Code",
			isValid: true,
		},
		{
			name:    "X-Geo-Latitude | success",
			in:      "X-Geo-Latitude",
			isValid: true,
		},
		{
			name:    "X-Geo-Longitude | success",
			in:      "X-Geo-Longitude",
			isValid: true,
		},
		{
			name:    "Request-Id | success",
			in:      "Request-Id",
			isValid: true,
		},
		{
			name:    "X-B3-TraceId | success",
			in:      "X-B3-TraceId",
			isValid: true,
		},
		{
			name:    "X-B3-ParentSpanId | success",
			in:      "X-B3-ParentSpanId",
			isValid: true,
		},
		{
			name:    "X-B3-SpanId | success",
			in:      "X-B3-SpanId",
			isValid: true,
		},
		{
			name:    "X-B3-Sampled | success",
			in:      "X-B3-Sampled",
			isValid: true,
		},
		{
			name:    "x-b3-sampled | success",
			in:      "x-b3-sampled",
			isValid: true,
		},
		{
			name:    "Failed-Header | failure",
			in:      "Failed-Header",
			isValid: false,
		},
	}
	for _, tt := range customMatcherTests {
		t.Run(tt.name, func(t *testing.T) {
			f := AtlasDefaultHeaderMatcher()
			_, ok := f(tt.in)
			if ok != tt.isValid {
				t.Errorf("got %v, want %v", ok, tt.isValid)
			}
		})
	}
}

func TestChainHeaderMatcher(t *testing.T) {
	chain := ChainHeaderMatcher(
		func(h string) (string, bool) {
			if h == "first" {
				return h, true
			}

			return "", false
		},
		func(h string) (string, bool) {
			if h == "second" {
				return h, true
			}

			return "", false
		},
		func(h string) (string, bool) {
			if h == "third" {
				return h, true
			}

			return "", false
		},
	)

	var customMatcherTests = []struct {
		name    string
		in      string
		isValid bool
	}{
		{
			name:    "first | success",
			in:      "first",
			isValid: true,
		},
		{
			name:    "second | success",
			in:      "second",
			isValid: true,
		},
		{
			name:    "third | success",
			in:      "third",
			isValid: true,
		},
		{
			name:    "fourth | success",
			in:      "fourth",
			isValid: false,
		},
	}

	for _, tt := range customMatcherTests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := chain(tt.in)
			if ok != tt.isValid {
				t.Errorf("got %v, want %v", ok, tt.isValid)
			}
		})
	}
}
