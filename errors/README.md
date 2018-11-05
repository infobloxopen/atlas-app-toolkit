
# Errors
To avoid burden of mapping errors returned from 3rd party libraries
you can gather all error mappings in one place and put an interceptor
provided by atlas-app-toolkit package in a middleware chain as following:
```go
interceptor := errors.UnaryServerInterceptor(
	// List of mappings

	// Base case: simply map error to an error container.
	errors.NewMapping(fmt.Errorf("Some Error"), errors.NewContainer(/* ... */).WithDetail(/* ... */)),
)
```

## Contents
1. <a href="#background">Background</a>
1. <a href="#mappers">Error Mappers</a>
1. <a href="#usage">Usage</a>
1. <a href="#validation">Validation Errors</a>
1. <a href="#pqerrors">PQ Errors</a>

<a name="intro"></a>

# Error Handling
<a name="background"></a>
## Background

This document is a brief overview of facilities provided by error handling
package. The rationale for implementing it are four noble reasons:

1. Provide ability to add specific details and field information to an error.
1. Provide ability to handle multiple errors without returning control to a callee.
1. Ability to map errors from 3-rd party libraries (gorm, to name one).
1. Mapping from error to container should be performed automatically in gRPC interceptor.

<a name="mappers"></a>
## Error Mappers

Error mapper performs conditional mapping from one error message to another. Error mapping functions are passed to a gRPC Error interceptor and called against error returned from handler.

Currently there are two mappers available: 
- ValidationErrors: error mapper and interceptor for protoc-gen-validate validation errors
- PQErrors:  error mapper for go postgres driver 

## Error Container

Error container is a data structure that implements Error interface and 
GRPCStatus method, enabling passing it around as a conventional error from one
side and as a protobuf Status to gRPC gateway from the other side.

There are several approaches exist to work with it:

1. Single error mode
1. Multiple errors mode


## Usage

### Single Error Return

This code snippet demonstrates the usage of error container as conventional
error:

```go
func validateNameLength(name string) error {
	if len(name) > 255 {
		return errors.NewContainer(
			codes.InvalidArgument, "Name validation error."
		).WithDetail(
			codes.InvalidArgument, "object", "Invalid name length."
		).WithField(
			"name", "Specify name with length less than 255.")
	}

	return nil
}
```

### Gather Multiple Errors

```go
func (svc *Service) validateName(name string) error {
	err := errors.InitContainer()
	if len(name) > 255 {
		err = err.WithDetail(codes.InvalidArgument, "object", "Invalid name length.")
		err = err.WithField("name", "Specify name with length less than 255.")
	}

	if strings.HasPrefix(name, "_") {
		err = err.WithDetail(codes.InvalidArgument, "object", "Invalid name.").WithField(
			"name", "Name cannot start with an underscore")
	}

	return err.IfSet(codes.InvalidArgument, "Invalid name.")
}
```

To gather multiple errors across several procedures use context functions:

```go
func (svc *Service) globalValidate(ctx context.Context, input *pb.Input) error {
	svc.validateName(ctx, input.Name)
	svc.validateIP(ctx, input.IP)

	// in some particular cases we expect that something really bad
	// should happen, so we can analyse it and throw instead of validation errors.
	if err := validateNamePermExternal(svc.AuthInfo); err != nil {
		return errors.New(ctx, codes.Unauthorized, "Client is not authorized.").
			WithDetails(/* ... */)
	}

	return errors.IfSet(ctx, codes.InvalidArgument, "Overall validation failed.")
	// Alternatively if we want to return the latest errCode/errMessage set instead
	// of overwriting it:
	// return errors.Error(ctx)
}

func (svc *Service) validateName(ctx context.Context, name string) {
	if len(name) > 255 {
		errors.Detail(ctx, codes.InvalidArgument, "object", "Invalid name length.")
		errors.Field(ctx, "name", "Specify name with length less than 255.")
	}
}

func (svc *Service) validateIP(ctx context.Context, ip string) { /* ip validation */ }
```

## Error Mapper

Error mapper performs conditional mapping from one error message to another.
Error mapping functions are passed to a gRPC Error interceptor and called against
error returned from handler.

Below we demonstrate a cases and customization techniques for mapping functions:

```go
interceptor := errors.UnaryServerInterceptor(
	// List of mappings
	
	// Base case: simply map error to an error container.
	errors.NewMapping(fmt.Errorf("Some Error"), errors.NewContainer(/* ... */).WithDetail(/* ... */)),

	// Extended Condition, mapped if error message contains "fk_contraint" or starts with "pg_sql:"
	// MapFunc calls logger and returns Internal Error, depending on debug mode it could add details.
	errors.NewMapping(
		errors.CondOr(
			errors.CondReMatch("fk_constraint"),
			errors.CondHasPrefix("pg_sql:"),
		),
		errors.MapFunc(func (ctx context.Context, err error) (error, bool) {
			logger.FromContext(ctx).Debug(fmt.Sprintf("DB Error: %v", err))
			err := errors.NewContainer(codes.Internal, "database error")

			if InDebugMode(ctx) {
				err.WithDetail(/* ... */)
			}

			// boolean flag indicates whether the mapping succeeded
			// it can be used to emulate fallthrough behavior (setting false)
			return err, true
		})
	),

	// Skip error
	errors.NewMapping(fmt.Errorf("Error to Skip"), nil)
)
```

