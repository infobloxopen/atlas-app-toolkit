package integration

import (
	"context"
	"fmt"
	"log"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc/metadata"
)

func TestAppendTokenToOutgoingContext(t *testing.T) {
	token, err := MakeTestJWT(
		jwt.SigningMethodHS256, jwt.MapClaims{
			"test-key": "test-value",
		},
	)
	if err != nil {
		t.Fatalf("unable to make test token: %v", err)
	}
	ctx := AppendTokenToOutgoingContext(context.Background(), "Bearer", token)
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		t.Fatal("unable to get metadata from context")
	}
	if md["authorization"][0] != fmt.Sprintf("Bearer %s", token) {
		t.Fatalf("context does not contain token in metadata")
	}
}

func ExampleAppendTokenToOutgoingContext() {
	// someFunc doesn't do anything, but in a real-world situation it might
	// send a request to some grpc service that requires an authorization
	// token
	someFunc := func(context.Context) {
		// send a request to some grpc service
	}
	authToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256, jwt.MapClaims{
			"user":  "user-test",
			"roles": "admin",
		},
	).SignedString([]byte("some-secret"))
	if err != nil {
		log.Fatalf("unable to build token: %v", err)
	}
	ctxBearer := AppendTokenToOutgoingContext(
		context.Background(), authToken, "Bearer",
	)
	someFunc(ctxBearer)
}

func TestStandardTestingContext(t *testing.T) {
	ctx, err := StandardTestingContext()
	if err != nil {
		t.Fatalf("unexpected error when getting standard context: %v", err)
	}
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		t.Fatal("unable to get metadata from context")
	}
	if md["authorization"][0] != fmt.Sprintf("token %s", standardToken) {
		t.Fatalf("context does not contain token in metadata")
	}
}
