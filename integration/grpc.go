package integration

import (
	"context"
	"fmt"

	jwt "github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc/metadata"
)

// AppendTokenToOutgoingContext adds an authorization token to the gRPC
// request context metadata. The user must provide a token field name like "token"
// or "bearer"
func AppendTokenToOutgoingContext(ctx context.Context, fieldName, token string) context.Context {
	c := metadata.AppendToOutgoingContext(
		ctx, "Authorization", fmt.Sprintf("%s %s", fieldName, token),
	)
	return c
}

// DefaultContext returns a context that has a jwt for basic testing purposes
func StandardTestingContext() (context.Context, error) {
	token, err := MakeTestJWT(jwt.SigningMethodHS256, StandardClaims)
	if err != nil {
		return nil, err
	}
	return AppendTokenToOutgoingContext(context.Background(), "token", token), nil
}
