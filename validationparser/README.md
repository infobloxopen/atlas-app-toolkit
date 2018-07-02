# validator

    import "github.com/infobloxopen/atlas-app-toolkit/errors/validator"

`validator` is a generic request contents validator server-side middleware for
gRPC.


### Request Validator Middleware

This middleware checks for the existence of a `Validate` method on each of the
messages of a gRPC request. In case of a
validation failure, an `InvalidArgument` gRPC status is returned, along with
the error that caused the validation failure.

While it is generic, it is intended to be used with plugins like 
https://github.com/mwitkow/go-proto-validators or https://github.com/lyft/protoc-gen-validate, Go protocol buffers codegen
plugins that create the `Validate` methods (including nested messages) based on declarative options in the `.proto` files themselves. 

## Usage

#### func  UnaryServerInterceptor

```go
func UnaryServerInterceptor() grpc.UnaryServerInterceptor
```
UnaryServerInterceptor returns a new unary server interceptor that validates
incoming messages.

Invalid messages will be rejected with `InvalidArgument` and the error before reaching any userspace handlers.


#### func  MapValidationError

```go
func MapValidationError() errors.MapFunc
```
MapValidationError returns a mapper that parses through the lyft protoc-gen-validate errors and only returns a user friendly error. 

Example return after MapValidationError on a invalid email: 

```json
{
    "error": {
        "status": 400,
        "code": "INVALID_ARGUMENT",
        "message": "Invalid PrimaryEmail: value must be a valid email address"
    },
    "fields": {
        "PrimaryEmail": [
            "value must be a valid email address"
        ]
    }
}
```