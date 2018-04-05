# NGP API Toolkit

1. [Getting Started](#getting-started)
    1. [Plugins](#plugins)
        1. [gRPC Protobuf](#grpc-protobuf)
        2. [gRPC Gateway](#grpc-gateway)
        3. [Middlewares](#middlewares)
        4. [Validation](#validation)
        5. [Documentation](#documentation)
        6. [Swagger](#swagger)
    2. [Build image](#build-image)
    3. [Kubernetes](#kubernetes)
2. [REST API Syntax Specification](#rest-api-syntax-specification)
    1. [Resources and Collections](#resources-and-collections)
    2. [HTTP Headers](#http-headers)
    3. [Responses](#responses)
    4. [Errors](#errors)
    5. [Collection Operators](#collection-operators)
        1. [Field Selection](#field-selection)
        2. [Sorting](#sorting)
        3. [Filtering](#filtering)
        4. [Pagination](#pagination)


## Getting Started

Before you get started please have a look at [this presentation](https://docs.google.com/presentation/d/1LuHrYp7E3KBVF4PcmNRLPrgsYtodsR3-R9odeQ2CJH0/edit#slide=id.p) and read through
[REST API Syntax Specification](https://docs.google.com/document/d/1gi4npvvaY_M1uP2i9LCmX8tOyvF6E5E7HAg1c9uAp_E/edit).

TL;DR

- Application may be composed from one or more independent services (micro-service architecture)
- Service is supposed to be a gRPC service
- REST API is presented by a separate service (gRPC Gateway) that serves as a reverse-proxy and
forwards incoming HTTP requests to gRPC services

### Plugins

NGP API Toolkit is not a framework it is a set of plugins for Google Protocol Buffer compiler.

#### gRPC Protobuf

See official documentation for [Protocol Buffer](https://developers.google.com/protocol-buffers/) and
for [gRPC](https://grpc.io/docs)

As an alternative you may use [this plugin](https://github.com/gogo/protobuf) to generate Golang code. That is the same
as official plugin but with [gadgets](https://github.com/gogo/protobuf/blob/master/extensions.md).

#### gRPC Gateway

See official [documentation](https://github.com/grpc-ecosystem/grpc-gateway)

#### Middlewares

The one of requirements for NGP API Toolkit was support of Pipeline model.
We recommend to use gRPC server interceptor as middleware. See [examples](https://github.com/grpc-ecosystem/go-grpc-middleware)

##### GetTenantID

We offer a convenient way to extract the TenantID field from an incoming authorization token. See the full description [here](https://docs.google.com/document/d/1kdCwdWoSfUbowkR6Ob2kW9a1a8PPWDPbMz69VXsigkM/edit#heading=h.1xm8o6fjrh57).

#### Validation
We recommend to use [this validation plugin](https://github.com/lyft/protoc-gen-validate) to generate
`Validate` method for your gRPC requests.

As an alternative you may use [this plugin](https://github.com/mwitkow/go-proto-validators) too.

Validation can be invoked "automatically" if you add [this](https://github.com/grpc-ecosystem/go-grpc-middleware/tree/master/validator) middleware as a gRPC server interceptor.

#### Documentation

We recommend to use [this plugin](https://github.com/pseudomuto/protoc-gen-doc) to generate documentation.

Documentation can be generated in different formats.

Here are several most used instructions used in documentation generation:

##### Leading comments

Leading comments can be used everywhere.

```proto
/**
 * This is a leading comment for a message
*/

message SomeMessage {
  // this is another leading comment
  string value = 1;
}
```

##### Trailing comments

Fields, Service Methods, Enum Values and Extensions support trailing comments.

```proto
enum MyEnum {
  DEFAULT = 0; // the default value
  OTHER   = 1; // the other value
}
```

##### Excluding comments

If you want to have some comment in your proto files, but don't want them to be part of the docs, you can simply prefix the comment with @exclude.

Example: include only the comment for the id field

```proto
/**
 * @exclude
 * This comment won't be rendered
 */
message ExcludedMessage {
  string id   = 1; // the id of this message.
  string name = 2; // @exclude the name of this message

  /* @exclude the value of this message. */
  int32 value = 3;
}
```

#### Swagger

Optionally you may generate [Swagger](https://swagger.io/) schema from your proto file.
To do so install [this plugin](https://github.com/grpc-ecosystem/grpc-gateway/tree/master/protoc-gen-swagger).

```sh
go get -u github.com/golang/protobuf/protoc-gen-go
```

Then invoke it as a plugin for Proto Compiler

```sh
protoc -I/usr/local/include -I. \
  -I$GOPATH/src \
  -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
  --swagger_out=logtostderr=true:. \
  path/to/your_service.proto
```

##### How to add Swagger definitions in my proto scheme?

```proto
import "protoc-gen-swagger/options/annotations.proto";

option (grpc.gateway.protoc_gen_swagger.options.openapiv2_swagger) = {
  info: {
    title: "My Service";
    version: "1.0";
  };
  schemes: HTTP;
  schemes: HTTPS;
  consumes: "application/json";
  produces: "application/json";
};

message MyMessage {
  option (grpc.gateway.protoc_gen_swagger.options.openapiv2_schema) = {
    external_docs: {
      url: "https://infoblox.com/docs/mymessage";
      description: "MyMessage description";
    }
};
```

For more Swagger options see [this scheme](https://github.com/grpc-ecosystem/grpc-gateway/blob/master/protoc-gen-swagger/options/openapiv2.proto)

See examples [addressbook](example/addressbook.proto) and [dnsconfig](example/dnsconfig.proto).
Here are generated Swagger schemes [addressbook](example/addressbook.swagger.json) and [dnsconfig](example/dnsconfig.swagger.json).

**NOTE** [Well Known Types](https://developers.google.com/protocol-buffers/docs/reference/google.protobuf) are
generated in a bit unusual way:

```json
    "protobufEmpty": {
      "type": "object",
      "description": "service Foo {\n      rpc Bar(google.protobuf.Empty) returns (google.protobuf.Empty);\n    }\n\nThe JSON representation for `Empty` is empty JSON object `{}`.",
      "title": "A generic empty message that you can re-use to avoid defining duplicated\nempty messages in your APIs. A typical example is to use it as the request\nor the response type of an API method. For instance:"
    },
```

### Build Image

See [README](gentool/README.md)

## Kubernetes

To make this example work in minikube you have to run the following command first:
```sh
make example-build
minikube start
eval $(minikube docker-env)
```

In case you do not have nginx already running on your minikube you can start it by running:
```sh
make nginx-up
```

After that:
```sh
make example-image
make example-up
```

Query addressbook application:
```sh
curl -k https://minikube/addressbook/v1/persons
```

Stop example application:
```sh
make example-down
```

## REST API Syntax Specification

All public REST API endpoints must follow [REST API Syntax Specification](https://docs.google.com/document/d/1gi4npvvaY_M1uP2i9LCmX8tOyvF6E5E7HAg1c9uAp_E/)

### Resources and Collections

#### How to define REST API Endpoints in my proto scheme?

You can map your gRPC service methods to one or more REST API endpoints.
See [this reference](https://cloud.google.com/service-management/reference/rpc/google.api#http) how to do it.

TL;DR

```proto
// It is possible to define multiple HTTP methods for one RPC by using
// the `additional_bindings` option. Example:
//
//     service Messaging {
//       rpc GetMessage(GetMessageRequest) returns (Message) {
//         option (google.api.http) = {
//           get: "/v1/messages/{message_id}"
//           additional_bindings {
//             get: "/v1/users/{user_id}/messages/{message_id}"
//           }
//         };
//       }
//     }
//     message GetMessageRequest {
//       string message_id = 1;
//       string user_id = 2;
//     }
//
//
// This enables the following two alternative HTTP JSON to RPC
// mappings:
//
// HTTP | RPC
// -----|-----
// `GET /v1/messages/123456` | `GetMessage(message_id: "123456")`
// `GET /v1/users/me/messages/123456` | `GetMessage(user_id: "me" message_id: "123456")`
```

### HTTP Headers

#### How are HTTP request headers mapped to gRPC client metadata?

[Answer](https://github.com/grpc-ecosystem/grpc-gateway/wiki/How-to-customize-your-gateway#mapping-from-http-request-headers-to-grpc-client-metadata)

#### How can I get HTTP request header on my gRPC service?

To extract headers from metadata all you need is to use
[FromIncomingContext](https://godoc.org/google.golang.org/grpc/metadata#FromIncomingContext) function

```golang
import (
    "context"

    "google.golang.org/grpc/metadata"
    "github.com/grpc-ecosystem/grpc-gateway/runtime"
)

func (s *myServiceImpl) MyMethod(ctx context.Context, req *MyRequest) (*MyResponse, error) {
    var userAgent string

    if md, ok := metadata.FromIncomingContext(ctx); ok {
        // Uppercase letters are automatically converted to lowercase, see metadata.New
        if u, ok [runtime.MetadataPrefix+"user-agent"]; ok {
            userAgen = u[0]
        }
    }
}
```

Also you can use our helper function `gw.Header()`

```golang
import (
    "context"

    "github.com/Infoblox-CTO/ngp.api.toolkit/gw"
)

func (s *myServiceImpl) MyMethod(ctx context.Context, req *MyRequest) (*MyResponse, error) {
    var userAgent string

    if h, ok := gw.Header(ctx, "user-agent"); ok {
        userAgent = h
    }
}
```

#### How can I send gRPC metadata?

To send metadata to gRPC-Gateway from your gRPC service you need to use [SetHeader](https://godoc.org/google.golang.org/grpc#SetHeader) function.

```golang
import (
    "context"

    "google.golang.org/grpc"
    "google.golang.org/grpc/metadata"
)

func (s *myServiceImpl) MyMethod(ctx context.Context, req *MyRequest) (*MyResponse, error) {
    md := metadata.Pairs("myheader", "myvalue")
    if err := grpc.SetHeader(ctx, md); err != nil {
        return nil, err
    }
    return nil, nil
}
```

If you do not use any custom outgoing header matcher you would see something like that:
```sh
> curl -i http://localhost:8080/addressbook/v1/persons

HTTP/1.1 200 OK
Content-Type: application/json
Grpc-Metadata-Myheader: myvalue
Date: Wed, 31 Jan 2018 15:28:52 GMT
Content-Length: 2

{}
```

### Responses

By default gRPC-Gateway translates non-error gRPC response into HTTP response
with status code set to `200 - OK`.

A HTTP response returned from gRPC-Gateway does not comform REST API Syntax
and has no `success` section.

In order to override this behavior gRPC-Gateway wiki recommends to overwrite
`ForwardResponseMessage` and `ForwardResponseStream` functions correspondingly.
See [this article](https://github.com/grpc-ecosystem/grpc-gateway/wiki/How-to-customize-your-gateway#replace-a-response-forwarder-per-method)

#### How can I overwrite default Forwarders?

See [example](example/addressbook/addressbook.overwrite.gw.pb.go).

#### Which forwarders I need to use to comply our REST API?

We made default [ForwardResponseMessage](gw/response.go#L36) and [ForwardResponseMessage](gw/response.go#L38)
implementations that conform REST API Syntax.

**NOTE** the forwarders still set `200 - OK` as HTTP status code if no errors encountered.

### How can I set 201/202/204/206 HTTP status codes?

In order to set HTTP status codes propely you need to send metadata from your
gRPC service so that default forwarders will be able to read them and set codes.
That is a common approach in gRPC to send extra information for response as
metadata.

We recommend use [gRPC status package](https://godoc.org/google.golang.org/grpc/status)
and our custom function [SetStatus](gw/status.go#L44) to add extra metadata
to the gRPC response.

See documentation in package [status](gw/status.go).

Also you may use shortcuts like: `SetCreated`, `SetUpdated` and `SetDeleted`.

```golang
import (
    "github.com/Infoblox-CTO/ngp.api.toolkit/gw"
)

func (s *myService) MyMethod(req *MyRequest) (*MyResponse, error) {
    err := gw.SetCreated(ctx, "created 1 item")
    return &MyResponse{Result: []*Item{item}}, err
}
```

See [example](example/addressbook/service.go#L64)

### Errors

#### How can I convert a gRPC error to a HTTP error response in accordance with REST API Syntax Specification?

You can write your own `ProtoErrorHandler` or use `gw.DefaultProtoErrorHandler` one.

How to handle error on gRPC-Gateway see [article](https://mycodesmells.com/post/grpc-gateway-error-handler)

How to use [gw.DefaultProtoErrorHandler](gw/errors.go#L25) see example below:

```golang
import (
    "github.com/grpc-ecosystem/grpc-gateway/runtime"
    "github.com/Infoblox-CTO/ngp.api.toolkit/gw"

    "github.com/Infoblox-CTO/yourapp"
)

func main() {
    // create error handler option
    errHandler := runtime.WithProtoErrorHandler(gw.DefaultProtoErrorHandler)

    // pass that option as a parameter
    mux := runtime.NewServeMux(errHandler)

    // register you app handler
    yourapp.RegisterAppHandlerFromEndpoint(ctx, mux, addr)

    ...

    // Profit!
}
```

You can find sample in example folder. See [code](example/cmd/gateway/main.go)

#### How can I send error with details from my gRPC service?

The  idiomatic way to send an error from you gRPC service is to simple return
it from you gRPC handler either as `status.Errorf()` or `errors.New()`.

```golang
import (
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

func (s *myServiceImpl) MyMethod(req *MyRequest) (*MyResponse, error) {
    return nil, status.Errorf(codes.Unimplemented, "method is not implemented: %v", req)
}
```

To attach details to your error you have to use `grpc/status` package.
You can use our default implementation of error details (`rpc/errdetails`) or your own one.

```golang
import (
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"

    "github.com/Infoblox-CTO/ngp.api.toolkit/rpc/errdetails"
)

func (s *myServiceImpl) MyMethod(req *MyRequest) (*MyResponse, error) {
    s := status.New(codes.Unimplemented, "MyMethod is not implemented")
    s = s.WithDetails(errdetails.New(codes.Internal), "myservice", "in progress")
    return nil, s.Err()
}
```

With `gw.DefaultProtoErrorHandler` enabled JSON response will look like:
```json
{
  "error": {
    "status": 501,
    "code": "NOT_IMPLEMENTED",
    "message": "MyMethod is not implemented"
  },
  "details": [
    {
      "code": "INTERNAL",
      "message": "in progress",
      "target": "myservice"
    }
  ]
}
```

You can find sample in example folder. See [code](example/addressbook/service.go:L72)

### Collection Operators

[REST API Syntax Specification](https://docs.google.com/document/d/1gi4npvvaY_M1uP2i9LCmX8tOyvF6E5E7HAg1c9uAp_E/edit#heading=h.jif2quft4mjc)
introduces a set of common request parameters that can be used to control
the way collections are returned. API toolkit provides some convenience methods
to support these parameters in your application.

#### How can I add support for collection operators in my gRPC-Gateway?

You can enable support of collection operators in your gRPC-Gateway by adding
a `runtime.ServeMuxOption` using `runtime.WithMetadata(gw.MetadataAnnotator)`.

```golang
import (
    "github.com/grpc-ecosystem/grpc-gateway/runtime"
    "github.com/Infoblox-CTO/ngp.api.toolkit/gw"

    "github.com/Infoblox-CTO/yourapp"
)

func main() {
    // create collection operator handler
    opHandler := runtime.WithMetadata(gw.MetadataAnnotator)

    // pass that option as a parameter
    mux := runtime.NewServeMux(opHandler)

    // register you app handler
    yourapp.RegisterAppHandlerFromEndpoint(ctx, mux, addr)

    ...
}
```

If you want to explicitly declare one of collection operators in your `proto`
scheme, to do so just import `collection_operators.proto`.

```proto
import "github.com/Infoblox-CTO/ngp.api.toolkit/op/collection_operators.proto";

message MyRequest {
    infoblox.api.Sorting sorting = 1;
}
```

After you declare one of collection operator in your `proto` message you need
to add `mw.WithCollectionOperator` server interceptor to the chain in your
gRPC service.

```golang
  server := grpc.NewServer(
    grpc.UnaryInterceptor(
      grpc_middleware.ChainUnaryServer( // middleware chain
        ...
        mw.WithCollectionOperator(), // collection operators
        ...
      ),
    ),
  )
```

Doing so all collection operators that defined in your proto message will be
populated in case if they provided in incoming HTTP request.

#### How can I apply collection operators passed to my GRPC service to a GORM query?

You can use `ApplyCollectionOperators` method from [op/gorm](op/gorm) package.

```golang
...
gormDB, err = ApplyCollectionOperators(gormDB, ctx)
if err != nil {
    ...
}
var people []Person
gormDB.Find(&people)
...
```

Separate methods per each collection operator are also available.

Check out [example](example/tagging/service.go) and [implementation](op/gorm/collection_operators.go).

#### Field Selection

A service may support field selection of collection data to reduce the volume of data in the result.
A collection of response resources can transformed by specifying a set of JSON tags to be returned.
[REST API Syntax Specification](https://docs.google.com/document/d/1gi4npvvaY_M1uP2i9LCmX8tOyvF6E5E7HAg1c9uAp_E/edit#heading=h.y734pt487r2u)
specifies that `_fields` request parameter that contains a comma-separated list of JSON tag names can be used for this purpose.
API toolkit provides a default support to strip fields in response. As it is not possible to completely remove all the fields
(such as primitives) from `proto.Message`. Because of this fields are additionally truncated on `grpc-gateway`. From gRPC it is also possible
to access `_fields` from request, use them to perform data fetch operations and control output. This can be done by setting
appropriate metadata keys that will be handled by `grpc-gateway`. See example below:

```
	fields := gw.FieldSelection(ctx)
	if fields != nil {
		// ... work with fields
		gw.SetFieldSelection(ctx, fields) //in case fields were changed comparing to what was in request
	}

```

You can find sample in example folder. See ListPersons implementation in the [code](example/addressbook/service.go:L33)

##### How to define field selection in my request?

```proto
import "github.com/Infoblox-CTO/ngp.api.toolkit/op/collection_operators.proto";

message MyRequest {
    infoblox.api.FieldSelection fields = 1;
}
```

#### Sorting

See [section](https://docs.google.com/document/d/1gi4npvvaY_M1uP2i9LCmX8tOyvF6E5E7HAg1c9uAp_E/edit#heading=h.usvz3nakezcg)
from REST API Syntax Specification.

##### How to define sorting in my request?

```proto
import "github.com/Infoblox-CTO/ngp.api.toolkit/op/collection_operators.proto";

message MyRequest {
    infoblox.api.Sorting sort = 1;
}
```

##### How can I get sorting operator on my gRPC service?

You may get it by using `gw.Sorting` function. Please note that if `_order_by`
has not been specified in an incoming HTTP request `gw.Sorting` returns `nil, nil`.

```golang
import (
    "context"

    "github.com/Infoblox-CTO/ngp.api.toolkit/gw"
    "github.com/Infoblox-CTO/ngp.api.toolkit/op"
)

func (s *myServiceImpl) MyMethod(ctx context.Context, req *MyRequest) (*MyResponse, error) {
    if sort, err := gw.Sorting(ctx); err != nil {
        return nil, err
    // check if sort has been specified!!!
    } else if sort != nil {
        // do sorting
        //
        // if you use gORM you may do the following
        // db.Order(sort.GoString())
    }
}
```

Also you may want to declare sorting parameter in your `proto` message.
In this case it will be populated automatically if you using
`mw.WithCollectionOperator` server interceptor.

See documentation in [op package](op/sorting.go)

#### Filtering

See [section](https://docs.google.com/document/d/1gi4npvvaY_M1uP2i9LCmX8tOyvF6E5E7HAg1c9uAp_E/edit#heading=h.lvxuni8pwk58)
from REST API Syntax Specification.

Usage of filtering features of the toolkit is similar to [sorting](#sorting). Check out [example](example/addressbook/service.go).

Note: if you decide to use toolkit provided `infoblox.api.Filtering` proto type, then you'll not be able to use swagger schema generation, since it's plugin doesn't work with recursive nature of `infoblox.api.Filtering`.

##### How to define filtering in my request?

```proto
import "github.com/Infoblox-CTO/ngp.api.toolkit/op/collection_operators.proto";

message MyRequest {
    infoblox.api.Filtering filter = 1;
}
```

#### Pagination

See [section](https://docs.google.com/document/d/1gi4npvvaY_M1uP2i9LCmX8tOyvF6E5E7HAg1c9uAp_E/edit#heading=h.u2ngqo8vu585)
from REST API Syntax Specification.

See [doc](op/pagination.go) for more details.

##### How to define pagination in my request/response?

```proto
import "github.com/Infoblox-CTO/ngp.api.toolkit/op/collection_operators.proto";

message MyRequest {
    infoblox.api.Pagination paging = 1;
}

message MyResponse {
  infoblox.api.PageInfo page = 1;
}
```
