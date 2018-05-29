package requestid

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/transport"
	"testing"
)

type testRequest struct{}

type testResponse struct{}

func DummyContextWithServerTransportStream() context.Context {
	expectedStream := &transport.Stream{}
	return grpc.NewContextWithServerTransportStream(context.Background(), expectedStream)
}

func TestUnaryServerInterceptorWithoutRequestId(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		exists, reqID := FromContext(ctx)
		if exists && reqID == "" {
			t.Errorf("requestId must be generated by interceptor")
		}
		return &testResponse{}, nil
	}
	ctx := DummyContextWithServerTransportStream()
	_, err := UnaryServerInterceptor()(ctx, testRequest{}, nil, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnaryServerInterceptorWithDummyRequestId(t *testing.T) {
	dummyRequestId := newRequestId()
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		exists, reqID := FromContext(ctx)
		if exists && reqID != dummyRequestId {
			t.Errorf("expected requestID: %q, returned requestId: %q", dummyRequestId, reqID)
		}
		return &testResponse{}, nil
	}
	ctx := DummyContextWithServerTransportStream()
	md := metadata.Pairs(DefaultRequestIDKey, dummyRequestId)
	newCtx := metadata.NewIncomingContext(ctx, md)
	_, err := UnaryServerInterceptor()(newCtx, testRequest{}, nil, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnaryServerInterceptorWithEmptyRequestId(t *testing.T) {
	emptyRequestId := ""
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		exists, reqID := FromContext(ctx)
		if exists && reqID == "" {
			t.Errorf("requestId must be generated by interceptor")
		}
		return &testResponse{}, nil
	}
	ctx := DummyContextWithServerTransportStream()
	md := metadata.Pairs(DefaultRequestIDKey, emptyRequestId)
	newCtx := metadata.NewIncomingContext(ctx, md)
	_, err := UnaryServerInterceptor()(newCtx, testRequest{}, nil, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
