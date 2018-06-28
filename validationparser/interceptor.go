package validator

import (
	"context"
	"reflect"
	"strings"

	"github.com/infobloxopen/atlas-app-toolkit/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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
				return nil, err
			}
		}
		return handler(ctx, req)
	}
}

// MapValidationError function returns a MapFunc that checks if the error is of type validation
func MapValidationError() errors.MapFunc {
	return errors.MapFunc(func(ctx context.Context, err error) (error, bool) {
		return ValidationHelper(err), true
	})
}

// ValidationHelper function returns a validation error from an error.
func ValidationHelper(err error) error {
	// Check if the error is of type of validation
	typeOfError := reflect.TypeOf(err)
	valueOfError := reflect.ValueOf(err)

	// Check if the error is a struct and a type of validation error that contains a cause.
	if valueOfError.Kind() == reflect.Struct && strings.Contains(typeOfError.String(), "ValidationError") && !valueOfError.FieldByName("Cause").IsNil() {
		cause := valueOfError.FieldByName("Cause")
		typeOfCause := reflect.TypeOf(cause.Interface())
		valueOfCause := reflect.ValueOf(cause.Interface())
		if strings.Contains(typeOfCause.String(), "ValidationError") && valueOfCause.FieldByName("Field").String() != "" && valueOfCause.FieldByName("Reason").String() != "" {
			// Retrieve the field and reason from the error
			field := valueOfCause.FieldByName("Field").String()
			reason := valueOfCause.FieldByName("Reason").String()
			return errors.NewContainer(codes.InvalidArgument, "Invalid %s: %s", field, reason).WithField(field, reason)
		}
	}
	return errors.NewContainer(codes.InvalidArgument, "Error : %s", err)
}
