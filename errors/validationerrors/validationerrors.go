package validationerrors

import (
	"context"

	"github.com/infobloxopen/atlas-app-toolkit/errors"
	"google.golang.org/grpc/codes"
)

// ToMapFunc function converts mapping function for *validationerrors.Error to a conventional
// MapFunc from atlas-app-toolkit/errors package.
func ToMapFunc(f func(context.Context, ValidationError) (error, bool)) errors.MapFunc {
	return func(ctx context.Context, err error) (error, bool) {
		if vErr, ok := err.(ValidationError); ok {
			return f(ctx, vErr)
		}
		return err, false
	}
}

// CondValidation function returns a condition function that matches standard
// Validation error and ensures that the error contains a field and a reason.
func CondValidation() errors.MapCond {
	return func(err error) bool {
		if vErr, ok := err.(ValidationError); ok {
			if vErr.Field != "" && vErr.Reason != "" {
				return true
			}
		}
		return false
	}
}

// CondFieldEq function returns a condition function that checks if the
// field matches the validation error field.
func CondFieldEq(theField string) errors.MapCond {
	return func(err error) bool {
		if vErr, ok := err.(ValidationError); ok {
			if vErr.Field == theField {
				return true
			}
		}
		return false
	}
}

// CondReasonEq function returns a condition function that checks if the
// reason matches the validation error reason.
func CondReasonEq(theReason string) errors.MapCond {
	return func(err error) bool {
		if vErr, ok := err.(ValidationError); ok {
			if vErr.Reason == theReason {
				return true
			}
		}
		return false
	}
}

// DefaultMapping the default behavior for validation error mapping.
func DefaultMapping() errors.MapFunc {
	return errors.NewMapping(
		CondValidation(),
		errors.MapFunc(func(ctx context.Context, err error) (error, bool) {
			vErr, _ := err.(ValidationError)
			return errors.NewContainer(codes.InvalidArgument, "Invalid %s: %s", vErr.Field, vErr.Reason).WithField(vErr.Field, vErr.Reason), true
		}),
	)
}
