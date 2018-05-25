package server_test

import (
	"fmt"

	"golang.org/x/net/context"
)

type HelloServerImpl struct{}

func (HelloServerImpl) SayHello(ctx context.Context, req *HelloRequest) (*HelloResponse, error) {
	return &HelloResponse{Greeting: fmt.Sprintf("hello, %s!", req.GetName())}, nil
}
