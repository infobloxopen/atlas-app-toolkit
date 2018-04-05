package auth

import (
	"context"
	"fmt"
	"testing"

	"google.golang.org/grpc/metadata"

	jwt "github.com/dgrijalva/jwt-go"
	pdp "github.com/infobloxopen/themis/pdp-service"
	"github.com/infobloxopen/themis/pep"
)

func TestWithJWT(t *testing.T) {
	var jwtTests = []struct {
		claims   jwt.MapClaims
		expected []*pdp.Attribute
	}{
		{
			jwt.MapClaims{
				"username":   "john",
				"department": "engineering",
			},
			[]*pdp.Attribute{
				&pdp.Attribute{"department", "string", "engineering"},
				&pdp.Attribute{"username", "string", "john"},
			},
		},
		{
			jwt.MapClaims{},
			[]*pdp.Attribute{},
		},
	}
	for _, test := range jwtTests {
		ctx, err := contextWithToken(test.claims)
		if err != nil {
			t.Errorf("Unexpected error when building request context: %v", err)
		}
		builder := NewBuilder(WithJWT())
		req, err := builder.build(ctx)
		if err != nil {
			t.Errorf("Unexpected error when building themis request: %v", err)
		}
		if !hasMatchingAttributes(req.Attributes, test.expected) {
			t.Errorf("Invalid request attributes: %v - expected %v", req.GetAttributes(), test.expected)
		}
	}
}

func TestWithRules(t *testing.T) {
	// needs to get merged: https://github.com/grpc/grpc-go/pull/1904
}

func TestWithCallback(t *testing.T) {
	var callbackTests = []struct {
		callback attributer
		expected []*pdp.Attribute
	}{
		{
			func(ctx context.Context) ([]*pdp.Attribute, error) {
				attributes := []*pdp.Attribute{
					&pdp.Attribute{"fruit", "string", "apple"},
					&pdp.Attribute{"vegetable", "string", "carrot"},
				}
				return attributes, nil
			},
			[]*pdp.Attribute{
				{"fruit", "string", "apple"},
				{"vegetable", "string", "carrot"},
			},
		},
		{
			func(ctx context.Context) ([]*pdp.Attribute, error) {
				return []*pdp.Attribute{}, nil
			},
			[]*pdp.Attribute{},
		},
	}
	for _, test := range callbackTests {
		builder := NewBuilder(WithCallback(test.callback))
		req, err := builder.build(context.Background())
		if err != nil {
			t.Errorf("Unexpected error when building request: %v", err)
		}
		if !hasMatchingAttributes(req.Attributes, test.expected) {
			t.Errorf("Invalid request attributes: %v - expected %v", req.GetAttributes(), test.expected)
		}
	}
}

// creates a context with a jwt
func contextWithToken(claims jwt.Claims) (context.Context, error) {
	token, err := makeToken(claims)
	if err != nil {
		return nil, err
	}
	md := metadata.Pairs(
		"authorization", fmt.Sprintf("token %s", token),
	)
	return metadata.NewIncomingContext(context.Background(), md), nil
}

// generates a token string based on the given jwt claims
func makeToken(claims jwt.Claims) (string, error) {
	method := jwt.SigningMethodHS256
	token := jwt.NewWithClaims(method, claims)
	signingString, err := token.SigningString()
	if err != nil {
		return "", err
	}
	signature, err := method.Sign(signingString, []byte("secret"))
	return fmt.Sprintf("%s.%s", signingString, signature), nil
}

// checks if first and second attribute lists contain identical elements
func hasMatchingAttributes(first, second []*pdp.Attribute) bool {
	if len(first) != len(second) {
		return false
	}
	for _, attr_first := range first {
		var hasAttribute bool
		for _, attr_second := range second {
			hasAttribute = hasAttribute || attr_first.String() == attr_second.String()
		}
		if !hasAttribute {
			return false
		}
	}
	return true
}

func TestAuthFunc(t *testing.T) {
	var authFuncTests = []struct {
		authorizer Authorizer
		client     mockClient
		expected   error
	}{
		{
			Authorizer{"", NewBuilder(), NewHandler()},
			mockClient{permitAll: true, withBadConnection: false},
			nil,
		},
		{
			Authorizer{"", NewBuilder(), NewHandler()},
			mockClient{permitAll: false, withBadConnection: false},
			ErrUnauthorized,
		},
		{
			Authorizer{"", NewBuilder(), NewHandler()},
			mockClient{permitAll: true, withBadConnection: true},
			ErrInternal,
		},
	}
	for _, test := range authFuncTests {
		authFunc := test.authorizer.authFunc(test.client)
		_, err := authFunc(context.Background())
		if test.expected != err {
			t.Errorf("Invalid authfunc error: %v - expected %v", err, test.expected)
		}
	}
}

type mockClient struct {
	permitAll         bool
	withBadConnection bool
}

func (m mockClient) Connect(string) error {
	if m.withBadConnection {
		return pep.ErrorConnected
	}
	return nil
}

func (m mockClient) Close() {}

func (m mockClient) Validate(in, out interface{}) error {
	o, ok := out.(*pdp.Response)
	if !ok {
		return pep.ErrorInvalidStruct
	}
	if m.permitAll {
		o.Effect = pdp.Response_PERMIT
	} else {
		o.Effect = pdp.Response_DENY
	}
	return nil
}
