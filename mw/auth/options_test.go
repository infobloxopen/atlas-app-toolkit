package auth

import (
	"context"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
	pdp "github.com/infobloxopen/themis/pdp-service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestWithJWT(t *testing.T) {
	var jwtTests = []struct {
		token    string
		expected []*pdp.Attribute
		keyfunc  jwt.Keyfunc
		err      error
	}{
		// parse and verify a valid token
		{
			token: makeToken(jwt.MapClaims{
				"username":   "john",
				"department": "engineering",
			}, t),
			expected: []*pdp.Attribute{
				&pdp.Attribute{Id: "department", Type: "string", Value: "engineering"},
				&pdp.Attribute{Id: "username", Type: "string", Value: "john"},
			},
			keyfunc: func(token *jwt.Token) (interface{}, error) {
				return []byte(TestSecret), nil
			},
			err: nil,
		},
		// parse and verify an invalid token
		{
			token: makeToken(jwt.MapClaims{
				"username":   "john",
				"department": "engineering",
			}, t),
			expected: []*pdp.Attribute{},
			keyfunc: func(token *jwt.Token) (interface{}, error) {
				return []byte("some-other-secret-123"), nil
			},
			err: ErrUnauthorized,
		},
		// parse a valid token, but do not verify
		{
			token:    makeToken(jwt.MapClaims{}, t),
			expected: []*pdp.Attribute{},
			keyfunc:  nil,
			err:      nil,
		},
		// parse an invalid token, but do not verify
		{
			token:    "some-nonsense-token",
			expected: []*pdp.Attribute{},
			keyfunc:  nil,
			err:      ErrUnauthorized,
		},
		// do not include a token in the request context
		{
			token:    "",
			expected: []*pdp.Attribute{},
			keyfunc:  nil,
			err:      ErrUnauthorized,
		},
	}
	for _, test := range jwtTests {
		ctx := context.Background()
		if test.token != "" {
			c, _ := contextWithToken(test.token)
			ctx = c
		}
		builder := NewBuilder(WithJWT(test.keyfunc))
		req, err := builder.build(ctx)
		if err != test.err {
			t.Errorf("Unexpected error when building request: %v", err)
		}
		if !hasMatchingAttributes(req.Attributes, test.expected) {
			t.Errorf("Invalid request attributes: %v - expected %v", req.GetAttributes(), test.expected)
		}
	}
}

func TestWithCallback(t *testing.T) {
	var callbackTests = []struct {
		callback attributer
		expected []*pdp.Attribute
	}{
		{
			callback: func(ctx context.Context) ([]*pdp.Attribute, error) {
				attributes := []*pdp.Attribute{
					&pdp.Attribute{Id: "fruit", Type: "string", Value: "apple"},
					&pdp.Attribute{Id: "vegetable", Type: "string", Value: "carrot"},
				}
				return attributes, nil
			},
			expected: []*pdp.Attribute{
				{Id: "fruit", Type: "string", Value: "apple"},
				{Id: "vegetable", Type: "string", Value: "carrot"},
			},
		},
		{
			callback: func(ctx context.Context) ([]*pdp.Attribute, error) {
				return []*pdp.Attribute{}, nil
			},
			expected: []*pdp.Attribute{},
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

type mockTransportStream struct{ method string }

func (m mockTransportStream) Method() string               { return m.method }
func (m mockTransportStream) SetHeader(metadata.MD) error  { return nil }
func (m mockTransportStream) SendHeader(metadata.MD) error { return nil }
func (m mockTransportStream) SetTrailer(metadata.MD) error { return nil }

func TestWithRequest(t *testing.T) {
	var tests = []struct {
		stream   *mockTransportStream
		expected []*pdp.Attribute
		err      error
	}{
		{
			stream: &mockTransportStream{method: "/PetStore/ListPets"},
			expected: []*pdp.Attribute{
				{Id: "operation", Type: "string", Value: "ListPets"},
				{Id: "application", Type: "string", Value: "petstore"},
			},
			err: nil,
		},
		{
			stream: &mockTransportStream{method: "/atlas.example.PetStore/ListPets"},
			expected: []*pdp.Attribute{
				{Id: "operation", Type: "string", Value: "ListPets"},
				{Id: "application", Type: "string", Value: "petstore"},
			},
			err: nil,
		},
		{
			stream:   nil,
			expected: []*pdp.Attribute{},
			err:      ErrInternal,
		},
	}
	for _, test := range tests {
		ctx := context.Background()
		if test.stream != nil {
			ctx = grpc.NewContextWithServerTransportStream(
				context.Background(),
				test.stream,
			)
		}
		builder := NewBuilder(WithRequest())
		req, err := builder.build(ctx)
		if err != test.err {
			t.Errorf("Unexpected error when building request: %v", err)
		}
		if !hasMatchingAttributes(req.Attributes, test.expected) {
			t.Errorf("Invalid request attributes: %v - expected %v", req.GetAttributes(), test.expected)
		}
	}
}

func TestStripPackageName(t *testing.T) {
	var tests = []struct {
		fullname string
		expected string
	}{
		{fullname: "ngp.api.toolkit.example.addressbook.AddressBook", expected: "AddressBook"},
		{fullname: "AddressBook", expected: "AddressBook"},
		{fullname: "", expected: ""},
	}
	for _, test := range tests {
		name := stripPackageName(test.fullname)
		if name != test.expected {
			t.Errorf("Unexpected service name: %s - expected %s", name, test.expected)
		}
	}
}

// checks if first and second attribute lists contain identical elements
func hasMatchingAttributes(first, second []*pdp.Attribute) bool {
	if len(first) != len(second) {
		return false
	}
	for _, attrFirst := range first {
		var hasAttribute bool
		for _, attrSecond := range second {
			hasAttribute = hasAttribute || attrFirst.String() == attrSecond.String()
		}
		if !hasAttribute {
			return false
		}
	}
	return true
}
