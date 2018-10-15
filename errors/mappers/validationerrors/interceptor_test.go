package validationerrors

import (
	"context"
	"fmt"
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
		actual   error
		expected error
	}{
		// Test cases
		{
			"ValidationErrorEmail",
			mocReqValidationError{
				field:  "Payload",
				reason: "embedded message failed validation",
				cause: mocReqValidationError{
					field:  "PrimaryEmail",
					reason: "value must be a valid email address",
					cause:  fmt.Errorf("mail: no angle-addr"),
					key:    true,
				},
				key: true,
			},
			&ValidationError{key: true, field: "primary_email", reason: "value must be a valid email address", cause: fmt.Errorf("mail: no angle-addr")},
		},
		{
			"ValidationErrorInt",
			mocReqValidationError{
				field:  "Payload",
				reason: "embedded message failed validation",
				cause: mocReqValidationError{
					field:  "Id",
					reason: "value must be greater than 50",
					cause:  fmt.Errorf("invalid Contact.Id"),
					key:    true,
				},
				key: true,
			},
			&ValidationError{key: true, field: "id", reason: "value must be greater than 50", cause: fmt.Errorf("invalid Contact.Id")},
		},
		{
			"ValidationErrorList",
			mocReqValidationError{
				field:  "Payload",
				reason: "embedded message failed validation",
				cause: mocReqValidationError{
					field:  "FirstName",
					reason: "value must not be in list [fizz buzz]",
					cause:  fmt.Errorf("invalid Contact.MiddleName"),
					key:    true,
				},
				key: true,
			},
			&ValidationError{key: true, field: "first_name", reason: "value must not be in list [fizz buzz]", cause: fmt.Errorf("invalid Contact.MiddleName")},
		},
		{
			"NotValidationError",
			mockNotValidation{
				Field:  "Other validator",
				Reason: "Not lyft validation",
				Key:    true,
			},
			fmt.Errorf("invalid key for Request.Other validator: Not lyft validation"),
		},
		{
			"ValidationErrorNoCause",
			mocReqValidationError{
				field:  "Payload",
				reason: "embedded message failed validation",
				key:    true,
			},
			fmt.Errorf("invalid key for ValidationError.Payload: embedded message failed validation"),
		},
		{
			"ValidationErrorBadCause",
			mocReqValidationError{
				field:  "Payload",
				reason: "embedded message failed validation",
				cause:  fmt.Errorf("Not validation"),
				key:    true,
			},
			fmt.Errorf("invalid key for ValidationError.Payload: embedded message failed validation | caused by: Not validation"),
		},
	}
	for _, tt := range tests {
		_, actual := UnaryServerInterceptor()(ctx, tt.actual, nil, nil)
		expected := tt.expected
		if actual.Error() != expected.Error() {
			t.Errorf("Error received was incorrect for test %s, expected: \"%s\", actual: \"%s\"", tt.name, expected, actual)
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

// mocReqValidationError is the validation error returned by
// mocReqValidationError.Validate if the designated constraints aren't met.
type mocReqValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e mocReqValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e mocReqValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e mocReqValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e mocReqValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e mocReqValidationError) ErrorName() string {
	return "ValidationError"
}

// Error satisfies the builtin error interface
func (e mocReqValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sValidationError.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = mocReqValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = mocReqValidationError{}

// Validate checks the field values on mocReqValidationError with the rules
// defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (e mocReqValidationError) Validate() error {
	return e
}
