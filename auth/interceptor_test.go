package auth

import (
	"context"
	"errors"
	"testing"

	pdp "github.com/infobloxopen/themis/pdp-service"
	"github.com/infobloxopen/themis/pep"
)

func TestAuthFunc(t *testing.T) {
	var authFuncTests = []struct {
		authorizer Authorizer
		factory    func() pep.Client
		expected   error
	}{
		{
			authorizer: Authorizer{"", NewBuilder(), NewHandler()},
			factory:    func() pep.Client { return mockClient{} },
			expected:   nil,
		},
		{
			authorizer: Authorizer{"", NewBuilder(), NewHandler()},
			factory:    func() pep.Client { return mockClient{deny: true} },
			expected:   ErrUnauthorized,
		},
		{
			authorizer: Authorizer{"", NewBuilder(), NewHandler()},
			factory:    func() pep.Client { return mockClient{errOnConnect: true} },
			expected:   ErrInternal,
		},
		{
			authorizer: Authorizer{"", NewBuilder(), NewHandler()},
			factory:    func() pep.Client { return mockClient{errOnValidate: true} },
			expected:   ErrInternal,
		},
		{
			authorizer: Authorizer{"", mockBuilder{errOnBuild: true}, NewHandler()},
			factory:    func() pep.Client { return mockClient{} },
			expected:   ErrInternal,
		},
		{
			authorizer: Authorizer{"", NewBuilder(), mockHandler{errOnHandle: true}},
			factory:    func() pep.Client { return mockClient{} },
			expected:   ErrInternal,
		},
	}
	for _, test := range authFuncTests {
		authFunc := test.authorizer.authFunc(test.factory)
		_, err := authFunc(context.Background())
		if test.expected != err {
			t.Errorf("Invalid authfunc error: %v - expected %v", err, test.expected)
		}
	}
}

type mockBuilder struct {
	errOnBuild bool
}

func (m mockBuilder) build(context.Context) (pdp.Request, error) {
	if m.errOnBuild {
		return pdp.Request{}, ErrInternal
	}
	return pdp.Request{}, nil
}

type mockHandler struct {
	errOnHandle bool
}

func (m mockHandler) handle(context.Context, pdp.Response) (bool, error) {
	if m.errOnHandle {
		return false, ErrInternal
	}
	return false, nil
}

type mockClient struct {
	errOnConnect  bool
	errOnValidate bool
	deny          bool
}

func (m mockClient) Connect(string) error {
	if m.errOnConnect {
		return pep.ErrorConnected
	}
	return nil
}

func (m mockClient) Close() {}

func (m mockClient) Validate(in, out interface{}) error {
	if m.errOnValidate {
		return errors.New("Unable to validate request")
	}
	o, ok := out.(*pdp.Response)
	if !ok {
		return pep.ErrorInvalidStruct
	}
	if !m.deny {
		o.Effect = pdp.Response_PERMIT
	}
	return nil
}
