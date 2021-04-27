package requestid

import (
	"context"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

// UnaryServerInterceptor returns grpc.UnaryServerInterceptor
// that should be used as a middleware to generate/include Request-Id in headers and context
// for tracing and tracking user's request.
//
//
// Returned middleware populates Request-Id from gRPC metadata if
// they defined in a testRequest message else creates a new one.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (res interface{}, err error) {

		reqID := HandleRequestID(ctx)

		// add request id to logger
		addRequestIDToLogger(ctx, reqID)

		ctx = NewContext(ctx, reqID)

		// returning from the request call
		res, err = handler(ctx, req)

		return
	}
}

func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {

		ctx := stream.Context()

		reqID := HandleRequestID(ctx)

		// add request id to logger
		addRequestIDToLogger(ctx, reqID)

		ctx = NewContext(ctx, reqID)

		wrapped := grpc_middleware.WrapServerStream(stream)
		wrapped.WrappedContext = ctx

		err = handler(srv, wrapped)

		return
	}
}
