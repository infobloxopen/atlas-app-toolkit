package params

import (
	"context"
	"testing"

	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/transport"
)

type testRequest struct{}

type testResponse struct{}

var handler = func(ctx context.Context, req interface{}) (interface{}, error) {
	return &testResponse{}, nil
}

func DummyContextWithServerTransportStream() context.Context {
	expectedStream := &transport.Stream{}
	return grpc.NewContextWithServerTransportStream(context.Background(), expectedStream)
}

func TestUnaryServerInterceptorValidCases(t *testing.T) {
	data := [][]string{
		[]string{"somekey", "somevalue"},
		[]string{gateway.SortMetaKey, "name"},
		[]string{gateway.SortMetaKey, "name", gateway.LimitMetaKey, "3"},
		[]string{gateway.SortMetaKey, "name", gateway.OffsetMetaKey, "3"},
		[]string{gateway.SortMetaKey, "name", gateway.OffsetMetaKey, "3", gateway.LimitMetaKey, "3"},
	}
	ctx := DummyContextWithServerTransportStream()
	for i, pms := range data {
		md := metadata.Pairs(pms...)
		newCtx := metadata.NewIncomingContext(ctx, md)
		_, err := UnaryServerInterceptor()(newCtx, testRequest{}, nil, handler)
		if err != nil {
			t.Fatalf("data item %d, unexpected error : %v", i, err)
		}
	}

}

func TestUnaryServerInterceptorErrorCases(t *testing.T) {
	data := [][]string{
		[]string{gateway.LimitMetaKey, "3"},
		[]string{gateway.OffsetMetaKey, "3"},
		[]string{gateway.OffsetMetaKey, "3", gateway.LimitMetaKey, "3"},
	}
	ctx := DummyContextWithServerTransportStream()
	for i, pms := range data {
		md := metadata.Pairs(pms...)
		newCtx := metadata.NewIncomingContext(ctx, md)
		_, err := UnaryServerInterceptor()(newCtx, testRequest{}, nil, handler)
		if err == nil {
			t.Fatalf("expected error for data item %d", i)
		}
	}
}
