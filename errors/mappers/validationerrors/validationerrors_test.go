package validationerrors

import (
	"context"
	"fmt"
	"testing"

	"github.com/infobloxopen/atlas-app-toolkit/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Test for the condition functions.
func TestCond(t *testing.T) {
	for _, tc := range []struct {
		in       error
		name     string
		cond     errors.MapCond
		expected bool
	}{
		{
			name: "Validation Error non pointer",
			cond: CondValidation(), in: &ValidationError{}, expected: false,
		},
		{
			name: "Validation Error base",
			cond: CondValidation(), in: ValidationError{key: true, field: "foo", reason: "bad foo", cause: fmt.Errorf("bad input")}, expected: true,
		},
		{
			name: "None Validation Error",
			cond: CondValidation(), in: fmt.Errorf("Not.Error"), expected: false,
		},
		{
			name: "CondFieldEq non pointer",
			cond: CondFieldEq("bar"), in: &ValidationError{field: "bar"}, expected: false,
		},
		{
			name: "CondFieldEq base",
			cond: CondFieldEq("foo"), in: ValidationError{field: "foo"}, expected: true,
		},
		{
			name: "CondFieldEq base bad",
			cond: CondFieldEq("foo"), in: ValidationError{field: "bar"}, expected: false,
		},
		{
			name: "CondFieldEq bad",
			cond: CondFieldEq("foo"), in: fmt.Errorf("foo"), expected: false,
		},
		{
			name: "CondReasonEq non pointer",
			cond: CondReasonEq("foo bar"), in: &ValidationError{reason: "foo bar"}, expected: false,
		},
		{
			name: "CondReasonEq base",
			cond: CondReasonEq("foo bar"), in: ValidationError{reason: "foo bar"}, expected: true,
		},
		{
			name: "CondReasonEq base bad",
			cond: CondReasonEq("foo"), in: ValidationError{reason: "foo foo"}, expected: false,
		},
		{
			name: "CondReasonEq bad",
			cond: CondReasonEq("foo bar"), in: fmt.Errorf("foo"), expected: false,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if actual := tc.cond(tc.in); actual != tc.expected {
				t.Errorf("Test failed for %v, expected: %v; actual: %v", tc.name, tc.expected, actual)
			}
		})
	}
}

