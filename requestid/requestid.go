package requestid

import (
	"context"
	"github.com/google/uuid"
	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// DefaultRequestIDKey is the metadata key name for request ID
var DefaultRequestIDKey = "Request-Id"

func HandleRequestId(ctx context.Context) (reqId string) {
	reqId, exists := gateway.Header(ctx, DefaultRequestIDKey)
	if !exists {
		reqId := newRequestId()
		return reqId
	}

	if reqId == "" {
		reqId := newRequestId()
		return reqId
	}

	return reqId
}

func newRequestId() string {
	return uuid.New().String()
}

func UpdateHeader(ctx context.Context, reqId string) error {
	md := metadata.Pairs(DefaultRequestIDKey, reqId)
	return grpc.SetHeader(ctx, md)
}
