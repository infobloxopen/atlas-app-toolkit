package requestid

import (
	"context"

	"github.com/google/uuid"
	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// DefaultRequestIDKey is the metadata key name for request ID
const DefaultRequestIDKey = "Request-Id"

func handleRequestID(ctx context.Context) (reqID string) {
	reqID, exists := gateway.Header(ctx, DefaultRequestIDKey)
	if !exists {
		reqID := newRequestID()
		return reqID
	}

	if reqID == "" {
		reqID := newRequestID()
		return reqID
	}

	return reqID
}

func newRequestID() string {
	return uuid.New().String()
}

func updateHeader(ctx context.Context, reqID string) error {
	md := metadata.Pairs(DefaultRequestIDKey, reqID)
	return grpc.SetHeader(ctx, md)
}
