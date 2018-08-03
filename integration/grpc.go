package integration

import (
	"context"
	"fmt"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/infobloxopen/atlas-app-toolkit/auth"
	"google.golang.org/grpc/metadata"
)

// AppendTokenToOutgoingContext adds an authorization token to the gRPC
// request context metadata. The user must provide a token field name like "token"
// or "bearer" to this function. It is intended specifically for gRPC testing.
func AppendTokenToOutgoingContext(ctx context.Context, fieldName, token string) context.Context {
	c := metadata.AppendToOutgoingContext(
		ctx, "Authorization", fmt.Sprintf("%s %s", fieldName, token),
	)
	return c
}

// StandardTestingContext returns an outgoing request context that includes the
// standard test JWT. It is intended specifically for gRPC testing.
func StandardTestingContext() (context.Context, error) {
	token, err := MakeTestJWT(jwt.SigningMethodHS256, StandardClaims)
	if err != nil {
		return nil, err
	}
	return AppendTokenToOutgoingContext(context.Background(), auth.DefaultTokenType, token), nil
}
