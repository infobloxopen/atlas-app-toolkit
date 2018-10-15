package auth

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// UnaryServerInterceptor returns grpc.UnaryServerInterceptor which populates request-scoped logrus logger with account_id field
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		addAccountIdToLogger(ctx)
		return handler(ctx, req)
	}
}

func addAccountIdToLogger(ctx context.Context) {
	if accountId, err := GetAccountID(ctx, nil); err == nil {
		ctxlogrus.AddFields(ctx, logrus.Fields{MultiTenancyField: accountId})
	}
}
