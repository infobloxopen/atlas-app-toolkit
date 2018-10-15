package validationerrors

import (
	"context"
	"fmt"

	"github.com/infobloxopen/atlas-app-toolkit/util"
	"google.golang.org/grpc"
)

type validator interface {
	Validate() error
}

// UnaryServerInterceptor returns a new unary server interceptor that validates incoming messages.
//
// Invalid messages will be rejected with `InvalidArgument` before reaching any userspace handlers.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if v, ok := req.(validator); ok {
			if err := v.Validate(); err != nil {
				// Fill ValidationError and throw it forward
				return nil, GetValidationError(err)
			}
		}
		return handler(ctx, req)
	}
}

// GetValidationError function returns a validation error from an error.
func GetValidationError(err error) error {
	// Check if the error is type of validation
	if vErr, ok := err.(RequestValidationError); ok {
		if causeErr, ok := vErr.Cause().(RequestValidationError); ok {
			return ValidationError{
				field:  util.CamelToSnake(causeErr.Field()),
				reason: causeErr.Reason(),
				key:    causeErr.Key(),
				cause:  causeErr.Cause(),
			}
		}
	}
	return err
}

// ValidationError represents the validation error that contains which field failed and the reasoning behind it.
type ValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ValidationError) ErrorName() string {
	return "ValidationError"
}

// Error satisfies the builtin error interface
func (e ValidationError) Error() string {
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

var _ error = ValidationError{}

// RequestValidationError represent a validation error
type RequestValidationError interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
}

var _ RequestValidationError = ValidationError{}
