# Authorization Interceptor

This package provides a server-side gRPC interceptor that interfaces with Themis, a policy engine that is developed and maintained by Infoblox. It is designed to give developers fine-grained control over who (e.g. specific users) or what (e.g. neighboring services) can access their gRPC service's business logic.

## How it Works
When building a gRPC server, a developer can supply his or her server with a collection of _gRPC server-side interceptors_. Each request that's sent to the gRPC server will traverse the gRPC interceptors before reaching the server's business logic. At any point, an interceptor can stop a given request from advancing to the application.

Here are some common usages for gRPC interceptors.

- Logging (e.g. log to `stdout` on each request)
- Metrics (e.g. increment a request-counting metric)
- **Authorization** (e.g. check if the request-sender is allowed to access a given resource)

When the authorization interceptor is enabled, it ensures that whoever or whatever sent the request is allowed to access some endpoint. If they're not, the interceptor issues a `DENY` response, otherwise the request moves to the application itself.

Behold the behavior authorization interceptor as if it were a living, breathing human: _"Hey Themis, is the sender of request XYZ allowed to access endpoint ABC in my service? They're not? Okay, I'll deny the request!"_

## Add Interceptors
If you already use gRPC with Go, your project probably has a bit of code like this.

```go
// Define your grpc server
myServer := grpc.NewServer()
```
This is how you would decorate your server with interceptors.

```go
// Define your grpc server
myServer := grpc.NewServer(
  rpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
    // Define your list of grpc interceptors
  )),
)
```

For more information about adding server-side gRPC interceptors, check out [the documentation](https://github.com/grpc-ecosystem/go-grpc-middleware) on the official gRPC GitHub repository.

## Add the Default AuthZ Interceptor

The `auth` package offers a default authorization interceptor.
```go
themisAddress := "default.themis:5555" 
applicationID := "shopping-mall"
// Define your grpc server
myServer := grpc.NewServer(
  grpc.UnaryInterceptor(grpc_middleware.ChainStreamServer(
    // Include the authorization interceptor
    auth.UnaryServerInterceptor(themisAddress, applicationID),
  )),
)
```
In the example above, the `themisAddress` variable corresponds to the host and port of Themis. The `applicationID`  maps the interceptor to a specific application. In a microservices environment, you might have multiple services that, as a unit, compose an application (e.g. a "petstore" service, a "coffee shop" service, and a "cellphone kiosk" service might belong to a "shopping mall" application). Each service has its own access control logic, but they're all under the same application umbrella.

When you define your gRPC server, start by using the `auth` package's `UnaryServerInterceptor` interceptor. The `applicationID` associates the interceptor with a specific set of policies.


## Customize the AuthZ Interceptor
The `auth` package provides a set of options that offer fine-grained control of the authorization interceptor. Options affect how the authorization middleware interfaces with Themis. For example, you might need to permit or deny request using the fields in a JWT. Well, there's an option to handle this use case!

Interceptor options are your friend because they abstract Themis' API behind a set of well-documented functions. Below is a detailed description of each option.


```go
import (
	"github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc"
)
...
// Build a new authorizer and specify some options. The authorizer will
// create authorization middleware based on options that get provided
// by the user.
authorizer := Authorizer{
  PDPAddress: "themis.default",
  Bldr: NewBuilder(
    // This is where options are provided by the user
    WithRequest("shopping-mall"),
    WithJWT(nil),
  ),
  Hdlr: NewHandler(),
}

// Create an authorization function. The authorization function is responsible
// for checking with Themis to determine if a request should be permitted
// or denied.
authFunc := authorizer.AuthFunc()
myServer := grpc.NewServer(
  grpc.UnaryInterceptor(grpc_middleware.ChainStreamServer(
    // include the authorization interceptor
    grpc_auth.UnaryServerInterceptor(authFunc)
  )),
)
```

You'll likely find the Go documentation helpful, too.

```sh
godoc -http :6060
```
Visit [this link](http://localhost:6060/pkg/github.com/infobloxopen/atlas-app-toolkit/auth/) in your browser.

### `WithJWT`

For token-based authorization with JWT, the `WithJWT` option will prove useful. When a request reaches the authorization interceptor, it will include the full JWT payload in a given authorization request to Themis.

```json
{
   "name":"john doe",
   "occupation":"astronaut",
   "group":"admin"
}
```
Each of the fields in the above example would be sent to Themis as part of the authorization request.

### `WithCallback`

The `WithCallback` option allows for application-specific authorization behavior. Although the toolkit offers a robust set of authorization options, you might need specialized, non-generalizable authorization logic to satisfy your access control requirements. Enter the `WithCallback` option.

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
In the example callback above, each request to the Themis will include an two additional attributes: the city and country (although both are hardcoded). It shows how you can modify the authorization logic without making changes to the toolkit code.

### `WithRequest`

The `WithRequest` option will
