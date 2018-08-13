package validationerrors

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"google.golang.org/grpc"

	"github.com/infobloxopen/atlas-app-toolkit/util"
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
	// Check if the error is of type of validation
	typeOfError := reflect.TypeOf(err)
	valueOfError := reflect.ValueOf(err)

	// Check if the error is a struct and a type of validation error that contains a cause.
	if valueOfError.Kind() == reflect.Struct && strings.Contains(typeOfError.String(), "ValidationError") &&
		!valueOfError.FieldByName("Cause").IsNil() {
		cause := valueOfError.FieldByName("Cause")
		typeOfCause := reflect.TypeOf(cause.Interface())

		valueOfCause := reflect.ValueOf(cause.Interface())
		if strings.Contains(typeOfCause.String(), "ValidationError") &&
			valueOfCause.FieldByName("Field").String() != "" &&
			valueOfCause.FieldByName("Reason").String() != "" &&
			!valueOfCause.FieldByName("Cause").IsNil() {
			// Retrieve the field and reason from the error
			field := valueOfCause.FieldByName("Field").String()
			field = util.CamelToSnake(field)
			reason := valueOfCause.FieldByName("Reason").String()
			key := valueOfCause.FieldByName("Key").Bool()
			causeValue := valueOfCause.FieldByName("Cause").Interface()
			if causeErr, ok := causeValue.(error); ok {
				return ValidationError{
					Field:         field,
					Reason:        reason,
					Key:           key,
					Cause:         causeErr,
					ErrorTypeName: typeOfCause.String(),
				}
			}
		}
	}
	// If it isn't a lyft validation error return the error
	return err
}

// ValidationError represents the validation error that contains which field failed and the reasoning behind it.
type ValidationError struct {
	Field         string
	Reason        string
	Key           bool
	Cause         error
	ErrorTypeName string // Error Name (“ABValidationError”)
}

// Error satisfies the builtin error interface.
func (e ValidationError) Error() string {
	cause := ""
	if e.Cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.Cause.Error())
	}

	key := ""
	if e.Key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %s%s.%s: %s%s",
		key,
		e.ErrorTypeName,
		e.Field,
		e.Reason,
		cause)
}

var _ error = ValidationError{}
