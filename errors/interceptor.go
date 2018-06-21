package errors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor returns grpc.UnaryServerInterceptor
// that should be used as a middleware to generate Error Messages
// with Details and Field Information with Mapping given.
func UnaryServerInterceptor(mapFuncs ...MapFunc) grpc.UnaryServerInterceptor {

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (res interface{}, err error) {

		// Handle panic.
		defer func() {
			if perr := recover(); perr != nil {
				err = status.Errorf(codes.Internal, "errors interceptor: %s", perr)
				grpclog.Errorln(err)
				res, err = nil, err
			}
		}()

		// Initialize container with mapping.
		container := InitContainer()
		mapper := container.AddMapping(mapFuncs...)

		// Save container in context.
		ctx = NewContext(ctx, container)

		// Execute handler.
		res, err = handler(ctx, req)

		if err != nil {
			// Return container as-is.
			if v, ok := err.(*Container); ok {
				return nil, v
			}

			// Perform mapping and return error if not nil.
			if err := mapper.Map(ctx, err); err != nil {
				return nil, err
			}
		}

		return res, nil
	}
}
