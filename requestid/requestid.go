package requestid

import (
	"context"

	"github.com/google/uuid"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
)

// DefaultRequestIDKey is the metadata key name for request ID
const DefaultRequestIDKey = "Request-Id"

func handleRequestID(ctx context.Context) (reqID string) {
	reqID, exists := FromContext(ctx)
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

// FromContext returns the Request-Id information from ctx if it exists.
func FromContext(ctx context.Context) (reqID string, exists bool) {
	reqID, exists = gateway.Header(ctx, DefaultRequestIDKey)
	return
}

// NewContext creates a new context with Request-Id attached if not exists.
func NewContext(ctx context.Context, reqID string) context.Context {
	md := metadata.Pairs(DefaultRequestIDKey, reqID)
	return metadata.NewOutgoingContext(ctx, md)
}

func addRequestIDToLogger(ctx context.Context, reqID string) {
	ctxlogrus.AddFields(ctx, logrus.Fields{DefaultRequestIDKey: reqID})
}
