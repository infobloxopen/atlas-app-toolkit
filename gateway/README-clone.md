# Gateway

This package contains helper functions that support creating, configuring, and running a gRPC REST gateway that is REST syntax compliant. Google already provides a lot of documentation related to the gRPC gateway, so this README will mostly serve to link to existing docs.


- Starting and stopping Docker containers inside Go tests
- Creating JSON Web Tokens for testing gRPC and REST requests
- Launching a Postgres database to tests against
- Building and executing Go binaries
 
The gRPC gateway `protoc` plugin allows a gRPC service to. The official gRPC gateway repository offers a nice, succinct explanation.

> gRPC is great — it generates API clients and server stubs in many programming languages, it is fast, easy-to-use, bandwidth-efficient and its design is combat-proven by Google. However, you might still want to provide a traditional RESTful API as well. Reasons can range from maintaining backwards-compatibility, supporting languages or clients not well supported by gRPC to simply maintaining the aesthetics and tooling involved with a RESTful architecture.


## Define REST Endpoints in Proto Schema

You can map your gRPC service methods to one or more REST API endpoints. Google's official gRPC documentation has several great examples [here](https://cloud.google.com/service-management/reference/rpc/google.api#http).

Note that it is possible to define multiple HTTP methods for one RPC by using the `additional_bindings` option.

```proto
service Messaging {
  rpc GetMessage(GetMessageRequest) returns (Message) {
    option (google.api.http) = {
      get: "/v1/messages/{message_id}"
        additional_bindings {
          get: "/v1/users/{user_id}/messages/{message_id}"
        }
      };
    }
  }
}

message GetMessageRequest {
  string message_id = 1;
  string user_id = 2;
}
```
This enables the following two alternative HTTP JSON to RPC mappings: 

| HTTP Verb | REST Endpoint                     | RPC                                                 |
| ----------|-----------------------------------|---------------------------------------------------- |
| `GET`     | `/v1/messages/123456`             | `GetMessage("123456")`                  |
| `GET`     | `/v1/users/me/messages/123456`    | `GetMessage("me", "123456")`    |


## HTTP Headers

