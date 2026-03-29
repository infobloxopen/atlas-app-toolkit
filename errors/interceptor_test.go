package errors

import (
	"context"
	"fmt"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestInterceptor(t *testing.T) {
	t.Run("handler success passes through", func(t *testing.T) {
		interceptor := UnaryServerInterceptor()
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "ok", nil
		}
		res, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if res != "ok" {
			t.Errorf("expected result %q, got %v", "ok", res)
		}
	})

	t.Run("handler returns Container error as-is", func(t *testing.T) {
		interceptor := UnaryServerInterceptor()
		expected := NewContainer(codes.InvalidArgument, "bad input")
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, expected
		}
		res, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)
		if res != nil {
			t.Errorf("expected nil result, got %v", res)
		}
		if err != expected {
			t.Errorf("expected Container error to pass through, got %v", err)
		}
	})

	t.Run("handler returns grpc status error as-is", func(t *testing.T) {
		interceptor := UnaryServerInterceptor()
		grpcErr := status.Error(codes.NotFound, "not found")
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, grpcErr
		}
		res, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)
		if res != nil {
			t.Errorf("expected nil result, got %v", res)
		}
		if err != grpcErr {
			t.Errorf("expected grpc status error to pass through, got %v", err)
		}
	})

	t.Run("handler error mapped by MapFunc", func(t *testing.T) {
		mapped := NewContainer(codes.Internal, "mapped error")
		mapFunc := NewMapping(
			CondEq("original error"),
			mapped,
		)
		interceptor := UnaryServerInterceptor(mapFunc)
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, fmt.Errorf("original error")
		}
		res, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)
		if res != nil {
			t.Errorf("expected nil result, got %v", res)
		}
		if err != mapped {
			t.Errorf("expected mapped error, got %v", err)
		}
	})

	t.Run("handler error with no matching map returns InitContainer", func(t *testing.T) {
		interceptor := UnaryServerInterceptor()
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, fmt.Errorf("unmapped error")
		}
		_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		c, ok := err.(*Container)
		if !ok {
			t.Fatalf("expected *Container, got %T", err)
		}
		if c.errCode != codes.Unknown {
			t.Errorf("expected code %v, got %v", codes.Unknown, c.errCode)
		}
	})

	t.Run("context contains container for handler use", func(t *testing.T) {
		interceptor := UnaryServerInterceptor()
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			c := FromContext(ctx)
			if c == nil {
				t.Error("expected container in context")
			}
			return "ok", nil
		}
		_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}
