package validationparser

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

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
		want     mockRequestValidationError
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
					Cause:  errors.New("caused by: mail: no angle-addr"),
					Key:    true,
				},
				Key: true,
			},
			status.Newf(codes.InvalidArgument, "PrimaryEmail value must be a valid email address").Err(),
		},
		{
			"ValidationErrorInt",
			mockRequestValidationError{
				Field:  "Payload",
				Reason: "embedded message failed validation",
				Cause: mockValidationError{
					Field:  "Id",
					Reason: "value must be greater than 50",
					Cause:  errors.New("caused by: invalid Contact.Id"),
					Key:    true,
				},
				Key: true,
			},
			status.Newf(codes.InvalidArgument, "Id value must be greater than 50").Err(),
		},
		{
			"ValidationErrorList",
			mockRequestValidationError{
				Field:  "Payload",
				Reason: "embedded message failed validation",
				Cause: mockValidationError{
					Field:  "FirstName",
					Reason: "value must not be in list [fizz buzz]",
					Cause:  errors.New("caused by: invalid Contact.MiddleName"),
					Key:    true,
				},
				Key: true,
			},
			status.Newf(codes.InvalidArgument, "FirstName value must not be in list [fizz buzz]").Err(),
		},
		{
			"ValidationErrorBad",
			mockRequestValidationError{
				Field:  "Payload",
				Reason: "embedded message failed validation",
				Cause:  errors.New("Bad test"),
				Key:    true,
			},
			status.Newf(codes.InvalidArgument, "invalid key for CreateRequest.Payload: embedded message failed validation | caused by: Bad test").Err(),
		},
	}
	for _, tt := range tests {
		_, actual := UnaryServerInterceptor()(ctx, tt.want, nil, nil)
		expected := tt.expected
		// verify that the errors match
		if actual.Error() != expected.Error() {
			t.Errorf("Error received was incorrect, expected: \"%s\", actual: \"%s\"", expected.Error(), actual.Error())
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

func TestUnaryServerInterceptor(t *testing.T) {
	tests := []struct {
		name string
		want grpc.UnaryServerInterceptor
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UnaryServerInterceptor(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UnaryServerInterceptor() = %v, want %v", got, tt.want)
			}
		})
	}
}
