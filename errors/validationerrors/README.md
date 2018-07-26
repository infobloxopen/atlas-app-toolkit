# Validation Errors

    import "github.com/infobloxopen/atlas-app-toolkit/errors/validationerrors"

`validationerrors` is a request contents validator server-side middleware for
gRPC.


### Request Validator Middleware

This middleware checks for the existence of a `Validate` method on each of the
messages of a gRPC request. In case of a
validation failure, an `InvalidArgument` gRPC status is returned, along with
the error that caused the validation failure.

It is intended to be used with plugins like https://github.com/lyft/protoc-gen-validate, Go protocol buffers codegen
plugins that create the `Validate` methods (including nested messages) based on declarative options in the `.proto` files themselves. 

## Usage

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

2. Create an ErrorMapping variable with all your mappings. 
3. Add DefaultMapping as part of your ErrorMapping variable

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

4. You can also add custom validation mappings:

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


For actual example usage look at:
https://github.com/infobloxopen/atlas-contacts-app