package errors

import (
	"context"
	"regexp"
	"strings"
)

// Mapper struct ...
type Mapper struct {
	mapFuncs []MapFunc
}

// Map function performs a mapping from error given following a chain of
// mappings that were defined prior to the Map call.
func (m *Mapper) Map(err error) error {
	for _, mapFunc := range mapFuncs {
		if resErr, ok := mapFunc(err); ok {
			return resErr
		}
	}
	return NewContainer()
}

// AddMapping function appends a list of mapping functions to a mapping chain.
func (m *Mapper) AddMapping(mf ...MapFunc) Mapper {
	if m.mapFuncs == nil {
		m.mapFuncs = []MapFunc{}
	}

	m.mapFuncs = append(m.mapFuncs, mf...)

	return m
}

// MapCond function takes an error and returns flag that indicates whether the
// map condition was met.
type MapCond func(error) bool
func (mc MapCond) Error() string { return "MapCond" }

// MapFunc function takes an error and returns mapped error and flag that
// indicates whether the mapping was performed successfully.
type MapFunc func(error) (error, bool)
func (mc MapFunc) Error() string { return "MapFunc" }

// NewMapping function ...
func NewMapping(src error, dst error) MapFunc {
	var mapCond MapCond
	var mapFunc MapFunc

	if v, ok := src.(MapCond); ok {
		mapCond = v
	} else {
		mapCond = CondEq(src.Error())
	}

	if v, ok := dst.(MapFunc); ok {
		mapFunc = v
	} else {
		mapFunc = func(err error) (error, bool) {
			return dst, true
		}
	}

	return func(err error) (error, bool) {
		if mapCond(err) {
			return mapFunc(err)
		}

		return nil, false
	}
}

// * -------------------------------------------- *
// * Various helper condition building functions. *
// * -------------------------------------------- *

// CondEq function takes a string as an input and returns a condition function
// that checks whether the error is equal to a string given.
func CondEq(src string) MapCond {
	return func(err error) bool {
		return src == err.Error()
	}
}

// CondReMatch function takes a string regexp pattern as an input and returns a
// condition function that checks whether the error matches the pattern given.
func CondReMatch(pattern string) MapCond {
	return func(err error) bool {
		matched, _ := regexp.MatchString(pattern, err.Error())
		return matched
	}
}

// CondHasSuffix function takes a string as an input and returns a condition
// function that checks whether the error ends with the string given.
func CondHasSuffix(suffix string) MapCond {
	return func(err error) bool {
		return strings.HasSuffix(err.Error(), suffix)
	}
}

// CondHasPrefix function takes a string as an input and returns a condition
// function that checks whether the error starts with the string given.
func CondHasPrefix(prefix string) MapCond {
	return func(err error) bool {
		return strings.HasPrefix(err.Error(), prefix)
	}
}

// CondNot function takes a condtion function as an input and returns a
// function that asserts inverse result.
func CondNot(mc MapCond) MapCond {
	return func(err error) bool {
		return !mc(err)
	}
}

// CondAnd function takes a list of condition function as an input and returns
// a function that asserts true if and only if all conditions are satisfied.
func CondAnd(mcs ...MapCond) MapCond {
	return func(err error) bool {
		for _, v := range mcs {
			if !v(err) {
				return false
			}
		}
		return true
	}
}

// CondOr function takes a list of condition function as an input and returns
// a function that asserts true if at least one of conditions is satisfied.
func CondOr(mcs ...MapCond) MapCond {
	return func(err error) bool {
		for _, v := range mcs {
			if v(err) {
				return true
			}
		}
		return false
	}
}
