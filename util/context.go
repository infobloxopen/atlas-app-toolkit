package util

import (
	"context"

	"github.com/grpc/grpc-go/metadata"

	"github.com/infobloxopen/atlas-app-toolkit/auth"
	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"github.com/infobloxopen/atlas-app-toolkit/requestid"
)

// ForwardContext set to outgoing context request_id, auth_token and X-Forwarded-For header value
func ForwardContext(ctx context.Context) context.Context {
	keys := []string{
		auth.AuthorizationHeader,
		requestid.DefaultRequestIDKey,
		gateway.XForwardedFor,
	}

	return ForwardContextWithCustomParams(ctx, keys)
}

// ForwardContextWithCustomParams set to outgoing context specified parameters from incoming context by keys
func ForwardContextWithCustomParams(ctx context.Context, keys []string) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}

	resultMD := make(metadata.MD, 0)

	for _, key := range keys {
		if params := md.Get(key); params != nil {
			resultMD.Append(key, params...)
		}
	}

	return metadata.NewOutgoingContext(ctx, resultMD)
}

// ReplaceParamsInContext replaces or adds the specified parameters to the outgoing context
func ReplaceParamsInContext(ctx context.Context, params map[string][]string) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return ctx
	}

	for key := range params {
		md[key] = params[key]
	}

	return metadata.NewOutgoingContext(ctx, md)
}