Such model allows us to define our own error classes and map them appropriately as in
example below:



```go
// service validation code.

type RequiredFieldErr string
(e RequiredFieldErr) Error() { return string(e) }

func RequiredFieldCond() errors.MapCond {
	return errors.MapCond(func(err error) bool {
		_, ok := err.(RequiredFieldErr)
		return ok
	})
}

func validateReqArgs(in *pb.Input) error {
	if in.Name == "" {
		return RequiredFieldErr("name")
	}

	return nil
}
```

```go
// interceptor init code
interceptor := errors.UnaryServerInterceptor(
	errors.NewMapping(
		RequiredFieldCond(),
		errors.MapFunc(func(ctx context.Context, err error) (error, bool) {
			return errors.NewContainer(
				codes.InvalidArgument, "Required field missing: %v", err
			).WithField(string(err), "%q argument is required.", string(err)
			), true
		}),
	)
)
```


<a name="validation"></a>
# Validation Errors

    import "github.com/infobloxopen/atlas-app-toolkit/errors/mappers/validationerrors"

`validationerrors` is a request contents validator server-side middleware for
gRPC.


### Request Validator Middleware

This middleware checks for the existence of a `Validate` method on each of the
messages of a gRPC request. In case of a
validation failure, an `InvalidArgument` gRPC status is returned, along with
the error that caused the validation failure.

It is intended to be used with plugins like https://github.com/lyft/protoc-gen-validate, Go protocol buffers codegen
plugins that create the `Validate` methods (including nested messages) based on declarative options in the `.proto` files themselves. 


#### func  UnaryServerInterceptor

```go
func UnaryServerInterceptor() grpc.UnaryServerInterceptor
```
UnaryServerInterceptor returns a new unary server interceptor that validates
incoming messages and returns a ValidationError.

Invalid messages will be rejected with `InvalidArgument` and the error before reaching any userspace handlers.


#### func  DefaultMapping

```go
func DefaultMapping() errors.MapFunc
```
DefaultMapping returns a mapper that parses through the lyft protoc-gen-validate errors and only returns a user friendly error. 

Example Usage: 

1. Add validationerrors and errors interceptors to your application:

    ```go
    errors.UnaryServerInterceptor(ErrorMappings...),
    validationerrors.UnaryServerInterceptor(),
    ```

1. Create an ErrorMapping variable with all your mappings. 
1. Add DefaultMapping as part of your ErrorMapping variable

     ```go
    var ErrorMappings = []errors.MapFunc{
        // Adding Default Validations Mapping
        validationerrors.DefaultMapping(), 

    }
    ```


    Example return after DefaultMapping on a invalid email: 

    ```json
    {
        "error": {
            "status": 400,
            "code": "INVALID_ARGUMENT",
            "message": "Invalid primary_email: value must be a valid email address"
        },
        "fields": {
            "primary_email": [
                "value must be a valid email address"
            ]
        }
    }
    ```

1. You can also add custom validation mappings:

    ```go
    var ErrorMappings = []errors.MapFunc{
        // Adding custom Validation Mapping based on the known field and reason from lyft
       errors.NewMapping(
			errors.CondAnd(
				validationerrors.CondValidation(),
                validationerrors.CondFieldEq("primary_email"),
				validationerrors.CondReasonEq("value must be a valid email address"),
			),
			errors.MapFunc(func(ctx context.Context, err error) (error, bool) {
				vErr, _ := err.(validationerrors.ValidationError)
				return errors.NewContainer(codes.InvalidArgument, "Custom error message for field: %v reason: %v", vErr.Field, vErr.Reason), true
            }),
       ),
    }

    ```

<a name="pqerrors"></a>

# PQ Errors

    import "github.com/infobloxopen/atlas-app-toolkit/errors/mappers/pqerrors"

`pqerrors` is a error mapper for postgres.



Dedicated error mapper for go postgres driver (lib/pq.Error) package is included under
the path of github.com/atlas-app-toolkit/errors/mappers/pqerrors. This package includes
following components:

 * Condition function CondPQ, CondConstraintEq, CondCodeEq for conditions involved in \*pq.Error detection,
specific constraint name and specific status code of postgres error respectively.

 * ToMapFunc function that converts mapping function that deals with pq.Error to avoid burden of
casting errors back and forth.

 * Default mapping function that can be included into errors interceptor for FK contraints (NewForeignKeyMapping),
RESTRICT (NewRestrictMapping), NOT NULL (NewNotNullMapping), PK/UNIQUE (NewUniqueMapping)

Example Usage: 

```go
import (
	...
	"github.com/atlas-app-toolkit/errors/mappers/pqerrors"
)

interceptor := errors.UnaryServerInterceptor(
	...
	pqerrors.NewUniqueMapping("emails_address_key", "Contacts", "Primary Email Address"),
	...
)
```

Any violation of UNIQUE constraint "email_address_key" will result in following error:


```json
{
  "error": {
    "status": 409,
    "code": "ALREADY_EXISTS",
    "message": "There is already an existing 'Contacts' object with the same 'Primary Email Address'."
  }
}
```