package auth

import (
	"context"
	"fmt"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc/metadata"
)

const (
	TEST_SECRET = "some-secret-123"
)

func TestGetJWTField(t *testing.T) {
	var jwtFieldTests = []struct {
		claims   jwt.MapClaims
		field    string
		expected string
		err      error
	}{
		{
			claims: jwt.MapClaims{
				"some-field": "id-abc-123",
			},
			field:    "some-field",
			expected: "id-abc-123",
			err:      nil,
		},
		{
			claims: jwt.MapClaims{
				"some-field": "id-abc-123",
			},
			field:    "some-other-field",
			expected: "",
			err:      errMissingField,
		},
		{
			claims:   jwt.MapClaims{},
			field:    "some-field",
			expected: "",
			err:      errMissingToken,
		},
	}
	for _, test := range jwtFieldTests {
		ctx := context.Background()
		if len(test.claims) != 0 {
			token := makeToken(test.claims, t)
			ctx, _ = contextWithToken(token)
		}
		actual, err := GetJWTField(ctx, test.field, nil)
		if err != test.err {
			t.Errorf("Invalid error value: %v - expected %v", err, test.err)
		}
		if actual != test.expected {
			t.Errorf("Invalid JWT field: %v - expected %v", actual, test.expected)
		}
	}
}

// creates a context with a jwt
func contextWithToken(token string) (context.Context, error) {
	md := metadata.Pairs(
		"authorization", fmt.Sprintf("token %s", token),
	)
	return metadata.NewIncomingContext(context.Background(), md), nil
}

// generates a token string based on the given jwt claims
func makeToken(claims jwt.Claims, t *testing.T) string {
	method := jwt.SigningMethodHS256
	token := jwt.NewWithClaims(method, claims)
	signingString, err := token.SigningString()
	if err != nil {
		t.Fatalf("Error when building token: %v", err)
	}
	signature, err := method.Sign(signingString, []byte(TEST_SECRET))
	if err != nil {
		t.Fatalf("Error when building token: %v", err)
	}
	return fmt.Sprintf("%s.%s", signingString, signature)
}