Your application or service might depend on HTTP headers from incoming REST requests. The official gRPC gateway documentation describes how to handle HTTP headers in detail, so check out the documentation [here](https://grpc-ecosystem.github.io/grpc-gateway/docs/customizingyourgateway.html).

### Using Headers in gRPC Service

To extract headers from metadata, you can use the [`FromIncomingContext`](https://godoc.org/google.golang.org/grpc/metadata#FromIncomingContext) function.

```go
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

You can also use the helper function provided in this package.

```go
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

### Adding Headers to REST Response

To send metadata from the gRPC server to the REST client, you need to use the [`SetHeader`](https://godoc.org/google.golang.org/grpc#SetHeader) function.

```go
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

If you do not use any custom outgoing header matcher, you will see something like this.

```sh
$ curl -i http://localhost:8080/contacts/v1/contacts

HTTP/1.1 200 OK
Content-Type: application/json
Grpc-Metadata-Myheader: myvalue
Date: Wed, 31 Jan 2018 15:28:52 GMT
Content-Length: 2

{}
```

## Responses

You may need to modify the HTTP response body returned by the gRPC gateway. For instance, the gRPC Gateway translates non-error gRPC responses into `200 - OK` HTTP responses, which might not suit your particular use case.

### Overwrite Default Response Forwarder

By default, an HTTP response returned by the gRPC Gateway doesn't conform to the Infoblox REST API Syntax (e.g. it has no `success` section).

To override this behavior, the gRPC Gateway documentation recommends overwriting `ForwardResponseMessage` and `ForwardResponseStream` functions correspondingly. See [this documentation](https://github.com/grpc-ecosystem/grpc-gateway/wiki/How-to-customize-your-gateway#replace-a-response-forwarder-per-method) for further information.

```go
import (
	"github.com/infobloxopen/atlas-app-toolkit/gateway"
)

func init() {
	forward_App_ListObjects_0 = gateway.ForwardResponseMessage
}
```

You can also refer [example app](https://github.com/infobloxopen/atlas-contacts-app/blob/master/pkg/pb/contacts.overwrite.pb.gw.go).

#### Conforming to REST API Syntax

We made default [`ForwardResponseMessageFunc`](response.go#L21) and [`ForwardResponseStreamFunc`](response.go#L21)
implementations that conform REST API Syntax.

_Note: the forwarders still set `200 - OK` as HTTP status code if no errors are encountered._

### Setting HTTP Status Codes

In order to set HTTP status codes properly, you need to send metadata from your gRPC service so that default forwarders will be able to read them and set codes. This is a common approach in gRPC to send extra information for response as metadata.

We recommend using the [gRPC status package](https://godoc.org/google.golang.org/grpc/status) and our custom function [`SetStatus`](status.go#L49) to add extra metadata to the gRPC response.

More documentation is available in the [`status`](status.go) package.

Also you may use shortcuts like `SetCreated`, `SetUpdated`, and `SetDeleted`.

```go
import (
    "github.com/infobloxopen/atlas-app-toolkit/gateway"
)

func (s *myService) MyMethod(req *MyRequest) (*MyResponse, error) {
    err := gateway.SetCreated(ctx, "created 1 item")
    return &MyResponse{Result: []*Item{item}}, err
}
```

### Response Format
Unless another format is specified in the request `Accept` header that the service supports, services render resources in responses in JSON format by default.

Services must embed their response in a Success JSON structure.

The Success JSON structure provides a uniform structure for expressing normal responses using a structure similar to the Error JSON structure used to render errors. The structure provides an enumerated set of codes and associated HTTP statuses (see Errors below) along with a message.

The Success JSON structure has the following format. The results tag is optional and appears when the response contains one or more resources.
```json
{
  "success": {
    "status": <http-status-code>,
    "code": <enumerated-error-code>,
    "message": <message-text>
  },
  "results": <service-response>
}
```

The `results` content follows the [Google model](https://cloud.google.com/apis/design/standard_methods): an object is returned for Get, Create and Update operations and list of objects for List operation.

To allow compatibility with existing systems, the results tag name can be changed to a service-defined tag. In this way the success data becomes just a tag added to an existing structure.

#### Example Success Responses

Response with no results
```json
{
  "success": {
    "status": 201,
    "message": "Account provisioned",
    "code": "CREATED"
  }
}
```

Response with results
```json
{
  "success": {
    "status": 200,
    "message": "Found 2 items",
    "code": "OK"
  },
  "results": [
    {
      "account_id": 4,
      "created_at": "2018-01-06T03:53:27.651Z",
      "updated_at": "2018-01-06T03:53:27.651Z",
      "account_number": null,
      "sfdc_account_id": "3",
      "id": 5
    },
    {
      "account_id": 31,
      "created_at": "2018-01-06T04:38:32.572Z",
      "updated_at": "2018-01-06T04:38:32.572Z",
      "account_number": null,
      "sfdc_account_id": "1",
      "id": 9
    }
  ]
}
```

Response for get by id operation
```json
{
  "success": {
    "status": 200,
    "message": "object found",
    "code": "OK"
  },
  "results": {
      "account_id": 4,
      "created_at": "2018-05-06T03:53:27.651Z",
      "updated_at": "2018-05-06T03:53:27.651Z",
      "id": 5
   }
}
```

Response with results and service-defined results tag `rpz_hits`
```json
{
  "success": {
    "status": 200,
    "message": "Read 360 items",
    "code": "OK"
  },
  "rpz_hits": [
    {
      "pid": "default",
      "rip": "10.35.205.4",
      "policy_name": "Default",
      "ttl": -1,
      "qtype": 1,
      "qip": "10.120.20.247",
      "confidence": 3,
      "network": "on-prem",
      "event_time": "2017-12-13T07:07:50.000Z",
      "feed_name": "rpz",
      "dsource": 1,
      "rcode": 3,
      "timestamp": "11e7-dfd4-54e564f0-0000-0000287cd227",
      "company": "302002|0",
      "feed_type": "0",
      "user": "unknown",
      "device": "10.120.20.247",
      "severity": 3,
      "country": "unknown",
      "policy_action": "Block",
      "qname": "barfywyjgx.com",
      "tproperty": "A",
      "tclass": "UNKNOWN"
    },
    ...
  ]
}
```

## Errors

### Format
Method error responses are rendered in the Error JSON format. The Error JSON format is similar to the Success JSON format.

The Error JSON structure has the following format. The details tag is optional and appears when the service provides more details about the error.
```json
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
  ],
  "fields": {
      "field1": [<message-1>, <message-2>, ...],
      "field2": ...,
  }
}
```

### Translating gRPC Errors to HTTP

To respond with an error message that is REST API syntax-compliant, you can write your own `ProtoErrorHandler` or use `DefaultProtoErrorHandler` provided in this package.

Passing errors from the gRPC service to the REST client is supported by the gRPC gateway, so see the gRPC gateway documentation [here](https://mycodesmells.com/post/grpc-gateway-error-handler).

Here's an example that shows how to use [`DefaultProtoErrorHandler`](gateway/errors.go#L25).

```go
import (
    "github.com/grpc-ecosystem/grpc-gateway/runtime"
    "github.com/infobloxopen/atlas-app-toolkit/gateway"

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

### Sending Error Details

The idiomatic way to send an error from you gRPC service is to simple return
it from you gRPC handler either as `status.Errorf()` or `errors.New()`.

```go
import (
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

func (s *myServiceImpl) MyMethod(req *MyRequest) (*MyResponse, error) {
    return nil, status.Errorf(codes.Unimplemented, "method is not implemented: %v", req)
}
```
