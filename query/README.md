# Query Package

## Collection Operators

For methods that return collections, operations may be implemented using the following conventions.
The operations are implied by request parameters in query strings.
 In some cases, stateful operational information may be passed in responses.
Toolkit introduces a set of common request parameters that can be used to control
the way collections are returned. API toolkit provides some convenience methods
to support these parameters in your application.

### How can I add support for collection operators in my gRPC-Gateway?

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
import "github.com/infobloxopen/atlas-app-toolkit/query/collection_operators.proto";

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

### How can I apply collection operators passed to my GRPC service to a GORM query?

You can use `ApplyCollectionOperators` methods from [op/gorm](gateway/gorm) package.

```golang
...
f := &query.Filtering{}
s := &query.Sorting{}
p := &query.Pagination{}
fs := &query.FieldSelection{}
gormDB, err = ApplyCollectionOperators(ctx, gormDB, &PersonORM{}, &Person{}, f, s, p, fs)
if err != nil {
    ...
}
var people []Person
gormDB.Find(&people)
...
```

Separate methods per each collection operator are also available.

Check out [example](example/tagging/service.go) and [implementation](gateway/gorm/collection_operators.go).

## Field Selection

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

### How to define field selection in my request?

```proto
import github.com/infobloxopen/atlas-app-toolkit/query/collection_operators.proto;

message MyRequest {
    infoblox.api.FieldSelection fields = 1;
}
```

## Sorting

A service may implement collection sorting. A collection of response resources can be sorted by their JSON tags. For a “flat” resource, the tag name is straightforward. If sorting is allowed on non-flat hierarchical resources, the service should implement a qualified naming scheme such as dot-qualification to reference data down the hierarchy. If a resource does not have the specified tag, its value is assumed to be null.

| Request Parameter | Description                              |
| ----------------- |------------------------------------------|
| _order_by         | A comma-separated list of JSON tag names. The sort direction can be specified by a suffix separated by whitespace before the tag name. The suffix “asc” sorts the data in ascending order. The suffix “desc” sorts the data in descending order. If no suffix is specified the data is sorted in ascending order. |

### How to define sorting in my request?

```proto
import github.com/infobloxopen/atlas-app-toolkit/query/collection_operators.proto;

message MyRequest {
    infoblox.api.Sorting sort = 1;
}
```

### How can I get sorting operator on my gRPC service?

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

## Filtering

A service may implement filtering. A collection of response resources can be filtered by a logical expression string that includes JSON tag references to values in each resource, literal values, and logical operators. If a resource does not have the specified tag, its value is assumed to be null.

| Request Parameter | Description                              |
| ----------------- |------------------------------------------|
| _filter           | A string expression containing JSON tags, literal values, and logical operators. |

Literal values include numbers (integer and floating-point), quoted (both single- or double-quoted) literal strings,  “null” , arrays with numbers (integer and floating-point) and arrays with quoted (both single- or double-quoted) literal strings. The following operators are commonly used in filter expressions.

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
| := | ieq     | Insensitive equal        | city := 'SaNtA ClArA'                                    |
| in           | Check existence in set   | city in [‘Santa Clara’, ‘New York’] or  price in [1,2,3] |

Usage of filtering features from the toolkit is similar to [sorting](#sorting).

Note: if you decide to use toolkit provided `infoblox.api.Filtering` proto type, then you'll not be able to use swagger schema generation, since it's plugin doesn't work with recursive nature of `infoblox.api.Filtering`.

## How to define filtering in my request?

```proto
import github.com/infobloxopen/atlas-app-toolkit/query/collection_operators.proto;

message MyRequest {
    infoblox.api.Filtering filter = 1;
}
```

## Pagination

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

### How to define pagination in my request/response?

```proto
import github.com/infobloxopen/atlas-app-toolkit/query/collection_operators.proto;

message MyRequest {
    infoblox.api.Pagination paging = 1;
}

message MyResponse {
  infoblox.api.PageInfo page = 1;
}
```

### Field Presence

Using the toolkit's Server Wrapper functionality, you can optionally enable
automatic filling of a FieldMask within the gateway.

As a prerequisite, the request passing through the gateway must match the list
of given HTTP methods (e.g. POST, PUT, PATCH) and contain a FieldMask at the 
top level.
```proto
import "google/protobuf/field_mask.proto";
message MyRequest {
  bytes data = 1;
  google.protobuf.FieldMask fields = 2;
}
```

To enable the functionality, use the following args in the `WithGateway` method:
```golang
server.WithGateway(
  gateway.WithGatewayOptions(
    runtime.WithMetadata(gateway.NewPresenceAnnotator("POST", ...)),
    ...
  ),
  gateway.WithDialOptions(
    []grpc.DialOption{grpc.WithInsecure(), grpc.WithUnaryInterceptor(
      grpc_middleware.ChainUnaryClient(
        []grpc.UnaryClientInterceptor{gateway.ClientUnaryInterceptor, gateway.PresenceClientInterceptor()}...)},
      ),
  )
)
```