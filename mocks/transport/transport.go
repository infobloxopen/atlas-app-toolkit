package transport

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type mockServerTransportStream struct {
	ctx context.Context
}

func (*mockServerTransportStream) Method() string {
	return "unimplemented"
}

func (*mockServerTransportStream) SetHeader(metadata.MD) error {
	return nil
}

func (*mockServerTransportStream) SendHeader(metadata.MD) error {
	return nil
}

func (*mockServerTransportStream) SetTrailer(metadata.MD) error { return nil }

func (m *mockServerTransportStream) Context() context.Context {
	return m.ctx
}

func (*mockServerTransportStream) SendMsg(m interface{}) error {
	return nil
}

func (*mockServerTransportStream) RecvMsg(m interface{}) error {
	return nil
}

type MockServerStream struct {
	ctx context.Context
}

func (*MockServerStream) Method() string {
	return "unimplemented"
}

func (*MockServerStream) SetHeader(metadata.MD) error {
	return nil
}

func (*MockServerStream) SendHeader(metadata.MD) error {
	return nil
}

func (*MockServerStream) SetTrailer(metadata.MD) {}

func (m *MockServerStream) Context() context.Context {
	return m.ctx
}

func (*MockServerStream) SendMsg(m interface{}) error {
	return nil
}

func (*MockServerStream) RecvMsg(m interface{}) error {
	return nil
}

func NewMockServerStream(ctx context.Context) *MockServerStream {
	return &MockServerStream{ctx}
}

func DummyContextWithServerTransportStream() context.Context {
	expectedStream := &mockServerTransportStream{}
	return grpc.NewContextWithServerTransportStream(context.Background(), expectedStream)
}
