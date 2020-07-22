package util

import (
	"context"

	"github.com/grpc/grpc-go/metadata"

	"github.com/infobloxopen/atlas-app-toolkit/auth"
	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"github.com/infobloxopen/atlas-app-toolkit/requestid"
)

// ForwardContext set to outgoing context request_id, auth_token and X-Forwarded-For header value
func ForwardContext(ctx context.Context, token string) context.Context {
	params := map[string][]string{}
	if token != "" {
		params[auth.AuthorizationHeader] = []string{token}
	}

	if rid, ok := requestid.FromContext(ctx); ok {
		params[requestid.DefaultRequestIDKey] = []string{rid}
	}

	if rIPs, ok := gateway.Header(ctx, gateway.XForwardedFor); ok {
		params[gateway.XForwardedFor] = []string{rIPs}
	}

	return ForwardContextWithCustomParams(ctx, params)
}

// ForwardContextWithCustomParams set to outgoing context all params from incoming context + custom params
func ForwardContextWithCustomParams(ctx context.Context, params map[string][]string) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}

	for key := range params {
		md[key] = params[key]
	}

	return metadata.NewOutgoingContext(ctx, md)
}
