package integration

import (
	"context"
	"fmt"
	"log"
	"strings"
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

func ExampleAppendTokenToOutgoingContext_output() {
	// make the jwt
	authToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256, jwt.MapClaims{
			"user":  "user-test",
			"roles": "admin",
		},
	).SignedString([]byte("some-secret"))
	if err != nil {
		log.Fatalf("unable to build token: %v", err)
	}
	// add the jwt to context
	ctxBearer := AppendTokenToOutgoingContext(
		context.Background(), "Bearer", authToken,
	)
	// check to make sure the token was added
	md, ok := metadata.FromOutgoingContext(ctxBearer)
	if !ok || len(md["authorization"]) < 1 {
		log.Fatalf("unable to get token from context: %v", err)
	}
	fields := strings.Split(md["authorization"][0], " ")
	if len(fields) < 2 {
		log.Fatalf("unexpected authorization metadata: %v", fields)
	}
	fmt.Println(fields[0] == "Bearer")
	fmt.Println(fields[1] == authToken)
	// Output: true
	// true
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
