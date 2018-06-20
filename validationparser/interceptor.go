package validationparser

import (
	"context"
	"reflect"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type validator interface {
	Validate() error
}

// UnaryServerInterceptor returns grpc.UnaryServerInterceptor
// that should be used as a middleware that parser through the protoc-gen-validate middleware and returns a
// user friendly error.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if v, ok := req.(validator); ok {
			if err := v.Validate(); err != nil {
				// Check what type of error we received
				typeOfError := reflect.TypeOf(err)
				valueOfError := reflect.ValueOf(err)
				// If the error is a type of validation error then parse
				if strings.Contains(typeOfError.String(), "ValidationError") && valueOfError.NumField() == 4 {
					// Get the cause of error from the error struct
					cause := valueOfError.FieldByName("Cause")
					typeOfCause := reflect.TypeOf(cause.Interface())
					valueOfCause := reflect.ValueOf(cause.Interface())
					if strings.Contains(typeOfCause.String(), "ValidationError") && valueOfCause.NumField() == 4 {
						// Retrieve the field and reason from the error
						field := valueOfCause.FieldByName("Field")
						reason := valueOfCause.FieldByName("Reason")
						st := status.Newf(codes.InvalidArgument, field.String()+" "+reason.String())
						return nil, st.Err()
					}
				}
				return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
			}
		}
		return handler(ctx, req)
	}
}
