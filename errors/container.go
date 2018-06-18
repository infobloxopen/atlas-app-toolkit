// The Error Container entity serves a purpose for keeping track of errors,
// details and fields information. This component can be used explicitly
// (for example when we want to sequentially fill it with details and fields),
// as well as implicitly (all errors that are returned from handler are
// transformed to an Error Container, and passed as GRPCStatus to a gRPC Gateway).
package errors

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/golang/protobuf/proto"

	"github.com/infobloxopen/atlas-app-toolkit/rpc/errdetails"
	"github.com/infobloxopen/atlas-app-toolkit/rpc/errfields"
)

const (
	// Context key for Error Container.
	DefaultErrorContainerKey = "Error-Container"
)

// Container type serves a purpose of an error container and
// error mapping entity.
type Container struct {
	// details field contains an array of error details.
	details []*errdetails.TargetInfo

	// fields field contains per-field error map.
	fields *errfields.FieldInfo

	// errCode, errMessage field contain the general error
	// message.
	errCode    codes.Code
	errMessage string
	errTarget  string

	errSet bool

	mapper MapperFunc
}

func NewContainer() Container {
	return Container{}.New(codes.Unknown, "Unknown")
}

// Error function ...
func (c Container) Error() string { return "Unknown" }

// GRPCStatus function ...
func (c Container) GRPCStatus() *status.Status {
	protoArr := make([]proto.Message, len(c.details)+1)

	protoArr[0] = proto.Message(c.fields)

	for i, d := range c.details {
		protoArr[i+1] = proto.Message(d)
	}

	if s, err := status.New(c.errCode, c.errMessage).WithDetails(protoArr...); err != nil {
		return s
	}

	return nil
}

func (c Container) IsSet() bool {
	return c.errSet
}

func (c Container) Set(code codes.Code, target string, format string, args ...interface{}) Container {
	c.errCode = code
	c.errMessage = fmt.Sprintf(format, args...)
	c.errTarget = target
	c.errSet = true

	c = c.WithDetail(c.errCode, c.errTarget, c.errMessage)

	return c
}

func (c Container) New(code codes.Code, format string, args ...interface{}) Container {
	c.errCode = code
	c.errMessage = fmt.Sprintf(format, args...)

	c.errSet = true

	return c
}

// WithDetail function appends a new detail to a list of details.
func (c Container) WithDetail(code codes.Code, target string, format string, args ...interface{}) Container {
	if c.details == nil {
		c.details = []*errdetails.TargetInfo{}
	}

	c.details = append(c.details, errdetails.Newf(code, target, format, args...))

	return c
}

func (c Container) WithDetails(details []*errdetails.TargetInfo) Container {
	if c.details == nil {
		c.details = []*errdetails.TargetInfo{}
	}

	c.details = append(c.details, details...)
	return c
}

// WithField function appends a error detail regarding particular field.
func (c Container) WithField(target string, format string, args ...interface{}) Container {
	if c.fields == nil {
		c.fields = &errfields.FieldInfo{}
	}

	c.fields.AddField(target, fmt.Sprintf(format, args...))
	return c
}

func (c Container) WithFields(fields map[string]string) Container {
	for k, v := range fields {
		c.WithField(k, v)
	}
	return c
}

// NewContext function saves a container to a context.
func NewContext(ctx context.Context, c Container) context.Context {
	return context.WithValue(ctx, DefaultErrorContainerKey, c)
}

// FromContext function restores container value from context.
func FromContext(ctx context.Context) Container {
	if c := ctx.Value(DefaultErrorContainerKey); c != nil {
		return c.(Container)
	}

	return NewContainer()
}

// Detail function appends a new detail to a list of details in container that is placed in a context.
func Detail(ctx context.Context, code codes.Code, target string, format string, args ...interface{}) Container {
	return FromContext(ctx).WithDetail(code, target, format, args...)
}

// Field function appends a error detail regarding particular field to a container that is placed in a context.
func Field(ctx context.Context, target string, format string, args ...interface{}) Container {
	return FromContext(ctx).WithField(target, format, args...)
}

func Set(ctx context.Context, code codes.Code, target string, format string, args ...interface{}) Container {
	return FromContext(ctx).Set(code, target, format, args...)
}

func ContextError(ctx context.Context) error {
	if FromContext(ctx).IsSet() {
		return FromContext(ctx)
	}

	return nil
}
