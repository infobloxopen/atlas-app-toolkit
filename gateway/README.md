# Gateway Package

### Resources and Collections

#### How to define REST API Endpoints in my proto scheme?

You can map your gRPC service methods to one or more REST API endpoints.
See [this reference](https://cloud.google.com/service-management/reference/rpc/google.api#http) how to do it.

It is possible to define multiple HTTP methods for one RPC by using the `additional_bindings` option. Example:

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

| HTTP | RPC
| -----|-----
| `GET /v1/messages/123456` | `GetMessage(message_id: "123456")`
| `GET /v1/users/me/messages/123456` | `GetMessage(user_id: "me" message_id: "123456")`


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

`results` content follows the [Google model](https://cloud.google.com/apis/design/standard_methods): an object is returned for Get, Create and Update operations and list of objects for List operation.

To allow compatibility with existing systems, the results tag name can be changed to a service-defined tag. In this way the success data becomes just a tag added to an existing structure.

#### Examples:

Response with no results:
```
{
  "success": {
    "status": 201,
    "message": "Account provisioned",
    "code": "CREATED"
  }
}
```

Response with results:
```
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

Response for get by id operation:
```
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

Response with results and service-defined results tag “rpz_hits”:
```
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
  ],
  "fields": {
      "field1": [<message-1>, <message-2>, ...],
      "field2": ...,
}
```

#### How can I convert a gRPC error to a HTTP error response in accordance with REST API Syntax Specification?

You can write your own `ProtoErrorHandler` or use `gateway.DefaultProtoErrorHandler` one.

How to handle error on gRPC-Gateway see [article](https://mycodesmells.com/post/grpc-gateway-error-handler)

How to use [gateway.DefaultProtoErrorHandler](gateway/errors.go#L25) see example below:

```golang
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
