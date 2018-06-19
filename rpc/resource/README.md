# Description
A number of application services require a mechanism to encode the identity of a particular resource across applications.
The resource identifiers consist of the application ID, an application-defined resource type, and an application-defined ID for that resource.
The reference format captures the same information as the three-tuple format, in a single string delimited by `/`:

```
<app_id>/<resource_type>/<resource_id>
```

# How to define in Proto?

The common representation of resource identifiers defined in Protocol Buffer format
you could find in [resource.proto](resourcepb/resource.proto) file.
The `message Identifier` implements `jsonpb.JSONPBMarshaler` and `jsonpb.JSONPBUnmarshaler`
interfaces so that it renders itself in JSON as a string in a single string delimited by `/`.

You could use it to define identifiers in your proto messages, e.g

```proto
import "github.com/infobloxopen/atlas-app-toolkit/rpc/resouce/resourcepb/resource.proto"

message MyMessage {
    infoblox.rpc.Identifier id = 1;
    string some_data = 2;
    infoblox.rpc.Identifier external_resource = 3;
}
```

# How to use in Golang?

The common interfaces and helper functions to work with identifiers are defined in [resource](resource.go) file.
There are also some common implementations. 

The [fq](fq/resource.go) package provides a codec to encode and decode [Protocol Buffer representation](resourcepb/resource.proto) of 
identifiers to and from `Identifier`.
The implementation of `identifier` supports `sql.Scanner` and `driver.Valuer` interfaces
so it could be stored in SQL DB as a single string delimited by `/` (in **f**ully **q**ualified from).

The [uuid](uuid/resource.go) package provides a codec to encode and decode [Protocol Buffer representation](resourcepb/resource.proto) of 
identifiers to and from `Identifier`.
The implementation of `identifier` supports `sql.Scanner` and `driver.Valuer` interfaces
so that could be stored in SQL DB as a string with **ONLY** Resource ID.

Please see [example](example_test.go) to see how to register codecs and encode/decode identifiers based on resource types of proto messages.
