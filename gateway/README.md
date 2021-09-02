# Gateway

This package contains helper functions that support creating, configuring, and running a gRPC REST gateway that is REST syntax compliant. Google already provides a lot of documentation related to the gRPC gateway, so this README will mostly serve to link to existing docs.

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
| `GET`     | `/v1/messages/123456`             | `GetMessage(message_id: "123456")`                  |
| `GET`     | `/v1/users/me/messages/123456`    | `GetMessage(user_id: "me", message_id: "123456")`    |


## HTTP Headers

Your application or service might depend on HTTP headers from incoming REST requests. The official gRPC gateway documentation describes how to handle HTTP headers in detail, so check out the documentation [here](https://grpc-ecosystem.github.io/grpc-gateway/docs/mapping/customizing_your_gateway/).

### Using Headers in gRPC Service

To extract headers from metadata, you can use the [`FromIncomingContext`](https://pkg.go.dev/google.golang.org/grpc/metadata#FromIncomingContext) function.

```go
import (
    "context"

    "google.golang.org/grpc/metadata"
    "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
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
$ curl -i http://localhost:8080/resource

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

#### Complying with Infoblox's REST API Syntax

We made default [`ForwardResponseMessageFunc`](response.go#L21) and [`ForwardResponseStreamFunc`](response.go#L21)
implementations that conform to Infoblox's REST API Syntax guidelines. These helper functions ensure that Infoblox teams who use toolkit follow the same REST API conventions. For non-Infoblox toolkit users, these are completely optional utilities.

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

By default for a successful RPC call only the proto response is rendered as JSON, however for a failed call a special format is used, and by calling special methods the response can include additional metadata.

The `WithSuccess(ctx context.Context, msg MessageWithFields)` function allows you to add a `success` block to the returned JSON.
By default this block only contains a message field, however arbitrary key-value pairs can also be included.
This is included at top level, alongside the assumed `result` or `results` field.

Ex.
```json
{
  "success": {
    "foo": "bar",
    "baz": 1,
    "message": <message-text>
  },
  "results": <service-response>
}
```

The `WithError(ctx context.Context, err error)` function allows you to add an extra `error` to the `errors` list in the returned JSON.
The `NewWithFields(message string, kvpairs ...interface{})` function can be used to create this error, which then includes additional fields in the error, otherwise only the error message will be included.
This is included at top level, alongside the assumed `result` or `results` field if the call succeeded despite the error, or alone otherwise.

Ex.
```json
{
  "errors": [
    {
      "foo": "bar",
      "baz": 1,
      "message": <message-text>
    }
  ],
  "results": <service-response>
}
```

To return an error with fields and fail the RPC, return an error from `NewResponseError(ctx context.Context, msg string, kvpairs ...interface{})` or `NewResponseErrorWithCode(ctx context.Context, c codes.Code, msg string, kvpairs ...interface{})` to also set the return code.

The function `IncludeStatusDetails(withDetails bool)` allows you to include the `success` block with fields `code` and `status` automatically for all responses,
and the first of the `errors` in failed responses will also include the fields.
Note that this choice affects all responses that pass through `gateway.ForwardResponseMessage`.

Ex:
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

#### Example Success Responses With IncludeStatusDetails(true)

Response with no results
```json
{
  "success": {
    "status": "CREATED",
    "message": "Account provisioned",
    "code": 201
  }
}
```

Response with results
```json
{
  "success": {
    "status": "OK",
    "message": "Found 2 items",
    "code": 200
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
    "status": "OK",
    "message": "object found",
    "code": 200
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
    "status": "OK",
    "message": "Read 360 items",
    "code": 200
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

## Query String Filtering
When using the collection operators with the grpc-gateway, extraneous errors may
be logged during rpcs as the query string is parsed that look like this:
```
field not found in *foo.ListFoobarRequest: _order_by
```
and the usage of any of the collection operator field names without the leading
underscore (`order_by`, `filter`,... instead of `_order_by`, `filter`,...) in
query strings may result in the error `unsupported field type reflect.Value`,
being returned.

This can be resolved by overwriting the default filter for each rpc with these
operators using the one defined in [filter.go](filter.go).
```golang
filter_Foobar_List_0 = gateway.DefaultQueryFilter
```


### Translating gRPC Errors to HTTP

To respond with an error message that is REST API syntax-compliant, you can write your own `ProtoErrorHandler` or use `DefaultProtoErrorHandler` provided in this package.

Passing errors from the gRPC service to the REST client is supported by the gRPC gateway, so see the gRPC gateway documentation [here](https://mycodesmells.com/post/grpc-gateway-error-handler).

Here's an example that shows how to use [`DefaultProtoErrorHandler`](gateway/errors.go#L25).

```go
import (
    "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
    "github.com/infobloxopen/atlas-app-toolkit/gateway"

    "github.com/yourrepo/yourapp"
)

func main() {
    // create error handler option
    errHandler := runtime.WithErrorHandler(gateway.DefaultProtoErrorHandler)

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
If additional fields are required, then use the
`NewResponseError(ctx context.Context, msg string, kvpairs ...interface{})` or
`NewResponseErrorWithCode(ctx context.Context, c codes.Code, msg string, kvpairs ...interface{})` functions instead.

```go
import (
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

func (s *myServiceImpl) MyMethod(req *MyRequest) (*MyResponse, error) {
    return nil, status.Errorf(codes.Unimplemented, "method is not implemented: %v", req)
}

func (s *myServiceImpl) MyMethod2(ctx context.Context, req *MyRequest) (*MyResponse, error) {
    return nil, NewResponseErrorWithCode(ctx, codes.Internal, "something broke on our end", "retry_in", 30)
}
```
