package pgxerrors

import (
	"context"

	"github.com/infobloxopen/atlas-app-toolkit/errors"
	"github.com/jackc/pgconn"
	"google.golang.org/grpc/codes"
)

const (
	msgForeignKeyViolation = "Cannot insert object '%s' as it does not refer to a valid '%s' object."
	msgRestrictViolation   = "Cannot update or delete an object '%s' as it is referenced by a '%s' object."
	msgNotNullViolation    = "The '%s' field for the '%s' object cannot be empty."
	msgUniqueViolation     = "There is already an existing '%s' object with the same '%s'."
)

// ToMapFunc function converts mapping function for *pgconn.PgError to a conventional
// MapFunc from atlas-app-toolkit/errors package.
func ToMapFunc(f func(context.Context, *pgconn.PgError) (error, bool)) errors.MapFunc {
	return func(ctx context.Context, err error) (error, bool) {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			return f(ctx, pgErr)
		}

		return err, false
	}
}

// CondPgError function returns a condition function that matches pgconn.PgError error.
func CondPgError() errors.MapCond {
	return func(err error) bool {
		_, ok := err.(*pgconn.PgError)
		return ok
	}
}

// CondConstraintEq function returns a condition function that matches a
// particular constraint name.
func CondConstraintEq(c string) errors.MapCond {
	return func(err error) bool {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.ConstraintName == c {
				return true
			}
		}

		return false
	}
}

// CondCodeEq function returns a condition function that matches
// a particular constraint code.
func CondCodeEq(code string) errors.MapCond {
	return func(err error) bool {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if string(pgErr.Code) == code {
				return true
			}
		}
		return false
	}
}

// NewForeignKeyMapping function returns a mapping function that performs
// mapping of a constraint name (c) to specific foreign key message with
// user-friendly referencing (t1) and referenced (t2) table names provided.
func NewForeignKeyMapping(c string, t1 string, t2 string) errors.MapFunc {
	return errors.NewMapping(
		errors.CondAnd(
			CondCodeEq("23503"),
			CondConstraintEq(c),
		),
		errors.NewContainer(codes.InvalidArgument, msgForeignKeyViolation, t1, t2),
	)
}

// NewRestrictMapping function returns a mapping function that performs
// mapping of a constraint name (c) to specific restrict violation error message
// with user-friendly referencing (t1) and referenced (t2) table names provided.
func NewRestrictMapping(c string, t1 string, t2 string) errors.MapFunc {
	return errors.NewMapping(
		errors.CondAnd(
			CondCodeEq("23001"),
			CondConstraintEq(c),
		),
		errors.NewContainer(codes.InvalidArgument, msgRestrictViolation, t1, t2),
	)
}

// NewNotNullMapping function returns a mapping function that performs
// mapping of a constraint name (c) to specific not-null violation error
// message with user-friendly table (t) name and column (col) name provided.
func NewNotNullMapping(c string, t string, col string) errors.MapFunc {
	return errors.NewMapping(
		errors.CondAnd(
			CondCodeEq("23502"),
			CondConstraintEq(c),
		),
		errors.NewContainer(codes.InvalidArgument, msgNotNullViolation, col, t),
	)
}

// NewUniqueMapping function returns a mapping function that performs
// mapping of a constraint name (c) to a specific unique violation error
// message with user-friendly table name (t) and column (col) name provided.
func NewUniqueMapping(c string, t string, col string) errors.MapFunc {
	return errors.NewMapping(
		errors.CondAnd(
			CondCodeEq("23505"),
			CondConstraintEq(c),
		),
		errors.NewContainer(codes.AlreadyExists, msgUniqueViolation, t, col),
	)
}
