package params

import (
	"context"

	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor returns grpc.UnaryServerInterceptor
// that should be used as a middleware to validate query parameter dependencies
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (res interface{}, err error) {

		_, exists := gateway.Header(ctx, gateway.SortMetaKey)
		if !exists {
			_, lexists := gateway.Header(ctx, gateway.LimitMetaKey)
			_, oexists := gateway.Header(ctx, gateway.OffsetMetaKey)

			if lexists || oexists {
				err = status.Errorf(codes.InvalidArgument, "parameters interceptor: paramters(s) %s/%s can't be used without %s",
					gateway.LimitQueryKey, gateway.OffsetQueryKey, gateway.SortQueryKey)
				return nil, err
			}
		}
		// returning from the request call
		return handler(ctx, req)
	}
}
