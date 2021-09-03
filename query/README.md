# Query

This package offers a set of predefined [protobuf](https://developers.google.com/protocol-buffers/) message types to be used as collection query parameters, which are also called *collection operators*.
These types are:
- `infoblox.api.Filtering`
- `infoblox.api.Sorting`
- `infoblox.api.Pagination`
- `infoblox.api.PageInfo`(used in response)
- `infoblox.api.FieldSelection`

## Enabling *collection operators* in your application

In order to get *collection operators* in your app you need the following:

- Add *collection operator* types to a request message. You're free to use any subset of them.

```proto
import "github.com/infobloxopen/atlas-app-toolkit/query/collection_operators.proto";

message MyListRequest {
    infoblox.api.Filtering filtering = 1;
    infoblox.api.Sorting sorting = 2;
    infoblox.api.Pagination pagination = 3;
    infoblox.api.FieldSelection fields = 4;
}
```

- Enable `gateway.ClientUnaryInterceptor` and `gateway.MetadataAnnotator` in your gRPC gateway. This will make *collection operators* to be automatically parsed on gRPC gateway side.
If you're using our [server](../server) wrapper, you need to explicitly set only `gateway.ClientUnaryInterceptor` since `gateway.MetadataAnnotator` is enabled by default.

```golang
server.WithGateway(
  gateway.WithDialOptions(
    []grpc.DialOption{grpc.WithInsecure(), grpc.WithUnaryInterceptor(
      grpc_middleware.ChainUnaryClient(
        []grpc.UnaryClientInterceptor{gateway.ClientUnaryInterceptor...})},
      ),
  )
)
```

## Filtering

The syntax of REST representation of `infoblox.api.Filtering` is the following.

| REST Query Parameter | Description                           |
| -------------------- |------------------------------------------|
| _filter              | A string expression containing JSON tags, literal values, and logical operators. |

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

In order to escape string literal delimiter duplicate it, e.g. for single-quoted string literals: `_filter=field == 'dup single quote '' '`, for double-quoted literals: `_filter=field == "dup double quote "" "`.


### JSONB Filtering
Also our filtering support custom `jsonb` conditions if you use `postgresql` as a database. In order to use this feathure you have to use special type [Jsonb](https://github.com/jinzhu/gorm/blob/master/dialects/postgres/postgres.go). If you plan to use special jsonb filtering on a field it's type should be `*dialects.Jsonb`.
Syntax for jsonb filtering is similar to the one you use for `StringCondition` and `StringArrayCondition`, as well as you can use operators from `StringCondition` and `StringArrayCondition`.

#### Example:
Consider that you have a following object where `info` field is a `jsonb` field:
```
{
 name: "my object",
 info: {
     Address: {
         "City": "Tacoma",
         "Country": "USA"
     }
 }
}
```
Filterting request with `jsonb` query will look like:
```
...?_filter=info.Address.City=='Tacoma'
```
if you want to compare enclosed object with json you are able to write:
```
...?_filter=info.Address=='{"City": "Tacoma", "Country": "USA"}'
```

Note: if you decide to use toolkit provided `infoblox.api.Filtering` proto type, then you'll not be able to use [vanilla](https://github.com/grpc-ecosystem/grpc-gateway/tree/master/protoc-gen-openapiv2) swagger schema generation, since this plugin doesn't work with recursive nature of `infoblox.api.Filtering`.
In this case you can use our [fork](https://github.com/infobloxopen/grpc-gateway/tree/v2/protoc-gen-openapiv2) which has a fix for this issue.
You can also use [atlas-gentool](https://github.com/infobloxopen/atlas-gentool) which contains both versions of the plugin.

## Sorting

The syntax of REST representation of `infoblox.api.Sorting` is the following.

| Request Parameter | Description                              | Example |
| ----------------- |------------------------------------------| ------- |
| _order_by         | A comma-separated list of JSON tag names. The sort direction can be specified by a suffix separated by whitespace before the tag name. The suffix “asc” sorts the data in ascending order. The suffix “desc” sorts the data in descending order. If no suffix is specified the data is sorted in ascending order. | work_address.addresss desc,first_name |

## Pagination

The syntax of REST representation of `infoblox.api.Pagination` and `infoblox.api.PageInfo` is the following.

| Paging Mode            | Request Parameters | Response Parameters | Description                                                  |
| ---------------------- |--------------------|---------------------|--------------------------------------------------------------|
| Client-driven paging   | _offset            |                     | The integer index (zero-origin) of the offset into a collection of resources. If omitted or null the value is assumed to be “0”. |
|                        | _limit             |                     | The integer number of resources to be returned in the response. The service may impose maximum value. If omitted the service may impose a default value. |
|                        |                    | _offset             | The service may optionally* include the offset of the next page of resources. A null value indicates no more pages. |
|                        |                    | _size               | The service may optionally include the total number of resources being paged. |
| Server-driven paging   | _page_token        |                     | The service-defined string used to identify a page of resources. A null value indicates the first page. |
|                        |                    | _page_token         | The service response should contain a string to indicate the next page of resources. A null value indicates no more pages. |
|                        |                    | _size               | The service may optionally include the total number of resources being paged. |

## Field Selection

The syntax of REST representation of `infoblox.api.FieldSelection` is the following.

| Request Parameter | Description                              | Example |
| ----------------- |------------------------------------------| ------- |
| _fields           | A comma-separated list of JSON tag names.| work_address.addresss,first_name |

As it is not possible to completely remove all the fields(such as primitives) from `proto.Message` on gRPC server side, fields are additionally truncated on gRPC Gateway side.
This is done by `gateway.ResponseForwarder`.

## Field Presence

Using the toolkit's [server](../server) package functionality, you can optionally enable automatic filling of a `google.protobuf.FieldMask` within the gRPC Gateway.

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
