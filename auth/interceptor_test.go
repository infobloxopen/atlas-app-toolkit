package auth

import (
	"context"
	"testing"

	"github.com/dgrijalva/jwt-go"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	mock_transport "github.com/infobloxopen/atlas-app-toolkit/mocks/transport"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	testJWT                 = `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWJqZWN0Ijp7ImlkIjoidGVzdElEIiwic3ViamVjdF90eXBlIjoidGVzdFVzZXIiLCJhdXRoZW50aWNhdGlvbl90eXBlIjoidGVzdCJ9LCJhY2NvdW50X2lkIjoidGVzdC1hY2MtaWQiLCJjdXN0b21fZmllbGQiOiJ0ZXN0LWN1c3RvbS1maWVsZCJ9.pEuJadBkY_twamJid9GKHGZWtIHsZ3cXv84sRqPG-vw`
	testAuthorizationHeader = "authorization"
	testAccountID           = "test-acc-id"
	testFullMethod          = "/app.Object/TestMethod"
)

type testRequest struct{}

type testResponse struct{}

func TestLogrusUnaryServerInterceptor(t *testing.T) {
	testHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		logger := ctxlogrus.Extract(ctx)
		assert.Equal(t, logger.Data[MultiTenancyField], testAccountID)
		return nil, nil
	}
	chain := grpc_middleware.ChainUnaryServer(
		grpc_logrus.UnaryServerInterceptor(logrus.NewEntry(logrus.StandardLogger())),
		LogrusUnaryServerInterceptor(),
	)
	ctx := contextWithToken(makeToken(jwt.MapClaims{MultiTenancyField: testAccountID}, t), DefaultTokenType)
	chain(ctx, nil, &grpc.UnaryServerInfo{}, testHandler)

	negativeTestHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		logger := ctxlogrus.Extract(ctx)
		_, ok := logger.Data[MultiTenancyField]
		assert.False(t, ok)
		return nil, nil
	}
	chain(context.Background(), nil, &grpc.UnaryServerInfo{}, negativeTestHandler)
}

func TestLogrusStreamServerInterceptor(t *testing.T) {
	handler := func(srv interface{}, stream grpc.ServerStream) error {
		logger := ctxlogrus.Extract(stream.Context())
		assert.Equal(t, logger.Data[MultiTenancyField], testAccountID)
		return nil
	}
	ctx := mock_transport.DummyContextWithServerTransportStream()
	md := metadata.Pairs(testAuthorizationHeader, testJWT)
	newCtx := metadata.NewIncomingContext(ctx, md)
	streamInterceptor := grpc_middleware.ChainStreamServer(
		grpc_logrus.StreamServerInterceptor(logrus.NewEntry(logrus.StandardLogger())),
		LogrusStreamServerInterceptor(),
	)
	if err := streamInterceptor(testRequest{}, mock_transport.NewMockServerStream(newCtx), &grpc.StreamServerInfo{FullMethod: testFullMethod}, handler); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
