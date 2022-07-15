package pgxerrors

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/infobloxopen/atlas-app-toolkit/errors"
)

func TestCond(t *testing.T) {
	for _, tc := range []struct {
		in       error
		name     string
		cond     errors.MapCond
		expected bool
	}{
		{
			name: "cond_pq_base",
			cond: CondPgError(), in: &pgconn.PgError{}, expected: true,
		},
		{
			name: "cond_pq_invalid_error",
			cond: CondPgError(), in: fmt.Errorf("pgconn.PgError"), expected: false,
		},
		{
			name: "cond_constr_base",
			cond: CondConstraintEq("foo"), in: &pgconn.PgError{ConstraintName: "foo"}, expected: true,
		},
		{
			name: "cond_constr_invalid_error",
			cond: CondConstraintEq("foo"), in: fmt.Errorf("foo"), expected: false,
		},
		{
			name: "cond_code_base",
			cond: CondCodeEq("1312"), in: &pgconn.PgError{Code: "1312"}, expected: true,
		},
		{
			name: "cond_code_invalid_error",
			cond: CondCodeEq("1312"), in: fmt.Errorf("foo"), expected: false,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if actual := tc.cond(tc.in); actual != tc.expected {
				t.Errorf("expected %v; got %v", tc.expected, actual)
			}
		})
	}
}

func TestMapping(t *testing.T) {

	// ToMapFunc Custom Mapping
	var f errors.MapFunc = ToMapFunc(func(ctx context.Context, err *pgconn.PgError) (error, bool) {
		if err.Detail == "yay" {
			return errors.NewContainer(codes.DataLoss, "data loss"), true
		}

		return nil, false
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
			name:       "base_restrict",
			in:         &pgconn.PgError{ConstraintName: "foo", Code: "23001"},
			expected:   true,
			mapping:    NewRestrictMapping("foo", "bar", "baz"),
			statusCode: codes.InvalidArgument,
			statusMsg:  fmt.Sprintf(msgRestrictViolation, "bar", "baz"),
		},
		{
			name:     "restrict_invalid_code",
			in:       &pgconn.PgError{ConstraintName: "foo", Code: "1312"},
			mapping:  NewRestrictMapping("foo", "baz", "bar"),
			expected: false,
		},
		{
			name:     "restrict_invalid_constraint",
			in:       &pgconn.PgError{ConstraintName: "hex", Code: "23503"},
			mapping:  NewRestrictMapping("foo", "bar", "baz"),
			expected: false,
		},
		{
			name:       "base_not_null",
			in:         &pgconn.PgError{ConstraintName: "foo", Code: "23502"},
			expected:   true,
			mapping:    NewNotNullMapping("foo", "bar", "baz"),
			statusCode: codes.InvalidArgument,
			statusMsg:  fmt.Sprintf(msgNotNullViolation, "baz", "bar"),
		},
		{
			name:     "not_null_invalid_code",
			in:       &pgconn.PgError{ConstraintName: "foo", Code: "1312"},
			mapping:  NewNotNullMapping("foo", "baz", "bar"),
			expected: false,
		},
		{
			name:     "not_null_invalid_constraint",
			in:       &pgconn.PgError{ConstraintName: "hex", Code: "23503"},
			mapping:  NewNotNullMapping("foo", "baz", "bar"),
			expected: false,
		},
		{
			name:       "base_foreign_key",
			in:         &pgconn.PgError{ConstraintName: "foo", Code: "23503"},
			expected:   true,
			mapping:    NewForeignKeyMapping("foo", "baz", "bar"),
			statusCode: codes.InvalidArgument,
			statusMsg:  fmt.Sprintf(msgForeignKeyViolation, "baz", "bar"),
		},
		{
			name:     "foreign_key_invalid_code",
			in:       &pgconn.PgError{ConstraintName: "foo", Code: "1312"},
			mapping:  NewForeignKeyMapping("foo", "baz", "bar"),
			expected: false,
		},
		{
			name:     "foreign_key_invalid_constraint",
			in:       &pgconn.PgError{ConstraintName: "hex", Code: "23503"},
			mapping:  NewForeignKeyMapping("foo", "baz", "bar"),
			expected: false,
		},
		{
			name:       "base_unique",
			in:         &pgconn.PgError{ConstraintName: "foo", Code: "23505"},
			expected:   true,
			mapping:    NewUniqueMapping("foo", "baz", "bar"),
			statusCode: codes.AlreadyExists,
			statusMsg:  fmt.Sprintf(msgUniqueViolation, "baz", "bar"),
		},
		{
			name:     "unique_invalid_code",
			in:       &pgconn.PgError{ConstraintName: "foo", Code: "1312"},
			mapping:  NewUniqueMapping("foo", "baz", "bar"),
			expected: false,
		},
		{
			name:     "unique_invalid_constraint",
			in:       &pgconn.PgError{ConstraintName: "hex", Code: "23505"},
			mapping:  NewUniqueMapping("foo", "baz", "bar"),
			expected: false,
		},
		{
			name:       "tomapfunc_base",
			in:         &pgconn.PgError{Detail: "yay"},
			mapping:    f,
			expected:   true,
			statusCode: codes.DataLoss,
			statusMsg:  "data loss",
		},
		{
			name:     "tomapfunc_invalid_err",
			in:       &pgconn.PgError{Detail: "ya"},
			mapping:  f,
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
