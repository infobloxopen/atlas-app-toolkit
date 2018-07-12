package pqerrors

import (
	"context"
	"fmt"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/lib/pq"

	"github.com/infobloxopen/atlas-app-toolkit/errors"
)

func TestCondPQ(t *testing.T) {
	for i, tc := range []struct {
		in       error
		expected bool
	}{
		{in: pq.Error{}, expected: false},
		{in: &pq.Error{}, expected: true},
		{in: fmt.Errorf("pq.Error"), expected: false},
	} {
		actual := CondPQ()(tc.in)
		if actual != tc.expected {
			t.Errorf("TestCase #%d: expected %v; got %v", i, tc.expected, actual)
		}
	}
}

func TestCondConstraintEq(t *testing.T) {
	testC := "constraint_foo"

	for i, tc := range []struct {
		in       error
		expected bool
	}{
		{in: pq.Error{Constraint: testC}, expected: false},
		{in: &pq.Error{Constraint: testC}, expected: true},
		{in: fmt.Errorf(testC), expected: false},
	} {
		actual := CondConstraintEq(testC)(tc.in)
		if actual != tc.expected {
			t.Errorf("TestCase #%d: expected %v; got %v", i, tc.expected, actual)
		}
	}
}

func TestCondConstraintCodeEq(t *testing.T) {
	testC := "1366"

	for i, tc := range []struct {
		in       error
		expected bool
	}{
		{in: pq.Error{Code: pq.ErrorCode(testC)}, expected: false},
		{in: &pq.Error{Code: pq.ErrorCode(testC)}, expected: true},
		{in: fmt.Errorf(testC), expected: false},
	} {
		actual := CondConstraintCodeEq(testC)(tc.in)
		if actual != tc.expected {
			t.Errorf("TestCase #%d: expected %v; got %v", i, tc.expected, actual)
		}
	}
}

func TestMapping(t *testing.T) {
	for _, tc := range []struct {
		in         error
		c, t1, t2  string
		expected   bool
		mapping    errors.MapFunc
		name       string
		statusCode codes.Code
		statusMsg  string
	}{
		{
			name:       "base_restrict",
			in:         &pq.Error{Constraint: "foo", Code: "23001"},
			expected:   true,
			mapping:    NewRestrictMapping("foo", "bar", "baz"),
			statusCode: codes.InvalidArgument,
			statusMsg:  fmt.Sprintf(msgRestrictViolation, "bar", "baz"),
		},
		{
			name:     "restrict_invalid_code",
			in:       &pq.Error{Constraint: "foo", Code: "1312"},
			mapping:  NewRestrictMapping("foo", "baz", "bar"),
			expected: false,
		},
		{
			name:     "restrict_invalid_constraint",
			in:       &pq.Error{Constraint: "hex", Code: "23503"},
			mapping:  NewRestrictMapping("foo", "bar", "baz"),
			expected: false,
		},
		{
			name:       "base_not_null",
			in:         &pq.Error{Constraint: "foo", Code: "23502"},
			expected:   true,
			mapping:    NewNotNullMapping("foo", "bar", "baz"),
			statusCode: codes.InvalidArgument,
			statusMsg:  fmt.Sprintf(msgNotNullViolation, "baz", "bar"),
		},
		{
			name:     "not_null_invalid_code",
			in:       &pq.Error{Constraint: "foo", Code: "1312"},
			mapping:  NewNotNullMapping("foo", "baz", "bar"),
			expected: false,
		},
		{
			name:     "not_null_invalid_constraint",
			in:       &pq.Error{Constraint: "hex", Code: "23503"},
			mapping:  NewNotNullMapping("foo", "baz", "bar"),
			expected: false,
		},
		{
			name:       "base_foreign_key",
			in:         &pq.Error{Constraint: "foo", Code: "23503"},
			expected:   true,
			mapping:    NewForeignKeyMapping("foo", "baz", "bar"),
			statusCode: codes.InvalidArgument,
			statusMsg:  fmt.Sprintf(msgForeignKeyViolation, "baz", "bar"),
		},
		{
			name:     "foreign_key_invalid_code",
			in:       &pq.Error{Constraint: "foo", Code: "1312"},
			mapping:  NewForeignKeyMapping("foo", "baz", "bar"),
			expected: false,
		},
		{
			name:     "foreign_key_invalid_constraint",
			in:       &pq.Error{Constraint: "hex", Code: "23503"},
			mapping:  NewForeignKeyMapping("foo", "baz", "bar"),
			expected: false,
		},
		{
			name:       "base_unique",
			in:         &pq.Error{Constraint: "foo", Code: "23505"},
			expected:   true,
			mapping:    NewUniqueMapping("foo", "baz", "bar"),
			statusCode: codes.AlreadyExists,
			statusMsg:  fmt.Sprintf(msgUniqueViolation, "baz", "bar"),
		},
		{
			name:     "unique_invalid_code",
			in:       &pq.Error{Constraint: "foo", Code: "1312"},
			mapping:  NewUniqueMapping("foo", "baz", "bar"),
			expected: false,
		},
		{
			name:     "unique_invalid_constraint",
			in:       &pq.Error{Constraint: "hex", Code: "23505"},
			mapping:  NewUniqueMapping("foo", "baz", "bar"),
			expected: false,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mapped, ok := tc.mapping(context.Background(), tc.in)
			if ok != tc.expected {
				t.Errorf("expected %v; got %v", tc.expected, ok)
			}

			if !ok {
				return
			}

			if actualCode := status.Code(mapped); actualCode != tc.statusCode {
				t.Errorf("status code: expected %v; got %v",
					tc.statusCode, actualCode)
			}

			if actualMsg := status.Convert(mapped).Message(); actualMsg != tc.statusMsg {
				t.Errorf("status message: expected %q; got %q",
					tc.statusMsg, actualMsg)
			}

		})
	}
}
