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
// with Details and Field Information.
func UnaryServerInterceptor(mcf MapCondFunc) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (res interface{}, err error) {
		// handle panic
		defer func() {
			if perr := recover(); perr != nil {
				err = status.Errorf(codes.Internal, "errors interceptor: %s", perr)
				grpclog.Errorln(err)
				res, err = nil, err
			}
		}()

		container := NewContainer().WithMapping(mcf)
		ctx = NewContext(ctx, container)

		res, err = handler(ctx, req)

		if err != nil {
			if v, ok := err.(Container); ok {
				return nil, v
			} else {
				if err := container.Map(err); err != nil {
					return nil, err
				}
			}
		}

		return res, nil
	}
}
