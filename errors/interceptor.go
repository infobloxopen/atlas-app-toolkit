package errors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor returns grpc.UnaryServerInterceptor
// that should be used as a middleware to generate Error Messages
// with Details and Field Information with Mapping given.
func UnaryServerInterceptor(mapFuncs ...MapFunc) grpc.UnaryServerInterceptor {

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (res interface{}, err error) {

		// Initialize container with mapping.
		container := InitContainer()
		mapper := container.AddMapping(mapFuncs...)

		// Save container in context.
		ctx = NewContext(ctx, container)

		// Execute handler.
		res, err = handler(ctx, req)

		if err != nil {
			// Return container as-is.
			if _, ok := err.(*Container); ok {
				return nil, err
			}

			// Pass protobuf status.
			if _, ok := status.FromError(err); ok {
				return nil, err
			}

			// Perform mapping and return error if not nil.
			if err := mapper.Map(ctx, err); err != nil {
				return nil, err
			}
		}

		return res, nil
	}
}
