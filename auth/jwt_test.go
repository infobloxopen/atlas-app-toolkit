package auth

import (
	"context"
	"fmt"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc/metadata"
)

const (
	// TestSecret dummy secret used for signing test JWTs
	TestSecret = "some-secret-123"
)

func TestGetJWTFieldWithTokenType(t *testing.T) {
	var jwtFieldTests = []struct {
		claims         jwt.MapClaims
		contextFactory func(t *testing.T) context.Context
		tokenType      string
		field          string
		expected       string
		err            error
	}{
		{
			contextFactory: func(t *testing.T) context.Context {
				token := makeToken(jwt.MapClaims{"some-field": "id-abc-123"}, t)
				ctx, err := contextWithToken(token, "Bearer")
				if err != nil {
					t.Fatalf("Error when building request context: %v", err)
				}
				return ctx
			},
			tokenType: "Bearer",
			field:     "some-field",
			expected:  "id-abc-123",
			err:       nil,
		},
		{
			contextFactory: func(t *testing.T) context.Context {
				token := makeToken(jwt.MapClaims{"some-field": "id-abc-123"}, t)
				ctx, err := contextWithToken(token, "Bearer")
				if err != nil {
					t.Fatalf("Error when building request context: %v", err)
				}
				return ctx
			},
			tokenType: "Bearer",
			field:     "some-other-field",
			expected:  "",
			err:       errMissingField,
		},
		{
			contextFactory: func(t *testing.T) context.Context {
				token := makeToken(jwt.MapClaims{"some-field": "id-abc-123"}, t)
				ctx, err := contextWithToken(token, "Bearer")
				if err != nil {
					t.Fatalf("Error when building request context: %v", err)
				}
				return ctx
			},
			tokenType: "token",
			field:     "some-field",
			expected:  "",
			err:       errMissingToken,
		},
		{
			contextFactory: func(t *testing.T) context.Context {
				return nil
			},
			tokenType: "Bearer",
			field:     "some-field",
			expected:  "",
			err:       errMissingToken,
		},
	}
	for _, test := range jwtFieldTests {
		ctx := test.contextFactory(t)

		actual, err := GetJWTFieldWithTokenType(ctx, test.tokenType, test.field, nil)

		if err != test.err {
			t.Errorf("Invalid error value: %v - expected %v", err, test.err)
		}
		if actual != test.expected {
			t.Errorf("Invalid JWT field: %v - expected %v", actual, test.expected)
		}
	}
}

func TestGetJWTField(t *testing.T) {
	var jwtFieldTests = []struct {
		claims   jwt.MapClaims
		field    string
		expected string
		err      error
	}{
		{
			claims:   jwt.MapClaims{"some-field": "id-abc-123"},
			field:    "some-field",
			expected: "id-abc-123",
			err:      nil,
		},
		{
			claims:   jwt.MapClaims{"some-field": "id-abc-123"},
			field:    "some-other-field",
			expected: "",
			err:      errMissingField,
		},
	}
	for _, test := range jwtFieldTests {
		ctx, err := contextWithToken(
			makeToken(test.claims, t), DefaultTokenType,
		)
		if err != nil {
			t.Fatalf("Error when building request context: %v", err)
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

func TestGetAccountID(t *testing.T) {
	var accountIDTests = []struct {
		claims   jwt.MapClaims
		expected string
		err      error
	}{
		{
			claims: jwt.MapClaims{
				MultiTenancyField: "id-abc-123",
			},
			expected: "id-abc-123",
			err:      nil,
		},
		{
			claims: jwt.MapClaims{
				"AccountID": "id-abc-123",
			},
			expected: "id-abc-123",
			err:      nil,
		},
		{
			claims:   jwt.MapClaims{},
			expected: "",
			err:      errMissingField,
		},
	}
	for _, test := range accountIDTests {
		token := makeToken(test.claims, t)
		ctx, err := contextWithToken(token, DefaultTokenType)
		if err != nil {
			t.Fatalf("Error when building request context: %v", err)
		}
		actual, err := GetAccountID(ctx, nil)
		if err != test.err {
			t.Errorf("Invalid error value: %v - expected %v", err, test.err)
		}
		if actual != test.expected {
			t.Errorf("Invalid AccountID: %v - expected %v", actual, test.expected)
		}
	}
}

// creates a context with a jwt
func contextWithToken(token, tokenType string) (context.Context, error) {
	md := metadata.Pairs(
		"authorization", fmt.Sprintf("%s %s", tokenType, token),
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
	signature, err := method.Sign(signingString, []byte(TestSecret))
	if err != nil {
		t.Fatalf("Error when building token: %v", err)
	}
	return fmt.Sprintf("%s.%s", signingString, signature)
}
