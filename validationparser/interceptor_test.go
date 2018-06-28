package validator

import (
	"context"
	"fmt"
	"testing"

	"github.com/infobloxopen/atlas-app-toolkit/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/transport"
)

// DummyContextWithServerTransportStream returns a dummy context for testing.
func DummyContextWithServerTransportStream() context.Context {
	expectedStream := &transport.Stream{}
	return grpc.NewContextWithServerTransportStream(context.Background(), expectedStream)
}

// TestUnaryServerInterceptor_ValidationErrors will run mock validation errors to see if it parses correctly.
func TestUnaryServerInterceptor_ValidationErrors(t *testing.T) {
	ctx := DummyContextWithServerTransportStream()
	tests := []struct {
		name     string
		actual   error
		expected error
	}{
		// Test cases
		{
			"ValidationErrorEmail",
			mockRequestValidationError{
				Field:  "Payload",
				Reason: "embedded message failed validation",
				Cause: mockValidationError{
					Field:  "PrimaryEmail",
					Reason: "value must be a valid email address",
					Cause:  fmt.Errorf("caused by: mail: no angle-addr"),
					Key:    true,
				},
				Key: true,
			},
			fmt.Errorf("Invalid %s: %s", "PrimaryEmail", "value must be a valid email address"),
		},
		{
			"ValidationErrorInt",
			mockRequestValidationError{
				Field:  "Payload",
				Reason: "embedded message failed validation",
				Cause: mockValidationError{
					Field:  "Id",
					Reason: "value must be greater than 50",
					Cause:  fmt.Errorf("caused by: invalid Contact.Id"),
					Key:    true,
				},
				Key: true,
			},
			fmt.Errorf("Invalid %s: %s", "Id", "value must be greater than 50"),
		},
		{
			"ValidationErrorList",
			mockRequestValidationError{
				Field:  "Payload",
				Reason: "embedded message failed validation",
				Cause: mockValidationError{
					Field:  "FirstName",
					Reason: "value must not be in list [fizz buzz]",
					Cause:  fmt.Errorf("caused by: invalid Contact.MiddleName"),
					Key:    true,
				},
				Key: true,
			},
			fmt.Errorf("Invalid %s: %s", "FirstName", "value must not be in list [fizz buzz]"),
		},
		{
			"NotValidationError",
			mockNotValidation{
				Field:  "Other validator",
				Reason: "Not lyft validation",
				Key:    true,
			},
			fmt.Errorf("Error : invalid key for Request.Other validator: Not lyft validation"),
		},
		{
			"ValidationErrorNoCause",
			mockRequestValidationError{
				Field:  "Payload",
				Reason: "embedded message failed validation",
				Key:    true,
			},
			fmt.Errorf("Error : invalid key for CreateRequest.Payload: embedded message failed validation"),
		},
		{
			"ValidationErrorBadCause",
			mockRequestValidationError{
				Field:  "Payload",
				Reason: "embedded message failed validation",
				Cause:  fmt.Errorf("Not validation"),
				Key:    true,
			},
			fmt.Errorf("Error : invalid key for CreateRequest.Payload: embedded message failed validation | caused by: Not validation"),
		},
		{
			"ValidationErrorBadField",
			mockRequestValidationError{
				Field:  "Payload",
				Reason: "embedded message failed validation",
				Cause: mockValidationError{
					Reason: "no field",
					Cause:  fmt.Errorf("bad test"),
					Key:    true,
				},
				Key: true,
			},
			fmt.Errorf("Error : invalid key for CreateRequest.Payload: embedded message failed validation | caused by: invalid key for CreateRequest.: no field | caused by: bad test"),
		},
		{
			"ValidationErrorBadReason",
			mockRequestValidationError{
				Field:  "Payload",
				Reason: "embedded message failed validation",
				Cause: mockValidationError{
					Field: "testField",
					Cause: fmt.Errorf("bad test"),
					Key:   true,
				},
				Key: true,
			},
			fmt.Errorf("Error : invalid key for CreateRequest.Payload: embedded message failed validation | caused by: invalid key for CreateRequest.testField:  | caused by: bad test"),
		},
	}
	for _, tt := range tests {
		_, err := UnaryServerInterceptor()(ctx, tt.actual, nil, nil)
		actual := ValidationHelper(err)
		expected := tt.expected
		if errCasted, ok := actual.(*errors.Container); ok {
			actualMessage := errCasted.GRPCStatus().Message()
			if actualMessage != expected.Error() {
				t.Errorf("Error received was incorrect for test %s, expected: \"%s\", actual: \"%s\"", tt.name, expected, actualMessage)
			}
		}
	}
}

// TestUnaryServerInterceptor_OtherError will run a regular error to see if it ignores it.
func TestUnaryServerInterceptor_OtherError(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return &testResponse{}, status.Newf(codes.Internal, "Internal error").Err()
	}
	ctx := DummyContextWithServerTransportStream()
	_, actual := UnaryServerInterceptor()(ctx, nil, nil, handler)
	expected := status.Newf(codes.Internal, "Internal error")

	// verify that the errors match
	if actual.Error() != expected.Err().Error() {
		t.Errorf("Error received was incorrect, expected: \"%s\", actual: \"%s\"", expected.Err(), actual.Error())
	}
}

// TestUnaryServerInterceptor_Success will run no errors.
func TestUnaryServerInterceptor_Success(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return &testResponse{}, nil
	}
	ctx := DummyContextWithServerTransportStream()
	_, actual := UnaryServerInterceptor()(ctx, nil, nil, handler)

	// verify that no errors get returned
	if actual != nil {
		t.Errorf("expected no error, but got: %v", actual.Error())
	}
}

// testResponse represents a mock response.
type testResponse struct{}

// mockRequestValidationError represents a validation request error.
type mockRequestValidationError struct {
	Field  string
	Reason string
	Cause  error
	Key    bool
}

// Error satisfies the builtin error interface
func (e mockRequestValidationError) Error() string {
	cause := ""
	if e.Cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.Cause)
	}

	key := ""
	if e.Key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sCreateRequest.%s: %s%s",
		key,
		e.Field,
		e.Reason,
		cause)
}

// Error satisfies the builtin error interface
func (e mockRequestValidationError) Validate() error {
	return e
}

var _ error = mockRequestValidationError{}

// mockValidationError represents a validation error.
type mockValidationError struct {
	Field  string
	Reason string
	Cause  error
	Key    bool
}

// Error satisfies the builtin error interface
func (e mockValidationError) Error() string {
	cause := ""
	if e.Cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.Cause)
	}

	key := ""
	if e.Key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sCreateRequest.%s: %s%s",
		key,
		e.Field,
		e.Reason,
		cause)
}

var _ error = mockValidationError{}

// mockNotValidation represents anoter validate error but not validation error.
type mockNotValidation struct {
	Field  string
	Reason string
	Cause  error
	Key    bool
}

// Error satisfies the builtin error interface
func (e mockNotValidation) Validate() error {
	return e
}

// Error satisfies the builtin error interface
func (e mockNotValidation) Error() string {
	cause := ""
	if e.Cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.Cause)
	}

	key := ""
	if e.Key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sRequest.%s: %s%s",
		key,
		e.Field,
		e.Reason,
		cause)
}

var _ error = mockNotValidation{}
