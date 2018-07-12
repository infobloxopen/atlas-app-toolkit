# Atlas App Toolkit Error Handling

This document is a brief overview of facilities provided by error handling
package. The rationale for implementing it are four noble reasons:

1. Provide ability to add specific details and field information to an error.
2. Provide ability to handle multiple errors without returning control to a callee.
3. Ability to map errors from 3-rd party libraries (gorm, to name one).
4. Mapping from error to container should be performed automatically in gRPC interceptor.

## Error Container

Error container is a data structure that implements Error interface and 
GRPCStatus method, enabling passing it around as a conventional error from one
side and as a protobuf Status to gRPC gateway from the other side.

There are several approaches exist to work with it:

1. Single error mode
2. Multiple errors mode

### Single Error Return

This code snippet demonstrates the usage of error container as conventional
error:

```
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

```
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

```
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

```
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



```

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

```
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
