package gateway

import (
	"context"
	"fmt"
	"net/http"
	"net/textproto"
	"strings"

	"google.golang.org/grpc/metadata"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

const XForwardedFor = "X-Forwarded-For"

// GetGeoHeaders returns a slice of x-geo- headers.
func GetGeoHeaders() []string {
	return []string{
		"x-geo-org",
		"x-geo-country-code",
		"x-geo-country-name",
		"x-geo-region-code",
		"x-geo-region-name",
		"x-geo-city-name",
		"x-geo-postal-code",
		"x-geo-latitude",
		"x-geo-longitude",
	}
}

// GetXB3Headers returns a slice of x-b3- headers.
func GetXB3Headers() []string {
	return []string{
		"x-b3-traceid",
		"x-b3-parentspanid",
		"x-b3-spanid",
		"x-b3-sampled",
	}
}

// Header returns first value for a given key if it exists in gRPC metadata
// from incoming or outcoming context, otherwise returns (nil, false)
//
// Calls HeaderN(ctx, key, 1)
//
// Provided key is converted to lowercase (see grpc/metadata.New).
// If key is not found the prefix "grpcgateway-" is added to the key and
// key is being searched once again.
func Header(ctx context.Context, key string) (string, bool) {
	if l, ok := HeaderN(ctx, key, 1); ok {
		return l[0], ok
	}
	return "", false
}

// HeaderN returns first n values for a given key if it exists in gRPC metadata
// from incoming or outcoming context, otherwise returns (nil, false)
//
// If n < 0 all values for a given key will be returned
// If n > 0 at least n values will be returned, or (nil, false)
// If n == 0 result is (nil, false)
//
// Provided key is converted to lowercase (see grpc/metadata.New).
// If key is not found the prefix "grpcgateway-" is added to the key and
// key is being searched once again.
func HeaderN(ctx context.Context, key string, n int) (val []string, found bool) {
	if n == 0 {
		return
	}

	if smd, ok := runtime.ServerMetadataFromContext(ctx); ok {
		ctx = metadata.NewIncomingContext(ctx, smd.HeaderMD)
	}

	imd, iok := metadata.FromIncomingContext(ctx)
	omd, ook := metadata.FromOutgoingContext(ctx)

	md := metadata.Join(imd, omd)

	if !iok && !ook {
		return nil, false
	}

	key = strings.ToLower(key)
	if v, ok := md[key]; ok {
		val = append(val, v...)
		found = true
	}
	// If md contains 'key' and 'runtime.MetadataPrefix + key'
	// collect them all
	key = runtime.MetadataPrefix + key
	if v, ok := md[key]; ok {
		val = append(val, v...)
		found = true
	}

	switch {
	case !found:
		return
	case n < 0 || len(val) == n:
		return
	case len(val) < n:
		return nil, false
	default:
		return val[:n], found
	}
}

// PrefixOutgoingHeaderMatcher discards all grpc header metadata.
func PrefixOutgoingHeaderMatcher(key string) (string, bool) {
	return "", false
}

func handleForwardResponseServerMetadata(matcher runtime.HeaderMatcherFunc, w http.ResponseWriter, md runtime.ServerMetadata) {
	for k, vs := range md.HeaderMD {
		if h, ok := matcher(k); ok {
			for _, v := range vs {
				w.Header().Add(h, v)
			}
		}
	}
}

func handleForwardResponseTrailerHeader(w http.ResponseWriter, md runtime.ServerMetadata) {
	for k := range md.TrailerMD {
		if strings.HasPrefix(k, "error-") || strings.HasPrefix(k, "success-") {
			continue
		}
		tKey := textproto.CanonicalMIMEHeaderKey(fmt.Sprintf("%s%s", runtime.MetadataTrailerPrefix, k))
		w.Header().Add("Trailer", tKey)
	}
}

func handleForwardResponseTrailer(w http.ResponseWriter, md runtime.ServerMetadata) {
	for k, vs := range md.TrailerMD {
		tKey := fmt.Sprintf("%s%s", runtime.MetadataTrailerPrefix, k)
		for _, v := range vs {
			w.Header().Add(tKey, v)
		}
	}
}

// GeoIPHeaderMatcher X-Geo-* headers are set of geo metadata from MaxMind DB injected on ingress nginx
func GeoIPHeaderMatcher() runtime.HeaderMatcherFunc {
	return ExtendedDefaultHeaderMatcher(GetGeoHeaders()...)
}

// RequestIDHeaderMatcher request id header contains unique identifier for request
func RequestIDHeaderMatcher() runtime.HeaderMatcherFunc {
	return ExtendedDefaultHeaderMatcher("request-id")
}

// TracingHeaderMatcher tracing headers
func TracingHeaderMatcher() runtime.HeaderMatcherFunc {
	return ExtendedDefaultHeaderMatcher(GetXB3Headers()...)
}

// AtlasDefaultHeaderMatcher func used to add all headers used by atlas-app-toolkit
// This function also passes through all the headers that runtime.DefaultHeaderMatcher handles.
// AtlasDefaultHeaderMatcher can be used as a Incoming/Outgoing header matcher.
func AtlasDefaultHeaderMatcher() func(string) (string, bool) {
	return ChainHeaderMatcher(
		GeoIPHeaderMatcher(),
		RequestIDHeaderMatcher(),
		TracingHeaderMatcher(),
	)
}

// ExtendedDefaultHeaderMatcher func is used to add custom headers to be matched
// from incoming http requests, If this returns true the header will be added to grpc context.
// This function also passes through all the headers that runtime.DefaultHeaderMatcher handles.
func ExtendedDefaultHeaderMatcher(headerNames ...string) func(string) (string, bool) {
	customHeaders := map[string]bool{}
	for _, name := range headerNames {
		customHeaders[strings.ToLower(name)] = true
	}
	return func(headerName string) (string, bool) {
		if key, ok := runtime.DefaultHeaderMatcher(headerName); ok {
			return key, ok
		}
		_, ok := customHeaders[strings.ToLower(headerName)]
		return headerName, ok
	}
}

// ChainHeaderMatcher func is used to build chain on header matcher funcitons
// this function can be used as incoming or outgoing header matcher
// keep in mind that gRPC metadata treat as case insensitive strings
func ChainHeaderMatcher(matchers ...runtime.HeaderMatcherFunc) runtime.HeaderMatcherFunc {
	return func(h string) (string, bool) {
		for _, m := range matchers {
			if k, allow := m(h); allow {
				return k, allow
			}
		}

		return "", false
	}
}
