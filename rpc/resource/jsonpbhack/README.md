
## Justification for this package

Some Github repos have been migrated to **gRPC-Gateway** [**v2**](https://github.com/grpc-ecosystem/grpc-gateway) that probides its own JSONPb marshaller. So did **gRPC-Gateway** of [**v1**](https://github.com/grpc-ecosystem/grpc-gateway/tree/v1) but they are breakingly different underneath.

JSONPb of v1 relies on [jsonpb](https://pkg.go.dev/github.com/golang/protobuf/jsonpb) library that had been deprecated. It uses Go's standard JSON library for JSON encoding/decoding which allowed us specifying our custom (un)marshaller for selected types. For example,  heavily used [**atlas-app-toolkit**](https://github.com/infobloxopen/atlas-app-toolkit) implements (Un)Marshaller interface for [Identifier](https://github.com/infobloxopen/atlas-app-toolkit/tree/master/rpc/resource) type that makes on-the-fly conversion of id object in requests' and responses' payload, so
```json
"id":{
    "applicationName":"atlas.contacts",
    "resourceId":"z7061a91-ae34-4aa5-a9b1-a9e9979736d7",
    "resourceType":"contact"
},
...
```
gets
```json
"id":"atlas.contacts/contact/z7061a91-ae34-4aa5-a9b1-a9e9979736d7",
```
and vice versa.

JSONPb of v2 discards jsonpb deprecated library and depends on [protojson](https://pkg.go.dev/google.golang.org/protobuf/encoding/protojson) library instead. Protojson uses its own implementation of JSON (un)marshaller. Because of this and also for the authors objective to stick to proto specification exactly, there is no way to wedge in our own (un)marshaller easily. They deliberately prohibit it. See [this](https://github.com/golang/protobuf/issues/1098) issue for example.

> The canonical JSON mapping is a portable format that can be consumed by many different protobuf implementations in many different languages. Producing a custom, non-portable encoding is out of scope for the protojson package.
The easiest option is to continue using deprecated jsonpb that we can specify via gateway options. The rightest approach is probably to follow along with protojson and update our services to use composite identifier.

To overcome such incompatibility where a lot of parties depend on JSON encoding/decoding adjustments of payload until they are refactored, the solution to the problem is to use JSONPb from gRPC-Gateway v1 and specify it as the Marshaller to use for the time being. **_jsonhack.go_** file is mostly a copy of [marshal_json.pb](https://github.com/grpc-ecosystem/grpc-gateway/blob/v1/runtime/marshal_jsonpb.go) file from v1 modified for to be accepted by v2 as an implementation of [Marshaller](https://github.com/grpc-ecosystem/grpc-gateway/blob/master/runtime/marshaler.go) interface:
```golang
runtime2.WithMarshalerOption(runtime2.MIMEWildcard, &jsonpbhack.JSONPb{
		OrigName:     true,
		EmitDefaults: true,
	},
)
```

Feel free to replace it with proper v2 Marshaller after aforementioned changes are implemented:
```golang 
runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
	MarshalOptions: protojson.MarshalOptions{
		UseProtoNames:   true,
		EmitUnpopulated: true,
	},
}),
```

