package auth

import (
	"context"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestLogrusUnaryServerInterceptor(t *testing.T) {
	testAccountId := "some-ib-customer"
	testHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		logger := ctxlogrus.Extract(ctx)
		assert.Equal(t, logger.Data[MultiTenancyField], testAccountId)
		return nil, nil
	}
	chain := grpc_middleware.ChainUnaryServer(grpc_logrus.UnaryServerInterceptor(logrus.NewEntry(logrus.StandardLogger())), LogrusUnaryServerInterceptor())
	ctx := contextWithToken(makeToken(jwt.MapClaims{MultiTenancyField: testAccountId}, t), DefaultTokenType)
	chain(ctx, nil, &grpc.UnaryServerInfo{}, testHandler)

	negativeTestHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		logger := ctxlogrus.Extract(ctx)
		_, ok := logger.Data[MultiTenancyField]
		assert.False(t, ok)
		return nil, nil
	}
	chain(context.Background(), nil, &grpc.UnaryServerInfo{}, negativeTestHandler)
}
