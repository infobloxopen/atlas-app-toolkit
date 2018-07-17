# Authorization Interceptor

This package provides a server-side gRPC interceptor that interfaces with [Themis](https://github.com/infobloxopen/themis), a policy engine that is developed and maintained by Infoblox. 

It is designed to give developers fine-grained control over who (e.g. specific users) or what (e.g. neighboring services) can access their gRPC service's business logic.

## Background

If you're unfamiliar with gRPC interceptors and their intended use, please consider reading [this](https://github.com/grpc-ecosystem/go-grpc-middleware#middleware) brief explanation.

The authorization interceptor determines whether or not an API consumer can access an endpoint. If the API consumer does not have appropriate permissions to access an endpoint (e.g. the consumer does not provide an API key), the interceptor will stop their request from advancing to the application.

Here's the authorization interceptor as if it were a living, breathing human: _"Hey Themis, is the sender of request XYZ allowed to access endpoint ABC in my service? They're not? Okay, I'll deny the request!"_

## Using Interceptors
If you already use gRPC with Go, your project probably has a bit of code like this.

```go
// define your grpc server
myServer := grpc.NewServer()
```
This is how you would decorate your server with interceptors.

```go
// define your grpc server
myServer := grpc.NewServer(
  grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
    // define your list of grpc interceptors
  )),
)
```

Again, the official gRPC GitHub repository [has documentation](https://github.com/grpc-ecosystem/go-grpc-middleware) to explain this process.

## Using the Default AuthZ Interceptor

The `auth` package offers a default authorization interceptor.
```go
themisAddress := "default.themis:5555" 
applicationID := "shopping-mall"
// define your grpc server
myServer := grpc.NewServer(
  grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
    // include the authorization interceptor
    auth.UnaryServerInterceptor(themisAddress, applicationID),
  )),
)
```
In the example above, the `themisAddress` variable corresponds to the host and port of Themis.


The `applicationID` maps the interceptor to a specific application. In a microservices environment, you might have multiple services that, as a unit, compose an application (e.g. a petstore service, a coffee shop service, and a cellphone kiosk service might belong to a shopping mall application). Each service has its own access control logic, but together they form an application.

When you define your gRPC server, start by using the `auth` package's `UnaryServerInterceptor` interceptor. The `applicationID` associates the interceptor with a specific set of policies.


## Customize the AuthZ Interceptor
The `auth` package provides a set of _options_ that offer fine-grained control of the authorization interceptor. Options affect how the authorization middleware interfaces with Themis. For example, you might need to permit or deny request using the fields in a JWT. Well, there's an option to handle this use case!

Interceptor options are helpful because they abstract Themis' API behind a set of well-documented functions.

### Applying Options

The example below shows how to would configure your application's authorization logic using options.

```go
import (
	"github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc"
)
...
// build a new authorizer and specify some options. the authorizer will
// create authorization middleware based on options that get provided
// by the user.
authorizer := Authorizer{
  PDPAddress: "themis.default:5555",
  Bldr: NewBuilder(
    // this is where options are provided by the user
    WithRequest("shopping-mall"),
    WithJWT(nil),
  ),
  Hdlr: NewHandler(),
}

// create an authorization function. the authorization function is responsible
// for checking with themis to determine if a request should be permitted
// or denied.
authFunc := authorizer.AuthFunc()
myServer := grpc.NewServer(
  grpc.UnaryInterceptor(grpc_middleware.ChainStreamServer(
    // include the authorization interceptor
    grpc_auth.UnaryServerInterceptor(authFunc)
  )),
)
```

This readme has a description of each option below, but you can also host a local GoDoc server.
```sh
godoc -http :6060
open http://localhost:6060/pkg/github.com/infobloxopen/atlas-app-toolkit/auth
```

### Interceptor Options

Here are the authorization options that are currently provided by the toolkit. 

#### `WithJWT`

This option enables token-based authorization with JWT. When requests reach the authorization interceptor, the interceptor will include the full JWT payload in a given authorization request to Themis.

```json
{
   "name":"john doe",
   "occupation":"astronaut",
   "group":"admin"
}
```
Each of the fields in the above example would be sent to Themis as part of the authorization request.

#### `WithRequest`

This option includes information about the gRPC request as part of the request to Themis. The interceptor will add the following attributes.

- The gRPC service name (e.g. `/PetStore`)
- The gRPC function name (e.g. `ListAllPets`)

#### `WithTLS`

This option uses metadata from a TLS-authenticated client. When included, the following options are included in a request to Themis.

- Whether or not the request-sender is TLS authenticated
- The TLS certificate issuer (if authenticated)
- The TLS common subject name (if authenticated)

#### `WithCallback`

This option allows for application-specific authorization behavior. The toolkit offers a robust set of authorization options, but you might need specialized, non-generalizable authorization logic to satisfy your access control requirements.

To use the `WithCallback` option, you must be somewhat familiar with the API for Themis, which is defined [here](https://github.com/infobloxopen/themis/blob/master/proto/service.proto).

```go
myCallBackFunction := func(ctx context.Context) ([]*pdp.Attribute, error){
  myAttributes := []*pdp.Attribute{
    {Id: "city", Type: "string", Value: "vancouver"},
    {Id: "country", Type: "string", Value: "canada"},
  }
  return myAttributes, nil
}
```
Each request to the Themis will include an two additional attributes: the city and country (although both are hardcoded). The point is, you can modify the authorization logic without making changes to the toolkit code.
