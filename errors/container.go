// The Error Container entity serves a purpose for keeping track of errors,
// details and fields information. This component can be used explicitly (for
// example when we want to sequentially fill it with details and fields), as
// well as implicitly (all errors that are returned from handler are
// transformed to an Error Container, and passed as GRPCStatus to a gRPC Gateway).
package errors

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/golang/protobuf/proto"

	"github.com/infobloxopen/atlas-app-toolkit/rpc/errdetails"
	"github.com/infobloxopen/atlas-app-toolkit/rpc/errfields"
)

// Container struct is an entity that servers a purpose of error container and
// consist of methods to append details, field errors and setting general
// error code/message.
type Container struct {
	// details field contains an array of error details.
	details []*errdetails.TargetInfo

	// fields field contains per-field error map.
	fields *errfields.FieldInfo

	// errCode, errMessage field contain the general error message.
	errCode    codes.Code
	errMessage string

	// errSet flag indicates whether the error was set by calling one of
	// following methods: Set, WithDetail(s), WithField(s).
	errSet bool

	// Mapper structure performs necessary mappings.
	Mapper
}

// NewContainer function returns a new entity of error container.
func NewContainer(code codes.Code, format string, args ...interface{}) *Container {
	return (&Container{}).New(code, format, args...)
}

func InitContainer() *Container {
	return (&Container{}).New(codes.Unknown, "Unknown")
}

// Error function returns error message currently associated with container.
func (c Container) Error() string { return c.errMessage }

// GRPCStatus function returns an error container as GRPC status.
func (c *Container) GRPCStatus() *status.Status {
	protoArr := []proto.Message{}

	if c.fields != nil {
		protoArr = append(protoArr, proto.Message(c.fields))
	}

	for _, d := range c.details {
		protoArr = append(protoArr, proto.Message(d))
	}

	if s, err := status.New(c.errCode, c.errMessage).WithDetails(protoArr...); err == nil {
		return s
	}

	return nil
}

// New function instantinates general error code and error message for error
// container.
func (c *Container) New(code codes.Code, format string, args ...interface{}) *Container {
	c.errCode = code
	c.errMessage = fmt.Sprintf(format, args...)

	c.details = nil
	c.fields = nil
	c.errSet = false

	return c
}

// IsSet function returns flag that determines whether the main error code and
// error message were set or not.
func (c *Container) IsSet() bool {
	return c.errSet
}

// Set function initializes general error code and error message for error
// container and also appends a detail with the same content to a an error
// container's 'details' section.
func (c *Container) Set(target string, code codes.Code, format string, args ...interface{}) *Container {
	c.errCode = code
	c.errMessage = fmt.Sprintf(format, args...)
	c.errSet = true

	c = c.WithDetail(c.errCode, target, c.errMessage)

	return c
}

// IfSet function initializes general error code and error message for error
// container if and only if any error was set previously by calling Set,
// WithField(s), WithDetail(s).
func (c *Container) IfSet(code codes.Code, format string, args ...interface{}) error {
	if c.errSet {
		c.errCode = code
		c.errMessage = fmt.Sprintf(format, args...)
		return c
	}

	return nil
}

// WithDetail function appends a new Detail to an error container's 'details'
// section.
func (c *Container) WithDetail(code codes.Code, target string, format string, args ...interface{}) *Container {
	c.errSet = true

	if c.details == nil {
		c.details = []*errdetails.TargetInfo{}
	}

	c.details = append(c.details, errdetails.Newf(code, target, format, args...))

	return c
}

// WithDetails function appends a list of error details to an error
// container's 'details' section.
func (c *Container) WithDetails(details ...*errdetails.TargetInfo) *Container {
	if len(details) == 0 {
		return c
	}

	c.errSet = true

	if c.details == nil {
		c.details = []*errdetails.TargetInfo{}
	}

	c.details = append(c.details, details...)
	return c
}

// WithField function appends a field error detail to an error container's
// 'fields' section.
func (c *Container) WithField(target string, format string, args ...interface{}) *Container {
	c.errSet = true

	if c.fields == nil {
		c.fields = &errfields.FieldInfo{}
	}

	c.fields.AddField(target, fmt.Sprintf(format, args...))
	return c
}

// WithFields function appends a several fields error details to an error
// container's 'fields' section.
func (c *Container) WithFields(fields map[string][]string) *Container {
	if len(fields) == 0 {
		return c
	}

	var hasDesc bool

	for k, v := range fields {
		for _, vVal := range v {
			if vVal != "" && k != "" {
				c.WithField(k, vVal)
				hasDesc = true
			}
		}
	}

	if hasDesc {
		c.errSet = true
	}

	return c
}
