package auth

import (
	"context"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// LogrusUnaryServerInterceptor returns grpc.UnaryServerInterceptor which populates request-scoped logrus logger with account_id field
func LogrusUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		addAccountIDToLogger(ctx)
		return handler(ctx, req)
	}
}

// LogrusStreamServerInterceptor returns grpc.StreamServerInterceptor which populates request-scoped logrus logger with account_id field
func LogrusStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		ctx := stream.Context()
		addAccountIDToLogger(ctx)
		wrapped := grpc_middleware.WrapServerStream(stream)
		wrapped.WrappedContext = ctx
		err = handler(srv, wrapped)
		return
	}
}

func addAccountIDToLogger(ctx context.Context) {
	if accountID, err := GetAccountID(ctx, nil); err == nil {
		ctxlogrus.AddFields(ctx, logrus.Fields{MultiTenancyField: accountID})
	}
}
