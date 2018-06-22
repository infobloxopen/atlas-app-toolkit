package errors

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
)

func testCond(t *testing.T, cnd MapCond, input error, e bool) {
	if cnd(input) != e {
		t.Errorf("Condition test failed with input %v: Expected %v, got %v", input, e, !e)
	}
}

func testMapping(t *testing.T, mf MapFunc, input error, expErr error, expOk bool) {
	err, ok := mf(nil, input)

	if !reflect.DeepEqual(err, expErr) || expOk != ok {
		t.Errorf("Mapping test failed with input %v, %v", input, expOk)
	}
}

func TestCond(t *testing.T) {
	var cnd MapCond

	if cnd.Error() != "MapCond" {
		t.Errorf("Invalid Error() result: expected 'MapCond', got %q.", cnd.Error())
	}

	// CondEq Test

	cnd = CondEq("foo")

	testCond(t, cnd, fmt.Errorf("foo"), true)
	testCond(t, cnd, fmt.Errorf("bar"), false)

	// CondReMatch Test

	cnd = CondReMatch(`^foo[^-]*bar$`)

	testCond(t, cnd, fmt.Errorf("foobar"), true)
	testCond(t, cnd, fmt.Errorf("foogazbar"), true)
	testCond(t, cnd, fmt.Errorf("foo"), false)
	testCond(t, cnd, fmt.Errorf("foo-bar"), false)
	testCond(t, cnd, fmt.Errorf("bar"), false)

	// CondHasSuffix Test

	cnd = CondHasSuffix("foo")

	testCond(t, cnd, fmt.Errorf("zfoo"), true)
	testCond(t, cnd, fmt.Errorf("foobar"), false)

	// CondHasPrefix Test

	cnd = CondHasPrefix("foo")

	testCond(t, cnd, fmt.Errorf("fooz"), true)
	testCond(t, cnd, fmt.Errorf("barfoo"), false)

	// CondNot Test

	cnd = CondNot(CondReMatch(`^foo[^-]*bar$`))

	testCond(t, cnd, fmt.Errorf("foobar"), false)
	testCond(t, cnd, fmt.Errorf("foogazbar"), false)
	testCond(t, cnd, fmt.Errorf("foo"), true)
	testCond(t, cnd, fmt.Errorf("foo-bar"), true)
	testCond(t, cnd, fmt.Errorf("bar"), true)

	// CondAnd Test

	cnd = CondAnd(CondHasSuffix("foo"), CondHasPrefix("bar"))

	testCond(t, cnd, fmt.Errorf("zfoo"), false)
	testCond(t, cnd, fmt.Errorf("foobar"), false)
	testCond(t, cnd, fmt.Errorf("barfoo"), true)

	// CondOr Test

	cnd = CondOr(CondHasSuffix("foo"), CondHasPrefix("bar"))

	testCond(t, cnd, fmt.Errorf("zfoo"), true)
	testCond(t, cnd, fmt.Errorf("foobar"), false)
	testCond(t, cnd, fmt.Errorf("baro"), true)
}

func TestNewMapping(t *testing.T) {
	var mf MapFunc

	someError := NewContainer(codes.InvalidArgument, "Some Error")

	// Test Mapping err -> container

	mf = NewMapping(fmt.Errorf("some error"), someError)

	if mf.Error() != "MapFunc" {
		t.Errorf("Invalid Error() result: expected 'MapFunc', got %q.", mf.Error())
	}

	testMapping(t, mf, fmt.Errorf("some error"), someError, true)
	testMapping(t, mf, fmt.Errorf("some other error"), nil, false)

	// Test Mapping with MapCond function.

	mf = NewMapping(CondReMatch(`^foo`), someError)

	testMapping(t, mf, fmt.Errorf("foo"), someError, true)
	testMapping(t, mf, fmt.Errorf("zfoo"), nil, false)

	// Test Mapping with MapCond and MapFunc.

	mf = NewMapping(
		CondReMatch(`^pg_sql:`),
		MapFunc(func(ctx context.Context, err error) (error, bool) {
			return NewContainer(codes.InvalidArgument, strings.TrimPrefix(err.Error(), "pg_sql:")), true
		}),
	)

	pgSqlErr := fmt.Errorf("pg_sql:Some Error")

	testMapping(t, mf, pgSqlErr, someError, true)
	testMapping(t, mf, fmt.Errorf("pg_sk"), nil, false)

	// Test Skip.

	mf = NewMapping(fmt.Errorf("Skip"), nil)

	testMapping(t, mf, fmt.Errorf("Skip"), nil, true)
	testMapping(t, mf, fmt.Errorf("NotSkip"), nil, false)
}

func TestMapper(t *testing.T) {
	m := Mapper{}

	// Test AddMapping.

	if len(m.mapFuncs) != 0 {
		t.Errorf("Unexpected mapFuncs count, expected %d, got %d", 0, len(m.mapFuncs))
	}

	err1 := NewContainer(codes.InvalidArgument, "Error One")

	m.AddMapping(
		NewMapping(fmt.Errorf("Err1"), err1),
		NewMapping(
			CondReMatch(`^Err[^1]{3}$`),
			MapFunc(func(ctx context.Context, err error) (error, bool) {
				return NewContainer(
					codes.Internal,
					"Internal Error "+strings.TrimPrefix(err.Error(), "Err"),
				), true
			}),
		),
		NewMapping(fmt.Errorf("To Skip"), nil),
	)

	if len(m.mapFuncs) != 3 {
		t.Errorf("Unexpected mapFuncs count, expected %d, got %d", 3, len(m.mapFuncs))
	}

	// Test Map.

	checkContainer(t, m.Map(nil, fmt.Errorf("Err1")).(*Container), err1)
	checkContainer(t, m.Map(nil, fmt.Errorf("Err666")).(*Container), NewContainer(codes.Internal, "Internal Error 666"))
	checkContainer(t, m.Map(nil, fmt.Errorf("Err61")).(*Container), InitContainer())
	checkContainer(t, m.Map(nil, fmt.Errorf("rEE")).(*Container), InitContainer())

	// Test Map with skip.

	if m.Map(nil, fmt.Errorf("To Skip")) != nil {
		t.Errorf("Expected nil, got non-nil value")
	}
}
