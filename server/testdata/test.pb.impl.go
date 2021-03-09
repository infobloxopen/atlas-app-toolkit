package testdata

import (
	"fmt"

	"golang.org/x/net/context"
)

type HelloServerImpl struct{}

func (HelloServerImpl) SayHello(ctx context.Context, req *SayHelloRequest) (*SayHelloResponse, error) {
	return &SayHelloResponse{Greeting: fmt.Sprintf("hello, %s!", req.GetName())}, nil
}
