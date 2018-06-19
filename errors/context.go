package errors

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
)

const (
	// Context key for Error Container.
	DefaultErrorContainerKey = "Error-Container"
)

// NewContext function creates a context with error container saved in it.
func NewContext(ctx context.Context, c *Container) context.Context {
	return context.WithValue(ctx, DefaultErrorContainerKey, c)
}

// FromContext function retrieves an error container value from context.
func FromContext(ctx context.Context) *Container {
	if c := ctx.Value(DefaultErrorContainerKey); c != nil {
		return c.(*Container)
	}

	return nil
}

// Detail function appends a new detail to a context stored error container's
// 'details' section.
func Detail(ctx context.Context, code codes.Code, target string, format string, args ...interface{}) *Container {
	return FromContext(ctx).WithDetail(code, target, format, args...)
}

// Field function appends a field error detail to a context stored error
// container's 'fields' section.
func Field(ctx context.Context, target string, format string, args ...interface{}) *Container {
	return FromContext(ctx).WithField(target, format, args...)
}

// Set function initializes a general error code and error message for context
// stored error container and also appends a details with the same content to
// an error container's 'details' section.
func Set(ctx context.Context, code codes.Code, target string, format string, args ...interface{}) *Container {
	return FromContext(ctx).Set(code, target, format, args...)
}

// IfSet function intializes general error code and error message for context
// stored error container if and onyl if any error was set previously by
// calling Set, WithField(s), WithDetails(s).
func IfSet(ctx context.Context, code codes.Code, format string, args ...interface{}) error {
	return FromContext(ctx).IfSet(code, format, args...)
}

// ContextError function returns an error container if any error field, detail
// or message was set, else it returns nil.
func ContextError(ctx context.Context) error {
	if FromContext(ctx).IsSet() {
		return FromContext(ctx)
	}

	return nil
}
