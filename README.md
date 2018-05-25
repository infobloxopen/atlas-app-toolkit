# API Toolkit

[![Build Status](https://img.shields.io/travis/infobloxopen/atlas-app-toolkit/master.svg?label=build)](https://travis-ci.org/infobloxopen/atlas-app-toolkit)

1. [Getting Started](#getting-started)
    1. [Plugins](#plugins)
        1. [gRPC Protobuf](#grpc-protobuf)
        2. [gRPC Gateway](#grpc-gateway)
        3. [Middlewares](#middlewares)
        4. [Validation](#validation)
        5. [Documentation](#documentation)
        6. [Swagger](#swagger)
    2. [Build image](#build-image)
    3. [Server Wrapper](#server-wrapper)
    4. [Example](#example)
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

Toolkit provides a means to have a generated code that supports a certain common functionality that 
is typicall requesred for any service.
Toolkit declares its own format for Resposes, Errors, Long Running operations, Collection operators.
More details on this can be found in appropriate section on this page.

Tollkit approach provides following features:
- Application may be composed from one or more independent services (micro-service architecture)
- Service is supposed to be a gRPC service
- REST API is presented by a separate service (gRPC Gateway) that serves as a reverse-proxy and
forwards incoming HTTP requests to gRPC services

### Initializing your Application

To get started with the toolkit, check out the [Atlas CLI](https://github.com/infobloxopen/atlas-cli) repository. The Atlas CLI's "bootstrap command" can generate new applications that make use of the toolkit. For Atlas newcomers, this is a great way to get up-to-speed with toolkit best practices.

### Plugins

API Toolkit is not a framework it is a set of plugins for Google Protocol Buffer compiler.

#### gRPC Protobuf

See official documentation for [Protocol Buffer](https://developers.google.com/protocol-buffers/) and
for [gRPC](https://grpc.io/docs)

As an alternative you may use [this plugin](https://github.com/gogo/protobuf) to generate Golang code. That is the same
as official plugin but with [gadgets](https://github.com/gogo/protobuf/blob/master/extensions.md).

#### gRPC Gateway

See official [documentation](https://github.com/grpc-ecosystem/grpc-gateway)

#### Middlewares

One of the requirements to the API Toolkit is to support a Pipeline model.
We recommend to use gRPC server interceptor as middleware. See [examples](https://github.com/grpc-ecosystem/go-grpc-middleware)

##### GetAccountID

We offer a convenient way to extract the AccountID field from an incoming authorization token.
For this purpose `auth.GetAccountID(ctx, nil)` function can be used:
```
func (s *contactsServer) Read(ctx context.Context, req *ReadRequest) (*ReadResponse, error) {
	input := req.GetContact()

	accountID, err := auth.GetAccountID(ctx, nil)
	if err == nil {
		input.AccountId = accountID
	} else if input.GetAccountId() == "" {
		return nil, err
	}

	c, err := DefaultReadContact(ctx, input, s.db)
	if err != nil {
		return nil, err
	}
	return &ReadResponse{Contact: c}, nil
}
```

When bootstrapping a gRPC server, add middleware that will extract the account_id token from the request context and set it in the request struct. The middleware will have to navigate the request struct via reflection, in the case that the account_id field is nested within the request (like if it's in a request wrapper as per our example above)

##### Transaction Management

We provide transaction management by offering `gorm.Transaction` wrapper and `gorm.UnaryServerInterceptor`.
The `gorm.Transaction` works as a singleton to prevent an application of creating more than one transaction instance per incoming request.
The `gorm.UnaryServerInterceptor` performs management on transactions. 
Interceptor creates new transaction on each incoming request and commits it if request finishes without error, otherwise transaction is aborted.
The created transaction is stored in `context.Context` and passed to the request handler as usual.
**NOTE** Client is responsible to call `txn.Begin()` to open transaction.

```go
// add gorm interceptor to the chain
  server := grpc.NewServer(
    grpc.UnaryInterceptor(
      grpc_middleware.ChainUnaryServer( // middleware chain
        ...
        gorm.UnaryServerInterceptor(), // transaction management
        ...
      ),
    ),
  )
```

```go
import (
	"github.com/infobloxopen/atlas-app-toolkit/gorm"
)

func (s *MyService) MyMethod(ctx context.Context, req *MyMethodRequest) (*MyMethodResponse, error) {
	// extract gorm transaction from context
	txn, ok := gorm.FromContext(ctx)
	if !ok {
		return panic("transaction is not opened") // don't panic in production!
	}
	// start transaction
	gormDB := txn.Begin()
	if err := gormDB.Error; err != nil {
		return nil, err
	}
	// do stuff with *gorm.DB
	return &MyMethodResponse{...}, nil
}
```

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

See example [contacts app](https://github.com/infobloxopen/atlas-contacts-app/blob/master/proto/contacts.proto).
Here is a [generated Swagger schema](https://github.com/infobloxopen/atlas-contacts-app/blob/master/proto/contacts.swagger.json).

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

For convenience purposes there is an atlas-gentool image available which contains a pre-installed set of often used plugins.
For more details see [infobloxopen/atlas-gentool](https://github.com/infobloxopen/atlas-gentool) repository.

### Server Wrapper

You can package your gRPC server along with your REST gateway, health checks and any other http endpoints using [`server.NewServer`](server/server.go):
```go
s, err := server.NewServer(
    server.WithGrpcServer(grpcServer),
    server.WithHealthChecks(healthChecks),
    server.WithGateway(
        gateway.WithEndpointRegistration("/v1/", server_test.RegisterHelloHandlerFromEndpoint),
        gateway.WithServerAddress(grpcL.Addr().String()),
    ),
)
if err != nil {
    log.Fatal(err)
}
// serve it by passing in net.Listeners for the respective servers
if err := s.Serve(grpcListener, httpListener); err != nil {
    log.Fatal(err)
}
```
You can see a full example [here](server/server_example_test.go).

## Example

An example app that is based on api-toolkit can be found [here](https://github.com/infobloxopen/atlas-contacts-app)

## REST API Syntax Specification

Toolkit enforces some of the API syntax requirements that are common for
applications that are written by Infoblox. All public REST API endpoints must follow the same guidelines mentioned below.

### Resources and Collections

#### How to define REST API Endpoints in my proto scheme?

You can map your gRPC service methods to one or more REST API endpoints.
See [this reference](https://cloud.google.com/service-management/reference/rpc/google.api#http) how to do it.

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

Also you can use our helper function `gateway.Header()`

```golang
import (
    "context"

    "github.com/infobloxopen/atlas-app-toolkit/gateway"
)

func (s *myServiceImpl) MyMethod(ctx context.Context, req *MyRequest) (*MyResponse, error) {
    var userAgent string

    if h, ok := gateway.Header(ctx, "user-agent"); ok {
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
> curl -i http://localhost:8080/contacts/v1/contacts

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

```
import (
	"github.com/infobloxopen/atlas-app-toolkit/gw"
)

func init() {
	forward_App_ListObjects_0 = gateway.ForwardResponseMessage
}
```

You can also refer [example app](https://github.com/github.com/infobloxopen/atlas-contacts-app/pb/contacts/contacts.overwrite.pb.gw.go)

#### Which forwarders I need to use to comply our REST API?

We made default [ForwardResponseMessage](gateway/response.go#L36) and [ForwardResponseMessage](gateway/response.go#L38)
implementations that conform REST API Syntax.

**NOTE** the forwarders still set `200 - OK` as HTTP status code if no errors encountered.

### How can I set 201/202/204/206 HTTP status codes?

In order to set HTTP status codes propely you need to send metadata from your
gRPC service so that default forwarders will be able to read them and set codes.
That is a common approach in gRPC to send extra information for response as
metadata.

We recommend use [gRPC status package](https://godoc.org/google.golang.org/grpc/status)
and our custom function [SetStatus](gateway/status.go#L44) to add extra metadata
to the gRPC response.

See documentation in package [status](gateway/status.go).

Also you may use shortcuts like: `SetCreated`, `SetUpdated` and `SetDeleted`.

```golang
import (
    "github.com/infobloxopen/atlas-app-toolkit/gateway"
)

func (s *myService) MyMethod(req *MyRequest) (*MyResponse, error) {
    err := gateway.SetCreated(ctx, "created 1 item")
    return &MyResponse{Result: []*Item{item}}, err
}
```

### Response format
Services render resources in responses in JSON format by default unless another format is specified in the request Accept header that the service supports.

Services must embed their response in a Success JSON structure.

The Success JSON structure provides a uniform structure for expressing normal responses using a structure similar to the Error JSON structure used to render errors. The structure provides an enumerated set of codes and associated HTTP statuses (see Errors below) along with a message.

The Success JSON structure has the following format. The results tag is optional and appears when the response contains one or more resources.
```
{
  "success": {
    "status": <http-status-code>,
    "code": <enumerated-error-code>,
    "message": <message-text>
  },
  "results": <service-response>
}
```

### Errors

#### Format
Method error responses are rendered in the Error JSON format. The Error JSON format is similar to the Success JSON format for error responses using a structure similar to the Success JSON structure for consistency.

The Error JSON structure has the following format. The details tag is optional and appears when the service provides more details about the error.
```
{
  "error": {
    "status": <http-status-code>,
    "code": <enumerated-error-code>,
    "message": <message-text>
  },
  "details": [
    {
      "message": <message-text>,
      "code": <enumerated-error-code>,
      "target": <resource-name>,
    },
    ...
  ]
}
```

#### How can I convert a gRPC error to a HTTP error response in accordance with REST API Syntax Specification?

You can write your own `ProtoErrorHandler` or use `gateway.DefaultProtoErrorHandler` one.

How to handle error on gRPC-Gateway see [article](https://mycodesmells.com/post/grpc-gateway-error-handler)

How to use [gateway.DefaultProtoErrorHandler](gateway/errors.go#L25) see example below:

```golang
import (
    "github.com/grpc-ecosystem/grpc-gateway/runtime"
    "github.com/infobloxopen/atlas-app-toolkit/gw"

    "github.com/yourrepo/yourapp"
)

func main() {
    // create error handler option
    errHandler := runtime.WithProtoErrorHandler(gateway.DefaultProtoErrorHandler)

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

    "github.com/infobloxopen/atlas-app-toolkit/rpc/errdetails"
)

func (s *myServiceImpl) MyMethod(req *MyRequest) (*MyResponse, error) {
    s := status.New(codes.Unimplemented, "MyMethod is not implemented")
    s = s.WithDetails(errdetails.New(codes.Internal), "myservice", "in progress")
    return nil, s.Err()
}
```

With `gateway.DefaultProtoErrorHandler` enabled JSON response will look like:
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

### Collection Operators

For methods that return collections, operations may be implemented using the following conventions.
The operations are implied by request parameters in query strings.
 In some cases, stateful operational information may be passed in responses.
Toolkit introduces a set of common request parameters that can be used to control
the way collections are returned. API toolkit provides some convenience methods
to support these parameters in your application.

#### How can I add support for collection operators in my gRPC-Gateway?

You can enable support of collection operators in your gRPC-Gateway by adding
a `runtime.ServeMuxOption` using `runtime.WithMetadata(gateway.MetadataAnnotator)`.

```golang
import (
    "github.com/grpc-ecosystem/grpc-gateway/runtime"
    "github.com/infobloxopen/atlas-app-toolkit/gateway"

    "github.com/yourrepo/yourapp"
)

func main() {
    // create collection operator handler
    opHandler := runtime.WithMetadata(gateway.MetadataAnnotator)

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
import "github.com/infobloxopen/atlas-app-toolkit/op/collection_operators.proto";

message MyRequest {
    infoblox.api.Sorting sorting = 1;
}
```

After you declare one of collection operator in your `proto` message you need
to add `gateway.UnaryServerInterceptor` server interceptor to the chain in your
gRPC service.

```golang
  server := grpc.NewServer(
    grpc.UnaryInterceptor(
      grpc_middleware.ChainUnaryServer( // middleware chain
        ...
        gateway.UnaryServerInterceptor(), // collection operators
        ...
      ),
    ),
  )
```

Doing so all collection operators that defined in your proto message will be
populated in case if they provided in incoming HTTP request.

#### How can I apply collection operators passed to my GRPC service to a GORM query?

You can use `ApplyCollectionOperators` method from [op/gorm](gateway/gorm) package.

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

Check out [example](example/tagging/service.go) and [implementation](gateway/gorm/collection_operators.go).

#### Field Selection


A service may implement field selection of collection data to reduce the volume of data in the result. A collection of response resources can be transformed by specifying a set of JSON tags to be returned. For a “flat” resource, the tag name is straightforward. If field selection is allowed on non-flat hierarchical resources, the service should implement a qualified naming scheme such as dot-qualification to reference data down the hierarchy. If a resource does not have the specified tag, the tag does not appear in the output resource.

| Request Parameter | Description                              |
| ----------------- |------------------------------------------|
| _fields           | A comma-separated list of JSON tag names.|

API toolkit provides a default support to strip fields in response. As it is not possible to completely remove all the fields
(such as primitives) from `proto.Message`. Because of this fields are additionally truncated on `grpc-gateway`. From gRPC it is also possible
to access `_fields` from request, use them to perform data fetch operations and control output. This can be done by setting
appropriate metadata keys that will be handled by `grpc-gateway`. See example below:

```
	fields := gateway.FieldSelection(ctx)
	if fields != nil {
		// ... work with fields
		gateway.SetFieldSelection(ctx, fields) //in case fields were changed comparing to what was in request
	}

```

##### How to define field selection in my request?

```proto
import github.com/infobloxopen/atlas-app-toolkit/query/collection_operators.proto;

message MyRequest {
    infoblox.api.FieldSelection fields = 1;
}
```

#### Sorting

A service may implement collection sorting. A collection of response resources can be sorted by their JSON tags. For a “flat” resource, the tag name is straightforward. If sorting is allowed on non-flat hierarchical resources, the service should implement a qualified naming scheme such as dot-qualification to reference data down the hierarchy. If a resource does not have the specified tag, its value is assumed to be null.

| Request Parameter | Description                              |
| ----------------- |------------------------------------------|
| _order_by         | A comma-separated list of JSON tag names. The sort direction can be specified by a suffix separated by whitespace before the tag name. The suffix “asc” sorts the data in ascending order. The suffix “desc” sorts the data in descending order. If no suffix is specified the data is sorted in ascending order. |

##### How to define sorting in my request?

```proto
import github.com/infobloxopen/atlas-app-toolkit/query/collection_operators.proto;

message MyRequest {
    infoblox.api.Sorting sort = 1;
}
```

##### How can I get sorting operator on my gRPC service?

You may get it by using `gateway.Sorting` function. Please note that if `_order_by`
has not been specified in an incoming HTTP request `gateway.Sorting` returns `nil, nil`.

```golang
import (
    "context"

    "github.com/infobloxopen/atlas-app-toolkit/gateway"
    "github.com/infobloxopen/atlas-app-toolkit/query"
)

func (s *myServiceImpl) MyMethod(ctx context.Context, req *MyRequest) (*MyResponse, error) {
    if sort, err := gateway.Sorting(ctx); err != nil {
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
`gateway.UnaryServerInterceptor` server interceptor.

See documentation in [query package](query/sorting.go)

#### Filtering

A service may implement filtering. A collection of response resources can be filtered by a logical expression string that includes JSON tag references to values in each resource, literal values, and logical operators. If a resource does not have the specified tag, its value is assumed to be null.


| Request Parameter | Description                              |
| ----------------- |------------------------------------------|
| _filter           | A string expression containing JSON tags, literal values, and logical operators. |

Literal values include numbers (integer and floating-point), and quoted (both single- or double-quoted) literal strings, and “null”. The following operators are commonly used in filter expressions.

| Operator     | Description              | Example                                                  |
| ------------ |--------------------------|----------------------------------------------------------|
| == | eq      | Equal                    | city == ‘Santa Clara’                                    |
| != | ne      | Not Equal                | city != null                                             |
| > | gt       | Greater Than             | price > 20                                               |
| >= | ge      | Greater Than or Equal To | price >= 10                                              |
| < | lt       | Less Than                | price < 20                                               |
| <= | le      | Less Than or Equal To    | price <= 100                                             |
| and          | Logical AND              | price <= 200 and price > 3.5                             |
| ~ | match    | Matches Regex            | name ~ “john .*”                                         |
| !~ | nomatch | Does Not Match Regex     | name !~ “john .*”                                        |
| or           | Logical OR               | price <= 3.5 or price > 200                              |
| not          | Logical NOT              | not price <= 3.5                                         |
| ()           | Grouping                 | (priority == 1 or city == ‘Santa Clara’) and price > 100 |

Usage of filtering features from the toolkit is similar to [sorting](#sorting).

Note: if you decide to use toolkit provided `infoblox.api.Filtering` proto type, then you'll not be able to use swagger schema generation, since it's plugin doesn't work with recursive nature of `infoblox.api.Filtering`.

##### How to define filtering in my request?

```proto
import github.com/infobloxopen/atlas-app-toolkit/query/collection_operators.proto;

message MyRequest {
    infoblox.api.Filtering filter = 1;
}
```

#### Pagination

A service may implement pagination of collections. Pagination of response resources can be client-driven, server-driven, or both. 

Client-driven pagination is a model in which rows are addressable by offset and page size. This scheme is similar to SQL query pagination where row offset and page size determine the rows in the query response. 

Server-driven pagination is a model in which the server returns some amount of data along with a token indicating there is more data and where subsequent queries can get the next page of data. This scheme is used by AWS Dynamo where, depending on the individual resource size, pages can run into thousands of resources. 

Some data sources can provide the number of resources a query will generate, while others cannot.

The paging model provided by the service is influenced by the expectations of the client. GUI clients prefer moderate page sizes, say no more than 1,000 resources per page. A “streaming” client may be able to consume tens of thousands of resources per page.

Consider the service behavior when no paging parameters are in the request. Some services may provide all the resources unpaged, while other services may have a default page size and provide the first page of data. In either case, the service should document its paging behavior in the absence of paging parameters.

Consider the service behavior when the query sorts or filters the data, and the underlying data is changing over time. A service may cache some amount of sorted and filtered data to be paged using client-driven paging, particularly in a GUI context. There is a trade-off between paged data coherence and changes to the data. The cache expiration time attempts to balance these competing factors.


| Paging Mode            | Request Parameters | Response Parameters | Description                                                  |
| ---------------------- |--------------------|---------------------|--------------------------------------------------------------|
| Client-driven paging   | _offset            |                     | The integer index (zero-origin) of the offset into a collection of resources. If omitted or null the value is assumed to be “0”. |
|                        | _limit             |                     | The integer number of resources to be returned in the response. The service may impose maximum value. If omitted the service may impose a default value. |
|                        |                    | _offset             | The service may optionally* include the offset of the next page of resources. A null value indicates no more pages. |
|                        |                    | _size               | The service may optionally include the total number of resources being paged. |
| Server-driven paging   | _page_token        |                     | The service-defined string used to identify a page of resources. A null value indicates the first page. |
|                        |                    | _page_token         | The service response should contain a string to indicate the next page of resources. A null value indicates no more pages. |
|                        |                    | _size               | The service may optionally include the total number of resources being paged. |
| Composite paging       | _page_token        |                     | The service-defined string used to identify a page of resources. A null value indicates the first page. |
|                        | _offset            |                     | The integer index (zero-origin) of the offset into a collection of resources in the page defined by the page token. If omitted or null the value is assumed to be “0”. |
|                        | _limit             |                     | The integer number of resources to be returned in the response. The service may impose maximum value. If omitted the service may impose a default value. |
|                        |                    | _page_token         | The service response should contain a string to indicate the next page of resources. A null value indicates no more pages. |
|                        |                    | _offset             | The service should include the offset of the next page of resources in the page defined by the page token. A null value indicates no more pages, at which point the client should request the page token in the response to get the next page. |
|                        |                    | _size               | The service may optionally include the total number of resources being paged. |

Note: Response offsets are optional since the client can often keep state on the last offset/limit request.

##### How to define pagination in my request/response?

```proto
import github.com/infobloxopen/atlas-app-toolkit/query/collection_operators.proto;

message MyRequest {
    infoblox.api.Pagination paging = 1;
}

message MyResponse {
  infoblox.api.PageInfo page = 1;
}
```
