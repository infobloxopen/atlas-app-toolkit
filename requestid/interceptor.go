package requestid

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
)

type requestIDKey struct{}

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (res interface{}, err error) {
		reqId := HandleRequestId(ctx)

		newCtx := context.WithValue(ctx, requestIDKey{}, reqId)

		// returning from the request call
		res, err = handler(newCtx, req)

		errd := UpdateHeader(newCtx, reqId)
		if errd != nil {
			errd = status.Errorf(codes.Internal, "request id interceptor: %s", errd)
			grpclog.Errorln(errd)
		}
		return
	}
}

func FromContext(ctx context.Context) (reqId string, exists bool) {
	reqId, exists = ctx.Value(requestIDKey{}).(string)
	if !exists {
		return "", false
	}
	return reqId, true
}
