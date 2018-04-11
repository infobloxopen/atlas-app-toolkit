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
			Authorizer{"", NewBuilder(), NewHandler()},
			func() pep.Client { return mockClient{} },
			nil,
		},
		{
			Authorizer{"", NewBuilder(), NewHandler()},
			func() pep.Client { return mockClient{deny: true} },
			ErrUnauthorized,
		},
		{
			Authorizer{"", NewBuilder(), NewHandler()},
			func() pep.Client { return mockClient{errOnConnect: true} },
			ErrInternal,
		},
		{
			Authorizer{"", NewBuilder(), NewHandler()},
			func() pep.Client { return mockClient{errOnValidate: true} },
			ErrInternal,
		},
		{
			Authorizer{"", mockBuilder{errOnBuild: true}, NewHandler()},
			func() pep.Client { return mockClient{} },
			ErrInternal,
		},
		{
			Authorizer{"", NewBuilder(), mockHandler{errOnHandle: true}},
			func() pep.Client { return mockClient{} },
			ErrInternal,
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
