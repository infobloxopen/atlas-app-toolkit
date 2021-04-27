package requestid

import (
	"context"

	"github.com/google/uuid"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"

	"github.com/infobloxopen/atlas-app-toolkit/gateway"
)

// DefaultRequestIDKey is the metadata key name for request ID
const (
	DeprecatedRequestIDKey = "Request-Id"
	DefaultRequestIDKey    = "X-Request-ID"
	RequestIDLogKey        = "request_id"
)

// HandleRequestID either extracts a existing and valid request ID from the context or generates a new one
func HandleRequestID(ctx context.Context) (reqID string) {
	reqID, exists := FromContext(ctx)
	if !exists || reqID == "" {
		reqID := newRequestID()
		return reqID
	}
	return reqID
}

func newRequestID() string {
	return uuid.New().String()
}

// FromContext returns the Request-Id information from ctx if it exists.
func FromContext(ctx context.Context) (string, bool) {
	if reqID, ok := gateway.Header(ctx, DefaultRequestIDKey); ok {
		return reqID, ok
	}

	if reqID, ok := gateway.Header(ctx, DeprecatedRequestIDKey); ok {
		return reqID, ok
	}

	return "", false
}

// NewContext creates a new context with Request-Id attached if not exists.
func NewContext(ctx context.Context, reqID string) context.Context {
	md := metadata.Pairs(DefaultRequestIDKey, reqID)
	return metadata.NewOutgoingContext(ctx, md)
}

func addRequestIDToLogger(ctx context.Context, reqID string) {
	ctxlogrus.AddFields(ctx, logrus.Fields{RequestIDLogKey: reqID})
}