// Test to ensure all mapping functions work as expected.
func TestMapping(t *testing.T) {

	// ToMapFunc Custom Mapping
	f := ToMapFunc(func(ctx context.Context, err ValidationError) (error, bool) {
		return errors.NewContainer(codes.InvalidArgument, "custom error message for field: %v", err.Field()), true
	})

	for _, tc := range []struct {
		in         error
		expected   bool
		mapping    errors.MapFunc
		name       string
		statusCode codes.Code
		statusMsg  string
	}{
		{
			name:       "DefaultMapping base",
			in:         ValidationError{key: true, field: "foo", reason: "bad foo", cause: fmt.Errorf("bad input")},
			expected:   true,
			mapping:    DefaultMapping(),
			statusCode: codes.InvalidArgument,
			statusMsg:  fmt.Sprintf("Invalid foo: bad foo"),
		},
		{
			name:       "DefaultMapping empty",
			in:         ValidationError{},
			expected:   false,
			mapping:    DefaultMapping(),
			statusCode: codes.InvalidArgument,
		},
		{
			name:     "DefaultMapping empty field",
			in:       ValidationError{key: true, reason: "bad foo", cause: fmt.Errorf("bad input")},
			expected: false,
			mapping:  DefaultMapping(),
		},
		{
			name:     "DefaultMapping empty reason",
			in:       ValidationError{key: true, field: "foo", cause: fmt.Errorf("bad input")},
			expected: false,
			mapping:  DefaultMapping(),
		},
		{
			name:       "DefaultMapping not ValidationError",
			in:         fmt.Errorf("Not validation Error"),
			expected:   false,
			mapping:    DefaultMapping(),
			statusCode: codes.InvalidArgument,
		},
		{
			name:       "DefaultMapping nil",
			in:         nil,
			expected:   false,
			mapping:    DefaultMapping(),
			statusCode: codes.InvalidArgument,
		},
		{
			name:     "CustomMapping CondFieldEq base",
			in:       ValidationError{key: true, field: "foo", reason: "bad foo", cause: fmt.Errorf("bad input")},
			expected: true,
			mapping: errors.NewMapping(
				errors.CondAnd(
					CondValidation(),
					CondFieldEq("foo"),
				),
				errors.MapFunc(func(ctx context.Context, err error) (error, bool) {
					vErr, _ := err.(ValidationError)
					return errors.NewContainer(codes.InvalidArgument, "custom error message for field: %v", vErr.Field()), true
				}),
			),
			statusCode: codes.InvalidArgument,
			statusMsg:  fmt.Sprintf("custom error message for field: foo"),
		},
		{
			name:     "CustomMapping CondFieldEq bad",
			in:       ValidationError{key: true, field: "foo", reason: "bad foo", cause: fmt.Errorf("bad input")},
			expected: false,
			mapping: errors.NewMapping(
				errors.CondAnd(
					CondValidation(),
					CondFieldEq("bar"),
				),
				errors.MapFunc(func(ctx context.Context, err error) (error, bool) {
					vErr, _ := err.(ValidationError)
					return errors.NewContainer(codes.InvalidArgument, "custom error message for field: %v", vErr.Field()), true
				}),
			),
		},
		{
			name:     "CustomMapping CondFieldEq empty",
			in:       ValidationError{key: true, reason: "bad foo", cause: fmt.Errorf("bad input")},
			expected: false,
			mapping: errors.NewMapping(
				errors.CondAnd(
					CondValidation(),
					CondFieldEq("bar"),
				),
				errors.MapFunc(func(ctx context.Context, err error) (error, bool) {
					vErr, _ := err.(ValidationError)
					return errors.NewContainer(codes.InvalidArgument, "custom error message for field: %v", vErr.Field()), true
				}),
			),
		},
		{
			name:     "CustomMapping CondFieldEq not validation error",
			in:       fmt.Errorf("Bad error"),
			expected: false,
			mapping: errors.NewMapping(
				CondFieldEq("bar"),
				errors.MapFunc(func(ctx context.Context, err error) (error, bool) {
					vErr, _ := err.(ValidationError)
					return errors.NewContainer(codes.InvalidArgument, "custom error message for field: %v", vErr.Field()), true
				}),
			),
		},
		{
			name:     "CustomMapping CondReasonEq base",
			in:       ValidationError{key: true, field: "foo", reason: "bad foo", cause: fmt.Errorf("bad input")},
			expected: true,
			mapping: errors.NewMapping(
				errors.CondAnd(
					CondValidation(),
					CondReasonEq("bad foo"),
				),
				errors.MapFunc(func(ctx context.Context, err error) (error, bool) {
					vErr, _ := err.(ValidationError)
					return errors.NewContainer(codes.InvalidArgument, "custom error message for reason: %s", vErr.Reason()), true
				}),
			),
			statusCode: codes.InvalidArgument,
			statusMsg:  fmt.Sprintf("custom error message for reason: bad foo"),
		},
		{
			name:     "CustomMapping CondReasonEq bad",
			in:       ValidationError{key: true, field: "foo", reason: "bad foo", cause: fmt.Errorf("bad input")},
			expected: false,
			mapping: errors.NewMapping(
				errors.CondAnd(
					CondValidation(),
					CondReasonEq("bar"),
				),
				errors.MapFunc(func(ctx context.Context, err error) (error, bool) {
					vErr, _ := err.(ValidationError)
					return errors.NewContainer(codes.InvalidArgument, "custom error message for reason: %s", vErr.Reason()), true
				}),
			),
		},
		{
			name:     "CustomMapping CondReasonEq empty",
			in:       ValidationError{key: true, field: "foo", cause: fmt.Errorf("bad input")},
			expected: false,
			mapping: errors.NewMapping(
				errors.CondAnd(
					CondValidation(),
					CondReasonEq("bar"),
				),
				errors.MapFunc(func(ctx context.Context, err error) (error, bool) {
					vErr, _ := err.(ValidationError)
					return errors.NewContainer(codes.InvalidArgument, "custom error message for reason: %v", vErr.Reason()), true
				}),
			),
		},
		{
			name:     "CustomMapping CondReasonEq not validation error",
			in:       fmt.Errorf("Bad error"),
			expected: false,
			mapping: errors.NewMapping(
				CondReasonEq("bar"),
				errors.MapFunc(func(ctx context.Context, err error) (error, bool) {
					vErr, _ := err.(ValidationError)
					return errors.NewContainer(codes.InvalidArgument, "custom error message for reason: %v", vErr.Reason()), true
				}),
			),
		},
		{
			name: "ToMapFunc base",
			in:   ValidationError{key: true, field: "foo", reason: "bad foo", cause: fmt.Errorf("bad input")},
			mapping: errors.NewMapping(
				errors.CondAnd(
					CondValidation(),
					CondFieldEq("foo"),
				),
				f,
			),
			expected:   true,
			statusCode: codes.InvalidArgument,
			statusMsg:  "custom error message for field: foo",
		},
		{
			name:       "ToMapFunc none validation error",
			in:         fmt.Errorf("Not validation error"),
			mapping:    f,
			expected:   false,
			statusCode: codes.InvalidArgument,
			statusMsg:  "custom error message for field: foo",
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mapped, ok := tc.mapping(context.Background(), tc.in)
			if ok != tc.expected {
				t.Errorf("Test failed for %v, expected: %v, actual: %v", tc.name, tc.expected, ok)
			}

			if !ok {
				return
			}

			if actualCode := status.Code(mapped); actualCode != tc.statusCode {
				t.Errorf("Test failed for %v, expected status code: %v, actual: %v", tc.name,
					tc.statusCode, actualCode)
			}

			if actualMsg := status.Convert(mapped).Message(); actualMsg != tc.statusMsg {
				t.Errorf("Test failed for %v, expected status message: %q, actual: %q", tc.name,
					tc.statusMsg, actualMsg)
			}

		})
	}
}
