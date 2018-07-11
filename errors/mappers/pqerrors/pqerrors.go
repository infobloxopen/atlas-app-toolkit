package pqerrors

import (
	"github.com/infobloxopen/atlas-app-toolkit/errors"
	"github.com/lib/pq"
)

const (
	msgForeignKeyViolation = "Cannot insert The '%s' object as it does not refer to a valid '%s' object."
	msgRestrictViolation   = "Cannot update or delete an '%s' object as it is referenced by a '%s' object."
	msgNotNullViolation    = "The '%s' field for the '%s' object cannot be empty."
	msgUniqueViolation     = "There is already an existing '%s' object with the same '%s'."
)

// ToMapFunc function converts mapping function for *pq.Error to a conventional
// MapFunc from atlas-app-toolkit/errors package.
func ToMapFunc(f func(context.Context, *pq.Error) (error, bool)) errors.MapFunc {
	return func(ctx context.Context, err error) (error, bool) {
		if pqErr, ok := err.(*pq.Error); ok {
			return f(ctx, pqErr)
		}

		return nil, false
	}
}

var CondPQ errors.MapCond = func(err error) bool {
	_, ok := err.(*pq.Error)
	return ok
}

// CondConstraint function returns a condition function that matches a
// particular constraint name.
func CondConstraintEq(c string) errors.MapCond {
	return func(err error) bool {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Constraint == c {
				return true
			}
		}

		return false
	}
}

// CondConstraintCode function returns a condition function that matches
// a particular constraint code.
func CondConstrintCodeEq(code string) errors.MapCond {
	return func(err error) bool {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == code {
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
			CondConstraintCodeEq("23503"),
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
			CondConstrainCodeEq("23001"),
			CondConstraintEq(c),
		),
		errors.NewContainer(codes.InvalidArgument, msgRestrictKeyViolation, t1, t2),
	)
}

// NewNotNullMapping function returns a mapping function that performs
// mapping of a constraint name (c) to specific not-null violation error
// message with user-friendly table (t) name and column (col) name provided.
func NewNotNullMapping(c string, t string, col string) errors.MapFunc {
	return errors.NewMapping(
		errors.CondAnd(
			CondConstraintCodeEq("23502"),
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
			CondConstraintCodeEq("23505"),
			CondConstraintEq(c),
		),
		errors.NewContainer(codes.InvalidArgument, msgUniqueViolation, t, col),
	)
}
