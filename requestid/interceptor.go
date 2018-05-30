package requestid

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
)

type requestIDKey struct{}

// UnaryServerInterceptor returns grpc.UnaryServerInterceptor
// that should be used as a middleware to generate/include Request-Id in headers and context
// for tracing and tracking user's request.
//
//
// Returned middleware populates Request-Id from gRPC metadata if
// they defined in a testRequest message else creates a new one.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (res interface{}, err error) {
		// handle panic
		defer func() {
			if perr := recover(); perr != nil {
				err = status.Errorf(codes.Internal, "request id interceptor: %s", perr)
				grpclog.Errorln(err)
				res, err = nil, err
			}
		}()

		reqID := handleRequestID(ctx)

		errd := updateHeader(ctx, reqID)
		if errd != nil {
			errd = status.Errorf(codes.Internal, "request id interceptor: unable to update metadata - %s", errd)
			grpclog.Errorln(errd)
		}

		newCtx := NewContext(ctx, reqID)

		// returning from the request call
		res, err = handler(newCtx, req)

		return
	}
}

// FromContext returns the Request-Id information from ctx if it exists.
func FromContext(ctx context.Context) (reqID string, exists bool) {
	reqID, exists = ctx.Value(requestIDKey{}).(string)
	if !exists {
		return "", false
	}
	return reqID, true
}

// NewContext creates a new context with Request-Id attached.
func NewContext(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, reqID)
}
