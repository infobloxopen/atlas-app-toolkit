package errors

import (
	"context"
	"regexp"
	"strings"
)

type MapperFunc func(err error) (error, bool)

// MapCondFunc function takes an error and returns flag that indicates
// whether the map succeeded and mapped error.
type MapCondFunc func(c *Container, err error) (error, bool)

// MapFunc type performs any necessary action on err and Container and returns an
// error.
type MapFunc func(c *Container, err error) error

// MapCond function takes an error and returns flag that indicates
// whether the map succeeded.
type MapCond func(error) bool

// WithMappingChain function combines several MapCondFunc and returns them as one function.
func (c *Container) WithMappingChain(mcf ...MapCondFunc) *Container {
	c.mapper = func(err error) (error, bool) {
		for _, v := range mcf {
			if mapped, ok := v(c, err); ok {
				return mapped, true
			}
		}

		return nil, false
	}

	return c
}

// WithMapping function sets error container function.
func (c *Container) WithMapping(mcf MapCondFunc) *Container {
	c.mapper = func(err error) (error, bool) {
		return mcf(c, err)
	}
	return c
}

func (c *Container) Map(err error) error {
	if c.mapper == nil {
		return NewContainer()
	}

	if mapped, ok := c.mapper(err); ok {
		return mapped
	}

	return NewContainer()
}

func Map(ctx context.Context, err error) error {
	return FromContext(ctx).Map(err)
}

// ToMapCondFunc takes an source error and desired transformation and
// condition function.
func ToMapCondFunc(mf MapFunc, cond MapCond) MapCondFunc {
	return func(c *Container, err error) (error, bool) {
		if cond(err) {
			return mf(c, err), true
		}
		return nil, false
	}
}

// CondEq function takes a string as an input and returns a condition
// function that checks whether the error is equal to a string given.
func CondEq(src string) MapCond {
	return func(err error) bool {
		return src == err.Error()
	}
}

// CondReMatch takes a string regexp pattern as an input and returns
// a condition function that checks whether the error matches the pattern given.
func CondReMatch(pattern string) MapCond {
	return func(err error) bool {
		matched, _ := regexp.MatchString(pattern, err.Error())
		return matched
	}
}

// CondHasSuffix function takes a string as an input and returns
// a condition function that checks whether the error ends with
// the string given.
func CondHasSuffix(suffix string) MapCond {
	return func(err error) bool {
		return strings.HasSuffix(err.Error(), suffix)
	}
}

// CondHasPrefix function takes a string as an input and returns
// a condition function that checks whether the error starts with
// the string given.
func CondHasPrefix(prefix string) MapCond {
	return func(err error) bool {
		return strings.HasPrefix(err.Error(), prefix)
	}
}

// CondNot takes a condtion function as an input and returns a function
// that asserts inverse result.
func CondNot(mc MapCond) MapCond {
	return func(err error) bool {
		return !mc(err)
	}
}

// CondAnd takes a list of condition function as an input and returns a function
// that asserts true if and only if all conditions are satisfied.
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

// CondOr takes a list of condition function as an input and returns a function
// that asserts true if at least one of conditions is satisfied.
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
